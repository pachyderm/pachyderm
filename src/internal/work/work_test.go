package work

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"golang.org/x/sync/errgroup"
)

var (
	errSubtaskFailure = errors.Errorf("subtask failure")
)

func seedStr(seed int64) string {
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}

func serializeTestData(testData *TestData) (*types.Any, error) {
	serializedTestData, err := proto.Marshal(testData)
	if err != nil {
		return nil, err
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(testData),
		Value:   serializedTestData,
	}, nil
}

func deserializeTestData(testDataAny *types.Any) (*TestData, error) {
	testData := &TestData{}
	if err := types.UnmarshalAny(testDataAny, testData); err != nil {
		return nil, err
	}
	return testData, nil
}

func processSubtask(t *testing.T, subtask *Task) error {
	testData, err := deserializeTestData(subtask.Data)
	if err != nil {
		return err
	}
	if testData.Processed {
		t.Logf("claimed subtask should not already be processed")
		t.Fail()
	}
	testData.Processed = true
	subtask.Data, err = serializeTestData(testData)
	if err != nil {
		return err
	}
	return nil
}

func collectSubtask(subtaskInfo *TaskInfo, collected map[string]bool) error {
	testData, err := deserializeTestData(subtaskInfo.Task.Data)
	if err != nil {
		return err
	}
	if !testData.Processed {
		return errors.Errorf("collected subtask should be processed")
	}
	collected[subtaskInfo.Task.ID] = true
	if subtaskInfo.State == State_FAILURE && subtaskInfo.Reason != errSubtaskFailure.Error() {
		return errors.Errorf("subtask failure reason does not equal subtask failure error message")
	}
	return nil
}

func test(t *testing.T, workerFailProb, taskCancelProb, subtaskFailProb float64) {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	env := testetcd.NewEnv(t)

	numTasks := 10
	numSubtasks := 10
	numWorkers := 5
	// Setup workers.
	workerCtx, workerCancel := context.WithCancel(context.Background())
	workerEg, errCtx := errgroup.WithContext(workerCtx)
	for i := 0; i < numWorkers; i++ {
		workerEg.Go(func() error {
			w := NewWorker(env.EtcdClient, "", "")
			for {
				ctx, cancel := context.WithCancel(errCtx)
				if err := w.Run(ctx, func(_ context.Context, subtask *Task) (*types.Any, error) {
					if rand.Float64() < workerFailProb {
						cancel()
						return nil, nil
					}
					if err := processSubtask(t, subtask); err != nil {
						return nil, err
					}
					if rand.Float64() < subtaskFailProb {
						return nil, errSubtaskFailure
					}
					return nil, nil
				}); err != nil {
					if errors.Is(workerCtx.Err(), context.Canceled) {
						return nil
					}
					if errors.Is(ctx.Err(), context.Canceled) {
						continue
					}
					return err
				}
			}
		})
	}
	tq, err := NewTaskQueue(errCtx, env.EtcdClient, "", "")
	require.NoError(t, err)
	taskMapsFunc := func() []map[string]bool {
		var taskMaps []map[string]bool
		for i := 0; i < numTasks; i++ {
			taskMaps = append(taskMaps, make(map[string]bool))
		}
		return taskMaps
	}
	created := taskMapsFunc()
	collected := taskMapsFunc()
	var taskEg errgroup.Group
	for i := 0; i < numTasks; i++ {
		i := i
		taskEg.Go(func() error {
			ctx, cancel := context.WithCancel(errCtx)
			if err := tq.RunTaskBlock(ctx, func(m *Master) error {
				if rand.Float64() < taskCancelProb {
					cancel()
					return nil
				}
				// Create subtasks.
				var subtasks []*Task
				for j := 0; j < numSubtasks; j++ {
					ID := strconv.Itoa(j)
					data, err := serializeTestData(&TestData{})
					if err != nil {
						return err
					}
					subtasks = append(subtasks, &Task{
						ID:   ID,
						Data: data,
					})
					created[i][ID] = true
				}
				return m.RunSubtasks(subtasks, func(_ context.Context, subtaskInfo *TaskInfo) error {
					return collectSubtask(subtaskInfo, collected[i])
				})
			}); err != nil && !errors.Is(ctx.Err(), context.Canceled) {
				return err
			}
			return nil
		})
	}
	require.NoError(t, taskEg.Wait(), msg)
	workerCancel()
	require.NoError(t, workerEg.Wait(), msg)
	require.Equal(t, created, collected, msg)
}

func TestBasic(t *testing.T) {
	t.Parallel()
	test(t, 0, 0, 0)
}

func TestWorkerCrashes(t *testing.T) {
	t.Parallel()
	test(t, 0.1, 0, 0)
}

func TestCancelTasks(t *testing.T) {
	t.Parallel()
	test(t, 0, 0.2, 0)
}

func TestSubtaskFailures(t *testing.T) {
	t.Parallel()
	test(t, 0, 0, 0.1)
}

func TestEverything(t *testing.T) {
	t.Parallel()
	test(t, 0.1, 0.2, 0.1)
}

func TestRunZeroSubtasks(t *testing.T) {
	t.Parallel()
	env := testetcd.NewEnv(t)

	tq, err := NewTaskQueue(context.Background(), env.EtcdClient, "", "")
	require.NoError(t, err)

	err = tq.RunTaskBlock(context.Background(), func(m *Master) error {
		return m.RunSubtasks(nil, func(_ context.Context, _ *TaskInfo) error {
			return nil
		})
	})
	require.NoError(t, err)
}
