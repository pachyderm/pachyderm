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

func TestWork(t *testing.T) {
	require.NoError(t, testutil.WithEtcdEnv(func(env *testutil.EtcdEnv) error {
		etcdPrefix := generateRandomString(32)
		taskID := "task"
		numSubtasks := 10
		numWorkers := 5
		// Setup maps for checking creation, processing, and collection.
		subtasksCreated := make(map[string]bool)
		subtasksProcessed := make(map[string]bool)
		var processMu sync.Mutex
		subtasksCollected := make(map[string]bool)
		// Setup workers.
		// (bryce) bump back up when changing subtask layout.
		failProb := 0.0
		for i := 0; i < numWorkers; i++ {
			go func() {
				for {
					ctx, cancel := context.WithCancel(context.Background())
					w := work.NewWorker(context.Background(), env.EtcdClient, etcdPrefix, "")
					if err := w.Run(ctx, func(_ context.Context, task, subtask *work.Task) error {
						if rand.Float64() < failProb {
							cancel()
						} else {
							processMu.Lock()
							subtasksProcessed[subtask.ID] = true
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
		// Setup master.
		m := work.NewMaster(context.Background(), env.EtcdClient, etcdPrefix, "")
		// Create task.
		var subtasks []*work.Task
		for i := 0; i < numSubtasks; i++ {
			id := strconv.Itoa(i)
			subtasks = append(subtasks, &work.Task{ID: id})
			subtasksCreated[id] = true
		}
		task := &work.Task{
			ID:       taskID,
			Subtasks: subtasks,
		}
		require.NoError(t, m.CreateTask(context.Background(), task, func(_ context.Context) {}))
		require.NoError(t, m.SendSubtaskBatch(context.Background(), task.ID, task, func(_ context.Context, subtaskInfo *work.TaskInfo) error {
			subtasksCollected[subtaskInfo.Task.ID] = true
			return nil
		}))
		require.NoError(t, m.WaitTask(context.Background(), task.ID))
		//require.NoError(t, m.WaitTask(context.Background(), task.ID))
		require.Equal(t, subtasksCreated, subtasksProcessed, subtasksCollected)
		return nil
	}))
}
