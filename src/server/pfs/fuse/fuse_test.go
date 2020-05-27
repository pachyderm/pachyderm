package fuse

import (
	"bytes"
	"crypto/sha256"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

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

func TestReadOnly(t *testing.T) {
	c := server.GetPachClient(t, server.GetBasicConfig())
	require.NoError(t, c.CreateRepo("repo"))
	withMount(t, c, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
	}, func(mountPoint string) {
		require.YesError(t, ioutil.WriteFile(filepath.Join(mountPoint, "repo", "foo"), []byte("foo\n"), 0644))
	})
}

func TestWrite(t *testing.T) {
	c := server.GetPachClient(t, server.GetBasicConfig())
	require.NoError(t, c.CreateRepo("repo"))
	// First, create a file
	withMount(t, c, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		Write: true,
	}, func(mountPoint string) {
		require.NoError(t, os.MkdirAll(filepath.Join(mountPoint, "repo", "dir"), 0777))
		require.NoError(t, ioutil.WriteFile(filepath.Join(mountPoint, "repo", "dir", "foo"), []byte("foo\n"), 0644))
	})
	var b bytes.Buffer
	require.NoError(t, c.GetFile("repo", "master", "dir/foo", 0, 0, &b))
	require.Equal(t, "foo\n", b.String())

	// Now append to the file
	withMount(t, c, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		Write: true,
	}, func(mountPoint string) {
		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "dir", "foo"))
		require.NoError(t, err)
		require.Equal(t, "foo\n", string(data))
		f, err := os.OpenFile(filepath.Join(mountPoint, "repo", "dir", "foo"), os.O_WRONLY, 0600)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, f.Close())
		}()
		_, err = f.Seek(0, 2)
		require.NoError(t, err)
		_, err = f.Write([]byte("foo\n"))
		require.NoError(t, err)
	})
	b.Reset()
	require.NoError(t, c.GetFile("repo", "master", "dir/foo", 0, 0, &b))
	require.Equal(t, "foo\nfoo\n", b.String())

	// Now overwrite that file
	withMount(t, c, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		Write: true,
	}, func(mountPoint string) {
		require.NoError(t, os.Remove(filepath.Join(mountPoint, "repo", "dir", "foo")))
		require.NoError(t, ioutil.WriteFile(filepath.Join(mountPoint, "repo", "dir", "foo"), []byte("bar\n"), 0644))
	})
	b.Reset()
	require.NoError(t, c.GetFile("repo", "master", "dir/foo", 0, 0, &b))
	require.Equal(t, "bar\n", b.String())

	// Now link it to another location
	withMount(t, c, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		Write: true,
	}, func(mountPoint string) {
		require.NoError(t, os.Link(filepath.Join(mountPoint, "repo", "dir", "foo"), filepath.Join(mountPoint, "repo", "dir", "bar")))
		require.NoError(t, os.Symlink(filepath.Join(mountPoint, "repo", "dir", "foo"), filepath.Join(mountPoint, "repo", "dir", "buzz")))
	})
	b.Reset()
	require.NoError(t, c.GetFile("repo", "master", "dir/bar", 0, 0, &b))
	require.Equal(t, "bar\n", b.String())
	b.Reset()
	require.NoError(t, c.GetFile("repo", "master", "dir/buzz", 0, 0, &b))
	require.Equal(t, "bar\n", b.String())

	// Now delete it
	withMount(t, c, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		Write: true,
	}, func(mountPoint string) {
		require.NoError(t, os.Remove(filepath.Join(mountPoint, "repo", "dir", "foo")))
	})
	b.Reset()
	require.YesError(t, c.GetFile("repo", "master", "dir/foo", 0, 0, &b))

	// Try writing to two repos at once
	require.NoError(t, c.CreateRepo("repo2"))
	withMount(t, c, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		Write: true,
	}, func(mountPoint string) {
		require.NoError(t, ioutil.WriteFile(filepath.Join(mountPoint, "repo", "file"), []byte("foo\n"), 0644))
		require.NoError(t, ioutil.WriteFile(filepath.Join(mountPoint, "repo2", "file"), []byte("foo\n"), 0644))
	})
}

func TestRepoOpts(t *testing.T) {
	c := server.GetPachClient(t, server.GetBasicConfig())
	require.NoError(t, c.CreateRepo("repo1"))
	require.NoError(t, c.CreateRepo("repo2"))
	require.NoError(t, c.CreateRepo("repo3"))
	_, err := c.PutFile("repo1", "master", "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	withMount(t, c, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		RepoOptions: map[string]*RepoOptions{
			"repo1": {},
		},
	}, func(mountPoint string) {
		repos, err := ioutil.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(repos))
		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo1", "foo"))
		require.NoError(t, err)
		require.Equal(t, "foo\n", string(data))
		require.YesError(t, ioutil.WriteFile(filepath.Join(mountPoint, "repo1", "bar"), []byte("bar\n"), 0644))
	})
	withMount(t, c, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		RepoOptions: map[string]*RepoOptions{
			"repo1": {Write: true},
		},
	}, func(mountPoint string) {
		repos, err := ioutil.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(repos))
		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo1", "foo"))
		require.NoError(t, err)
		require.Equal(t, "foo\n", string(data))
		require.NoError(t, ioutil.WriteFile(filepath.Join(mountPoint, "repo1", "bar"), []byte("bar\n"), 0644))
	})

	_, err = c.PutFile("repo1", "staging", "buzz", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	withMount(t, c, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		RepoOptions: map[string]*RepoOptions{
			"repo1": {Branch: "staging", Write: true},
		},
	}, func(mountPoint string) {
		repos, err := ioutil.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 1, len(repos))
		_, err = ioutil.ReadFile(filepath.Join(mountPoint, "repo1", "foo"))
		require.YesError(t, err)
		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo1", "buzz"))
		require.NoError(t, err)
		require.Equal(t, "buzz\n", string(data))
		require.NoError(t, ioutil.WriteFile(filepath.Join(mountPoint, "repo1", "fizz"), []byte("fizz\n"), 0644))
	})
	var b bytes.Buffer
	require.NoError(t, c.GetFile("repo1", "staging", "fizz", 0, 0, &b))
	require.Equal(t, "fizz\n", b.String())
}

func withMount(tb testing.TB, c *client.APIClient, opts *Options, f func(mountPoint string)) {
	dir, err := ioutil.TempDir("", "pfs-mount")
	require.NoError(tb, err)
	defer os.RemoveAll(dir)
	if opts == nil {
		opts = &Options{}
	}
	if opts.Unmount == nil {
		opts.Unmount = make(chan struct{})
	}
	unmounted := make(chan struct{})
	var mountErr error
	defer func() {
		close(opts.Unmount)
		<-unmounted
		require.NoError(tb, mountErr)
	}()
	defer func() {
		// recover because panics leave the mount in a weird state that makes
		// it hard to rerun the tests, mostly relevent when you're iterating on
		// these tests, or the code they test.
		if r := recover(); r != nil {
			tb.Fatal(r)
		}
	}()
	go func() {
		mountErr = Mount(c, dir, opts)
		close(unmounted)
	}()
	// Gotta give the fuse mount time to come up.
	time.Sleep(2 * time.Second)
	f(dir)
}
