package minikubetestenv

import (
	"context"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"golang.org/x/sync/semaphore"
)

const poolSize = 5

var clusterPool []*client.APIClient

var (
	clientFactory *ClientFactory
	setup         sync.Once
)

type ClientFactory struct {
	sem               *semaphore.Weighted
	managedClusters   map[string]*client.APIClient
	availableClusters map[string]bool
}

// AcquireClient returns a pachyderm APIClient for one of a pool of
// managed pachyderm clusters deployed in namespaces
func AcquireClient(t *testing.T, ctx context.Context) (*client.APIClient, func()) {
	setup.Do(func() {
		clientFactory = &ClientFactory{
			managedClusters:   map[string]*client.APIClient{},
			availableClusters: map[string]bool{},
			sem:               semaphore.NewWeighted(poolSize),
		}
	})
	require.NoError(t, clientFactory.sem.Acquire(ctx, 1))
	var assigned string
	if len(clientFactory.availableClusters) > 0 {
		for ns := range clientFactory.availableClusters {
			assigned = ns
			delete(clientFactory.availableClusters, ns)
			break
		}
	} else if len(clientFactory.managedClusters) < poolSize {
		// deploy
		assigned = testutil.UniqueString("testenv-cluster-")
		clientFactory.managedClusters[assigned] = nil
	} else {
		t.Fatal("Ahhhh")
	}

	return nil, func() {
		clientFactory.sem.Release(1)
		clientFactory.availableClusters[assigned] = true
	}
}
