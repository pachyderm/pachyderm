package pipeline

import (
	"path"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/pachyderm/pfs/lib/btrfs"
)

func check(err error, t *testing.T) {
	if err != nil {
		debug.PrintStack()
		t.Fatal(err)
	}
}

// TestOutput tests a simple job that outputs some data.
func TestOutput(t *testing.T) {
	outRepo := "TestOuput"
	check(btrfs.Init(outRepo), t)
	pipeline := NewPipeline("", outRepo, "", "master")
	pachfile := `
image ubuntu

run touch /out/foo
run touch /out/bar
`
	err := pipeline.RunPachFile(strings.NewReader(pachfile))
	check(err, t)

	exists, err := btrfs.FileExists(path.Join(outRepo, "0", "foo"))
	check(err, t)
	if exists != true {
		t.Fatal("File `foo` doesn't exist when it should.")
	}

	exists, err = btrfs.FileExists(path.Join(outRepo, "1", "bar"))
	check(err, t)
	if exists != true {
		t.Fatal("File `bar` doesn't exist when it should.")
	}
}

func TestInputOutput(t *testing.T) {
	// create the in repo
	inRepo := "TestInputOutput_in"
	check(btrfs.Init(inRepo), t)

	// add data to it
	err := btrfs.WriteFile(path.Join(inRepo, "master", "data", "foo"), []byte("foo"))
	check(err, t)

	// commit data
	err = btrfs.Commit(inRepo, "commit", "master")
	check(err, t)

	outRepo := "TestInputOutput_out"
	check(btrfs.Init(outRepo), t)

	pipeline := NewPipeline(inRepo, outRepo, "commit", "master")

	pachfile := `
image ubuntu

import data

run cp /in/data/foo /out/foo
`
	err = pipeline.RunPachFile(strings.NewReader(pachfile))
	check(err, t)

	exists, err := btrfs.FileExists(path.Join(outRepo, "0", "foo"))
	check(err, t)
	if exists != true {
		t.Fatal("File `foo` doesn't exist when it should.")
	}
}

func TestLog(t *testing.T) {
	outRepo := "TestLog"
	check(btrfs.Init(outRepo), t)
	pipeline := NewPipeline("", outRepo, "", "master")
	pachfile := `
image ubuntu

run echo "foo"
`
	err := pipeline.RunPachFile(strings.NewReader(pachfile))
	check(err, t)

	exists, err := btrfs.FileExists(path.Join(outRepo, "0", ".log"))
	check(err, t)
	if exists != true {
		t.Fatal("File .log should exist.")
	}
}
