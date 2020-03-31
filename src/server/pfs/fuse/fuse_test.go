package fuse

import (
	"crypto/sha256"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/workload"
)

const (
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)

func TestBasic(t *testing.T) {
	c := server.GetPachClient(t, server.GetBasicConfig())
	require.NoError(t, c.CreateRepo("repo"))
	_, err := c.PutFile("repo", "master", "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	_, err = c.PutFile("repo", "master", "dir/file2", strings.NewReader("foo"))
	require.NoError(t, err)
	withMount(t, c, nil, func(mountPoint string) {
		repos, err := ioutil.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(repos))
		require.Equal(t, "repo", filepath.Base(repos[0].Name()))

		files, err := ioutil.ReadDir(filepath.Join(mountPoint, "repo"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		require.Equal(t, "dir", filepath.Base(files[0].Name()))

		files, err = ioutil.ReadDir(filepath.Join(mountPoint, "repo", "dir"))
		require.NoError(t, err)
		require.Equal(t, 2, len(files))
		require.Equal(t, "file1", filepath.Base(files[0].Name()))
		require.Equal(t, "file2", filepath.Base(files[1].Name()))

		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "dir", "file1"))
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))
	})
}

func TestChunkSize(t *testing.T) {
	c := server.GetPachClient(t, server.GetBasicConfig())
	require.NoError(t, c.CreateRepo("repo"))
	_, err := c.PutFile("repo", "master", "file", strings.NewReader(strings.Repeat("p", int(pfs.ChunkSize))))
	require.NoError(t, err)
	withMount(t, c, nil, func(mountPoint string) {
		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "file"))
		require.NoError(t, err)
		require.Equal(t, int(pfs.ChunkSize), len(data))
	})
}

func TestLargeFile(t *testing.T) {
	c := server.GetPachClient(t, server.GetBasicConfig())
	require.NoError(t, c.CreateRepo("repo"))
	src := workload.RandString(rand.New(rand.NewSource(123)), GB+17)
	_, err := c.PutFile("repo", "master", "file", strings.NewReader(src))
	require.NoError(t, err)
	withMount(t, c, nil, func(mountPoint string) {
		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "file"))
		require.NoError(t, err)
		require.Equal(t, sha256.Sum256([]byte(src)), sha256.Sum256(data))
	})
}

func BenchmarkLargeFile(b *testing.B) {
	c := server.GetPachClient(b, server.GetBasicConfig())
	require.NoError(b, c.CreateRepo("repo"))
	src := workload.RandString(rand.New(rand.NewSource(123)), GB)
	_, err := c.PutFile("repo", "master", "file", strings.NewReader(src))
	require.NoError(b, err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		withMount(b, c, nil, func(mountPoint string) {
			data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "file"))
			require.NoError(b, err)
			require.Equal(b, GB, len(data))
			b.SetBytes(GB)
		})
	}
}

func TestSeek(t *testing.T) {
	c := server.GetPachClient(t, server.GetBasicConfig())
	require.NoError(t, c.CreateRepo("repo"))
	data := strings.Repeat("foo", MB)
	_, err := c.PutFile("repo", "master", "file", strings.NewReader(data))
	require.NoError(t, err)
	withMount(t, c, nil, func(mountPoint string) {
		f, err := os.Open(filepath.Join(mountPoint, "repo", "file"))
		require.NoError(t, err)
		defer func() {
			require.NoError(t, f.Close())
		}()

		testSeek := func(offset int64) {
			_, err = f.Seek(offset, 0)
			require.NoError(t, err)
			d, err := ioutil.ReadAll(f)
			require.NoError(t, err)
			require.Equal(t, data[offset:], string(d))
		}

		testSeek(0)
		testSeek(MB)
		testSeek(2 * MB)
		testSeek(3 * MB)
	})
}

func TestHeadlessBranch(t *testing.T) {
	c := server.GetPachClient(t, server.GetBasicConfig())
	require.NoError(t, c.CreateRepo("repo"))
	require.NoError(t, c.CreateBranch("repo", "master", "", nil))
	withMount(t, c, nil, func(mountPoint string) {
		fis, err := ioutil.ReadDir(filepath.Join(mountPoint, "repo"))
		require.NoError(t, err)
		// Headless branches display with 0 files.
		require.Equal(t, 0, len(fis))
	})
}

func withMount(tb testing.TB, c *client.APIClient, commits map[string]string, f func(mountPoint string)) {
	dir, err := ioutil.TempDir("", "pfs")
	require.NoError(tb, err)
	defer os.RemoveAll(dir)
	opts := &Options{
		Commits: commits,
		Unmount: make(chan struct{}),
	}
	defer func() {
		close(opts.Unmount)
	}()
	go func() {
		Mount(c, dir, opts)
	}()
	// Gotta give the fuse mount time to come up.
	time.Sleep(2 * time.Second)
	f(dir)
}
