package testutil

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
)

var (
	// pachClient is the "template" client that contains the original GRPC
	// connection and that other clients returned by this library are based on. It
	// has no ctx and therefore no auth credentials or cancellation
	pachClient *client.APIClient

	// pachClientOnce ensures that 'pachClient' is only initialized once
	pachClientOnce sync.Once

	// etcdAddress is a hostport pointing (hopefully) to etcd. This library
	// guesses an etcd hostport from the environment
	etcdAddress string

	// etcdAddressOnce ensures that etcdAddress is only initialized once
	etcdAddressOnce sync.Once
)

// GetPachClient returns a pachd client to tests. It's initialized once and then
// reused for future tests
func GetPachClient(tb testing.TB) *client.APIClient {
	tb.Helper()
	pachClientOnce.Do(func() {
		var err error
		if _, ok := os.LookupEnv("PACHD_PORT_650_TCP_ADDR"); ok {
			pachClient, err = client.NewInCluster()
		} else {
			pachClient, err = client.NewOnUserMachine(false, "user")
		}
		if err != nil {
			panic(fmt.Sprintf("pps failed to initialize pach client: %v", err))
		}
	})
	return pachClient
}

// GetEtcdAddress guesses an etcd hostport based on this process's environment
// variables, caches it, and then returns it in this and subsequent calls
func GetEtcdAddress() string {
	etcdAddressOnce.Do(func() {
		if val, ok := os.LookupEnv("ETCD_PORT_2379_TCP_ADDR"); ok {
			etcdAddress = val + ":2379"
		} else if val, ok := os.LookupEnv("ADDRESS"); ok {
			idx := strings.LastIndex(val, ":")
			if idx < 0 {
				panic(fmt.Sprintf("ADDRESS does not contain \":\": %s", val))
			}
			etcdAddress = val[:idx] + ":32379"
		} else {
			etcdAddress = "127.0.0.1:32379"
		}
	})
	return etcdAddress
}
