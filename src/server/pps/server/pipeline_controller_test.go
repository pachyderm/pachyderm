package server

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	logrus "github.com/sirupsen/logrus"
)

func ppsMasterHandles(t *testing.T) (*mockStateDriver, *mockInfraDriver, *testpachd.MockPachd) {
	sDriver := newMockStateDriver()
	iDriver := newMockInfraDriver()
	mockEnv := testpachd.NewMockEnv(t)
	env := Env{
		BackgroundContext: context.Background(),
		Logger:            logrus.StandardLogger(),
		EtcdClient:        mockEnv.EtcdClient,
		GetPachClient: func(context.Context) *client.APIClient {
			return mockEnv.PachClient
		},
		Config:     newConfig(t),
		DB:         nil,
		KubeClient: nil,
		AuthServer: nil,
		PFSServer:  nil,
		TxnEnv:     nil,
	}
	ctx, cf := context.WithCancel(context.Background())
	t.Cleanup(func() { cf() })
	master := newMaster(ctx, env, env.EtcdPrefix, iDriver, sDriver)
	go master.run()
	return sDriver, iDriver, mockEnv.MockPachd
}

func waitForPipelineState(t testing.TB, stateDriver *mockStateDriver, pipeline string, state pps.PipelineState) {
	require.NoErrorWithinT(t, 10*time.Second, func() error {
		return backoff.Retry(func() error {
			if stateDriver.pipelines[pipeline].State == state {
				return nil
			}
			return errors.New("change hasn't reflected")
		}, backoff.NewTestingBackOff())
	})
}

func TestPipelineController(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t)
	mockPachd.PFS.SubscribeCommit.Use(func(*pfs.SubscribeCommitRequest, pfs.API_SubscribeCommitServer) error {
		return nil
	})
	name := tu.UniqueString("TestPipelineController_")
	stateDriver.upsertPipeline(&pps.PipelineInfo{
		Pipeline: client.NewPipeline(name),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	})
	waitForPipelineState(t, stateDriver, name, pps.PipelineState_PIPELINE_RUNNING)
	require.ElementsEqual(t, []pps.PipelineState{
		pps.PipelineState_PIPELINE_STARTING,
		pps.PipelineState_PIPELINE_RUNNING,
	}, stateDriver.states[name])
	require.Equal(t, 1, len(infraDriver.rcs))
	rc, err := infraDriver.ReadReplicationController(context.Background(), stateDriver.pipelines[name])
	require.NoError(t, err)
	require.True(t, rcIsFresh(stateDriver.pipelines[name], &rc.Items[0]))

}
