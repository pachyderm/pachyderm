package minikubetestenv

import (
	"context"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"golang.org/x/sync/semaphore"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	poolSize        = 5
	namespacePrefix = "test-cluster-"
)

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

func (cf *ClusterFactory) acquireFreeCluster() (string, *client.APIClient) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	if len(cf.availableClusters) > 0 {
		var assigned string
		for ns := range cf.availableClusters {
			assigned = ns
			delete(cf.availableClusters, ns)
			break
		}
		return assigned, cf.managedClusters[assigned]
	}
	return "", nil
}

func (cf *ClusterFactory) acquireNewCluster(t testing.TB) (string, *client.APIClient) {
	assigned := testutil.UniqueString(namespacePrefix)
	kube := testutil.GetKubeClient(t)
	_, err := kube.CoreV1().Namespaces().Create(context.Background(),
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: assigned,
			},
		},
		metav1.CreateOptions{})
	require.NoError(t, err)

	cf.mu.Lock()
	cf.managedClusters[assigned] = nil // hold my place in line
	managedCount := len(cf.managedClusters)
	cf.mu.Unlock()

	c := InstallRelease(t,
		context.Background(),
		assigned,
		kube,
		&DeployOpts{
			PortOffset: uint16(managedCount * 11),
		})
	cf.mu.Lock()
	cf.managedClusters[assigned] = c
	cf.mu.Unlock()
	return assigned, c
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
	var assigned string
	assigned, c := clusterFactory.acquireFreeCluster()
	if assigned == "" {
		assigned, c = clusterFactory.acquireNewCluster(t)
	}
	t.Cleanup(func() {
		clusterFactory.sem.Release(1)
		clusterFactory.mu.Lock()
		clusterFactory.availableClusters[assigned] = true
		clusterFactory.mu.Unlock()
	})
	return c
}
