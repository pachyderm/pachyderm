package server

import (
	"time"

	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
)

type logAPIServer struct {
	protorpclog.Logger
	delegate persist.APIServer
}

func newLogAPIServer(delegate persist.APIServer) *logAPIServer {
	return &logAPIServer{protorpclog.NewLogger("pachyderm.pps.persist.API"), delegate}
}

func (a *logAPIServer) CreateJob(ctx context.Context, request *persist.Job) (response *persist.Job, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.CreateJob(ctx, request)
}

func (a *logAPIServer) GetJobByID(ctx context.Context, request *google_protobuf.StringValue) (response *persist.Job, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJobByID(ctx, request)
}

func (a *logAPIServer) GetJobsByPipelineID(ctx context.Context, request *google_protobuf.StringValue) (response *persist.Jobs, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJobsByPipelineID(ctx, request)
}

func (a *logAPIServer) CreateJobStatus(ctx context.Context, request *persist.JobStatus) (response *persist.JobStatus, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.CreateJobStatus(ctx, request)
}

func (a *logAPIServer) GetJobStatusesByJobID(ctx context.Context, request *google_protobuf.StringValue) (response *persist.JobStatuses, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJobStatusesByJobID(ctx, request)
}

func (a *logAPIServer) CreateJobLog(ctx context.Context, request *persist.JobLog) (response *persist.JobLog, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.CreateJobLog(ctx, request)
}

func (a *logAPIServer) GetJobLogsByJobID(ctx context.Context, request *google_protobuf.StringValue) (response *persist.JobLogs, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJobLogsByJobID(ctx, request)
}

func (a *logAPIServer) CreatePipeline(ctx context.Context, request *persist.Pipeline) (response *persist.Pipeline, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.CreatePipeline(ctx, request)
}

func (a *logAPIServer) UpdatePipeline(ctx context.Context, request *persist.Pipeline) (response *persist.Pipeline, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.UpdatePipeline(ctx, request)
}

func (a *logAPIServer) GetPipelineByID(ctx context.Context, request *google_protobuf.StringValue) (response *persist.Pipeline, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetPipelineByID(ctx, request)
}

func (a *logAPIServer) GetPipelinesByName(ctx context.Context, request *google_protobuf.StringValue) (response *persist.Pipelines, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetPipelinesByName(ctx, request)
}

func (a *logAPIServer) GetAllPipelines(ctx context.Context, request *google_protobuf.Empty) (response *persist.Pipelines, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetAllPipelines(ctx, request)
}
