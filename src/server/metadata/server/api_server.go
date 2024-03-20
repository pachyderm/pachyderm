package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	metadatapb "github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/server/metadata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Env struct {
	DB *pachsql.DB
}

type APIServer struct {
	metadatapb.UnsafeAPIServer
	env Env
}

var _ metadatapb.APIServer = (*APIServer)(nil)

func NewMetadataServer(env Env) *APIServer {
	return &APIServer{
		env: env,
	}
}

// EditMetadata transactionally mutates metadata.  All operations are attempted, in order, but if
// any fail, the entire operation fails.
func (s *APIServer) EditMetadata(ctx context.Context, req *metadatapb.EditMetadataRequest) (*metadatapb.EditMetadataResponse, error) {
	res := &metadatapb.EditMetadataResponse{}
	if err := metadata.EditMetadata(ctx, s.env.DB, req); err != nil {
		return res, status.Errorf(codes.FailedPrecondition, "apply edits: %v", err)
	}
	return res, nil
}
