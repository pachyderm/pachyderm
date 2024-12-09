// Package server implements the identity service gRPC server.
package server

import (
	"context"
	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type apiServer struct{}

func (a *apiServer) Heartbeat(ctx context.Context, req *ec.HeartbeatRequest) (resp *ec.HeartbeatResponse, retErr error) {
	return &ec.HeartbeatResponse{}, errors.New("enterprise server is deprecated")
}

func (a *apiServer) Activate(ctx context.Context, req *ec.ActivateRequest) (resp *ec.ActivateResponse, retErr error) {
	return &ec.ActivateResponse{}, errors.New("enterprise server is deprecated")
}

func (a *apiServer) GetState(ctx context.Context, req *ec.GetStateRequest) (resp *ec.GetStateResponse, retErr error) {
	return &ec.GetStateResponse{}, errors.New("enterprise server is deprecated")
}

func (a *apiServer) GetActivationCode(ctx context.Context, req *ec.GetActivationCodeRequest) (resp *ec.GetActivationCodeResponse, retErr error) {
	return &ec.GetActivationCodeResponse{}, errors.New("enterprise server is deprecated")
}

func (a *apiServer) getEnterpriseRecord() (*ec.GetActivationCodeResponse, error) {
	return &ec.GetActivationCodeResponse{}, errors.New("enterprise server is deprecated")
}

func (a *apiServer) Deactivate(ctx context.Context, req *ec.DeactivateRequest) (resp *ec.DeactivateResponse, retErr error) {
	return &ec.DeactivateResponse{}, errors.New("enterprise server is deprecated")
}

func (a *apiServer) Pause(ctx context.Context, req *ec.PauseRequest) (resp *ec.PauseResponse, retErr error) {
	return &ec.PauseResponse{}, errors.New("enterprise server is deprecated")
}

func (a *apiServer) rollPachd(ctx context.Context, paused bool) error {
	return errors.New("enterprise server is deprecated")
}

func (a *apiServer) Unpause(ctx context.Context, req *ec.UnpauseRequest) (resp *ec.UnpauseResponse, retErr error) {
	return &ec.UnpauseResponse{}, errors.New("enterprise server is deprecated")
}

func (a *apiServer) PauseStatus(ctx context.Context, req *ec.PauseStatusRequest) (resp *ec.PauseStatusResponse, retErr error) {
	return &ec.PauseStatusResponse{}, errors.New("enterprise server is deprecated")
}
