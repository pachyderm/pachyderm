package pjs_test

import (
	"encoding/hex"
	"testing"
	"testing/fstest"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjs"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
)

func TestHashFS(t *testing.T) {
	var testFS = fstest.MapFS{
		"test":   &fstest.MapFile{Data: []byte("123")},
		"abc123": &fstest.MapFile{Data: []byte("foo")},
	}
	h, err := pjs.HashFS(testFS)
	require.NoError(t, err, "hashing filesystem")
	require.Equal(t, "25ec7afdab18e8e5808730def751575fdb964d293a36ce21103415c747ca06db", hex.EncodeToString(h))
}

func TestHashFileset(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	var testFS = fstest.MapFS{
		"test":   &fstest.MapFile{Data: []byte("123")},
		"abc123": &fstest.MapFile{Data: []byte("foo")},
	}
	id, err := env.PachClient.FileSystemToFileset(ctx, testFS)
	require.NoError(t, err, "FileSystemToFileset")
	h, err := pjs.HashFileset(ctx, env.PachClient.FilesetClient, id)
	require.NoError(t, err, "hashing fileset")
	require.Equal(t, "25ec7afdab18e8e5808730def751575fdb964d293a36ce21103415c747ca06db", hex.EncodeToString(h))
}
