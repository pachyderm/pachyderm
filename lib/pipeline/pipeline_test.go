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
	pipeline := NewPipeline("", outRepo, "commit", "master")
	pachfile := `
image ubuntu

run touch /out/foo
run touch /out/bar
`
	err := pipeline.RunPachFile(strings.NewReader(pachfile))
	check(err, t)

	exists, err := btrfs.FileExists(path.Join(outRepo, "commit-0", "foo"))
	check(err, t)
	if exists != true {
		t.Fatal("File `foo` doesn't exist when it should.")
	}

	exists, err = btrfs.FileExists(path.Join(outRepo, "commit-1", "bar"))
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

	exists, err := btrfs.FileExists(path.Join(outRepo, "commit-0", "foo"))
	check(err, t)
	if exists != true {
		t.Fatal("File `foo` doesn't exist when it should.")
	}
}

func TestLog(t *testing.T) {
	outRepo := "TestLog"
	check(btrfs.Init(outRepo), t)
	pipeline := NewPipeline("", outRepo, "commit", "master")
	pachfile := `
image ubuntu

run echo "foo"
`
	err := pipeline.RunPachFile(strings.NewReader(pachfile))
	check(err, t)

	exists, err := btrfs.FileExists(path.Join(outRepo, "commit-0", ".log"))
	check(err, t)
	if exists != true {
		t.Fatal("File .log should exist.")
	}
}

func TestPipelines(t *testing.T) {
	inRepo := "TestPipelines_in"
	check(btrfs.Init(inRepo), t)
	outRepo := "TestPipelines_out"
	check(btrfs.Init(outRepo), t)

	// Create a data file:
	check(btrfs.WriteFile(path.Join(inRepo, "master", "data", "foo"), []byte("foo")), t)

	// Create the Pachfile
	check(btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "cp"), []byte(`
image ubuntu

import data

run cp /in/data/foo /out/foo
run echo "foo"
`)), t)
	check(btrfs.Commit(inRepo, "commit", "master"), t)

	check(RunPipelines("pipeline", inRepo, outRepo, "commit", "master"), t)

	data, err := btrfs.ReadFile(path.Join(outRepo, "commit", "foo"))
	check(err, t)
	if string(data) != "foo" {
		t.Fatal("Incorrect file content.")
	}
}
