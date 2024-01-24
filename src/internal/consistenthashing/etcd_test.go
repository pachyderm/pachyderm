package consistenthashing

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/internal/collection"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"

	etcd "go.etcd.io/etcd/client/v3"

	"golang.org/x/sync/errgroup"
)

type testRingConfig struct {
	client *etcd.Client
	ctx    context.Context
	cancel context.CancelFunc
}

type lockTestConfig struct {
	ctx            context.Context
	cancel         context.CancelFunc
	eg             *errgroup.Group
	locks          int
	keys           *sync.Map
	workers        int
	workerIds      map[int]string
	workersReady   chan struct{}
	ringConfig     testRingConfig
	beginLocking   chan struct{}
	doneLocking    chan struct{}
	beginUnlocking chan struct{}
	doneUnlocking  chan struct{}
}

func setupTest(t *testing.T) testRingConfig {
	ctx, cancel := pctx.WithCancel(pctx.TestContext(t))
	etcdEnv := testetcd.NewEnv(ctx, t)
	return testRingConfig{
		client: etcdEnv.EtcdClient,
		ctx:    ctx,
		cancel: cancel,
	}
}

func TestCleanShutdown(t *testing.T) {
	config := setupTest(t)
	err := WithRing(config.ctx, config.client, "master", func(ctx context.Context, ring *Ring) error { return nil })
	require.NoError(t, err, "should not fail")
}

func TestGracefulShutdown(t *testing.T) {
	config := setupTest(t)
	err := WithRing(config.ctx, config.client, "master", func(ctx context.Context, ring *Ring) error {
		return errors.EnsureStack(errors.New("fail test"))
	})
	require.YesError(t, err)
	require.Equal(t, err.Error(), "fail test")
}

func TestWatch(t *testing.T) {
	config := setupTest(t)
	defer config.cancel()
	collection.DefaultTTL = 1
	err := WithRing(config.ctx, config.client, "master", func(ctx context.Context, ring *Ring) error {
		err := WithRing(config.ctx, config.client, "master", func(ctx context.Context, innerRing *Ring) error {
			time.Sleep(2 * time.Second)
			require.Len(t, ring.MemberIds(), 2, "there should be 2 total members")
			return nil
		})
		require.NoError(t, err, "should be able to create second ring")
		time.Sleep(2 * time.Second)
		require.Len(t, ring.MemberIds(), 1, "there should be only 1 member")
		return nil
	})
	require.NoError(t, err, "should be able to create first ring")
}

func TestLocking(t *testing.T) {
	test := setupLockTest(t, 3, 9)
	defer func() {
		test.cancel()
	}()
	for i := 0; i < test.workers; i++ {
		nodeId := test.workerIds[i]
		test.eg.Go(func() error {
			test.runWorker(test.ctx, t, nodeId)
			return nil
		})
	}
	test.lockAllLocks()
	locksPerNode := test.locksPerWorker()
	test.unlockAllLocks()
	for _, num := range locksPerNode {
		require.Equal(t, num, test.locks/test.workers)
	}
}

func TestLockingWithDeleteWorker(t *testing.T) {
	test := setupLockTest(t, 3, 9)
	defer func() {
		test.cancel()
		require.NoError(t, test.eg.Wait())
	}()
	nodeToDelete := strconv.Itoa(100)
	deleteCtx, deleteCancel := pctx.WithCancel(test.ctx)
	defer deleteCancel()
	for i := 0; i < test.workers; i++ {
		nodeId := test.workerIds[i]
		test.eg.Go(func() error {
			if nodeId == nodeToDelete {
				test.runWorker(deleteCtx, t, nodeId)
			} else {
				test.runWorker(test.ctx, t, nodeId)
			}

			return nil
		})
	}
	test.lockAllLocks()
	locksPerNode := test.locksPerWorker()
	require.True(t, locksPerNode[nodeToDelete] > 0, "node to delete should have some locks")
	rebalancedLocks := locksPerNode[nodeToDelete]
	deleteCancel()
	for i := 0; i < rebalancedLocks; i++ {
		<-test.doneLocking
	}
	locksPerNode = test.locksPerWorker()
	require.True(t, locksPerNode[nodeToDelete] == 0, "deleted node should have 0 locks")
	test.unlockAllLocks()
}

func TestLockingWithAddWorker(t *testing.T) {
	test := setupLockTest(t, 3, 9)
	test.workers-- // allows test to generate keys with a hash that should rebalance locks/members locks
	defer func() {
		test.cancel()
		require.NoError(t, test.eg.Wait())
	}()
	for i := 0; i < test.workers; i++ {
		nodeId := test.workerIds[i]
		test.eg.Go(func() error {
			test.runWorker(test.ctx, t, nodeId)
			return nil
		})
	}
	test.lockAllLocks()
	test.workers++
	memberToAdd := strconv.Itoa((test.workers - 1) * 100)
	test.workerIds[test.workers-1] = memberToAdd
	test.beginLocking <- struct{}{}
	test.eg.Go(func() error {
		test.runWorker(test.ctx, t, memberToAdd)
		return nil
	})
	rebalancedLocks := test.locks / test.workers
	for i := 0; i < rebalancedLocks; i++ {
		<-test.doneLocking
	}
	locksPerMember := test.locksPerWorker()
	require.Equal(t, locksPerMember[memberToAdd], rebalancedLocks)
	test.unlockAllLocks()
}

// This hash function should get us pretty close to an even distribution,
// given how lock tests define our key and node names.
func hashFnForTests(t *testing.T) func([]byte) uint32 {
	return func(key []byte) uint32 {
		key = []byte(path.Base(string(key)))
		keyAsInt, err := strconv.Atoi(string(key))
		require.NoError(t, err, "conversion should succeed")
		return uint32(keyAsInt)
	}
}

func setupLockTest(t *testing.T, numNodes, numLocks int) lockTestConfig {
	hashFn = hashFnForTests(t)
	// number of goroutines is test.Nodes * test.locks, so we should test with fairly small numbers.
	test := lockTestConfig{
		locks:      numLocks,
		keys:       &sync.Map{},
		workers:    numNodes,
		workerIds:  map[int]string{},
		ringConfig: setupTest(t),
	}
	test.eg, test.ctx = errgroup.WithContext(test.ringConfig.ctx)
	test.ctx, test.cancel = pctx.WithCancel(test.ctx)
	test.workersReady = make(chan struct{}, numNodes)
	test.beginLocking = make(chan struct{}, numNodes)
	test.doneLocking = make(chan struct{}, numLocks)
	test.beginUnlocking = make(chan struct{}, numLocks)
	test.doneUnlocking = make(chan struct{}, numLocks)
	// starting from 1 is intentional here, allows the lock distribution to be even.
	for i := 1; i <= test.locks; i++ {
		key := strconv.Itoa(i%test.workers*100 + i)
		test.keys.Store(key, "")
	}
	for i := 0; i < test.workers; i++ {
		test.workerIds[i] = strconv.Itoa(i * 100)
	}
	return test
}

func (test *lockTestConfig) lockAllLocks() {
	for i := 0; i < test.workers; i++ {
		<-test.workersReady
	}
	for i := 0; i < test.workers; i++ {
		test.beginLocking <- struct{}{}
	}
	for i := 0; i < test.locks; i++ {
		<-test.doneLocking
	}
}

func (test *lockTestConfig) unlockAllLocks() {
	for i := 0; i < test.locks; i++ {
		test.beginUnlocking <- struct{}{}
	}
	for i := 0; i < test.locks; i++ {
		<-test.doneUnlocking
	}
}

func (test *lockTestConfig) locksPerWorker() map[string]int {
	locksPerWorker := map[string]int{}
	for _, worker := range test.workerIds {
		locksPerWorker[worker] = 0
	}
	test.keys.Range(func(k any, v any) bool {
		node := v.(string)
		locksPerWorker[node]++

		return true // return true acts as continue in a sync.Map.Range()
	})
	return locksPerWorker
}

func (test *lockTestConfig) runWorker(ctx context.Context, t *testing.T, id string) {
	collection.DefaultTTL = 1
	eg, ctx := errgroup.WithContext(pctx.Child(ctx, "worker."+id))
	err := withRing(ctx, test.ringConfig.client, "master", id, func(ctx context.Context, ring *Ring) error {
		time.Sleep(time.Second * 1)
		test.workersReady <- struct{}{}
		<-test.beginLocking
		test.keys.Range(func(k any, v any) bool {
			key := k.(string)
			eg.Go(func() error {
				return errors.EnsureStack(test.lockAndUnlock(ctx, t, ring, key, id))
			})
			return true // return true acts as continue in a sync.Map.Range()
		})
		return errors.EnsureStack(eg.Wait())
	})
	require.NoError(t, err, "withRing should not fail right?")
}

func (test *lockTestConfig) lockAndUnlock(ctx context.Context, t *testing.T, ring *Ring, key string, memberId string) error {
	_, err := ring.Lock(ctx, key)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
	test.keys.Store(key, memberId)
	log.Debug(ctx, fmt.Sprintf("worker %s got lock %s", memberId, key))

	test.doneLocking <- struct{}{}
	select {
	case <-test.beginUnlocking:
	case <-ctx.Done():
		return nil
	}
	err = ring.Unlock(key)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		log.Error(ctx, "should be able to unlock key", zap.Error(err))
	}
	require.NoError(t, err, "should be able to unlock key")
	log.Debug(ctx, fmt.Sprintf("worker %s unlocked %s", memberId, key), zap.Error(err))
	test.doneUnlocking <- struct{}{}
	return nil
}
