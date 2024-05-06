package server

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	v1 "k8s.io/api/core/v1"
)

type pipelineTest struct {
	key            pipelineKey
	expectedStates []pps.PipelineState
}

func ppsMasterHandles(t *testing.T, taskCount int64) (*mockStateDriver, *mockInfraDriver, *testpachd.MockPachd) {
	ctx := pctx.TestContext(t)
	sDriver := newMockStateDriver()
	iDriver := newMockInfraDriver()
	mockEnv := testpachd.NewMockEnv(ctx, t)
	env := Env{
		BackgroundContext: context.Background(),
		EtcdClient:        mockEnv.EtcdClient,
		GetPachClient: func(ctx context.Context) *client.APIClient {
			return mockEnv.PachClient.WithCtx(ctx)
		},
		Config:      *pachconfig.ConfigFromOptions(),
		TaskService: newMockTaskService(taskCount),
		DB:          nil,
		KubeClient:  nil,
		AuthServer:  nil,
		PFSServer:   nil,
		TxnEnv:      nil,
	}
	ctx, cf := pctx.WithCancel(context.Background())
	t.Cleanup(func() { cf() })
	master := newMaster(ctx, env, env.EtcdPrefix, iDriver, sDriver)
	master.scaleUpInterval = 2 * time.Second
	go master.run()
	return sDriver, iDriver, mockEnv.MockPachd
}

type mockTaskService struct {
	count int64
}

func newMockTaskService(count int64) *mockTaskService {
	return &mockTaskService{
		count: count,
	}
}

func (mts *mockTaskService) NewDoer(_, _ string, _ task.Cache) task.Doer {
	panic("NewDoer not implemented by mockTaskService")
}
func (mts *mockTaskService) NewSource(_ string) task.Source {
	panic("NewSource not implemented by mockTaskService")
}
func (mts *mockTaskService) List(_ context.Context, _, _ string, _ func(string, string, *task.Task, bool) error) error {
	panic("List not implemented by mockTaskService")
}
func (mts *mockTaskService) Count(_ context.Context, _ string) (int64, error) {
	return mts.count, nil
}

func waitForPipelineState(t testing.TB, stateDriver *mockStateDriver, pipeline *pps.Pipeline, state pps.PipelineState) {
	require.NoErrorWithinT(t, 10*time.Second, func() error {
		return backoff.Retry(func() error {
			actualStates := stateDriver.states[toKey(pipeline)]
			for _, s := range actualStates {
				if s == state {
					return nil
				}
			}
			return errors.New("change hasn't reflected")
		}, backoff.NewTestingBackOff())
	})
}

func mockJobRunning(mockPachd *testpachd.MockPachd, commitCount int) (done chan struct{}) {
	mockPachd.PFS.InspectCommit.Use(func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
		return &pfs.CommitInfo{}, nil
	})
	testDone := make(chan struct{})
	mockPachd.PFS.SubscribeCommit.Use(func(req *pfs.SubscribeCommitRequest, server pfs.API_SubscribeCommitServer) error {
		for i := 0; i < commitCount; i++ {
			if err := server.Send(&pfs.CommitInfo{Commit: client.NewCommit(req.Repo.Project.GetName(), req.Repo.Name, "", uuid.NewWithoutDashes())}); err != nil {
				return errors.EnsureStack(err)
			}
		}
		<-testDone
		return nil
	})
	return testDone
}

func validate(t testing.TB, sDriver *mockStateDriver, iDriver *mockInfraDriver, tests []pipelineTest) {
	for _, test := range tests {
		require.NoErrorWithinT(t, 10*time.Second, func() error {
			return backoff.Retry(func() error {
				return require.ElementsEqualOrErr(test.expectedStates, sDriver.states[test.key])
			}, backoff.NewTestingBackOff())
		})
		require.Equal(t, 1, len(iDriver.rcs))
		rc, err := iDriver.ReadReplicationController(context.Background(), sDriver.currentPipelineInfo(test.key))
		require.NoError(t, err)
		require.True(t, rcIsFresh(context.TODO(), sDriver.currentPipelineInfo(test.key), &rc.Items[0]))
	}
}

func TestBasic(t *testing.T) {
	stateDriver, infraDriver, _ := ppsMasterHandles(t, 0)
	pipelineName := tu.UniqueString(t.Name())
	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	stateDriver.upsertPipeline(&pps.PipelineInfo{
		Pipeline: pipeline,
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	})
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			key: toKey(pipeline),
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[toKey(pipeline)][mockInfraOp_CREATE])
}

func TestDeletePipeline(t *testing.T) {
	stateDriver, infraDriver, _ := ppsMasterHandles(t, 0)
	pipelineName := tu.UniqueString(t.Name())
	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	pi := &pps.PipelineInfo{
		Pipeline: pipeline,
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	}
	stateDriver.upsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			key: toKey(pipeline),
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[toKey(pipeline)][mockInfraOp_CREATE])
	// clear state and trigger the delete event
	stateDriver.reset()
	defer stateDriver.cancelWatch()
	stateDriver.pushWatchEvent(pi, watch.EventDelete)
	require.NoErrorWithinT(t, 5*time.Second, func() error {
		return backoff.Retry(func() error {
			if infraDriver.calls[toKey(pipeline)][mockInfraOp_DELETE] == 1 {
				return nil
			}
			return errors.New("change hasn't reflected")
		}, backoff.NewTestingBackOff())
	})
}

func TestDeleteRC(t *testing.T) {
	stateDriver, infraDriver, _ := ppsMasterHandles(t, 0)
	pipelineName := tu.UniqueString(t.Name())
	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	pi := &pps.PipelineInfo{
		Pipeline: pipeline,
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	}
	stateDriver.upsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			key: toKey(pipeline),
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[toKey(pipeline)][mockInfraOp_CREATE])
	// remove RCs, and nudge with an event
	infraDriver.resetRCs()
	stateDriver.pushWatchEvent(pi, watch.EventPut)
	// verify restart side effects were requested
	require.NoErrorWithinT(t, 5*time.Second, func() error {
		return backoff.Retry(func() error {
			if infraDriver.calls[toKey(pipeline)][mockInfraOp_CREATE] == 2 {
				return nil
			}
			return errors.New("change hasn't reflected")
		}, backoff.NewTestingBackOff())
	})

}

func TestAutoscalingBasic(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t, 1)
	pipelineName := tu.UniqueString(t.Name())
	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	done := mockJobRunning(mockPachd, 1)
	defer close(done)
	stateDriver.upsertPipeline(&pps.PipelineInfo{
		Pipeline: pipeline,
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
			key: toKey(pipeline),
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[toKey(pipeline)][mockInfraOp_CREATE])
}

func TestAutoscalingManyCommits(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t, 1)
	pipelineName := tu.UniqueString(t.Name())
	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	done := mockJobRunning(mockPachd, 100)
	defer close(done)
	inspectCount := 0
	mockPachd.PFS.InspectCommit.Use(func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
		inspectCount++
		return &pfs.CommitInfo{}, nil
	})
	stateDriver.upsertPipeline(&pps.PipelineInfo{
		Pipeline: pipeline,
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
			key: toKey(pipeline),
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[toKey(pipeline)][mockInfraOp_CREATE])
	// 2 * number of commits, to capture the user + meta repos
	require.Equal(t, 200, inspectCount)
}

func TestAutoscalingManyTasks(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t, 100)
	pipelineName := tu.UniqueString(t.Name())
	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	done := mockJobRunning(mockPachd, 1)
	defer close(done)
	mockPachd.PFS.InspectCommit.Use(func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
		// wait for pipeline replica scale to update before closing the commit
		// the first two elements should always be [0, 1]
		require.NoError(t, backoff.Retry(func() error {
			if len(infraDriver.scaleHistory[toKey(pipeline)]) > 2 {
				return nil
			}
			return errors.New("waiting for scaleHistory to update")
		}, backoff.NewTestingBackOff()))
		return &pfs.CommitInfo{}, nil
	})
	pi := &pps.PipelineInfo{
		Pipeline: pipeline,
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details: &pps.PipelineInfo_Details{
			Autoscaling: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 300,
			},
		},
		Version: 1,
	}
	stateDriver.upsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			key: toKey(pipeline),
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[toKey(pipeline)][mockInfraOp_CREATE])
	require.ElementsEqual(t, []int32{0, 1, 100, 0}, infraDriver.scaleHistory[toKey(pi.Pipeline)])
}

func TestAutoscalingNoCommits(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t, 0)
	pipelineName := tu.UniqueString(t.Name())
	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	done := mockJobRunning(mockPachd, 0)
	defer close(done)
	stateDriver.upsertPipeline(&pps.PipelineInfo{
		Pipeline: pipeline,
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
			key: toKey(pipeline),
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[toKey(pipeline)][mockInfraOp_CREATE])
}

func TestPause(t *testing.T) {
	stateDriver, infraDriver, _ := ppsMasterHandles(t, 0)
	pipelineName := tu.UniqueString(t.Name())
	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	pi := &pps.PipelineInfo{
		Pipeline: pipeline,
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	}
	spec := stateDriver.upsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			key: toKey(pipeline),
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	// pause pipeline
	stateDriver.specCommits[spec.Id].Stopped = true
	stateDriver.pushWatchEvent(pi, watch.EventPut)
	waitForPipelineState(t, stateDriver, pi.Pipeline, pps.PipelineState_PIPELINE_PAUSED)
	// unpause pipeline
	stateDriver.specCommits[spec.Id].Stopped = false
	stateDriver.pushWatchEvent(pi, watch.EventPut)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			key: toKey(pipeline),
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
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t, 1)
	pipelineName := tu.UniqueString(t.Name())
	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	done := mockJobRunning(mockPachd, 1)
	defer close(done)
	pi := &pps.PipelineInfo{
		Pipeline: pipeline,
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details: &pps.PipelineInfo_Details{
			Autoscaling: true,
			ParallelismSpec: &pps.ParallelismSpec{
				Constant: 1,
			},
		},
		Version: 1,
	}
	spec := stateDriver.upsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			key: toKey(pipeline),
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
	stateDriver.specCommits[spec.Id].Stopped = true
	stateDriver.pushWatchEvent(pi, watch.EventPut)
	waitForPipelineState(t, stateDriver, pi.Pipeline, pps.PipelineState_PIPELINE_PAUSED)
	// unpause pipeline
	stateDriver.specCommits[spec.Id].Stopped = false
	stateDriver.pushWatchEvent(pi, watch.EventPut)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			key: toKey(pipeline),
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
	stateDriver, infraDriver, _ := ppsMasterHandles(t, 0)
	pipelineName := tu.UniqueString(t.Name())
	pipeline := client.NewPipeline(pfs.DefaultProjectName, pipelineName)
	pi := &pps.PipelineInfo{
		Pipeline: pipeline,
		State:    pps.PipelineState_PIPELINE_STARTING,
		Details:  &pps.PipelineInfo_Details{},
		Version:  1,
	}
	stateDriver.upsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			key: toKey(pipeline),
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[toKey(pipeline)][mockInfraOp_CREATE])
	pi.Version = 2
	pi.State = pps.PipelineState_PIPELINE_RUNNING
	stateDriver.upsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			key: toKey(pipeline),
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_RESTARTING,
				pps.PipelineState_PIPELINE_RUNNING,
			},
		},
	})
	require.Equal(t, 2, infraDriver.calls[toKey(pipeline)][mockInfraOp_CREATE])

}

type evalTest struct {
	state       pps.PipelineState
	sideEffects []sideEffect
}

// explicitly tests the logic of pipeline_controller.go#evaluate()
func TestEvaluate(t *testing.T) {
	// conceptual params: State (8 opts), Autoscaling (2 opts), Stopped (2 opts), rc=?nil (2 opts)
	pi := &pps.PipelineInfo{
		Details: &pps.PipelineInfo_Details{
			Autoscaling: false,
		},
	}
	rc := &v1.ReplicationController{}
	test := func(startState pps.PipelineState, expectedState pps.PipelineState, expectedSideEffects []sideEffect) {
		pi.State = startState
		actualState, actualSideEffects, _, err := evaluate(pi, rc)
		require.NoError(t, err)
		require.Equal(t, expectedState, actualState, "for start state %v, expected state was %v but got %v", startState, expectedState, actualState)
		require.Equal(t, len(expectedSideEffects), len(actualSideEffects), "for start state %v, expected side effects are %v but got %v", startState, expectedSideEffects, actualSideEffects)
		for i := 0; i < len(expectedSideEffects); i++ {
			require.True(t, expectedSideEffects[i].equals(actualSideEffects[i]))
		}
		require.Equal(t, expectedState, actualState)
	}
	stackMaps := func(maps ...map[pps.PipelineState]evalTest) map[pps.PipelineState]evalTest {
		newMap := make(map[pps.PipelineState]evalTest)
		for _, m := range maps {
			for k, v := range m {
				newMap[k] = v
			}
		}
		return newMap
	}
	// Autoscaling = False, Stopped = False, With RC != nil
	tests := map[pps.PipelineState]evalTest{
		pps.PipelineState_PIPELINE_STARTING: {
			state: pps.PipelineState_PIPELINE_RUNNING,
			sideEffects: []sideEffect{
				CrashingMonitorSideEffect(sideEffectToggle_DOWN),
			},
		},
		pps.PipelineState_PIPELINE_RUNNING: {
			state: pps.PipelineState_PIPELINE_RUNNING,
			sideEffects: []sideEffect{
				CrashingMonitorSideEffect(sideEffectToggle_DOWN),
				PipelineMonitorSideEffect(sideEffectToggle_UP),
				ScaleWorkersSideEffect(sideEffectToggle_UP),
			},
		},
		pps.PipelineState_PIPELINE_RESTARTING: {
			state: pps.PipelineState_PIPELINE_RUNNING,
			sideEffects: []sideEffect{
				CrashingMonitorSideEffect(sideEffectToggle_DOWN),
				PipelineMonitorSideEffect(sideEffectToggle_DOWN),
			},
		},
		pps.PipelineState_PIPELINE_FAILURE: {
			state: pps.PipelineState_PIPELINE_FAILURE,
			sideEffects: []sideEffect{
				FinishCommitsSideEffect(),
				ResourcesSideEffect(sideEffectToggle_DOWN),
			},
		},
		pps.PipelineState_PIPELINE_PAUSED: {
			state:       pps.PipelineState_PIPELINE_RUNNING,
			sideEffects: []sideEffect{},
		},
		pps.PipelineState_PIPELINE_STANDBY: {
			state: pps.PipelineState_PIPELINE_STANDBY,
			sideEffects: []sideEffect{
				CrashingMonitorSideEffect(sideEffectToggle_DOWN),
				PipelineMonitorSideEffect(sideEffectToggle_UP),
				ScaleWorkersSideEffect(sideEffectToggle_DOWN),
			},
		},
		pps.PipelineState_PIPELINE_CRASHING: {
			state: pps.PipelineState_PIPELINE_CRASHING,
			sideEffects: []sideEffect{
				CrashingMonitorSideEffect(sideEffectToggle_UP),
				PipelineMonitorSideEffect(sideEffectToggle_UP),
				ScaleWorkersSideEffect(sideEffectToggle_UP),
			},
		},
	}
	for initState, expectedResult := range tests {
		test(initState, expectedResult.state, expectedResult.sideEffects)
	}

	// Autoscaling == true, Stopped == false, RC != nil
	pi.Details.Autoscaling = true
	tests[pps.PipelineState_PIPELINE_STARTING] = evalTest{
		state: pps.PipelineState_PIPELINE_STANDBY,
		sideEffects: []sideEffect{
			CrashingMonitorSideEffect(sideEffectToggle_DOWN),
		},
	}
	tests[pps.PipelineState_PIPELINE_PAUSED] = evalTest{
		state:       pps.PipelineState_PIPELINE_STANDBY,
		sideEffects: []sideEffect{},
	}
	for initState, expectedResult := range tests {
		test(initState, expectedResult.state, expectedResult.sideEffects)
	}

	// Autoscaling == true, Stopped == true, RC != nil
	pi.Stopped = true
	testsStop := stackMaps(tests, tests, map[pps.PipelineState]evalTest{
		pps.PipelineState_PIPELINE_STARTING: {
			state:       pps.PipelineState_PIPELINE_PAUSED,
			sideEffects: []sideEffect{},
		},
		pps.PipelineState_PIPELINE_RESTARTING: {
			state:       pps.PipelineState_PIPELINE_PAUSED,
			sideEffects: []sideEffect{},
		},
		pps.PipelineState_PIPELINE_PAUSED: {
			state: pps.PipelineState_PIPELINE_PAUSED,
			sideEffects: []sideEffect{
				// TODO: Could we just do this when we pause?
				PipelineMonitorSideEffect(sideEffectToggle_DOWN),
				CrashingMonitorSideEffect(sideEffectToggle_DOWN),
				ScaleWorkersSideEffect(sideEffectToggle_DOWN),
			},
		},
		pps.PipelineState_PIPELINE_RUNNING: {
			state:       pps.PipelineState_PIPELINE_PAUSED,
			sideEffects: []sideEffect{},
		},
		pps.PipelineState_PIPELINE_STANDBY: {
			state:       pps.PipelineState_PIPELINE_PAUSED,
			sideEffects: []sideEffect{},
		},
		pps.PipelineState_PIPELINE_CRASHING: {
			state:       pps.PipelineState_PIPELINE_PAUSED,
			sideEffects: []sideEffect{},
		},
	})

	for initState, expectedResult := range testsStop {
		test(initState, expectedResult.state, expectedResult.sideEffects)
	}

	// Stopped == false, Autoscaling == true, RC == nil
	rc = nil
	pi.Stopped = false
	testsNoRC := stackMaps(tests, map[pps.PipelineState]evalTest{
		pps.PipelineState_PIPELINE_STARTING: {
			state: pps.PipelineState_PIPELINE_STANDBY,
			sideEffects: []sideEffect{
				ResourcesSideEffect(sideEffectToggle_UP),
				CrashingMonitorSideEffect(sideEffectToggle_DOWN),
			},
		},
		pps.PipelineState_PIPELINE_RESTARTING: {
			state: pps.PipelineState_PIPELINE_RUNNING,
			sideEffects: []sideEffect{
				ResourcesSideEffect(sideEffectToggle_UP),
				CrashingMonitorSideEffect(sideEffectToggle_DOWN),
				PipelineMonitorSideEffect(sideEffectToggle_DOWN),
			},
		},
		pps.PipelineState_PIPELINE_PAUSED: {
			state:       pps.PipelineState_PIPELINE_RESTARTING,
			sideEffects: []sideEffect{RestartSideEffect()},
		},
		pps.PipelineState_PIPELINE_RUNNING: {
			state:       pps.PipelineState_PIPELINE_RESTARTING,
			sideEffects: []sideEffect{RestartSideEffect()},
		},
		pps.PipelineState_PIPELINE_STANDBY: {
			state:       pps.PipelineState_PIPELINE_RESTARTING,
			sideEffects: []sideEffect{RestartSideEffect()},
		},
		pps.PipelineState_PIPELINE_CRASHING: {
			state:       pps.PipelineState_PIPELINE_RESTARTING,
			sideEffects: []sideEffect{RestartSideEffect()},
		},
	})
	for initState, expectedResult := range testsNoRC {
		test(initState, expectedResult.state, expectedResult.sideEffects)
	}
}
