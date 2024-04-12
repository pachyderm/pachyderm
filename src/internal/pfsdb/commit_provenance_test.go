package pfsdb_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	v2_6_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.6.0"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func TestCommitSetProvenance(suite *testing.T) {
	ctx := pctx.TestContext(suite)
	db, _ := dockertestenv.NewEphemeralPostgresDB(ctx, suite)
	defer db.Close()
	// setup schema
	withTx(suite, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
		_, err := tx.ExecContext(ctx, `CREATE SCHEMA collections`)
		require.NoError(suite, err)
		require.NoError(suite, col.SetupPostgresV0(ctx, tx))
		require.NoError(suite, col.SetupPostgresCollections(ctx, tx, col.NewPostgresCollection("commits", db, nil, &pfs.CommitInfo{}, nil)))
		_, err = tx.ExecContext(ctx, `CREATE SCHEMA pfs`)
		require.NoError(suite, err)
		require.NoError(suite, v2_6_0.SetupCommitProvenanceV0(ctx, tx))
	})
	suite.Cleanup(func() {
		db.Close()
	})
	proj := "my_project"
	suite.Run("Basic", func(t *testing.T) {
		// setup basic DAG
		//              -- C
		//             /
		// A <- B <---
		//             \
		//              -- D
		//             /
		// E <--------
		td, cf := NewTestDAG(proj)
		defer require.NoError(t, cf(db))
		withTx(t, ctx, db, func(_ context.Context, tx *pachsql.Tx) {
			require.NoError(t, td.addRepo(tx, "A"))
			require.NoError(t, td.addRepo(tx, "B", "A"))
			require.NoError(t, td.addRepo(tx, "C", "B"))
			require.NoError(t, td.addRepo(tx, "E"))
			require.NoError(t, td.addRepo(tx, "D", "B", "E"))
		})
		// setup basic commit graph of
		//                -- C@y
		//               /
		// A@v <- B@x <--
		//               \
		//                -- D@z
		//               /
		// E@w <---------
		var a, b, c, d, e *pfs.Commit
		withTx(t, ctx, db, func(_ context.Context, tx *pachsql.Tx) {
			var err error
			a, err = td.addCommitSet(tx, "v", "A")
			require.NoError(t, err)
			b, err = td.addCommitSet(tx, "x", "B")
			require.NoError(t, err)
			c, err = td.addCommitSet(tx, "y", "C")
			require.NoError(t, err)
			e, err = td.addCommitSet(tx, "w", "E")
			require.NoError(t, err)
			d, err = td.addCommitSet(tx, "z", "D")
			require.NoError(t, err)
		})
		// assert commit set provenance
		// check y's commit set provenance
		withTx(t, ctx, db, func(__ context.Context, tx *pachsql.Tx) {
			yProv, err := pfsdb.CommitSetProvenance(tx, "y")
			require.NoError(t, err)
			checkCommitsEqual(t, []*pfs.Commit{a, b}, yProv)
		})
		// check y's commit set subvenance
		withTx(t, ctx, db, func(_ context.Context, tx *pachsql.Tx) {
			ySubv, err := pfsdb.CommitSetSubvenance(tx, "y")
			require.NoError(t, err)
			checkCommitsEqual(t, []*pfs.Commit{}, ySubv)
		})
		// check z's commit set provenance
		withTx(t, ctx, db, func(_ context.Context, tx *pachsql.Tx) {
			zProv, err := pfsdb.CommitSetProvenance(tx, "z")
			require.NoError(t, err)
			checkCommitsEqual(t, []*pfs.Commit{a, b, e}, zProv)
		})
		// check x's commit set subvenance
		withTx(t, ctx, db, func(_ context.Context, tx *pachsql.Tx) {
			xSubv, err := pfsdb.CommitSetSubvenance(tx, "x")
			require.NoError(t, err)
			dAtW := client.NewCommit(proj, "D", "", "w")
			checkCommitsEqual(t, []*pfs.Commit{c, dAtW, d}, xSubv)
		})
	})
	suite.Run("SameCommitSet", func(t *testing.T) {
		// opencv DAG
		//           ---- edges <-
		//          /             \
		// images <-               \
		//          \               \
		//           -----------  montage
		td, cf := NewTestDAG(proj)
		defer require.NoError(t, cf(db))
		withTx(t, ctx, db, func(_ context.Context, tx *pachsql.Tx) {
			require.NoError(t, td.addRepo(tx, "images"))
			require.NoError(t, td.addRepo(tx, "edges", "images"))
			require.NoError(t, td.addRepo(tx, "montage", "images", "edges"))
		})
		withTx(t, ctx, db, func(_ context.Context, tx *pachsql.Tx) {
			var err error
			_, err = td.addCommitSet(tx, "x", "images")
			require.NoError(t, err)
		})
		withTx(t, ctx, db, func(_ context.Context, tx *pachsql.Tx) {
			var err error
			xSubv, err := pfsdb.CommitSetSubvenance(tx, "x")
			require.NoError(t, err)
			require.Len(t, xSubv, 0)
			xProv, err := pfsdb.CommitSetProvenance(tx, "x")
			require.NoError(t, err)
			require.Len(t, xProv, 0)
		})
	})
}

func TestGetCommitWithIDProvenance(t *testing.T) {
	commits := make(map[int]*pfsdb.CommitWithID)
	size := 10
	withDB(t, func(ctx context.Context, t *testing.T, db *pachsql.DB) {
		withTx(t, ctx, db, func(ctx context.Context, tx *pachsql.Tx) {
			for i := 1; i <= size; i++ { // row ID in postgres starts at 1.
				commit := testCommitWithCommitKey(ctx, t, tx, fmt.Sprintf("r%d", i), fmt.Sprintf("c%d", i))
				id, err := pfsdb.CreateCommit(ctx, tx, commit)
				require.NoError(t, err, "should be able to create commit")
				if i > 1 { // make every commit provenant on the commit before it.
					commit.DirectProvenance = []*pfs.Commit{commits[i-1].Commit}
					createCommitProvenance(ctx, t, tx, id, commit.DirectProvenance)
				}
				commits[i] = &pfsdb.CommitWithID{
					ID:         id,
					CommitInfo: commit,
				}
			}
			provenantCommits, err := pfsdb.GetCommitWithIDProvenance(ctx, tx, commits[size].ID)
			require.NoError(t, err, "should be able to get provenant commits")
			require.Equal(t, size-1, len(commits)-1) // ignore commit whose id is size.
			for _, commit := range provenantCommits {
				_, ok := commits[int(commit.ID)]
				require.True(t, ok, "found provenant commit should exist in map of created commits")
			}
			subvenantCommits, err := pfsdb.GetCommitWithIDSubvenance(ctx, tx, commits[1].ID)
			require.NoError(t, err, "should be able to get subvenant commits")
			require.Equal(t, size-1, len(commits)-1) // ignore commit whose id is size.
			for _, commit := range subvenantCommits {
				_, ok := commits[int(commit.ID)]
				require.True(t, ok, "found subvenant commit should exist in map of created commits")
			}
			// test options
			limit := uint64(3)
			provenantCommits, err = pfsdb.GetCommitWithIDProvenance(ctx, tx, commits[size].ID, pfsdb.WithLimit(limit))
			require.NoError(t, err, "should be able to get provenant commits")
			require.Equal(t, uint64(len(provenantCommits)), limit) // ignore commit whose id is size.
			maxDepth := uint64(2)
			subvenantCommits, err = pfsdb.GetCommitWithIDSubvenance(ctx, tx, commits[1].ID, pfsdb.WithMaxDepth(maxDepth))
			require.NoError(t, err, "should be able to get subvenant commits")
			require.Equal(t, uint64(len(subvenantCommits)), maxDepth) // ignore commit whose id is 1.
		})
	})
}

func createCommitProvenance(ctx context.Context, t *testing.T, tx *pachsql.Tx, fromId pfsdb.CommitID, provenantCommits []*pfs.Commit) {
	createCommitProvenanceFromRows(ctx, t, tx, fromId, getProvenantCommitRows(ctx, t, tx, provenantCommits))
}

func getProvenantCommitRows(ctx context.Context, t *testing.T, tx *pachsql.Tx, provenantCommits []*pfs.Commit) []pfsdb.Commit {
	provCommits := ""
	for _, provCommit := range provenantCommits {
		provCommits = provCommits + fmt.Sprintf("'%s', ", pfsdb.CommitKey(provCommit))
	}
	if len(provCommits) > 0 {
		provCommits = provCommits[:len(provCommits)-2] // remove ", " from last element.
	}
	rows := make([]pfsdb.Commit, 0)
	require.NoError(t, tx.SelectContext(ctx, &rows,
		fmt.Sprintf("SELECT DISTINCT commit.commit_id, commit.int_id FROM pfs.commits commit WHERE commit_id IN (%s)", provCommits)),
		"should be able to get commit ids")
	return rows
}

func createCommitProvenanceFromRows(ctx context.Context, t *testing.T, tx *pachsql.Tx, fromId pfsdb.CommitID, provCommitRows []pfsdb.Commit) {
	provInsertQuery := `
		INSERT INTO pfs.commit_provenance
		(from_id, to_id)
		VALUES %s
		ON CONFLICT DO NOTHING;`
	values := ""
	for _, provCommitRow := range provCommitRows {
		values += fmt.Sprintf("(%d, %d), ", fromId, provCommitRow.ID)
	}
	if len(values) > 0 {
		values = values[:len(values)-2] // remove ", " from last element.
	}
	_, err := tx.ExecContext(ctx, fmt.Sprintf(provInsertQuery, values))
	require.NoError(t, err, "should be able to create provenance from rows")
}

type testDAG struct {
	project string
	// maps a repo to all of its repos in direct provenance
	// (we simplify branches for the purpose of simplicity in this test
	provDag map[string][]string
	// inverse of provDag
	subvDag map[string][]string
	// maps repo to its latest commit
	heads map[string]*pfs.Commit
}

func NewTestDAG(project string) (*testDAG, func(*pachsql.DB) error) {
	return &testDAG{
			project: project,
			provDag: make(map[string][]string),
			subvDag: make(map[string][]string),
			heads:   make(map[string]*pfs.Commit),
		}, func(db *pachsql.DB) error {
			stmt := `DELETE FROM pfs.commits`
			_, err := db.Exec(stmt)
			return errors.Wrapf(err, "delete pfs.commits")
		}
}

func (td *testDAG) addRepo(tx *pachsql.Tx, repo string, provRepos ...string) error {
	if _, ok := td.provDag[repo]; ok {
		return errors.Errorf("repo %q already exists", repo)
	}
	commitID := uuid.New()
	c := client.NewCommit(td.project, repo, "", commitID)
	if err := addCommitWrapper(tx, c); err != nil {
		return err
	}
	td.heads[repo] = c
	for _, r := range provRepos {
		if _, ok := td.provDag[r]; ok {
			td.subvDag[r] = append(td.subvDag[r], repo)
		} else {
			return errors.Errorf("prov repo %q must exist", r)
		}
		if err := pfsdb.AddCommitProvenance(tx, c, td.heads[r]); err != nil {
			return err
		}
	}
	td.provDag[repo] = provRepos
	if len(provRepos) == 0 {
		td.provDag[repo] = []string{}
	}
	td.subvDag[repo] = []string{}
	return nil
}

// simulates commit propagation
func (td *testDAG) addCommitSet(tx *pachsql.Tx, commitID string, repo string) (*pfs.Commit, error) {
	if _, ok := td.provDag[repo]; !ok {
		return nil, errors.Errorf("repo %q must exist", repo)
	}
	var bfsQueue []string
	bfsQueue = append(bfsQueue, repo)
	seen := make(map[string]struct{})
	for len(bfsQueue) > 0 {
		var r string
		r, bfsQueue = bfsQueue[0], bfsQueue[1:]
		c := client.NewCommit(td.project, r, "", commitID)
		if _, ok := seen[pfsdb.CommitKey(c)]; !ok {
			if err := addCommitWrapper(tx, c); err != nil {
				return nil, err
			}
			seen[pfsdb.CommitKey(c)] = struct{}{}
		}
		td.heads[r] = c
		for _, prov := range td.provDag[r] {
			if err := pfsdb.AddCommitProvenance(tx, c, td.heads[prov]); err != nil {
				return nil, err
			}
		}
		bfsQueue = append(bfsQueue, td.subvDag[r]...)
	}
	return client.NewCommit(td.project, repo, "", commitID), nil
}

func addCommitWrapper(tx *pachsql.Tx, c *pfs.Commit) error {
	if _, err := tx.Exec(`INSERT INTO collections.commits(key) VALUES ($1);`, pfsdb.CommitKey(c)); err != nil {
		return errors.Wrapf(err, "insert %q to collections.commits", pfsdb.CommitKey(c))
	}
	return pfsdb.AddCommit(tx, c)
}

func withTx(t *testing.T, ctx context.Context, db *pachsql.DB, f func(context.Context, *pachsql.Tx)) {
	t.Helper()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	f(ctx, tx)
	require.NoError(t, tx.Commit())
}

func withFailedTx(t *testing.T, ctx context.Context, db *pachsql.DB, f func(context.Context, *pachsql.Tx)) {
	t.Helper()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	f(ctx, tx)
	require.YesError(t, tx.Commit())
}

func checkCommitsEqual(t *testing.T, expecteds, unsortedActuals []*pfs.Commit) {
	t.Helper()
	require.Equal(t, len(expecteds), len(unsortedActuals))
	sort.Slice(unsortedActuals, func(i, j int) bool {
		return pfsdb.CommitKey(unsortedActuals[i]) < pfsdb.CommitKey(unsortedActuals[j])
	})
	for i := range expecteds {
		require.Equal(t, pfsdb.CommitKey(expecteds[i]), pfsdb.CommitKey(unsortedActuals[i]))
	}
}
