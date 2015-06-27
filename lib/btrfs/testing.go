package btrfs

import (
	"bufio"
	"runtime/debug"
	"testing"

	"github.com/pachyderm/pfs/lib/utils"
)

// This file contains convenience functions for testing the btrfs module.

// CheckFile checks if a file on disk contains a given string.
func CheckFile(name, content string, t *testing.T) {
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

// CheckNoExists checks that no file is present.
func CheckNoExists(name string, t *testing.T) {
	exists, err := FileExists(name)
	check(err, t)
	if exists {
		t.Fatalf("File %s shouldn't exist.", name)
	}
}

// CheckExists checks that a file is present
func CheckExists(name string, t *testing.T) {
	exists, err := FileExists(name)
	check(err, t)
	if !exists {
		t.Fatalf("File %s should exist.", name)
	}
}
