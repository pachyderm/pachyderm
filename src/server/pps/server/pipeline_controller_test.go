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
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/task"
	logrus "github.com/sirupsen/logrus"
)

type pipelineTest struct {
	pipeline       string
	assertWhen     []pps.PipelineState
	expectedStates []pps.PipelineState
}

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
		Config:      newConfig(t),
		TaskService: nil,
		DB:          nil,
		KubeClient:  nil,
		AuthServer:  nil,
		PFSServer:   nil,
		TxnEnv:      nil,
	}
	ctx, cf := context.WithCancel(context.Background())
	t.Cleanup(func() { cf() })
	master := newMaster(ctx, env, env.EtcdPrefix, iDriver, sDriver)
	go master.run()
	return sDriver, iDriver, mockEnv.MockPachd
}

func waitForPipelineStates(t testing.TB, stateDriver *mockStateDriver, pipeline string, states []pps.PipelineState) {
	require.NoErrorWithinT(t, 10*time.Second, func() error {
		return backoff.Retry(func() error {
			actualStates := stateDriver.states[pipeline]
			for i := range actualStates {
				k := i
				for j, expected := range states {
					if k >= len(actualStates) || actualStates[k] != expected {
						break
					}
					k++
					if j == len(states)-1 {
						return nil
					}
				}
			}
			return errors.New("change hasn't reflected")
		}, backoff.NewTestingBackOff())
	})
}

func mockAutoscaling(mockPachd *testpachd.MockPachd, taskCount, commitCount int) {
	mockPachd.PFS.ListTask.Use(func(req *task.ListTaskRequest, srv pfs.API_ListTaskServer) error {
		for i := 0; i < taskCount; i++ {
			srv.Send(&task.TaskInfo{})
		}
		return nil
	})
	mockPachd.PFS.InspectCommit.Use(func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
		return &pfs.CommitInfo{}, nil
	})
	mockPachd.PFS.SubscribeCommit.Use(func(req *pfs.SubscribeCommitRequest, server pfs.API_SubscribeCommitServer) error {
		for i := 0; i < commitCount; i++ {
			server.Send(&pfs.CommitInfo{
				Commit: client.NewCommit(req.Repo.Name, "", uuid.NewWithoutDashes()),
			})
		}
		// keep the subscribe stream open longer
		time.Sleep(3 * time.Second)
		return nil
	})
}

func validate(t testing.TB, sDriver *mockStateDriver, iDriver *mockInfraDriver, tests []pipelineTest) {
	for _, test := range tests {
		waitForPipelineStates(t, sDriver, test.pipeline, test.assertWhen)
		require.ElementsEqual(t, test.expectedStates, sDriver.states[test.pipeline])
		require.Equal(t, 1, len(iDriver.rcs))
		rc, err := iDriver.ReadReplicationController(context.Background(), sDriver.pipelines[test.pipeline])
		require.NoError(t, err)
		require.True(t, rcIsFresh(sDriver.pipelines[test.pipeline], &rc.Items[0]))
	}
}

func TestBasic(t *testing.T) {
	stateDriver, infraDriver, _ := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	stateDriver.upsertPipeline(&pps.PipelineInfo{
		Pipeline: client.NewPipeline(pipeline),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	})
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline:   pipeline,
			assertWhen: []pps.PipelineState{pps.PipelineState_PIPELINE_RUNNING},
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[pipeline][mockInfraOp_CREATE])
}

func TestDeletePipeline(t *testing.T) {
	stateDriver, infraDriver, _ := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	pi := &pps.PipelineInfo{
		Pipeline: client.NewPipeline(pipeline),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	}
	stateDriver.upsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline:   pipeline,
			assertWhen: []pps.PipelineState{pps.PipelineState_PIPELINE_RUNNING},
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[pipeline][mockInfraOp_CREATE])

	stateDriver.reset()
	stateDriver.pushWatchEvent(pi, watch.EventDelete)
	require.NoErrorWithinT(t, 5*time.Second, func() error {
		return backoff.Retry(func() error {
			if infraDriver.calls[pipeline][mockInfraOp_DELETE] == 1 {
				return nil
			}
			return errors.New("change hasn't reflected")
		}, backoff.NewTestingBackOff())
	})

}

func TestAutoscalingBasic(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	mockAutoscaling(mockPachd, 1, 1)
	stateDriver.upsertPipeline(&pps.PipelineInfo{
		Pipeline: client.NewPipeline(pipeline),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details: &pps.PipelineInfo_Details{
			Autoscaling: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
		},
		Version: 1,
	})
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			assertWhen: []pps.PipelineState{
				pps.PipelineState_PIPELINE_RUNNING,
			},
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[pipeline][mockInfraOp_CREATE])
}

func TestAutoscalingNoCommits(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	mockAutoscaling(mockPachd, 0, 0)
	stateDriver.upsertPipeline(&pps.PipelineInfo{
		Pipeline: client.NewPipeline(pipeline),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details: &pps.PipelineInfo_Details{
			Autoscaling: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
		},
		Version: 1,
	})
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			assertWhen: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_STANDBY,
			},
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[pipeline][mockInfraOp_CREATE])
}
