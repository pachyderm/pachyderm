package consistenthashing

import (
	"context"
	"hash/crc32"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
)

var (
	// hashFn is overridable for testing purposes.
	hashFn = crc32.ChecksumIEEE
)

// Ring is a consistent hash Ring. Each process should only create one instance with a given prefix. Nodes can be
// added locally to a Ring. A Ring watches for changes to the prefix to determine if new nodes have been added
// or deleted.
type Ring struct {
	client    *etcd.Client
	stateLock sync.Mutex
	logger    *logrus.Logger
	members   []member
	node      node // node represents the member instance in a Ring that is local to the current process.
	prefix    string
}

// node is a local member of a Ring. Upon creation, a node requests a lease from etcd and continuously refreshes it
// until it is shutdown.
type node struct {
	member
	locks map[string]lockInfo
}

// member contains metadata about members in a Ring.
type member struct {
	Id   string
	hash uint32
}

// lockInfo instances are created for keys by a ring's node.
// A context lockInfo.ctx is stored so the proper context is used when unlocking lockInfo.lock
type lockInfo struct {
	lock dlock.DLock
	ctx  context.Context
}

// MemberIds returns the id of each member in the ring.
func (ring *Ring) MemberIds() []string {
	ring.stateLock.Lock()
	defer ring.stateLock.Unlock()
	var ids []string
	for _, member := range ring.members {
		ids = append(ids, member.Id)
	}
	return ids
}

// WithRing instantiates a ring that lives for the duration of the callback 'cb'. While a ring instance is live,
// it refreshes a lease in etcd and watches the ring prefix for add member and delete member events.
func WithRing(ctx context.Context, client *etcd.Client, logger *logrus.Logger, prefix string, cb func(ctx context.Context, ring *Ring) error) error {
	return withRing(ctx, client, logger, prefix, uuid.New(), cb)
}

func withRing(ctx context.Context, client *etcd.Client, logger *logrus.Logger, prefix, id string, cb func(ctx context.Context, ring *Ring) error) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	eg, ctx := errgroup.WithContext(cancelCtx)
	ring := ring(client, logger, prefix, id)
	defer ring.logger.WithFields(ring.memberLogFields()).Infof("shutting down ring")
	etcdCol := collection.NewEtcdCollection(ring.client, path.Join(ring.prefix, "nodes"), nil, nil, nil, nil)
	eg.Go(func() error { return ring.watch(ctx, etcdCol) })
	eg.Go(func() error { return ring.createNode(ctx, etcdCol, ring.node.Id) })
	eg.Go(func() error {
		ring.logger.WithFields(ring.memberLogFields()).Infof("started ring")
		if err := cb(ctx, ring); err != nil {
			return err
		}
		cancel()
		return nil
	})
	err := eg.Wait()
	if errors.Is(cancelCtx.Err(), context.Canceled) {
		err = nil
	}
	return errors.EnsureStack(err)
}

func ring(client *etcd.Client, logger *logrus.Logger, prefix string, id string) *Ring {
	localMember := member{
		Id:   id,
		hash: hashFn([]byte(id)),
	}
	return &Ring{
		client:  client,
		logger:  logger,
		members: []member{localMember},
		node: node{
			member: localMember,
			locks:  map[string]lockInfo{},
		},
		prefix: prefix,
	}
}

// createNode creates a lease, inserts itself as a key to etcd, keeps the lease alive in the background.
func (ring *Ring) createNode(ctx context.Context, col collection.EtcdCollection, id string) error {
	if err := col.Claim(ctx, id, &types.BoolValue{Value: true},
		func(ctx context.Context) error {
			// keep the lease alive until the context is canceled.
			<-ctx.Done()
			return nil
		}); err != nil {
		ring.logger.WithFields(ring.memberLogFields()).Errorf("failed to renew etcd lease: %v", err)
		return errors.EnsureStack(err)
	}
	return nil
}

// watch watches etcd keys with prefix Ring.prefix to determine if a member has been added or removed.
// When a member is added, the ring calls rebalance to release any locks that no longer associate to its node. Rebalance
// on delete happens by default since each member attempts to retrieve all locks. When a member is deleted,
// the pending calls to Lock by each member will go through on the nodes that associates to given lock.
func (ring *Ring) watch(ctx context.Context, col collection.EtcdCollection) error {
	if err := col.ReadOnly(ctx).WatchF(func(event *watch.Event) error {
		ring.stateLock.Lock()
		defer ring.stateLock.Unlock()
		nodeKey := string(event.Key)
		switch event.Type {
		case watch.EventDelete:
			ring.removeById(ring.etcdToMember(nodeKey))
		case watch.EventPut:
			ring.insertById(ring.etcdToMember(nodeKey))
			// Need to release locks that no longer associate to the ring's node.
			if err := ring.rebalance(); err != nil {
				ring.logger.WithField("ring", ring.prefix).Errorf("failed rebalancing: %v", err)
				return err
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		ring.logger.WithFields(ring.memberLogFields()).Errorf("failed watch: %v", err)
		return errors.EnsureStack(err)
	}
	return nil
}

func (ring *Ring) etcdToMember(key string) string {
	return path.Base(key)
}

func (ring *Ring) keyWithRingPrefix(key string) string {
	return path.Join(ring.prefix, key)
}

// rebalance iterates through held locks and releases locks that no longer associate to a ring's node.
func (ring *Ring) rebalance() error {
	var errs error
	for key, lockInfo := range ring.node.locks {
		if ring.get(key).Id == ring.node.Id {
			continue
		}
		if err := lockInfo.lock.Unlock(lockInfo.ctx); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "error unlocking %v", key))
		}
		delete(ring.node.locks, key)
		ring.logger.WithFields(ring.lockLogFields(key)).Info("rebalanced lock")
	}
	return errors.EnsureStack(errs)
}

func (ring *Ring) get(key string) member {
	return ring.members[ring.getIndex(key)]
}

// getIndex returns the index of the member associated with key. A key belongs to the first member whose hash is
// greater than the key. If no Members are greater, then the Ring wraps around to the first member.
func (ring *Ring) getIndex(key string) int {
	keyHash := hashFn([]byte(key))
	index := sort.Search(len(ring.members), func(i int) bool {
		return ring.members[i].hash >= keyHash
	})
	// wrap around to the first member in the Ring.
	if index >= len(ring.members) {
		index = 0
	}
	return index
}

func (ring *Ring) insertById(id string) {
	member := member{
		Id:   id,
		hash: hashFn([]byte(id)),
	}
	ring.insert(member)
}

// insert updates Ring.members, and sorts Ring.members by the member hash.
// If the member already exists, insert is a no-op. The member can be mocked for testing purposes.
func (ring *Ring) insert(member member) {
	index := sort.Search(len(ring.members), func(i int) bool {
		return ring.members[i].hash >= member.hash
	})
	if index < len(ring.members) && ring.members[index].hash == member.hash {
		return // member already exists.
	}
	ring.logger.WithFields(ring.memberLogFields()).Infof("adding member %s to ring", member.Id)
	if index >= len(ring.members) {
		ring.members = append(ring.members, member)
		return
	}
	ring.members = append(ring.members[:index+1], ring.members[index:]...)
	ring.members[index] = member
}

func (ring *Ring) removeById(id string) {
	member := member{
		Id:   id,
		hash: hashFn([]byte(id)),
	}
	ring.remove(member)
}

// remove deletes the member from Ring.members if it exists.
func (ring *Ring) remove(member member) {
	index := sort.Search(len(ring.members), func(i int) bool {
		return ring.members[i].hash >= member.hash
	})
	if ring.members == nil || index >= len(ring.members) || ring.members[index].Id != member.Id {
		return // member doesn't exist.
	}
	ring.logger.WithFields(ring.memberLogFields()).Infof("deleting member %s from ring", member.Id)
	ring.members = append(ring.members[:index], ring.members[index+1:]...)
}

// Lock attempts to lock a key in etcd using consistent hashing. If key does not belong to the ring's node,
// the call blocks until context is cancelled.
func (ring *Ring) Lock(ctx context.Context, key string) (context.Context, error) {
	key = ring.keyWithRingPrefix(key)
	ring.stateLock.Lock()
	defer ring.stateLock.Unlock()
	info, exists := ring.node.locks[key]
	if exists && info.lock != nil {
		return info.ctx, nil
	}
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	// TODO: this should be event driven instead of by polling for better scalability.
	for {
		if len(ring.members) != 0 && ring.node.Id == ring.get(key).Id {
			nodeLock := dlock.NewDLock(ring.client, key)
			lockCtx, err := nodeLock.TryLock(ctx)
			if err != nil && !errors.Is(err, concurrency.ErrLocked) {
				return nil, errors.EnsureStack(err)
			}
			if err == nil { // lock() must fallthrough in the case where err == concurrency.ErrLocked
				ring.node.locks[key] = lockInfo{
					lock: nodeLock,
					ctx:  ctx,
				}
				ring.logger.WithFields(ring.lockLogFields(key)).Info("claimed lock")
				return lockCtx, nil
			}
		}
		ring.stateLock.Unlock()
		select {
		case <-ctx.Done():
			ring.stateLock.Lock()
			return nil, errors.EnsureStack(ctx.Err())
		case <-ticker.C:
			ring.stateLock.Lock()
		}
	}
}

// Unlock attempts to unlock a key in etcd if and only if a ring's node owns a lock reference to it.
// Remote member instances must give up their own locks.
func (ring *Ring) Unlock(key string) error {
	ring.stateLock.Lock()
	defer ring.stateLock.Unlock()
	key = ring.keyWithRingPrefix(key)
	info, exists := ring.node.locks[key]
	if !exists {
		return nil // lock does not exist in this local process.
	}
	if err := info.lock.Unlock(info.ctx); err != nil {
		return errors.EnsureStack(err)
	}
	delete(ring.node.locks, key)
	ring.logger.WithFields(ring.lockLogFields(key)).Info("released lock")
	return nil
}

// memberLogFields are used by a ring logger to annotate log events with the ring and node.
func (ring *Ring) memberLogFields() map[string]interface{} {
	return map[string]interface{}{
		"ring": ring.prefix,
		"node": ring.node.Id,
	}
}

// lockLogFields are used by a ring logger to annotate log events with info from memberLogFields and a lock.
func (ring *Ring) lockLogFields(lock string) map[string]interface{} {
	fields := ring.memberLogFields()
	fields["lock"] = lock
	return fields
}
