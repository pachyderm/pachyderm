package pjsdb_test

import (
	"database/sql"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestCreateJobContext(t *testing.T) {
	t.Run("valid/not_exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, 0)
			require.NoError(t, err)
			job, err := pjsdb.GetJob(d.ctx, d.tx, id)
			require.NoError(t, err)
			require.Nil(t, job.ContextHash, "no job context should exist")
		})
	})
	t.Run("valid/exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, 0)
			require.NoError(t, err)
			jobCtx, err := pjsdb.CreateJobContext(d.ctx, d.tx, id)
			require.NoError(t, err)
			job, err := pjsdb.GetJob(d.ctx, d.tx, id)
			require.NoError(t, err)
			require.Equal(t, jobCtx.Hash, job.ContextHash)
		})
	})
	t.Run("invalid/double_create", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, 0)
			require.NoError(t, err)
			_, err = pjsdb.CreateJobContext(d.ctx, d.tx, id)
			require.NoError(t, err)
			_, err = pjsdb.CreateJobContext(d.ctx, d.tx, id)
			require.YesError(t, err)
			require.ErrorIs(t, err, pjsdb.ErrJobContextAlreadyExists)
		})
	})
}

func TestResolveJobContext(t *testing.T) {
	t.Run("invalid/not_exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			_, err := pjsdb.ResolveJobContext(d.ctx, d.tx, []byte("token"))
			require.ErrorIs(t, err, sql.ErrNoRows)
		})
	})
	t.Run("valid/exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, 0)
			require.NoError(t, err)
			jobCtx, err := pjsdb.CreateJobContext(d.ctx, d.tx, id)
			require.NoError(t, err)
			newId, err := pjsdb.ResolveJobContext(d.ctx, d.tx, jobCtx.Token)
			require.NoError(t, err)
			require.NoDiff(t, id, newId, nil)
		})
	})
	t.Run("invalid/revoked", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, 0)
			require.NoError(t, err)
			jobCtx, err := pjsdb.CreateJobContext(d.ctx, d.tx, id)
			require.NoError(t, err)
			newId, err := pjsdb.ResolveJobContext(d.ctx, d.tx, jobCtx.Token)
			require.NoError(t, err)
			require.NoDiff(t, id, newId, nil)
			require.NoError(t, pjsdb.RevokeJobContext(d.ctx, d.tx, id))
			_, err = pjsdb.ResolveJobContext(d.ctx, d.tx, jobCtx.Token)
			require.YesError(t, err)
			require.ErrorIs(t, err, sql.ErrNoRows)
		})
	})
}

func TestRevokeJobContext(t *testing.T) {
	t.Run("invalid/not_exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			require.ErrorIs(t, pjsdb.RevokeJobContext(d.ctx, d.tx, 1), sql.ErrNoRows)
			id, err := createJob(t, d, 0)
			require.NoError(t, err)
			require.ErrorIs(t, pjsdb.RevokeJobContext(d.ctx, d.tx, id), sql.ErrNoRows)
		})
	})
	t.Run("valid/exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			id, err := createJob(t, d, 0)
			require.NoError(t, err)
			_, err = pjsdb.CreateJobContext(d.ctx, d.tx, id)
			require.NoError(t, err)
			require.NoError(t, pjsdb.RevokeJobContext(d.ctx, d.tx, id))
			jobCtx := &pjsdb.JobContext{}
			err = d.tx.GetContext(d.ctx, jobCtx, `SELECT context_hash FROM pjs.jobs WHERE id = $1 AND context_hash IS NOT NULL`, id)
			t.Log(jobCtx, err)
			require.NoDiff(t, *jobCtx, pjsdb.JobContext{}, nil)
			require.YesError(t, err)
			require.ErrorIs(t, err, sql.ErrNoRows)
		})
	})
}
