package ppsdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

func newTestDB(t testing.TB, ctx context.Context) *pachsql.DB {
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	return db
}

func comparePipelines(t *testing.T, expected, got *pps.PipelineInfo) {
	t.Helper()
	expected = protoutil.Clone(expected)
	got = protoutil.Clone(got)
	require.NoDiff(t, expected, got, []cmp.Option{protocmp.Transform()})
}

func withTx(t *testing.T, ctx context.Context, db *pachsql.DB, f func(context.Context, *pachsql.Tx)) {
	t.Helper()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	f(ctx, tx)
	require.NoError(t, tx.Commit())
}

func TestGetPipeline(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		expectedInfo := &pps.PipelineInfo{
			Pipeline: &pps.Pipeline{
				Project: &pfs.Project{Name: "foo"},
				Name:    "bar",
			},
			SpecCommit: &pfs.Commit{
				Repo: &pfs.Repo{
					Project: &pfs.Project{Name: "foo"},
					Name:    "bar",
					Type:    pfs.SpecRepoType,
				},
				Id: uuid.NewWithoutDashes(),
			},
		}
		err := pfsdb.UpsertProject(ctx, tx, &pfs.ProjectInfo{Project: expectedInfo.Pipeline.Project})
		require.NoError(t, err, "should be able to create project")
		repo := &pfs.Repo{
			Project: expectedInfo.Pipeline.Project,
			Name:    "bar",
			Type:    pfs.SpecRepoType,
		}
		_, err = pfsdb.UpsertRepo(ctx, tx, &pfs.RepoInfo{
			Repo: repo,
		})
		require.NoError(t, err, "should be able to create repo")
		_, err = pfsdb.CreateCommit(ctx, tx, &pfs.CommitInfo{
			Commit:  expectedInfo.SpecCommit,
			Origin:  &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO},
			Started: timestamppb.New(time.Now()),
		})
		require.NoError(t, err, "should be able to create commit")
		_, err = pfsdb.UpsertBranch(ctx, tx, &pfs.BranchInfo{
			Branch: &pfs.Branch{
				Repo: repo,
				Name: "master",
			},
			Head: expectedInfo.SpecCommit,
		})
		require.NoError(t, err, "should be able to create branch")

		err = ppsdb.UpsertPipeline(ctx, tx, expectedInfo)
		require.NoError(t, err, "should be able to create pipeline")
		got, err := ppsdb.GetPipeline(ctx, tx, expectedInfo.Pipeline.Project.Name, expectedInfo.Pipeline.Name)
		require.NoError(t, err, "should be able to get pipeline")
		comparePipelines(t, expectedInfo, got.PipelineInfo)
		// Validate error for attempting to get non-existent pipeline.
		_, err = ppsdb.GetPipeline(ctx, tx, "foobar", "baz")
		require.YesError(t, err, "getting a nonexistent pipeline should fail")
	})
}

func TestUpsertPipeline(t *testing.T) {

	t.Parallel()
	ctx := pctx.TestContext(t)
	db := newTestDB(t, ctx)
	expectedInfo := &pps.PipelineInfo{
		Pipeline: &pps.Pipeline{
			Project: &pfs.Project{Name: "foo"},
			Name:    "bar",
		},
		SpecCommit: &pfs.Commit{
			Repo: &pfs.Repo{
				Project: &pfs.Project{Name: "foo"},
				Name:    "bar",
				Type:    pfs.SpecRepoType,
			},
			Id: uuid.NewWithoutDashes(),
		},
		Details: &pps.PipelineInfo_Details{},
	}
	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		err := pfsdb.UpsertProject(ctx, tx, &pfs.ProjectInfo{Project: expectedInfo.Pipeline.Project})
		require.NoError(t, err, "should be able to create project")
		repo := &pfs.Repo{
			Project: expectedInfo.Pipeline.Project,
			Name:    "bar",
			Type:    pfs.SpecRepoType,
		}
		_, err = pfsdb.UpsertRepo(ctx, tx, &pfs.RepoInfo{
			Repo: repo,
		})
		require.NoError(t, err, "should be able to create repo")
		_, err = pfsdb.CreateCommit(ctx, tx, &pfs.CommitInfo{
			Commit:  expectedInfo.SpecCommit,
			Origin:  &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO},
			Started: timestamppb.New(time.Now()),
		})
		require.NoError(t, err, "should be able to create commit")
		_, err = pfsdb.UpsertBranch(ctx, tx, &pfs.BranchInfo{
			Branch: &pfs.Branch{
				Repo: repo,
				Name: "master",
			},
			Head: expectedInfo.SpecCommit,
		})
		require.NoError(t, err, "should be able to create branch")

		err = ppsdb.UpsertPipeline(ctx, tx, expectedInfo)
		require.NoError(t, err, "should be able to create pipeline")
		got, err := ppsdb.GetPipeline(ctx, tx, expectedInfo.Pipeline.Project.Name, expectedInfo.Pipeline.Name)
		require.NoError(t, err, "should be able to get pipeline")
		comparePipelines(t, expectedInfo, got.PipelineInfo)
		// Validate error for attempting to get non-existent pipeline.
		_, err = ppsdb.GetPipeline(ctx, tx, "foobar", "baz")
		require.YesError(t, err, "getting a nonexistent pipeline should fail")
	})

	withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		expectedInfo.Details.Description = "quux"
		err := ppsdb.UpsertPipeline(ctx, tx, expectedInfo)
		require.NoError(t, err, "should be able to update pipeline")

		got, err := ppsdb.GetPipeline(ctx, tx, expectedInfo.Pipeline.Project.Name, expectedInfo.Pipeline.Name)
		require.NoError(t, err, "should be able to get pipeline")
		comparePipelines(t, expectedInfo, got.PipelineInfo)
	})

}
