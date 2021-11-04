package minikubetestenv

import (
	"os"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

var clusterLock sync.Mutex

// NewPachClient gets a pachyderm client for use in tests.
// The client will be connected to a full Pachyderm cluster deployed in minikube
// NewPachClient only allows one test to be run at a time.  After calling it with test t,
// subsequent calls will block until t finishes.
//
// TODO: eventually we can manage multiple deployments across different namespaces.
func NewPachClient(t testing.TB) *client.APIClient {
	clusterLock.Lock()
	t.Cleanup(clusterLock.Unlock)
	var pachClient *client.APIClient
	var err error
	if _, ok := os.LookupEnv("PACHD_PORT_1650_TCP_ADDR"); ok {
		pachClient, err = client.NewInCluster()
	} else {
		pachClient, err = client.NewForTest()
	}
	if err != nil {
		t.Fatalf("error getting Pachyderm client: %s", err.Error())
	}
	require.NoError(t, pachClient.DeleteAll())
	t.Cleanup(func() { require.NoError(t, pachClient.DeleteAll()) })
	return pachClient
}
