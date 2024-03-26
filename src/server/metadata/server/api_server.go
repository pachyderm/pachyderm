package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	metadatapb "github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/server/metadata"

	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
)

// Auth is the subset of the auth server needed by the metadata service.
type Env struct {
	TxnEnv *transactionenv.TransactionEnv
	Auth   metadata.Auth
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
	if err := s.env.TxnEnv.WithWriteContext(ctx, func(tc *txncontext.TransactionContext) error {
		if err := metadata.EditMetadataInTransaction(ctx, tc, s.env.Auth, req); err != nil {
			return status.Errorf(codes.FailedPrecondition, "apply edits: %v", err)
		}
		return nil
	}); err != nil {
		return res, status.Errorf(codes.Unknown, "run write txn: %v", err)
	}
	return res, nil
}
