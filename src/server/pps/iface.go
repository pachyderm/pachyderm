package pps

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/context"
	pfs_client "github.com/pachyderm/pachyderm/v2/src/pfs"
	pps_client "github.com/pachyderm/pachyderm/v2/src/pps"
)

type APIServer interface {
	pps_client.APIServer

	UpdateJobStateInTransaction(*txncontext.TransactionContext, *pps_client.UpdateJobStateRequest) error
	CreatePipelineInTransaction(*txncontext.TransactionContext, *pps_client.CreatePipelineRequest, **pfs_client.Commit) error
}
