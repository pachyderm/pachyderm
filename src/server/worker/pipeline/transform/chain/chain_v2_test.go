package chain

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
)

type testHasherV2 struct{}

func (th *testHasherV2) Hash(inputs []*common.Input) string {
	return common.HashDatumV2("", "", inputs)
}

type testIteratorV2 struct {
	metas []*datum.Meta
}

func newTestIteratorV2(metas []*datum.Meta) *testIteratorV2 {
	return &testIteratorV2{metas: metas}
}

func (ti *testIteratorV2) Iterate(cb func(*datum.Meta) error) error {
	for _, meta := range ti.metas {
		if err := cb(meta); err != nil {
			return err
		}
	}
	return nil
}

func newTestChainV2(metas ...*datum.Meta) *JobChainV2 {
	hasher := &testHasherV2{}
	if len(metas) > 0 {
		baseDit := newTestIteratorV2(metas)
		return NewJobChainV2(hasher, baseDit)
	}
	return NewJobChainV2(hasher)
}

func newMeta(jobID, name, hash string) *datum.Meta {
	return &datum.Meta{
		JobID: jobID,
		Inputs: []*common.Input{
			&common.Input{
				FileInfo: &pfs.FileInfo{
					File: &pfs.File{
						Path: name,
					},
				},
			},
		},
		Hash: hash,
	}
}

func newTestMetas(jobID string) []*datum.Meta {
	return []*datum.Meta{
		newMeta(jobID, "a", "ahash"),
		newMeta(jobID, "b", "bhash"),
		newMeta(jobID, "c", "chash"),
	}
}

func requireIteratorContentsV2(t *testing.T, jdi *JobDatumIteratorV2, metas []*datum.Meta) {
	require.NoError(t, jdi.Iterate(func(meta *datum.Meta) error {
		require.Equal(t, metas[0], meta)
		metas = metas[1:]
		return nil
	}))
}

func TestEmptyBaseV2(t *testing.T) {
	chain := newTestChainV2()
	jobID := uuid.NewWithoutDashes()
	jobMetas := newTestMetas(jobID)
	ti := newTestIteratorV2(jobMetas)
	jdi := chain.CreateJob(jobID, ti, ti)
	requireIteratorContentsV2(t, jdi, jobMetas)
}

func TestAdditiveOnBaseV2(t *testing.T) {
	chain := newTestChainV2(newTestMetas(uuid.NewWithoutDashes())...)
	jobID := uuid.NewWithoutDashes()
	jobMetas := newTestMetas(jobID)[1:]
	ti := newTestIteratorV2(jobMetas)
	jdi := chain.CreateJob(jobID, ti, ti)
	requireIteratorContentsV2(t, jdi, jobMetas)
}

func TestSubtractiveOnBaseV2(t *testing.T) {
	chain := newTestChainV2(newTestMetas(uuid.NewWithoutDashes())...)
	jobID := uuid.NewWithoutDashes()
	jobMetas := newTestMetas(jobID)[1:]
	ti := newTestIteratorV2(jobMetas)
	jdi := chain.CreateJob(jobID, ti, ti)
	requireIteratorContentsV2(t, jdi, jobMetas[1:])
}

func TestAdditiveSubtractiveOnBaseV2(t *testing.T) {
	chain := newTestChainV2(newTestMetas(uuid.NewWithoutDashes())[1:]...)
	jobID := uuid.NewWithoutDashes()
	jobMetas := newTestMetas(jobID)[:2]
	ti := newTestIteratorV2(jobMetas)
	jdi := chain.CreateJob(jobID, ti, ti)
	requireIteratorContentsV2(t, jdi, jobMetas[0:1])
}
