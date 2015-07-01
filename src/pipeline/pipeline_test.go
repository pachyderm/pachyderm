package pipeline

import (
	"path"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pfs/src/btrfs"
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
	pipeline := newPipeline("output", "", outRepo, "commit", "master", "0-1", "")
	pachfile := `
image ubuntu

# touch foo
run touch /out/foo
# touch bar
run touch /out/bar
`
	err := pipeline.runPachFile(strings.NewReader(pachfile))
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

func TestEcho(t *testing.T) {
	outRepo := "TestEcho"
	check(btrfs.Init(outRepo), t)
	pipeline := newPipeline("echo", "", outRepo, "commit", "master", "0-1", "")
	pachfile := `
image ubuntu

run echo foo >/out/foo
run echo foo >/out/bar
`
	err := pipeline.runPachFile(strings.NewReader(pachfile))
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

	pipeline := newPipeline("input_output", inRepo, outRepo, "commit", "master", "0-1", "")

	pachfile := `
image ubuntu

input data

run cp /in/data/foo /out/foo
`
	err = pipeline.runPachFile(strings.NewReader(pachfile))
	check(err, t)

	exists, err := btrfs.FileExists(path.Join(outRepo, "commit-0", "foo"))
	check(err, t)
	if exists != true {
		t.Fatal("File `foo` doesn't exist when it should.")
	}
}

// TODO(jd), this test is falsely passing I think that however we're getting
// logs from Docker doesn't work.
func TestLog(t *testing.T) {
	outRepo := "TestLog"
	check(btrfs.Init(outRepo), t)
	pipeline := newPipeline("log", "", outRepo, "commit1", "master", "0-1", "")
	pachfile := `
image ubuntu

run echo "foo"
`
	err := pipeline.runPachFile(strings.NewReader(pachfile))
	check(err, t)

	log, err := btrfs.ReadFile(path.Join(outRepo, "commit1-0", ".log"))
	check(err, t)
	if string(log) != "foo\n" {
		t.Fatal("Expect foo, got: ", string(log))
	}

	pipeline = newPipeline("log", "", outRepo, "commit2", "master", "0-1", "")
	pachfile = `
image ubuntu

run echo "bar" >&2
`
	err = pipeline.runPachFile(strings.NewReader(pachfile))
	check(err, t)

	log, err = btrfs.ReadFile(path.Join(outRepo, "commit2-0", ".log"))
	check(err, t)
	if string(log) != "bar\n" {
		t.Fatal("Expect bar, got: ", string(log))
	}
}

// TestScrape tests a the scraper pipeline
func TestScrape(t *testing.T) {
	inRepo := "TestScrape_in"
	check(btrfs.Init(inRepo), t)
	outRepo := "TestScrape_out"
	check(btrfs.Init(outRepo), t)

	// Create a url to scrape
	check(btrfs.WriteFile(path.Join(inRepo, "master", "urls", "1"), []byte("pachyderm.io")), t)

	// Commit the data
	check(btrfs.Commit(inRepo, "commit", "master"), t)

	// Create a pipeline to run
	pipeline := newPipeline("scrape", inRepo, outRepo, "commit", "master", "0-1", "")
	pachfile := `
image busybox

input urls

run cat /in/urls/* | xargs  wget -P /out
`
	err := pipeline.runPachFile(strings.NewReader(pachfile))

	exists, err := btrfs.FileExists(path.Join(outRepo, "commit", "index.html"))
	check(err, t)
	if !exists {
		t.Fatal("pachyderm.io should exists")
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

input data

run cp /in/data/foo /out/foo
run echo "foo"
`)), t)
	check(btrfs.Commit(inRepo, "commit", "master"), t)

	check(RunPipelines("pipeline", inRepo, outPrefix, "commit", "master", "0-1"), t)

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

	err := RunPipelines("pipeline", inRepo, outPrefix, "commit", "master", "0-1")
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
	err := RunPipelines("pipeline", inRepo, outPrefix, "commit1", "master", "0-1")
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
	err = RunPipelines("pipeline", inRepo, outPrefix, "commit2", "master", "0-1")
	// this time the pipelines should not err
	check(err, t)

	// These are the most important 2 checks:

	// If this one fails it means that dirty state isn't properly saved
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit1-fail/bar"), t)
	// If this one fails it means that dirty state isn't properly cleared
	btrfs.CheckNoExists(path.Join(outPrefix, "recover", "commit2-0/bar"), t)

	// These commits are mostly covered by other tests
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit1-fail/foo"), t)
	btrfs.CheckExists(path.Join(outPrefix, "recover", "commit1-0/foo"), t)
	btrfs.CheckNoExists(path.Join(outPrefix, "recover", "commit1-1"), t)
	btrfs.CheckNoExists(path.Join(outPrefix, "recover", "commit1"), t)
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

	r := NewRunner("pipeline", inRepo, outPrefix, "commit", "master", "0-1")
	go func() {
		err := r.Run()
		if err != ErrCancelled {
			t.Fatal("Should get `ErrCancelled` error.")
		}
	}()

	// This is just to make sure we don't trigger the early exit case in Run
	// and actually exercise the code.
	time.Sleep(time.Second * 2)
	check(r.Cancel(), t)
}

// TestWrap tests a simple job that uses line wrapping in it's Pachfile
func TestWrap(t *testing.T) {
	outRepo := "TestWrap"
	check(btrfs.Init(outRepo), t)
	pipeline := newPipeline("output", "", outRepo, "commit", "master", "0-1", "")
	pachfile := `
image ubuntu

# touch foo and bar
run touch /out/foo \
          /out/bar
`
	err := pipeline.runPachFile(strings.NewReader(pachfile))
	check(err, t)

	exists, err := btrfs.FileExists(path.Join(outRepo, "commit", "foo"))
	check(err, t)
	if exists != true {
		t.Fatal("File `foo` doesn't exist when it should.")
	}

	exists, err = btrfs.FileExists(path.Join(outRepo, "commit", "bar"))
	check(err, t)
	if exists != true {
		t.Fatal("File `bar` doesn't exist when it should.")
	}
}

func TestDependency(t *testing.T) {
	inRepo := "TestDependencyin"
	check(btrfs.Init(inRepo), t)
	p1 := `
image ubuntu

run echo foo >/out/foo
`
	check(btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "p1"), []byte(p1)), t)
	p2 := `
image ubuntu

input pps://p1

run cp /in/p1/foo /out/foo
`
	check(btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "p2"), []byte(p2)), t)
	check(btrfs.Commit(inRepo, "commit", "master"), t)

	outPrefix := "TestDependency"
	runner := NewRunner("pipeline", inRepo, outPrefix, "commit", "master", "0-1")
	check(runner.Run(), t)

	res, err := btrfs.ReadFile(path.Join(outPrefix, "p2", "commit", "foo"))
	check(err, t)
	if string(res) != "foo\n" {
		t.Fatal("Expected foo, got: ", string(res))
	}
}

func TestRunnerInputs(t *testing.T) {
	inRepo := "TestRunnerInputsin"
	check(btrfs.Init(inRepo), t)
	p1 := `
image ubuntu

input foo
input bar
`
	check(btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "p1"), []byte(p1)), t)
	p2 := `
image ubuntu

input fizz
input buzz
`
	check(btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "p2"), []byte(p2)), t)
	check(btrfs.Commit(inRepo, "commit", "master"), t)

	outPrefix := "TestRunnerInputs"
	runner := NewRunner("pipeline", inRepo, outPrefix, "commit", "master", "0-1")
	inputs, err := runner.Inputs()
	check(err, t)
	if strings.Join(inputs, " ") != "foo bar fizz buzz" {
		t.Fatal("Incorrect inputs: ", inputs, " expected: ", []string{"foo", "bar", "fizz", "buzz"})
	}
}

// TestInject tests that s3 injections works
func TestInject(t *testing.T) {
	outRepo := "TestInject"
	check(btrfs.Init(outRepo), t)
	pipeline := newPipeline("output", "", outRepo, "commit", "master", "0-1", "")
	check(pipeline.inject("s3://pachyderm-test/pipeline"), t)
	check(pipeline.finish(), t)
	res, err := btrfs.ReadFile(path.Join(outRepo, "commit", "file"))
	check(err, t)
	if string(res) != "foo\n" {
		t.Fatal("Expected foo, got: ", string(res))
	}
}
