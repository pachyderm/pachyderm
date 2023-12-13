package server

import (
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// Make sure that pipeline validation requires:
// - No dash in pipeline name
// - Input must have branch and glob
func TestInvalidCreatePipeline(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	c := env.PachClient

	projectName := tu.UniqueString("prj-")
	err := c.CreateProject(projectName)
	require.NoError(t, err)
	// Set up repo
	dataRepo := tu.UniqueString("TestDuplicatedJob_data")
	require.NoError(t, c.CreateRepo(projectName, dataRepo))

	pipelineName := tu.UniqueString("pipeline")
	cmd := []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"}

	// Create pipeline with input named "out"
	err = c.CreatePipeline(projectName,
		pipelineName,
		"",
		cmd,
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInputOpts("out", projectName, dataRepo, "", "/*", "", "", false, false, nil),
		"master",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "out", err.Error())

	// Create pipeline with no glob
	err = c.CreatePipeline(projectName,
		pipelineName,
		"",
		cmd,
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInputOpts("input", projectName, dataRepo, "", "", "", "", false, false, nil),
		"master",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "glob", err.Error())

	// Create pipeline with input commit
	err = c.CreatePipeline(projectName,
		pipelineName,
		"",
		cmd,
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		&pps.Input{
			Pfs: &pps.PFSInput{
				Name:    "input",
				Project: projectName,
				Repo:    dataRepo,
				Glob:    "/*",
				Commit:  "not_supported",
			},
		},
		"master",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "input cannot come from a commit", err.Error())
}

// Make sure that pipeline validation checks that all inputs exist
func TestPipelineThatUseNonexistentInputs(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	c := env.PachClient
	pipelineName := tu.UniqueString("pipeline")
	require.YesError(t, c.CreatePipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInputOpts("whatever", pfs.DefaultProjectName, "nonexistent", "", "/*", "", "", false, false, nil),
		"master",
		false,
	))
}

// Make sure that pipeline validation checks that all inputs exist
func TestPipelineNamesThatContainUnderscoresAndHyphens(t *testing.T) {
	t.Parallel()
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	c := env.PachClient

	projectName := tu.UniqueString("prj-")
	err := c.CreateProject(projectName)
	require.NoError(t, err)

	dataRepo := tu.UniqueString("TestPipelineNamesThatContainUnderscoresAndHyphens")
	require.NoError(t, c.CreateRepo(projectName, dataRepo))

	require.NoError(t, c.CreatePipeline(projectName,
		tu.UniqueString("pipeline-hyphen"),
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(projectName, dataRepo, "/*"),
		"",
		false,
	))

	require.NoError(t, c.CreatePipeline(projectName,
		tu.UniqueString("pipeline_underscore"),
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewPFSInput(projectName, dataRepo, "/*"),
		"",
		false,
	))
}
