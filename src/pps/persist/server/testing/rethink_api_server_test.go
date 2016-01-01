package testing

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"golang.org/x/net/context"
)

func TestBasicRethink(t *testing.T) {
	RunTestWithRethinkAPIServer(t, testBasicRethink)
}

func TestBlock(t *testing.T) {
	RunTestWithRethinkAPIServer(t, testBlock)
}

func testBasicRethink(t *testing.T, apiServer persist.APIServer) {
	jobInfo, err := apiServer.CreateJobInfo(
		context.Background(),
		&persist.JobInfo{
			PipelineName: "foo",
		},
	)
	require.NoError(t, err)
	getJobInfo, err := apiServer.InspectJob(
		context.Background(),
		&pps.InspectJobRequest{
			Job: &pps.Job{
				Id: jobInfo.JobId,
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, jobInfo.JobId, getJobInfo.JobId)
	require.Equal(t, "foo", getJobInfo.PipelineName)
}

func testBlock(t *testing.T, apiServer persist.APIServer) {
	jobInfo, err := apiServer.CreateJobInfo(context.Background(), &persist.JobInfo{})
	require.NoError(t, err)
	jobID := jobInfo.JobId
	go func() {
		_, err := apiServer.CreateJobOutput(
			context.Background(),
			&persist.JobOutput{
				JobId:        jobID,
				OutputCommit: pfsutil.NewCommit("foo", "bar"),
			})
		require.NoError(t, err)
		_, err = apiServer.CreateJobState(
			context.Background(),
			&persist.JobState{
				JobId: jobID,
				State: pps.JobState_JOB_STATE_SUCCESS,
			})
		require.NoError(t, err)
	}()
	_, err = apiServer.InspectJob(
		context.Background(),
		&pps.InspectJobRequest{
			Job:         &pps.Job{Id: jobID},
			BlockOutput: true,
			BlockState:  true,
		},
	)
	require.NoError(t, err)
}
