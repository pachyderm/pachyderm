package discovery

import (
	"fmt"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestEtcdClient(t *testing.T) {
	if os.Getenv("ETCD_PORT_2379_TCP_ADDR") == "" {
		t.Skip("skipping test; $ETCD_PORT_2379_TCP_ADDR not set")
	}

	t.Parallel()
	client, err := getEtcdClient()
	require.NoError(t, err)
	runTest(t, client)
}

func TestEtcdWatch(t *testing.T) {
	if os.Getenv("ETCD_PORT_2379_TCP_ADDR") == "" {
		t.Skip("skipping test; $ETCD_PORT_2379_TCP_ADDR not set")
	}

	t.Parallel()
	client, err := getEtcdClient()
	require.NoError(t, err)
	runWatchTest(t, client)
}

func runTest(t *testing.T, client Client) {
	err := client.Set("foo", "one", 0)
	require.NoError(t, err)
	value, err := client.Get("foo")
	require.NoError(t, err)
	require.Equal(t, "one", value)

	err = client.Set("a/b/foo", "one", 0)
	require.NoError(t, err)
	err = client.Set("a/b/bar", "two", 0)
	require.NoError(t, err)
	values, err := client.GetAll("a/b")
	require.NoError(t, err)
	require.Equal(t, map[string]string{"a/b/foo": "one", "a/b/bar": "two"}, values)

	require.NoError(t, client.Close())
}

func runWatchTest(t *testing.T, client Client) {
	cancel := make(chan bool)
	err := client.WatchAll(
		"watchAll/foo",
		cancel,
		func(value map[string]string) error {
			if value == nil {
				return client.Set("watchAll/foo/bar", "quux", 0)
			}
			require.Equal(t, map[string]string{"watchAll/foo/bar": "quux"}, value)
			close(cancel)
			return nil
		},
	)
	require.True(t, errors.Is(err, ErrCancelled))
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
	return fmt.Sprintf("http://%s:2379", etcdAddr), nil
}
