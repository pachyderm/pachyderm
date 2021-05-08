package obj

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestLocalClient(t *testing.T) {
	t.Parallel()
	TestSuite(t, newTestLocalClient)
}

func newTestLocalClient(t testing.TB) Client {
	dir := t.TempDir()
	t.Log("testing local client in", dir)
	c, err := NewLocalClient(dir)
	require.NoError(t, err)
	return c
}
