package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/metadata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Env struct {
	DB *pachsql.DB
}

type APIServer struct {
	metadata.UnsafeAPIServer
	env Env
}

var _ metadata.APIServer = (*APIServer)(nil)

func NewMetadataServer(env Env) *APIServer {
	return &APIServer{
		env: env,
	}
}

func (s *APIServer) EditMetadata(ctx context.Context, req *metadata.EditMetadataRequest) (*metadata.EditMetadataResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
