package work

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cevaris/ordered_map"
)

const (
	waitTime = 100 * time.Millisecond
)

type subtaskFunc func(context.Context) error

type taskEntry struct {
	ctx             context.Context
	cancel          context.CancelFunc
	subtaskFuncChan chan subtaskFunc
	errChan         chan error
}

// runSubtask blocks on the task queue receiving the subtask function, executing it, then returning the error state.
// The task queue will not attempt to receive a subtask function until it will be processed.
func (te *taskEntry) runSubtask(subtask subtaskFunc) error {
	select {
	case te.subtaskFuncChan <- subtask:
	case <-te.ctx.Done():
		return te.ctx.Err()
	}
	select {
	case err := <-te.errChan:
		return err
	case <-te.ctx.Done():
		return te.ctx.Err()
	}
}

// The task queue data structure is an ordered map that stores task entries.
// Subtasks are sent through the subtask function channel in the task entries.
// The reason this design was chosen (as compared to a priority queue where the subtasks are the entries) is because it
// has a much lower memory footprint at scale, and our use case is such that the number of tasks in general will be
// significantly lower than the number of subtasks. Also, we are not concerned with the ordering of subtasks within a task,
// only the ordering of subtasks across tasks.
// (bryce) need a remap operation for task queue because the in memory maps
// do not free memory after deletions.
type taskQueue struct {
	tasks *ordered_map.OrderedMap
	mu    sync.Mutex
}

func newTaskQueue(ctx context.Context) *taskQueue {
	tq := &taskQueue{
		tasks: ordered_map.NewOrderedMap(),
	}
	// The next subtask to process is determined by iterating through the ordered map and checking the
	// subtask function channel for each task entry to see if the next subtask is ready to be processed.
	// If a subtask function is received, then it is executed and the error state is returned through
	// the task entry error channel.
	// After processing a subtask, the iteration starts from the beginning (new subtasks from earlier
	// tasks should be processed first).
	go func() {
	NextSubtask:
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("task queue context canceled: %v\n", ctx.Err())
				return
			default:
			}
			tq.mu.Lock()
			iter := tq.tasks.IterFunc()
			for kv, ok := iter(); ok; kv, ok = iter() {
				te := kv.Value.(*taskEntry)
				select {
				case f := <-te.subtaskFuncChan:
					tq.mu.Unlock()
					te.errChan <- f(te.ctx)
					continue NextSubtask
				default:
				}
			}
			tq.mu.Unlock()
			time.Sleep(waitTime)
		}
	}()
	return tq
}

// runTask runs a new task in the task queue.
// The task code should be contained within the passed in callback.
// The callback will receive a taskEntry, which should be used for running subtasks in the task queue.
// The task state will be cleaned up upon return of the callback.
func (tq *taskQueue) runTask(ctx context.Context, taskID string, f func(*taskEntry)) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if _, ok := tq.tasks.Get(taskID); ok {
		return fmt.Errorf("errored creating task %v, which already exists", taskID)
	}
	ctx, cancel := context.WithCancel(ctx)
	te := &taskEntry{
		ctx:             ctx,
		cancel:          cancel,
		subtaskFuncChan: make(chan subtaskFunc, 1),
		errChan:         make(chan error, 1),
	}
	tq.tasks.Set(taskID, te)
	go func() {
		defer tq.deleteTask(taskID)
		f(te)
	}()
	return nil
}

func (tq *taskQueue) deleteTask(taskID string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	tc, ok := tq.tasks.Get(taskID)
	if !ok {
		return
	}
	tc.(*taskEntry).cancel()
	tq.tasks.Delete(taskID)
}
