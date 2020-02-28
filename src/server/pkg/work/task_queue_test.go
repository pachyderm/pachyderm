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
	testTasks := []*testTask{}
	for i := 0; i < numTasks; i++ {
		testTasks = append(testTasks, &testTask{
			subtaskChan: make(chan struct{}),
		})
	}
	tq := newTaskQueue(context.Background())
	var tcs []*taskChans
	for i := 0; i < numTasks; i++ {
		tc, err := tq.createTask(context.Background(), strconv.Itoa(i))
		require.NoError(t, err)
		tcs = append(tcs, tc)
	}
	for j := 0; j < numSubtasks; j++ {
		// Set up a subtask that sleeps for a bit to allow us to queue up some subtasks.
		readyChan := make(chan struct{})
		go func() {
			tcs[0].processSubtask(func(_ context.Context) error {
				close(readyChan)
				time.Sleep(1 * time.Second)
				return nil
			})
		}()
		<-readyChan
		// Queue up some subtasks, and confirm that they are completed in the correct order.
		for i := 1; i < len(tcs); i++ {
			i := i
			go func() {
				tcs[i].processSubtask(func(_ context.Context) error {
					testTasks[i].subtaskChan <- struct{}{}
					return nil
				})
			}()
		}
		for i := 1; i < len(tcs); i++ {
			select {
			case <-testTasks[i].subtaskChan:
			}
		}
	}
}
