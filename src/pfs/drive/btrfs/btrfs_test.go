package btrfs

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/require"
)

var (
	counter int32
)

func TestSimple(t *testing.T) {
	driver, err := NewDriver(getBtrfsRootDir(t), "drive.TestSimple")
	require.NoError(t, err)
	shards := make(map[uint64]bool)
	shards[0] = true
	repo := &pfs.Repo{Name: "drive.TestSimple"}
	require.NoError(t, driver.CreateRepo(repo))
	commit1 := &pfs.Commit{
		Repo: repo,
		Id:   "commit1",
	}
	require.NoError(t, driver.StartCommit(nil, commit1, shards))
	file1 := &pfs.File{
		Commit: commit1,
		Path:   "foo",
	}
	require.NoError(t, driver.PutFile(file1, 0, 0, strings.NewReader("foo")))
	require.NoError(t, driver.FinishCommit(commit1, shards))
	reader, err := driver.GetFile(file1, 0)
	require.NoError(t, err)
	contents, err := ioutil.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, string(contents), "foo")
	commit2 := &pfs.Commit{
		Repo: repo,
		Id:   "commit2",
	}
	require.NoError(t, driver.StartCommit(commit1, commit2, shards))
	file2 := &pfs.File{
		Commit: commit2,
		Path:   "bar",
	}
	require.NoError(t, driver.PutFile(file2, 0, 0, strings.NewReader("bar")))
	require.NoError(t, driver.FinishCommit(commit2, shards))
	changes, err := driver.ListChange(file2, commit1, 0)
	require.NoError(t, err)
	require.Equal(t, len(changes), 1)
	require.Equal(t, changes[0].File, file2)
	require.Equal(t, changes[0].OffsetBytes, uint64(0))
	require.Equal(t, changes[0].SizeBytes, uint64(3))
}

func getBtrfsRootDir(tb testing.TB) string {
	// TODO(pedge)
	rootDir := os.Getenv("PFS_DRIVER_ROOT")
	if rootDir == "" {
		tb.Fatal("PFS_DRIVER_ROOT not set")
	}
	return rootDir
}
