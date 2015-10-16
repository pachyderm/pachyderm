package server

import (
	"time"

	"go.pachyderm.com/pachyderm/src/pps/watch"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
)

type logAPIServer struct {
	protorpclog.Logger
	delegate watch.APIServer
}

func newLogAPIServer(delegate watch.APIServer) *logAPIServer {
	return &logAPIServer{protorpclog.NewLogger("pachyderm.pps.watch.API"), delegate}
}

func (a *logAPIServer) Start(ctx context.Context, request *google_protobuf.Empty) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.Start(ctx, request)
}

func (a *logAPIServer) RegisterChangeEvent(ctx context.Context, request *watch.ChangeEvent) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	return a.delegate.RegisterChangeEvent(ctx, request)
}
