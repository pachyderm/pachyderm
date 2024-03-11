package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/logs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	logservice "github.com/pachyderm/pachyderm/v2/src/server/logs"
)

type APIServer = *apiServer

type apiServer struct {
	logs.UnsafeAPIServer
	service logservice.LogService
}

func NewAPIServer() (*apiServer, error) {
	return &apiServer{}, nil
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
