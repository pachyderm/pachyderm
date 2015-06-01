package pipeline

import (
	"path"
	"runtime/debug"
	"strings"
	"testing"
	"time"

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

// TestPipelines runs a 2 step pipeline.
func TestPipelines(t *testing.T) {
	inRepo := "TestPipelines_in"
	check(btrfs.Init(inRepo), t)
	outPrefix := "TestPipelines_out"

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

	check(RunPipelines("pipeline", inRepo, outPrefix, "commit", "master"), t)

	data, err := btrfs.ReadFile(path.Join(outPrefix, "cp", "commit", "foo"))
	check(err, t)
	if string(data) != "foo" {
		t.Fatal("Incorrect file content.")
	}
}

// TestError makes sure that we handle commands that error correctly.
func TestError(t *testing.T) {
	inRepo := "TestError_in"
	check(btrfs.Init(inRepo), t)
	outPrefix := "TestError_out"

	// Create the Pachfile
	check(btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "error"), []byte(`
image ubuntu

run touch /out/foo
run cp /in/foo /out/bar
`)), t)
	// Last line should fail here.

	// Commit to the inRepo
	check(btrfs.Commit(inRepo, "commit", "master"), t)

	err := RunPipelines("pipeline", inRepo, outPrefix, "commit", "master")
	if err == nil {
		t.Fatal("Running pipeline should error.")
	}

	// Check that foo exists
	exists, err := btrfs.FileExists(path.Join(outPrefix, "error", "commit-0", "foo"))
	check(err, t)
	if !exists {
		t.Fatal("File foo should exist.")
	}

	// Check that commit doesn't exist
	exists, err = btrfs.FileExists(path.Join(outPrefix, "error", "commit"))
	check(err, t)
	if exists {
		t.Fatal("Commit \"commit\" should not get created when a command fails.")
	}
}

// TestRecover runs a pipeline with an error. Then fixes the pipeline to not
// include an error and reruns it.
func TestRecover(t *testing.T) {
	inRepo := "TestRecover_in"
	check(btrfs.Init(inRepo), t)
	outPrefix := "TestRecover_out"

	// Create the Pachfile
	check(btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "recover"), []byte(`
image ubuntu

run touch /out/foo
run touch /out/bar && cp /in/foo /out/bar
`)), t)
	// Last line should fail here.

	// Commit to the inRepo
	check(btrfs.Commit(inRepo, "commit1", "master"), t)

	// Run the pipelines
	err := RunPipelines("pipeline", inRepo, outPrefix, "commit1", "master")
	if err == nil {
		t.Fatal("Running pipeline should error.")
	}

	// Fix the Pachfile
	check(btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "recover"), []byte(`
image ubuntu

run touch /out/foo
run touch /out/bar
`)), t)

	// Commit to the inRepo
	check(btrfs.Commit(inRepo, "commit2", "master"), t)

	// Run the pipelines
	err = RunPipelines("pipeline", inRepo, outPrefix, "commit2", "master")
	// this time the pipelines should not err
	check(err, t)

	// These are the most important 2 checks:

	// If this one fails it means that dirty state isn't properly saved
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit2-pre/bar"), t)
	// If this one fails it means that dirty state isn't properly cleared
	btrfs.CheckNoExists(path.Join(outPrefix, "recover", "commit2-0/bar"), t)

	// These commits are mostly covered by other tests
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit1-0/foo"), t)
	btrfs.CheckNoExists(path.Join(outPrefix, "recover", "commit1-1"), t)
	btrfs.CheckNoExists(path.Join(outPrefix, "recover", "commit1"), t)
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit2-pre/foo"), t)
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit2-0/foo"), t)
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit2-1/foo"), t)
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit2-1/bar"), t)
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit2/foo"), t)
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit2/bar"), t)
}

func TestCancel(t *testing.T) {
	inRepo := "TestCancel_in"
	check(btrfs.Init(inRepo), t)
	outPrefix := "TestCancel_out"

	// Create the Pachfile
	check(btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "cancel"), []byte(`
image ubuntu

run sleep 100
`)), t)
	check(btrfs.Commit(inRepo, "commit", "master"), t)

	r := NewRunner("pipeline", inRepo, outPrefix, "commit", "master")
	go func() {
		err := r.Run()
		if err != Cancelled {
			t.Fatal("Should get `Cancelled` error.")
		}
	}()

	// This is just to make sure we don't trigger the early exit case in Run
	// and actually exercise the code.
	time.Sleep(time.Second * 2)
	check(r.Cancel(), t)
}
