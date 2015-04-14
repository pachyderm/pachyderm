package btrfs

import (
	"bufio"
	"io"
	"os"
	"path"
	"testing"
)

var run_string string

func check(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

// writeFile quickly writes a string to disk.
func writeFile(name, content string, t *testing.T) {
	f, err := Create(name)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content + "\n")
	f.Close()

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
	// Create the repo:
	check(Init("repo"), t)

	// Write a file "file" and create a commit "commit1":
	writeFile("repo/master/file", "foo", t)
	err := Commit("repo", "commit1", "master")
	check(err, t)
	checkFile(path.Join("repo", "commit1", "file"), "foo", t)

	// Create a new branch "branch" from commit "commit1", and check that
	// it contains the file "file":
	check(Branch("repo", "commit1", "branch"), t)
	checkFile("repo/branch/file", "foo", t)

	// Create a file "file2" in branch "branch", and commit it to
	// "commit2":
	writeFile("repo/branch/file2", "foo", t)
	err = Commit("repo", "commit2", "branch")
	check(err, t)
	checkFile(path.Join("repo", "commit2", "file2"), "foo", t)

	// Print BTRFS hierarchy data for humans:
	check(Log("repo", "0", func(r io.ReadCloser) error {
		_, err := io.Copy(os.Stdout, r)
		return err
	}), t)
}

// TestNewRepoIsEmpty
// TestCommitsAreReadOnly // implement by setting readonly
// TestBranchesAreReadWrite

// TestReplication checks that replication is correct when using local BTRFS.
// Uses `Pull`
// This is heavier and hairier, do it last.
func TestReplication(t *testing.T) {
	t.Skip("implement this")
}

// TestSendRecv // low-level
// TestSendBaseRecv // low-level

// TestSendWithMissingIntermediateCommitIsCorrect(?) // ? means we don't know what the behavior is.

// TestBranchesAreNotReplicated // this is a known property, but not desirable long term
// TestCommitsAreReplicated // Uses Send and Recv

// TestHoldRelease // Creates one-off commit named after a UUID, to ensure a data consumer can always access data in a commit, even if the original commit is deleted.

// Test for `Commits`: check that the sort order of CommitInfo objects is structured correctly.
// Start from:
//	// Print BTRFS hierarchy data for humans:
//	check(Log("repo", "0", func(r io.ReadCloser) error {
//		_, err := io.Copy(os.Stdout, r)
//		return err
//	}), t)

// TestFindNew, which is basically like `git diff`. Corresponds to `find-new` in btrfs.
// Case: spaces in filenames
// Case: create, delete, edit files and check that the filenames correspond to the changes ones.


// go test coverage
