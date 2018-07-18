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
