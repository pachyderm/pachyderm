package testing

import (
	"context"
	"fmt"
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

var (
	subtaskFailure = fmt.Errorf("subtask failure")
)

func seedStr(seed int64) string {
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}

func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('a' + rand.Intn(26))
	}
	return string(b)
}

func test(t *testing.T, workerFailProb, taskCancelProb, subtaskFailProb float64) {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
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
		for i := 0; i < numWorkers; i++ {
			go func() {
				for {
					ctx, cancel := context.WithCancel(context.Background())
					w := work.NewWorker(context.Background(), env.EtcdClient, etcdPrefix, "")
					if err := w.Run(ctx, func(_ context.Context, task, subtask *work.Task) error {
						if rand.Float64() < workerFailProb {
							cancel()
						} else {
							i, err := strconv.Atoi(task.ID)
							require.NoError(t, err, msg)
							processMu.Lock()
							defer processMu.Unlock()
							if subtasksProcessed[i] != nil {
								require.False(t, subtasksProcessed[i][subtask.ID], msg)
								if rand.Float64() < subtaskFailProb {
									delete(subtasksCreated[i], subtask.ID)
									return subtaskFailure
								}
								subtasksProcessed[i][subtask.ID] = true
							}
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
			ctx, cancel := context.WithCancel(context.Background())
			m, err := tq.CreateTask(ctx, &work.Task{ID: taskID})
			require.NoError(t, err, msg)
			eg.Go(func() error {
				m.Collect(func(_ context.Context, subtaskInfo *work.TaskInfo) error {
					if subtaskInfo.State == work.State_FAILURE {
						require.Equal(t, subtaskFailure.Error(), subtaskInfo.Reason, msg)
						return nil
					}
					subtasksCollected[i][subtaskInfo.Task.ID] = true
					return nil
				})
				// Create subtasks.
				for j := 0; j < numSubtasks; j++ {
					subtaskID := strconv.Itoa(j)
					require.NoError(t, m.CreateSubtask(&work.Task{ID: subtaskID}), msg)
					subtasksCreated[i][subtaskID] = true
				}
				if rand.Float64() < taskCancelProb {
					cancel()
					require.Equal(t, ctx.Err(), m.Wait(), msg)
					subtasksCreated[i] = nil
					processMu.Lock()
					subtasksProcessed[i] = nil
					processMu.Unlock()
					subtasksCollected[i] = nil
				} else {
					require.NoError(t, m.Wait(), msg)
				}
				require.NoError(t, tq.DeleteTask(context.Background(), taskID), msg)
				return nil
			})
		}
		require.NoError(t, eg.Wait(), msg)
		require.Equal(t, subtasksCreated, subtasksProcessed, subtasksCollected, msg)
		return nil
	}), msg)
}

func TestBasic(t *testing.T) {
	test(t, 0, 0, 0)
}

func TestWorkerCrashes(t *testing.T) {
	test(t, 0.1, 0, 0)
}

func TestCancelTasks(t *testing.T) {
	test(t, 0, 0.2, 0)
}

func TestSubtaskFailures(t *testing.T) {
	test(t, 0, 0, 0.1)
}

func TestEverything(t *testing.T) {
	test(t, 0.1, 0.2, 0.1)
}
