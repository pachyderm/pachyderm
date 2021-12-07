package task

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
	errTaskFailure = errors.Errorf("task failure")
)

func seedStr(seed int64) string {
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}

func serializeTestTask(testTask *TestTask) (*types.Any, error) {
	serializedTestTask, err := proto.Marshal(testTask)
	if err != nil {
		return nil, err
	}
	return &types.Any{
		TypeUrl: "/" + string(proto.MessageName(testTask)),
		Value:   serializedTestTask,
	}, nil
}

func deserializeTestTask(any *types.Any) (*TestTask, error) {
	testTask := &TestTask{}
	if err := types.UnmarshalAny(any, testTask); err != nil {
		return nil, err
	}
	return testTask, nil
}

func test(t *testing.T, workerFailProb, groupCancelProb, taskFailProb float64) {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	env := testetcd.NewEnv(t)
	numGroups := 10
	numTasks := 10
	numWorkers := 5
	// Set up task service.
	s := NewEtcdService(env.EtcdClient, "")
	// Set up workers.
	workerCtx, workerCancel := context.WithCancel(context.Background())
	workerEg, errCtx := errgroup.WithContext(workerCtx)
	for i := 0; i < numWorkers; i++ {
		workerEg.Go(func() error {
			maker := s.Maker("")
			for {
				ctx, cancel := context.WithCancel(errCtx)
				if err := maker.Make(ctx, func(_ context.Context, input *types.Any) (*types.Any, error) {
					if rand.Float64() < workerFailProb {
						cancel()
						return nil, nil
					}
					if rand.Float64() < taskFailProb {
						return nil, errTaskFailure
					}
					testTask, err := deserializeTestTask(input)
					if err != nil {
						return nil, err
					}
					return serializeTestTask(testTask)
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
	created := [][]bool{}
	collected := [][]bool{}
	for i := 0; i < numGroups; i++ {
		created = append(created, make([]bool, numTasks))
		collected = append(collected, make([]bool, numTasks))
	}
	// Create groups.
	var groupEg errgroup.Group
	for i := 0; i < numGroups; i++ {
		i := i
		groupEg.Go(func() error {
			var inputs []*types.Any
			for j := 0; j < numTasks; j++ {
				input, err := serializeTestTask(&TestTask{ID: strconv.Itoa(j)})
				if err != nil {
					return err
				}
				inputs = append(inputs, input)
				created[i][j] = true
			}
			ctx, cancel := context.WithCancel(errCtx)
			return s.Doer(ctx, "", strconv.Itoa(i), func(ctx context.Context, d Doer) error {
				if err := d.DoBatch(ctx, inputs, func(j int64, output *types.Any, err error) error {
					if rand.Float64() < groupCancelProb {
						created[i] = nil
						collected[i] = nil
						cancel()
						return nil
					}
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
					collected[i][j] = true
					return nil
				}); err != nil && !errors.Is(ctx.Err(), context.Canceled) {
					return err
				}
				return nil
			})
		})
	}
	require.NoError(t, groupEg.Wait(), msg)
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

func TestCancelGroups(t *testing.T) {
	t.Parallel()
	test(t, 0, 0.05, 0)
}

func TestTaskFailures(t *testing.T) {
	t.Parallel()
	test(t, 0, 0, 0.1)
}

func TestEverything(t *testing.T) {
	t.Parallel()
	test(t, 0.1, 0.2, 0.1)
}

func TestRunZeroTasks(t *testing.T) {
	t.Parallel()
	env := testetcd.NewEnv(t)
	s := NewEtcdService(env.EtcdClient, "")
	require.NoError(t, s.Doer(context.Background(), "", "", func(ctx context.Context, d Doer) error {
		return d.DoBatch(ctx, nil, func(_ int64, _ *types.Any, _ error) error { return errors.New("no tasks should exist") })
	}))
}
