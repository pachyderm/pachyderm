package work

import (
	"context"
	"fmt"
	"sync"

	"github.com/cevaris/ordered_map"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"golang.org/x/sync/errgroup"
)

type taskFunc func(context.Context, *Task) error
type subtaskFunc func()

type task struct {
	taskCtx, errCtx context.Context
	cancel          context.CancelFunc
	eg              *errgroup.Group
	subtaskChan     chan subtaskFunc
}

// (bryce) need a remap operation for task queue because the in memory maps
// do not free memory after deletions.
type taskQueue struct {
	tasks   *ordered_map.OrderedMap
	limiter limit.ConcurrencyLimiter
	ready   chan struct{}
	mu      sync.Mutex
}

func newTaskQueue(ctx context.Context, taskParallelism int) *taskQueue {
	tq := &taskQueue{
		tasks:   ordered_map.NewOrderedMap(),
		limiter: limit.New(taskParallelism),
		ready:   make(chan struct{}, taskParallelism),
	}
	// (bryce) safety measure for this returning?
	go func() {
		for {
			select {
			case <-tq.ready:
				tq.processNextSubtask()
			case <-ctx.Done():
				return
			}
		}
	}()
	return tq
}

func (tq *taskQueue) processNextSubtask() {
	tq.mu.Lock()
	iter := tq.tasks.IterFunc()
	for kv, ok := iter(); ok; kv, ok = iter() {
		task := kv.Value.(*task)
		select {
		case f := <-task.subtaskChan:
			tq.mu.Unlock()
			f()
			return
		default:
		}
	}
	tq.mu.Unlock()
}

func (tq *taskQueue) createTask(ctx context.Context, taskID string, f func(context.Context)) {
	if _, ok := tq.getTask(taskID); ok {
		return
	}
	// Acquire the right to proceed.
	tq.limiter.Acquire()
	// Create a context and cancel function for the new task.
	// Also, create an errgroup and context for errors.
	taskCtx, cancel := context.WithCancel(ctx)
	eg, errCtx := errgroup.WithContext(taskCtx)
	task := &task{
		taskCtx:     taskCtx,
		errCtx:      errCtx,
		cancel:      cancel,
		eg:          eg,
		subtaskChan: make(chan subtaskFunc, 1),
	}
	tq.setTask(taskID, task)
	go f(task.errCtx)
}

func (tq *taskQueue) sendSubtaskBatch(taskID string, subtasks *Task, f taskFunc) error {
	task, ok := tq.getTask(taskID)
	if !ok {
		return fmt.Errorf("error sending subtasks for task %v, which does not exist", taskID)
	}
	task.eg.Go(func() error {
		return filterTaskContextCanceledError(task.taskCtx, func() error {
			for _, subtask := range subtasks.Subtasks {
				select {
				case task.subtaskChan <- tq.subtaskFunc(task, subtask, f):
				case <-task.errCtx.Done():
					return task.errCtx.Err()
				}
				select {
				case tq.ready <- struct{}{}:
				case <-task.errCtx.Done():
					return task.errCtx.Err()
				}
			}
			return nil
		})
	})
	return nil
}

func (tq *taskQueue) subtaskFunc(task *task, subtask *Task, f taskFunc) subtaskFunc {
	return func() {
		task.eg.Go(func() error {
			return filterTaskContextCanceledError(task.taskCtx, func() error {
				return f(task.errCtx, subtask)
			})
		})
	}
}

func (tq *taskQueue) waitTask(taskID string) error {
	task, ok := tq.getTask(taskID)
	if !ok {
		return fmt.Errorf("error waiting task %v, which does not exist", taskID)
	}
	return task.eg.Wait()
}

func (tq *taskQueue) deleteTask(taskID string) error {
	task, ok := tq.getTask(taskID)
	if !ok {
		return fmt.Errorf("error deleting task %v, which does not exist", taskID)
	}
	defer func() {
		tq.mu.Lock()
		tq.tasks.Delete(taskID)
		tq.mu.Unlock()
		tq.limiter.Release()
	}()
	// Cancel the context for the task and wait for the goroutines.
	task.cancel()
	return task.eg.Wait()
}

func (tq *taskQueue) getTask(taskID string) (*task, bool) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	val, ok := tq.tasks.Get(taskID)
	if !ok {
		return nil, ok
	}
	return val.(*task), ok
}

func (tq *taskQueue) setTask(taskID string, task *task) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	tq.tasks.Set(taskID, task)
}

func filterTaskContextCanceledError(ctx context.Context, f func() error) error {
	if err := f(); err != nil && ctx.Err() != context.Canceled {
		return err
	}
	return nil
}
