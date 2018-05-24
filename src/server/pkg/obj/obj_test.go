package obj

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func TestSimple(t *testing.T) {
	t.Run("local", func(t *testing.T) {
		localClient, err := NewLocalClient(tu.UniqueString("/tmp/"))
		require.NoError(t, err)
		testClient(t, localClient)
	})
}

func requireContent(t *testing.T, c Client, name, content string) {
	t.Helper()
	data, err := ReadObj(c, name)
	require.NoError(t, err)
	require.Equal(t, content, data)
}

func testClient(t *testing.T, c Client) {
	require.NoError(t, WriteObj(c, "1", "foo"))
	requireContent(t, c, "1", "foo")
	require.NoError(t, c.Copy("1", "2"))
	requireContent(t, c, "1", "foo")
	requireContent(t, c, "2", "foo")
	files, err := WalkAll(c, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(files))
	require.Equal(t, files[0], "1")
	require.Equal(t, files[1], "2")
	require.NoError(t, c.Delete("1"))
	requireContent(t, c, "2", "foo")
	_, err = ReadObj(c, "1")
	require.YesError(t, err)
	files, err = WalkAll(c, "")
	require.NoError(t, err)
	require.Equal(t, 1, len(files))
	require.Equal(t, files[0], "2")
}
