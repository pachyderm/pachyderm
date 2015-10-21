package jobserver

import (
	"time"

	"go.pachyderm.com/pachyderm/src/pps"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
)

type logAPIServer struct {
	protorpclog.Logger
	delegate pps.JobAPIServer
}

func newLogAPIServer(delegate pps.JobAPIServer) *logAPIServer {
	return &logAPIServer{protorpclog.NewLogger("pachyderm.pps.JobAPI"), delegate}
}

func (a *logAPIServer) CreateJob(ctx context.Context, request *pps.CreateJobRequest) (response *pps.Job, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.CreateJob(ctx, request)
}

func (a *logAPIServer) GetJob(ctx context.Context, request *pps.GetJobRequest) (response *pps.Job, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJob(ctx, request)
}

func (a *logAPIServer) GetJobsByPipelineName(ctx context.Context, request *pps.GetJobsByPipelineNameRequest) (response *pps.Jobs, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJobsByPipelineName(ctx, request)
}

func (a *logAPIServer) StartJob(ctx context.Context, request *pps.StartJobRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.StartJob(ctx, request)
}

func (a *logAPIServer) GetJobStatus(ctx context.Context, request *pps.GetJobStatusRequest) (response *pps.JobStatus, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetJobStatus(ctx, request)
}

func (a *logAPIServer) GetJobLogs(request *pps.GetJobLogsRequest, responseServer pps.JobAPI_GetJobLogsServer) (err error) {
	// TODO(pedge): log
	return a.delegate.GetJobLogs(request, responseServer)
}
