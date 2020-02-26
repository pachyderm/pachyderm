package work

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"
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
	tq := newTaskQueue(context.Background(), numTasks)
	for i := 0; i < numTasks; i++ {
		tq.createTask(context.Background(), strconv.Itoa(i), func(_ context.Context) {})
	}
	for j := 0; j < numSubtasks; j++ {
		for i := 0; i < numTasks; i++ {
			i := i
			tq.sendSubtaskBatch(strconv.Itoa(i), &Task{Subtasks: []*Task{&Task{ID: strconv.Itoa(j)}}}, func(_ context.Context, _ *Task) error {
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				testTasks[i].subtaskChan <- struct{}{}
				return nil
			})
		}
		for i := 0; i < numTasks; i++ {
			select {
			case <-testTasks[i].subtaskChan:
			}
		}
	}
}
