package minikubetestenv

import (
	"context"
	"flag"
	"fmt"
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
	namespacePrefix = "test-cluster-"
)

var (
	clusterFactory      *ClusterFactory
	setup               sync.Once
	poolSize            *int  = flag.Int("clusters.pool", 6, "maximum size of managed pachyderm clusters")
	useLeftoverClusters *bool = flag.Bool("clusters.reuse", false, "reuse leftover pachyderm clusters if available")
)

type ClusterFactory struct {
	// ever growing registry of managed clusters. Removing registries would break the current PortOffset logic
	managedClusters   map[string]*client.APIClient
	availableClusters map[string]struct{}
	mu                sync.Mutex         // guards modifications to the ClusterFactory maps
	sem               semaphore.Weighted // enforces max concurrency
}

func (cf *ClusterFactory) assignClient(assigned string, c *client.APIClient) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.managedClusters[assigned] = c
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
	assigned, clusterIdx := func() (string, int) {
		cf.mu.Lock()
		defer cf.mu.Unlock()
		idx := len(cf.managedClusters) + 1
		v := fmt.Sprintf("%s%v", namespacePrefix, idx)
		cf.managedClusters[v] = nil // hold my place in line
		return v, idx
	}()
	kube := testutil.GetKubeClient(t)
	if _, err := kube.CoreV1().Namespaces().Get(context.Background(), assigned, metav1.GetOptions{}); err != nil {
		_, err := kube.CoreV1().Namespaces().Create(context.Background(),
			&v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: assigned,
				},
			},
			metav1.CreateOptions{})
		require.NoError(t, err)
	}
	c := InstallRelease(t,
		context.Background(),
		assigned,
		kube,
		&DeployOpts{
			PortOffset:         uint16(clusterIdx * 10),
			UseLeftoverCluster: *useLeftoverClusters,
		})
	cf.assignClient(assigned, c)
	return assigned, c
}

// AcquireCluster returns a pachyderm APIClient from one of a pool of managed pachyderm
// clusters deployed in separate namespace, along with the associated namespace
func AcquireCluster(t testing.TB) (*client.APIClient, string) {
	setup.Do(func() {
		clusterFactory = &ClusterFactory{
			managedClusters:   map[string]*client.APIClient{},
			availableClusters: map[string]struct{}{},
			sem:               *semaphore.NewWeighted(int64(*poolSize)),
		}
	})

	require.NoError(t, clusterFactory.sem.Acquire(context.Background(), 1))
	var assigned string
	// TODO(acohen4): client.DeleteAll() during cleanup
	t.Cleanup(func() {
		clusterFactory.mu.Lock()
		clusterFactory.availableClusters[assigned] = struct{}{}
		clusterFactory.mu.Unlock()
		clusterFactory.sem.Release(1)
	})
	assigned, c := clusterFactory.acquireFreeCluster()
	if assigned == "" {
		assigned, c = clusterFactory.acquireNewCluster(t)
	}
	return c, assigned
}
