package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func wrap(t testing.TB, ps *pps.ParallelismSpec) *pps.PipelineInfo {
	return &pps.PipelineInfo{
		Pipeline: &pps.Pipeline{
			// TODO: should this be project-aware?
			Name: t.Name() + "-pipeline",
		},
		Details: &pps.PipelineInfo_Details{
			ParallelismSpec: ps,
		},
	}
}

func TestGetExpectedNumWorkers(t *testing.T) {
	// An empty parallelism spec should default to 1 worker
	workers, err := getExpectedNumWorkers(wrap(t,
		&pps.ParallelismSpec{}))
	require.NoError(t, err)
	require.Equal(t, 1, workers)

	// A constant should literally be returned
	workers, err = getExpectedNumWorkers(wrap(t,
		&pps.ParallelismSpec{
			Constant: 1,
		}))
	require.NoError(t, err)
	require.Equal(t, 1, workers)
	workers, err = getExpectedNumWorkers(wrap(t,
		&pps.ParallelismSpec{
			Constant: 3,
		}))
	require.NoError(t, err)
	require.Equal(t, 3, workers)

	// No parallelism spec should default to 1 worker
	workers, err = getExpectedNumWorkers(wrap(t, nil))
	require.NoError(t, err)
	require.Equal(t, 1, workers)
}
