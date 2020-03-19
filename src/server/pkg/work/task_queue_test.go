package work

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

type testTask struct {
	count       int
	subtaskChan chan struct{}
}

func TestTaskQueue(t *testing.T) {
	numTasks := 5
	numSubtasks := 5
	// Just a sanity check for the remap operation.
	remapThreshold = 2
	testTasks := []*testTask{}
	for i := 0; i < numTasks; i++ {
		testTasks = append(testTasks, &testTask{
			subtaskChan: make(chan struct{}),
		})
	}
	tq := newTaskQueue(context.Background())
	var readyChans, doneChans []chan struct{}
	for i := 0; i < numSubtasks; i++ {
		readyChans = append(readyChans, make(chan struct{}))
		doneChans = append(doneChans, make(chan struct{}))
	}
	for i := 0; i < numTasks; i++ {
		i := i
		require.NoError(t, tq.runTask(context.Background(), strconv.Itoa(i), func(taskEntry *taskEntry) {
			for j := 0; j < numSubtasks; j++ {
				if i == 0 {
					// The first task will create subtasks that sleep a bit to allow the the subtasks
					// from the subsequent tasks to queue up.
					require.NoError(t, taskEntry.runSubtaskBlock(func(_ context.Context) error {
						close(readyChans[j])
						time.Sleep(1 * time.Second)
						return nil
					}))
				} else {
					<-readyChans[j]
					// The subtasks in the subsequent tasks should execute in task creation order.
					require.NoError(t, taskEntry.runSubtaskBlock(func(_ context.Context) error {
						testTasks[i].subtaskChan <- struct{}{}
						return nil
					}))
					if i == numTasks-1 {
						close(doneChans[j])
					}
				}
				<-doneChans[j]
			}
		}))
	}
	for j := 0; j < numSubtasks; j++ {
		for i := 1; i < numTasks; i++ {
			select {
			case <-testTasks[i].subtaskChan:
			}
		}
	}
}
