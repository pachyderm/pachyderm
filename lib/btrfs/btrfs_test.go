package btrfs

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
)

var run_string string

func check(err error, t *testing.T) {
	if err != nil {
		debug.PrintStack()
		t.Fatal(err)
	}
}

// writeFile quickly writes a string to disk.
func writeFile(name, content string, t *testing.T) {
	f, err := Create(name)
	check(err, t)
	f.WriteString(content + "\n")
	check(f.Close(), t)
}

var suffix int = 0

// writeLots writes a lots of data to disk in 128 MB files
func writeLots(prefix string, nFiles int, t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait()
	for i := 0; i < nFiles; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f, err := Create(fmt.Sprintf("%s-%d", prefix, suffix))
			check(err, t)
			suffix++
			_, err = io.Copy(f, io.LimitReader(rand.Reader, (1<<20)*16))
			check(err, t)
			check(f.Close(), t)
		}(i)
	}
}

// checkFile checks if a file on disk contains a given string.
func checkFile(name, content string, t *testing.T) {
	exists, err := FileExists(name)
	check(err, t)
	if !exists {
		t.Fatalf("File %s should exist.", name)
	}

	f, err := Open(name)
	if err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(f)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if line != content+"\n" {
		t.Fatal("File contained the wrong value.")
	}
	f.Close()
}

// checkNoFile checks that no file is present.
func checkNoFile(name string, t *testing.T) {
	exists, err := FileExists(name)
	check(err, t)
	if exists {
		t.Fatalf("File %s shouldn't exist.", name)
	}
}

func removeFile(name string, t *testing.T) {
	err := Remove(name)
	if err != nil {
		t.Fatal(err)
	}
}

// TestOsOps checks that reading, writing, and deletion are correct on BTRFS.
func TestOsOps(t *testing.T) {
	writeFile("foo", "foo", t)
	checkFile("foo", "foo", t)
	removeFile("foo", t)
	checkNoFile("foo", t)
}

// TestGit checks that the Git-style interface to BTRFS is correct.
func TestGit(t *testing.T) {
	srcRepo := "repo_TestGit"
	// Create the repo:
	check(Init(srcRepo), t)

	// Write a file "file" and create a commit "commit1":
	writeFile(fmt.Sprintf("%s/master/file", srcRepo), "foo", t)
	if !testing.Short() {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 3, t)
	}
	err := Commit(srcRepo, "commit1", "master")
	check(err, t)
	checkFile(path.Join(srcRepo, "commit1", "file"), "foo", t)

	// Create a new branch "branch" from commit "commit1", and check that
	// it contains the file "file":
	check(Branch(srcRepo, "commit1", "branch"), t)
	checkFile(fmt.Sprintf("%s/branch/file", srcRepo), "foo", t)

	// Create a file "file2" in branch "branch", and commit it to
	// "commit2":
	writeFile(fmt.Sprintf("%s/branch/file2", srcRepo), "foo", t)
	if !testing.Short() {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 3, t)
	}
	err = Commit(srcRepo, "commit2", "branch")
	check(err, t)
	checkFile(path.Join(srcRepo, "commit2", "file2"), "foo", t)

	// Print BTRFS hierarchy data for humans:
	check(Log(srcRepo, "t0", Desc, func(r io.Reader) error {
		_, err := io.Copy(os.Stdout, r)
		return err
	}), t)
}

func TestNewRepoIsEmpty(t *testing.T) {
	srcRepo := "repo_TestNewRepoIsEmpty"
	check(Init(srcRepo), t)

	// ('master' is the default branch)
	dirpath := path.Join(srcRepo, "master")
	descriptors, err := ReadDir(dirpath)
	check(err, t)
	if len(descriptors) != 1 || descriptors[0].Name() != ".meta" {
		t.Fatalf("expected empty repo")
	}
}

func TestCommitsAreReadOnly(t *testing.T) {
	srcRepo := "repo_TestCommitsAreReadOnly"
	check(Init(srcRepo), t)

	err := Commit(srcRepo, "commit1", "master")
	check(err, t)

	_, err = Create(fmt.Sprintf("%s/commit1/file", srcRepo))
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "read-only file system") {
		t.Fatalf("expected read-only filesystem error, got %s", err)
	}
}

func TestBranchesAreReadWrite(t *testing.T) {
	srcRepo := "repo_TestBranchesAreReadWrite"
	check(Init(srcRepo), t)

	err := Branch(srcRepo, "t0", "my_branch")
	check(err, t)

	fn := fmt.Sprintf("%s/my_branch/file", srcRepo)
	writeFile(fn, "some content", t)
	checkFile(fn, "some content", t)
}

// TestReplication checks that replication is correct when using local BTRFS.
// Uses `Pull`
// This is heavier and hairier, do it last.
func TestReplication(t *testing.T) {
	t.Skip("implement this")
}

// TestSendRecv checks the Send and Recv replication primitives.
func TestSendRecv(t *testing.T) {
	// Create a source repo:
	srcRepo := "repo_TestSendRecv_src"
	check(Init(srcRepo), t)

	// Create a file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile1", srcRepo), "foo", t)
	if testing.Short() {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 1, t)
	} else {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 10, t)
	}

	// Create a commit in the source repo:
	check(Commit(srcRepo, "mycommit1", "master"), t)

	// Create another file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile2", srcRepo), "bar", t)
	if testing.Short() {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 1, t)
	} else {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 10, t)
	}

	// Create a another commit in the source repo:
	check(Commit(srcRepo, "mycommit2", "master"), t)

	// Create a destination repo:
	dstRepo := "repo_TestSendRecv_dst"
	check(InitReplica(dstRepo), t)
	repo2Recv := func(r io.Reader) error { return Recv(dstRepo, r) }
	check(Send(srcRepo, "t0", repo2Recv), t)

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoFile(fmt.Sprintf("%s/mycommit1", dstRepo), t)
	checkNoFile(fmt.Sprintf("%s/mycommit2", dstRepo), t)

	// Run a Send/Recv operation to fetch data from the older "mycommit1".
	// This verifies that tree copying works:
	check(Send(srcRepo, "mycommit1", repo2Recv), t)

	// Check that the file from mycommit1 exists, but not from mycommit2:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo", t)
	checkNoFile(fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), t)

	// Send again, this time starting from mycommit1 and going to mycommit2:
	check(Send(srcRepo, "mycommit2", repo2Recv), t)

	// Verify that files from both commits are present:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), "bar", t)
}

// TestSendWithMissingIntermediateCommitIsCorrect(?) // ? means we don't know what the behavior is.

// TestBranchesAreNotReplicated // this is a known property, but not desirable long term
// TestCommitsAreReplicated // Uses Send and Recv
func TestCommitsAreReplicated(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	// Create a source repo:
	srcRepo := "repo_TestCommitsAreReplicated_src"
	check(Init(srcRepo), t)

	// Create a file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile1", srcRepo), "foo", t)
	if testing.Short() {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 1, t)
	} else {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 10, t)
	}

	// Create a commit in the source repo:
	check(Commit(srcRepo, "mycommit1", "master"), t)

	// Create another file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile2", srcRepo), "bar", t)
	if testing.Short() {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 1, t)
	} else {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 10, t)
	}

	// Create a another commit in the source repo:
	check(Commit(srcRepo, "mycommit2", "master"), t)

	// Create a destination repo:
	dstRepo := "repo_TestSendCommitsAreReplicated_dst"
	check(InitReplica(dstRepo), t)

	// Verify that the commits "mycommit1" and "mycommit2" do exist in source:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", srcRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/mycommit2/myfile2", srcRepo), "bar", t)

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoFile(fmt.Sprintf("%s/mycommit1", dstRepo), t)
	checkNoFile(fmt.Sprintf("%s/mycommit2", dstRepo), t)

	// Run a Pull/Recv operation to fetch all commits:
	err := Pull2(srcRepo, "", NewLocalReplica(dstRepo))
	check(err, t)

	// Verify that files from both commits are present:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), "bar", t)

	// Now check that we can use dstRepo as the source for replication
	// Create a second dest repo:
	dstRepo2 := "repo_TestSendCommitsAreReplicated_dst2"
	check(InitReplica(dstRepo2), t)

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoFile(fmt.Sprintf("%s/mycommit1", dstRepo2), t)
	checkNoFile(fmt.Sprintf("%s/mycommit2", dstRepo2), t)

	// Run a Pull/Recv operation to fetch all commits:
	err = Pull2(dstRepo, "", NewLocalReplica(dstRepo2))
	check(err, t)

	// Verify that files from both commits are present:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", dstRepo2), "foo", t)
	checkFile(fmt.Sprintf("%s/mycommit2/myfile2", dstRepo2), "bar", t)
	checkFile(fmt.Sprintf("%s/master/myfile1", dstRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/master/myfile2", dstRepo), "bar", t)
}

func TestS3Replica(t *testing.T) {
	// Create a source repo:
	srcRepo := "repo_TestS3Replica_src"
	check(Init(srcRepo), t)

	// Create a file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile1", srcRepo), "foo", t)
	if testing.Short() {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 1, t)
	} else {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 10, t)
	}

	// Create a commit in the source repo:
	check(Commit(srcRepo, "mycommit1", "master"), t)

	// Create another file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile2", srcRepo), "bar", t)
	if testing.Short() {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 1, t)
	} else {
		writeLots(fmt.Sprintf("%s/master/big_file", srcRepo), 10, t)
	}

	// Create a another commit in the source repo:
	check(Commit(srcRepo, "mycommit2", "master"), t)

	// Create a destination repo:
	dstRepo := "repo_TestS3Replica_dst"
	check(InitReplica(dstRepo), t)

	// Verify that the commits "mycommit1" and "mycommit2" do in source:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", srcRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/mycommit2/myfile2", srcRepo), "bar", t)

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoFile(fmt.Sprintf("%s/mycommit1", dstRepo), t)
	checkNoFile(fmt.Sprintf("%s/mycommit2", dstRepo), t)

	// Run a Pull/Recv operation to fetch all commits:
	s3Replica := NewS3Replica(path.Join("pachyderm-test", RandSeq(20)))
	err := Pull2(srcRepo, "", s3Replica)
	check(err, t)

	err = s3Replica.Pull("", NewLocalReplica(dstRepo))
	check(err, t)

	// Verify that files from both commits are present:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), "bar", t)
	checkFile(fmt.Sprintf("%s/master/myfile1", dstRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/master/myfile2", dstRepo), "bar", t)
}

// TestHoldRelease creates one-off commit named after a UUID, to ensure a data consumer can always access data in a commit, even if the original commit is deleted.
func TestHoldRelease(t *testing.T) {
	srcRepo := "repo_TestHoldRelease"
	check(Init(srcRepo), t)

	// Write a file "myfile" with contents "foo":
	master_fn := fmt.Sprintf("%s/master/myfile", srcRepo)
	writeFile(master_fn, "foo", t)
	checkFile(master_fn, "foo", t)

	// Create a commit "mycommit" and verify "myfile" exists:
	mycommit_fn := fmt.Sprintf("%s/mycommit/myfile", srcRepo)
	check(Commit(srcRepo, "mycommit", "master"), t)
	checkFile(mycommit_fn, "foo", t)

	// Grab a snapshot:
	snapshot_path, err := Hold(srcRepo, "mycommit")
	check(err, t)

	// Delete the commit from the snapshot.
	// (uses the lower-level btrfs command for now):
	mycommit_path := fmt.Sprintf("%s/mycommit", srcRepo)
	check(SubvolumeDelete(mycommit_path), t)

	// Verify that the commit path doesn't exist:
	checkNoFile(mycommit_path, t)

	// Verify that the file still exists in our snapshot:
	snapshot_fn := fmt.Sprintf("%s/myfile", snapshot_path)
	checkFile(snapshot_fn, "foo", t)
}

// Test for `Commits`: check that the sort order of CommitInfo objects is structured correctly.
// Start from:
//	// Print BTRFS hierarchy data for humans:
//	check(Log("repo", "0", func(r io.Reader) error {
//		_, err := io.Copy(os.Stdout, r)
//		return err
//	}), t)

// TestFindNew, which is basically like `git diff`. Corresponds to `find-new` in btrfs.
// Case: spaces in filenames
// Case: create, delete, edit files and check that the filenames correspond to the changes ones.

// go test coverage
