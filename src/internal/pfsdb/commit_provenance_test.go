//go:build unit_test

package pfsdb

import (
	"context"
	"sort"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func TestCommitSetProvenance(t *testing.T) {
	db, _ := dockertestenv.NewEphemeralPostgresDB(t)
	defer db.Close()
	ctx := context.Background()
	// setup schema
	tx, err := db.Beginx()
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `CREATE SCHEMA pfs`)
	require.NoError(t, err)
	require.NoError(t, SetupCommitProvenanceV0(ctx, tx))
	require.NoError(t, tx.Commit())
	// create some commits
	tx, err = db.Beginx()
	require.NoError(t, err)
	proj := "my_project"
	a := client.NewProjectCommit(proj, "A", "", "v")
	require.NoError(t, AddCommit(ctx, tx, a))
	b := client.NewProjectCommit(proj, "B", "", "x")
	require.NoError(t, AddCommit(ctx, tx, b))
	c := client.NewProjectCommit(proj, "C", "", "y")
	require.NoError(t, AddCommit(ctx, tx, c))
	d := client.NewProjectCommit(proj, "D", "", "z")
	require.NoError(t, AddCommit(ctx, tx, d))
	e := client.NewProjectCommit(proj, "E", "", "w")
	require.NoError(t, AddCommit(ctx, tx, e))
	// setup basic commit graph of
	//                -- C@y
	//               /
	// A@v <- B@x <--
	//               \
	//                -- D@z
	//               /
	// E@w <---------
	require.NoError(t, AddCommitProvenance(ctx, tx, b, a))
	require.NoError(t, AddCommitProvenance(ctx, tx, c, b))
	require.NoError(t, AddCommitProvenance(ctx, tx, d, b))
	require.NoError(t, AddCommitProvenance(ctx, tx, d, e))
	require.NoError(t, tx.Commit())
	// assert commit set provenance
	tx, err = db.Beginx()
	require.NoError(t, err)
	// check y's commit set provenance
	yProv, err := CommitSetProvenance(ctx, tx, "y")
	require.NoError(t, err)
	sort.Slice(yProv, func(i, j int) bool {
		return CommitKey(yProv[i]) < CommitKey(yProv[j])
	})
	checkCommitsEqual(t, []*pfs.Commit{a, b}, yProv)
	// check y's commit set subvenance
	ySubv, err := CommitSetSubvenance(ctx, tx, "y")
	require.NoError(t, err)
	sort.Slice(ySubv, func(i, j int) bool {
		return CommitKey(ySubv[i]) < CommitKey(ySubv[j])
	})
	checkCommitsEqual(t, []*pfs.Commit{}, ySubv)
	// check z's commit set provenance
	zProv, err := CommitSetProvenance(ctx, tx, "z")
	require.NoError(t, err)
	sort.Slice(zProv, func(i, j int) bool {
		return CommitKey(zProv[i]) < CommitKey(zProv[j])
	})
	checkCommitsEqual(t, []*pfs.Commit{a, b, e}, zProv)
	// check x's commit set subvenance
	xSubv, err := CommitSetSubvenance(ctx, tx, "x")
	require.NoError(t, err)
	sort.Slice(xSubv, func(i, j int) bool {
		return CommitKey(xSubv[i]) < CommitKey(xSubv[j])
	})
	checkCommitsEqual(t, []*pfs.Commit{c, d}, xSubv)
	require.NoError(t, tx.Commit())
}

func checkCommitsEqual(t *testing.T, as, bs []*pfs.Commit) {
	require.Equal(t, len(as), len(bs))
	for i := range as {
		require.Equal(t, CommitKey(as[i]), CommitKey(bs[i]))
	}
}
