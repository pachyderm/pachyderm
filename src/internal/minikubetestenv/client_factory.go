//go:build k8s

package minikubetestenv

import (
	"context"
	"crypto/x509"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/semaphore"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	namespacePrefix = "test-cluster-"
	leasePrefix     = "minikubetestenv-"
)

var (
	clusterFactory *ClusterFactory = &ClusterFactory{
		managedClusters:   map[string]*managedCluster{},
		availableClusters: map[string]struct{}{},
		sem:               *semaphore.NewWeighted(int64(*poolSize)),
	}
	setup               sync.Once
	poolSize            *int  = flag.Int("clusters.pool", 15, "maximum size of managed pachyderm clusters")
	useLeftoverClusters *bool = flag.Bool("clusters.reuse", true, "reuse leftover pachyderm clusters if available")
	cleanupDataAfter    *bool = flag.Bool("clusters.data.cleanup", false, "cleanup the data following each test")
	forceLocal          *bool = flag.Bool("clusters.local", false, "use whatever is in your pachyderm context as the target")
)

type acquireSettings struct {
	SkipLoki         bool
	TLS              bool
	EnterpriseMember bool
	CertPool         *x509.CertPool
	ValueOverrides   map[string]string
	UseNewCluster    bool
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
var UseNewClusterOption Option = func(as *acquireSettings) {
	as.UseNewCluster = true
}

type managedCluster struct {
	client   *client.APIClient
	settings *acquireSettings
}

type ClusterFactory struct {
	// ever growing registry of managed clusters. Removing registries would break the current PortOffset logic
	managedClusters   map[string]*managedCluster
	availableClusters map[string]struct{} // DNJ TOD - remove probably
	mu                sync.Mutex          // guards modifications to the ClusterFactory maps
	sem               semaphore.Weighted  // enforces max concurrency
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
		PortOffset:         uint16(clusterIdx * 100),
		UseLeftoverCluster: *useLeftoverClusters && !as.UseNewCluster,
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
	// DNJ TODO - lock namespace
	return "", nil
}

func (cf *ClusterFactory) assignCluster(t testing.TB) (string, int) {
	kube := testutil.GetKubeClient(t)
	var idx int
	var ns string
	timeout := time.Second * 300
	startAssign := time.Now()
	for idx = 1; idx <= *poolSize+1; idx++ { //DNJ TODO -cleanup
		if idx == *poolSize+1 {
			if time.Since(startAssign) > timeout {
				require.True(t, false, "could not assign a cluster within timeout: %s", timeout.String())
			}
			// DNJ TODO block until we can go, then restart the loop
			time.Sleep(5 * time.Second)
			idx = 0
			continue
		}
		ns = fmt.Sprintf("%s%v", namespacePrefix, idx)
		_, err := kube.CoordinationV1().
			Leases(ns).
			Get(context.Background(), fmt.Sprintf("%s%v", leasePrefix, idx), v1.GetOptions{})
		if k8serrors.IsNotFound(err) { // DNJ TODO this probably needs to PutNamespace here now - is that ok with tests?
			PutNamespace(t, ns)
			err = putLease(t, ns)
			if k8serrors.IsAlreadyExists(err) { // if it already exists, but didn't before we were racing, so don't break to take the next namespaace
				continue
			}
			require.NoError(t, err)
			break
		} else {
			require.NoError(t, err)
		}
	}
	// cf.mu.Lock() // DNJ TODSO -clean
	// defer cf.mu.Unlock()
	// idx := len(cf.managedClusters) + 1
	// v := fmt.Sprintf("%s%v", namespacePrefix, idx)
	cf.managedClusters[ns] = nil // hold my place in line
	return ns, idx
}

func (cf *ClusterFactory) acquireNewCluster(t testing.TB, as *acquireSettings) (string, *managedCluster) {
	assigned, clusterIdx := cf.assignCluster(t)
	kube := testutil.GetKubeClient(t)
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
// assigning clusters to test clients, creating the namespace, and reserving the lease on that namespace.
// Unlike AcquireCluster, ClaimCluster doesn't deploy the cluster.
func ClaimCluster(t testing.TB) (string, uint16) {
	require.NoError(t, clusterFactory.sem.Acquire(context.Background(), 1)) // DNJ TODO should semchanges be in helm mutex to prevent starting new cluster before semaphore decrements?
	assigned, clusterIdx := clusterFactory.assignCluster(t)
	t.Cleanup(func() {
		clusterFactory.mu.Lock()
		clusterFactory.availableClusters[assigned] = struct{}{}
		clusterFactory.mu.Unlock()
		clusterFactory.sem.Release(1)
	})
	portOffset := uint16(clusterIdx * 100)
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
	// assigned, mc := clusterFactory.acquireFreeCluster() // DNJ TODO - anything we can do for perf so we don't always install?
	// if assigned == "" {
	assigned, mc := clusterFactory.acquireNewCluster(t, as)
	// }

	// If the cluster settings have changed, upgrade the cluster to make them take effect.
	// if !reflect.DeepEqual(mc.settings, &acquireSettings{}) { // DNJ TODO - fix - mc is no longer reflective of the current cluster
	// 	t.Logf("%v: cluster settings have changed; upgrading cluster", assigned)
	// 	mc.client = UpgradeRelease(t,
	// 		context.Background(),
	// 		assigned,
	// 		testutil.GetKubeClient(t),
	// 		deployOpts(clusterIdx(t, assigned), as),
	// 	)
	// 	mc.settings = as
	// }
	deleteAll(t, mc.client)
	return mc.client, assigned
}
