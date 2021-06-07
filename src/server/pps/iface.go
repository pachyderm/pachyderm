package pps

import (
	"context"

	pfs_client "github.com/pachyderm/pachyderm/src/client/pfs"
	pps_client "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/transactionenv/txncontext"
)

// TransactionServer is an interface for the transactionally-supported
// methods that can be called through the PPS server.
type TransactionServer interface {
	UpdateJobStateInTransaction(*txncontext.TransactionContext, *pps_client.UpdateJobStateRequest) error
	CreatePipelineInTransaction(*txncontext.TransactionContext, *pps_client.CreatePipelineRequest, **pfs_client.Commit) error
}

type APIServer interface {
	pps_client.APIServer
	TransactionServer

	ListPipelineNoAuth(context.Context, *pps_client.ListPipelineRequest) (*pps_client.PipelineInfos, error)
}
