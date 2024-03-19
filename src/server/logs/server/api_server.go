package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	logservice "github.com/pachyderm/pachyderm/v2/src/server/logs"
)

type APIServer = *apiServer

type Env struct {
	GetLokiClient func() (*loki.Client, error)
}

type apiServer struct {
	logs.UnsafeAPIServer
	env     Env
	service logservice.LogService
}

func NewAPIServer(env Env) (*apiServer, error) {
	return &apiServer{
		env: env,
		service: logservice.LogService{
			GetLokiClient: env.GetLokiClient,
		},
	}, nil
}

type getLogsServerPublisher struct {
	server logs.API_GetLogsServer
}

func (glsp getLogsServerPublisher) Publish(ctx context.Context, response *logs.GetLogsResponse) error {
	return glsp.server.Send(response)
}

func (l *apiServer) GetLogs(request *logs.GetLogsRequest, apiGetLogsServer logs.API_GetLogsServer) error {
	if err := l.service.GetLogs(apiGetLogsServer.Context(), request, getLogsServerPublisher{apiGetLogsServer}); err != nil {
		if errors.Is(err, logservice.ErrUnimplemented) {
			return status.Error(codes.Unimplemented, err.Error())
		}
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}
