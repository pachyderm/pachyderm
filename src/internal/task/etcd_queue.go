package task

import (
	"context"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	"github.com/cevaris/ordered_map"
)

const (
	waitTime = 10 * time.Millisecond
	bufSize  = 10
)

type taskFunc func()

type groupEntry struct {
	cancel       context.CancelFunc
	taskFuncChan chan taskFunc
}

type taskQueue struct {
	groups *ordered_map.OrderedMap
	mu     sync.Mutex
}

func newTaskQueue(ctx context.Context) *taskQueue {
	tq := &taskQueue{
		groups: ordered_map.NewOrderedMap(),
	}
	go func() {
		var cycleKey string
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			// Get the next group.
			tq.mu.Lock()
			iter := tq.groups.IterFunc()
			kv, ok := iter()
			tq.mu.Unlock()
			var key string
			if ok {
				key = kv.Key.(string)
			}
			// Wait if there are no group entries or a cycle has been completed without finding a task.
			if !ok || key == cycleKey {
				time.Sleep(waitTime)
				cycleKey = ""
				continue
			}
			if cycleKey == "" {
				cycleKey = key
			}
			// Check if the next task is ready for the group and process it.
			ge := kv.Value.(*groupEntry)
			select {
			case cb := <-ge.taskFuncChan:
				cb()
				cycleKey = ""
			default:
			}
			tq.requeueGroup(key)
		}
	}()
	return tq
}

func (tq *taskQueue) group(ctx context.Context, groupID string, cb func(context.Context, chan taskFunc)) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if _, ok := tq.groups.Get(groupID); ok {
		return errors.Errorf("errored creating group %v, which already exists", groupID)
	}
	ctx, cancel := context.WithCancel(ctx)
	taskFuncChan := make(chan taskFunc, bufSize)
	ge := &groupEntry{
		cancel:       cancel,
		taskFuncChan: taskFuncChan,
	}
	tq.groups.Set(groupID, ge)
	go func() {
		cb(ctx, taskFuncChan)
	}()
	return nil
}

func (tq *taskQueue) requeueGroup(groupID string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	ge, ok := tq.groups.Get(groupID)
	if !ok {
		return
	}
	tq.groups.Delete(groupID)
	tq.groups.Set(groupID, ge)
}

func (tq *taskQueue) deleteGroup(groupID string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	ge, ok := tq.groups.Get(groupID)
	if !ok {
		return
	}
	ge.(*groupEntry).cancel()
	tq.groups.Delete(groupID)
}
