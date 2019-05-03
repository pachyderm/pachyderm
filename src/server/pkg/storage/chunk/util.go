package chunk

import (
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

const (
	// Prefix is the default chunk storage prefix.
	Prefix = "chunks"
)

// LocalStorage creates a local chunk storage instance.
// Useful for storage layer tests.
func LocalStorage(tb testing.TB) (obj.Client, *Storage) {
	wd, err := os.Getwd()
	require.NoError(tb, err)
	objC, err := obj.NewLocalClient(wd)
	require.NoError(tb, err)
	return objC, NewStorage(objC, Prefix)
}
