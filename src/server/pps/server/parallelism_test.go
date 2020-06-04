package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func wrap(t testing.TB, ps *pps.ParallelismSpec) *pps.PipelineInfo {
	return &pps.PipelineInfo{
		Pipeline: &pps.Pipeline{
			Name: t.Name() + "-pipeline",
		},
		ParallelismSpec: ps,
	}
}

func TestGetExpectedNumWorkers(t *testing.T) {
	kubeClient := tu.GetKubeClient(t)

	// An empty parallelism spec should default to 1 worker
	workers, err := getExpectedNumWorkers(kubeClient, wrap(t,
		&pps.ParallelismSpec{}))
	require.NoError(t, err)
	require.Equal(t, 1, workers)

	// A constant should literally be returned
	workers, err = getExpectedNumWorkers(kubeClient, wrap(t,
		&pps.ParallelismSpec{
			Constant: 1,
		}))
	require.NoError(t, err)
	require.Equal(t, 1, workers)
	workers, err = getExpectedNumWorkers(kubeClient, wrap(t,
		&pps.ParallelismSpec{
			Constant: 3,
		}))
	require.NoError(t, err)
	require.Equal(t, 3, workers)

	// Constant and Coefficient cannot both be non-zero
	_, err = getExpectedNumWorkers(kubeClient, wrap(t,
		&pps.ParallelismSpec{
			Constant:    3,
			Coefficient: 0.5,
		}))
	require.YesError(t, err)

	// No parallelism spec should default to 1 worker
	workers, err = getExpectedNumWorkers(kubeClient, wrap(t, nil))
	require.NoError(t, err)
	require.Equal(t, 1, workers)

	nodes, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	require.NoError(t, err)
	numNodes := len(nodes.Items)

	// Coefficient == 1
	parellelism, err := getExpectedNumWorkers(kubeClient, wrap(t,
		&pps.ParallelismSpec{
			Coefficient: 1,
		}))
	require.NoError(t, err)
	require.Equal(t, numNodes, parellelism)

	// Coefficient > 1
	parellelism, err = getExpectedNumWorkers(kubeClient, wrap(t,
		&pps.ParallelismSpec{
			Coefficient: 2,
		}))
	require.NoError(t, err)
	require.Equal(t, 2*numNodes, parellelism)

	// Make sure we start at least one worker
	parellelism, err = getExpectedNumWorkers(kubeClient, wrap(t,
		&pps.ParallelismSpec{
			Coefficient: 0.01,
		}))
	require.NoError(t, err)
	require.Equal(t, 1, parellelism)
}
