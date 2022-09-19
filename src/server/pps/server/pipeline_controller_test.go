package server_test

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/server/pps/server"

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
	expectedStates []pps.PipelineState
}

func ppsMasterHandles(t *testing.T) (*server.MockStateDriver, *server.MockInfraDriver, *testpachd.MockPachd) {
	sDriver := server.NewMockStateDriver()
	iDriver := server.NewMockInfraDriver()
	mockEnv := testpachd.NewMockEnv(t)
	env := server.Env{
		BackgroundContext: context.Background(),
		Logger:            logrus.StandardLogger(),
		EtcdClient:        mockEnv.EtcdClient,
		GetPachClient: func(ctx context.Context) *client.APIClient {
			return mockEnv.PachClient.WithCtx(ctx)
		},
		Config:      *serviceenv.ConfigFromOptions(),
		TaskService: nil,
		DB:          nil,
		KubeClient:  nil,
		AuthServer:  nil,
		PFSServer:   nil,
		TxnEnv:      nil,
	}
	ctx, cf := context.WithCancel(context.Background())
	t.Cleanup(func() { cf() })
	server.MockMaster(ctx, env, iDriver, sDriver)
	return sDriver, iDriver, mockEnv.MockPachd
}

func waitForPipelineState(t testing.TB, stateDriver *server.MockStateDriver, pipeline string, state pps.PipelineState) {
	require.NoErrorWithinT(t, 10*time.Second, func() error {
		return backoff.Retry(func() error {
			actualStates := stateDriver.States[pipeline]
			for _, s := range actualStates {
				if s == state {
					return nil
				}
			}
			return errors.New("change hasn't reflected")
		}, backoff.NewTestingBackOff())
	})
}

func mockJobRunning(mockPachd *testpachd.MockPachd, taskCount, commitCount int) (done chan struct{}) {
	mockPachd.PPS.ListTask.Use(func(_ *task.ListTaskRequest, srv pps.API_ListTaskServer) error {
		for i := 0; i < taskCount; i++ {
			if err := srv.Send(&task.TaskInfo{State: task.State_RUNNING}); err != nil {
				return errors.EnsureStack(err)
			}
		}
		return nil
	})
	mockPachd.PFS.InspectCommit.Use(func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
		return &pfs.CommitInfo{}, nil
	})
	testDone := make(chan struct{})
	mockPachd.PFS.SubscribeCommit.Use(func(req *pfs.SubscribeCommitRequest, server pfs.API_SubscribeCommitServer) error {
		for i := 0; i < commitCount; i++ {
			if err := server.Send(&pfs.CommitInfo{Commit: client.NewCommit(req.Repo.Name, "", uuid.NewWithoutDashes())}); err != nil {
				return errors.EnsureStack(err)
			}
		}
		<-testDone
		return nil
	})
	return testDone
}

func validate(t testing.TB, sDriver *server.MockStateDriver, iDriver *server.MockInfraDriver, tests []pipelineTest) {
	for _, test := range tests {
		require.NoErrorWithinT(t, 10*time.Second, func() error {
			return backoff.Retry(func() error {
				return require.ElementsEqualOrErr(test.expectedStates, sDriver.States[test.pipeline])
			}, backoff.NewTestingBackOff())
		})
		require.Equal(t, 1, len(iDriver.Rcs))
		rc, err := iDriver.ReadReplicationController(context.Background(), sDriver.CurrentPipelineInfo(test.pipeline))
		require.NoError(t, err)
		require.True(t, server.RcIsFresh(sDriver.CurrentPipelineInfo(test.pipeline), &rc.Items[0]))
	}
}

func TestBasic(t *testing.T) {
	stateDriver, infraDriver, _ := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	stateDriver.UpsertPipeline(&pps.PipelineInfo{
		Pipeline: client.NewPipeline(pipeline),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	})
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 1, infraDriver.Calls[pipeline][server.MockInfraOp_CREATE])
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
	stateDriver.UpsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 1, infraDriver.Calls[pipeline][server.MockInfraOp_CREATE])
	// clear state and trigger the delete event
	stateDriver.Reset()
	defer stateDriver.CancelWatch()
	stateDriver.PushWatchEvent(pi, watch.EventDelete)
	require.NoErrorWithinT(t, 5*time.Second, func() error {
		return backoff.Retry(func() error {
			if infraDriver.Calls[pipeline][server.MockInfraOp_DELETE] == 1 {
				return nil
			}
			return errors.New("change hasn't reflected")
		}, backoff.NewTestingBackOff())
	})
}

func TestDeleteRC(t *testing.T) {
	stateDriver, infraDriver, _ := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	pi := &pps.PipelineInfo{
		Pipeline: client.NewPipeline(pipeline),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	}
	stateDriver.UpsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 1, infraDriver.Calls[pipeline][server.MockInfraOp_CREATE])
	// remove RCs, and nudge with an event
	infraDriver.ResetRCs()
	stateDriver.PushWatchEvent(pi, watch.EventPut)
	// verify restart side effects were requested
	require.NoErrorWithinT(t, 5*time.Second, func() error {
		return backoff.Retry(func() error {
			if infraDriver.Calls[pipeline][server.MockInfraOp_CREATE] == 2 {
				return nil
			}
			return errors.New("change hasn't reflected")
		}, backoff.NewTestingBackOff())
	})

}

func TestAutoscalingBasic(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	done := mockJobRunning(mockPachd, 1, 1)
	defer close(done)
	stateDriver.UpsertPipeline(&pps.PipelineInfo{
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
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.Calls[pipeline][server.MockInfraOp_CREATE])
}

func TestAutoscalingManyCommits(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	done := mockJobRunning(mockPachd, 1, 100)
	defer close(done)
	inspectCount := 0
	mockPachd.PFS.InspectCommit.Use(func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
		inspectCount++
		return &pfs.CommitInfo{}, nil
	})
	stateDriver.UpsertPipeline(&pps.PipelineInfo{
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
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.Calls[pipeline][server.MockInfraOp_CREATE])
	// 2 * number of commits, to capture the user + meta repos
	require.Equal(t, 200, inspectCount)
}

func TestAutoscalingManyTasks(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	done := mockJobRunning(mockPachd, 100, 1)
	defer close(done)
	mockPachd.PFS.InspectCommit.Use(func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
		// wait for pipeline replica scale to update before closing the commit
		// the first two elements should always be [0, 1]
		require.NoError(t, backoff.Retry(func() error {
			if len(infraDriver.ScaleHistory[pipeline]) > 2 {
				return nil
			}
			return errors.New("waiting for scaleHistory to update")
		}, backoff.NewTestingBackOff()))
		return &pfs.CommitInfo{}, nil
	})
	pi := &pps.PipelineInfo{
		Pipeline: client.NewPipeline(pipeline),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details: &pps.PipelineInfo_Details{
			Autoscaling: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 300,
			},
		},
		Version: 1,
	}
	stateDriver.UpsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.Calls[pipeline][server.MockInfraOp_CREATE])
	require.ElementsEqual(t, []int32{0, 1, 100, 0}, infraDriver.ScaleHistory[pi.Pipeline.Name])
}

func TestAutoscalingNoCommits(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	done := mockJobRunning(mockPachd, 0, 0)
	defer close(done)
	stateDriver.UpsertPipeline(&pps.PipelineInfo{
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
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.Calls[pipeline][server.MockInfraOp_CREATE])
}

func TestPause(t *testing.T) {
	stateDriver, infraDriver, _ := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	pi := &pps.PipelineInfo{
		Pipeline: client.NewPipeline(pipeline),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	}
	spec := stateDriver.UpsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	// pause pipeline
	stateDriver.SpecCommits[spec.ID].Stopped = true
	stateDriver.PushWatchEvent(pi, watch.EventPut)
	waitForPipelineState(t, stateDriver, pi.Pipeline.Name, pps.PipelineState_PIPELINE_PAUSED)
	// unpause pipeline
	stateDriver.SpecCommits[spec.ID].Stopped = false
	stateDriver.PushWatchEvent(pi, watch.EventPut)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_PAUSED,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
}

func TestPauseAutoscaling(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	done := mockJobRunning(mockPachd, 1, 1)
	defer close(done)
	pi := &pps.PipelineInfo{
		Pipeline: client.NewPipeline(pipeline),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details: &pps.PipelineInfo_Details{
			Autoscaling: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
		},
		Version: 1,
	}
	spec := stateDriver.UpsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	testDone := make(chan struct{})
	defer close(testDone)
	// simulate no more commits
	mockPachd.PFS.SubscribeCommit.Use(func(req *pfs.SubscribeCommitRequest, server pfs.API_SubscribeCommitServer) error {
		// keep the subscribe open for the duration of the test
		<-testDone
		return nil
	})
	// pause pipeline
	stateDriver.SpecCommits[spec.ID].Stopped = true
	stateDriver.PushWatchEvent(pi, watch.EventPut)
	waitForPipelineState(t, stateDriver, pi.Pipeline.Name, pps.PipelineState_PIPELINE_PAUSED)
	// unpause pipeline
	stateDriver.SpecCommits[spec.ID].Stopped = false
	stateDriver.PushWatchEvent(pi, watch.EventPut)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_PAUSED,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
}

func TestStaleRestart(t *testing.T) {
	stateDriver, infraDriver, _ := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	pi := &pps.PipelineInfo{
		Pipeline: client.NewPipeline(pipeline),
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	}
	stateDriver.UpsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 1, infraDriver.Calls[pipeline][server.MockInfraOp_CREATE])
	pi.Version = 2
	pi.State = pps.PipelineState_PIPELINE_RUNNING
	stateDriver.UpsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_RESTARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 2, infraDriver.Calls[pipeline][server.MockInfraOp_CREATE])

}
