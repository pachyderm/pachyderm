package testing

import (
	"context"
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/testetcd"
	"github.com/pachyderm/pachyderm/src/server/pkg/work2"

	types "github.com/gogo/protobuf/types"
	"golang.org/x/sync/errgroup"
)

func wrapCallback(cb func(context.Context, *TestData) error) func(context.Context, *types.Any) error {
	return func(ctx context.Context, any *types.Any) error {
		testData := &TestData{}
		if err := types.UnmarshalAny(any, testData); err != nil {
			return err
		}
		return cb(ctx, testData)
	}
}

func runWorkerMasterPair(
	numTasks int,
	taskCb func(context.Context, work2.Task) error,
	workerCb func(context.Context, *TestData) error,
) error {
	return testetcd.WithEtcdEnv(func(env *testetcd.EtcdEnv) error {
		ctx, cancel := context.WithCancel(env.Context)
		eg, ctx := errgroup.WithContext(ctx)

		eg.Go(func() error {
			worker := work2.NewWorker(env.EtcdClient, "", "")
			return worker.Run(ctx, wrapCallback(workerCb))
		})

		eg.Go(func() error {
			defer cancel()
			taskEg, ctx := errgroup.WithContext(ctx)

			master := work2.NewMaster(env.EtcdClient, "", "")
			for i := 0; i < numTasks; i++ {
				taskEg.Go(func() error {
					return master.WithTask(func(task work2.Task) error {
						defer fmt.Printf("task done\n")
						fmt.Printf("running task\n")
						return taskCb(ctx, task)
					})
				})
			}

			return taskEg.Wait()
		})

		return eg.Wait()
	})
}

func TestNoSubtasks(t *testing.T) {
	err := runWorkerMasterPair(
		1,
		func(ctx context.Context, task work2.Task) error {
			subtaskChan := make(chan *types.Any)
			close(subtaskChan)

			return task.Run(ctx, subtaskChan, wrapCallback(func(ctx context.Context, data *TestData) error {
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

func TestBasic(t *testing.T) {
}

func TestMultiRun(t *testing.T) {
	// use task.Run multiple times
}

func TestMultiTask(t *testing.T) {
}

func TestMultiWorker(t *testing.T) {
}

func TestExecutionOrder(t *testing.T) {
}

func TestClear(t *testing.T) {
}
