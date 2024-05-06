package v2_10_0_test

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	v2_10_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.10.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/proto"
)

func TestPipelineVersionsDeduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark in short mode")
	}
	// - setup database with collections.pipelines, collections.jobs
	// - migrate pipeline A from versions [1, 1, 2] -> [1, 2, 3]
	type pipelineVersion struct {
		createdAt time.Time
		version   int
	}
	type testCase struct {
		desc    string
		initial []pipelineVersion
		want    []pipelineVersion
		wantErr bool
	}
	var (
		baseTime = time.Now().Add(-time.Hour * 24 * 365)
		tcs      = []testCase{
			{
				desc: "no changes needed result in no changes made",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
				},
			},
			{
				desc: "fail when there are two pipeline versions at the very same instant",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   1,
					},
				},
				wantErr: true,
			},
			{
				desc: "do not fail when there are two pipeline versions with different version numbers at the very same instant",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
				},
				wantErr: true,
			},
			{
				desc: "renumber simple duplication at beginning",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
				},
			},
			{
				desc: "renumber out-of-order simple duplication at beginning",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
				},
			},
			{
				desc: "renumber simple duplication in the middle",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(4 * time.Second),
						version:   3,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
					{
						createdAt: baseTime.Add(4 * time.Second),
						version:   4,
					},
				},
			},
			{
				desc: "renumber simple duplication in the middle with duplicate timestamps but different versions",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   4,
					},
				},
			},
			{
				desc: "renumber out-of-order simple duplication in the middle",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(4 * time.Second),
						version:   3,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
					{
						createdAt: baseTime.Add(4 * time.Second),
						version:   4,
					},
				},
			},
			{
				desc: "renumber simple duplication at end",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
				},
			},
			{
				desc: "renumber out-of-order simple duplication at end",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					},
				},
			},
			{
				desc: "renumber complex deduplication",
				initial: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					}, {
						createdAt: baseTime.Add(5 * time.Second),
						version:   3,
					},
					{
						createdAt: baseTime.Add(4 * time.Second),
						version:   4,
					},
					{
						createdAt: baseTime.Add(6 * time.Second),
						version:   5,
					}, {
						createdAt: baseTime.Add(7 * time.Second),
						version:   5,
					},
					{
						createdAt: baseTime.Add(8 * time.Second),
						version:   5,
					},
					{
						createdAt: baseTime.Add(9 * time.Second),
						version:   6,
					},
				},
				want: []pipelineVersion{
					{
						createdAt: baseTime.Add(1 * time.Second),
						version:   1,
					},
					{
						createdAt: baseTime.Add(2 * time.Second),
						version:   2,
					},
					{
						createdAt: baseTime.Add(3 * time.Second),
						version:   3,
					}, {
						createdAt: baseTime.Add(4 * time.Second),
						version:   4,
					},
					{
						createdAt: baseTime.Add(5 * time.Second),
						version:   5,
					},
					{
						createdAt: baseTime.Add(6 * time.Second),
						version:   6,
					}, {
						createdAt: baseTime.Add(7 * time.Second),
						version:   7,
					},
					{
						createdAt: baseTime.Add(8 * time.Second),
						version:   8,
					},
					{
						createdAt: baseTime.Add(9 * time.Second),
						version:   9,
					},
				},
			},
		}
	)
	var db *pachsql.DB
	c := pachd.NewTestPachd(t, pachd.TestPachdOption{
		MutateFullOption: func(fullOption *pachd.FullOption) {
			// run all but the last migration step
			s := v2_10_0.MigratePartial(clusterstate.State_2_8_0, 1)
			fullOption.DesiredState = &s
		},
		MutateEnv: func(env *pachd.Env) {
			db = env.DB
		},
	})
	for _, tc := range tcs {
		repo := "input"
		pipeline := &pps.Pipeline{Name: "pipeline", Project: &pfs.Project{Name: pfs.DefaultProjectName}}
		require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repo))
		require.NoError(t, c.CreatePipeline(
			pfs.DefaultProjectName,
			"pipeline",
			"", /* default image*/
			[]string{"cp", "-r", "/pfs/in", "/pfs/out"},
			nil, /* stdin */
			nil, /* spec */
			&pps.Input{Pfs: &pps.PFSInput{Project: pfs.DefaultProjectName, Repo: repo, Glob: "/*", Name: "in"}},
			"",    /* output */
			false, /* update */
		))

		// update the pipeline to create the necessary number of versions
		for i := 0; i < len(tc.initial)-1; i++ {
			c.RerunPipeline(c.Ctx(), &pps.RerunPipelineRequest{Pipeline: pipeline})
		}

		// set database to bad state
		tx := db.MustBeginTx(c.Ctx(), nil)
		pKeys, pis := listPipelineVersions(t, c.Ctx(), tx, pipeline)
		var pipUpdates []*v2_10_0.PipUpdateRow
		for i, v := range tc.initial {
			if v.version != i+1 {
				pis[i].Version = uint64(i)
				data, err := proto.Marshal(pis[i])
				require.NoError(t, err)
				vIdx := v2_10_0.VersionKey(pipeline.Project.Name, pipeline.Name, pis[i].Version)
				update := &v2_10_0.PipUpdateRow{Key: pKeys[i], Proto: data, IdxVersion: vIdx}
				pipUpdates = append(pipUpdates, update)
			}
		}
		require.NoError(t, v2_10_0.UpdatePipelineRows(c.Ctx(), tx, pipUpdates))
		// TODO update jobs as well
		require.NoError(t, tx.Commit())

		// verify bad state
		pis, err := c.ListPipeline()
		require.NoError(t, err)
		require.Len(t, pis, 1)
		jis, err := c.ListJob(pfs.DefaultProjectName, pipeline.Name, nil, 0, false)
		require.NoError(t, err)
		require.Len(t, jis, 1)

		// test
		tx = db.MustBeginTx(c.Ctx(), nil)
		require.NoError(t, v2_10_0.DeduplicatePipelineVersions(c.Ctx(), migrations.Env{Tx: tx}))
		require.NoError(t, tx.Commit())

		// assert that there should be no duplicate pipelines detected in query
		// assert that pipelines and jobs should list without error
		pis, err = c.ListPipeline()
		require.NoError(t, err)
		require.Len(t, pis, 1)
		jis, err = c.ListJob(pfs.DefaultProjectName, pipeline.Name, nil, 0, false)
		require.NoError(t, err)
		require.Len(t, jis, 1)

		// clean up
		require.NoError(t, c.DeleteAll(c.Ctx()))
		undoDeduplicatePipelineVersions(t, db)
	}
}

func undoDeduplicatePipelineVersions(t *testing.T, db *pachsql.DB) {
	_, err := db.Exec("DROP INDEX pip_version_idx;")
	require.NoError(t, err)
}

type pipelineRow struct {
	Key   string `db:"key"`
	Proto []byte `db:"proto"`
}

func listPipelineVersions(t *testing.T, ctx context.Context, tx *pachsql.Tx, pipeline *pps.Pipeline) ([]string, []*pps.PipelineInfo) {
	query := `SELECT key, proto FROM collections.pipelines ORDER BY createdat`
	rr, err := tx.QueryxContext(ctx, query)
	require.NoError(t, err)
	defer rr.Close()
	var keys []string
	var pis []*pps.PipelineInfo
	for rr.Next() {
		var row pipelineRow
		pi := &pps.PipelineInfo{}
		require.NoError(t, rr.Err())
		require.NoError(t, rr.StructScan(&row))
		require.NoError(t, proto.Unmarshal(row.Proto, pi))
		keys = append(keys, row.Key)
		pis = append(pis, pi)
	}
	return keys, pis
}
