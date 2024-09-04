package client_test

import (
	"testing"
	"testing/fstest"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
)

func TestFileSystemToFileSet(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	var testFS = fstest.MapFS{
		"foo":      &fstest.MapFile{Data: []byte("bar")},
		"baz/quux": &fstest.MapFile{Data: []byte("quuux")},
	}
	fs, err := env.PachClient.FileSystemToFileset(ctx, testFS)
	require.NoError(t, err, "must be able to upload filesystem and get fileset ID")
	t.Log("fileset", fs)
}
