package chain

import (
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
)

type DatumHasherV2 interface {
	Hash([]*common.InputV2) string
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
		stats:     &datum.Stats{},
		dit:       dit,
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

func (jdi *JobDatumIteratorV2) WithDeleter(deleter func(*datum.Meta) error, cb func() error) error {
	jdi.deleter = deleter
	defer func() {
		jdi.deleter = nil
	}()
	return cb()
}

func (jdi *JobDatumIteratorV2) Iterate(cb func(*datum.Meta) error) error {
	if jdi.parent == nil {
		return jdi.dit.Iterate(cb)
	}
	// Generate datum sets for the new datums (datums that do not exist in the parent job).
	// TODO: Logging?
	if err := datum.Merge([]datum.IteratorV2{jdi.dit, jdi.parent.dit}, func(metas []*datum.Meta) error {
		return jdi.maybeSkip(metas, cb)
	}); err != nil {
		return err
	}
	<-jdi.parent.done
	// Generate datum sets for the skipped datums that were not processed by the parent (failed, recovered, etc.).
	return datum.Merge([]datum.IteratorV2{jdi.parent.dit, jdi.parent.outputDit}, func(metas []*datum.Meta) error {
		jdi.stats.Skipped++
		return jdi.maybeSkip(metas, func(meta *datum.Meta) error {
			jdi.stats.Skipped--
			return cb(meta)
		})
	})
}

func (jdi *JobDatumIteratorV2) maybeSkip(metas []*datum.Meta, cb func(*datum.Meta) error) error {
	// Handle the case when the datum appears in both streams.
	if len(metas) > 1 {
		// If the hashes are not equal, then delete the datum and process it.
		if jdi.jc.hasher.Hash(metas[0].Inputs) != jdi.jc.hasher.Hash(metas[1].Inputs) {
			if jdi.deleter != nil {
				if err := jdi.deleter(metas[1]); err != nil {
					return err
				}
			}
			return cb(metas[0])
		}
		// If the hashes are equal, but the datum was not processed, then process it.
		if metas[1].State != datum.State_PROCESSED {
			return cb(metas[0])
		}
		// If the hashes are equal and the datum was processed, then skip it.
		return nil
	}
	// Handle the case when the datum only exists in the parent.
	// TODO: Use job id not set as sentinel for current job datum iterator?
	if metas[0].JobID != "" {
		if jdi.deleter != nil {
			return jdi.deleter(metas[0])
		}
		return nil
	}
	return cb(metas[0])
}

func (jdi *JobDatumIteratorV2) Stats() *datum.Stats {
	return jdi.stats
}

func (jdi *JobDatumIteratorV2) Finish() {
	close(jdi.done)
}
