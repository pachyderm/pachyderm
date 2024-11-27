// Package server implements the license service gRPC server.
package server

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	lc "github.com/pachyderm/pachyderm/v2/src/license"
)

type apiServer struct{}

func (a *apiServer) Activate(ctx context.Context, req *lc.ActivateRequest) (*lc.ActivateResponse, error) {
	return &lc.ActivateResponse{}, errors.New("license server is deprecated")
}

func (a *apiServer) GetActivationCode(ctx context.Context, req *lc.GetActivationCodeRequest) (resp *lc.GetActivationCodeResponse, retErr error) {
	return &lc.GetActivationCodeResponse{}, errors.New("license server is deprecated")
}

func (a *apiServer) AddCluster(ctx context.Context, req *lc.AddClusterRequest) (resp *lc.AddClusterResponse, retErr error) {
	return &lc.AddClusterResponse{}, errors.New("license server is deprecated")
}

func (a *apiServer) Heartbeat(ctx context.Context, req *lc.HeartbeatRequest) (resp *lc.HeartbeatResponse, retErr error) {
	return &lc.HeartbeatResponse{}, errors.New("license server is deprecated")
}

func (a *apiServer) DeleteAll(ctx context.Context, req *lc.DeleteAllRequest) (resp *lc.DeleteAllResponse, retErr error) {
	return &lc.DeleteAllResponse{}, errors.New("license server is deprecated")
}

func (a *apiServer) ListClusters(ctx context.Context, req *lc.ListClustersRequest) (resp *lc.ListClustersResponse, retErr error) {
	return &lc.ListClustersResponse{}, errors.New("license server is deprecated")
}

func (a *apiServer) DeleteCluster(ctx context.Context, req *lc.DeleteClusterRequest) (resp *lc.DeleteClusterResponse, retErr error) {
	return &lc.DeleteClusterResponse{}, errors.New("license server is deprecated")
}

func (a *apiServer) UpdateCluster(ctx context.Context, req *lc.UpdateClusterRequest) (resp *lc.UpdateClusterResponse, retErr error) {
	return &lc.UpdateClusterResponse{}, errors.New("license server is deprecated")
}

func (a *apiServer) ListUserClusters(ctx context.Context, req *lc.ListUserClustersRequest) (resp *lc.ListUserClustersResponse, retErr error) {
	return &lc.ListUserClustersResponse{}, nil
}
