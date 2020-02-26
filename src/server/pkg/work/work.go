package work

import (
	"context"
	"fmt"
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
)

const (
	taskPrefix    = "/task"
	subtaskPrefix = "/subtask"
)

// CollectFunc is a callback that is used for collecting the results
// from a subtask that has been processed.
type CollectFunc func(context.Context, *TaskInfo) error

type Master struct {
	etcdClient          *etcd.Client
	taskCol, subtaskCol col.Collection
	taskQueue           *taskQueue
}

func NewMaster(ctx context.Context, etcdClient *etcd.Client, etcdPrefix string, taskNamespace string) *Master {
	return &Master{
		etcdClient: etcdClient,
		taskCol:    collection(etcdClient, path.Join(etcdPrefix, taskPrefix, taskNamespace)),
		subtaskCol: collection(etcdClient, path.Join(etcdPrefix, subtaskPrefix, taskNamespace)),
		taskQueue:  newTaskQueue(ctx, 100),
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

func (m *Master) CreateTask(ctx context.Context, task *Task, f func(context.Context)) error {
	// (bryce) probably should validate that subtasks are not set here.
	// or change up the protos.
	if _, err := col.NewSTM(ctx, m.etcdClient, func(stm col.STM) error {
		return m.taskCol.ReadWrite(stm).Put(task.ID, &TaskInfo{Task: task})
	}); err != nil {
		return err
	}
	m.taskQueue.createTask(ctx, task.ID, f)
	return nil
}

func (m *Master) SendSubtaskBatch(ctx context.Context, taskID string, subtasks *Task, collectFunc CollectFunc) error {
	taskInfo := &TaskInfo{}
	if _, err := col.NewSTM(ctx, m.etcdClient, func(stm col.STM) error {
		return m.taskCol.ReadWrite(stm).Update(taskID, taskInfo, func() error {
			taskInfo.Task.Subtasks = append(taskInfo.Task.Subtasks, subtasks)
			return nil
		})
	}); err != nil {
		return err
	}
	m.taskQueue.sendSubtaskBatch(taskID, subtasks, func(ctx context.Context, subtask *Task) error {
		// Collect the results from the processing of the subtasks.
		subtaskKey := path.Join(taskID, subtask.ID)
		return m.subtaskCol.ReadOnly(ctx).WatchOneF(subtaskKey, func(e *watch.Event) error {
			var key string
			subtaskInfo := &TaskInfo{}
			if err := e.Unmarshal(&key, subtaskInfo); err != nil {
				return err
			}
			// Check that the full key matched the subtask key.
			if key != subtaskKey {
				return nil
			}
			// Check that the subtask state is terminal.
			if subtaskInfo.State == State_RUNNING {
				return nil
			}
			if collectFunc != nil {
				if err := collectFunc(ctx, subtaskInfo); err != nil {
					return err
				}
			}
			return errutil.ErrBreak
		}, watch.WithFilterDelete())
	})
	return nil
}

func (m *Master) WaitTask(ctx context.Context, taskID string) (retErr error) {
	return m.taskQueue.waitTask(taskID)
}

func (m *Master) DeleteTask(ctx context.Context, taskID string) (retErr error) {
	defer func() {
		if _, err := col.NewSTM(ctx, m.etcdClient, func(stm col.STM) error {
			return m.taskCol.ReadWrite(stm).Delete(taskID)
		}); retErr == nil {
			retErr = err
		}
	}()
	return m.taskQueue.deleteTask(taskID)
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
	taskQueue           *taskQueue
}

// NewWorker creates a new worker.
func NewWorker(ctx context.Context, etcdClient *etcd.Client, etcdPrefix string, taskNamespace string) *Worker {
	return &Worker{
		etcdClient: etcdClient,
		taskCol:    collection(etcdClient, path.Join(etcdPrefix, taskPrefix, taskNamespace)),
		subtaskCol: collection(etcdClient, path.Join(etcdPrefix, subtaskPrefix, taskNamespace)),
		taskQueue:  newTaskQueue(ctx, 100),
	}
}

// Run runs the worker with the given context.
// The worker will continue to watch the task collection until the context is cancelled.
func (w *Worker) Run(ctx context.Context, processFunc ProcessFunc) error {
	return w.taskCol.ReadOnly(ctx).WatchF(func(e *watch.Event) (retErr error) {
		var taskID string
		taskInfo := &TaskInfo{}
		if err := e.Unmarshal(&taskID, taskInfo); err != nil {
			return err
		}
		if e.Type == watch.EventDelete {
			defer func() {
				// Attempt to cleanup the subtask entries.
				//if _, err := col.NewSTM(ctx, w.etcdClient, func(stm col.STM) error {
				//	w.subtaskCol.ReadWrite(stm).DeleteAllPrefix(taskID)
				//	return nil
				//}); retErr == nil {
				//	retErr = err
				//}
			}()
			return w.taskQueue.deleteTask(taskID)
		}
		// Create the task if it has not already been created.
		w.taskQueue.createTask(ctx, taskID, func(ctx context.Context) {
			if err := w.subtaskCol.ReadOnly(ctx).WatchOneF(taskID, func(e *watch.Event) error {
				if e.Type == watch.EventDelete {
					var subtaskKey string
					subtaskInfo := &TaskInfo{}
					if err := e.Unmarshal(&subtaskKey, subtaskInfo); err != nil {
						return err
					}
					w.taskQueue.sendSubtaskBatch(taskID, &Task{Subtasks: []*Task{subtaskInfo.Task}}, w.taskFunc(taskInfo.Task, processFunc))
				}
				return nil
			}, watch.WithFilterPut()); err != nil && ctx.Err() != context.Canceled {
				// (bryce) how should logging work?
				fmt.Printf("errored in task callback")
			}
		})
		// Send the subtask batch to the task queue.
		if len(taskInfo.Task.Subtasks) > 0 {
			nextSubtasks := taskInfo.Task.Subtasks[len(taskInfo.Task.Subtasks)-1]
			w.taskQueue.sendSubtaskBatch(taskID, nextSubtasks, w.taskFunc(taskInfo.Task, processFunc))
		}
		return nil
	})
}

func (w *Worker) taskFunc(task *Task, processFunc ProcessFunc) taskFunc {
	return func(ctx context.Context, subtask *Task) error {
		subtaskKey := path.Join(task.ID, subtask.ID)
		subtaskInfo := &TaskInfo{}
		if err := w.subtaskCol.Claim(ctx, subtaskKey, subtaskInfo, func(claimCtx context.Context) (retErr error) {
			defer func() {
				// If the task was cancelled or the claim was lost, just return with no error.
				// (bryce) figure out what logging this should look like.
				if claimCtx.Err() == context.Canceled {
					retErr = nil
					return
				}
				subtaskInfo.Task = subtask
				subtaskInfo.State = State_SUCCESS
				if retErr != nil {
					subtaskInfo.State = State_FAILURE
					subtaskInfo.Reason = retErr.Error()
					retErr = nil
				}
				if _, err := col.NewSTM(claimCtx, w.etcdClient, func(stm col.STM) error {
					return w.subtaskCol.ReadWrite(stm).Put(subtaskKey, subtaskInfo)
				}); retErr == nil {
					retErr = err
				}
			}()
			return processFunc(claimCtx, task, subtask)
		}); err != nil && err != col.ErrNotClaimed {
			return err
		}
		return nil
	}
}
