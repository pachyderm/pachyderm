package store

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	runTest(t, testBasic)
}

func testBasic(t *testing.T, client Client) {
	run := &pps.PipelineRun{
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
	require.NoError(t, client.AddPipelineRun(run))
	runResp, err := client.GetPipelineRun("id")
	require.NoError(t, err)
	require.Equal(t, run, runResp)
}

func runTest(t *testing.T, testFunc func(*testing.T, Client)) {
	client, err := getRethinkSession()
	require.NoError(t, err)
	testFunc(t, client)
}

func getRethinkSession() (Client, error) {
	address, err := getRethinkAddress()
	if err != nil {
		return nil, err
	}
	return NewRethinkClient(address)
}

func getRethinkAddress() (string, error) {
	rethinkAddr := os.Getenv("RETHINK_PORT_28015_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("RETHINK_PORT_28015_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:28015", rethinkAddr), nil
}
