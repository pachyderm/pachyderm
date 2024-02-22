package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

type APIServer = *apiServer

type apiServer struct {
	logs.UnsafeAPIServer
}

func NewAPIServer() (*apiServer, error) {
	return &apiServer{}, nil
}

func (l *apiServer) GetLogs(request *logs.GetLogsRequest, apiGetLogsServer logs.API_GetLogsServer) (retErr error) {
	msg := new(pps.LogMessage)
	msg.Message = "GetLogs dummy response"

	resp := &logs.GetLogsResponse{ResponseType: &logs.GetLogsResponse_Log{Log: &logs.LogMessage{LogType: &logs.LogMessage_PpsLogMessage{PpsLogMessage: msg}}}}

	if err := apiGetLogsServer.Send(resp); err != nil {
		return errors.EnsureStack(err)
	}

	return nil
}
