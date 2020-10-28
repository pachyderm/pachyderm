package server

import (
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

// Make sure that pipeline validation requires:
// - No dash in pipeline name
// - Input must have branch and glob
func TestInvalidCreatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := tu.GetPachClient(t)

	// Set up repo
	dataRepo := tu.UniqueString("TestDuplicatedJob_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	pipelineName := tu.UniqueString("pipeline")
	cmd := []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"}

	// Create pipeline with input named "out"
	err := c.CreatePipeline(
		pipelineName,
		"",
		cmd,
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInputOpts("out", dataRepo, "", "/*", "", "", false),
		"master",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "out", err.Error())

	// Create pipeline with no glob
	err = c.CreatePipeline(
		pipelineName,
		"",
		cmd,
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInputOpts("input", dataRepo, "", "", "", "", false),
		"master",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "glob", err.Error())
}

// Make sure that pipeline validation checks that all inputs exist
func TestPipelineThatUseNonexistentInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := tu.GetPachClient(t)
	pipelineName := tu.UniqueString("pipeline")
	require.YesError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInputOpts("whatever", "nonexistent", "", "/*", "", "", false),
		"master",
		false,
	))
}

// Make sure that pipeline validation checks that all inputs exist
func TestPipelineNamesThatContainUnderscoresAndHyphens(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := tu.GetPachClient(t)

	dataRepo := tu.UniqueString("TestPipelineNamesThatContainUnderscoresAndHyphens")
	require.NoError(t, c.CreateRepo(dataRepo))

	require.NoError(t, c.CreatePipeline(
		tu.UniqueString("pipeline-hyphen"),
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	))

	require.NoError(t, c.CreatePipeline(
		tu.UniqueString("pipeline_underscore"),
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(dataRepo, "/*"),
		"",
		false,
	))
}

func TestPipelineInvalidParallelism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := tu.GetPachClient(t)

	// Set up repo
	dataRepo := tu.UniqueString("TestPipelineInvalidParallelism")
	require.NoError(t, c.CreateRepo(dataRepo))

	// Create pipeline named "out"
	err := c.CreatePipeline(
		"invalid-parallelism-pipeline",
		"",
		[]string{"bash", "-c"},
		[]string{"echo hello"},
		&pps.ParallelismSpec{
			Constant:    1,
			Coefficient: 1.0,
		},
		client.NewPFSInput(dataRepo, "/*"),
		"master",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "parallelism", err.Error())
}
