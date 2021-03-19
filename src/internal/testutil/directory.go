package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// MkdirTemp creates a temporary directory with a random name under the OS's
// temp directory and deletes it at the end of the test.
func MkdirTemp(t testing.TB) string {
	basePath := filepath.Join(os.TempDir(), "pachyderm_test")
	require.NoError(t, os.MkdirAll(basePath, 0700))

	// TODO: change this to os.MkdirTemp when we move to go1.16+
	dir, err := ioutil.TempDir(basePath, "")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
	})

	return dir
}
