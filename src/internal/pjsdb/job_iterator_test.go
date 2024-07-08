package pjsdb_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

func TestForEachJob(t *testing.T) {
	ctx, db := DB(t)
	s := FilesetStorage(t, db)
	pageSize := uint64(20)
	expected := make(map[pjsdb.JobID]bool)
	got := make(map[pjsdb.JobID]bool)
	withTx(t, ctx, db, s, func(d dependencies) {
		parent := pjsdb.JobID(0)
		var err error
		for i := uint64(0); i <= pageSize+pageSize/2; i++ {
			parent, err = createJob(t, d, parent)
			expected[parent] = true
			require.NoError(t, err)
		}
	})
	err := pjsdb.ForEachJob(ctx, db, pjsdb.IterateJobsRequest{
		IteratorConfiguration: pjsdb.IteratorConfiguration{PageSize: pageSize},
	}, func(job pjsdb.Job) error {
		got[job.ID] = true
		return nil
	})
	require.NoError(t, err)
	require.NoDiff(t, expected, got, nil)
}

func TestForEachJobTxByFilter(t *testing.T) {
	t.Run("valid/parent", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			expected := make(map[pjsdb.JobID]bool)
			left, err := createJob(t, d, 0)
			require.NoError(t, err)
			right, err := createJob(t, d, 0)
			require.NoError(t, err)
			withForEachJob(t, d, expected,
				pjsdb.IterateJobsRequest{Filter: pjsdb.IterateJobsFilter{Parent: left}},
				func(expected map[pjsdb.JobID]bool) {
					for i := 0; i < 25; i++ {
						leftChild, err := createJob(t, d, left)
						require.NoError(t, err)
						_, err = createJob(t, d, right)
						require.NoError(t, err)
						expected[leftChild] = true
					}
				})
		})
	})
	t.Run("valid/spec_hash", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			expected := make(map[pjsdb.JobID]bool)
			targetFs := mockFileset(t, d, "/spec", "#!/bin/bash; echo 'hello';")
			targetHash := hash(t, d, targetFs)
			withForEachJob(t, d, expected,
				pjsdb.IterateJobsRequest{Filter: pjsdb.IterateJobsFilter{SpecHash: targetHash}},
				func(expected map[pjsdb.JobID]bool) {
					for i := 0; i < 25; i++ {
						included := createJobWithFilesets(t, d, 0, targetFs, nil)
						_, err := createJob(t, d, 0)
						require.NoError(t, err)
						expected[included] = true
					}
				})
		})
	})
	t.Run("valid/spec", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			expected := make(map[pjsdb.JobID]bool)
			targetFs := mockFileset(t, d, "/spec", "#!/bin/bash; echo 'hello';")
			withForEachJob(t, d, expected,
				pjsdb.IterateJobsRequest{Filter: pjsdb.IterateJobsFilter{Spec: []byte(targetFs.HexString())}},
				func(expected map[pjsdb.JobID]bool) {
					for i := 0; i < 25; i++ {
						included := createJobWithFilesets(t, d, 0, targetFs, nil)
						_, err := createJob(t, d, 0)
						require.NoError(t, err)
						expected[included] = true
					}
				})
		})
	})
}

func TestIterateJobsFilterIsEmpty(t *testing.T) {
	filter := pjsdb.IterateJobsFilter{}
	require.True(t, filter.IsEmpty())
	filter.Input = []byte("") // empty, but non-nil slices should also be considered empty.
	require.True(t, filter.IsEmpty())
	filter.Input = []byte("dummy-input")
	require.False(t, filter.IsEmpty())
}

func withForEachJob(t *testing.T, d dependencies, expected map[pjsdb.JobID]bool, req pjsdb.IterateJobsRequest, f func(expected map[pjsdb.JobID]bool)) {
	got := make(map[pjsdb.JobID]bool)
	f(expected)
	err := pjsdb.ForEachJobTxByFilter(d.ctx, d.tx, req,
		func(job pjsdb.Job) error {
			got[job.ID] = true
			return nil
		})
	require.NoError(t, err)
	require.NoDiff(t, expected, got, nil)
}

func hash(t *testing.T, d dependencies, id *fileset.ID) []byte {
	fs, err := d.s.Open(d.ctx, []fileset.ID{*id})
	require.NoError(t, err)
	hash := pachhash.New()
	require.NoError(t, fs.Iterate(d.ctx, func(f fileset.File) error {
		fHash, err := f.Hash(d.ctx)
		require.NoError(t, err)
		_, err = hash.Write(fHash)
		require.NoError(t, err)
		return nil
	}))
	return hash.Sum(nil)
}
