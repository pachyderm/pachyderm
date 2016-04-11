package fuse_test

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"bazil.org/fuse/fs/fstestutil"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/server/pfs/server"
	"google.golang.org/grpc"
)

func TestRootReadDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped because of short mode")
	}

	t.Parallel()

	// don't leave goroutines running
	var wg sync.WaitGroup
	defer wg.Wait()

	tmp, err := ioutil.TempDir("", "pachyderm-test-")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmp))
	}()

	// closed on successful termination
	quit := make(chan struct{})
	defer close(quit)
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	// TODO try to share more of this setup code with various main
	// functions
	localAddress := listener.Addr().String()
	srv := grpc.NewServer()
	const (
		numShards = 1
	)
	sharder := shard.NewLocalSharder([]string{localAddress}, numShards)
	hasher := pfsserver.NewHasher(numShards, 1)
	router := shard.NewRouter(
		sharder,
		grpcutil.NewDialer(
			grpc.WithInsecure(),
		),
		localAddress,
	)

	blockDir := filepath.Join(tmp, "blocks")
	blockServer, err := server.NewLocalBlockAPIServer(blockDir)
	require.NoError(t, err)
	pfsclient.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	require.NoError(t, err)

	apiServer := server.NewAPIServer(
		hasher,
		router,
	)
	pfsclient.RegisterAPIServer(srv, apiServer)

	internalAPIServer := server.NewInternalAPIServer(
		hasher,
		router,
		driver,
	)
	pfsclient.RegisterInternalAPIServer(srv, internalAPIServer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Serve(listener); err != nil {
			select {
			case <-quit:
				// orderly shutdown
				return
			default:
				t.Errorf("grpc serve: %v", err)
			}
		}
	}()

	clientConn, err := grpc.Dial(localAddress, grpc.WithInsecure())
	require.NoError(t, err)
	apiClient := pfsclient.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	require.NoError(t, os.Mkdir(mountpoint, 0700))
	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, mounter.Mount(mountpoint, nil, nil, ready))
	}()

	<-ready
	defer func() {
		require.NoError(t, mounter.Unmount(mountpoint))
	}()

	require.NoError(t, pfsclient.CreateRepo(apiClient, "one"))
	require.NoError(t, pfsclient.CreateRepo(apiClient, "two"))

	require.NoError(t, fstestutil.CheckDir(mountpoint, map[string]fstestutil.FileInfoCheck{
		"one": func(fi os.FileInfo) error {
			if g, e := fi.Mode(), os.ModeDir|0555; g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			// TODO show repoSize in repo stat?
			if g, e := fi.Size(), int64(0); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show RepoInfo.Created as time
			// if g, e := fi.ModTime().UTC(), repoModTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
		"two": func(fi os.FileInfo) error {
			if g, e := fi.Mode(), os.ModeDir|0555; g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			// TODO show repoSize in repo stat?
			if g, e := fi.Size(), int64(0); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show RepoInfo.Created as time
			// if g, e := fi.ModTime().UTC(), repoModTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
	}))
}

func TestRepoReadDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped because of short mode")
	}

	t.Parallel()

	// don't leave goroutines running
	var wg sync.WaitGroup
	defer wg.Wait()

	tmp, err := ioutil.TempDir("", "pachyderm-test-")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmp))
	}()

	// closed on successful termination
	quit := make(chan struct{})
	defer close(quit)
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	// TODO try to share more of this setup code with various main
	// functions
	localAddress := listener.Addr().String()
	srv := grpc.NewServer()
	const (
		numShards = 1
	)
	sharder := shard.NewLocalSharder([]string{localAddress}, numShards)
	hasher := pfsserver.NewHasher(numShards, 1)
	router := shard.NewRouter(
		sharder,
		grpcutil.NewDialer(
			grpc.WithInsecure(),
		),
		localAddress,
	)

	blockDir := filepath.Join(tmp, "blocks")
	blockServer, err := server.NewLocalBlockAPIServer(blockDir)
	require.NoError(t, err)
	pfsclient.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	require.NoError(t, err)

	apiServer := server.NewAPIServer(
		hasher,
		router,
	)
	pfsclient.RegisterAPIServer(srv, apiServer)

	internalAPIServer := server.NewInternalAPIServer(
		hasher,
		router,
		driver,
	)
	pfsclient.RegisterInternalAPIServer(srv, internalAPIServer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Serve(listener); err != nil {
			select {
			case <-quit:
				// orderly shutdown
				return
			default:
				t.Errorf("grpc serve: %v", err)
			}
		}
	}()

	clientConn, err := grpc.Dial(localAddress, grpc.WithInsecure())
	require.NoError(t, err)
	apiClient := pfsclient.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	require.NoError(t, os.Mkdir(mountpoint, 0700))

	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, mounter.Mount(mountpoint, nil, nil, ready))
	}()

	<-ready
	defer func() {
		require.NoError(t, mounter.Unmount(mountpoint))
	}()

	const (
		repoName = "foo"
	)
	require.NoError(t, pfsclient.CreateRepo(apiClient, repoName))
	commitA, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	require.NoError(t, err)
	require.NoError(t, pfsclient.FinishCommit(apiClient, repoName, commitA.ID))
	t.Logf("finished commit %v", commitA.ID)

	commitB, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	require.NoError(t, err)
	t.Logf("open commit %v", commitB.ID)

	require.NoError(t, fstestutil.CheckDir(filepath.Join(mountpoint, repoName), map[string]fstestutil.FileInfoCheck{
		commitA.ID: func(fi os.FileInfo) error {
			if g, e := fi.Mode(), os.ModeDir|0555; g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			// TODO show commitSize in commit stat?
			if g, e := fi.Size(), int64(0); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show CommitInfo.StartTime as ctime, CommitInfo.Finished as mtime
			// TODO test ctime via .Sys
			// if g, e := fi.ModTime().UTC(), commitFinishTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
		commitB.ID: func(fi os.FileInfo) error {
			if g, e := fi.Mode(), os.ModeDir|0775; g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			// TODO show commitSize in commit stat?
			if g, e := fi.Size(), int64(0); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show CommitInfo.StartTime as ctime, ??? as mtime
			// TODO test ctime via .Sys
			// if g, e := fi.ModTime().UTC(), commitFinishTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
	}))
}

func TestCommitOpenReadDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped because of short mode")
	}

	t.Parallel()

	// don't leave goroutines running
	var wg sync.WaitGroup
	defer wg.Wait()

	tmp, err := ioutil.TempDir("", "pachyderm-test-")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmp))
	}()

	// closed on successful termination
	quit := make(chan struct{})
	defer close(quit)
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	// TODO try to share more of this setup code with various main
	// functions
	localAddress := listener.Addr().String()
	srv := grpc.NewServer()
	const (
		numShards = 1
	)
	sharder := shard.NewLocalSharder([]string{localAddress}, numShards)
	hasher := pfsserver.NewHasher(numShards, 1)
	router := shard.NewRouter(
		sharder,
		grpcutil.NewDialer(
			grpc.WithInsecure(),
		),
		localAddress,
	)

	blockDir := filepath.Join(tmp, "blocks")
	blockServer, err := server.NewLocalBlockAPIServer(blockDir)
	require.NoError(t, err)
	pfsclient.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	require.NoError(t, err)

	apiServer := server.NewAPIServer(
		hasher,
		router,
	)
	pfsclient.RegisterAPIServer(srv, apiServer)

	internalAPIServer := server.NewInternalAPIServer(
		hasher,
		router,
		driver,
	)
	pfsclient.RegisterInternalAPIServer(srv, internalAPIServer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Serve(listener); err != nil {
			select {
			case <-quit:
				// orderly shutdown
				return
			default:
				t.Errorf("grpc serve: %v", err)
			}
		}
	}()

	clientConn, err := grpc.Dial(localAddress, grpc.WithInsecure())
	require.NoError(t, err)
	apiClient := pfsclient.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	require.NoError(t, os.Mkdir(mountpoint, 0700))

	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, mounter.Mount(mountpoint, nil, nil, ready))
	}()

	<-ready
	defer func() {
		require.NoError(t, mounter.Unmount(mountpoint))
	}()

	const (
		repoName = "foo"
	)
	require.NoError(t, pfsclient.CreateRepo(apiClient, repoName))
	commit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	require.NoError(t, err)
	t.Logf("open commit %v", commit.ID)

	const (
		greetingName = "greeting"
		greeting     = "Hello, world\n"
		greetingPerm = 0644
	)
	require.NoError(t, ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.ID, greetingName), []byte(greeting), greetingPerm))
	const (
		scriptName = "script"
		script     = "#!/bin/sh\necho foo\n"
		scriptPerm = 0750
	)
	require.NoError(t, ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.ID, scriptName), []byte(script), scriptPerm))

	// open mounts look empty, so that mappers cannot accidentally use
	// them to communicate in an unreliable fashion
	require.NoError(t, fstestutil.CheckDir(filepath.Join(mountpoint, repoName, commit.ID), nil))
}

func TestCommitFinishedReadDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped because of short mode")
	}

	t.Parallel()

	// don't leave goroutines running
	var wg sync.WaitGroup
	defer wg.Wait()

	tmp, err := ioutil.TempDir("", "pachyderm-test-")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	defer func() {
		require.NoError(t, os.RemoveAll(tmp))
	}()

	// closed on successful termination
	quit := make(chan struct{})
	defer close(quit)
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	// TODO try to share more of this setup code with various main
	// functions
	localAddress := listener.Addr().String()
	srv := grpc.NewServer()
	const (
		numShards = 1
	)

	sharder := shard.NewLocalSharder([]string{localAddress}, numShards)
	hasher := pfsserver.NewHasher(numShards, 1)
	router := shard.NewRouter(
		sharder,
		grpcutil.NewDialer(
			grpc.WithInsecure(),
		),
		localAddress,
	)

	blockDir := filepath.Join(tmp, "blocks")
	blockServer, err := server.NewLocalBlockAPIServer(blockDir)
	require.NoError(t, err)
	pfsclient.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	require.NoError(t, err)

	apiServer := server.NewAPIServer(
		hasher,
		router,
	)
	pfsclient.RegisterAPIServer(srv, apiServer)

	internalAPIServer := server.NewInternalAPIServer(
		hasher,
		router,
		driver,
	)
	pfsclient.RegisterInternalAPIServer(srv, internalAPIServer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Serve(listener); err != nil {
			select {
			case <-quit:
				// orderly shutdown
				return
			default:
				t.Errorf("grpc serve: %v", err)
			}
		}
	}()

	clientConn, err := grpc.Dial(localAddress, grpc.WithInsecure())
	require.NoError(t, err)
	apiClient := pfsclient.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	require.NoError(t, os.Mkdir(mountpoint, 0700))

	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, mounter.Mount(mountpoint, nil, nil, ready))
	}()

	<-ready
	defer func() {
		require.NoError(t, mounter.Unmount(mountpoint))
	}()

	const (
		repoName = "foo"
	)
	require.NoError(t, pfsclient.CreateRepo(apiClient, repoName))
	commit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	require.NoError(t, err)
	t.Logf("open commit %v", commit.ID)

	const (
		greetingName = "greeting"
		greeting     = "Hello, world\n"
		greetingPerm = 0644
	)
	require.NoError(t, ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.ID, greetingName), []byte(greeting), greetingPerm))
	const (
		scriptName = "script"
		script     = "#!/bin/sh\necho foo\n"
		scriptPerm = 0750
	)
	require.NoError(t, ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.ID, scriptName), []byte(script), scriptPerm))
	require.NoError(t, pfsclient.FinishCommit(apiClient, repoName, commit.ID))

	require.NoError(t, fstestutil.CheckDir(filepath.Join(mountpoint, repoName, commit.ID), map[string]fstestutil.FileInfoCheck{
		greetingName: func(fi os.FileInfo) error {
			// TODO respect greetingPerm
			if g, e := fi.Mode(), os.FileMode(0666); g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			if g, e := fi.Size(), int64(len(greeting)); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show fileModTime as mtime
			// if g, e := fi.ModTime().UTC(), fileModTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
		scriptName: func(fi os.FileInfo) error {
			// TODO respect scriptPerm
			if g, e := fi.Mode(), os.FileMode(0666); g != e {
				return fmt.Errorf("wrong mode: %v != %v", g, e)
			}
			if g, e := fi.Size(), int64(len(script)); g != e {
				t.Errorf("wrong size: %v != %v", g, e)
			}
			// TODO show fileModTime as mtime
			// if g, e := fi.ModTime().UTC(), fileModTime; g != e {
			// 	t.Errorf("wrong mtime: %v != %v", g, e)
			// }
			return nil
		},
	}))
}

func TestWriteAndRead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped because of short mode")
	}

	t.Parallel()

	// don't leave goroutines running
	var wg sync.WaitGroup
	defer wg.Wait()

	tmp, err := ioutil.TempDir("", "pachyderm-test-")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tmp))
	}()

	// closed on successful termination
	quit := make(chan struct{})
	defer close(quit)
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()

	// TODO try to share more of this setup code with various main
	// functions
	localAddress := listener.Addr().String()
	srv := grpc.NewServer()
	const (
		numShards = 1
	)
	sharder := shard.NewLocalSharder([]string{localAddress}, numShards)
	hasher := pfsserver.NewHasher(numShards, 1)
	router := shard.NewRouter(
		sharder,
		grpcutil.NewDialer(
			grpc.WithInsecure(),
		),
		localAddress,
	)

	blockDir := filepath.Join(tmp, "blocks")
	blockServer, err := server.NewLocalBlockAPIServer(blockDir)
	require.NoError(t, err)
	pfsclient.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	require.NoError(t, err)

	apiServer := server.NewAPIServer(
		hasher,
		router,
	)
	pfsclient.RegisterAPIServer(srv, apiServer)

	internalAPIServer := server.NewInternalAPIServer(
		hasher,
		router,
		driver,
	)
	pfsclient.RegisterInternalAPIServer(srv, internalAPIServer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Serve(listener); err != nil {
			select {
			case <-quit:
				// orderly shutdown
				return
			default:
				t.Errorf("grpc serve: %v", err)
			}
		}
	}()

	clientConn, err := grpc.Dial(localAddress, grpc.WithInsecure())
	require.NoError(t, err)
	apiClient := pfsclient.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	require.NoError(t, os.Mkdir(mountpoint, 0700))
	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.NoError(t, mounter.Mount(mountpoint, nil, nil, ready))
	}()

	<-ready
	defer func() {
		if err := mounter.Unmount(mountpoint); err != nil {
			t.Errorf("unmount: %v", err)
		}
	}()

	const (
		repoName = "foo"
	)
	require.NoError(t, pfsclient.CreateRepo(apiClient, repoName))
	commit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	require.NoError(t, err)
	const (
		greeting = "Hello, world\n"
	)
	filePath := filepath.Join(mountpoint, repoName, commit.ID, "greeting")
	require.NoError(t, ioutil.WriteFile(filePath, []byte(greeting), 0644))
	data, err := ioutil.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, nil, data)
	require.NoError(t, pfsclient.FinishCommit(apiClient, repoName, commit.ID))
}
