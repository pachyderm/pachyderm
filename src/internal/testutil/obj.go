package testutil

import (
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// NewObjectClient creates a Client which is cleaned up after the test exists
func NewObjectClient(t testing.TB) (Client, string) {
	dir := testutil.MkdirTemp(t)
	objC, err := obj.NewLocalClient(dir)
	require.NoError(t, err)
	return objC, strings.ReplaceAll(strings.Trim(dir, "/"), "/", ".")
}
