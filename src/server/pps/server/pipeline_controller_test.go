package server

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/task"
	logrus "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
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
	master.scaleUpInterval = 2 * time.Second
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
	mockPachd.PPS.ListTask.Use(func(req *task.ListTaskRequest, srv pps.API_ListTaskServer) error {
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
		time.Sleep(5 * time.Second)
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
	// clear state and trigger the delete event
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

func TestDeleteRC(t *testing.T) {
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
	// remove RCs, and nudge with an event
	infraDriver.resetRCs()
	stateDriver.pushWatchEvent(pi, watch.EventPut)
	// verify restart side effects were requested
	require.NoErrorWithinT(t, 5*time.Second, func() error {
		return backoff.Retry(func() error {
			if infraDriver.calls[pipeline][mockInfraOp_CREATE] == 2 {
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
				// STANDBY is set consecutively because ppsutil.SetPipelineState
				// will still update the PipelineInfo even if the desired state
				// is already set
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
	require.Equal(t, 1, infraDriver.calls[pipeline][mockInfraOp_CREATE])
}

func TestAutoscalingManyCommits(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	mockAutoscaling(mockPachd, 1, 100)
	inspectCount := 0
	mockPachd.PFS.InspectCommit.Use(func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
		inspectCount++
		return &pfs.CommitInfo{}, nil
	})
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
	// 2 * number of commits, to capture the user + meta repos
	require.Equal(t, 200, inspectCount)
}

func TestAutoscalingManyTasks(t *testing.T) {
	stateDriver, infraDriver, mockPachd := ppsMasterHandles(t)
	pipeline := tu.UniqueString(t.Name())
	mockAutoscaling(mockPachd, 100, 1)
	mockPachd.PFS.InspectCommit.Use(func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
		// add some delay so that Scaling Up will kick in before the commit is otherwise considered done
		time.Sleep(time.Second)
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
	stateDriver.upsertPipeline(pi)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			assertWhen: []pps.PipelineState{
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
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
	require.ElementsEqual(t, []int32{0, 1, 100, 0}, infraDriver.elapsedScales[ppsutil.PipelineRcName(pi.Pipeline.Name, pi.Version)])
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

func TestPause(t *testing.T) {
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
	// pause pipeline
	pi.Stopped = true
	stateDriver.pushWatchEvent(pi, watch.EventPut)
	waitForPipelineStates(t, stateDriver, pi.Pipeline.Name, []pps.PipelineState{pps.PipelineState_PIPELINE_PAUSED})
	// unpause pipeline
	pi.Stopped = false
	stateDriver.pushWatchEvent(pi, watch.EventPut)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			assertWhen: []pps.PipelineState{
				pps.PipelineState_PIPELINE_PAUSED,
				pps.PipelineState_PIPELINE_RUNNING,
			},
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
	mockAutoscaling(mockPachd, 1, 1)
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
	stateDriver.upsertPipeline(pi)
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
	// simulate no more commits
	mockPachd.PFS.SubscribeCommit.Use(func(req *pfs.SubscribeCommitRequest, server pfs.API_SubscribeCommitServer) error {
		// keep the subscribe open
		time.Sleep(3 * time.Second)
		return nil
	})
	// pause pipeline
	pi.Stopped = true
	stateDriver.pushWatchEvent(pi, watch.EventPut)
	waitForPipelineStates(t, stateDriver, pi.Pipeline.Name, []pps.PipelineState{pps.PipelineState_PIPELINE_PAUSED})
	// unpause pipeline
	pi.Stopped = false
	stateDriver.pushWatchEvent(pi, watch.EventPut)
	validate(t, stateDriver, infraDriver, []pipelineTest{
		{
			pipeline: pipeline,
			assertWhen: []pps.PipelineState{
				pps.PipelineState_PIPELINE_PAUSED,
				pps.PipelineState_PIPELINE_STANDBY,
			},
			expectedStates: []pps.PipelineState{
				pps.PipelineState_PIPELINE_STARTING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_RUNNING,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_PAUSED,
				pps.PipelineState_PIPELINE_STANDBY,
				pps.PipelineState_PIPELINE_STANDBY,
			},
		},
	})
}

func TestCrashing(t *testing.T) {

}

func TestRestarts(t *testing.T) {

}

// Tests that monitor goros update
func TestUpdateAutoscalingSpec(t *testing.T) {

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
		actualState, actualSideEffects, err := evaluate(pi, rc)
		require.NoError(t, err)
		require.Equal(t, expectedState, actualState)
		require.Equal(t, len(expectedSideEffects), len(actualSideEffects))
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
