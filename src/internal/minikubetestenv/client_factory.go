//go:build k8s

package minikubetestenv

import (
	"context"
	"crypto/x509"
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"

	"golang.org/x/sync/semaphore"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	namespacePrefix = "test-cluster-"
)

var (
	clusterFactory *ClusterFactory = &ClusterFactory{
		managedClusters:   map[string]*managedCluster{},
		availableClusters: map[string]struct{}{},
		sem:               *semaphore.NewWeighted(int64(*poolSize)),
	}
	setup               sync.Once
	poolSize            *int  = flag.Int("clusters.pool", 1, "maximum size of managed pachyderm clusters")
	useLeftoverClusters *bool = flag.Bool("clusters.reuse", false, "reuse leftover pachyderm clusters if available")
	cleanupDataAfter    *bool = flag.Bool("clusters.data.cleanup", false, "cleanup the data following each test")
	forceLocal          *bool = flag.Bool("clusters.local", false, "use whatever is in your pachyderm context as the target")
)

type acquireSettings struct {
	SkipLoki         bool
	TLS              bool
	EnterpriseMember bool
	CertPool         *x509.CertPool
	ValueOverrides   map[string]string
}

type Option func(*acquireSettings)

var SkipLokiOption Option = func(as *acquireSettings) {
	as.SkipLoki = true
}

var WithTLS Option = func(as *acquireSettings) {
	as.TLS = true
}

func WithCertPool(pool *x509.CertPool) Option {
	return func(as *acquireSettings) {
		as.CertPool = pool
	}
}

func WithValueOverrides(v map[string]string) Option {
	return func(as *acquireSettings) {
		as.ValueOverrides = v
	}
}

var EnterpriseMemberOption Option = func(as *acquireSettings) {
	as.EnterpriseMember = true
}

type managedCluster struct {
	client   *client.APIClient
	settings *acquireSettings
}

type ClusterFactory struct {
	// ever growing registry of managed clusters. Removing registries would break the current PortOffset logic
	managedClusters   map[string]*managedCluster
	availableClusters map[string]struct{}
	mu                sync.Mutex         // guards modifications to the ClusterFactory maps
	sem               semaphore.Weighted // enforces max concurrency
}

func (cf *ClusterFactory) assignClient(assigned string, mc *managedCluster) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.managedClusters[assigned] = mc
}

func clusterIdx(t testing.TB, name string) int {
	s := strings.Split(name, "-")
	r, err := strconv.Atoi(s[len(s)-1])
	require.NoError(t, err)
	return r
}

func deployOpts(clusterIdx int, as *acquireSettings) *DeployOpts {
	return &DeployOpts{
		PortOffset:         uint16(clusterIdx * 10),
		UseLeftoverCluster: *useLeftoverClusters,
		DisableLoki:        as.SkipLoki,
		TLS:                as.TLS,
		CertPool:           as.CertPool,
		ValueOverrides:     as.ValueOverrides,
	}
}

func deleteAll(t testing.TB, c *client.APIClient) {
	if c == nil {
		return
	}
	tok := c.AuthToken()
	c.SetAuthToken(testutil.RootToken)
	require.NoError(t, c.DeleteAll())
	c.SetAuthToken(tok)
}

func (cf *ClusterFactory) acquireFreeCluster() (string, *managedCluster) {
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

func (cf *ClusterFactory) assignCluster() (string, int) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	idx := len(cf.managedClusters) + 1
	v := fmt.Sprintf("%s%v", namespacePrefix, idx)
	cf.managedClusters[v] = nil // hold my place in line
	return v, idx
}

func (cf *ClusterFactory) acquireNewCluster(t testing.TB, as *acquireSettings) (string, *managedCluster) {
	assigned, clusterIdx := cf.assignCluster()
	kube := testutil.GetKubeClient(t)
	PutNamespace(t, assigned)
	c := InstallRelease(t,
		context.Background(),
		assigned,
		kube,
		deployOpts(clusterIdx, as),
	)
	mc := &managedCluster{
		client:   c,
		settings: as,
	}
	cf.assignClient(assigned, mc)
	return assigned, mc
}

// ClaimCluster returns an unused kubernetes namespace name that can be deployed. It is only responsible for
// assigning clusters to test clients. Unlike AcquireCluster, ClaimCluster doesn't deploy the cluster.
func ClaimCluster(t testing.TB) (string, uint16) {
	// setup.Do(func() {
	// 	clusterFactory = &ClusterFactory{
	// 		managedClusters:   map[string]*managedCluster{},
	// 		availableClusters: map[string]struct{}{},
	// 		sem:               *semaphore.NewWeighted(int64(*poolSize)),
	// 	}
	// })
	require.NoError(t, clusterFactory.sem.Acquire(context.Background(), 1)) // DNJ TODO should semchanges be in helm mutex to prevent starting new cluster before semaphore decrements?
	assigned, clusterIdx := clusterFactory.assignCluster()
	t.Cleanup(func() {
		clusterFactory.mu.Lock()
		clusterFactory.availableClusters[assigned] = struct{}{}
		clusterFactory.mu.Unlock()
		clusterFactory.sem.Release(1)
	})
	portOffset := uint16(clusterIdx * 10)
	return assigned, portOffset
}

var localLock sync.Mutex

// AcquireCluster returns a pachyderm APIClient from one of a pool of managed pachyderm
// clusters deployed in separate namespace, along with the associated namespace
func AcquireCluster(t testing.TB, opts ...Option) (*client.APIClient, string) {
	t.Helper()
	if *forceLocal {
		c, err := client.NewOnUserMachine(pctx.TODO(), "")
		if err != nil {
			t.Fatalf("create local client: %v", err)
		}
		t.Log("waiting for local cluster lock")
		localLock.Lock()
		t.Log("got local cluster lock")
		t.Cleanup(localLock.Unlock)
		return c, ""
	}
	// setup.Do(func() {
	// 	clusterFactory = &ClusterFactory{
	// 		managedClusters:   map[string]*managedCluster{},
	// 		availableClusters: map[string]struct{}{},
	// 		sem:               *semaphore.NewWeighted(int64(*poolSize)),
	// 	}
	// })

	require.NoError(t, clusterFactory.sem.Acquire(context.Background(), 1))
	var assigned string
	t.Cleanup(func() {
		clusterFactory.mu.Lock()
		if mc := clusterFactory.managedClusters[assigned]; mc != nil {
			collectMinikubeCodeCoverage(t, mc.client, mc.settings.ValueOverrides)
			if *cleanupDataAfter {
				deleteAll(t, mc.client)
			}
		}
		clusterFactory.availableClusters[assigned] = struct{}{}
		clusterFactory.mu.Unlock()
		clusterFactory.sem.Release(1)
	})
	as := &acquireSettings{}
	for _, o := range opts {
		o(as)
	}
	assigned, mc := clusterFactory.acquireFreeCluster()
	if assigned == "" {
		assigned, mc = clusterFactory.acquireNewCluster(t, as)
	}

	// If the cluster settings have changed, upgrade the cluster to make them take effect.
	if !reflect.DeepEqual(mc.settings, as) {
		t.Logf("%v: cluster settings have changed; upgrading cluster", assigned)
		mc.client = UpgradeRelease(t,
			context.Background(),
			assigned,
			testutil.GetKubeClient(t),
			deployOpts(clusterIdx(t, assigned), as),
		)
		mc.settings = as
	}
	deleteAll(t, mc.client)
	return mc.client, assigned
}
