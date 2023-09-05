package server

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"golang.org/x/sync/errgroup"
)

type WorkerEnv struct {
	DB          *sqlx.DB
	ObjClient   obj.Client
	TaskService task.Service
}

type WorkerConfig struct {
	Storage pachconfig.StorageConfiguration
}

type Worker struct {
	env    WorkerEnv
	config WorkerConfig

	storage *storage.Server
}

func NewWorker(env WorkerEnv, config WorkerConfig) (*Worker, error) {
	ss, err := storage.New(storage.Env{
		ObjectStore: env.ObjClient,
		DB:          env.DB,
	}, config.Storage)
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
	// compactionWorker
	eg.Go(func() error {
		ctx := pctx.Child(ctx, "compactionWorker")
		return compactionWorker(ctx, w.env.TaskService.NewSource(StorageTaskNamespace), w.storage.Filesets)
	})
	// urlWorker
	eg.Go(func() error {
		ctx := pctx.Child(ctx, "urlWorker")
		w.URLWorker(ctx)
		return nil
	})
	return eg.Wait()
}
