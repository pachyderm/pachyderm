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
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	"golang.org/x/sync/errgroup"
)

const (
	taskPrefix    = "/task"
	subtaskPrefix = "/subtask"
	claimPrefix   = "/claim"
)

// TaskQueue manages a set of parallel tasks, and provides an interface for running tasks.
// Priority of tasks (and therefore subtasks) is based on task creation time, so tasks created
// earlier will be prioritized over tasks that were created later.
type TaskQueue struct {
	*taskEtcd
	taskQueue *taskQueue
}

type taskEtcd struct {
	etcdClient                    *etcd.Client
	taskCol, subtaskCol, claimCol col.Collection
}

// NewTaskQueue sets up a new task queue.
func NewTaskQueue(ctx context.Context, etcdClient *etcd.Client, etcdPrefix string, taskNamespace string) (*TaskQueue, error) {
	tq := &TaskQueue{
		taskEtcd:  newTaskEtcd(etcdClient, etcdPrefix, taskNamespace),
		taskQueue: newTaskQueue(ctx),
	}
	// Clear etcd key space.
	if err := tq.deleteAllTasks(); err != nil {
		return nil, err
	}
	return tq, nil
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

// RunTask runs a task in the task queue.
// The task code should be contained within the passed in callback.
// The callback will receive a Master, which should be used for running subtasks in the task queue.
// The task state will be cleaned up upon return of the callback.
func (tq *TaskQueue) RunTask(ctx context.Context, f func(*Master)) (retErr error) {
	task := &Task{ID: uuid.NewWithoutDashes()}
	if _, err := col.NewSTM(ctx, tq.etcdClient, func(stm col.STM) error {
		return tq.taskCol.ReadWrite(stm).Put(task.ID, task)
	}); err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			if err := tq.deleteTask(task.ID); err != nil {
				fmt.Printf("errored deleting task %v: %v\n", task.ID, err)
			}
		}
	}()
	return tq.taskQueue.runTask(ctx, task.ID, func(te *taskEntry) {
		defer func() {
			if err := tq.deleteTask(task.ID); err != nil {
				fmt.Printf("errored deleting task %v: %v\n", task.ID, err)
			}
		}()
		f(&Master{
			taskEtcd:  tq.taskEtcd,
			taskID:    task.ID,
			taskEntry: te,
		})
	})
}

// RunTaskBlock is similar to RunTask, but blocks on the callback.
func (tq *TaskQueue) RunTaskBlock(ctx context.Context, f func(*Master) error) error {
	errChan := make(chan error)
	if err := tq.RunTask(ctx, func(master *Master) {
		errChan <- f(master)
	}); err != nil {
		return err
	}
	return <-errChan
}

// Master manages subtasks in the task queue, and provides an interface for running subtasks.
type Master struct {
	*taskEtcd
	taskID    string
	taskEntry *taskEntry
}

// Ctx returns the context for the master.
func (m *Master) Ctx() context.Context {
	return m.taskEntry.ctx
}

// CollectFunc is a callback that is used for collecting the results
// from a subtask that has been processed.
type CollectFunc func(context.Context, *TaskInfo) error

// RunSubtasks runs a set of subtasks and collects the results with the passed in callback.
func (m *Master) RunSubtasks(subtasks []*Task, collectFunc CollectFunc) (retErr error) {
	var eg errgroup.Group
	subtaskChan := make(chan *Task)
	eg.Go(func() error {
		return m.RunSubtasksChan(subtaskChan, collectFunc)
	})
	defer func() {
		close(subtaskChan)
		if err := eg.Wait(); retErr == nil {
			retErr = err
		}
	}()
	for _, subtask := range subtasks {
		select {
		case subtaskChan <- subtask:
		case <-m.taskEntry.ctx.Done():
			return m.taskEntry.ctx.Err()
		}
	}
	return nil
}

// RunSubtasksChan runs a set of subtasks (provided through a channel) and collects the results with the passed in callback.
func (m *Master) RunSubtasksChan(subtaskChan chan *Task, collectFunc CollectFunc) (retErr error) {
	var eg errgroup.Group
	var count int64
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(m.taskEntry.ctx)
	eg.Go(func() error {
		return m.subtaskCol.ReadOnly(ctx).WatchOneF(m.taskID, func(e *watch.Event) error {
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
				if err := m.taskEntry.runSubtaskBlock(func(ctx context.Context) error {
					return collectFunc(ctx, subtaskInfo)
				}); err != nil {
					return err
				}
			}
			atomic.AddInt64(&count, -1)
			select {
			case <-done:
				if count == 0 {
					return errutil.ErrBreak
				}
			default:
			}
			return nil
		})
	})
	defer func() {
		// If the subtasks have already been collected (or there were none), then
		// cancel the ctx for the collect goroutine.
		if atomic.LoadInt64(&count) == 0 {
			cancel()
		}
		close(done)
		if err := eg.Wait(); retErr == nil && ctx.Err() != context.Canceled {
			retErr = err
		}
		if err := m.deleteSubtasks(); err != nil {
			fmt.Printf("errored deleting subtasks for task %v: %v\n", m.taskID, err)
		}
	}()
	for subtask := range subtaskChan {
		if err := m.createSubtask(subtask); err != nil {
			return err
		}
		atomic.AddInt64(&count, 1)
	}
	return nil
}

func (m *Master) createSubtask(subtask *Task) error {
	if subtask.ID == "" {
		subtask.ID = uuid.NewWithoutDashes()
	}
	subtaskKey := path.Join(m.taskID, subtask.ID)
	subtaskInfo := &TaskInfo{Task: subtask}
	if _, err := col.NewSTM(m.taskEntry.ctx, m.etcdClient, func(stm col.STM) error {
		return m.subtaskCol.ReadWrite(stm).Put(subtaskKey, subtaskInfo)
	}); err != nil {
		return err
	}
	return nil
}

func (m *Master) deleteSubtasks() error {
	_, err := col.NewSTM(context.Background(), m.etcdClient, func(stm col.STM) error {
		m.subtaskCol.ReadWrite(stm).DeleteAllPrefix(m.taskID)
		return nil
	})
	return err
}

func (tq *TaskQueue) deleteTask(taskID string) error {
	_, err := col.NewSTM(context.Background(), tq.etcdClient, func(stm col.STM) error {
		tq.subtaskCol.ReadWrite(stm).DeleteAllPrefix(taskID)
		return tq.taskCol.ReadWrite(stm).Delete(taskID)
	})
	return err
}

func (tq *TaskQueue) deleteAllTasks() error {
	_, err := col.NewSTM(context.Background(), tq.etcdClient, func(stm col.STM) error {
		tq.subtaskCol.ReadWrite(stm).DeleteAll()
		tq.taskCol.ReadWrite(stm).DeleteAll()
		return nil
	})
	return err
}

// Worker is a worker that will process subtasks in a task.
// A worker watches the task collection for tasks to be created / deleted and appropriately
// runs / deletes tasks in the internal task queue with a function that watches the
// subtask and claim collections for subtasks that need to be processed.
// The processFunc callback will be called for each subtask that needs to be processed
// in the task.
type Worker struct {
	*taskEtcd
}

// NewWorker creates a new worker.
func NewWorker(etcdClient *etcd.Client, etcdPrefix string, taskNamespace string) *Worker {
	return &Worker{taskEtcd: newTaskEtcd(etcdClient, etcdPrefix, taskNamespace)}
}

// ProcessFunc is a callback that is used for processing a subtask in a task.
type ProcessFunc func(context.Context, *Task) error

// Run runs the worker with the given context.
// The worker will continue to watch the task collection until the context is canceled.
func (w *Worker) Run(ctx context.Context, processFunc ProcessFunc) error {
	taskQueue := newTaskQueue(ctx)
	return w.taskCol.ReadOnly(ctx).WatchF(func(e *watch.Event) error {
		var taskID string
		task := &Task{}
		if err := e.Unmarshal(&taskID, task); err != nil {
			return err
		}
		if e.Type == watch.EventDelete {
			taskQueue.deleteTask(taskID)
			return nil
		}
		return taskQueue.runTask(ctx, taskID, func(taskEntry *taskEntry) {
			if err := w.taskFunc(task, taskEntry, processFunc); err != nil && taskEntry.ctx.Err() != context.Canceled {
				fmt.Printf("errored in task callback: %v\n", err)
			}
		})
	})
}

func (w *Worker) taskFunc(task *Task, taskEntry *taskEntry, processFunc ProcessFunc) error {
	claimWatch, err := w.claimCol.ReadOnly(taskEntry.ctx).WatchOne(task.ID, watch.WithFilterPut())
	if err != nil {
		return err
	}
	subtaskWatch, err := w.subtaskCol.ReadOnly(taskEntry.ctx).WatchOne(task.ID, watch.WithFilterDelete())
	if err != nil {
		return err
	}
	for {
		select {
		case e := <-claimWatch.Watch():
			if e.Type == watch.EventError {
				return e.Err
			}
			if e.Type != watch.EventDelete {
				continue
			}
			var subtaskKey string
			if err := e.Unmarshal(&subtaskKey, &Claim{}); err != nil {
				return err
			}
			taskEntry.runSubtask(w.subtaskFunc(subtaskKey, processFunc))
		case e := <-subtaskWatch.Watch():
			if e.Type == watch.EventError {
				return e.Err
			}
			var subtaskKey string
			if err := e.Unmarshal(&subtaskKey, &TaskInfo{}); err != nil {
				return err
			}
			taskEntry.runSubtask(w.subtaskFunc(subtaskKey, processFunc))
		case <-taskEntry.ctx.Done():
			return taskEntry.ctx.Err()
		}
	}
}

func (w *Worker) subtaskFunc(subtaskKey string, processFunc ProcessFunc) subtaskFunc {
	return func(ctx context.Context) {
		if err := func() error {
			// (bryce) this should be refactored to have the check and claim in the same stm.
			// there is a rare race condition that does not affect correctness, but it is less
			// than ideal because a subtask could get run once more than necessary.
			subtaskInfo := &TaskInfo{}
			if _, err := col.NewSTM(ctx, w.etcdClient, func(stm col.STM) error {
				return w.subtaskCol.ReadWrite(stm).Get(subtaskKey, subtaskInfo)
			}); err != nil {
				return err
			}
			if subtaskInfo.State != State_RUNNING {
				return nil
			}
			return w.claimCol.Claim(ctx, subtaskKey, &Claim{}, func(claimCtx context.Context) (retErr error) {
				subtask := subtaskInfo.Task
				defer func() {
					// If the task context was canceled or the claim was lost, just return with no error.
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
							subtaskInfo.Task = subtask
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
				return processFunc(claimCtx, subtask)
			})
		}(); err != nil {
			// If the task context was canceled or the subtask was deleted / not claimed, then no error should be logged.
			if ctx.Err() == context.Canceled ||
				col.IsErrNotFound(err) || err == col.ErrNotClaimed {
				return
			}
			fmt.Printf("errored in subtask callback: %v\n", err)
		}
	}
}
