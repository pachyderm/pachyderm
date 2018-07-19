package fuse

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/stretchr/testify/require"
)

const (
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)

func TestBasic(t *testing.T) {
	c := server.GetPachClient(t)
	require.NoError(t, c.CreateRepo("repo"))
	_, err := c.PutFile("repo", "master", "file", strings.NewReader("foo"))
	require.NoError(t, err)
	mount(t, c, nil, func(mountPoint string) {
		repos, err := ioutil.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(repos))
		require.Equal(t, "repo", filepath.Base(repos[0].Name()))

		files, err := ioutil.ReadDir(filepath.Join(mountPoint, "repo"))
		require.NoError(t, err)
		require.Equal(t, 1, len(files))
		require.Equal(t, "file", filepath.Base(files[0].Name()))

		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "file"))
		require.NoError(t, err)
		require.Equal(t, "foo", string(data))
	})
}

func TestLargeFile(t *testing.T) {
	c := server.GetPachClient(t)
	require.NoError(t, c.CreateRepo("repo"))
	_, err := c.PutFile("repo", "master", "file", strings.NewReader(strings.Repeat("p", GB)))
	require.NoError(t, err)
	mount(t, c, nil, func(mountPoint string) {
		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "file"))
		require.NoError(t, err)
		require.Equal(t, GB, len(data))
	})
}

func TestSeek(t *testing.T) {
	c := server.GetPachClient(t)
	require.NoError(t, c.CreateRepo("repo"))
	data := strings.Repeat("foo", MB)
	_, err := c.PutFile("repo", "master", "file", strings.NewReader(data))
	require.NoError(t, err)
	mount(t, c, nil, func(mountPoint string) {
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

func mount(t *testing.T, c *client.APIClient, commits map[string]string, f func(mountPoint string)) {
	dir, err := ioutil.TempDir("", "pfs")
	require.NoError(t, err)
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
