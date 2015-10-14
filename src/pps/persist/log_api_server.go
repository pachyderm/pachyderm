package persist

import (
	"time"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
)

type logAPIServer struct {
	protorpclog.Logger
	delegate APIServer
}

func newLogAPIServer(delegate APIServer) *logAPIServer {
	return &logAPIServer{protorpclog.NewLogger("pachyderm.pps.persist.API"), delegate}
}

func (a *logAPIServer) CreateJob(ctx context.Context, request *Job) (response *Job, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.CreateJob(ctx, request)
}

func (a *logAPIServer) GetJobByID(ctx context.Context, request *google_protobuf.StringValue) (response *Job, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJobByID(ctx, request)
}

func (a *logAPIServer) GetJobsByPipelineID(ctx context.Context, request *google_protobuf.StringValue) (response *Jobs, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJobsByPipelineID(ctx, request)
}

func (a *logAPIServer) CreateJobStatus(ctx context.Context, request *JobStatus) (response *JobStatus, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.CreateJobStatus(ctx, request)
}

func (a *logAPIServer) GetJobStatusesByJobID(ctx context.Context, request *google_protobuf.StringValue) (response *JobStatuses, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJobStatusesByJobID(ctx, request)
}

func (a *logAPIServer) CreateJobLog(ctx context.Context, request *JobLog) (response *JobLog, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.CreateJobLog(ctx, request)
}

func (a *logAPIServer) GetJobLogsByJobID(ctx context.Context, request *google_protobuf.StringValue) (response *JobLogs, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJobLogsByJobID(ctx, request)
}

func (a *logAPIServer) CreatePipeline(ctx context.Context, request *Pipeline) (response *Pipeline, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.CreatePipeline(ctx, request)
}

func (a *logAPIServer) UpdatePipeline(ctx context.Context, request *Pipeline) (response *Pipeline, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.UpdatePipeline(ctx, request)
}

func (a *logAPIServer) GetPipelineByID(ctx context.Context, request *google_protobuf.StringValue) (response *Pipeline, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetPipelineByID(ctx, request)
}

func (a *logAPIServer) GetPipelinesByName(ctx context.Context, request *google_protobuf.StringValue) (response *Pipelines, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetPipelinesByName(ctx, request)
}
