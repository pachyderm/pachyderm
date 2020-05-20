package ppsutil

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

func TestGetExpectedNumWorkers(t *testing.T) {
	kubeClient := tu.GetKubeClient(t)
	parallelismSpec := &pps.ParallelismSpec{}

	// An empty parallelism spec should default to 1 worker
	parallelismSpec.Constant = 0
	parallelismSpec.Coefficient = 0
	workers, err := GetExpectedNumWorkers(kubeClient, parallelismSpec)
	require.NoError(t, err)
	require.Equal(t, 1, workers)

	// A constant should literally be returned
	parallelismSpec.Constant = 1
	workers, err = GetExpectedNumWorkers(kubeClient, parallelismSpec)
	require.NoError(t, err)
	require.Equal(t, 1, workers)

	parallelismSpec.Constant = 3
	workers, err = GetExpectedNumWorkers(kubeClient, parallelismSpec)
	require.NoError(t, err)
	require.Equal(t, 3, workers)

	// Constant and Coefficient cannot both be non-zero
	parallelismSpec.Coefficient = 0.5
	_, err = GetExpectedNumWorkers(kubeClient, parallelismSpec)
	require.YesError(t, err)

	// No parallelism spec should default to 1 worker
	parallelismSpec = nil
	workers, err = GetExpectedNumWorkers(kubeClient, parallelismSpec)
	require.NoError(t, err)
	require.Equal(t, 1, workers)

	// TODO: test a non-zero coefficient - requires setting up a number of nodes with the kubeClient
}
