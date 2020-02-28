package work

import (
	"context"
	"fmt"
	"path"
	"sync/atomic"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	"golang.org/x/sync/errgroup"
)

const (
	taskPrefix    = "/task"
	subtaskPrefix = "/subtask"
	claimPrefix   = "/claim"
)

type TaskQueue struct {
	*taskEtcd
	taskQueue *taskQueue
}

type taskEtcd struct {
	etcdClient                    *etcd.Client
	taskCol, subtaskCol, claimCol col.Collection
}

func NewTaskQueue(ctx context.Context, etcdClient *etcd.Client, etcdPrefix string, taskNamespace string) *TaskQueue {
	return &TaskQueue{
		taskEtcd:  newTaskEtcd(etcdClient, etcdPrefix, taskNamespace),
		taskQueue: newTaskQueue(ctx),
	}
}

func newTaskEtcd(etcdClient *etcd.Client, etcdPrefix string, taskNamespace string) *taskEtcd {
	return &taskEtcd{
		etcdClient: etcdClient,
		taskCol:    newCollection(etcdClient, path.Join(etcdPrefix, taskPrefix, taskNamespace), &Task{}),
		subtaskCol: newCollection(etcdClient, path.Join(etcdPrefix, subtaskPrefix, taskNamespace), &TaskInfo{}),
		claimCol:   newCollection(etcdClient, path.Join(etcdPrefix, claimPrefix, taskNamespace), &Claim{}),
	}
}

func newCollection(etcdClient *etcd.Client, etcdPrefix string, template proto.Message) col.Collection {
	return col.NewCollection(
		etcdClient,
		etcdPrefix,
		nil,
		template,
		nil,
		nil,
	)
}

func (tq *TaskQueue) CreateTask(ctx context.Context, task *Task) (*Master, error) {
	if _, err := col.NewSTM(ctx, tq.etcdClient, func(stm col.STM) error {
		return tq.taskCol.ReadWrite(stm).Put(task.ID, task)
	}); err != nil {
		return nil, err
	}
	taskChans, err := tq.taskQueue.createTask(ctx, task.ID)
	if err != nil {
		return nil, err
	}
	return &Master{
		taskEtcd:  tq.taskEtcd,
		taskID:    task.ID,
		taskChans: taskChans,
		done:      make(chan struct{}),
	}, nil
}

type Master struct {
	*taskEtcd
	taskID    string
	taskChans *taskChans
	eg        errgroup.Group
	count     int64
	done      chan struct{}
}

// CollectFunc is a callback that is used for collecting the results
// from a subtask that has been processed.
type CollectFunc func(context.Context, *TaskInfo) error

func (m *Master) Collect(collectFunc CollectFunc) {
	m.eg.Go(func() error {
		return m.subtaskCol.ReadOnly(m.taskChans.ctx).WatchOneF(m.taskID, func(e *watch.Event) error {
			var key string
			subtaskInfo := &TaskInfo{}
			if err := e.Unmarshal(&key, subtaskInfo); err != nil {
				return err
			}
			// Check that the subtask state is terminal.
			if subtaskInfo.State == State_RUNNING {
				return nil
			}
			if collectFunc != nil {
				if err := m.taskChans.processSubtask(func(ctx context.Context) error {
					return collectFunc(ctx, subtaskInfo)
				}); err != nil {
					return err
				}
			}
			atomic.AddInt64(&m.count, -1)
			select {
			case <-m.done:
				// (bryce) it might be worth the memory cost to just keep a map of ids for subtasks that were created,
				// rather than relying on the correct etcd events?
				if m.count == 0 {
					return errutil.ErrBreak
				}
			default:
			}
			return nil
		})
	})
}

func (m *Master) CreateSubtask(subtask *Task) error {
	subtaskKey := path.Join(m.taskID, subtask.ID)
	subtaskInfo := &TaskInfo{Task: subtask}
	if _, err := col.NewSTM(m.taskChans.ctx, m.etcdClient, func(stm col.STM) error {
		return m.subtaskCol.ReadWrite(stm).Put(subtaskKey, subtaskInfo)
	}); err != nil {
		return err
	}
	atomic.AddInt64(&m.count, 1)
	return nil
}

func (m *Master) Wait() error {
	close(m.done)
	return m.eg.Wait()
}

func (tq *TaskQueue) DeleteTask(ctx context.Context, taskID string) error {
	if err := tq.taskQueue.deleteTask(taskID); err != nil {
		return err
	}
	// Cleanup task entry.
	if _, err := col.NewSTM(ctx, tq.etcdClient, func(stm col.STM) error {
		return tq.taskCol.ReadWrite(stm).Delete(taskID)
	}); err != nil {
		return err
	}
	// Cleanup subtask entries.
	_, err := col.NewSTM(ctx, tq.etcdClient, func(stm col.STM) error {
		tq.subtaskCol.ReadWrite(stm).DeleteAllPrefix(taskID)
		return nil
	})
	return err
}

// Worker is a worker that will process subtasks in a task.
// The worker will watch the task collection for tasks added by a master.
// When a task is added, the worker will claim and process subtasks associated
// with that task.
// The processFunc callback will be called for each subtask that needs to be processed
// in the task.
type Worker struct {
	*taskEtcd
	taskQueue *taskQueue
}

// NewWorker creates a new worker.
func NewWorker(ctx context.Context, etcdClient *etcd.Client, etcdPrefix string, taskNamespace string) *Worker {
	return &Worker{
		taskEtcd:  newTaskEtcd(etcdClient, etcdPrefix, taskNamespace),
		taskQueue: newTaskQueue(ctx),
	}
}

// ProcessFunc is a callback that is used for processing a subtask in a task.
type ProcessFunc func(context.Context, *Task, *Task) error

// Run runs the worker with the given context.
// The worker will continue to watch the task collection until the context is cancelled.
func (w *Worker) Run(ctx context.Context, processFunc ProcessFunc) error {
	return w.taskCol.ReadOnly(ctx).WatchF(func(e *watch.Event) (retErr error) {
		var taskID string
		task := &Task{}
		if err := e.Unmarshal(&taskID, task); err != nil {
			return err
		}
		if e.Type == watch.EventDelete {
			return w.taskQueue.deleteTask(taskID)
		}
		taskChans, err := w.taskQueue.createTask(ctx, taskID)
		if err != nil {
			return err
		}
		go func() {
			if err := w.taskFunc(task, taskChans, processFunc); err != nil && taskChans.ctx.Err() != context.Canceled {
				// (bryce) how should logging work?
				fmt.Printf("errored in task callback: %v\n", err)
			}
		}()
		return nil
	})
}

func (w *Worker) taskFunc(task *Task, taskChans *taskChans, processFunc ProcessFunc) error {
	claimWatch, err := w.claimCol.ReadOnly(taskChans.ctx).WatchOne(task.ID, watch.WithFilterPut())
	if err != nil {
		return err
	}
	subtaskWatch, err := w.subtaskCol.ReadOnly(taskChans.ctx).WatchOne(task.ID, watch.WithFilterDelete())
	if err != nil {
		return err
	}
	for {
		select {
		case e := <-claimWatch.Watch():
			if e.Type == watch.EventError {
				return err
			}
			if e.Type != watch.EventDelete {
				continue
			}
			var subtaskKey string
			if err := e.Unmarshal(&subtaskKey, &Claim{}); err != nil {
				return err
			}
			if err := taskChans.processSubtask(w.subtaskFunc(task, subtaskKey, processFunc)); err != nil {
				return err
			}
		case e := <-subtaskWatch.Watch():
			if e.Type == watch.EventError {
				return err
			}
			var subtaskKey string
			if err := e.Unmarshal(&subtaskKey, &TaskInfo{}); err != nil {
				return err
			}
			if err := taskChans.processSubtask(w.subtaskFunc(task, subtaskKey, processFunc)); err != nil {
				return err
			}
		case <-taskChans.ctx.Done():
			return taskChans.ctx.Err()
		}
	}
}

func (w *Worker) subtaskFunc(task *Task, subtaskKey string, processFunc ProcessFunc) subtaskFunc {
	return func(ctx context.Context) error {
		// (bryce) this should be refactored to have the check and claim in the same stm.
		// there is a rare race condition that does not affect correctness, but it is less
		// than ideal because a subtask could get run once more than necessary.
		subtaskInfo := &TaskInfo{}
		if _, err := col.NewSTM(ctx, w.etcdClient, func(stm col.STM) error {
			return w.subtaskCol.ReadWrite(stm).Get(subtaskKey, subtaskInfo)
		}); err != nil {
			return err
		}
		if subtaskInfo.State == State_RUNNING {
			if err := w.claimCol.Claim(ctx, subtaskKey, &Claim{}, func(claimCtx context.Context) (retErr error) {
				defer func() {
					// If the task was cancelled or the claim was lost, just return with no error.
					if claimCtx.Err() == context.Canceled {
						retErr = nil
						return
					}
					subtaskInfo := &TaskInfo{}
					if _, err := col.NewSTM(claimCtx, w.etcdClient, func(stm col.STM) error {
						return w.subtaskCol.ReadWrite(stm).Update(subtaskKey, subtaskInfo, func() error {
							// (bryce) remove when check and claim are in the same stm.
							if subtaskInfo.State != State_RUNNING {
								return nil
							}
							subtaskInfo.State = State_SUCCESS
							if retErr != nil {
								subtaskInfo.State = State_FAILURE
								subtaskInfo.Reason = retErr.Error()
								retErr = nil
							}
							return nil
						})
					}); retErr == nil {
						retErr = err
					}
				}()
				return processFunc(claimCtx, task, subtaskInfo.Task)
			}); err != nil && err != col.ErrNotClaimed {
				return err
			}
		}
		return nil
	}
}
