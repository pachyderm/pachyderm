package btrfs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
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
	repoName := "repo_TestGit"
	// Create the repo:
	check(Init(repoName), t)

	// Write a file "file" and create a commit "commit1":
	writeFile(fmt.Sprintf("%s/master/file", repoName), "foo", t)
	err := Commit(repoName, "commit1", "master")
	check(err, t)
	checkFile(path.Join(repoName, "commit1", "file"), "foo", t)

	// Create a new branch "branch" from commit "commit1", and check that
	// it contains the file "file":
	check(Branch(repoName, "commit1", "branch"), t)
	checkFile(fmt.Sprintf("%s/branch/file", repoName), "foo", t)

	// Create a file "file2" in branch "branch", and commit it to
	// "commit2":
	writeFile(fmt.Sprintf("%s/branch/file2", repoName), "foo", t)
	err = Commit(repoName, "commit2", "branch")
	check(err, t)
	checkFile(path.Join(repoName, "commit2", "file2"), "foo", t)

	// Print BTRFS hierarchy data for humans:
	check(Log(repoName, "0", func(r io.ReadCloser) error {
		_, err := io.Copy(os.Stdout, r)
		return err
	}), t)
}

func TestNewRepoIsEmpty(t *testing.T) {
	repoName := "repo_TesNewRepoIsEmpty"
	check(Init(repoName), t)

	// ('master' is the default branch)
	dirpath := path.Join(repoName, "master")
	descriptors, err := ReadDir(dirpath)
	check(err, t)
	if len(descriptors) != 0 {
		t.Fatalf("expected empty repo")
	}
}

func TestCommitsAreReadOnly(t *testing.T) {
	repoName := "repo_TestCommitsAreReadOnly"
	check(Init(repoName), t)

	err := Commit(repoName, "commit1", "master")
	check(err, t)

	_, err = Create(fmt.Sprintf("%s/commit1/file", repoName))
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "read-only file system") {
		t.Fatalf("expected read-only filesystem error, got %s", err)
	}
}

func TestBranchesAreReadWrite(t *testing.T) {
	repoName := "repo_TestBranchesAreReadWrite"
	check(Init(repoName), t)

	err := Branch(repoName, "t0", "my_branch")
	check(err, t)

	fn := fmt.Sprintf("%s/my_branch/file", repoName)
	writeFile(fn, "some content", t)
	checkFile(fn, "some content", t)
}

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

// TestHoldRelease creates one-off commit named after a UUID, to ensure a data consumer can always access data in a commit, even if the original commit is deleted.
func TestHoldRelease(t *testing.T) {
	repoName := "repo_TestHoldRelease"
	check(Init(repoName), t)

	// Write a file "myfile" with contents "foo":
	master_fn := fmt.Sprintf("%s/master/myfile", repoName)
	writeFile(master_fn, "foo", t)
	checkFile(master_fn, "foo", t)

	// Create a commit "mycommit" and verify "myfile" exists:
        mycommit_fn := fmt.Sprintf("%s/mycommit/myfile", repoName)
	check(Commit(repoName, "mycommit", "master"), t)
	checkFile(mycommit_fn, "foo", t)

	// Grab a snapshot:
	snapshot_path, err := Hold(repoName, "mycommit")
	check(err, t)

	// Delete the commit from the snapshot.
	// (uses the lower-level btrfs command for now):
	mycommit_path := fmt.Sprintf("%s/mycommit", repoName)
	check(SubvolumeDelete(mycommit_path), t)

	// Verify that the file still exists in our snapshot:
	snapshot_fn := fmt.Sprintf("%s/myfile", snapshot_path)
	checkFile(snapshot_fn, "foo", t)
}


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
