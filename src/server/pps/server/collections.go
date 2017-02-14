package server

import (
	"context"
	"path"

	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

const (
	pipelinesPrefix     = "/pipelines"
	jobsRunningPrefix   = "/jobs/running"
	jobsCompletedPrefix = "/jobs/completed"
)

func (a *apiServer) pipelines(stm col.STM) *col.Collection {
	return col.NewCollection(
		a.etcdClient,
		path.Join(a.etcdPrefix, pipelinesPrefix),
		stm,
	)
}

func (a *apiServer) jobsRunning(stm col.STM) *col.Collection {
	return col.NewCollection(
		a.etcdClient,
		path.Join(a.etcdPrefix, jobsRunningPrefix),
		stm,
	)
}

func (a *apiServer) jobsCompleted(stm col.STM) *col.Collection {
	return col.NewCollection(
		a.etcdClient,
		path.Join(a.etcdPrefix, jobsCompletedPrefix),
		stm,
	)
}

func (a *apiServer) pipelinesReadonly(ctx context.Context) *col.ReadonlyCollection {
	return col.NewReadonlyCollection(
		ctx,
		a.etcdClient,
		path.Join(a.etcdPrefix, pipelinesPrefix),
	)
}

func (a *apiServer) jobRunningReadonly(ctx context.Context) *col.ReadonlyCollection {
	return col.NewReadonlyCollection(
		ctx,
		a.etcdClient,
		path.Join(a.etcdPrefix, jobsRunningPrefix),
	)
}

func (a *apiServer) jobCompletedReadonly(ctx context.Context) *col.ReadonlyCollection {
	return col.NewReadonlyCollection(
		ctx,
		a.etcdClient,
		path.Join(a.etcdPrefix, jobsCompletedPrefix),
	)
}
