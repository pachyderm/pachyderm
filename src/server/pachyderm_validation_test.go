package server

import (
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

// Make sure that pipeline validation requires:
// - No dash in pipeline name
// - Input must have branch and glob
func TestInvalidCreatePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c := getPachClient(t)

	// Set up repo
	dataRepo := uniqueString("TestDuplicatedJob_data")
	require.NoError(t, c.CreateRepo(dataRepo))

	pipelineName := uniqueString("pipeline")
	cmd := []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"}

	// Create pipeline named "out"
	err := c.CreatePipeline(
		pipelineName,
		"",
		cmd,
		nil,
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Name:   "out",
			Repo:   &pfs.Repo{Name: dataRepo},
			Branch: "master",
			Glob:   "/*",
		}},
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
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{{
			Name:   "input",
			Repo:   &pfs.Repo{Name: dataRepo},
			Branch: "master",
		}},
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
	c := getPachClient(t)
	pipelineName := uniqueString("pipeline")
	require.YesError(t, c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Strategy: pps.ParallelismSpec_CONSTANT,
			Constant: 1,
		},
		[]*pps.PipelineInput{
			{
				Name:   "whatever",
				Branch: "master",
				Glob:   "/*",
				Repo:   &pfs.Repo{Name: "nonexistent"},
			},
		},
		"master",
		false,
	))
}
