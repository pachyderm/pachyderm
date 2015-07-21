package pipeline

import (
	"path"
	"strings"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/btrfs"
	"github.com/pachyderm/pachyderm/src/etcache"
	"github.com/stretchr/testify/require"
)

// TestOutput tests a simple pipeline that outputs some data.
func TestOutput(t *testing.T) {
	t.Parallel()
	pipeline := newTestPipeline(t, "output", "commit", "master", "0-1", true)
	pachfile := `
image ubuntu

# touch foo
run touch /out/foo
# touch bar
run touch /out/bar
`
	err := pipeline.runPachFile(strings.NewReader(pachfile))
	require.NoError(t, err)

	exists, err := btrfs.FileExists(path.Join(pipeline.outRepo, "commit-0", "foo"))
	require.NoError(t, err)
	require.True(t, exists, "File `foo` doesn't exist when it should.")

	exists, err = btrfs.FileExists(path.Join(pipeline.outRepo, "commit-1", "bar"))
	require.NoError(t, err)
	require.True(t, exists, "File `bar` doesn't exist when it should.")
}

func TestEcho(t *testing.T) {
	t.Parallel()
	pipeline := newTestPipeline(t, "echo", "commit", "master", "0-1", true)
	pachfile := `
image ubuntu

run echo foo >/out/foo
run echo foo >/out/bar
`
	err := pipeline.runPachFile(strings.NewReader(pachfile))
	require.NoError(t, err)

	exists, err := btrfs.FileExists(path.Join(pipeline.outRepo, "commit-0", "foo"))
	require.NoError(t, err)
	require.True(t, exists, "File `foo` doesn't exist when it should.")

	exists, err = btrfs.FileExists(path.Join(pipeline.outRepo, "commit-1", "bar"))
	require.NoError(t, err)
	require.True(t, exists, "File `bar` doesn't exist when it should.")
}

func TestInputOutput(t *testing.T) {
	t.Parallel()
	// create the in repo
	pipeline := newTestPipeline(t, "inputOutput", "commit", "master", "0-1", true)

	// add data to it
	err := btrfs.WriteFile(path.Join(pipeline.inRepo, "master", "data", "foo"), []byte("foo"))
	require.NoError(t, err)

	// commit data
	err = btrfs.Commit(pipeline.inRepo, "commit", "master")
	require.NoError(t, err)

	pachfile := `
image ubuntu

input data

run cp /in/data/foo /out/foo
`
	err = pipeline.runPachFile(strings.NewReader(pachfile))
	require.NoError(t, err)

	exists, err := btrfs.FileExists(path.Join(pipeline.outRepo, "commit-0", "foo"))
	require.NoError(t, err)
	require.True(t, exists, "File `foo` doesn't exist when it should.")
}

// TODO(jd), this test is falsely passing I think that however we're getting
// logs from Docker doesn't work.
func TestLog(t *testing.T) {
	t.Parallel()
	pipeline := newTestPipeline(t, "log", "commit1", "master", "0-1", true)
	pachfile := `
image ubuntu

run echo "foo"
`
	err := pipeline.runPachFile(strings.NewReader(pachfile))
	require.NoError(t, err)

	log, err := btrfs.ReadFile(path.Join(pipeline.outRepo, "commit1-0", ".log"))
	require.NoError(t, err)
	require.Equal(t, "foo\n", string(log))

	pipeline = newTestPipeline(t, "log", "commit2", "master", "0-1", false)
	pachfile = `
image ubuntu

run echo "bar" >&2
`
	err = pipeline.runPachFile(strings.NewReader(pachfile))
	require.NoError(t, err)

	log, err = btrfs.ReadFile(path.Join(pipeline.outRepo, "commit2-0", ".log"))
	require.NoError(t, err)
	require.Equal(t, "bar\n", string(log))
}

// TestScrape tests a the scraper pipeline
func TestScrape(t *testing.T) {
	// TODO(any): what?? wget is not found in the container if parallel is set
	//t.Parallel()
	pipeline := newTestPipeline(t, "scrape", "commit", "master", "0-1", true)

	// Create a url to scrape
	require.NoError(t, btrfs.WriteFile(path.Join(pipeline.inRepo, "master", "urls", "1"), []byte("pachyderm.io")))

	// Commit the data
	require.NoError(t, btrfs.Commit(pipeline.inRepo, "commit", "master"))

	// Create a pipeline to run
	pachfile := `
image pachyderm/scraper

input urls

run cat /in/urls/* | xargs wget -P /out
`
	err := pipeline.runPachFile(strings.NewReader(pachfile))

	exists, err := btrfs.FileExists(path.Join(pipeline.outRepo, "commit", "index.html"))
	require.NoError(t, err)
	require.True(t, exists, "pachyderm.io should exist")
}

// TestPipelines runs a 2 step pipeline.
func TestPipelines(t *testing.T) {
	t.Parallel()
	inRepo := "TestPipelines_in"
	require.NoError(t, btrfs.Init(inRepo))
	outPrefix := "TestPipelines_out"

	// Create a data file:
	require.NoError(t, btrfs.WriteFile(path.Join(inRepo, "master", "data", "foo"), []byte("foo")))

	// Create the Pachfile
	require.NoError(t, btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "cp"), []byte(`
image ubuntu

input data

run cp /in/data/foo /out/foo
run echo "foo"
`)))
	require.NoError(t, btrfs.Commit(inRepo, "commit", "master"))

	require.NoError(t, RunPipelines("pipeline", inRepo, outPrefix, "commit", "master", "0-1", etcache.NewCache()))

	data, err := btrfs.ReadFile(path.Join(outPrefix, "cp", "commit", "foo"))
	require.NoError(t, err)
	require.Equal(t, "foo", string(data))
}

// TestError makes sure that we handle commands that error correctly.
func TestError(t *testing.T) {
	t.Parallel()
	inRepo := "TestError_in"
	require.NoError(t, btrfs.Init(inRepo))
	outPrefix := "TestError_out"

	// Create the Pachfile
	require.NoError(t, btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "error"), []byte(`
image ubuntu

run touch /out/foo
run cp /in/foo /out/bar
`)))
	// Last line should fail here.

	// Commit to the inRepo
	require.NoError(t, btrfs.Commit(inRepo, "commit", "master"))

	err := RunPipelines("pipeline", inRepo, outPrefix, "commit", "master", "0-1", etcache.NewCache())
	require.Error(t, err, "Running pipeline should error.")

	// Check that foo exists
	exists, err := btrfs.FileExists(path.Join(outPrefix, "error", "commit-0", "foo"))
	require.NoError(t, err)
	require.True(t, exists, "File foo should exist.")

	// Check that commit doesn't exist
	exists, err = btrfs.FileExists(path.Join(outPrefix, "error", "commit"))
	require.NoError(t, err)
	require.False(t, exists, "Commit \"commit\" should not get created when a command fails.")
}

// TestRecover runs a pipeline with an error. Then fixes the pipeline to not
// include an error and reruns it.
func TestRecover(t *testing.T) {
	t.Parallel()
	inRepo := "TestRecover_in"
	require.NoError(t, btrfs.Init(inRepo))
	outPrefix := "TestRecover_out"

	// Create the Pachfile
	require.NoError(t, btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "recover"), []byte(`
image ubuntu

run touch /out/foo
run touch /out/bar && cp /in/foo /out/bar
`)))
	// Last line should fail here.

	// Commit to the inRepo
	require.NoError(t, btrfs.Commit(inRepo, "commit1", "master"))

	// Run the pipelines
	err := RunPipelines("pipeline", inRepo, outPrefix, "commit1", "master", "0-1", etcache.NewCache())
	require.Error(t, err, "Running pipeline should error.")

	// Fix the Pachfile
	require.NoError(t, btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "recover"), []byte(`
image ubuntu

run touch /out/foo
run touch /out/bar
`)))

	// Commit to the inRepo
	require.NoError(t, btrfs.Commit(inRepo, "commit2", "master"))

	// Run the pipelines
	err = RunPipelines("pipeline", inRepo, outPrefix, "commit2", "master", "0-1", etcache.NewCache())
	// this time the pipelines should not err
	require.NoError(t, err)

	// These are the most important 2 checks:

	// If this one fails it means that dirty state isn't properly saved
	checkExists(t, path.Join(outPrefix, "recover", "commit1-fail/bar"))
	// If this one fails it means that dirty state isn't properly cleared
	checkNoExists(t, path.Join(outPrefix, "recover", "commit2-0/bar"))

	// These commits are mostly covered by other tests
	checkExists(t, path.Join(outPrefix, "recover", "commit1-fail/foo"))
	checkExists(t, path.Join(outPrefix, "recover", "commit1-0/foo"))
	checkNoExists(t, path.Join(outPrefix, "recover", "commit1-1"))
	checkNoExists(t, path.Join(outPrefix, "recover", "commit1"))
	checkExists(t, path.Join(outPrefix, "recover", "commit2-0/foo"))
	checkExists(t, path.Join(outPrefix, "recover", "commit2-1/foo"))
	checkExists(t, path.Join(outPrefix, "recover", "commit2-1/bar"))
	checkExists(t, path.Join(outPrefix, "recover", "commit2/foo"))
	checkExists(t, path.Join(outPrefix, "recover", "commit2/bar"))
}

func TestCancel(t *testing.T) {
	t.Parallel()
	inRepo := "TestCancel_in"
	require.NoError(t, btrfs.Init(inRepo))
	outPrefix := "TestCancel_out"

	// Create the Pachfile
	require.NoError(t, btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "cancel"), []byte(`
image ubuntu

run sleep 100
`)))
	require.NoError(t, btrfs.Commit(inRepo, "commit", "master"))

	r := NewRunner("pipeline", inRepo, outPrefix, "commit", "master", "0-1", etcache.NewCache())
	go func() {
		err := r.Run()
		require.Equal(t, ErrCancelled, err)
	}()

	// This is just to make sure we don't trigger the early exit case in Run
	// and actually exercise the code.
	time.Sleep(time.Second * 2)
	require.NoError(t, r.Cancel())
}

// TestWrap tests a simple pipeline that uses line wrapping in it's Pachfile
func TestWrap(t *testing.T) {
	t.Parallel()
	outRepo := "TestWrap_out"
	require.NoError(t, btrfs.Init(outRepo))
	pipeline := newPipeline("output", "", outRepo, "commit", "master", "0-1", "", etcache.NewCache())
	pachfile := `
image ubuntu

# touch foo and bar
run touch /out/foo \
          /out/bar
`
	err := pipeline.runPachFile(strings.NewReader(pachfile))
	require.NoError(t, err)

	exists, err := btrfs.FileExists(path.Join(outRepo, "commit", "foo"))
	require.NoError(t, err)
	require.True(t, exists, "File `foo` doesn't exist when it should.")

	exists, err = btrfs.FileExists(path.Join(outRepo, "commit", "bar"))
	require.NoError(t, err)
	require.True(t, exists, "File `bar` doesn't exist when it should.")
}

func TestDependency(t *testing.T) {
	t.Parallel()
	inRepo := "TestDependency_in"
	require.NoError(t, btrfs.Init(inRepo))
	p1 := `
image ubuntu

run echo foo >/out/foo
`
	require.NoError(t, btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "p1"), []byte(p1)))
	p2 := `
image ubuntu

input pps://p1

run cp /in/p1/foo /out/foo
`
	require.NoError(t, btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "p2"), []byte(p2)))
	require.NoError(t, btrfs.Commit(inRepo, "commit", "master"))

	outPrefix := "TestDependency"
	runner := NewRunner("pipeline", inRepo, outPrefix, "commit", "master", "0-1", etcache.NewCache())
	require.NoError(t, runner.Run())

	res, err := btrfs.ReadFile(path.Join(outPrefix, "p2", "commit", "foo"))
	require.NoError(t, err)
	require.Equal(t, "foo\n", string(res))
}

func TestRunnerInputs(t *testing.T) {
	t.Parallel()
	inRepo := "TestRunnerInputs_in"
	require.NoError(t, btrfs.Init(inRepo))
	p1 := `
image ubuntu

input foo
input bar
`
	require.NoError(t, btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "p1"), []byte(p1)))
	p2 := `
image ubuntu

input fizz
input buzz
`
	require.NoError(t, btrfs.WriteFile(path.Join(inRepo, "master", "pipeline", "p2"), []byte(p2)))
	require.NoError(t, btrfs.Commit(inRepo, "commit", "master"))

	outPrefix := "TestRunnerInputs"
	runner := NewRunner("pipeline", inRepo, outPrefix, "commit", "master", "0-1", etcache.NewCache())
	inputs, err := runner.Inputs()
	require.NoError(t, err)
	require.Equal(t, []string{"foo", "bar", "fizz", "buzz"}, inputs)
}

// TestInject tests that s3 injections works
func TestInject(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip()
	}
	outRepo := "TestInject_out"
	require.NoError(t, btrfs.Init(outRepo))
	pipeline := newPipeline("output", "", outRepo, "commit", "master", "0-1", "", etcache.NewCache())
	require.NoError(t, pipeline.inject("s3://pachyderm-test/pipeline", true))
	require.NoError(t, pipeline.finish())
	res, err := btrfs.ReadFile(path.Join(outRepo, "commit", "file"))
	require.NoError(t, err)
	require.Equal(t, "foo\n", string(res))
}

func TestExternalOutput(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		//t.Skip()
	}
	outRepo := "Tests3Output_out"
	require.NoError(t, btrfs.Init(outRepo))
	pipeline := newPipeline("output", "", outRepo, "commit", "master", "0-1", "", etcache.NewCache())
	require.NoError(t, pipeline.output("s3://pachyderm-test/pipeline-out"))
	pachfile := `
image ubuntu
output s3://pachyderm-test/pipeline-output

run echo foo >/out/foo
`
	require.NoError(t, pipeline.runPachFile(strings.NewReader(pachfile)))
}

func newTestPipeline(
	t *testing.T,
	repoPrefix string,
	commit string,
	branch string,
	shard string,
	init bool,
) *pipeline {
	if init {
		require.NoError(t, btrfs.Init(repoPrefix+"-in"))
		require.NoError(t, btrfs.Init(repoPrefix+"-out"))
	}
	return newPipeline(
		"pipeline",
		repoPrefix+"-in",
		repoPrefix+"-out",
		commit,
		branch,
		shard,
		"pipelineDir",
		etcache.NewCache(),
	)
}

func checkNoExists(t *testing.T, name string) {
	exists, err := btrfs.FileExists(name)
	require.NoError(t, err)
	require.False(t, exists)
}

func checkExists(t *testing.T, name string) {
	exists, err := btrfs.FileExists(name)
	require.NoError(t, err)
	require.True(t, exists)
}
