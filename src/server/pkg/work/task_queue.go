package work

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cevaris/ordered_map"
)

const (
	waitTime = 10 * time.Millisecond
)

var (
	remapThreshold = 10000
)

type subtaskFunc func(context.Context)
type subtaskBlockFunc func(context.Context) error

type taskEntry struct {
	ctx             context.Context
	cancel          context.CancelFunc
	subtaskFuncChan chan subtaskFunc
}

// runSubtask sends a subtask to be run in the task queue.
// The task queue will not attempt to receive a subtask function until it will be processed.
func (te *taskEntry) runSubtask(subtask subtaskFunc) {
	select {
	case te.subtaskFuncChan <- subtask:
	case <-te.ctx.Done():
	}
}

// runSubtaskBlock is similar to runSubtask, but blocks on the subtask.
func (te *taskEntry) runSubtaskBlock(subtask subtaskBlockFunc) error {
	errChan := make(chan error)
	te.runSubtask(func(ctx context.Context) {
		errChan <- subtask(ctx)
	})
	return <-errChan
}

// The task queue data structure is an ordered map that stores task entries.
// Subtasks are sent through the subtask function channel in the task entries.
// The reason this design was chosen (as compared to a priority queue where the subtasks are the entries) is because it
// has a much lower memory footprint at scale, and our use case is such that the number of tasks in general will be
// significantly lower than the number of subtasks. Also, we are not concerned with the ordering of subtasks within a task,
// only the ordering of subtasks across tasks.
type taskQueue struct {
	tasks                  *ordered_map.OrderedMap
	mu                     sync.Mutex
	tasksDeletedSinceRemap int
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
					f(te.ctx)
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
	}
	tq.tasks.Set(taskID, te)
	go func() {
		defer tq.deleteTask(taskID)
		f(te)
	}()
	return nil
}

// maybeRemap copies the entries in the ordered map to a new ordered map after a certain number of
// tasks have been deleted. This is to prevent unbounded memory usage due to maps not freeing
// memory after deletions.
func (tq *taskQueue) maybeRemap() {
	tq.tasksDeletedSinceRemap++
	if tq.tasksDeletedSinceRemap >= remapThreshold {
		var kvs []*ordered_map.KVPair
		iter := tq.tasks.IterFunc()
		for kv, ok := iter(); ok; kv, ok = iter() {
			kvs = append(kvs, kv)
		}
		tq.tasks = ordered_map.NewOrderedMapWithArgs(kvs)
		tq.tasksDeletedSinceRemap = 0
	}
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
	tq.maybeRemap()
}
