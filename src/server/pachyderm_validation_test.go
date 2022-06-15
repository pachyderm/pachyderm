package server

import (
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// Make sure that pipeline validation requires:
// - No dash in pipeline name
// - Input must have branch and glob
func TestInvalidCreatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)

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
		client.NewPFSInputOpts("out", dataRepo, "", "/*", "", "", false, false, nil),
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
		client.NewPFSInputOpts("input", dataRepo, "", "", "", "", false, false, nil),
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
	c, _ := minikubetestenv.AcquireCluster(t)
	pipelineName := tu.UniqueString("pipeline")
	require.YesError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInputOpts("whatever", "nonexistent", "", "/*", "", "", false, false, nil),
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
	c, _ := minikubetestenv.AcquireCluster(t)

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
