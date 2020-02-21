package chain

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
)

// Interface - put job into black box
// only for jobs in the running state
// black box returns datum.Iterator of datums to be processed as they are safe to be processed
// Notify black box when a job succeeds or fails so it can propagate datums to downstream jobs

type DatumHasher interface {
	Hash([]*common.Input) string
}

type JobData interface {
	Iterator() (datum.Iterator, error)
}

type JobDatumIterator interface {
	NextBatch(context.Context) (uint64, error)
	NextDatum() []*common.Input
	AdditiveOnly() bool
	DatumSet() DatumSet
	Reset()
}

type JobChain interface {
	Initialized() bool
	Initialize(baseDatums DatumSet) error

	Start(jd JobData) (JobDatumIterator, error)
	Succeed(jd JobData, recoveredDatums DatumSet) error
	Fail(jd JobData) error
}

type DatumSet map[string]struct{}

type jobDatumIterator struct {
	data JobData
	jc   *jobChain

	// TODO: lower memory consumption - all these datumsets might result in a
	// really large memory footprint. See if we can do a streaming interface to
	// replace these - will likely require the new storage layer, as additive-only
	// jobs need this stuff the most.
	yielding  DatumSet // Datums that may be yielded as the iterator progresses
	yielded   DatumSet // Datums that have been yielded
	allDatums DatumSet // All datum hashes from the datum iterator

	ancestors []*jobDatumIterator
	dit       datum.Iterator

	finished     bool
	additiveOnly bool
	done         chan struct{}
}

type jobChain struct {
	mutex      sync.Mutex
	hasher     DatumHasher
	jobs       []*jobDatumIterator
	baseDatums DatumSet
}

func NewJobChain(hasher DatumHasher) JobChain {
	return &jobChain{
		hasher:     hasher,
		jobs:       []*jobDatumIterator{},
		baseDatums: nil,
	}
}

func (jc *jobChain) Initialized() bool {
	return len(jc.jobs) > 0
}

func (jc *jobChain) Initialize(baseDatums DatumSet) error {
	if jc.Initialized() {
		return fmt.Errorf("cannot reinitialize JobChain")
	}

	// Insert a dummy job representing the given base datum set
	jdi := &jobDatumIterator{
		data:      nil,
		jc:        jc,
		allDatums: baseDatums,
		finished:  true,
		done:      make(chan struct{}),
	}

	close(jdi.done)
	jc.jobs = append(jc.jobs, jdi)

	return nil
}

func (jdi *jobDatumIterator) recalculate(baseDatums DatumSet, allAncestors []*jobDatumIterator) {
	jdi.ancestors = []*jobDatumIterator{}
	interestingAncestors := map[*jobDatumIterator]struct{}{}
	for hash := range jdi.allDatums {
		if _, ok := jdi.yielded[hash]; ok {
			continue
		}

		safeToProcess := true
		// interestingAncestors should be _all_ unfinished previous jobs which have
		// _any_ datum overlap with this job
		for _, ancestor := range allAncestors {
			if !ancestor.finished {
				if _, ok := ancestor.allDatums[hash]; ok {
					interestingAncestors[ancestor] = struct{}{}
					safeToProcess = false
				}
			}
		}

		if safeToProcess {
			jdi.yielding[hash] = struct{}{}
		}
	}

	var parentJob *jobDatumIterator
	for i := len(allAncestors) - 1; i >= 0; i-- {
		// Skip all failed jobs
		if allAncestors[i].allDatums != nil {
			parentJob = allAncestors[i]
			break
		}
	}

	// If this job is additive-only from the parent job, we should mark it now -
	// loop over parent datums to see if they are all present
	jdi.additiveOnly = true
	for hash := range parentJob.allDatums {
		if _, ok := jdi.allDatums[hash]; !ok {
			jdi.additiveOnly = false
			break
		}
	}

	if jdi.additiveOnly {
		// If this is additive-only, we only need to enqueue new datums (since the parent job)
		for hash := range jdi.yielding {
			if _, ok := parentJob.allDatums[hash]; ok {
				delete(jdi.yielding, hash)
			}
		}
		// An additive-only job can only progress once its parent job has finished.
		// At that point it will re-evaluate what datums to process in case of a
		// failed job or recovered datums.
		if !parentJob.finished {
			jdi.ancestors = append(jdi.ancestors, parentJob)
		}
	} else {
		for ancestor := range interestingAncestors {
			jdi.ancestors = append(jdi.ancestors, ancestor)
		}
	}
}

func (jc *jobChain) Start(jd JobData) (JobDatumIterator, error) {
	if !jc.Initialized() {
		return nil, fmt.Errorf("JobChain is not initialized")
	}

	dit, err := jd.Iterator()
	if err != nil {
		return nil, err
	}

	jdi := &jobDatumIterator{
		data:      jd,
		jc:        jc,
		yielding:  make(DatumSet),
		yielded:   make(DatumSet),
		allDatums: make(DatumSet),
		ancestors: []*jobDatumIterator{},
		dit:       dit,
		done:      make(chan struct{}),
	}

	jdi.dit.Reset()
	for i := 0; i < jdi.dit.Len(); i++ {
		inputs := jdi.dit.DatumN(i)
		hash := jc.hasher.Hash(inputs)
		jdi.allDatums[hash] = struct{}{}
	}

	jc.mutex.Lock()
	defer jc.mutex.Unlock()

	jdi.recalculate(jc.baseDatums, jc.jobs)

	jc.jobs = append(jc.jobs, jdi)
	return jdi, nil
}

func (jc *jobChain) indexOf(jd JobData) (int, error) {
	for i, x := range jc.jobs {
		if x.data == jd {
			return i, nil
		}
	}
	panic("job not found in job chain")
	return 0, fmt.Errorf("job not found in job chain")
}

func (jc *jobChain) cleanFinishedJobs() {
	for len(jc.jobs) > 1 && jc.jobs[1].finished {
		if jc.jobs[1].allDatums != nil {
			jc.jobs[0].allDatums = jc.jobs[1].allDatums
		}
		jc.jobs = append(jc.jobs[:1], jc.jobs[2:]...)
	}
}

func (jc *jobChain) Fail(jd JobData) error {
	jc.mutex.Lock()
	defer jc.mutex.Unlock()

	index, err := jc.indexOf(jd)
	if err != nil {
		return err
	}

	jdi := jc.jobs[index]
	jdi.allDatums = nil
	jdi.finished = true

	close(jdi.done)

	jc.cleanFinishedJobs()

	return nil
}

func (jc *jobChain) Succeed(jd JobData, recoveredDatums DatumSet) error {
	jc.mutex.Lock()
	defer jc.mutex.Unlock()

	index, err := jc.indexOf(jd)
	if err != nil {
		return err
	}

	jdi := jc.jobs[index]

	if len(jdi.yielding) != 0 || len(jdi.ancestors) > 0 {
		return fmt.Errorf(
			"cannot succeed a job with items remaining on the iterator: %d datums and %d ancestor jobs",
			len(jdi.yielding), len(jdi.ancestors),
		)
	}

	for hash := range recoveredDatums {
		delete(jdi.allDatums, hash)
	}

	jdi.finished = true

	if index == 0 {
		jc.jobs = jc.jobs[1:]
		jc.baseDatums = jdi.allDatums
	}

	close(jdi.done)

	jc.cleanFinishedJobs()

	return nil
}

func safeToProcess(hash string, ancestors []*jobDatumIterator) bool {
	for _, ancestor := range ancestors {
		if _, ok := ancestor.allDatums[hash]; ok {
			return false
		}
	}
	return true
}

// TODO: iteration should return a chunk of 'known' new datums before other
// datums (to optimize for distributing processing across workers). This should
// still be true even after resetting the iterator.
func (jdi *jobDatumIterator) NextBatch(ctx context.Context) (uint64, error) {
	for len(jdi.yielding) == 0 {
		if len(jdi.ancestors) == 0 {
			return 0, nil
		}

		// Wait on an ancestor job
		cases := make([]reflect.SelectCase, 0, len(jdi.ancestors)+1)
		for _, x := range jdi.ancestors {
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(x.done)})
		}
		cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())})

		// Wait for an ancestor job to finish, then remove it from our dependencies
		selectIndex, _, _ := reflect.Select(cases)
		if selectIndex == len(cases)-1 {
			return 0, ctx.Err()
		}

		if err := func() error {
			jdi.jc.mutex.Lock()
			defer jdi.jc.mutex.Unlock()

			if jdi.finished {
				return fmt.Errorf("stopping datum iteration because job failed")
			}

			index, err := jdi.jc.indexOf(jdi.data)
			if err != nil {
				return err
			}

			jdi.recalculate(jdi.jc.baseDatums, jdi.jc.jobs[:index])
			return nil
		}(); err != nil {
			return 0, err
		}

		jdi.dit.Reset()
	}

	return uint64(len(jdi.yielding)), nil
}

func (jdi *jobDatumIterator) NextDatum() []*common.Input {
	for jdi.dit.Next() {
		inputs := jdi.dit.Datum()
		hash := jdi.jc.hasher.Hash(inputs)
		if _, ok := jdi.yielding[hash]; ok {
			delete(jdi.yielding, hash)
			jdi.yielded[hash] = struct{}{}
			return inputs
		}
	}

	return nil
}

func (jdi *jobDatumIterator) Reset() {
	jdi.dit.Reset()
	for hash := range jdi.yielded {
		delete(jdi.yielded, hash)
		jdi.yielding[hash] = struct{}{}
	}
}

func (jdi *jobDatumIterator) Datum() []*common.Input {
	return jdi.dit.Datum()
}

func (jdi *jobDatumIterator) DatumSet() DatumSet {
	return jdi.allDatums
}

func (jdi *jobDatumIterator) AdditiveOnly() bool {
	return jdi.additiveOnly
}
