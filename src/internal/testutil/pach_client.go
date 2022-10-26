package testutil

import (
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
)

//
const DefaultTransformImage = "ubuntu:20.04"

// NewPachClient gets a pachyderm client for use in tests. It works for tests
// running both inside and outside the cluster. Unlike GetPachClient,
// multiple calls will return new pach client instances
func NewPachClient(t testing.TB) *client.APIClient {
	var c *client.APIClient
	var pachErr error
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
