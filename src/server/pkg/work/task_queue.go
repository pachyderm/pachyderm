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
	go func() {
	NextSubtask:
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("task queue context canceled: %v\n", ctx.Err())
			default:
			}
			tq.mu.Lock()
			iter := tq.tasks.IterFunc()
			for kv, ok := iter(); ok; kv, ok = iter() {
				taskChans := kv.Value.(*taskChans)
				select {
				case f := <-taskChans.subtaskFuncChan:
					tq.mu.Unlock()
					taskChans.errChan <- f(taskChans.ctx)
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

func (tq *taskQueue) createTask(ctx context.Context, taskID string) (*taskChans, error) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if _, ok := tq.tasks.Get(taskID); ok {
		return nil, fmt.Errorf("errored creating task %v, which already exists", taskID)
	}
	ctx, cancel := context.WithCancel(ctx)
	taskChans := &taskChans{
		ctx:             ctx,
		cancel:          cancel,
		subtaskFuncChan: make(chan subtaskFunc, 1),
		errChan:         make(chan error, 1),
	}
	tq.tasks.Set(taskID, taskChans)
	return taskChans, nil
}

type taskChans struct {
	ctx             context.Context
	cancel          context.CancelFunc
	subtaskFuncChan chan subtaskFunc
	errChan         chan error
}

func (t *taskChans) processSubtask(subtask subtaskFunc) error {
	select {
	case t.subtaskFuncChan <- subtask:
	case <-t.ctx.Done():
		return t.ctx.Err()
	}
	select {
	case err := <-t.errChan:
		return err
	case <-t.ctx.Done():
		return t.ctx.Err()
	}
}

func (tq *taskQueue) deleteTask(taskID string) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	tc, ok := tq.tasks.Get(taskID)
	if !ok {
		return fmt.Errorf("errored deleting task %v, which does not exist", taskID)
	}
	tc.(*taskChans).cancel()
	tq.tasks.Delete(taskID)
	return nil
}
