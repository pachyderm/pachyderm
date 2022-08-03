package consistenthashing

import (
	"context"
	"hash/crc32"
	"math/rand"
	"path"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"

	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"

	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
)

type TestRingConfig struct {
	Client *clientv3.Client
	Ctx    context.Context
	Cancel context.CancelFunc
}

type LockTestConfig struct {
	Cancel   context.CancelFunc
	Ctx      context.Context
	ErrGroup *errgroup.Group

	HashFn  func([]byte) uint32
	Locks   int
	Keys    *sync.Map
	Nodes   int
	NodeIds map[int]string

	RingConfig TestRingConfig

	NodesReady chan bool

	BeginLocking chan bool
	DoneLocking  chan bool

	BeginUnlocking chan bool
	DoneUnlocking  chan bool
}

func TestAddNodes(t *testing.T) {
	config := ringConfig(t)
	defer config.Cancel()

	ring, err := New(config.Ctx, "master", config.Client, &logrus.Logger{}, crc32.ChecksumIEEE)
	require.NoError(t, err, "should be able to create Ring")

	node1Id := uuid.New()
	err = ring.AddNode(config.Ctx, node1Id)
	require.NoError(t, err, "should be able to create a node")
	require.Len(t, ring.nodeInfos, 1, "Ring.nodeInfos should have 1 node")

	node2Id := uuid.New()
	err = ring.AddNode(config.Ctx, node2Id)
	require.NoError(t, err, "should be able to create a node")
	require.Len(t, ring.nodeInfos, 2, "Ring.nodeInfos should have 2 nodeInfos")

	err = ring.DeleteNode(config.Ctx, node2Id)
	require.NoError(t, err, "should be able to create a node")
	require.Len(t, ring.nodeInfos, 1, "Ring.nodeInfos should have 1 node")
}

func TestAddNodeShutdown(t *testing.T) {
	config := ringConfig(t)

	nodeReady := make(chan bool)

	testRing, err := New(config.Ctx, "master", config.Client, &logrus.Logger{}, crc32.ChecksumIEEE)
	defer func() {
		testRing.Cancel()
		for _, node := range testRing.localNodes {
			node.Cancel()
		}
		require.NoError(t, testRing.ErrGroup.Wait(), "no errors are expected")

	}()

	require.NoError(t, err, "should be able to create testRing")

	testRing.ErrGroup.Go(func() error {
		err := testRing.AddNode(config.Ctx, uuid.New())
		nodeReady <- true
		return err
	})

	<-nodeReady
}

func TestDoubleAddNode(t *testing.T) {
	config := ringConfig(t)
	defer config.Cancel()

	ring, err := New(config.Ctx, "master", config.Client, &logrus.Logger{}, crc32.ChecksumIEEE)
	require.NoError(t, err, "should be able to create Ring")

	node1 := uuid.New()
	err = ring.AddNode(config.Ctx, node1)
	require.NoError(t, err, "should be able to create a node")

	err = ring.AddNode(config.Ctx, node1)
	require.NoError(t, err, "should be a no-op")

	require.True(t, len(ring.nodeInfos) == 1, "node should not have been re-inserted")
}

func TestAddNodeWatch(t *testing.T) {
	config := ringConfig(t)
	defer config.Cancel()

	ring, err := New(config.Ctx, "master", config.Client, &logrus.Logger{}, crc32.ChecksumIEEE)
	require.NoError(t, err, "should be able to create Ring")

	err = ring.AddNode(config.Ctx, uuid.New())
	require.NoError(t, err, "should be able to create a node")

	_, err = config.Client.Put(config.Ctx, "master/fake", "true")
	require.NoError(t, err, "should be able to put")
	time.Sleep(1 * time.Second)
	require.Len(t, ring.nodeInfos, 2, "there should be 2 total nodeInfos")

	_, err = config.Client.Delete(config.Ctx, "master/fake")
	require.NoError(t, err, "should be able to delete")
	time.Sleep(1 * time.Second)
	require.Len(t, ring.nodeInfos, 1, "there should be 1 nodeInfo")
}

func TestInsertAndRemoveNodeInfos(t *testing.T) {
	ring := Ring{}

	var nodesToInsert []NodeInfo
	nodeCount := 10
	for i := 0; i < nodeCount; i++ {
		nodesToInsert = append(nodesToInsert, NodeInfo{
			Id:   strconv.Itoa(i),
			Hash: uint32(i),
		})
	}

	// randomize the array, so we don't insert in sorted order.
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(nodesToInsert), func(i, j int) {
		nodesToInsert[i], nodesToInsert[j] = nodesToInsert[j], nodesToInsert[i]
	})

	for _, nodeInfo := range nodesToInsert {
		ring.insertNodeInfo(nodeInfo)
	}

	isSorted := sort.SliceIsSorted(ring.nodeInfos, func(i, j int) bool {
		return ring.nodeInfos[i].Hash <= ring.nodeInfos[j].Hash
	})

	require.True(t, isSorted, "Ring.nodeInfos should be sorted")

	var nodesToRemove []NodeInfo
	deleteCount := 3
	indices := map[int]bool{}
	for i := 0; i < deleteCount; i++ {
		randIndex := rand.Intn(nodeCount)
		_, alreadyUsed := indices[randIndex]

		for alreadyUsed {
			randIndex = rand.Intn(nodeCount)
			_, alreadyUsed = indices[randIndex]
		}
		indices[randIndex] = true

		nodesToRemove = append(nodesToRemove, nodesToInsert[randIndex])
	}

	for _, nodeInfo := range nodesToRemove {
		ring.removeNodeInfo(nodeInfo)
	}

	isSorted = sort.SliceIsSorted(ring.nodeInfos, func(i, j int) bool {
		return ring.nodeInfos[i].Hash <= ring.nodeInfos[j].Hash
	})

	require.True(t, isSorted, "Ring.nodeInfos should be sorted")
}

func TestDeleteNode(t *testing.T) {
	config := ringConfig(t)
	defer config.Cancel()

	ring, err := New(config.Ctx, "master", config.Client, &logrus.Logger{}, crc32.ChecksumIEEE)
	require.NoError(t, err, "should be able to create Ring")

	nodeId := uuid.New()
	err = ring.AddNode(config.Ctx, nodeId)
	require.NoError(t, err, "should be able to create a nodeId")
	require.Len(t, ring.nodeInfos, 1, "Ring.nodeInfos should have 1 nodeId")

	err = ring.DeleteNode(config.Ctx, nodeId)
	require.NoError(t, err, "should be able to delete a nodeId")
	require.Len(t, ring.nodeInfos, 0, "Ring.nodeInfos should be empty")

	resp, err := config.Client.Get(config.Ctx, ring.nodeIdToEtcdKey(nodeId))
	require.NoError(t, err, "should be able to get key from etcd")
	for _, kv := range resp.Kvs {
		require.False(t, string(kv.Key) == ring.nodeIdToEtcdKey(nodeId), "key should be deleted")
	}
}

func TestDeleteNodeDoesntExist(t *testing.T) {
	config := ringConfig(t)
	defer config.Cancel()

	ring, err := New(config.Ctx, "master", config.Client, &logrus.Logger{}, crc32.ChecksumIEEE)
	require.NoError(t, err, "should be able to create Ring")

	err = ring.DeleteNode(config.Ctx, uuid.New())
	require.YesError(t, err, "should fail since node is already deleted")
	require.Equal(t, err.Error(), ErrNotLocal)
}

func TestDeleteRemoteNode(t *testing.T) {
	config := ringConfig(t)
	defer config.Cancel()

	localRing, err := New(config.Ctx, "master", config.Client, &logrus.Logger{}, crc32.ChecksumIEEE)
	require.NoError(t, err, "should be able to create Ring")

	remoteRing, err := New(config.Ctx, "master", config.Client, &logrus.Logger{}, crc32.ChecksumIEEE)
	require.NoError(t, err, "should be able to create Ring")

	err = localRing.AddNode(config.Ctx, uuid.New())
	require.NoError(t, err, "should be able to create a node")
	require.Len(t, localRing.nodeInfos, 1, "localRing.nodeInfos should have 1 nodeInfo")
	require.Len(t, localRing.localNodes, 1, "localRing.localNodes should have 1 node")

	node2Id := uuid.New()
	err = remoteRing.AddNode(config.Ctx, node2Id)
	require.NoError(t, err, "should be able to create a node")

	time.Sleep(time.Second * 3) // give some time for the rings to sync their status.

	require.Len(t, localRing.nodeInfos, 2, "localRing.nodeInfos should have 2 nodeInfos")
	require.Len(t, remoteRing.nodeInfos, 2, "remoteRing.nodeInfos should have 2 nodeInfos")
	require.Len(t, localRing.localNodes, 1, "localRing.localNodes should have 1 node")
	require.Len(t, remoteRing.localNodes, 1, "remoteRing.localNodes should have 1 node")

	err = localRing.DeleteNode(config.Ctx, node2Id)
	require.YesError(t, err, "should fail since node is remote")
	require.Equal(t, err.Error(), ErrNotLocal)
}

func TestDoubleDelete(t *testing.T) {
	config := ringConfig(t)
	defer config.Cancel()

	ring, err := New(config.Ctx, "master", config.Client, &logrus.Logger{}, crc32.ChecksumIEEE)
	require.NoError(t, err, "should be able to create Ring")

	nodeId := uuid.New()
	err = ring.AddNode(config.Ctx, nodeId)
	require.NoError(t, err, "should be able to create a nodeId")

	err = ring.DeleteNode(config.Ctx, nodeId)
	require.NoError(t, err, "should be able to delete a nodeId")

	err = ring.DeleteNode(config.Ctx, nodeId)
	require.YesError(t, err, "should fail since nodeId is already deleted")
	require.Equal(t, err.Error(), ErrNotLocal)
}

func TestLockAndUnlock(t *testing.T) {
	test := lockTest(t, 3, 9, testHashFn(t))
	defer func() {
		test.Cancel()
		require.NoError(t, test.ErrGroup.Wait())
	}()

	for i := 0; i < test.Nodes; i++ {
		nodeId := test.NodeIds[i]
		test.ErrGroup.Go(func() error {
			test.launchNode(test.Ctx, test.ErrGroup, nodeId, t)
			return nil
		})
	}

	test.lockAllLocks()

	locksPerNode := test.locksPerNode()

	test.unlockAllLocks()

	for _, num := range locksPerNode {
		require.Equal(t, num, test.Locks/test.Nodes)
	}
}

func TestRebalanceDelete(t *testing.T) {
	test := lockTest(t, 3, 9, testHashFn(t))
	defer func() {
		test.Cancel()
		close(test.BeginUnlocking)
		require.NoError(t, test.ErrGroup.Wait())
	}()

	var ringToDelete *Ring
	nodeToDelete := strconv.Itoa(100)

	for i := 0; i < test.Nodes; i++ {
		nodeId := test.NodeIds[i]
		test.ErrGroup.Go(func() error {
			remoteRing := test.launchNode(test.Ctx, test.ErrGroup, nodeId, t)
			if nodeId == nodeToDelete {
				ringToDelete = remoteRing
			}

			return nil
		})
	}

	test.lockAllLocks()

	locksPerNode := test.locksPerNode()
	require.True(t, locksPerNode[nodeToDelete] > 0, "node to delete should have some locks")

	rebalancedLocks := locksPerNode[nodeToDelete]
	err := ringToDelete.DeleteNode(test.Ctx, nodeToDelete)
	require.NoError(t, err, "should be able to delete node")

	for i := 0; i < rebalancedLocks; i++ {
		<-test.DoneLocking
	}

	locksPerNode = test.locksPerNode()
	require.True(t, locksPerNode[nodeToDelete] == 0, "deleted node should have 0 locks")

	test.unlockAllLocks()
}

func TestRebalanceAdd(t *testing.T) {
	test := lockTest(t, 3, 9, testHashFn(t))
	test.Nodes-- // allows test to generate keys with a hash that should rebalance locks/nodes locks

	defer func() {
		test.Cancel()
		close(test.BeginUnlocking)
		require.NoError(t, test.ErrGroup.Wait())
	}()

	for i := 0; i < test.Nodes; i++ {
		nodeId := test.NodeIds[i]
		test.ErrGroup.Go(func() error {
			test.launchNode(test.Ctx, test.ErrGroup, nodeId, t)
			return nil
		})
	}

	test.lockAllLocks()

	test.Nodes++
	nodeToAdd := strconv.Itoa((test.Nodes - 1) * 100)
	test.NodeIds[test.Nodes-1] = nodeToAdd
	test.BeginLocking <- true

	test.ErrGroup.Go(func() error {
		test.launchNode(test.Ctx, test.ErrGroup, nodeToAdd, t)
		return nil
	})

	rebalancedLocks := test.Locks / test.Nodes
	for i := 0; i < rebalancedLocks; i++ {
		<-test.DoneLocking
	}

	locksPerNode := test.locksPerNode()
	require.Equal(t, locksPerNode[nodeToAdd], rebalancedLocks)

	test.unlockAllLocks()
}

// This hash function should get us pretty close to an even distribution,
// given how lock tests define our key and node names.
func testHashFn(t *testing.T) func([]byte) uint32 {
	return func(key []byte) uint32 {
		keyAsInt, err := strconv.Atoi(string(key))
		require.NoError(t, err, "conversion should succeed")

		return uint32(keyAsInt)
	}
}

func ringConfig(t *testing.T) TestRingConfig {
	etcdEnv := testetcd.NewEnv(t)
	ctx, cancel := context.WithCancel(context.Background())

	return TestRingConfig{
		Client: etcdEnv.EtcdClient,
		Ctx:    ctx,
		Cancel: cancel,
	}

}

func lockTest(t *testing.T, numNodes, numLocks int, hashFn func([]byte) uint32) LockTestConfig {
	// number of goroutines is test.Nodes * test.Locks, so we should test with fairly small numbers.
	test := LockTestConfig{
		HashFn:     hashFn,
		Locks:      numLocks,
		Keys:       &sync.Map{},
		Nodes:      numNodes,
		NodeIds:    map[int]string{},
		RingConfig: ringConfig(t),
	}

	test.ErrGroup, test.Ctx = errgroup.WithContext(test.RingConfig.Ctx)
	test.Ctx, test.Cancel = context.WithCancel(test.Ctx)

	test.NodesReady = make(chan bool, numNodes)
	test.BeginLocking = make(chan bool, numNodes)
	test.DoneLocking = make(chan bool, numLocks)

	test.BeginUnlocking = make(chan bool, numLocks)
	test.DoneUnlocking = make(chan bool, numLocks)

	// starting from 1 is intentional here, allows the lock distribution to be even.
	for i := 1; i <= test.Locks; i++ {
		key := strconv.Itoa(i%test.Nodes*100 + i)
		test.Keys.Store(key, "")
	}

	for i := 0; i < test.Nodes; i++ {
		test.NodeIds[i] = strconv.Itoa(i * 100)
	}

	return test
}

func (test *LockTestConfig) lockAllLocks() {
	for i := 0; i < test.Nodes; i++ {
		<-test.NodesReady
	}
	for i := 0; i < test.Nodes; i++ {
		test.BeginLocking <- true
	}
	for i := 0; i < test.Locks; i++ {
		<-test.DoneLocking
	}
}

func (test *LockTestConfig) unlockAllLocks() {
	for i := 0; i < test.Locks; i++ {
		test.BeginUnlocking <- true
	}
	for i := 0; i < test.Locks; i++ {
		<-test.DoneUnlocking
	}
}

func (test *LockTestConfig) locksPerNode() map[string]int {
	locksPerNode := map[string]int{}
	for _, node := range test.NodeIds {
		locksPerNode[node] = 0
	}

	test.Keys.Range(func(k any, v any) bool {
		node := v.(string)
		locksPerNode[node]++

		return true // return true acts as continue in a sync.Map.Range()
	})

	return locksPerNode
}

func (test *LockTestConfig) launchNode(ctx context.Context, eg *errgroup.Group, nodeId string, t *testing.T) *Ring {
	remoteRing, err := New(ctx, "master", test.RingConfig.Client, &logrus.Logger{}, test.HashFn)
	require.NoError(t, err, "should be able to create Ring")

	err = remoteRing.AddNode(ctx, nodeId)
	require.NoError(t, err, "should be able to create a node")

	test.NodesReady <- true
	<-test.BeginLocking

	test.Keys.Range(func(k any, v any) bool {
		key := k.(string)

		eg.Go(func() error {
			_, err := remoteRing.Lock(ctx, key)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					logrus.WithField("node", nodeId).WithField("lock", key).Infof("test shutting down node")
					return nil
				}
				return err
			}

			test.Keys.Store(key, nodeId)

			test.DoneLocking <- true

			if _, ok := <-test.BeginUnlocking; !ok {
				return nil // channel was closed at end of test
			}

			err = remoteRing.Unlock(key)
			require.NoError(t, err, "should be able to unlock key")

			test.DoneUnlocking <- true

			return nil
		})

		return true // return true acts as continue in a sync.Map.Range()
	})

	return remoteRing
}

func ExampleNew() {
	etcdEnv := testetcd.NewEnv(&testing.T{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ring, err := New(ctx, path.Join("master", "ring"), etcdEnv.EtcdClient, &logrus.Logger{}, crc32.ChecksumIEEE)
	if err != nil {
		// handle err
	}

	if err := ring.AddNode(ctx, uuid.New()); err != nil {
		// handle err
	}
}
