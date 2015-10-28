package testing

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"golang.org/x/net/context"
)

func TestBasicRethink(t *testing.T) {
	RunTestWithRethinkAPIServer(t, testBasicRethink)
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
