package pps

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	pfs_client "github.com/pachyderm/pachyderm/v2/src/pfs"
	pps_client "github.com/pachyderm/pachyderm/v2/src/pps"
)

// APIServer is the internal interface for other services to call this one.
// This includes all the public RPC methods and additional internal-only methods for use within pachd.
// These methods *do not* check that a user is authorized unless otherwise noted.
type APIServer interface {
	pps_client.APIServer

	NewPropagater(*txncontext.TransactionContext) txncontext.PpsPropagater

	StopPipelineJobInTransaction(*txncontext.TransactionContext, *pps_client.StopPipelineJobRequest) error
	UpdatePipelineJobStateInTransaction(*txncontext.TransactionContext, *pps_client.UpdatePipelineJobStateRequest) error
	CreatePipelineInTransaction(*txncontext.TransactionContext, *pps_client.CreatePipelineRequest, *string, **pfs_client.Commit) error
}
