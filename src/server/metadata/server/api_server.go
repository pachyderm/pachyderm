// Package server implements the metadata service gRPC server.
package server

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	metadatapb "github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/server/metadata"

	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
)

type Env struct {
	TxnEnv *transactionenv.TransactionEnv
	// Auth is the subset of the auth server needed by the metadata service.
	Auth metadata.Auth
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
	if err := s.env.TxnEnv.WithWriteContext(ctx, func(ctx context.Context, tc *txncontext.TransactionContext) error {
		if err := metadata.EditMetadataInTransaction(ctx, tc, s.env.Auth, req); err != nil {
			return err
		}
		return nil
	}); err != nil {
		switch {
		case strings.Contains(err.Error(), "SQLSTATE"):
			return res, status.Errorf(codes.Aborted, "apply edits: %v", err)
		case strings.Contains(err.Error(), "is not authorized"):
			return res, status.Errorf(codes.PermissionDenied, "apply edits: %v", err)
		case strings.Contains(err.Error(), "already exists"):
			return res, status.Errorf(codes.FailedPrecondition, "apply edits: %v", err)
		default:
			return res, status.Errorf(codes.Unknown, "apply edits: %v", err)
		}
	}
	return res, nil
}
