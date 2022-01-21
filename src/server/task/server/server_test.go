package server

import (
	"context"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"

	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/taskapi"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"golang.org/x/sync/errgroup"
)

var (
	errTaskFailure = errors.Errorf("task failure")
)

func serializeTestTask(testTask *task.TestTask) (*types.Any, error) {
	serializedTestTask, err := proto.Marshal(testTask)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + string(proto.MessageName(testTask)),
		Value:   serializedTestTask,
	}, nil
}

func deserializeTestTask(any *types.Any) (*task.TestTask, error) {
	testTask := &task.TestTask{}
	if err := types.UnmarshalAny(any, testTask); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return testTask, nil
}

func computeTaskID(input *types.Any) (string, error) {
	val, err := proto.Marshal(input)
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	sum := pachhash.Sum(val)
	return pachhash.EncodeHash(sum[:]), nil
}

func TestListTask(t *testing.T) {
	t.Parallel()
	env := testpachd.NewMockEnv(t)
	testNamespace := tu.UniqueString(t.Name())
	s := task.NewEtcdService(env.EtcdClient, "")
	api := NewAPIServer("", "", "", env.EtcdClient)
	env.MockPachd.Task.ListTask.Use(api.ListTask)

	numGroups := 10
	numTasks := 10
	numWorkers := 5

	claimedChan := make(chan struct{}, numWorkers)
	finishChan := make(chan struct{}, numWorkers)

	// deterministic failure for easy checking
	shouldFail := func(id string) bool {
		asInt, err := strconv.Atoi(id)
		require.NoError(t, err)
		return asInt%3 == 0
	}

	flushTasksAndVerify := func(flush bool) {
		if flush {
			for i := 0; i < numWorkers; i++ {
				<-claimedChan
			}
		}
		var groupTotalClaimed, totalClaimed int
		for g := 0; g < numGroups; g++ {
			tasks, err := env.PachClient.ListTask(path.Join(testNamespace, strconv.Itoa(g)))
			require.NoError(t, err)
			for i, info := range tasks {
				switch info.State {
				case taskapi.State_SUCCESS, taskapi.State_FAILURE:
					// tasks should be most-recently-created first, per group
					order := len(tasks) - 1 - i
					require.Equal(t, info.State == taskapi.State_FAILURE, (g*numTasks+order)%3 == 0)
				case taskapi.State_CLAIMED:
					groupTotalClaimed++
				default:
					require.Equal(t, taskapi.State_RUNNING, info.State)
				}
				// namespace/group/random/taskID
				pathParts := strings.Split(info.FullKey, "/")
				require.Equal(t, 4, len(pathParts))
				asInt, err := strconv.Atoi(pathParts[1])
				require.NoError(t, err)
				require.Equal(t, asInt, g)
			}
		}
		allTasks, err := env.PachClient.ListTask(testNamespace)
		require.NoError(t, err)

		for _, info := range allTasks {
			if info.State == taskapi.State_CLAIMED {
				totalClaimed++
			}
		}
		// it's possible for per-group results to be deleted between the API calls,
		// but claimed, unfinished tasks must still be present
		require.Equal(t, totalClaimed, groupTotalClaimed)

		if flush {
			require.Equal(t, numWorkers, groupTotalClaimed)
			for i := 0; i < numWorkers; i++ {
				finishChan <- struct{}{}
			}
		} else {
			require.Equal(t, 0, groupTotalClaimed)
		}
	}

	var groupEg errgroup.Group
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	workerEg, errCtx := errgroup.WithContext(workerCtx)
	for g := 0; g < numGroups; g++ {
		g := g
		groupEg.Go(func() error {
			var inputs []*types.Any
			for j := 0; j < numTasks; j++ {
				input, err := serializeTestTask(&task.TestTask{ID: strconv.Itoa(g*numTasks + j)})
				if err != nil {
					return err
				}
				inputs = append(inputs, input)
			}
			ctx, cancel := context.WithCancel(errCtx)
			defer cancel()
			d := s.NewDoer(testNamespace, strconv.Itoa(g))
			if err := task.DoBatch(ctx, d, inputs, func(j int64, output *types.Any, err error) error {
				if err != nil {
					if err.Error() != errTaskFailure.Error() {
						return errors.Errorf("task error message (%v) does not equal expected error message (%v)", err.Error(), errTaskFailure.Error())
					}
				} else {
					_, err = deserializeTestTask(output)
					if err != nil {
						return err
					}
				}
				return nil
			}); err != nil && !errors.Is(ctx.Err(), context.Canceled) {
				return err
			}
			return nil
		})
	}

	// check there's no task progress
	flushTasksAndVerify(false)

	for i := 0; i < numWorkers; i++ {
		workerEg.Go(func() error {
			src := s.NewSource("")
			for {
				if err := func() error {
					ctx, cancel := context.WithCancel(errCtx)
					defer cancel()
					err := src.Iterate(ctx, func(_ context.Context, input *types.Any) (*types.Any, error) {
						testTask, err := deserializeTestTask(input)
						if err != nil {
							return nil, err
						}
						// use channels to control task progress
						claimedChan <- struct{}{}
						<-finishChan
						if shouldFail(testTask.ID) {
							return nil, errTaskFailure
						}
						return serializeTestTask(testTask)
					})
					if errors.Is(ctx.Err(), context.Canceled) {
						return nil
					}
					return errors.EnsureStack(err)
				}(); err != nil {
					return err
				}
				if errors.Is(workerCtx.Err(), context.Canceled) {
					return nil
				}
			}
		})
	}

	// allow numWorkers tasks at a time to progress, and check ListTask results
	for i := 0; i < numTasks*numGroups; i += numWorkers {
		flushTasksAndVerify(true)
	}
	close(claimedChan)
	close(finishChan)
	require.NoError(t, groupEg.Wait())
	workerCancel()
	require.NoError(t, workerEg.Wait())
}
