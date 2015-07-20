package btrfs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pachyderm/pachyderm/src/pkg/executil"
	"github.com/stretchr/testify/require"
)

func init() {
	// TODO(pedge): needed in tests? will not be needed for golang 1.5 for sure
	runtime.GOMAXPROCS(runtime.NumCPU())
	executil.SetDebug(true)
}

func TestFFI(t *testing.T) {
	t.Parallel()
	test(t, "ffi", NewFFIAPI(), "exec", NewExecAPI())
}

func TestExec(t *testing.T) {
	t.Parallel()
	test(t, "exec", NewExecAPI(), "ffi", NewFFIAPI())
}

func test(t *testing.T, testName string, testAPI API, referenceName string, referenceAPI API) {
	rootDir := filepath.Join(getBtrfsRootDir(t), "test-btrfs", fmt.Sprintf("%s-%s", testName, referenceName))
	err := os.RemoveAll(rootDir)
	require.NoError(t, err)
	err = os.MkdirAll(rootDir, 0700)
	require.NoError(t, err)

	volume1 := filepath.Join(rootDir, "volume1")
	err = testAPI.SubvolumeCreate(volume1)
	require.NoError(t, err)
	checkDirExists(t, volume1)

	readonly, err := testAPI.PropertyGetReadonly(volume1)
	require.NoError(t, err)
	require.False(t, readonly)
	referenceReadonly, err := referenceAPI.PropertyGetReadonly(volume1)
	require.NoError(t, err)
	require.False(t, referenceReadonly)

	err = testAPI.PropertySetReadonly(volume1, true)
	require.NoError(t, err)
	readonly, err = testAPI.PropertyGetReadonly(volume1)
	require.NoError(t, err)
	require.True(t, readonly)
	referenceReadonly, err = referenceAPI.PropertyGetReadonly(volume1)
	require.NoError(t, err)
	require.True(t, referenceReadonly)

	err = testAPI.PropertySetReadonly(volume1, false)
	require.NoError(t, err)
	readonly, err = testAPI.PropertyGetReadonly(volume1)
	require.NoError(t, err)
	require.False(t, readonly)
	referenceReadonly, err = referenceAPI.PropertyGetReadonly(volume1)
	require.NoError(t, err)
	require.False(t, referenceReadonly)

	volume2 := filepath.Join(rootDir, "volume2")
	err = testAPI.SubvolumeSnapshot(volume1, volume2, true)
	require.NoError(t, err)
	readonly, err = testAPI.PropertyGetReadonly(volume2)
	require.NoError(t, err)
	require.True(t, readonly)
	referenceReadonly, err = referenceAPI.PropertyGetReadonly(volume2)
	require.NoError(t, err)
	require.True(t, referenceReadonly)

	err = testAPI.PropertySetReadonly(volume2, false)
	require.NoError(t, err)
	readonly, err = testAPI.PropertyGetReadonly(volume2)
	require.NoError(t, err)
	require.False(t, readonly)
	referenceReadonly, err = referenceAPI.PropertyGetReadonly(volume2)
	require.NoError(t, err)
	require.False(t, referenceReadonly)

	volume3 := filepath.Join(rootDir, "volume3")
	err = testAPI.SubvolumeSnapshot(volume1, volume3, false)
	require.NoError(t, err)
	readonly, err = testAPI.PropertyGetReadonly(volume3)
	require.NoError(t, err)
	require.False(t, readonly)
	referenceReadonly, err = referenceAPI.PropertyGetReadonly(volume3)
	require.NoError(t, err)
	require.False(t, referenceReadonly)

	err = testAPI.PropertySetReadonly(volume3, true)
	require.NoError(t, err)
	readonly, err = testAPI.PropertyGetReadonly(volume3)
	require.NoError(t, err)
	require.True(t, readonly)
	referenceReadonly, err = referenceAPI.PropertyGetReadonly(volume3)
	require.NoError(t, err)
	require.True(t, referenceReadonly)

	volume4 := filepath.Join(rootDir, "volume4")
	err = testAPI.SubvolumeSnapshot(volume2, volume4, false)
	require.NoError(t, err)
	readonly, err = testAPI.PropertyGetReadonly(volume4)
	require.NoError(t, err)
	require.False(t, readonly)
	referenceReadonly, err = referenceAPI.PropertyGetReadonly(volume4)
	require.NoError(t, err)
	require.False(t, referenceReadonly)

	err = testAPI.PropertySetReadonly(volume4, true)
	require.NoError(t, err)
	readonly, err = testAPI.PropertyGetReadonly(volume4)
	require.NoError(t, err)
	require.True(t, readonly)
	referenceReadonly, err = referenceAPI.PropertyGetReadonly(volume4)
	require.NoError(t, err)
	require.True(t, referenceReadonly)
}

func getBtrfsRootDir(t *testing.T) string {
	// TODO(pedge)
	rootDir := os.Getenv("PFS_BTRFS_ROOT")
	if rootDir == "" {
		t.Fatal("PFS_BTRFS_ROOT not set")
	}
	return rootDir
}

func checkDirExists(t *testing.T, path string) {
	stat, err := os.Stat(path)
	require.NoError(t, err)
	require.True(t, stat.IsDir())
}
