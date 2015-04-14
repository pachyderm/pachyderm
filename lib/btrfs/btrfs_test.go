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

func TestOsOps(t *testing.T) {
	writeFile("foo", "foo", t)
	checkFile("foo", "foo", t)
	removeFile("foo", t)
	checkNoFile("foo", t)
}

func TestGit(t *testing.T) {
	check(Init("repo"), t)
	writeFile("repo/master/file", "foo", t)
	err := Commit("repo", "commit1", "master")
	check(err, t)
	checkFile(path.Join("repo", "commit1", "file"), "foo", t)

	check(Branch("repo", "commit1", "branch"), t)
	checkFile("repo/branch/file", "foo", t)

	writeFile("repo/branch/file2", "foo", t)
	err = Commit("repo", "commit2", "branch")
	check(err, t)
	checkFile(path.Join("repo", "commit2", "file2"), "foo", t)

	check(Log("repo", "0", func(r io.ReadCloser) error {
		_, err := io.Copy(os.Stdout, r)
		return err
	}), t)
}
