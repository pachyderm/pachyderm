package pjsdb_test

import (
	"math"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

func createRootJob(t *testing.T, d dependencies) pjsdb.JobID {
	fs, hash := mockAndHashFileset(t, d, "/", "")
	req := pjsdb.CreateJobRequest{
		Parent:      0,
		Inputs:      nil,
		Program:     fs,
		ProgramHash: hash,
	}
	id, err := pjsdb.CreateJob(d.ctx, d.tx, req)
	require.NoError(t, err)
	return id
}

func createJob(t *testing.T, d dependencies, parent pjsdb.JobID) (pjsdb.JobID, error) {
	req := makeReq(t, d, parent, nil)
	id, err := pjsdb.CreateJob(d.ctx, d.tx, req)
	return id, err
}

func TestCreateAndGetJob(t *testing.T) {
	t.Run("valid/parent/nil", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, 0)
			require.NoError(t, err)
			_, err = pjsdb.GetJob(d.ctx, d.tx, id)
			require.NoError(t, err)
		})
	})
	t.Run("valid/parent/exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, createRootJob(t, d))
			require.NoError(t, err)
			_, err = pjsdb.GetJob(d.ctx, d.tx, id)
			require.NoError(t, err)
		})
	})
	t.Run("valid/cache/exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, createRootJob(t, d))
			require.NoError(t, err)
			hasher := pachhash.New()
			_, err = hasher.Write([]byte("hello, world!"))
			require.NoError(t, err)
			hash := hasher.Sum(nil)
			_, err = d.tx.ExecContext(d.ctx, `
				UPDATE pjs.job_cache SET job_hash = $1, cache_read=true, cache_write=true
				WHERE job_id = $2;`,
				hash, id)
			require.NoError(t, err)
			j, err := pjsdb.GetJob(d.ctx, d.tx, id)
			require.NoError(t, err)
			require.NoDiff(t, hash, j.JobHash, nil)
		})
	})
}

func TestCancelJob(t *testing.T) {
	t.Run("valid/cancel/single", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, createRootJob(t, d))
			require.NoError(t, err)
			canceledJobs, err := pjsdb.CancelJob(d.ctx, d.tx, id)
			require.NoError(t, err)
			require.Equal(t, 1, len(canceledJobs))
		})
	})
	t.Run("valid/cancel/all", func(t *testing.T) {
		maxDepth := 4
		numJobs := int(math.Pow(float64(2), float64(maxDepth)) - 1)
		numJobs += 1 // for the root job.
		withDependencies(t, func(d dependencies) {
			fullBinaryJobTree(t, d, maxDepth)
			// cancel all the jobs, including root.
			canceledJobs, err := pjsdb.CancelJob(d.ctx, d.tx, 1)
			require.NoError(t, err)
			require.Equal(t, numJobs, len(canceledJobs))
		})
	})
	t.Run("valid/cancel/subset", func(t *testing.T) {
		maxDepth := 3
		/* The tree looks like this at a depth of 3:
		2
		├── 3
		│   ├── 5
		│   └── 6
		└── 4
		    ├── 7
		    └── 8
		*/
		withDependencies(t, func(d dependencies) {
			fullBinaryJobTree(t, d, maxDepth)
			// cancel 3 and children of 3 (5, and 6).
			canceledJobs, err := pjsdb.CancelJob(d.ctx, d.tx, 3)
			require.NoError(t, err)
			require.Equal(t, 3, len(canceledJobs))
			require.ElementsEqual(t, []pjsdb.JobID{3, 5, 6}, canceledJobs)
		})
	})
	t.Run("invalid/cancel/not_exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			_, err := pjsdb.CancelJob(d.ctx, d.tx, 4)
			require.YesError(t, err)
			if !errors.As(err, &pjsdb.JobNotFoundError{}) {
				t.Fatalf("expected to get job not found error, got: %s", err)
			}
		})
	})
}

func jobIDs(jobs []pjsdb.Job) (ids []pjsdb.JobID) {
	for _, job := range jobs {
		ids = append(ids, job.ID)
	}
	return ids
}

func TestWalkJob(t *testing.T) {
	t.Run("valid/walk/all", func(t *testing.T) {
		maxDepth := 3
		numJobs := int(math.Pow(float64(2), float64(maxDepth)) - 1)
		numJobs += 1 // for the root job.
		withDependencies(t, func(d dependencies) {
			fullBinaryJobTree(t, d, maxDepth)
			// walk all jobs.
			t.Run("level_order", func(t *testing.T) {
				jobs, err := pjsdb.WalkJob(d.ctx, d.tx, 1, pjsdb.LevelOrder, 10_000)
				require.NoError(t, err)
				require.Equal(t, numJobs, len(jobs))
				ids := jobIDs(jobs)
				require.NoDiff(t, []pjsdb.JobID{1, 2, 3, 4, 5, 6, 7, 8}, ids, nil)
			})
			t.Run("pre_order", func(t *testing.T) {
				jobs, err := pjsdb.WalkJob(d.ctx, d.tx, 1, pjsdb.PreOrder, 10_000)
				require.NoError(t, err)
				require.Equal(t, numJobs, len(jobs))
				ids := jobIDs(jobs)
				require.NoDiff(t, []pjsdb.JobID{1, 2, 3, 5, 6, 4, 7, 8}, ids, nil)
			})
			t.Run("post_order", func(t *testing.T) {
				jobs, err := pjsdb.WalkJob(d.ctx, d.tx, 1, pjsdb.MirroredPostOrder, 10_000)
				require.NoError(t, err)
				require.Equal(t, numJobs, len(jobs))
				ids := jobIDs(jobs)
				require.NoDiff(t, []pjsdb.JobID{8, 7, 4, 6, 5, 3, 2, 1}, ids, nil)
			})
		})
	})
	t.Run("valid/walk/subset", func(t *testing.T) {
		maxDepth := 3
		/* The tree looks like this at a depth of 3:
		2
		├── 3
		│   ├── 5
		│   └── 6
		└── 4
		    ├── 7
		    └── 8
		*/
		withDependencies(t, func(d dependencies) {
			fullBinaryJobTree(t, d, maxDepth)
			t.Run("level_order", func(t *testing.T) {
				jobs, err := pjsdb.WalkJob(d.ctx, d.tx, 3, pjsdb.LevelOrder, 10_000)
				require.NoError(t, err)
				require.Equal(t, 3, len(jobs))
				ids := jobIDs(jobs)
				require.NoDiff(t, []pjsdb.JobID{3, 5, 6}, ids, nil)
			})
			t.Run("pre_order", func(t *testing.T) {
				jobs, err := pjsdb.WalkJob(d.ctx, d.tx, 3, pjsdb.PreOrder, 10_000)
				require.NoError(t, err)
				require.Equal(t, 3, len(jobs))
				ids := jobIDs(jobs)
				require.NoDiff(t, []pjsdb.JobID{3, 5, 6}, ids, nil)
			})
			t.Run("mirrored_post_order", func(t *testing.T) {
				jobs, err := pjsdb.WalkJob(d.ctx, d.tx, 3, pjsdb.MirroredPostOrder, 10_000)
				require.NoError(t, err)
				require.Equal(t, 3, len(jobs))
				ids := jobIDs(jobs)
				require.NoDiff(t, []pjsdb.JobID{6, 5, 3}, ids, nil)
			})
			t.Run("level_order/depth=1", func(t *testing.T) {
				jobs, err := pjsdb.WalkJob(d.ctx, d.tx, 2, pjsdb.LevelOrder, 1)
				require.NoError(t, err)
				require.Equal(t, 3, len(jobs))
				ids := jobIDs(jobs)
				require.NoDiff(t, []pjsdb.JobID{2, 3, 4}, ids, nil)
			})
		})
	})
}

func TestDeleteJob(t *testing.T) {
	t.Run("valid/delete/single/queued", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, createRootJob(t, d))
			require.NoError(t, err)
			deletedIds, err := pjsdb.DeleteJob(d.ctx, d.tx, id)
			require.NoError(t, err)
			require.Equal(t, deletedIds[0], id)
		})
	})
	t.Run("valid/delete/single", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, createRootJob(t, d))
			require.NoError(t, err)
			canceledIds, err := pjsdb.CancelJob(d.ctx, d.tx, id)
			require.NoError(t, err)
			require.Equal(t, canceledIds[0], id)
			deletedIds, err := pjsdb.DeleteJob(d.ctx, d.tx, id)
			require.NoError(t, err)
			require.Equal(t, deletedIds[0], id)
		})
	})
	t.Run("valid/all", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			root := createRootJob(t, d)
			id, err := createJob(t, d, root)
			require.NoError(t, err)
			id2, err := createJob(t, d, id)
			require.NoError(t, err)
			expected := []pjsdb.JobID{root, id, id2}
			canceledIds, err := pjsdb.CancelJob(d.ctx, d.tx, root)
			require.NoError(t, err)
			require.ElementsEqual(t, expected, canceledIds)
			deletedIds, err := pjsdb.DeleteJob(d.ctx, d.tx, root)
			require.NoError(t, err)
			require.ElementsEqual(t, expected, deletedIds)
		})
	})
	t.Run("valid/delete/subset", func(t *testing.T) {
		maxDepth := 3
		/* The tree looks like this at a depth of 3:
		2
		├── 3
		│   ├── 5
		│   └── 6
		└── 4
		    ├── 7
		    └── 8
		*/
		withDependencies(t, func(d dependencies) {
			fullBinaryJobTree(t, d, maxDepth)
			canceledJobs, err := pjsdb.CancelJob(d.ctx, d.tx, 3)
			require.NoError(t, err)
			require.Equal(t, 3, len(canceledJobs))
			require.ElementsEqual(t, []pjsdb.JobID{3, 5, 6}, canceledJobs)
			deletedJobs, err := pjsdb.DeleteJob(d.ctx, d.tx, 3)
			require.NoError(t, err)
			require.Equal(t, 3, len(deletedJobs))
			require.ElementsEqual(t, canceledJobs, deletedJobs)
		})
	})
	t.Run("invalid/delete/single", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, createRootJob(t, d))
			require.NoError(t, err)
			_, err = d.tx.ExecContext(d.ctx, `UPDATE pjs.jobs SET processing = CURRENT_TIMESTAMP where id = $1`, id)
			require.NoError(t, err)
			deletedIds, err := pjsdb.DeleteJob(d.ctx, d.tx, id)
			require.NoError(t, err)
			require.Len(t, deletedIds, 0)
		})
	})
	t.Run("invalid/delete/done_parent", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			fullBinaryJobTree(t, d, 3)
			_, err := d.tx.ExecContext(d.ctx, `UPDATE pjs.jobs SET done = CURRENT_TIMESTAMP WHERE id = 3;`)
			require.NoError(t, err)
			_, err = pjsdb.DeleteJob(d.ctx, d.tx, 3)
			require.YesError(t, err)
			require.True(t, strings.Contains(err.Error(), "is done before child"))
		})
	})
}

func TestListJobTxByFilter(t *testing.T) {
	t.Run("valid/program", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			expected := make([]pjsdb.Job, 0)
			var err error
			targetFs, targetHash := mockAndHashFileset(t, d, "/program", "#!/bin/bash; echo 'hello';")
			for i := 0; i < 5; i++ {
				included, err := pjsdb.GetJob(d.ctx, d.tx, createJobWithFilesets(t, d, 0, targetFs, targetHash))
				require.NoError(t, err)
				_, err = createJob(t, d, 0)
				require.NoError(t, err)
				expected = append(expected, included)
			}
			jobs, err := pjsdb.ListJobTxByFilter(d.ctx, d.tx,
				pjsdb.IterateJobsRequest{Filter: pjsdb.IterateJobsFilter{Program: []byte(fileset.Token(targetFs).HexString())}})
			require.NoError(t, err)
			require.NoDiff(t, expected, jobs, nil)
		})
	})
}

func fullBinaryJobTree(t *testing.T, d dependencies, maxDepth int) {
	if maxDepth == 0 {
		return
	}
	edges := make(map[pjsdb.JobID][]pjsdb.JobID)
	parents := make([]pjsdb.JobID, 0)
	// create node at depth == 1
	parent, err := createJob(t, d, createRootJob(t, d))
	require.NoError(t, err)
	parents = append(parents, parent)
	// create nodes at depth > 1
	for depth := 2; depth <= maxDepth; depth++ {
		newParents := make([]pjsdb.JobID, 0)
		for _, p := range parents {
			leftChild, err := createJob(t, d, p)
			require.NoError(t, err)
			rightChild, err := createJob(t, d, p)
			require.NoError(t, err)
			children := []pjsdb.JobID{leftChild, rightChild}
			edges[p] = children
			newParents = append(newParents, children...)
		}
		parents = newParents
	}
}
