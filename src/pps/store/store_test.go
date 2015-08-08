package store

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/stretchr/testify/require"
)

func TestBasicRethink(t *testing.T) {
	runTestRethink(t, testBasic)
}

func TestBasicInMem(t *testing.T) {
	runTestInMem(t, testBasic)
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
	pipelineRunStatusResponse, err := client.GetPipelineRunStatusLatest("id")
	require.NoError(t, err)
	require.Equal(t, pipelineRunStatusResponse.PipelineRunStatusType, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_NONE)
	require.NoError(t, client.AddPipelineRunStatus("id", pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS))
	pipelineRunStatusResponse, err = client.GetPipelineRunStatusLatest("id")
	require.NoError(t, err)
	require.Equal(t, pipelineRunStatusResponse.PipelineRunStatusType, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS)

	require.NoError(t, client.AddPipelineRunContainerIDs("id", "container"))
	containerIDs, err := client.GetPipelineRunContainers("id")
	require.NoError(t, err)
	require.Equal(t, []*PipelineContainer{&PipelineContainer{"id", "container"}}, containerIDs)
}

func runTestInMem(t *testing.T, testFunc func(*testing.T, Client)) {
	testFunc(t, NewInMemoryClient())
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
	databaseName := strings.Replace(common.NewUUID(), "-", "", -1)
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
