package chain

import (
	"context"

	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
)

type noSkipJobDatumIterator struct {
	dit       datum.Iterator
	ditIndex  int
	allDatums DatumSet
	done      bool
}

type noSkipJobChain struct {
	hasher DatumHasher
}

// NewNoSkipJobChain constructs a JobChain that will always yield all datums in each
// job's datum iterator, and will therefore not block.
func NewNoSkipJobChain(hasher DatumHasher) JobChain {
	return &noSkipJobChain{hasher: hasher}
}

// Start adds a new job to the chain and returns the corresponding
// JobDatumIterator
func (jc *noSkipJobChain) Start(jd JobData) (JobDatumIterator, error) {
	dit, err := jd.Iterator()
	if err != nil {
		return nil, err
	}

	allDatums := make(DatumSet)

	dit.Reset()
	for dit.Next() {
		allDatums[jc.hasher.Hash(dit.Datum())]++
	}
	dit.Reset()

	return &noSkipJobDatumIterator{
		allDatums: allDatums,
		dit:       dit,
		ditIndex:  -1,
		done:      false,
	}, nil
}

// RecoveredDatums indicates the set of recovered datums for the job. This can
// be called multiple times.
func (jc *noSkipJobChain) RecoveredDatums(jd JobData, recoveredDatums DatumSet) error {
	return nil
}

// Succeed indicates that the job has finished successfully
func (jc *noSkipJobChain) Succeed(jd JobData) error {
	return nil
}

// Fail indicates that the job has finished unsuccessfully
func (jc *noSkipJobChain) Fail(jd JobData) error {
	return nil
}

func (jdi *noSkipJobDatumIterator) NextBatch(ctx context.Context) (uint64, error) {
	if !jdi.done {
		jdi.done = true
		return jdi.MaxLen(), nil
	}
	return 0, nil
}

func (jdi *noSkipJobDatumIterator) NextDatum() ([]*common.Input, int64) {
	jdi.ditIndex++
	if jdi.ditIndex < jdi.dit.Len() {
		return jdi.dit.Datum(), int64(jdi.ditIndex)
	}
	return nil, 0
}

func (jdi *noSkipJobDatumIterator) Reset() {
	jdi.dit.Reset()
	jdi.done = false
}

func (jdi *noSkipJobDatumIterator) MaxLen() uint64 {
	return uint64(jdi.dit.Len())
}

func (jdi *noSkipJobDatumIterator) DatumSet() DatumSet {
	return jdi.allDatums
}

func (jdi *noSkipJobDatumIterator) AdditiveOnly() bool {
	return false
}
