package discovery

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockClient(t *testing.T) {
	runTest(t, NewMockClient())
}

func TestEtcdClient(t *testing.T) {
	client, err := getEtcdClient()
	require.NoError(t, err)
	runTest(t, client)
}

func runTest(t *testing.T, client Client) {
	require.NoError(t, client.Close())
}

func getEtcdClient() (Client, error) {
	etcdAddress, err := getEtcdAddress()
	if err != nil {
		return nil, err
	}
	return NewEtcdClient(etcdAddress), nil
}

func getEtcdAddress() (string, error) {
	etcdAddr := os.Getenv("ETCD_PORT_2379_TCP_ADDR")
	if etcdAddr == "" {
		return "", errors.New("ETCD_PORT_2379_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:2379", etcdAddr), nil
}
