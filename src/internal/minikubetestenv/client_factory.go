//go:build k8s

package minikubetestenv

import (
	"context"
	"crypto/x509"
	"flag"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	namespacePrefix = "test-cluster-"
)

var (
	clusterFactory = &ClusterFactory{
		managedClusters: map[string]*managedCluster{},
	}
	poolSize            = flag.Int("clusters.pool", 9, "maximum size of managed pachyderm clusters")
	useLeftoverClusters = flag.Bool("clusters.reuse", true, "reuse leftover pachyderm clusters if available")
	cleanupDataAfter    = flag.Bool("clusters.data.cleanup", false, "cleanup the data following each test")
	forceLocal          = flag.Bool("clusters.local", false, "use whatever is in your pachyderm context as the target")
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
	managedClusters map[string]*managedCluster
	mu              sync.Mutex // guards modifications to the ClusterFactory maps
}

func (cf *ClusterFactory) assignClient(assigned string, mc *managedCluster) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.managedClusters[assigned] = mc
}

func deployOpts(clusterIdx int, as *acquireSettings) *DeployOpts {
	return &DeployOpts{
		PortOffset:         uint16(clusterIdx * 150),
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

func (cf *ClusterFactory) assignCluster(t testing.TB) (string, int) {
	var idx int
	var ns string
	// if we exhausted allowed namespaces, wait and then try again to see if one is available
	require.NoErrorWithinTRetryConstant(t, time.Second*300, func() error {
		var ok bool
		for idx = 1; idx <= *poolSize; idx++ {
			ns = fmt.Sprintf("%s%v", namespacePrefix, idx)
			if ok = LeaseNamespace(t, ns); ok {
				break
			}
		}
		if !ok {
			return errors.Errorf("No namespaces available to provision, waiting to try again.")
		}
		return nil
	}, time.Second*5, "Could not assign a test namespace within timeout")
	// Take a slot in managedClusters where we will cache the pachclient if available.
	// Useful for getting code coverage from the cluster at the end of the test
	cf.managedClusters[ns] = nil
	return ns, idx
}

func (cf *ClusterFactory) acquireInstalledCluster(t testing.TB, as *acquireSettings) (string, *managedCluster) {
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
// Unlike AcquireCluster, ClaimCluster doesn't do helm install on the cluster.
func ClaimCluster(t testing.TB) (string, uint16) {
	assigned, clusterIdx := clusterFactory.assignCluster(t)
	portOffset := uint16(clusterIdx * 150)
	return assigned, portOffset
}

var localLock sync.Mutex

// AcquireCluster returns a pachyderm APIClient from one of a pool of managed pachyderm
// clusters deployed in separate namespace, along with the associated namespace
func AcquireCluster(t testing.TB, opts ...Option) (*client.APIClient, string) {
	t.Helper()
	ctx := pctx.TestContext(t)
	if *forceLocal {
		c, err := client.NewOnUserMachine(ctx, "")
		if err != nil {
			t.Fatalf("create local client: %v", err)
		}
		t.Log("waiting for local cluster lock")
		localLock.Lock()
		t.Log("got local cluster lock")
		t.Cleanup(localLock.Unlock)
		return c, ""
	}

	as := &acquireSettings{}
	for _, o := range opts {
		o(as)
	}
	assigned, mc := clusterFactory.acquireInstalledCluster(t, as)
	t.Cleanup(func() { // must come after assignment to run cleanmup with code coverage before the lease is removed.
		clusterFactory.mu.Lock()
		if mc := clusterFactory.managedClusters[assigned]; mc != nil {
			collectMinikubeCodeCoverage(t, mc.client, mc.settings.ValueOverrides)
			if *cleanupDataAfter {
				deleteAll(t, mc.client)
			}
		}
		clusterFactory.mu.Unlock()
	})
	deleteAll(t, mc.client)
	return mc.client, assigned
}
