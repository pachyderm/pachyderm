package server

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"go.pachyderm.com/pachyderm/src/pkg/require"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"golang.org/x/net/context"

	"github.com/satori/go.uuid"
)

func TestBasicRethink(t *testing.T) {
	runTestRethink(t, testBasicRethink)
}

func testBasicRethink(t *testing.T, apiServer persist.APIServer) {
	jobInfo, err := apiServer.CreateJobInfo(
		context.Background(),
		&persist.JobInfo{
			Spec: &persist.JobInfo_PipelineName{
				PipelineName: "foo",
			},
		},
	)
	require.NoError(t, err)
	getJobInfo, err := apiServer.GetJobInfo(
		context.Background(),
		&pps.Job{
			Id: jobInfo.JobId,
		},
	)
	require.NoError(t, err)
	require.Equal(t, jobInfo.JobId, getJobInfo.JobId)
	require.Equal(t, "foo", getJobInfo.GetPipelineName())
	require.NotNil(t, getJobInfo.CreatedAt)
}

func runTestRethink(t *testing.T, testFunc func(*testing.T, persist.APIServer)) {
	apiServer, err := getTestRethinkAPIServer()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, apiServer.Close())
	}()
	testFunc(t, apiServer)
}

func getTestRethinkAPIServer() (*rethinkAPIServer, error) {
	address, err := getTestRethinkAddress()
	if err != nil {
		return nil, err
	}
	databaseName := strings.Replace(uuid.NewV4().String(), "-", "", -1)
	if err := InitDBs(address, databaseName); err != nil {
		return nil, err
	}
	return newRethinkAPIServer(address, databaseName)
}

func getTestRethinkAddress() (string, error) {
	rethinkAddr := os.Getenv("RETHINK_PORT_28015_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("RETHINK_PORT_28015_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:28015", rethinkAddr), nil
}
