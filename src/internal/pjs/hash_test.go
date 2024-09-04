package pjs_test

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"path"
	"testing"
	"testing/fstest"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjs"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
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

func randInt(max int) int {
	i, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic("could not generate random int") // this should never happen
	}
	return int(i.Int64())
}

func populateDir(fs fstest.MapFS, dirName string, max int) {
	if max <= 0 {
		return
	}
	var count = randInt(max) + 1
	for i := 0; i < count; i++ {
		// generate a file or a directory
		if randInt(2) == 0 {
			name := testutil.UniqueString("d")
			populateDir(fs, path.Join(dirName, name), max/2)
		} else {
			name := testutil.UniqueString("f")
			var buf = make([]byte, max)
			if _, err := rand.Read(buf); err != nil {
				panic("could not read random data") // this should never happen
			}
			fs[path.Join(dirName, name)] = &fstest.MapFile{Data: buf}
		}
	}
}

func TestHashing(t *testing.T) {
	var (
		testFS = make(fstest.MapFS)
		ctx    = pctx.TestContext(t)
		env    = realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	)
	populateDir(testFS, ".", 8)
	populateDir(testFS, ".", 8)

	id, err := env.PachClient.FileSystemToFileset(ctx, testFS)
	require.NoError(t, err, "FileSystemToFileset")
	filesetHash, err := pjs.HashFileset(ctx, env.PachClient.FilesetClient, id)
	require.NoError(t, err, "hashing fileset")

	fsHash, err := pjs.HashFS(testFS)
	require.NoError(t, err, "hashing filesystem")

	require.Equal(t, hex.EncodeToString(filesetHash), hex.EncodeToString(fsHash), "fileset and filesystem hashes should be identical %v got", testFS)
	testFS["*new-file*"] = &fstest.MapFile{Data: []byte("test")}
	newFSHash, err := pjs.HashFS(testFS)
	require.NoError(t, err, "hashing filesystem")
	require.NotEqual(t, hex.EncodeToString(fsHash), hex.EncodeToString(newFSHash), "altered filesystem should have different hash")
}
