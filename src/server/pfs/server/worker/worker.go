package worker

import (
	"context"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
)

type Env struct {
	DB          *sqlx.DB
	Bucket      *obj.Bucket
	TaskService task.Service
}

type Config struct {
	Storage pachconfig.StorageConfiguration
}

type Worker struct {
	env    Env
	config Config

	storage *storage.Server
}

func NewWorker(ctx context.Context, env Env, config Config) (*Worker, error) {
	ss, err := storage.New(ctx, storage.Env{
		DB:     env.DB,
		Bucket: env.Bucket,
		Config: config.Storage,
	})
	if err != nil {
		return nil, err
	}
	return &Worker{
		env:    env,
		config: config,

		storage: ss,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	log.Info(ctx, "started worker")
	defer log.Info(ctx, "exited worker")
	eg.Go(func() error {
		ctx := pctx.Child(ctx, "compactionWorker")
		return server.CompactionWorker(ctx, w.env.TaskService.NewSource(server.StorageTaskNamespace), w.storage.Filesets)
	})
	eg.Go(func() error {
		ctx := pctx.Child(ctx, "urlWorker")
		return w.URLWorker(ctx)
	})
	return errors.EnsureStack(eg.Wait())
}
