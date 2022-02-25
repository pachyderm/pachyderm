package minikubetestenv

import (
	"context"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"golang.org/x/sync/semaphore"
)

const poolSize = 5

var clusterPool []*client.APIClient

var (
	clusterFactory *ClusterFactory
	setup          sync.Once
)

type ClusterFactory struct {
	managedClusters   map[string]*client.APIClient
	availableClusters map[string]bool
	mu                sync.Mutex         // guards modifications to the ClusterFactory maps
	sem               semaphore.Weighted // enforces max concurrency
}

// AcquireCluster returns a pachyderm APIClient for one of a pool of
// managed pachyderm clusters deployed in namespaces
func AcquireCluster(t testing.TB, ctx context.Context) *client.APIClient {
	setup.Do(func() {
		clusterFactory = &ClusterFactory{
			managedClusters:   map[string]*client.APIClient{},
			availableClusters: map[string]bool{},
			sem:               *semaphore.NewWeighted(poolSize),
		}
	})
	clusterFactory.sem.Acquire(ctx, 1)
	clusterFactory.mu.Lock()
	var assigned string
	if len(clusterFactory.availableClusters) > 0 {
		for ns := range clusterFactory.availableClusters {
			assigned = ns
			delete(clusterFactory.availableClusters, ns)
			break
		}
		return clusterFactory.managedClusters[assigned]
	} else if len(clusterFactory.managedClusters) < poolSize {
		assigned = testutil.UniqueString("testenv-cluster-")
		newClusterClient := InstallRelease(t, context.Background(), assigned, testutil.GetKubeClient(t), &DeployOpts{})
		clusterFactory.managedClusters[assigned] = newClusterClient
	} else {
		clusterFactory.mu.Unlock()
		clusterFactory.sem.Release(1)
		t.Fatal("A test's goroutine shouldn't proceed if the ClusterFactory is running at maximum concurrency.")
	}
	clusterFactory.mu.Unlock()
	t.Cleanup(func() {
		clusterFactory.sem.Release(1)
		clusterFactory.availableClusters[assigned] = true
	})
	return clusterFactory.managedClusters[assigned]
}
