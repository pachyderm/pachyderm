package consistenthashing

import (
	"context"
	"hash/crc32"
	"path"
	"sort"
	"sync"
	"time"

	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
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
func WithRing(ctx context.Context, client *etcd.Client, prefix string, cb func(ctx context.Context, ring *Ring) error) error {
	return withRing(ctx, client, prefix, uuid.New(), cb)
}

func withRing(rctx context.Context, client *etcd.Client, prefix, id string, cb func(ctx context.Context, ring *Ring) error) error {
	ring := ring(client, prefix, id)

	cancelCtx, cancel := pctx.WithCancel(pctx.Child(rctx, "ring", pctx.WithFields(zap.Inline(ring))))
	defer cancel()

	eg, ctx := errgroup.WithContext(cancelCtx)
	defer log.Info(ctx, "shutting down ring")

	etcdCol := collection.NewEtcdCollection(ring.client, path.Join(ring.prefix, "nodes"), nil, nil, nil, nil)
	eg.Go(func() error { return ring.watch(ctx, etcdCol) })
	eg.Go(func() error { return ring.createNode(ctx, etcdCol, ring.node.Id) })
	eg.Go(func() error {
		log.Info(ctx, "started ring")
		if err := cb(ctx, ring); err != nil {
			return err
		}
		cancel()
		return nil
	})
	err := eg.Wait()
	if errors.Is(context.Cause(cancelCtx), context.Canceled) {
		err = nil
	}
	return errors.EnsureStack(err)
}

func ring(client *etcd.Client, prefix string, id string) *Ring {
	localMember := member{
		Id:   id,
		hash: hashFn([]byte(id)),
	}
	return &Ring{
		client:  client,
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
	if err := col.Claim(ctx, id, wrapperspb.Bool(true),
		func(ctx context.Context) error {
			// keep the lease alive until the context is canceled.
			<-ctx.Done()
			return nil
		}); err != nil {
		log.Info(ctx, "failed to renew etcd lease", zap.Error(err))
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
			ring.removeById(ctx, ring.etcdToMember(nodeKey))
		case watch.EventPut:
			ring.insertById(ctx, ring.etcdToMember(nodeKey))
			// Need to release locks that no longer associate to the ring's node.
			if err := ring.rebalance(); err != nil {
				log.Error(ctx, "failed rebalancing", zap.Error(err))
				return err
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		log.Error(ctx, "failed watch", zap.Error(err))
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
	for key := range ring.node.locks {
		if ring.get(key).Id == ring.node.Id {
			continue
		}
		if err := ring.releaseLock(key); err != nil {
			return err
		}
	}
	return nil
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

func (ring *Ring) releaseLock(key string) error {
	lockInfo, exists := ring.node.locks[key]
	if !exists {
		return nil
	}
	delete(ring.node.locks, key)
	// The lock is released when either:
	// - The lock context is canceled.
	// - The unlock call completes successfully.
	ctx := lockInfo.ctx
	return backoff.RetryNotify(func() error {
		if errors.Is(context.Cause(ctx), context.Canceled) {
			return nil
		}
		return lockInfo.lock.Unlock(ctx)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "releasing lock; retrying", zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	})
}

func (ring *Ring) insertById(ctx context.Context, id string) {
	member := member{
		Id:   id,
		hash: hashFn([]byte(id)),
	}
	ring.insert(ctx, member)
}

// insert updates Ring.members, and sorts Ring.members by the member hash.
// If the member already exists, insert is a no-op. The member can be mocked for testing purposes.
func (ring *Ring) insert(ctx context.Context, member member) {
	index := sort.Search(len(ring.members), func(i int) bool {
		return ring.members[i].hash >= member.hash
	})
	if index < len(ring.members) && ring.members[index].hash == member.hash {
		return // member already exists.
	}
	log.Info(ctx, "adding member to ring", zap.String("memberID", member.Id))
	if index >= len(ring.members) {
		ring.members = append(ring.members, member)
		return
	}
	ring.members = append(ring.members[:index+1], ring.members[index:]...)
	ring.members[index] = member
}

func (ring *Ring) removeById(ctx context.Context, id string) {
	member := member{
		Id:   id,
		hash: hashFn([]byte(id)),
	}
	ring.remove(ctx, member)
}

// remove deletes the member from Ring.members if it exists.
func (ring *Ring) remove(ctx context.Context, member member) {
	index := sort.Search(len(ring.members), func(i int) bool {
		return ring.members[i].hash >= member.hash
	})
	if ring.members == nil || index >= len(ring.members) || ring.members[index].Id != member.Id {
		return // member doesn't exist.
	}
	log.Info(ctx, "deleting member from ring", zap.String("memberID", member.Id))
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
				l := lockInfo{
					lock: nodeLock,
					ctx:  pctx.Child(lockCtx, "lock", pctx.WithFields(zap.String("lock", key))),
				}
				ring.node.locks[key] = l
				return lockCtx, nil
			}
		}
		ring.stateLock.Unlock()
		select {
		case <-ctx.Done():
			ring.stateLock.Lock()
			return nil, errors.EnsureStack(context.Cause(ctx))
		case <-ticker.C:
			ring.stateLock.Lock()
		}
	}
}

// Unlock attempts to unlock a key in etcd if and only if a ring's node owns a lock reference to it.
// Remote member instances must give up their own locks.
func (ring *Ring) Unlock(key string) error {
	key = ring.keyWithRingPrefix(key)
	ring.stateLock.Lock()
	defer ring.stateLock.Unlock()
	return ring.releaseLock(key)
}

// MarshalLogObject implements zapcore.ObjectMarshaler.
func (ring *Ring) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("ring", ring.prefix)
	enc.AddString("node", ring.node.Id)
	return nil
}
