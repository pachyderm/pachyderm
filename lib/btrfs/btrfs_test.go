package btrfs

import (
	"bufio"
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
func writeFile(fs *FS, name, content string, t *testing.T) {
	f, err := fs.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(content + "\n")
	f.Close()

}

// checkFile checks if a file on disk contains a given string.
func checkFile(fs *FS, name, content string, t *testing.T) {
	exists, err := fs.FileExists(name)
	check(err, t)
	if !exists {
		t.Fatalf("File %s should exist.", name)
	}

	f, err := fs.Open(name)
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
func checkNoFile(fs *FS, name string, t *testing.T) {
	exists, err := fs.FileExists(name)
	check(err, t)
	if exists {
		t.Fatalf("File %s shouldn't exist.", name)
	}
}

func removeFile(fs *FS, name string, t *testing.T) {
	err := fs.Remove(name)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOsOps(t *testing.T) {
	fs := NewFSWithRandSeq("TestOsOps")
	fs.EnsureNamespace()
	writeFile(fs, "foo", "foo", t)
	checkFile(fs, "foo", "foo", t)
	removeFile(fs, "foo", t)
	checkNoFile(fs, "foo", t)
}

func TestGit(t *testing.T) {
	fs := NewFSWithRandSeq("TestGit")
	fs.EnsureNamespace()

	check(fs.Init("repo"), t)
	writeFile(fs, "repo/branches/master/foo", "foo", t)
	commit, err := fs.Commit("repo", "master")
	check(err, t)
	checkFile(fs, path.Join("repo", "commits", commit, "foo"), "foo", t)
}
