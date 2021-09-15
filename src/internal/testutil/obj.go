package testutil

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// NewObjectClient creates a obj.Client which is cleaned up after the test exists
func NewObjectClient(t testing.TB) (obj.Client, string) {
	dir := t.TempDir()
	objC, err := obj.NewLocalClient(dir)
	require.NoError(t, err)
	return objC, objC.(interface{ BucketURL() string }).BucketURL()
}
