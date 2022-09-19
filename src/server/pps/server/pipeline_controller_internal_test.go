package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	v1 "k8s.io/api/core/v1"
)

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
