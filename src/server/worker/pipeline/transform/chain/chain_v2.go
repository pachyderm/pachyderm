package chain

import (
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
)

type DatumHasherV2 interface {
	Hash([]*common.Input) string
}

// TODO: We should probably pipe a context through here.

type JobChainV2 struct {
	hasher  DatumHasherV2
	prevJob *JobDatumIteratorV2
}

func NewJobChainV2(hasher DatumHasherV2, baseDit ...datum.IteratorV2) *JobChainV2 {
	jc := &JobChainV2{hasher: hasher}
	if len(baseDit) > 0 {
		// Insert a dummy job representing the given base datum set
		jdi := &JobDatumIteratorV2{
			jc:        jc,
			dit:       baseDit[0],
			outputDit: baseDit[0],
			done:      make(chan struct{}),
		}
		close(jdi.done)
		jc.prevJob = jdi
	}
	return jc
}

func (jc *JobChainV2) CreateJob(jobID string, dit, outputDit datum.IteratorV2) *JobDatumIteratorV2 {
	jdi := &JobDatumIteratorV2{
		jc:        jc,
		parent:    jc.prevJob,
		jobID:     jobID,
		stats:     &datum.Stats{ProcessStats: &pps.ProcessStats{}},
		dit:       datum.NewJobIterator(dit, jobID),
		outputDit: outputDit,
		done:      make(chan struct{}),
	}
	jc.prevJob = jdi
	return jdi
}

type JobDatumIteratorV2 struct {
	jc             *JobChainV2
	parent         *JobDatumIteratorV2
	jobID          string
	stats          *datum.Stats
	dit, outputDit datum.IteratorV2
	done           chan struct{}
	deleter        func(*datum.Meta) error
}

// TODO: There should be a way to handle this through callbacks, but this would require some more changes to the registry.
func (jdi *JobDatumIteratorV2) SetDeleter(deleter func(*datum.Meta) error) {
	jdi.deleter = deleter
}

func (jdi *JobDatumIteratorV2) Iterate(cb func(*datum.Meta) error) error {
	if jdi.parent == nil {
		return jdi.dit.Iterate(cb)
	}
	// Generate datum sets for the new datums (datums that do not exist in the parent job).
	// TODO: Logging?
	if err := datum.Merge([]datum.IteratorV2{jdi.dit, jdi.parent.dit}, func(metas []*datum.Meta) error {
		if len(metas) == 1 {
			if metas[0].JobID != jdi.jobID {
				return nil
			}
			return cb(metas[0])
		}
		if jdi.skippableDatum(metas[0], metas[1]) {
			jdi.stats.Skipped++
			return nil
		}
		return cb(metas[0])
	}); err != nil {
		return err
	}
	// TODO: Need a context here.
	<-jdi.parent.done
	// Generate datum sets for the skipped datums that were not processed by the parent (failed, recovered, etc.).
	// Also generate deletion operations appropriately.
	return datum.Merge([]datum.IteratorV2{jdi.dit, jdi.parent.dit, jdi.parent.outputDit}, func(metas []*datum.Meta) error {
		// Datum only exists in the current job or only exists in the parent job and was not processed.
		if len(metas) == 1 {
			return nil
		}
		if len(metas) == 2 {
			// Datum only exists in the parent job and was processed.
			if metas[0].JobID != jdi.jobID {
				return jdi.deleteDatum(metas[1])
			}
			// Datum exists in both jobs, but was not processed by the parent.
			if jdi.skippableDatum(metas[0], metas[1]) {
				jdi.stats.Skipped--
				return cb(metas[0])
			}
		}
		// Check if a skipped datum was not successfully processed by the parent.
		if jdi.skippableDatum(metas[0], metas[1]) {
			if jdi.skippableDatum(metas[1], metas[2]) {
				return nil
			}
			jdi.stats.Skipped--
			if err := cb(metas[0]); err != nil {
				return err
			}
		}
		return jdi.deleteDatum(metas[2])
	})
}

func (jdi *JobDatumIteratorV2) deleteDatum(meta *datum.Meta) error {
	if jdi.deleter == nil {
		return nil
	}
	return jdi.deleter(meta)
}

func (jdi *JobDatumIteratorV2) skippableDatum(meta1, meta2 *datum.Meta) bool {
	// If the hashes are equal and the second datum was processed, then skip it.
	return jdi.jc.hasher.Hash(meta1.Inputs) == jdi.jc.hasher.Hash(meta2.Inputs) && meta2.State == datum.State_PROCESSED
}

func (jdi *JobDatumIteratorV2) Stats() *datum.Stats {
	return jdi.stats
}

func (jdi *JobDatumIteratorV2) Finish() {
	close(jdi.done)
}
