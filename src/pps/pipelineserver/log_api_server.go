package pipelineserver

import (
	"time"

	"go.pachyderm.com/pachyderm/src/pps"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
)

type logAPIServer struct {
	protorpclog.Logger
	delegate APIServer
}

func newLogAPIServer(delegate APIServer) *logAPIServer {
	return &logAPIServer{protorpclog.NewLogger("pachyderm.pps.PipelineAPI"), delegate}
}

func (a *logAPIServer) Start() error {
	return a.delegate.Start()
}

func (a *logAPIServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *pps.Pipeline, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.CreatePipeline(ctx, request)
}

func (a *logAPIServer) GetPipeline(ctx context.Context, request *pps.GetPipelineRequest) (response *pps.Pipeline, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetPipeline(ctx, request)
}

func (a *logAPIServer) GetAllPipelines(ctx context.Context, request *google_protobuf.Empty) (response *pps.Pipelines, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.GetAllPipelines(ctx, request)
}
