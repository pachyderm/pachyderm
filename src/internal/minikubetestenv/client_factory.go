package minikubetestenv

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"
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

type acquireSettings struct {
	WaitForLoki      bool
	EnterpriseMember bool
}

type Option func(*acquireSettings)

var WaitForLokiOption = func(ds *acquireSettings) {
	ds.WaitForLoki = true
}

var EnterpriseMemberOption = func(ds *acquireSettings) {
	ds.EnterpriseMember = true
}

var (
	clusterFactory      *ClusterFactory
	setup               sync.Once
	poolSize            *int  = flag.Int("clusters.pool", 6, "maximum size of managed pachyderm clusters")
	useLeftoverClusters *bool = flag.Bool("clusters.reuse", false, "reuse leftover pachyderm clusters if available")
	cleanupDataAfter    *bool = flag.Bool("clusters.data.cleanup", true, "cleanup the data following each test")
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

func clusterIdx(t testing.TB, name string) int {
	s := strings.Split(name, "-")
	r, err := strconv.Atoi(s[len(s)-1])
	require.NoError(t, err)
	return r
}

func deployOpts(clusterIdx int, loki bool) *DeployOpts {
	return &DeployOpts{
		PortOffset:         uint16(clusterIdx * 10),
		UseLeftoverCluster: *useLeftoverClusters,
		Loki:               loki,
	}
}

func deleteAll(t testing.TB, c *client.APIClient) {
	tok := c.AuthToken()
	c.SetAuthToken(testutil.RootToken)
	require.NoError(t, c.DeleteAll())
	c.SetAuthToken(tok)
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

func (cf *ClusterFactory) acquireNewCluster(t testing.TB, as *acquireSettings) (string, *client.APIClient) {
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
		deployOpts(clusterIdx, as.WaitForLoki),
	)
	cf.assignClient(assigned, c)
	return assigned, c
}

// AcquireCluster returns a pachyderm APIClient from one of a pool of managed pachyderm
// clusters deployed in separate namespace, along with the associated namespace
func AcquireCluster(t testing.TB, opts ...Option) (*client.APIClient, string) {
	setup.Do(func() {
		clusterFactory = &ClusterFactory{
			managedClusters:   map[string]*client.APIClient{},
			availableClusters: map[string]struct{}{},
			sem:               *semaphore.NewWeighted(int64(*poolSize)),
		}
	})

	require.NoError(t, clusterFactory.sem.Acquire(context.Background(), 1))
	var assigned string
	t.Cleanup(func() {
		clusterFactory.mu.Lock()
		if *cleanupDataAfter {
			deleteAll(t, clusterFactory.managedClusters[assigned])
		}
		clusterFactory.availableClusters[assigned] = struct{}{}
		clusterFactory.mu.Unlock()
		clusterFactory.sem.Release(1)
	})
	as := &acquireSettings{}
	for _, o := range opts {
		o(as)
	}
	assigned, c := clusterFactory.acquireFreeCluster()
	if assigned == "" {
		assigned, c = clusterFactory.acquireNewCluster(t, as)
	}
	// in the case loki is requested, upgrade the cluster to include it
	if as.WaitForLoki {
		c = UpgradeRelease(t,
			context.Background(),
			assigned,
			testutil.GetKubeClient(t),
			deployOpts(clusterIdx(t, assigned), as.WaitForLoki),
		)
	}
	deleteAll(t, c)
	return c, assigned
}
