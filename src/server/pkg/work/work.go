package work

import (
	"context"
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	"golang.org/x/sync/errgroup"
)

const (
	workPrefix  = "/work"
	shardPrefix = "/shard"
)

// CollectFunc is a callback that is used for collecting the results
// from a shard that has been processed.
type CollectFunc func(context.Context, *Shard) error

// Master is the master for a unit of work.
// The master will layout the shards for the work in etcd and collect
// them upon completion. The collectFunc callback will be called for
// each shard that is collected.
type Master struct {
	etcdClient        *etcd.Client
	workCol, shardCol col.Collection
	collectFunc       CollectFunc
}

// NewMaster creates a new master.
func NewMaster(etcdClient *etcd.Client, etcdPrefix string, collectFunc CollectFunc) *Master {
	return &Master{
		etcdClient:  etcdClient,
		workCol:     workCollection(etcdClient, etcdPrefix),
		shardCol:    shardCollection(etcdClient, etcdPrefix),
		collectFunc: collectFunc,
	}
}

func workCollection(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, workPrefix),
		nil,
		&WorkInfo{},
		nil,
		nil,
	)
}

func shardCollection(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, shardPrefix),
		nil,
		&ShardInfo{},
		nil,
		nil,
	)
}

// Run runs the master with a given context and unit of work.
func (m *Master) Run(ctx context.Context, work *Work) (retErr error) {
	workInfo := &WorkInfo{Work: work}
	if err := m.updateWorkInfo(ctx, workInfo); err != nil {
		return err
	}
	defer func() {
		workInfo.State = State_SUCCESS
		// (bryce) might also want a killed state.
		if retErr != nil {
			workInfo.State = State_FAILURE
		}
		if err := m.updateWorkInfo(ctx, workInfo); err != nil && retErr != nil {
			retErr = err
		}
		// (bryce) what should the work etcd entry cleanup look like?
		// Need all of the workers to receive the termination signal
		// before it can be cleaned up.
	}()
	// Collect the results from the processing of the shards.
	for _, shard := range work.Shards {
		shardKey := path.Join(work.Id, shard.Id)
		if err := m.shardCol.ReadOnly(ctx).WatchOneF(shardKey, func(e *watch.Event) error {
			var key string
			shardInfo := &ShardInfo{}
			if err := e.Unmarshal(&key, shardInfo); err != nil {
				return err
			}
			// Check that the shard state is terminal.
			if shardInfo.State == State_RUNNING {
				return nil
			}
			// Check that the full key matched the shard key.
			if key != shardKey {
				return nil
			}
			// (bryce) need to figure out error propagation if shard fails.
			// if shardInfo.State == State_FAILURE {}
			if err := m.collectFunc(ctx, shardInfo.Shard); err != nil {
				return err
			}
			return errutil.ErrBreak
		}, watch.WithFilterDelete()); err != nil {
			return err
		}
	}
	return nil
}

func (m *Master) updateWorkInfo(ctx context.Context, workInfo *WorkInfo) error {
	_, err := col.NewSTM(ctx, m.etcdClient, func(stm col.STM) error {
		return m.workCol.ReadWrite(stm).Put(workInfo.Work.Id, workInfo)
	})
	return err
}

// ProcessFunc is a callback that is used for processing a shard in a unit of work.
type ProcessFunc func(context.Context, *Work, *Shard) error

// Worker is a worker that will process shards in a unit of work.
// The worker will watch the work collection for units of work added by a master.
// When a unit of work is added, the worker will claim and process shards associated
// with that unit of work.
// The processFunc callback will be called for each shard that needs to be processed
// in the unit of work.
type Worker struct {
	etcdClient        *etcd.Client
	workCol, shardCol col.Collection
	processFunc       ProcessFunc
}

// NewWorker creates a new worker.
func NewWorker(etcdClient *etcd.Client, etcdPrefix string, processFunc ProcessFunc) *Worker {
	return &Worker{
		etcdClient:  etcdClient,
		workCol:     workCollection(etcdClient, etcdPrefix),
		shardCol:    shardCollection(etcdClient, etcdPrefix),
		processFunc: processFunc,
	}
}

// Run runs the worker with the given context.
// The worker will continue to watch the work collection until the context is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	return w.workCol.ReadOnly(ctx).WatchF(func(e *watch.Event) (retErr error) {
		var id string
		workInfo := &WorkInfo{}
		if err := e.Unmarshal(&id, workInfo); err != nil {
			return err
		}
		if workInfo.State == State_RUNNING {
			return w.processWork(ctx, workInfo.Work)
		}
		return nil
	})
}

func (w *Worker) processWork(ctx context.Context, work *Work) error {
	var eg errgroup.Group
	ctx, cancel := context.WithCancel(ctx)
	// Watch for work termination.
	// This will handle completion, failure, and cancellation.
	eg.Go(func() error {
		return w.workCol.ReadOnly(ctx).WatchOneF(work.Id, func(e *watch.Event) error {
			var id string
			workInfo := &WorkInfo{}
			if err := e.Unmarshal(&id, workInfo); err != nil {
				return err
			}
			if workInfo.State != State_RUNNING {
				cancel()
				return errutil.ErrBreak
			}
			return nil
		})
	})
	// Process the shards.
	eg.Go(func() (retErr error) {
		// (bryce) cleanup of shards needs to be re-thought.
		// Cannot cleanup in this manner, because it messes up shard claims when a worker errors.
		// Cleaning up in the master results with a race condition with workers
		// waiting on work termination signal and shard claims to expire.
		//defer func() {
		//	// Attempt to cleanup shard entries.
		//	if _, err := col.NewSTM(context.Background(), w.etcdClient, func(stm col.STM) error {
		//		w.shardCol.ReadWrite(stm).DeleteAllPrefix(work.Id)
		//		return nil
		//	}); err != nil && retErr == nil {
		//		retErr = err
		//	}
		//}()
		if err := w.processShards(ctx, work); err != nil {
			return err
		}
		// Wait for a deletion event (ttl expired) before attempting to process the shards again.
		return w.shardCol.ReadOnly(ctx).WatchOneF(work.Id, func(e *watch.Event) error {
			return w.processShards(ctx, work)
		}, watch.WithFilterPut())
	})
	if err := eg.Wait(); err != nil {
		// Work terminated.
		if ctx.Err() == context.Canceled {
			return nil
		}
		return err
	}
	return nil
}

func (w *Worker) processShards(ctx context.Context, work *Work) error {
	for _, shard := range work.Shards {
		shardKey := path.Join(work.Id, shard.Id)
		shardInfo := &ShardInfo{}
		if err := w.shardCol.Claim(ctx, shardKey, shardInfo, func(ctx context.Context) error {
			if err := w.processFunc(ctx, work, shard); err != nil {
				return err
			}
			shardInfo.Shard = shard
			shardInfo.State = State_SUCCESS
			_, err := col.NewSTM(ctx, w.etcdClient, func(stm col.STM) error {
				return w.shardCol.ReadWrite(stm).Put(shardKey, shardInfo)
			})
			return err
		}); err != nil && err != col.ErrNotClaimed {
			return err
		}
	}
	return nil
}
