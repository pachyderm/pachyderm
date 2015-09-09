package discovery

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/pkg/require"
)

func TestEtcdClient(t *testing.T) {
	t.Parallel()
	client, err := getEtcdClient()
	require.NoError(t, err)
	runTest(t, client)
}

func TestEtcdWatch(t *testing.T) {
	t.Parallel()
	client, err := getEtcdClient()
	require.NoError(t, err)
	runWatchTest(t, client)
}

func runTest(t *testing.T, client Client) {
	_, err := client.Set("foo", "one", 0)
	require.NoError(t, err)
	value, ok, err := client.Get("foo")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "one", value)
	//values, err := client.GetAll("foo")
	//require.NoError(t, err)
	//require.Equal(t, map[string]string{"foo": "one"}, values)

	_, err = client.Set("a/b/foo", "one", 0)
	require.NoError(t, err)
	_, err = client.Set("a/b/bar", "two", 0)
	require.NoError(t, err)
	values, err := client.GetAll("a/b")
	require.NoError(t, err)
	require.Equal(t, map[string]string{"a/b/foo": "one", "a/b/bar": "two"}, values)

	require.NoError(t, client.Close())
}

func runWatchTest(t *testing.T, client Client) {
	cancel := make(chan bool)
	err := client.Watch(
		"watch/foo",
		cancel,
		func(value string) (uint64, error) {
			if value == "" {
				return client.Set("watch/foo", "bar", 0)
			}
			require.Equal(t, "bar", value)
			close(cancel)
			return 0, nil
		},
	)
	require.Equal(t, ErrCancelled, err)

	cancel = make(chan bool)
	err = client.WatchAll(
		"watchAll/foo",
		cancel,
		func(value map[string]string) (uint64, error) {
			if value == nil {
				return client.Set("watchAll/foo/bar", "quux", 0)
			}
			require.Equal(t, map[string]string{"watchAll/foo/bar": "quux"}, value)
			close(cancel)
			return 0, nil
		},
	)
	require.Equal(t, ErrCancelled, err)
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
