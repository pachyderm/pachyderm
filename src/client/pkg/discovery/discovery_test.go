package discovery

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

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
	ctx := context.TODO()
	err := client.Set(ctx, "foo", "one", 0)
	require.NoError(t, err)
	value, err := client.Get(ctx, "foo")
	require.NoError(t, err)
	require.Equal(t, "one", value)

	err = client.Set(ctx, "a/b/foo", "one", 0)
	require.NoError(t, err)
	err = client.Set(ctx, "a/b/bar", "two", 0)
	require.NoError(t, err)
	values, err := client.GetAll(ctx, "a/b")
	require.NoError(t, err)
	require.Equal(t, map[string]string{"a/b/foo": "one", "a/b/bar": "two"}, values)

	require.NoError(t, client.Close())
}

func runWatchTest(t *testing.T, client Client) {
	cancel := make(chan bool)
	ctx := context.TODO()
	err := client.Watch(
		ctx,
		"watch/foo",
		cancel,
		func(value string) error {
			if value == "" {
				return client.Set(ctx, "watch/foo", "bar", 0)
			}
			require.Equal(t, "bar", value)
			close(cancel)
			return nil
		},
	)
	require.Equal(t, ErrCancelled, err)

	cancel = make(chan bool)
	err = client.WatchAll(
		ctx,
		"watchAll/foo",
		cancel,
		func(value map[string]string) error {
			if value == nil {
				return client.Set(ctx, "watchAll/foo/bar", "quux", 0)
			}
			require.Equal(t, map[string]string{"watchAll/foo/bar": "quux"}, value)
			close(cancel)
			return nil
		},
	)
	require.Equal(t, ErrCancelled, err)
}

func getEtcdClient() (Client, error) {
	etcdAddress, err := getEtcdAddress()
	if err != nil {
		return nil, err
	}
	return NewEtcdClient(etcdAddress)
}

func getEtcdAddress() (string, error) {
	etcdAddr := os.Getenv("ETCD_PORT_2379_TCP_ADDR")
	if etcdAddr == "" {
		return "", errors.New("ETCD_PORT_2379_TCP_ADDR not set")
	}
	return fmt.Sprintf("http://%s:2379", etcdAddr), nil
}
