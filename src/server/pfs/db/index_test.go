package persist

import (
	"fmt"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pfs/db/persist"

	"github.com/dancannon/gorethink"
)

var (
	RethinkAddress = "localhost:28015"
)

func TestDiffPathIndex(t *testing.T) {
	dbClient := getClient(t)
	path := "/foo/bar/fizz/buzz"
	cursor, err := gorethink.Expr(DiffPathIndex.CreateFunction(gorethink.Expr(&persist.Diff{
		Repo: "repo",
		Path: path,
		Clock: &persist.Clock{
			Branch: "branch",
			Clock:  1,
		},
	}))).Run(dbClient)
	require.NoError(t, err)
	var key []interface{}
	require.NoError(t, cursor.All(&key))
	require.Equal(t, fmt.Sprintf("%v", []interface{}{"repo", path, []interface{}{"branch", 1}}), fmt.Sprintf("%v", key))
}

func TestDiffPrefixIndex(t *testing.T) {
	dbClient := getClient(t)
	cursor, err := gorethink.Expr(DiffPrefixIndex.CreateFunction(gorethink.Expr(&persist.Diff{
		Repo: "repo",
		Path: "/foo/bar/fizz/buzz",
		Clock: &persist.Clock{
			Branch: "branch",
			Clock:  1,
		},
	}))).Run(dbClient)
	require.NoError(t, err)
	keys := []interface{}{}
	expectedPrefixes := []interface{}{
		"/",
		"/foo",
		"/foo/bar",
		"/foo/bar/fizz",
	}
	require.NoError(t, cursor.All(&keys))
	for i, key := range keys {
		// This is kinda hacky, but we have to compare the string
		// representations of these two values because DeepEqual doesn't
		// work when the underlying types are different.  In this case,
		// key is an interface{} and the expected value is an []interface{},
		// even though their underlying values are equal.
		require.Equal(t, fmt.Sprintf("%v", []interface{}{"repo", expectedPrefixes[i], []interface{}{"branch", 1}}), fmt.Sprintf("%v", key))
	}
}

func TestDiffParentIndex(t *testing.T) {
	dbClient := getClient(t)
	path := "/foo/bar/fizz/buzz"
	cursor, err := gorethink.Expr(DiffParentIndex.CreateFunction(gorethink.Expr(&persist.Diff{
		Repo: "repo",
		Path: path,
		Clock: &persist.Clock{
			Branch: "branch",
			Clock:  1,
		},
	}))).Run(dbClient)
	require.NoError(t, err)
	var key []interface{}
	require.NoError(t, cursor.All(&key))
	require.Equal(t, fmt.Sprintf("%v", []interface{}{"repo", "/foo/bar/fizz", []interface{}{"branch", 1}}), fmt.Sprintf("%v", key))
}

func TestDiffClockIndex(t *testing.T) {
	dbClient := getClient(t)
	path := "/foo/bar/fizz/buzz"
	cursor, err := gorethink.Expr(DiffClockIndex.CreateFunction(gorethink.Expr(&persist.Diff{
		Repo: "repo",
		Path: path,
		Clock: &persist.Clock{
			Branch: "branch",
			Clock:  1,
		},
	}))).Run(dbClient)
	require.NoError(t, err)
	var key []interface{}
	require.NoError(t, cursor.All(&key))
	require.Equal(t, fmt.Sprintf("%v", []interface{}{"repo", "branch", 1}), fmt.Sprintf("%v", key))
}

func TestCommitClockIndex(t *testing.T) {
	dbClient := getClient(t)
	cursor, err := gorethink.Expr(CommitClockIndex.CreateFunction(gorethink.Expr(&persist.Commit{
		Repo: "repo",
		FullClock: []*persist.Clock{
			&persist.Clock{
				Branch: "master",
				Clock:  0,
			},
			&persist.Clock{
				Branch: "branch",
				Clock:  1,
			},
		},
	}))).Run(dbClient)
	require.NoError(t, err)
	var key []interface{}
	require.NoError(t, cursor.All(&key))
	require.Equal(t, fmt.Sprintf("%v", []interface{}{"repo", "branch", 1}), fmt.Sprintf("%v", key))
}

func getClient(t *testing.T) *gorethink.Session {
	dbClient, err := gorethink.Connect(gorethink.ConnectOpts{
		Address: RethinkAddress,
		Timeout: connectTimeoutSeconds * time.Second,
	})
	require.NoError(t, err)
	return dbClient
}
