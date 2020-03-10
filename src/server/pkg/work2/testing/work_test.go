package testing

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/testetcd"
	"github.com/pachyderm/pachyderm/src/server/pkg/work2"

	types "github.com/gogo/protobuf/types"
	"golang.org/x/sync/errgroup"
)

func TestNoSubtasks(t *testing.T) {
	err := testetcd.WithEtcdEnv(func(env *testetcd.EtcdEnv) error {
		eg, ctx := errgroup.WithContext(env.Context)

		eg.Go(func() error {
			worker := work2.NewWorker(env.EtcdClient, "", "")
			return worker.Run(ctx, func(ctx context.Context, data *types.Any) error {
				require.True(t, false, "no subtasks were made, work should not happen")
				return nil
			})
		})

		eg.Go(func() error {
			master := work2.NewMaster(env.EtcdClient, "", "")
			return master.WithTask(func(task work2.Task) error {
				subtaskChan := make(chan *types.Any)
				close(subtaskChan)

				return task.Run(ctx, subtaskChan, func(ctx context.Context, data *types.Any) error {
					require.True(t, false, "no subtasks were made, collection should not happen")
					return nil
				})
			})
		})

		return eg.Wait()
	})
	require.NoError(t, err)
}
