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
	taskPrefix    = "/task"
	subtaskPrefix = "/subtask"
)

// CollectFunc is a callback that is used for collecting the results
// from a subtask that has been processed.
type CollectFunc func(context.Context, *Task) error

// Master is the master for a task.
// The master will layout the subtasks for the task in etcd and collect
// them upon completion. The collectFunc callback will be called for
// each subtask that is collected.
type Master struct {
	etcdClient          *etcd.Client
	taskCol, subtaskCol col.Collection
	collectFunc         CollectFunc
}

// NewMaster creates a new master.
func NewMaster(etcdClient *etcd.Client, etcdPrefix string, collectFunc CollectFunc) *Master {
	return &Master{
		etcdClient:  etcdClient,
		taskCol:     collection(etcdClient, path.Join(etcdPrefix, taskPrefix)),
		subtaskCol:  collection(etcdClient, path.Join(etcdPrefix, subtaskPrefix)),
		collectFunc: collectFunc,
	}
}

func collection(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		etcdPrefix,
		nil,
		&TaskInfo{},
		nil,
		nil,
	)
}

// Run runs the master with a given context and task.
func (m *Master) Run(ctx context.Context, task *Task) (retErr error) {
	taskInfo := &TaskInfo{Task: task}
	if err := m.updateTaskInfo(ctx, taskInfo); err != nil {
		return err
	}
	defer func() {
		taskInfo.State = State_SUCCESS
		// (bryce) might also want a killed state.
		if retErr != nil {
			taskInfo.State = State_FAILURE
		}
		if err := m.updateTaskInfo(ctx, taskInfo); err != nil && retErr != nil {
			retErr = err
		}
		// (bryce) what should the work etcd entry cleanup look like?
		// Need all of the workers to receive the termination signal
		// before it can be cleaned up.
	}()
	// Collect the results from the processing of the subtasks.
	for _, subtask := range task.Subtasks {
		subtaskKey := path.Join(task.Id, subtask.Id)
		if err := m.subtaskCol.ReadOnly(ctx).WatchOneF(subtaskKey, func(e *watch.Event) error {
			var key string
			subtaskInfo := &TaskInfo{}
			if err := e.Unmarshal(&key, subtaskInfo); err != nil {
				return err
			}
			// Check that the subtask state is terminal.
			if subtaskInfo.State == State_RUNNING {
				return nil
			}
			// Check that the full key matched the subtask key.
			if key != subtaskKey {
				return nil
			}
			// (bryce) need to figure out error propagation if subtask fails.
			// if subtaskInfo.State == State_FAILURE {}
			if err := m.collectFunc(ctx, subtaskInfo.Task); err != nil {
				return err
			}
			return errutil.ErrBreak
		}, watch.WithFilterDelete()); err != nil {
			return err
		}
	}
	return nil
}

func (m *Master) updateTaskInfo(ctx context.Context, taskInfo *TaskInfo) error {
	_, err := col.NewSTM(ctx, m.etcdClient, func(stm col.STM) error {
		return m.taskCol.ReadWrite(stm).Put(taskInfo.Task.Id, taskInfo)
	})
	return err
}

// ProcessFunc is a callback that is used for processing a subtask in a task.
type ProcessFunc func(context.Context, *Task, *Task) error

// Worker is a worker that will process subtasks in a task.
// The worker will watch the task collection for tasks added by a master.
// When a task is added, the worker will claim and process subtasks associated
// with that task.
// The processFunc callback will be called for each subtask that needs to be processed
// in the task.
type Worker struct {
	etcdClient          *etcd.Client
	taskCol, subtaskCol col.Collection
	processFunc         ProcessFunc
}

// NewWorker creates a new worker.
func NewWorker(etcdClient *etcd.Client, etcdPrefix string, processFunc ProcessFunc) *Worker {
	return &Worker{
		etcdClient:  etcdClient,
		taskCol:     collection(etcdClient, path.Join(etcdPrefix, taskPrefix)),
		subtaskCol:  collection(etcdClient, path.Join(etcdPrefix, subtaskPrefix)),
		processFunc: processFunc,
	}
}

// Run runs the worker with the given context.
// The worker will continue to watch the task collection until the context is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	return w.taskCol.ReadOnly(ctx).WatchF(func(e *watch.Event) (retErr error) {
		var id string
		taskInfo := &TaskInfo{}
		if err := e.Unmarshal(&id, taskInfo); err != nil {
			return err
		}
		if taskInfo.State == State_RUNNING {
			return w.processTask(ctx, taskInfo.Task)
		}
		return nil
	})
}

func (w *Worker) processTask(ctx context.Context, task *Task) error {
	var eg errgroup.Group
	ctx, cancel := context.WithCancel(ctx)
	// Watch for task termination.
	// This will handle completion, failure, and cancellation.
	eg.Go(func() error {
		return w.taskCol.ReadOnly(ctx).WatchOneF(task.Id, func(e *watch.Event) error {
			var id string
			taskInfo := &TaskInfo{}
			if err := e.Unmarshal(&id, taskInfo); err != nil {
				return err
			}
			if taskInfo.State != State_RUNNING {
				cancel()
				return errutil.ErrBreak
			}
			return nil
		})
	})
	// Process the subtasks.
	eg.Go(func() (retErr error) {
		// (bryce) cleanup of subtasks needs to be re-thought.
		// Cannot cleanup in this manner, because it messes up subtasks claims when a worker errors.
		// Cleaning up in the master results with a race condition with workers
		// waiting on work termination signal and subtask claims to expire.
		//defer func() {
		//	// Attempt to cleanup subtask entries.
		//	if _, err := col.NewSTM(context.Background(), w.etcdClient, func(stm col.STM) error {
		//		w.subtaskCol.ReadWrite(stm).DeleteAllPrefix(work.Id)
		//		return nil
		//	}); err != nil && retErr == nil {
		//		retErr = err
		//	}
		//}()
		if err := w.processSubtasks(ctx, task); err != nil {
			return err
		}
		// Wait for a deletion event (ttl expired) before attempting to process the subtasks again.
		return w.subtaskCol.ReadOnly(ctx).WatchOneF(task.Id, func(e *watch.Event) error {
			return w.processSubtasks(ctx, task)
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

func (w *Worker) processSubtasks(ctx context.Context, task *Task) error {
	for _, subtask := range task.Subtasks {
		subtaskKey := path.Join(task.Id, subtask.Id)
		subtaskInfo := &TaskInfo{}
		if err := w.subtaskCol.Claim(ctx, subtaskKey, subtaskInfo, func(ctx context.Context) error {
			if err := w.processFunc(ctx, task, subtask); err != nil {
				return err
			}
			subtaskInfo.Task = subtask
			subtaskInfo.State = State_SUCCESS
			_, err := col.NewSTM(ctx, w.etcdClient, func(stm col.STM) error {
				return w.subtaskCol.ReadWrite(stm).Put(subtaskKey, subtaskInfo)
			})
			return err
		}); err != nil && err != col.ErrNotClaimed {
			return err
		}
	}
	return nil
}
