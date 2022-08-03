package consistenthashing

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"

	"github.com/pachyderm/pachyderm/v2/src/internal/watch"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
)

const (
	// LeaseTTL is used by nodes in a ring to refresh their liveness state.
	LeaseTTL = int64(60)
	// ErrNotLocal is returned when attempting to delete a node that is remote.
	ErrNotLocal = "remote node cannot be deleted. Only local nodes may be deleted"
)

// Ring is a consistent hash Ring. Each process should only create one instance with a given prefix. Nodes can be
// added locally to a Ring. A Ring watches for changes to the prefix to determine if new nodes have been added
// or deleted.
// A Ring should be initialized with a call to New().
type Ring struct {
	Client *etcd.Client        // Client will be used to make api calls against etcd.
	Prefix string              // Prefix is prepended to node keys stored in etcd to avoid naming conflicts.
	HashFn func([]byte) uint32 // HashFn will be called to determine the hash for a given key.
	Logger *logrus.Logger      // Logger is used to output log messages if defined.

	// localNodes represents the nodes in a Ring that are local to the current process. In practice for most use cases,
	// a given process is only going to have one node. However, multiple local nodes are supported for testing
	// purposes or if a user wants to abstract a node as multiple goroutines.
	localNodes map[string]*node

	nodeInfos []NodeInfo // nodeInfos is a string array of all node Ids in the Ring sorted by crc32 checksum.

	Cancel   context.CancelFunc // Cancel can be used bya caller to stop Ring subroutines.
	Ctx      context.Context    // Ctx is the context used to create the Ring.
	ErrGroup *errgroup.Group    // ErrGroup is used for waiting on subroutines to terminate.
}

// node is a member of a Ring. Upon creation, a node requests a lease from etcd and continuously refreshes it until it
// is shutdown.
type node struct {
	Cancel context.CancelFunc // Cancel can be used by a caller to stop node subroutines.
	Ctx    context.Context    // Ctx is the context used to create the node.
	NodeInfo
	Locks sync.Map // Locks is a thread-safe map of lockInfos associated with a node.
}

// NodeInfo contains metadata about nodes in a Ring.
type NodeInfo struct {
	Id   string // Id is the name of a node.
	Hash uint32 // Hash is the output of the ring's HashFn(Id).
}

// lockInfo instances are created for keys by local nodes in a ring.
// A context lockInfo.Ctx is stored so the proper context is used when unlocking lockInfo.Lock
type lockInfo struct {
	Lock dlock.DLock     // Lock is a dlock that locks a key with a prefix in etcd.
	Ctx  context.Context // Ctx is the context returned by Lock.
}

// New is the default constructor for Ring. Creating a Ring, starts a goroutine to watch prefix in etcd.
func New(ctx context.Context, prefix string, client *etcd.Client, logger *logrus.Logger, hashFn func([]byte) uint32) (*Ring, error) {
	ring := &Ring{
		Client: client,
		Prefix: prefix,
		HashFn: hashFn,
		Logger: logger,

		localNodes: map[string]*node{},
	}

	ring.ErrGroup, ctx = errgroup.WithContext(ctx)
	ring.Ctx, ring.Cancel = context.WithCancel(ctx)

	ring.ErrGroup.Go(func() error {
		return ring.watch(ring.Ctx, ring.Cancel)
	})

	return ring, nil
}

// AddNode idempotently creates a node in a consistent hashing Ring. The 'namespace' of the node is the prefix.
// The ID is the actual ID of the node. In the backend, we create an etcd key to represent the node.
// The node currently: creates a lease, inserts itself as a key to etcd, keeps the lease alive in the
// background, and watches all etcd keys with the prefix 'prefix' for changes. AddNode is idempotent, so if a node
// already exists with the same id and prefix, no change is made to etcd and no additional goroutines are started.
func (ring *Ring) AddNode(ctx context.Context, id string) error {
	nodeKey := ring.nodeIdToEtcdKey(id)

	// Check to see if a node by this name already exists, if so return since the node has already been added.
	getResp, err := ring.Client.Get(ctx, nodeKey)
	if err != nil {
		return errors.EnsureStack(err)
	}

	for _, kv := range getResp.Kvs {
		if string(kv.Key) == nodeKey {
			return nil // key already exists.
		}
	}

	grantResp, err := ring.Client.Grant(ctx, LeaseTTL)
	if err != nil {
		return errors.EnsureStack(err)
	}

	ctx, cancel := context.WithCancel(ctx)

	keepAliveChan, err := ring.Client.KeepAlive(ctx, grantResp.ID)
	if err != nil {
		cancel()
		return errors.EnsureStack(err)
	}

	ring.ErrGroup.Go(func() error {
		defer cancel()

		for {
			_, more := <-keepAliveChan
			if !more {
				if ctx.Err() == nil {
					ring.Logger.WithFields(
						map[string]interface{}{
							"node": nodeKey,
							"ring": ring.Prefix,
						}).Errorf("failed to renew etcd lease")
				}

				ring.Logger.WithFields(
					map[string]interface{}{
						"node": nodeKey,
						"ring": ring.Prefix,
					}).Info("shutting down")
				return nil
			}
		}
	})

	if _, err := ring.Client.Put(ctx, nodeKey, "true", etcd.WithLease(grantResp.ID)); err != nil {
		return errors.EnsureStack(err)
	}

	localNodeInfo := NodeInfo{
		Id:   id,
		Hash: ring.HashFn([]byte(id)),
	}

	ring.insertNodeInfo(localNodeInfo)

	ring.localNodes[id] = &node{
		Cancel:   cancel,
		Ctx:      ctx,
		NodeInfo: localNodeInfo,
		Locks:    sync.Map{},
	}

	return nil
}

// DeleteNode deletes the node entry from etcd if and only if the node is a local node.
// If the node exists, then it should trigger a graceful terminate of the node's goroutines.
func (ring *Ring) DeleteNode(ctx context.Context, localId string) error {
	if ring.Client == nil {
		return errors.New("etcd client is nil")
	}

	node, isLocal := ring.localNodes[localId]
	if !isLocal {
		return errors.New(ErrNotLocal)
	}

	nodeKey := ring.nodeIdToEtcdKey(localId)

	if _, err := ring.Client.Delete(ctx, nodeKey); err != nil {
		return errors.EnsureStack(err)
	}

	ring.removeNodeInfoById(localId)

	var errs []error

	node.Locks.Range(func(prefix any, info any) bool {
		lockInfo := info.(lockInfo)
		if err := lockInfo.Lock.Unlock(ctx); err != nil {
			errs = append(errs, err)
		}

		return true // return true acts as continue in a sync.Map.Range()
	})

	if errs != nil {
		return errors.New(fmt.Sprintf("errors while unlocking locks while deleting node: %v", errs))
	}

	delete(ring.localNodes, localId)

	node.Cancel()

	return nil
}

// GetIndex returns the index of the node associated with key. A key belongs to the first node whose hash is
// greater than the key. If no nodeInfos are greater, then the Ring wraps around to the first node.
func (ring *Ring) GetIndex(key string) int {
	keyHash := ring.HashFn([]byte(key))
	index := sort.Search(len(ring.nodeInfos), func(i int) bool {
		return ring.nodeInfos[i].Hash >= keyHash
	})

	// wrap around to the first node in the Ring.
	if index >= len(ring.nodeInfos) {
		index = 0
	}

	return index
}

// Get is a convenience wrapper around GetIndex that returns the NodeInfo that associates to a key.
func (ring *Ring) Get(key string) NodeInfo {
	return ring.nodeInfos[ring.GetIndex(key)]
}

// LocalNodes is an accessor for Ring.localNodes.
func (ring *Ring) LocalNodes() map[string]*node {
	return ring.localNodes
}

// Lock attempts to lock a key in etcd using consistent hashing. If the key belongs to a remote node,
// the call to lock blocks until the context is cancelled or expires.
func (ring *Ring) Lock(ctx context.Context, key string) (context.Context, error) {
	var node *node
	var nodeIsLocal bool

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		index := ring.GetIndex(key)
		node, nodeIsLocal = ring.localNodes[ring.nodeInfos[index].Id]
		if nodeIsLocal {
			break
		}

		select {
		case <-ctx.Done():
			return nil, errors.EnsureStack(ctx.Err())
		case <-ticker.C:
		}
	}

	if info, alreadyExists := node.Locks.Load(key); alreadyExists {
		lockCtx := info.(lockInfo).Ctx
		return lockCtx, nil
	}

	nodeLock := dlock.NewDLock(ring.Client, key)
	lockCtx, err := nodeLock.Lock(ctx)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	node.Locks.Store(key, lockInfo{
		Lock: nodeLock,
		Ctx:  lockCtx,
	})

	ring.localNodes[node.Id] = node

	return lockCtx, nil
}

// Unlock attempts to unlock a key in etcd if and only if a local node owns a lock reference to it.
// The node must be running locally in our process. Remote nodes must give up their own locks.
func (ring *Ring) Unlock(key string) error {
	index := ring.GetIndex(key)

	node, isLocal := ring.localNodes[ring.nodeInfos[index].Id]
	if !isLocal {
		// lock should not exist in this local process. This might happen if a lock was rebalanced away.
		// remove lock with key from local nodes
		for _, localNode := range ring.localNodes {
			info, staleLock := localNode.Locks.Load(key)
			if !staleLock {
				continue
			}

			lockInfo := info.(lockInfo)

			if err := lockInfo.Lock.Unlock(lockInfo.Ctx); err != nil {
				return errors.EnsureStack(err)
			}

			localNode.Locks.Delete(key)
		}

		return nil
	}

	info, localExists := node.Locks.Load(key)
	if !localExists {
		return nil // lock does not exist in this local process.
	}

	lockInfo := info.(lockInfo)
	if err := lockInfo.Lock.Unlock(lockInfo.Ctx); err != nil {
		return errors.EnsureStack(err)
	}

	node.Locks.Delete(key)

	ring.localNodes[node.Id] = node

	return nil
}

// insertNodeInfoById is a convenience wrapper for insertNodeInfo that constructs the node on behalf of the caller.
func (ring *Ring) insertNodeInfoById(id string) {
	node := NodeInfo{
		Id:   id,
		Hash: ring.HashFn([]byte(id)),
	}

	ring.insertNodeInfo(node)
}

// insertNodeInfo updates Ring.nodeInfos, and sorts Ring.nodeInfos by the node hash.
// If the node already exists, insertNodeInfoById is a no-op. node can be mocked for testing purposes.
func (ring *Ring) insertNodeInfo(node NodeInfo) {
	index := sort.Search(len(ring.nodeInfos), func(i int) bool {
		return ring.nodeInfos[i].Hash >= node.Hash
	})

	if index < len(ring.nodeInfos) && ring.nodeInfos[index].Hash == node.Hash {
		return // node already exists.
	}

	if index >= len(ring.nodeInfos) {
		ring.nodeInfos = append(ring.nodeInfos, node)
		return
	}

	ring.nodeInfos = append(ring.nodeInfos[:index+1], ring.nodeInfos[index:]...)
	ring.nodeInfos[index] = node
}

func (ring *Ring) nodeIdToEtcdKey(id string) string {
	return path.Join(ring.Prefix, id)
}

func (ring *Ring) nodeEtcdKeyToId(fullId string) string {
	return strings.TrimPrefix(fullId, ring.Prefix+"/")
}

// removeNodeInfoById is a convenience wrapper for removeNodeInfo that constructs the node on behalf of the caller.
func (ring *Ring) removeNodeInfoById(id string) {
	node := NodeInfo{
		Id:   id,
		Hash: ring.HashFn([]byte(id)),
	}

	ring.removeNodeInfo(node)
}

// removeNodeInfo removes the node from NodeHash and nodeInfos if the node exists.
func (ring *Ring) removeNodeInfo(node NodeInfo) {
	index := sort.Search(len(ring.nodeInfos), func(i int) bool {
		return ring.nodeInfos[i].Hash >= node.Hash
	})

	if ring.nodeInfos == nil || index >= len(ring.nodeInfos) || ring.nodeInfos[index].Id != node.Id {
		return // node doesn't exist.
	}

	ring.nodeInfos = append(ring.nodeInfos[:index], ring.nodeInfos[index+1:]...)
}

// watch watches etcd keys with prefix Ring.Prefix to determine if a node has been added or removed.
// If so, it triggers a rebalance event to redistribute locks.
func (ring *Ring) watch(ctx context.Context, cancel context.CancelFunc) error {
	if ring.Client == nil {
		return errors.New("etcd client is nil")
	}

	nodeWatcher, err := watch.NewEtcdWatcher(ctx, ring.Client, "", ring.Prefix+"/", nil)
	if err != nil {
		return errors.EnsureStack(err)
	}

	// Again, not sure if this should live here or somewhere else
	go func() {
		defer cancel()

		for event := range nodeWatcher.Watch() {
			if event.Type == watch.EventError {
				if errors.Is(event.Err, context.Canceled) || errors.Is(event.Err, context.DeadlineExceeded) {
					ring.Logger.WithField("ring", ring.Prefix).Infof("shutting down ring")
					return
				}

				ring.Logger.WithField("ring", ring.Prefix).Errorf("encountered an error: %v", event.Err)
				continue
			}

			nodeKey := string(event.Key)

			switch event.Type {
			case watch.EventDelete:
				ring.Logger.WithField("ring", ring.Prefix).Infof("deleting node %s from ring", nodeKey)
				ring.removeNodeInfoById(ring.nodeEtcdKeyToId(nodeKey))

				// Rebalance on delete happens by default since each node attempts to get all locks. When a node
				// is deleted, the pending calls to lock() by each node will go through on the nodes that should now
				// own a given lock.

			case watch.EventPut:
				ring.Logger.WithField("ring", ring.Prefix).Infof("adding node %s to ring", nodeKey)
				ring.insertNodeInfoById(ring.nodeEtcdKeyToId(nodeKey))

				// Need to check our localNodes to release locks that the nodes own when another remote node is added.
				if err := ring.rebalanceLocks(); err != nil {
					ring.Logger.WithField("ring", ring.Prefix).Errorf("encountered and error while rebalancing: %v", err)
				}
			}
		}
	}()

	return nil
}

func (ring *Ring) rebalanceLocks() error {
	var errs []error
	for _, node := range ring.localNodes {
		node.Locks.Range(func(k, v any) bool {
			prefix := k.(string)
			if ring.Get(prefix).Id == node.Id {
				return true // return true acts as continue in a sync.Map.Range()
			}

			lockInfo := v.(lockInfo)
			if err := lockInfo.Lock.Unlock(lockInfo.Ctx); err != nil {
				errs = append(errs, errors.EnsureStack(err))
			}

			node.Locks.Delete(prefix)
			return true
		})
	}

	if errs != nil {
		return errors.New(fmt.Sprintf("errors while unlocking locks while rebalancing node: %v", errs))
	}

	return nil
}
