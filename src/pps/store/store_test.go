package store

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/satori/go.uuid"
)

func TestBasicRethink(t *testing.T) {
	runTestRethink(t, testBasic)
}

func testBasic(t *testing.T, client Client) {
	pipelineRun := &pps.PipelineRun{
		Id: "id",
		PipelineSource: &pps.PipelineSource{
			GithubPipelineSource: &pps.GithubPipelineSource{
				ContextDir:  "dir",
				User:        "user",
				Repository:  "repo",
				Branch:      "branch",
				AccessToken: "token",
			},
		},
	}
	require.NoError(t, client.AddPipelineRun(pipelineRun))
	pipelineRunResponse, err := client.GetPipelineRun("id")
	require.NoError(t, err)
	require.Equal(t, pipelineRun, pipelineRunResponse)
	require.NoError(t, client.AddPipelineRunStatus("id", pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_NONE))
	pipelineRunStatusResponse, err := client.GetAllPipelineRunStatuses("id")
	require.NoError(t, err)
	require.Equal(t, pipelineRunStatusResponse[0].PipelineRunStatusType, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_NONE)
	require.NoError(t, client.AddPipelineRunStatus("id", pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS))
	pipelineRunStatusResponse, err = client.GetAllPipelineRunStatuses("id")
	require.NoError(t, err)
	require.Equal(t, pipelineRunStatusResponse[0].PipelineRunStatusType, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS)

	require.NoError(t, client.AddPipelineRunContainers(&pps.PipelineRunContainer{PipelineRunId: "id", ContainerId: "container", Node: "node"}))
	containerIDs, err := client.GetPipelineRunContainers("id")
	require.NoError(t, err)
	require.Equal(t, []*pps.PipelineRunContainer{&pps.PipelineRunContainer{PipelineRunId: "id", ContainerId: "container", Node: "node"}}, containerIDs)
}

func runTestRethink(t *testing.T, testFunc func(*testing.T, Client)) {
	client, err := getRethinkSession()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, client.Close())
	}()
	testFunc(t, client)
}

func getRethinkSession() (Client, error) {
	address, err := getRethinkAddress()
	if err != nil {
		return nil, err
	}
	databaseName := strings.Replace(uuid.NewV4().String(), "-", "", -1)
	if err := InitDBs(address, databaseName); err != nil {
		return nil, err
	}
	return NewRethinkClient(address, databaseName)
}

func getRethinkAddress() (string, error) {
	rethinkAddr := os.Getenv("RETHINK_PORT_28015_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("RETHINK_PORT_28015_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:28015", rethinkAddr), nil
}
