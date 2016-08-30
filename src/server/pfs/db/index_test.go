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
	port int32 = 40651
)

func TestDiffPrefixIndex(t *testing.T) {
	testSetup(t, func(dbClient *gorethink.Session) {
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
		fmt.Printf("keys: %v", keys)
		for i, key := range keys {
			// This is kinda hacky, but we have to compare the string
			// representations of these two values because DeepEqual doesn't
			// work when the underlying types are different.  In this case,
			// key is an interface{} and the expected value is an []interface{},
			// even though their underlying values are equal.
			require.Equal(t, fmt.Sprintf("%v", []interface{}{"repo", expectedPrefixes[i], []interface{}{"branch", 1}}), fmt.Sprintf("%v", key))
		}
	})
}

func testSetup(t *testing.T, testCode func(*gorethink.Session)) {
	dbClient, err := gorethink.Connect(gorethink.ConnectOpts{
		Address: RethinkAddress,
		Timeout: connectTimeoutSeconds * time.Second,
	})
	require.NoError(t, err)

	testCode(dbClient)
}
