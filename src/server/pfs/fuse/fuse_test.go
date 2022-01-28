package fuse

import (
	"bytes"
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil/random"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)

func TestBasic(t *testing.T) {
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	commit := client.NewCommit("repo", "master", "")
	err := env.PachClient.PutFile(commit, "dir/file1", strings.NewReader("foo"))
	require.NoError(t, err)
	err = env.PachClient.PutFile(commit, "dir/file2", strings.NewReader("foo"))
	require.NoError(t, err)
	withMount(t, env.PachClient, nil, func(mountPoint string) {
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	err := env.PachClient.PutFile(client.NewCommit("repo", "master", ""), "file", strings.NewReader(strings.Repeat("p", int(pfs.ChunkSize))))
	require.NoError(t, err)
	withMount(t, env.PachClient, nil, func(mountPoint string) {
		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "file"))
		require.NoError(t, err)
		require.Equal(t, int(pfs.ChunkSize), len(data))
	})
}

func TestLargeFile(t *testing.T) {
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	random.SeedRand(123)
	src := random.String(GB + 17)
	err := env.PachClient.PutFile(client.NewCommit("repo", "master", ""), "file", strings.NewReader(src))
	require.NoError(t, err)
	withMount(t, env.PachClient, nil, func(mountPoint string) {
		data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "file"))
		require.NoError(t, err)
		require.Equal(t, sha256.Sum256([]byte(src)), sha256.Sum256(data))
	})
}

func BenchmarkLargeFile(b *testing.B) {
	env := testpachd.NewRealEnv(b, dockertestenv.NewTestDBConfig(b))
	require.NoError(b, env.PachClient.CreateRepo("repo"))
	random.SeedRand(123)
	src := random.String(GB)
	err := env.PachClient.PutFile(client.NewCommit("repo", "master", ""), "file", strings.NewReader(src))
	require.NoError(b, err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		withMount(b, env.PachClient, nil, func(mountPoint string) {
			data, err := ioutil.ReadFile(filepath.Join(mountPoint, "repo", "file"))
			require.NoError(b, err)
			require.Equal(b, GB, len(data))
			b.SetBytes(GB)
		})
	}
}

func TestSeek(t *testing.T) {
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	data := strings.Repeat("foo", MB)
	err := env.PachClient.PutFile(client.NewCommit("repo", "master", ""), "file", strings.NewReader(data))
	require.NoError(t, err)
	withMount(t, env.PachClient, nil, func(mountPoint string) {
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	require.NoError(t, env.PachClient.CreateBranch("repo", "master", "", "", nil))
	withMount(t, env.PachClient, nil, func(mountPoint string) {
		fis, err := ioutil.ReadDir(filepath.Join(mountPoint, "repo"))
		require.NoError(t, err)
		// Headless branches display with 0 files.
		require.Equal(t, 0, len(fis))
	})
}

func TestReadOnly(t *testing.T) {
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	withMount(t, env.PachClient, &Options{
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo"))
	commit := client.NewCommit("repo", "master", "")
	// First, create a file
	withMount(t, env.PachClient, &Options{
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
	require.NoError(t, env.PachClient.GetFile(commit, "dir/foo", &b))
	require.Equal(t, "foo\n", b.String())

	// Now append to the file
	withMount(t, env.PachClient, &Options{
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
	require.NoError(t, env.PachClient.GetFile(commit, "dir/foo", &b))
	require.Equal(t, "foo\nfoo\n", b.String())

	// Now overwrite that file
	withMount(t, env.PachClient, &Options{
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
	require.NoError(t, env.PachClient.GetFile(commit, "dir/foo", &b))
	require.Equal(t, "bar\n", b.String())

	// Now link it to another location
	withMount(t, env.PachClient, &Options{
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
	require.NoError(t, env.PachClient.GetFile(commit, "dir/bar", &b))
	require.Equal(t, "bar\n", b.String())
	b.Reset()
	require.NoError(t, env.PachClient.GetFile(commit, "dir/buzz", &b))
	require.Equal(t, "bar\n", b.String())

	// Now delete it
	withMount(t, env.PachClient, &Options{
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
	require.YesError(t, env.PachClient.GetFile(commit, "dir/foo", &b))

	// Try writing to two repos at once
	require.NoError(t, env.PachClient.CreateRepo("repo2"))
	withMount(t, env.PachClient, &Options{
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
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("repo1"))
	require.NoError(t, env.PachClient.CreateRepo("repo2"))
	require.NoError(t, env.PachClient.CreateRepo("repo3"))
	err := env.PachClient.PutFile(client.NewCommit("repo1", "master", ""), "foo", strings.NewReader("foo\n"))
	require.NoError(t, err)
	withMount(t, env.PachClient, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		RepoOptions: map[string]*RepoOptions{
			"repo1": {Name: "repo1", Repo: "repo1"},
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
	withMount(t, env.PachClient, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		RepoOptions: map[string]*RepoOptions{
			"repo1": {Name: "repo1", Repo: "repo1", Write: true},
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
	stagingCommit := client.NewCommit("repo1", "staging", "")
	err = env.PachClient.PutFile(stagingCommit, "buzz", strings.NewReader("buzz\n"))
	require.NoError(t, err)
	withMount(t, env.PachClient, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		RepoOptions: map[string]*RepoOptions{
			"repo1": {Name: "repo1", Repo: "repo1", Branch: "staging", Write: true},
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
	require.NoError(t, env.PachClient.GetFile(stagingCommit, "fizz", &b))
	require.Equal(t, "fizz\n", b.String())
}

func TestOpenCommit(t *testing.T) {
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	require.NoError(t, env.PachClient.CreateRepo("in"))
	require.NoError(t, env.PachClient.CreateRepo("out"))
	require.NoError(t, env.PachClient.CreateBranch("out", "master", "", "", []*pfs.Branch{client.NewBranch("in", "master")}))
	_, err := env.PachClient.StartCommit("in", "master")
	require.NoError(t, err)

	withMount(t, env.PachClient, &Options{
		Fuse: &fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
		Write: true,
	}, func(mountPoint string) {
		repos, err := ioutil.ReadDir(mountPoint)
		require.NoError(t, err)
		require.Equal(t, 2, len(repos))
		files, err := ioutil.ReadDir(filepath.Join(mountPoint, "in"))
		require.NoError(t, err)
		require.Equal(t, 0, len(files))
		files, err = ioutil.ReadDir(filepath.Join(mountPoint, "out"))
		require.NoError(t, err)
		require.Equal(t, 0, len(files))
	})
}

func withMount(tb testing.TB, c *client.APIClient, opts *Options, f func(mountPoint string)) {
	dir := tb.TempDir()
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
