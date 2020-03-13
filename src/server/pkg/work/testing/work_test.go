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

type subtaskMaps struct {
	created, processed, collected []map[string]bool
	mu                            sync.Mutex
}

func newSubtaskMaps(numTasks int) *subtaskMaps {
	// Setup maps for checking creation, processing, and collection.
	taskMapsFunc := func() []map[string]bool {
		var taskMaps []map[string]bool
		for i := 0; i < numTasks; i++ {
			taskMaps = append(taskMaps, make(map[string]bool))
		}
		return taskMaps
	}
	return &subtaskMaps{
		created:   taskMapsFunc(),
		processed: taskMapsFunc(),
		collected: taskMapsFunc(),
	}
}

func test(t *testing.T, workerFailProb, taskCancelProb, subtaskFailProb float64) {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	require.NoError(t, testutil.WithEtcdEnv(func(env *testutil.EtcdEnv) error {
		numTasks := 10
		numSubtasks := 10
		numWorkers := 5
		sms := newSubtaskMaps(numTasks)
		// Setup workers.
		workerCtx, workerCancel := context.WithCancel(context.Background())
		eg, errCtx := errgroup.WithContext(workerCtx)
		for i := 0; i < numWorkers; i++ {
			eg.Go(func() error {
				w := work.NewWorker(errCtx, env.EtcdClient, "", "")
				for {
					ctx, cancel := context.WithCancel(errCtx)
					if err := w.Run(ctx, func(_ context.Context, task, subtask *work.Task) error {
						if rand.Float64() < workerFailProb {
							cancel()
						} else {
							i, err := strconv.Atoi(task.ID)
							if err != nil {
								t.Logf("errored parsing task ID: %v", err)
								t.Fail()
							}
							sms.mu.Lock()
							defer sms.mu.Unlock()
							if sms.processed[i] != nil {
								if sms.processed[i][subtask.ID] {
									t.Logf("claimed subtask should not be processed")
									t.Fail()
								}
								if rand.Float64() < subtaskFailProb {
									delete(sms.created[i], subtask.ID)
									return subtaskFailure
								}
								sms.processed[i][subtask.ID] = true
							}
						}
						return nil
					}); err != nil {
						if workerCtx.Err() == context.Canceled {
							return nil
						}
						if ctx.Err() == context.Canceled {
							continue
						}
						return err
					}
				}
				return nil
			})
		}
		tq, err := work.NewTaskQueue(errCtx, env.EtcdClient, "", "")
		require.NoError(t, err)
		var doneChans []chan struct{}
		for i := 0; i < numTasks; i++ {
			i := i
			taskID := strconv.Itoa(i)
			ctx, cancel := context.WithCancel(errCtx)
			doneChans = append(doneChans, make(chan struct{}))
			require.NoError(t, tq.RunTask(ctx, &work.Task{ID: taskID}, func(m *work.Master) {
				eg.Go(func() error {
					defer close(doneChans[i])
					// Create subtasks.
					var subtasks []*work.Task
					for j := 0; j < numSubtasks; j++ {
						subtaskID := strconv.Itoa(j)
						subtasks = append(subtasks, &work.Task{ID: subtaskID})
						sms.created[i][subtaskID] = true
					}
					//
					if rand.Float64() < taskCancelProb {
						cancel()
						sms.mu.Lock()
						defer sms.mu.Unlock()
						sms.created[i] = nil
						sms.processed[i] = nil
						sms.collected[i] = nil
						return nil
					}
					return m.RunSubtasks(subtasks, func(_ context.Context, subtaskInfo *work.TaskInfo) error {
						if subtaskInfo.State == work.State_FAILURE {
							if subtaskInfo.Reason != subtaskFailure.Error() {
								return fmt.Errorf("subtask failure reason does not equal subtask failure error message")
							}
							return nil
						}
						sms.collected[i][subtaskInfo.Task.ID] = true
						return nil
					})
				})
				<-doneChans[i]
			}), msg)
		}
		for _, doneChan := range doneChans {
			<-doneChan
		}
		workerCancel()
		require.NoError(t, eg.Wait(), msg)
		require.Equal(t, sms.created, sms.processed, sms.collected, msg)
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
