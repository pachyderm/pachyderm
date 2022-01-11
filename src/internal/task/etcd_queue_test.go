package task

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

type testGroup struct {
	taskChan chan struct{}
}

func TestTaskQueue(t *testing.T) {
	numGroups := 5
	numTasks := 5
	testGroups := []*testGroup{}
	for i := 0; i < numGroups; i++ {
		testGroups = append(testGroups, &testGroup{
			taskChan: make(chan struct{}),
		})
	}
	tq := newTaskQueue(context.Background())
	ready := make(chan struct{})
	for i := 0; i < numGroups; i++ {
		i := i
		require.NoError(t, tq.group(context.Background(), strconv.Itoa(i), func(_ context.Context, taskFuncChan chan taskFunc) {
			// The first group will create a task that sleeps a bit to allow the tasks
			// from the subsequent groups to queue up.
			if i == 0 {
				taskFuncChan <- func() {
					close(ready)
					time.Sleep(1 * time.Second)
					testGroups[0].taskChan <- struct{}{}
				}
				for j := 1; j < numTasks; j++ {
					taskFuncChan <- func() {
						testGroups[0].taskChan <- struct{}{}
					}
				}
			} else {
				<-ready
				for j := 0; j < numTasks; j++ {
					taskFuncChan <- func() {
						testGroups[i].taskChan <- struct{}{}
					}
				}
			}
		}))
	}
	for j := 0; j < numTasks; j++ {
		for i := 0; i < numGroups; i++ {
			<-testGroups[i].taskChan
		}
	}
}
