package testing

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"golang.org/x/sync/errgroup"
)

// generateRandomString is a helper function for getPachClient
func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('a' + rand.Intn(26))
	}
	return string(b)
}

func TestTaskQueue(t *testing.T) {
	require.NoError(t, testutil.WithEtcdEnv(func(env *testutil.EtcdEnv) error {
		etcdPrefix := generateRandomString(32)
		numTasks := 10
		numSubtasks := 10
		numWorkers := 5
		// Setup maps for checking creation, processing, and collection.
		taskMapsFunc := func() []map[string]bool {
			var taskMaps []map[string]bool
			for i := 0; i < numTasks; i++ {
				taskMaps = append(taskMaps, make(map[string]bool))
			}
			return taskMaps
		}
		subtasksCreated := taskMapsFunc()
		subtasksProcessed := taskMapsFunc()
		var processMu sync.Mutex
		subtasksCollected := taskMapsFunc()
		// Setup workers.
		failProb := 0.1
		for i := 0; i < numWorkers; i++ {
			go func() {
				for {
					ctx, cancel := context.WithCancel(context.Background())
					w := work.NewWorker(context.Background(), env.EtcdClient, etcdPrefix, "")
					if err := w.Run(ctx, func(_ context.Context, task, subtask *work.Task) error {
						if rand.Float64() < failProb {
							cancel()
						} else {
							i, err := strconv.Atoi(task.ID)
							require.NoError(t, err)
							processMu.Lock()
							require.False(t, subtasksProcessed[i][subtask.ID])
							subtasksProcessed[i][subtask.ID] = true
							processMu.Unlock()
						}
						return nil
					}); err != nil {
						if ctx.Err() == context.Canceled {
							continue
						}
						break
					}
				}
			}()
		}
		tq := work.NewTaskQueue(context.Background(), env.EtcdClient, etcdPrefix, "")
		var eg errgroup.Group
		for i := 0; i < numTasks; i++ {
			i := i
			taskID := strconv.Itoa(i)
			m, err := tq.CreateTask(context.Background(), &work.Task{ID: taskID})
			require.NoError(t, err)
			eg.Go(func() error {
				m.Collect(func(_ context.Context, subtaskInfo *work.TaskInfo) error {
					subtasksCollected[i][subtaskInfo.Task.ID] = true
					return nil
				})
				// Create subtasks.
				for j := 0; j < numSubtasks; j++ {
					subtaskID := strconv.Itoa(j)
					require.NoError(t, m.CreateSubtask(&work.Task{ID: subtaskID}))
					subtasksCreated[i][subtaskID] = true
				}
				require.NoError(t, m.Wait())
				require.NoError(t, tq.DeleteTask(context.Background(), taskID))
				return nil
			})
		}
		require.NoError(t, eg.Wait())
		require.Equal(t, subtasksCreated, subtasksProcessed, subtasksCollected)
		return nil
	}))
}
