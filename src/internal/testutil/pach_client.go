package testutil

import (
	"os"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
)

const DefaultTransformImage = "pachyderm/testuser:local"

var (
	pachClient *client.APIClient
	pachErr    error
	clientOnce sync.Once
)

// GetPachClient gets a pachyderm client for use in tests. It works for tests
// running both inside and outside the cluster. Note that multiple calls to
// GetPachClient will return the same instance.
func GetPachClient(t testing.TB) *client.APIClient {
	clientOnce.Do(func() {
		if _, ok := os.LookupEnv("PACHD_PORT_1650_TCP_ADDR"); ok {
			pachClient, pachErr = client.NewInCluster()
		} else {
			pachClient, pachErr = client.NewForTest()
		}
	})
	if pachErr != nil {
		t.Fatalf("error getting Pachyderm client: %s", pachErr.Error())
	}
	return pachClient
}

// NewPachClient gets a pachyderm client for use in tests. It works for tests
// running both inside and outside the cluster. Unlike GetPachClient,
// multiple calls will return new pach client instances
func NewPachClient(t testing.TB) *client.APIClient {
	var c *client.APIClient
	if _, ok := os.LookupEnv("PACHD_PORT_1650_TCP_ADDR"); ok {
		t.Log("creating pach client using cluster config")
		c, pachErr = client.NewInCluster()
	} else {
		t.Log("creating pach client using NewForTest")
		c, pachErr = client.NewForTest()
	}
	if pachErr != nil {
		t.Fatalf("error getting Pachyderm client: %s", pachErr.Error())
	}
	return c
}
