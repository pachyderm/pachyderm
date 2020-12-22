package testutil

import (
	"os"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
)

var (
	pachClient *client.APIClient
	clientOnce sync.Once
)

// GetPachClient gets a pachyderm client for use in tests. It works for tests
// running both inside and outside the cluster. Note that multiple calls to
// GetPachClient will return the same instance.
func GetPachClient(t testing.TB) *client.APIClient {
	clientOnce.Do(func() {
		var err error
		if _, ok := os.LookupEnv("PACHD_PORT_650_TCP_ADDR"); ok {
			pachClient, err = client.NewInCluster()
		} else {
			pachClient, err = client.NewForTest()
		}
		if err != nil {
			t.Fatalf("error getting Pachyderm client: %s", err.Error())
		}
	})
	return pachClient
}
