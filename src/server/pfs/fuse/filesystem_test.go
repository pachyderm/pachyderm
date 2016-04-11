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
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			t.Errorf("cannot remove tempdir: %v", err)
		}
	}()

	// closed on successful termination
	quit := make(chan struct{})
	defer close(quit)
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("cannot listen: %v", err)
	}
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
	if err != nil {
		t.Fatalf("NewLocalBlockAPIServer: %v", err)
	}
	pfsclient.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}

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
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	apiClient := pfsclient.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	if err := os.Mkdir(mountpoint, 0700); err != nil {
		t.Fatalf("mkdir mountpoint: %v", err)
	}

	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mounter.Mount(mountpoint, nil, nil, ready); err != nil {
			t.Errorf("mount and serve: %v", err)
		}
	}()

	<-ready
	defer func() {
		if err := mounter.Unmount(mountpoint); err != nil {
			t.Errorf("unmount: %v", err)
		}
	}()

	if err := pfsclient.CreateRepo(apiClient, "one"); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}
	if err := pfsclient.CreateRepo(apiClient, "two"); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	if err := fstestutil.CheckDir(mountpoint, map[string]fstestutil.FileInfoCheck{
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
	}); err != nil {
		t.Errorf("wrong directory content: %v", err)
	}
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
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			t.Errorf("cannot remove tempdir: %v", err)
		}
	}()

	// closed on successful termination
	quit := make(chan struct{})
	defer close(quit)
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("cannot listen: %v", err)
	}
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
	if err != nil {
		t.Fatalf("NewLocalBlockAPIServer: %v", err)
	}
	pfsclient.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}

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
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	apiClient := pfsclient.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	if err := os.Mkdir(mountpoint, 0700); err != nil {
		t.Fatalf("mkdir mountpoint: %v", err)
	}

	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mounter.Mount(mountpoint, nil, nil, ready); err != nil {
			t.Errorf("mount and serve: %v", err)
		}
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
	if err := pfsclient.CreateRepo(apiClient, repoName); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}
	commitA, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	if err != nil {
		t.Fatalf("StartCommit: %v", err)
	}
	if err := pfsclient.FinishCommit(apiClient, repoName, commitA.ID); err != nil {
		t.Fatalf("FinishCommit: %v", err)
	}
	t.Logf("finished commit %v", commitA.ID)

	commitB, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	if err != nil {
		t.Fatalf("StartCommit: %v", err)
	}
	t.Logf("open commit %v", commitB.ID)

	if err := fstestutil.CheckDir(filepath.Join(mountpoint, repoName), map[string]fstestutil.FileInfoCheck{
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
	}); err != nil {
		t.Errorf("wrong directory content: %v", err)
	}
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
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			t.Errorf("cannot remove tempdir: %v", err)
		}
	}()

	// closed on successful termination
	quit := make(chan struct{})
	defer close(quit)
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("cannot listen: %v", err)
	}
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
	if err != nil {
		t.Fatalf("NewLocalBlockAPIServer: %v", err)
	}
	pfsclient.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}

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
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	apiClient := pfsclient.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	if err := os.Mkdir(mountpoint, 0700); err != nil {
		t.Fatalf("mkdir mountpoint: %v", err)
	}

	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mounter.Mount(mountpoint, nil, nil, ready); err != nil {
			t.Errorf("mount and serve: %v", err)
		}
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
	if err := pfsclient.CreateRepo(apiClient, repoName); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}
	commit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	if err != nil {
		t.Fatalf("StartCommit: %v", err)
	}
	t.Logf("open commit %v", commit.ID)

	const (
		greetingName = "greeting"
		greeting     = "Hello, world\n"
		greetingPerm = 0644
	)
	if err := ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.ID, greetingName), []byte(greeting), greetingPerm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	const (
		scriptName = "script"
		script     = "#!/bin/sh\necho foo\n"
		scriptPerm = 0750
	)
	if err := ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.ID, scriptName), []byte(script), scriptPerm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// open mounts look empty, so that mappers cannot accidentally use
	// them to communicate in an unreliable fashion
	if err := fstestutil.CheckDir(filepath.Join(mountpoint, repoName, commit.ID), nil); err != nil {
		t.Errorf("wrong directory content: %v", err)
	}
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
		if err := os.RemoveAll(tmp); err != nil {
			t.Errorf("cannot remove tempdir: %v", err)
		}
	}()

	// closed on successful termination
	quit := make(chan struct{})
	defer close(quit)
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("cannot listen: %v", err)
	}
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
	if err != nil {
		t.Fatalf("NewLocalBlockAPIServer: %v", err)
	}
	pfsclient.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}

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
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	apiClient := pfsclient.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	if err := os.Mkdir(mountpoint, 0700); err != nil {
		t.Fatalf("mkdir mountpoint: %v", err)
	}

	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mounter.Mount(mountpoint, nil, nil, ready); err != nil {
			t.Errorf("mount and serve: %v", err)
		}
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
	if err := pfsclient.CreateRepo(apiClient, repoName); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}
	commit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	if err != nil {
		t.Fatalf("StartCommit: %v", err)
	}
	t.Logf("open commit %v", commit.ID)

	const (
		greetingName = "greeting"
		greeting     = "Hello, world\n"
		greetingPerm = 0644
	)
	if err := ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.ID, greetingName), []byte(greeting), greetingPerm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	const (
		scriptName = "script"
		script     = "#!/bin/sh\necho foo\n"
		scriptPerm = 0750
	)
	if err := ioutil.WriteFile(filepath.Join(mountpoint, repoName, commit.ID, scriptName), []byte(script), scriptPerm); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := pfsclient.FinishCommit(apiClient, repoName, commit.ID); err != nil {
		t.Fatalf("FinishCommit: %v", err)
	}

	if err := fstestutil.CheckDir(filepath.Join(mountpoint, repoName, commit.ID), map[string]fstestutil.FileInfoCheck{
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
	}); err != nil {
		t.Errorf("wrong directory content: %v", err)
	}
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
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			t.Errorf("cannot remove tempdir: %v", err)
		}
	}()

	// closed on successful termination
	quit := make(chan struct{})
	defer close(quit)
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("cannot listen: %v", err)
	}
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
	if err != nil {
		t.Fatalf("NewLocalBlockAPIServer: %v", err)
	}
	pfsclient.RegisterBlockAPIServer(srv, blockServer)

	driver, err := drive.NewDriver(localAddress)
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}

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
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	apiClient := pfsclient.NewAPIClient(clientConn)
	mounter := fuse.NewMounter(localAddress, apiClient)

	mountpoint := filepath.Join(tmp, "mnt")
	if err := os.Mkdir(mountpoint, 0700); err != nil {
		t.Fatalf("mkdir mountpoint: %v", err)
	}

	ready := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mounter.Mount(mountpoint, nil, nil, ready); err != nil {
			t.Errorf("mount and serve: %v", err)
		}
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
	if err := pfsclient.CreateRepo(apiClient, repoName); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}
	commit, err := pfsclient.StartCommit(apiClient, repoName, "", "")
	if err != nil {
		t.Fatalf("StartCommit: %v", err)
	}

	const (
		greeting = "Hello, world\n"
	)
	filePath := filepath.Join(mountpoint, repoName, commit.ID, "greeting")

	if err := ioutil.WriteFile(filePath, []byte(greeting), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := ioutil.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, nil, data)
}
