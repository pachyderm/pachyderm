package testing

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/testetcd"
	"github.com/pachyderm/pachyderm/src/server/pkg/work2"

	types "github.com/gogo/protobuf/types"
	"golang.org/x/sync/errgroup"
)

// Helper type to make sure callbacks do not get run concurrently
type nonConcurrentCheck struct {
	t         *testing.T
	mutex     sync.Mutex
	happening bool
}

func (ncc *nonConcurrentCheck) Check(cb func()) {
	flip := func(newVal bool) {
		ncc.mutex.Lock()
		defer ncc.mutex.Unlock()
		require.Equal(ncc.t, !newVal, ncc.happening, "callback is running concurrently")
		ncc.happening = newVal
	}

	flip(true)
	defer flip(false)

	cb()
}

func newConcurrentCheck(t *testing.T) *nonConcurrentCheck {
	return &nonConcurrentCheck{t: t}
}

func wrapCallback(ncc *nonConcurrentCheck, cb func(context.Context, *TestData) error) func(context.Context, *types.Any) error {
	return func(ctx context.Context, any *types.Any) error {
		testData := &TestData{}
		if err := types.UnmarshalAny(any, testData); err != nil {
			fmt.Printf("wrapped callback deserialization err: %v\n", err)
			return err
		}
		var err error
		ncc.Check(func() {
			err = cb(ctx, testData)
		})
		fmt.Printf("wrapped callback returning err: %v\n", err)
		return err
	}
}

type TestOptions struct {
	Tasks   int
	Workers int
}

func runWorkerMasterPair(
	t *testing.T,
	options TestOptions,
	taskCb func(context.Context, int, work2.Task, *nonConcurrentCheck) error,
	workerCb func(context.Context, *TestData) error,
) error {
	return testetcd.WithEtcdEnv(func(env *testetcd.EtcdEnv) error {
		ctx, cancel := context.WithCancel(env.Context)
		eg, ctx := errgroup.WithContext(ctx)

		for i := 0; i < options.Workers; i++ {
			eg.Go(func() error {
				workerCheck := newConcurrentCheck(t)
				worker := work2.NewWorker(env.EtcdClient, "", "")
				return worker.Run(ctx, wrapCallback(workerCheck, workerCb))
			})
		}

		eg.Go(func() error {
			defer cancel()
			taskEg, ctx := errgroup.WithContext(ctx)

			master := work2.NewMaster(env.EtcdClient, "", "")
			for i := 0; i < options.Tasks; i++ {
				i := i
				taskEg.Go(func() error {
					taskCheck := newConcurrentCheck(t)
					return master.WithTask(func(task work2.Task) error {
						return taskCb(ctx, i, task, taskCheck)
					})
				})
			}

			err := taskEg.Wait()
			fmt.Printf("master tasks returned: %v\n", err)
			return err
		})

		return eg.Wait()
	})
}

func TestNoSubtasks(t *testing.T) {
	err := runWorkerMasterPair(
		t, TestOptions{Tasks: 1, Workers: 1},
		func(ctx context.Context, i int, task work2.Task, ncc *nonConcurrentCheck) error {
			subtaskChan := make(chan *types.Any)
			close(subtaskChan)

			return task.Run(ctx, subtaskChan, wrapCallback(ncc, func(ctx context.Context, data *TestData) error {
				require.True(t, false, "no subtasks were made, collection should not happen")
				return nil
			}))
		},
		func(ctx context.Context, data *TestData) error {
			require.True(t, false, "no subtasks were made, work should not happen")
			return nil
		},
	)
	require.NoError(t, err)
}

func sendSubtasks(keys []string, subtaskChan chan *types.Any) error {
	for _, key := range keys {
		any, err := types.MarshalAny(&TestData{Key: key})
		if err != nil {
			return err
		}
		fmt.Printf("master adding subtask: %v\n", key)
		subtaskChan <- any
	}
	return nil
}

func doTest(t *testing.T, numWorkers int, tasks [][]string) {
	allTasks := []string{}
	for _, subtasks := range tasks {
		allTasks = append(allTasks, subtasks...)
	}
	runTasks := []string{}
	collectedTasks := []string{}

	err := runWorkerMasterPair(
		t, TestOptions{Tasks: len(tasks), Workers: numWorkers},
		func(ctx context.Context, i int, task work2.Task, ncc *nonConcurrentCheck) error {
			eg := errgroup.Group{}
			eg.Go(func() error {
				subtaskChan := make(chan *types.Any, 10)
				require.NoError(t, sendSubtasks(tasks[i], subtaskChan))
				close(subtaskChan)

				return task.Run(ctx, subtaskChan, wrapCallback(ncc, func(ctx context.Context, data *TestData) error {
					fmt.Printf("master collecting first-round task: %v\n", data.Key)
					collectedTasks = append(collectedTasks, data.Key)
					time.Sleep(5 * time.Millisecond) // Leave a window to check that this isn't run concurrently
					return nil
				}))
			})

			return eg.Wait()
		},
		func(ctx context.Context, data *TestData) error {
			fmt.Printf("worker running task: %v\n", data.Key)
			runTasks = append(runTasks, data.Key)
			time.Sleep(5 * time.Millisecond) // Leave a window to check that this isn't run concurrently
			return nil
		},
	)
	require.NoError(t, err)
	require.ElementsEqual(t, allTasks, runTasks)
	require.ElementsEqual(t, allTasks, collectedTasks)
}

func TestBasic(t *testing.T) {
	doTest(t, 1, [][]string{
		{"a", "b"},
	})
}

func TestMultiTask(t *testing.T) {
	doTest(t, 1, [][]string{
		{"a", "b"},
		{"c", "d"},
		{"e"},
		{},
		{"f", "g", "h", "i"},
	})
}

func TestMultiWorker(t *testing.T) {
	doTest(t, 4, [][]string{
		{"a", "b", "c", "d", "e"},
	})
}

func TestMultiTaskAndWorker(t *testing.T) {
	doTest(t, 4, [][]string{
		{"a", "b"},
		{"c", "d"},
		{"e"},
		{},
		{"f", "g", "h", "i"},
	})
}

func TestMultiRun(t *testing.T) {
	expectedTasks := []string{"a", "b", "c", "d", "e"}
	runTasks := []string{}
	collectedTasks := []string{}

	err := runWorkerMasterPair(
		t, TestOptions{Tasks: 1, Workers: 1},
		func(ctx context.Context, i int, task work2.Task, ncc *nonConcurrentCheck) error {
			subtaskChan := make(chan *types.Any, 10)
			require.NoError(t, sendSubtasks(expectedTasks[0:2], subtaskChan))
			close(subtaskChan)

			if err := task.Run(ctx, subtaskChan, wrapCallback(ncc, func(ctx context.Context, data *TestData) error {
				fmt.Printf("master collecting first-round task: %v\n", data.Key)
				collectedTasks = append(collectedTasks, data.Key)
				return nil
			})); err != nil {
				return err
			}

			subtaskChan = make(chan *types.Any, 10)
			require.NoError(t, sendSubtasks(expectedTasks[2:], subtaskChan))
			close(subtaskChan)

			return task.Run(ctx, subtaskChan, wrapCallback(ncc, func(ctx context.Context, data *TestData) error {
				fmt.Printf("master collecting second-round task: %v\n", data.Key)
				collectedTasks = append(collectedTasks, data.Key)
				return nil
			}))
		},
		func(ctx context.Context, data *TestData) error {
			fmt.Printf("worker running task: %v\n", data.Key)
			runTasks = append(runTasks, data.Key)
			return nil
		},
	)
	require.NoError(t, err)
	require.ElementsEqual(t, expectedTasks, runTasks)
	require.ElementsEqual(t, expectedTasks, collectedTasks)
}

func TestParallelRun(t *testing.T) {
	expectedTasks := []string{"a", "b", "c", "d", "e"}
	runTasks := []string{}
	collectedTasks := []string{}

	err := runWorkerMasterPair(
		t, TestOptions{Tasks: 1, Workers: 1},
		func(ctx context.Context, i int, task work2.Task, ncc *nonConcurrentCheck) error {
			eg := errgroup.Group{}
			eg.Go(func() error {
				subtaskChan := make(chan *types.Any, 10)
				require.NoError(t, sendSubtasks(expectedTasks[0:2], subtaskChan))
				close(subtaskChan)

				return task.Run(ctx, subtaskChan, wrapCallback(ncc, func(ctx context.Context, data *TestData) error {
					fmt.Printf("master collecting first-round task: %v\n", data.Key)
					collectedTasks = append(collectedTasks, data.Key)
					return nil
				}))
			})

			eg.Go(func() error {
				subtaskChan := make(chan *types.Any, 10)
				require.NoError(t, sendSubtasks(expectedTasks[2:], subtaskChan))
				close(subtaskChan)

				return task.Run(ctx, subtaskChan, wrapCallback(ncc, func(ctx context.Context, data *TestData) error {
					fmt.Printf("master collecting second-round task: %v\n", data.Key)
					collectedTasks = append(collectedTasks, data.Key)
					return nil
				}))
			})

			return eg.Wait()
		},
		func(ctx context.Context, data *TestData) error {
			fmt.Printf("worker running task: %v\n", data.Key)
			runTasks = append(runTasks, data.Key)
			return nil
		},
	)
	require.NoError(t, err)
	require.ElementsEqual(t, expectedTasks, runTasks)
	require.ElementsEqual(t, expectedTasks, collectedTasks)
}

func TestExecutionOrder(t *testing.T) {
}

func TestClear(t *testing.T) {
}
