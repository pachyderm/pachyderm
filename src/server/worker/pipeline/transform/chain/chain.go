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
	Next(context.Context) (bool, error)
	Datum() []*common.Input
	NumAvailable() int
	AdditiveOnly() bool
	DatumSet() DatumSet
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
	unyielded       DatumSet // Datums that are waiting on an ancestor job
	yielding        DatumSet // Datums that may be yielded as the iterator progresses
	yielded         DatumSet // Datums that have been yielded
	allDatums       DatumSet // All datum hashes from the datum iterator
	recoveredDatums DatumSet // Recovered datums from a completed job

	ancestors []*jobDatumIterator
	dit       datum.Iterator

	finished     bool
	additiveOnly bool
	// TODO: have a 'doneProcessing' (for additive-subtractive descendents) and 'doneMerging' (for additive-only decendents)
	done chan struct{}
}

type jobChain struct {
	mutex      sync.Mutex
	hasher     DatumHasher
	jobs       []*jobDatumIterator
	baseDatums DatumSet
}

func NewJobChain(hasher DatumHasher) (JobChain, error) {
	return &jobChain{
		hasher:     hasher,
		jobs:       []*jobDatumIterator{},
		baseDatums: nil,
	}, nil
}

func (jc *jobChain) Initialized() bool {
	return jc.baseDatums != nil
}

func (jc *jobChain) Initialize(baseDatums DatumSet) error {
	if jc.Initialized() {
		return fmt.Errorf("cannot reinitialize JobChain")
	}
	jc.baseDatums = baseDatums
	return nil
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
		unyielded: make(DatumSet),
		yielding:  make(DatumSet),
		yielded:   make(DatumSet),
		allDatums: make(DatumSet),
		ancestors: []*jobDatumIterator{},
		dit:       dit,
		done:      make(chan struct{}),
	}

	jc.mutex.Lock()
	defer jc.mutex.Unlock()

	dit.Reset()
	ancestors := map[*jobDatumIterator]struct{}{}
	for i := 0; i < dit.Len(); i++ {
		inputs := jdi.dit.DatumN(i)
		hash := jc.hasher.Hash(inputs)
		jdi.allDatums[hash] = struct{}{}

		safeToProcess := true
		// ancestors should be _all_ previous jobs which have _any_ datum overlap with this job
		for _, ancestor := range jc.jobs {
			if !ancestor.finished {
				if _, ok := ancestor.allDatums[hash]; ok {
					ancestors[ancestor] = struct{}{}
					safeToProcess = false
				}
			}
		}

		if safeToProcess {
			jdi.yielding[hash] = struct{}{}
		} else {
			jdi.unyielded[hash] = struct{}{}
		}
	}

	// If this job is additive-only from the parent job, we should mark it now - loop over parent datums to see if they are all present
	parentDatums := jc.baseDatums
	if len(jc.jobs) > 0 {
		parentDatums = jc.jobs[len(jc.jobs)-1].allDatums
	}
	jdi.additiveOnly = true
	for hash := range parentDatums {
		if _, ok := jdi.allDatums[hash]; !ok {
			jdi.additiveOnly = false
			break
		}
	}

	if jdi.additiveOnly {
		// If this is additive-only, we only need to enqueue new datums (since the parent job)
		for hash := range jdi.allDatums {
			if _, ok := parentDatums[hash]; ok {
				delete(jdi.yielding, hash)
				delete(jdi.unyielded, hash)
			}
		}
	}

	for ancestor := range ancestors {
		jdi.ancestors = append(jdi.ancestors, ancestor)
	}

	fmt.Printf("Started job with %d dependencies\n", len(jdi.ancestors))

	jc.jobs = append(jc.jobs, jdi)
	return jdi, nil
}

func (jc *jobChain) indexOf(jd JobData) (int, error) {
	for i, x := range jc.jobs {
		if x.data == jd {
			return i, nil
		}
	}
	return 0, fmt.Errorf("job not found in job chain")
}

func (jc *jobChain) cleanFinishedJobs() {
	var newBaseDatums DatumSet
	index := -1
	for i, job := range jc.jobs {
		if !job.finished {
			break
		}
		index = i
		if job.allDatums != nil {
			newBaseDatums = job.allDatums
		}
	}

	jc.jobs = jc.jobs[index+1:]
	jc.baseDatums = newBaseDatums
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

	jc.cleanFinishedJobs()

	close(jdi.done)

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

	if len(jdi.yielding) != 0 || len(jdi.unyielded) != 0 {
		return fmt.Errorf(
			"cannot succeed a job with remaining datums: %d + %d of %d",
			len(jdi.unyielded), len(jdi.yielding), len(jdi.unyielded)+len(jdi.yielding)+len(jdi.yielded),
		)
	}

	for hash := range recoveredDatums {
		delete(jdi.allDatums, hash)
	}

	jdi.recoveredDatums = recoveredDatums
	jdi.finished = true

	if index == 0 {
		jc.jobs = jc.jobs[1:]
		jc.baseDatums = jdi.allDatums
	}

	jc.cleanFinishedJobs()

	close(jdi.done)

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

func (jdi *jobDatumIterator) Next(ctx context.Context) (bool, error) {
	for {
		for len(jdi.yielding) == 0 {
			if len(jdi.unyielded) == 0 || len(jdi.ancestors) == 0 {
				return false, nil
			}

			// Wait on an ancestor job
			cases := make([]reflect.SelectCase, 0, len(jdi.ancestors)+1)
			for _, x := range jdi.ancestors {
				cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(x.done)})
			}
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())})

			// Wait for an ancestor job to finish, then remove it from our dependencies
			index, _, _ := reflect.Select(cases)
			if index == len(cases)-1 {
				return false, ctx.Err()
			}
			// ancestor = jdi.ancestors[index]
			jdi.ancestors = append(jdi.ancestors[:index], jdi.ancestors[index+1:]...)

			// TODO: special handling if jdi is additive-only

			// TODO: update 'yielding' from 'unyielded'
			for hash := range jdi.unyielded {
				if safeToProcess(hash, jdi.ancestors) {
					delete(jdi.unyielded, hash)
					jdi.yielding[hash] = struct{}{}
				}
			}

			jdi.dit.Reset()
		}

		for jdi.dit.Next() {
			inputs := jdi.dit.Datum()
			hash := jdi.jc.hasher.Hash(inputs)
			if _, ok := jdi.yielding[hash]; ok {
				delete(jdi.yielding, hash)
				jdi.yielded[hash] = struct{}{}
				return true, nil
			}
		}
		// TODO: assert that len(jdi.yielding) == 0
	}
}

func (jdi *jobDatumIterator) NumAvailable() int {
	return len(jdi.yielding)
}

func (jdi *jobDatumIterator) Datum() []*common.Input {
	return jdi.dit.Datum()
}

func (jdi *jobDatumIterator) DatumSet() DatumSet {
	return jdi.allDatums
}

func (jdi *jobDatumIterator) AdditiveOnly() bool {
	return false // TODO: implement
}
