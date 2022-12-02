//go:build k8s

package server

import (
	"path"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
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
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	c := env.PachClient
	// Set up repo
	dataRepo := tu.UniqueString("TestDuplicatedJob_data")
	require.NoError(t, c.CreateProjectRepo(pfs.DefaultProjectName, dataRepo))

	projectName := tu.UniqueString("_prj-")
	err := c.CreateProject(projectName)
	require.NoError(t, err)

	pipelineName := tu.UniqueString("pipeline")
	cmd := []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"}

	// Create pipeline with input named "out"
	err = c.CreateProjectPipeline(projectName,
		pipelineName,
		"",
		cmd,
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewProjectPFSInputOpts("out", projectName, dataRepo, "", "/*", "", "", false, false, nil),
		"master",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "out", err.Error())

	// Create pipeline with no glob
	err = c.CreateProjectPipeline(projectName,
		pipelineName,
		"",
		cmd,
		nil,
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewProjectPFSInputOpts("input", projectName, dataRepo, "", "", "", "", false, false, nil),
		"master",
		false,
	)
	require.YesError(t, err)
	require.Matches(t, "glob", err.Error())
}

// Make sure that pipeline validation checks that all inputs exist
func TestPipelineThatUseNonexistentInputs(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	c := env.PachClient
	pipelineName := tu.UniqueString("pipeline")
	require.YesError(t, c.CreateProjectPipeline(pfs.DefaultProjectName,
		pipelineName,
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewProjectPFSInputOpts("whatever", pfs.DefaultProjectName, "nonexistent", "", "/*", "", "", false, false, nil),
		"master",
		false,
	))
}

// Make sure that pipeline validation checks that all inputs exist
func TestPipelineNamesThatContainUnderscoresAndHyphens(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	c := env.PachClient

	projectName := tu.UniqueString("_prj-")
	err := c.CreateProject(projectName)
	require.NoError(t, err)

	dataRepo := tu.UniqueString("TestPipelineNamesThatContainUnderscoresAndHyphens")
	require.NoError(t, c.CreateProjectRepo(projectName, dataRepo))

	require.NoError(t, c.CreateProjectPipeline(projectName,
		tu.UniqueString("pipeline-hyphen"),
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewProjectPFSInput(projectName, dataRepo, "/*"),
		"",
		false,
	))

	require.NoError(t, c.CreateProjectPipeline(projectName,
		tu.UniqueString("pipeline_underscore"),
		"",
		[]string{"bash"},
		[]string{""},
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewProjectPFSInput(projectName, dataRepo, "/*"),
		"",
		false,
	))
}

func TestProjectNameValidation(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	c := env.PachClient
	badFormatErr := "only alphanumeric characters"
	testCases := []struct {
		projectName string
		errMatch    string // "" means no error
	}{
		{tu.UniqueString("PROJECT"), ""},
		{tu.UniqueString("0123456789"), ""},
		{tu.UniqueString("_"), ""},
		{tu.UniqueString("project-"), ""},
		{"default", "already exists"},
		{tu.UniqueString("!project"), badFormatErr},
		{tu.UniqueString("\""), badFormatErr},
		{tu.UniqueString("\\"), badFormatErr},
		{tu.UniqueString("'"), badFormatErr},
		{tu.UniqueString("[]{}"), badFormatErr},
		{tu.UniqueString("|"), badFormatErr},
		{tu.UniqueString("new->project"), badFormatErr},
		{tu.UniqueString("project?"), badFormatErr},
		{tu.UniqueString("project:1"), badFormatErr},
		{tu.UniqueString("project;"), badFormatErr},
		{tu.UniqueString("project."), badFormatErr},
	}
	for _, testCase := range testCases {
		t.Run(testCase.projectName, func(t *testing.T) {
			t.Parallel()
			err := c.CreateProject(testCase.projectName)
			if testCase.errMatch != "" {
				require.YesError(t, err)
				require.Matches(t, testCase.errMatch, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}

}
