package chain

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/datum"
)

// TODO: More documentation.

// JobChain manages a chain of jobs.
type JobChain struct {
	pachClient *client.APIClient
	hasher     datum.Hasher
	base       datum.Iterator
	noSkip     bool
	prevJob    *JobDatumIterator
}

// NewJobChain creates a new job chain.
// TODO: We should probably pipe a context through here.
func NewJobChain(pachClient *client.APIClient, hasher datum.Hasher, opts ...JobChainOption) *JobChain {
	jc := &JobChain{
		pachClient: pachClient,
		hasher:     hasher,
	}
	for _, opt := range opts {
		opt(jc)
	}
	if jc.base != nil {
		// Insert a dummy job representing the given base datum set
		jdi := &JobDatumIterator{
			jc:        jc,
			dit:       jc.base,
			outputDit: jc.base,
			done:      make(chan struct{}),
		}
		close(jdi.done)
		jc.prevJob = jdi
	}
	return jc
}

// CreateJob creates a job in the job chain.
// TODO: Context should be associated with the iteration, but need to change datum iterator interface for that.
func (jc *JobChain) CreateJob(ctx context.Context, pipelineJobID string, dit, outputDit datum.Iterator) *JobDatumIterator {
	jdi := &JobDatumIterator{
		ctx:           ctx,
		jc:            jc,
		parent:        jc.prevJob,
		pipelineJobID: pipelineJobID,
		stats:         &datum.Stats{ProcessStats: &pps.ProcessStats{}},
		dit:           datum.NewPipelineJobIterator(dit, pipelineJobID, jc.hasher),
		outputDit:     outputDit,
		done:          make(chan struct{}),
	}
	jc.prevJob = jdi
	return jdi
}

// JobDatumIterator provides a way to iterate through the datums in a job.
type JobDatumIterator struct {
	ctx            context.Context
	jc             *JobChain
	parent         *JobDatumIterator
	pipelineJobID  string
	stats          *datum.Stats
	dit, outputDit datum.Iterator
	finishOnce     sync.Once
	done           chan struct{}
	deleter        func(*datum.Meta) error
}

// SetDeleter sets the deleter callback for the iterator.
// TODO: There should be a way to handle this through callbacks, but this would require some more changes to the registry.
func (jdi *JobDatumIterator) SetDeleter(deleter func(*datum.Meta) error) {
	jdi.deleter = deleter
}

// Iterate iterates through the datums for the job.
// This algorithm is split into two parts: before the parent job finishes and after.
// For each part, we create an output datum fileset that is lexicographically ordered with respect to the order in which the datums were output by the datum iterator (this is implemented through the datum index prefix).
// For the part before the parent job finishes, we create an output datum fileset for the datums that only exist in the current job and record the datums we could potentially skip in a datum fileset.
// For the part after the parent job finishes, we use the potentially skipped datum fileset and check it against the parent's output datum fileset to see if we can actually skip the datums. We create an output datum fileset for the datums that we are unable to skip.
// TODO: There is probably a clean way to cache the output datum filesets so we do not need to recompute them across iterations.
func (jdi *JobDatumIterator) Iterate(cb func(*datum.Meta) error) error {
	jdi.stats.Skipped = 0
	for {
		if jdi.parent == nil {
			return jdi.dit.Iterate(cb)
		}
		if err := jdi.parent.dit.Iterate(func(_ *datum.Meta) error { return nil }); err != nil {
			if pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) {
				jdi.parent = jdi.parent.parent
				continue
			}
			return err
		}
		break
	}
	pachClient := jdi.jc.pachClient.WithCtx(jdi.ctx)
	return pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		pachClient = pachClient.WithCtx(ctx)
		// Upload the datums from the current job and parent job datum iterators into the datum fileset format.
		filesetID, err := jdi.uploadDatumFileset(pachClient, jdi.dit)
		if err != nil {
			return err
		}
		renewer.Add(filesetID)
		parentFilesetID, err := jdi.uploadDatumFileset(pachClient, jdi.parent.dit)
		if err != nil {
			return err
		}
		renewer.Add(parentFilesetID)
		// Create the output datum fileset for the new datums (datums that do not exist in the parent job).
		// TODO: Logging?
		var outputFilesetID string
		skippedFilesetID, err := jdi.withDatumFileset(pachClient, func(skippedSet *datum.Set) error {
			filesetIterator := datum.NewFileSetIterator(pachClient, filesetID)
			parentFilesetIterator := datum.NewFileSetIterator(pachClient, parentFilesetID)
			var err error
			if outputFilesetID, err = jdi.withDatumFileset(pachClient, func(outputSet *datum.Set) error {
				return datum.Merge([]datum.Iterator{parentFilesetIterator, filesetIterator}, func(metas []*datum.Meta) error {
					if len(metas) == 1 {
						if metas[0].PipelineJobID != jdi.pipelineJobID {
							return nil
						}
						return outputSet.UploadMeta(metas[0], datum.WithPrefixIndex())
					}
					if jdi.skippableDatum(metas[0], metas[1]) {
						jdi.stats.Skipped++
						return skippedSet.UploadMeta(metas[0])
					}
					return outputSet.UploadMeta(metas[0], datum.WithPrefixIndex())
				})
			}); err != nil {
				return err
			}
			renewer.Add(outputFilesetID)
			return nil
		})
		if err != nil {
			return err
		}
		renewer.Add(skippedFilesetID)
		if err := datum.NewFileSetIterator(pachClient, outputFilesetID).Iterate(cb); err != nil {
			return err
		}
		select {
		case <-jdi.parent.done:
		case <-jdi.ctx.Done():
			return jdi.ctx.Err()
		}
		// Create the output datum fileset for the skipped datums that were not processed by the parent (failed, recovered, etc.).
		// Also create deletion operations appropriately.
		skippedFilesetIterator := datum.NewFileSetIterator(pachClient, skippedFilesetID)
		outputFilesetID, err = jdi.withDatumFileset(pachClient, func(s *datum.Set) error {
			return datum.Merge([]datum.Iterator{jdi.parent.outputDit, skippedFilesetIterator}, func(metas []*datum.Meta) error {
				if len(metas) == 1 {
					// Datum was skipped, but does not exist in the parent job output.
					if metas[0].PipelineJobID == jdi.pipelineJobID {
						jdi.stats.Skipped--
						return s.UploadMeta(metas[0], datum.WithPrefixIndex())
					}
					// Datum only exists in the parent job.
					return jdi.deleteDatum(metas[0])
				}
				// Check if a skipped datum was not successfully processed by the parent.
				if !jdi.skippableDatum(metas[0], metas[1]) {
					jdi.stats.Skipped--
					if err := jdi.deleteDatum(metas[1]); err != nil {
						return err
					}
					return s.UploadMeta(metas[0], datum.WithPrefixIndex())
				}
				return nil
			})
		})
		if err != nil {
			return err
		}
		renewer.Add(outputFilesetID)
		return datum.NewFileSetIterator(pachClient, outputFilesetID).Iterate(cb)
	})
}

func (jdi *JobDatumIterator) uploadDatumFileset(pachClient *client.APIClient, dit datum.Iterator) (string, error) {
	return jdi.withDatumFileset(pachClient, func(s *datum.Set) error {
		return dit.Iterate(func(meta *datum.Meta) error {
			return s.UploadMeta(meta)
		})
	})
}

func (jdi *JobDatumIterator) withDatumFileset(pachClient *client.APIClient, cb func(*datum.Set) error) (string, error) {
	resp, err := pachClient.WithCreateFilesetClient(func(mf client.ModifyFile) error {
		storageRoot := filepath.Join(os.TempDir(), "pachyderm-skipped-tmp", uuid.NewWithoutDashes())
		return datum.WithSet(nil, storageRoot, cb, datum.WithMetaOutput(mf))
	})
	if err != nil {
		return "", err
	}
	return resp.FilesetId, nil
}

func (jdi *JobDatumIterator) deleteDatum(meta *datum.Meta) error {
	if jdi.deleter == nil {
		return nil
	}
	return jdi.deleter(meta)
}

func (jdi *JobDatumIterator) skippableDatum(meta1, meta2 *datum.Meta) bool {
	if jdi.jc.noSkip {
		return false
	}
	// If the hashes are equal and the second datum was processed, then skip it.
	return meta1.Hash == meta2.Hash && meta2.State == datum.State_PROCESSED
}

// Stats returns the stats for the most recent iteration.
func (jdi *JobDatumIterator) Stats() *datum.Stats {
	return jdi.stats
}

// Finish finishes the job in the job chain.
func (jdi *JobDatumIterator) Finish() {
	jdi.finishOnce.Do(func() {
		close(jdi.done)
	})
}
