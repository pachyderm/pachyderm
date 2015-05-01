package btrfs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
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
	check(Log(repoName, "0", func(r io.Reader) error {
		_, err := io.Copy(os.Stdout, r)
		return err
	}), t)
}

func TestNewRepoIsEmpty(t *testing.T) {
	repoName := "repo_TestNewRepoIsEmpty"
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

// TestSendRecv checks the Send and Recv replication primitives.
func TestSendRecv(t *testing.T) {
	// Create a source repo:
	srcRepo := "repo_TestSendRecv_src"
	check(Init(srcRepo), t)

	// Create a file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile1", srcRepo), "foo", t)

	// Create a commit in the source repo:
	check(Commit(srcRepo, "mycommit1", "master"), t)

	// Create another file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile2", srcRepo), "bar", t)

	// Create a another commit in the source repo:
	check(Commit(srcRepo, "mycommit2", "master"), t)

	// Create a destination repo:
	dstRepo := "repo_TestSendRecv_dst"
	check(Init(dstRepo), t)

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoFile(fmt.Sprintf("%s/mycommit1", dstRepo), t)
	checkNoFile(fmt.Sprintf("%s/mycommit2", dstRepo), t)

	// Run a Send/Recv operation to fetch data from the older "mycommit1".
	// This verifies that tree copying works:
	repo2Recv := func(r io.Reader) error { return Recv(dstRepo, r) }
	check(Send(fmt.Sprintf("%s/t0", srcRepo), fmt.Sprintf("%s/mycommit1", srcRepo), repo2Recv), t)

	// Check that the file from mycommit1 exists, but not from mycommit2:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo", t)
	checkNoFile(fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), t)

	// Send again, this time starting from mycommit1 and going to mycommit2:
	repo2Recv = func(r io.Reader) error { return Recv(dstRepo, r) }
	check(Send(fmt.Sprintf("%s/mycommit1", srcRepo), fmt.Sprintf("%s/mycommit2", srcRepo), repo2Recv), t)

	// Verify that files from both commits are present:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), "bar", t)
}

// TestSendBaseRecv checks the SendBase and Recv replication primitives.
func TestSendBaseRecv(t *testing.T) {
	// Create a source repo:
	srcRepo := "repo_TestSendBaseRecv_src"
	check(Init(srcRepo), t)

	// Create a file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile1", srcRepo), "foo", t)

	// Create a commit in the source repo:
	check(Commit(srcRepo, "mycommit1", "master"), t)

	// Create another file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile2", srcRepo), "bar", t)

	// Create a another commit in the source repo:
	check(Commit(srcRepo, "mycommit2", "master"), t)

	// Create a destination repo:
	dstRepo := "repo_TestSendBaseRecv_dst"
	check(Init(dstRepo), t)

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoFile(fmt.Sprintf("%s/mycommit1", dstRepo), t)
	checkNoFile(fmt.Sprintf("%s/mycommit2", dstRepo), t)

	// Run a SendBase/Recv operation to fetch data from the source commits:
	// This verifies that tree copying works:
	repo2Recv := func(r io.Reader) error { return Recv(dstRepo, r) }
	check(SendBase(fmt.Sprintf("%s/mycommit1", srcRepo), repo2Recv), t)
	check(SendBase(fmt.Sprintf("%s/mycommit2", srcRepo), repo2Recv), t)

	// Verify that files from both commits are present:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), "bar", t)
}

// TestSendWithMissingIntermediateCommitIsCorrect(?) // ? means we don't know what the behavior is.

// TestBranchesAreNotReplicated // this is a known property, but not desirable long term
// TestCommitsAreReplicated // Uses Send and Recv
func TestCommitsAreReplicated(t *testing.T) {
	// Create a source repo:
	srcRepo := "repo_TestSendCommitsAreReplicated_src"
	check(Init(srcRepo), t)

	// Create a file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile1", srcRepo), "foo", t)

	// Create a commit in the source repo:
	check(Commit(srcRepo, "mycommit1", "master"), t)

	// Create another file in the source repo:
	writeFile(fmt.Sprintf("%s/master/myfile2", srcRepo), "bar", t)

	// Create a another commit in the source repo:
	check(Commit(srcRepo, "mycommit2", "master"), t)

	// Create a destination repo:
	dstRepo := "repo_TestSendCommitsAreReplicated_dst"
	check(InitBare(dstRepo), t)

	// Verify that the commits "mycommit1" and "mycommit2" do in source:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", srcRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/mycommit2/myfile2", srcRepo), "bar", t)

	// Verify that the commits "mycommit1" and "mycommit2" do not exist in destination:
	checkNoFile(fmt.Sprintf("%s/mycommit1", dstRepo), t)
	checkNoFile(fmt.Sprintf("%s/mycommit2", dstRepo), t)

	// Run a Pull/Recv operation to fetch all commits:
	repo2Recv := func(r io.Reader) error { return Recv(dstRepo, r) }
	transid, err := Transid(srcRepo, "t0") // TODO(rw,jd): test other transids here
	check(err, t)
	check(Pull(srcRepo, transid, repo2Recv), t)

	// Verify that files from both commits are present:
	checkFile(fmt.Sprintf("%s/mycommit1/myfile1", dstRepo), "foo", t)
	checkFile(fmt.Sprintf("%s/mycommit2/myfile2", dstRepo), "bar", t)
}

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
func TestFindNew(t *testing.T) {
	repoName := "repo_TestFindNew"
	check(Init(repoName), t)

	checkFindNew := func(want []string, repo, branch, transid string) {
		got, err := FindNew(repo, branch, transid)
		check(err, t)
		t.Logf("checkFindNew(%v, %v, %v) -> %v", repo, branch, transid, got)

		// handle nil and empty slice the same way:
		if len(want) == 0 && len(got) == 0 {
			return
		}

		if !reflect.DeepEqual(want, got) {
			t.Fatalf("wanted %v, got %v for FindNew(%v, %v, %v)", want, got, repo, branch, transid)
		}

	}

	// Get a transaction ID for the first commit:
	transid0, err := Transid(repoName, "t0")
	check(err, t)

	// There are no new files upon repo creation:
	checkFindNew([]string{}, repoName, "t0", transid0)

	// A new, uncommited file is returned in the list:
	writeFile(fmt.Sprintf("%s/master/myfile1", repoName), "foo", t)
	checkFindNew([]string{"myfile1"}, repoName, "master", transid0)

	// When that file is commited, then it still shows up in the delta since transid0:
	check(Commit(repoName, "mycommit1", "master"), t)
	// TODO(rw, jd) Shouldn't this pass?
	checkFindNew([]string{"myfile1"}, repoName, "mycommit1", transid0)

	// Get a transaction ID for the second commit:
	transid1, err := Transid(repoName, "mycommit1")
	check(err, t)

	// The file doesn't show up in the delta since the new transaction:
	checkFindNew([]string{}, repoName, "mycommit1", transid1)

	// Sanity check: the old delta still gives the same result:
	checkFindNew([]string{"myfile1"}, repoName, "master", transid0)
}

// Case: spaces in filenames
// Case: create, delete, edit files and check that the filenames correspond to the changes ones.

// go test coverage
