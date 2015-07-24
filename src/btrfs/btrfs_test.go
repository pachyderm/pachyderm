package btrfs

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/log"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	if err := CheckVersion(); err != nil {
		t.Fatal(err)
	}
}

// TestOsOps checks that reading, writing, and deletion are correct on BTRFS.
func TestOsOps(t *testing.T) {
	t.Parallel()
	writeFile(t, "foo", "foo")
	checkFile(t, "foo", "foo")
	require.NoError(t, Remove("foo"))
	checkNoExists(t, "foo")
}

// TestGit checks that the Git-style interface to BTRFS is correct.
func TestGit(t *testing.T) {
	t.Parallel()
	srcRepo := "repo_TestGit"
	// Create the repo:
	require.NoError(t, Init(srcRepo))

	// Write a file "file" and create a commit "commit1":
	writeFile(t, fmt.Sprintf("%s/master/file", srcRepo), "foo")
	if !testing.Short() {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 3)
	}
	err := Commit(srcRepo, "commit1", "master")
	require.NoError(t, err)
	checkFile(t, path.Join(srcRepo, "commit1", "file"), "foo")

	// Create a new branch "branch" from commit "commit1", and check that
	// it contains the file "file":
	require.NoError(t, Branch(srcRepo, "commit1", "branch"))
	checkFile(t, fmt.Sprintf("%s/branch/file", srcRepo), "foo")

	// Create a file "file2" in branch "branch", and commit it to
	// "commit2":
	writeFile(t, fmt.Sprintf("%s/branch/file2", srcRepo), "foo")
	if !testing.Short() {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 3)
	}
	err = Commit(srcRepo, "commit2", "branch")
	require.NoError(t, err)
	checkFile(t, path.Join(srcRepo, "commit2", "file2"), "foo")

	// Print BTRFS hierarchy data for humans:
	require.NoError(t, _log(srcRepo, "", Desc, func(r io.Reader) error {
		_, err := io.Copy(os.Stdout, r)
		return err
	}))
}

func TestNewRepoIsEmpty(t *testing.T) {
	t.Parallel()
	srcRepo := "repo_TestNewRepoIsEmpty"
	require.NoError(t, Init(srcRepo))

	// ('master' is the default branch)
	dirpath := path.Join(srcRepo, "master")
	descriptors, err := ReadDir(dirpath)
	require.NoError(t, err)
	if len(descriptors) != 1 || descriptors[0].Name() != ".meta" {
		t.Fatalf("expected empty repo")
	}
}

func TestCommitsAreReadOnly(t *testing.T) {
	t.Parallel()
	srcRepo := "repo_TestCommitsAreReadOnly"
	require.NoError(t, Init(srcRepo))

	err := Commit(srcRepo, "commit1", "master")
	require.NoError(t, err)

	_, err = Create(fmt.Sprintf("%s/commit1/file", srcRepo))
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "read-only file system") {
		t.Fatalf("expected read-only filesystem error, got %s", err)
	}
}

func TestBranchesAreReadWrite(t *testing.T) {
	t.Parallel()
	srcRepo := "repo_TestBranchesAreReadWrite"
	require.NoError(t, Init(srcRepo))

	// Make a commit
	require.NoError(t, Commit(srcRepo, "my_commit", "master"))

	// Make a branch
	require.NoError(t, Branch(srcRepo, "my_commit", "my_branch"))

	fn := fmt.Sprintf("%s/my_branch/file", srcRepo)
	writeFile(t, fn, "some content")
	checkFile(t, fn, "some content")
}

// TestReplication checks that replication is correct when using local BTRFS.
// Uses `Pull`
// This is heavier and hairier, do it last.
func TestReplication(t *testing.T) {
	t.Parallel()
	t.Skip("implement this")
}

// TestSendRecv checks the Send and Recv replication primitives.
func TestSendRecv(t *testing.T) {
	t.Parallel()
	// Create a source repo:
	srcRepo := "repo_TestSendRecv_src"
	require.NoError(t, Init(srcRepo))

	// Create a file in the source repo:
	writeFile(t, fmt.Sprintf("%s/master/myfile1", srcRepo), "foo")
	if testing.Short() {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 1)
	} else {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 10)
	}

	// Create a commit in the source repo:
	require.NoError(t, Commit(srcRepo, "mycommit1", "master"))

	// Create another file in the source repo:
	writeFile(t, fmt.Sprintf("%s/master/myfile2", srcRepo), "bar")
	if testing.Short() {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 1)
	} else {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 10)
	}

	// Create a another commit in the source repo:
	require.NoError(t, Commit(srcRepo, "mycommit2", "master"))

	// Create a destination repo:
	dstRepo := "repo_TestSendRecv_dst"
	require.NoError(t, Init(dstRepo))
	repo2Recv := func(r io.Reader) error { return recv(dstRepo, r) }

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoExists(t, fmt.Sprintf("%s/master/mycommit1", dstRepo))
	checkNoExists(t, fmt.Sprintf("%s/master/mycommit2", dstRepo))

	// Run a Send/Recv operation to fetch data from the older "mycommit1".
	// This verifies that tree copying works:
	require.NoError(t, send(srcRepo, "mycommit1", repo2Recv))

	// Check that the file from mycommit1 exists, but not from mycommit2:
	checkFile(t, fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo")
	checkNoExists(t, fmt.Sprintf("%s/mycommit2/myfile2", dstRepo))

	// Send again, this time starting from mycommit1 and going to mycommit2:
	require.NoError(t, send(srcRepo, "mycommit2", repo2Recv))

	// Verify that files from both commits are present:
	checkFile(t, fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo")
	checkFile(t, fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), "bar")
}

// TestBranchesAreNotReplicated // this is a known property, but not desirable long term
// TestCommitsAreReplicated // Uses Send and Recv
func TestCommitsAreReplicated(t *testing.T) {
	t.Parallel()
	// Create a source repo:
	srcRepo := "repo_TestCommitsAreReplicated_src"
	require.NoError(t, Init(srcRepo))

	// Create a file in the source repo:
	writeFile(t, fmt.Sprintf("%s/master/myfile1", srcRepo), "foo")
	if testing.Short() {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 1)
	} else {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 10)
	}

	// Create a commit in the source repo:
	require.NoError(t, Commit(srcRepo, "mycommit1", "master"))

	// Create another file in the source repo:
	writeFile(t, fmt.Sprintf("%s/master/myfile2", srcRepo), "bar")
	if testing.Short() {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 1)
	} else {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 10)
	}

	// Create a another commit in the source repo:
	require.NoError(t, Commit(srcRepo, "mycommit2", "master"))

	// Create a destination repo:
	dstRepo := "repo_TestCommitsAreReplicated_dst"
	require.NoError(t, Init(dstRepo))

	// Verify that the commits "mycommit1" and "mycommit2" do exist in source:
	checkFile(t, fmt.Sprintf("%s/mycommit1/myfile1", srcRepo), "foo")
	checkFile(t, fmt.Sprintf("%s/mycommit2/myfile2", srcRepo), "bar")

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoExists(t, fmt.Sprintf("%s/mycommit1", dstRepo))
	checkNoExists(t, fmt.Sprintf("%s/mycommit2", dstRepo))

	// Run a Pull/Recv operation to fetch all commits:
	err := Pull(srcRepo, "", NewLocalReplica(dstRepo))
	require.NoError(t, err)

	// Verify that files from both commits are present:
	checkFile(t, fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo")
	checkFile(t, fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), "bar")

	// Now check that we can use dstRepo as the source for replication
	// Create a second dest repo:
	dstRepo2 := "repo_TestCommitsAreReplicated_dst2"
	require.NoError(t, Init(dstRepo2))

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoExists(t, fmt.Sprintf("%s/mycommit1", dstRepo2))
	checkNoExists(t, fmt.Sprintf("%s/mycommit2", dstRepo2))

	// Run a Pull/Recv operation to fetch all commits:
	err = Pull(dstRepo, "", NewLocalReplica(dstRepo2))
	require.NoError(t, err)

	// Verify that files from both commits are present:
	checkFile(t, fmt.Sprintf("%s/mycommit1/myfile1", dstRepo2), "foo")
	checkFile(t, fmt.Sprintf("%s/mycommit2/myfile2", dstRepo2), "bar")
	checkFile(t, fmt.Sprintf("%s/master/myfile1", dstRepo), "foo")
	checkFile(t, fmt.Sprintf("%s/master/myfile2", dstRepo), "bar")
}

// TestSendWithMissingIntermediateCommitIsCorrect(?) // ? means we don't know what the behavior is.
func TestSendWithMissingIntermediateCommitIsCorrect(t *testing.T) {
	t.Parallel()
	//FIXME: https://github.com/pachyderm/pachyderm/issues/60
	t.Skip("Removing commits currently breaks replication, this is ok for now because users can't remove commits.")
	// Create a source repo:
	srcRepo := "repo_TestSendWithMissingIntermediateCommitIsCorrect_src"
	require.NoError(t, Init(srcRepo))

	// Create a file in the source repo:
	writeFile(t, fmt.Sprintf("%s/master/myfile1", srcRepo), "foo")

	// Create a commit in the source repo:
	require.NoError(t, Commit(srcRepo, "mycommit1", "master"))

	// Create another file in the source repo:
	writeFile(t, fmt.Sprintf("%s/master/myfile2", srcRepo), "bar")

	// Create a another commit in the source repo:
	require.NoError(t, Commit(srcRepo, "mycommit2", "master"))

	// Delete intermediate commit "mycommit1":
	require.NoError(t, subvolumeDelete(fmt.Sprintf("%s/mycommit1", srcRepo)))

	// Verify that the commit "mycommit1" does not exist and "mycommit2" does in the source repo:
	checkNoExists(t, fmt.Sprintf("%s/mycommit1", srcRepo))
	checkFile(t, fmt.Sprintf("%s/mycommit2/myfile2", srcRepo), "bar")

	// Create a destination repo:
	dstRepo := "repo_TestSendWithMissingIntermediateCommitIsCorrect_dst"
	require.NoError(t, Init(dstRepo))

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoExists(t, fmt.Sprintf("%s/mycommit1", dstRepo))
	checkNoExists(t, fmt.Sprintf("%s/mycommit2", dstRepo))

	// Run a Pull/Recv operation to fetch all commits:
	require.NoError(t, Pull(srcRepo, "", NewLocalReplica(dstRepo)))

	// Verify that the commit "mycommit1" does not exist and "mycommit2" does in the destination repo:
	t.Skipf("TODO(jd,rw): no files were synced")
	checkNoExists(t, fmt.Sprintf("%s/mycommit1/myfile1", dstRepo))
	checkFile(t, fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), "bar")
}

// TestBranchesAreNotImplicitlyReplicated // this is a known property, but not desirable long term
func TestBranchesAreNotImplicitlyReplicated(t *testing.T) {
	t.Parallel()
	// Create a source repo:
	srcRepo := "repo_TestBranchesAreNotImplicitlyReplicated_src"
	require.NoError(t, Init(srcRepo))

	// Create a commit in the source repo:
	require.NoError(t, Commit(srcRepo, "mycommit", "master"))

	// Create a branch in the source repo:
	require.NoError(t, Branch(srcRepo, "mycommit", "mybranch"))

	// Create a destination repo:
	dstRepo := "repo_TestBranchesAreNotImplicitlyReplicated_dst"
	require.NoError(t, Init(dstRepo))

	// Run a Pull/Recv operation to fetch all commits on master:
	require.NoError(t, Pull(srcRepo, "", NewLocalReplica(dstRepo)))

	// Verify that only the commits are replicated, not branches:
	commitFilename := fmt.Sprintf("%s/mycommit", dstRepo)
	exists, err := FileExists(commitFilename)
	require.NoError(t, err)
	if !exists {
		t.Fatalf("File %s should exist.", commitFilename)
	}
	checkNoExists(t, fmt.Sprintf("%s/mybranch", dstRepo))
}

func TestS3Replica(t *testing.T) {
	t.Parallel()
	t.Skip("This test is periodically failing to reach s3.")
	// Create a source repo:
	srcRepo := "repo_TestS3Replica_src"
	require.NoError(t, Init(srcRepo))

	// Create a file in the source repo:
	writeFile(t, fmt.Sprintf("%s/master/myfile1", srcRepo), "foo")
	if testing.Short() {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 1)
	} else {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 10)
	}

	// Create a commit in the source repo:
	require.NoError(t, Commit(srcRepo, "mycommit1", "master"))

	// Create another file in the source repo:
	writeFile(t, fmt.Sprintf("%s/master/myfile2", srcRepo), "bar")
	if testing.Short() {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 1)
	} else {
		writeLots(t, fmt.Sprintf("%s/master/big_file", srcRepo), 10)
	}

	// Create a another commit in the source repo:
	require.NoError(t, Commit(srcRepo, "mycommit2", "master"))

	// Create a destination repo:
	dstRepo := "repo_TestS3Replica_dst"
	require.NoError(t, Init(dstRepo))

	// Verify that the commits "mycommit1" and "mycommit2" do exist in source:
	checkFile(t, fmt.Sprintf("%s/mycommit1/myfile1", srcRepo), "foo")
	checkFile(t, fmt.Sprintf("%s/mycommit2/myfile2", srcRepo), "bar")

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoExists(t, fmt.Sprintf("%s/mycommit1", dstRepo))
	checkNoExists(t, fmt.Sprintf("%s/mycommit2", dstRepo))

	// Run a Pull to push all commits to s3
	s3Replica := NewS3Replica(path.Join("pachyderm-test", common.NewUUID()))
	err := Pull(srcRepo, "", s3Replica)
	require.NoError(t, err)

	// Pull commits from s3 to a new local replica
	err = s3Replica.Pull("", NewLocalReplica(dstRepo))
	require.NoError(t, err)

	// Verify that files from both commits are present:
	checkFile(t, fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo")
	checkFile(t, fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), "bar")
	checkFile(t, fmt.Sprintf("%s/master/myfile1", dstRepo), "foo")
	checkFile(t, fmt.Sprintf("%s/master/myfile2", dstRepo), "bar")
}

// Test for `Commits`: check that the sort order of CommitInfo objects is structured correctly.
// Start from:
//	// Print BTRFS hierarchy data for humans:
//	check(Log("repo", "0", func(r io.Reader) error {
//		_, err := io.Copy(os.Stdout, r)
//		return err
//	}), t)

// TestFindNew, which is basically like `git diff`. Corresponds to `find-new` in btrfs.
func TestFindNew(t *testing.T) {
	t.Parallel()
	repoName := "repo_TestFindNew"
	require.NoError(t, Init(repoName))

	checkFindNew := func(want []string, repo, from, to string) {
		got, err := FindNew(repo, from, to)
		require.NoError(t, err)
		t.Logf("checkFindNew(%v, %v, %v) -> %v", repo, from, to, got)

		// handle nil and empty slice the same way:
		if len(want) == 0 && len(got) == 0 {
			return
		}

		if !reflect.DeepEqual(want, got) {
			t.Fatalf("wanted %v, got %v for FindNew(%v, %v, %v)", want, got, repo, from, to)
		}

	}

	// There are no new files upon repo creation:
	require.NoError(t, Commit(repoName, "mycommit0", "master"))
	checkFindNew([]string{}, repoName, "mycommit0", "mycommit0")

	// A new, uncommited file is returned in the list:
	writeFile(t, fmt.Sprintf("%s/master/myfile1", repoName), "foo")
	checkFindNew([]string{"myfile1"}, repoName, "mycommit0", "master")

	// When that file is commited, then it still shows up in the delta since transid0:
	require.NoError(t, Commit(repoName, "mycommit1", "master"))
	// TODO(rw, jd) Shouldn't this pass?
	checkFindNew([]string{"myfile1"}, repoName, "mycommit0", "mycommit1")

	// The file doesn't show up in the delta since the new transaction:
	checkFindNew([]string{}, repoName, "mycommit1", "mycommit1")

	// Sanity check: the old delta still gives the same result:
	checkFindNew([]string{"myfile1"}, repoName, "mycommit0", "master")
}

func TestFilenamesWithSpaces(t *testing.T) {
	t.Parallel()
	repoName := "repo_TestFilenamesWithSpaces"
	require.NoError(t, Init(repoName))

	fn := fmt.Sprintf("%s/master/my file", repoName)
	writeFile(t, fn, "some content")
	checkFile(t, fn, "some content")
}

func TestFilenamesWithSlashesFail(t *testing.T) {
	t.Parallel()
	repoName := "repo_TestFilenamesWithSlashesFail"
	require.NoError(t, Init(repoName))

	fn := fmt.Sprintf("%s/master/my/file", repoName)
	_, err := Create(fn)
	if err == nil {
		t.Fatalf("expected filename with slash to fail")
	}
}

func TestTwoSources(t *testing.T) {
	t.Parallel()
	src1 := "repo_TestTwoSources_src1"
	require.NoError(t, Init(src1))
	src2 := "repo_TestTwoSources_src2"
	require.NoError(t, Init(src2))
	dst := "repo_TestTwoSources_dst"
	require.NoError(t, Init(dst))

	// write a file to src1
	writeFile(t, fmt.Sprintf("%s/master/file1", src1), "file1")
	// commit it
	require.NoError(t, Commit(src1, "commit1", "master"))
	// push it to src2
	require.NoError(t, NewLocalReplica(src1).Pull("", NewLocalReplica(src2)))
	// push it to dst
	require.NoError(t, NewLocalReplica(src1).Pull("", NewLocalReplica(dst)))

	writeFile(t, fmt.Sprintf("%s/master/file2", src2), "file2")
	require.NoError(t, Commit(src2, "commit2", "master"))
	require.NoError(t, NewLocalReplica(src2).Pull("commit1", NewLocalReplica(dst)))

	checkFile(t, fmt.Sprintf("%s/commit1/file1", dst), "file1")
	checkFile(t, fmt.Sprintf("%s/commit2/file2", dst), "file2")
}

func TestWaitFile(t *testing.T) {
	t.Parallel()
	src := "repo_TestWaitFile"
	require.NoError(t, Init(src))
	complete := make(chan struct{})
	go func() {
		require.NoError(t, WaitFile(src+"/file", nil))
		complete <- struct{}{}
	}()
	WriteFile(src+"/file", nil)
	select {
	case <-complete:
		// we passed the test
		return
	case <-time.After(time.Second * 10):
		t.Fatal("Timeout waiting for file.")
	}
}

func TestCancelWaitFile(t *testing.T) {
	t.Parallel()
	src := "repo_TestCancelWaitFile"
	require.NoError(t, Init(src))
	complete := make(chan struct{})
	cancel := make(chan struct{})
	go func() {
		err := WaitFile(src+"/file", cancel)
		if err != ErrCancelled {
			t.Fatal("Got the wrong error. Expected ErrCancelled.")
		}
		complete <- struct{}{}
	}()
	cancel <- struct{}{}
	select {
	case <-complete:
		// we passed the test
		return
	case <-time.After(time.Second * 10):
		t.Fatal("Timeout waiting for file.")
	}
}

func TestWaitAnyFile(t *testing.T) {
	t.Parallel()
	src := "repo_TestWaitAnyFile"
	require.NoError(t, Init(src))
	complete := make(chan struct{})
	go func() {
		file, err := WaitAnyFile(src+"/file1", src+"/file2")
		log.Print("WaitedOn: ", file)
		require.NoError(t, err)
		if file != src+"/file2" {
			t.Fatal("Got the wrong file.")
		}
		complete <- struct{}{}
	}()
	WriteFile(src+"/file2", nil)
	select {
	case <-complete:
		// we passed the test
		return
	case <-time.After(time.Second * 10):
		t.Fatal("Timeout waiting for file.")
	}
}

// Case: create, delete, edit files and check that the filenames correspond to the changes ones.

// go test coverage

func checkFile(t *testing.T, name string, content string) {
	checkExists(t, name)
	file, err := Open(name)
	require.NoError(t, err)
	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	require.NoError(t, err)
	require.Equal(t, content+"\n", line)
	require.NoError(t, file.Close())
}

func checkExists(t *testing.T, name string) {
	exists, err := FileExists(name)
	require.NoError(t, err)
	require.True(t, exists)
}
func checkNoExists(t *testing.T, name string) {
	exists, err := FileExists(name)
	require.NoError(t, err)
	require.False(t, exists)
}

// writeFile quickly writes a string to disk.
func writeFile(t *testing.T, name, content string) {
	f, err := Create(name)
	require.NoError(t, err)
	f.WriteString(content + "\n")
	require.NoError(t, f.Close())
}

// writeLots writes a lots of data to disk in 128 MB files
func writeLots(t *testing.T, prefix string, nFiles int) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < nFiles; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f, err := Create(fmt.Sprintf("%s-%d", prefix, common.NewUUID()))
			require.NoError(t, err)
			_, err = io.Copy(f, io.LimitReader(rand.Reader, (1<<20)*16))
			require.NoError(t, err)
			require.NoError(t, f.Close())
		}(i)
	}
}
