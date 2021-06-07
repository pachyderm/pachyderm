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

// InternalAPI is the set of non-RPC methods that are used internally within pachd
type InternalAPI interface {
	TransactionServer

	ListPipelineNoAuth(context.Context, *pps_client.ListPipelineRequest) (*pps_client.PipelineInfos, error)
}

// APIServer is the set of all methods exposed by PPS
type APIServer interface {
	pps_client.APIServer
	InternalAPI
}
