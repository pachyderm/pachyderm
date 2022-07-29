package dlock

import (
	"context"
	"testing"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"

	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"

	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
)

func TestUnlockBeforeLock(t *testing.T) {
	env := testetcd.NewEnv(t)
	lock := NewDLock(env.EtcdClient, "test")
	err := lock.Unlock(context.Background())
	require.NoError(t, err)
}

func TestDoubleUnlock(t *testing.T) {
	env := testetcd.NewEnv(t)
	lock := NewDLock(env.EtcdClient, "test")
	_, err := lock.Lock(context.Background())
	require.NoError(t, err)
	err = lock.Unlock(context.Background())
	require.NoError(t, err)
	err = lock.Unlock(context.Background())
	require.ErrorIs(t, err, rpctypes.ErrLeaseNotFound)
}

func TestIsLocked(t *testing.T) {
	env := testetcd.NewEnv(t)
	lock1 := NewDLock(env.EtcdClient, "test")
	lock2 := NewDLock(env.EtcdClient, "test2")
	_, err := lock1.Lock(context.Background())
	require.NoError(t, err)
	require.True(t, lock1.IsLocked())
	require.False(t, lock2.IsLocked())
}

func TestTryLock(t *testing.T) {
	env := testetcd.NewEnv(t)
	lock1 := NewDLock(env.EtcdClient, "test")
	lock2 := NewDLock(env.EtcdClient, "test")
	_, err := lock1.Lock(context.Background())
	require.NoError(t, err)
	_, err = lock2.TryLock(context.Background())
	require.ErrorIs(t, err, concurrency.ErrLocked)
}
