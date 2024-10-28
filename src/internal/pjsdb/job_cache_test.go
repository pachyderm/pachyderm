package pjsdb_test

import (
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/pjs"

	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

func requireCacheCount(t *testing.T, expected int, d dependencies) {
	numRows := 0
	require.NoError(t, sqlx.GetContext(d.ctx, d.tx, &numRows, `SELECT count(*) FROM pjs.job_cache`))
	require.Equal(t, expected, numRows)
}

func TestCaching(t *testing.T) {
	t.Run("valid/cache_read_and_write", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			// first job should miss cache read, but successfully write.
			createRequest := makeReq(t, d, 0, nil)
			id, err := pjsdb.CreateJob(d.ctx, d.tx, createRequest)
			require.NoError(t, err)
			err = pjsdb.CompleteJob(d.ctx, d.tx, id, []fileset.Pin{}, pjsdb.WriteToCacheOption{
				ProgramHash: createRequest.ProgramHash,
				InputHashes: createRequest.InputHashes,
			})
			require.NoError(t, err)
			requireCacheCount(t, 1, d) // there should be a cache entry now.

			// second job should get a read hit and then write that result back.
			expected, err := pjsdb.CreateJob(d.ctx, d.tx, createRequest)
			require.NoError(t, err)
			got, err := pjsdb.GetJob(d.ctx, d.tx, expected)
			require.NoError(t, err)
			require.True(t, !got.Done.IsZero())
			requireCacheCount(t, 2, d) // there should be two entries in the cache now.

			// last job should get a read hit, but not write the result back.
			createRequest.CacheWriteEnabled = false
			expected, err = pjsdb.CreateJob(d.ctx, d.tx, createRequest)
			require.NoError(t, err)
			got, err = pjsdb.GetJob(d.ctx, d.tx, expected)
			require.NoError(t, err)
			require.True(t, !got.Done.IsZero())
			requireCacheCount(t, 2, d) // there should still be two entries in the cache now.
		})
	})
	t.Run("valid/cache_error", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			// first job should miss cache read, but successfully write.
			createRequest := makeReq(t, d, 0, nil)
			id, err := pjsdb.CreateJob(d.ctx, d.tx, createRequest)
			require.NoError(t, err)
			err = pjsdb.ErrorJob(d.ctx, d.tx, id, pjs.JobErrorCode_FAILED, pjsdb.WriteToCacheOption{
				ProgramHash: createRequest.ProgramHash,
				InputHashes: createRequest.InputHashes,
			})
			require.NoError(t, err)
			requireCacheCount(t, 1, d) // there should be a cache entry now.

			// second job should get a read hit and then write that result back.
			expected, err := pjsdb.CreateJob(d.ctx, d.tx, createRequest)
			require.NoError(t, err)
			got, err := pjsdb.GetJob(d.ctx, d.tx, expected)
			require.NoError(t, err)
			require.NotEqual(t, 0, len(got.Error))
			requireCacheCount(t, 2, d) // there should be two entries in the cache now.
		})
	})
}
