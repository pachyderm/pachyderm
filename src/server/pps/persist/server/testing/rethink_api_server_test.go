package testing

import (
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"
	"golang.org/x/net/context"
)

func TestSubscribePipelineInfos(t *testing.T) {
	RunTestWithPersistClient(t, testSubscribePipelineInfos)
}

func testSubscribePipelineInfos(t *testing.T, client persist.APIClient) {
	gotFirstChange := make(chan bool)
	go func() {
		_, err := client.CreatePipelineInfo(
			context.Background(),
			&persist.PipelineInfo{
				PipelineName: "foo",
				Transform: &ppsclient.Transform{
					Cmd: []string{"foo"},
				},
			})
		require.NoError(t, err)
		_, err = client.UpdatePipelineState(
			context.Background(),
			&persist.UpdatePipelineStateRequest{
				PipelineName: "foo",
				State:        ppsclient.PipelineState_PIPELINE_RUNNING,
			},
		)
		require.NoError(t, err)
		<-gotFirstChange
		_, err = client.UpdatePipelineInfo(
			context.Background(),
			&persist.PipelineInfo{
				PipelineName: "foo",
				Transform: &ppsclient.Transform{
					Cmd: []string{"bar"},
				},
			},
		)
		require.NoError(t, err)
		_, err = client.UpdatePipelineState(
			context.Background(),
			&persist.UpdatePipelineStateRequest{
				PipelineName: "foo",
				State:        ppsclient.PipelineState_PIPELINE_RUNNING,
			},
		)
		require.NoError(t, err)
	}()
	subscribeCtx, cancel := context.WithCancel(context.Background())
	pipelineClient, err := client.SubscribePipelineInfos(
		subscribeCtx,
		&persist.SubscribePipelineInfosRequest{
			IncludeInitial: true,
		})
	require.NoError(t, err)
	pipelineChange, err := pipelineClient.Recv()
	require.NoError(t, err)
	require.Equal(t, persist.ChangeType_CREATE, pipelineChange.Type)
	require.Equal(t, []string{"foo"}, pipelineChange.Pipeline.Transform.Cmd)
	close(gotFirstChange)
	pipelineChange, err = pipelineClient.Recv()
	require.NoError(t, err)
	require.Equal(t, persist.ChangeType_UPDATE, pipelineChange.Type)
	require.Equal(t, []string{"bar"}, pipelineChange.Pipeline.Transform.Cmd)
	go func() {
		time.Sleep(10)
		cancel()
	}()
	// We expect there to not be any more pipeline updates since we only did 2
	// operations on it.
	pipelineChange, err = pipelineClient.Recv()
	require.YesError(t, err, "Got superflous change: %+v", pipelineChange)
}
