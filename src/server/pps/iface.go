package pps

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pps_client "github.com/pachyderm/pachyderm/v2/src/pps"
)

// APIServer is the internal interface for other services to call this one.
// This includes all the public RPC methods and additional internal-only methods for use within pachd.
// These methods *do not* check that a user is authorized unless otherwise noted.
type APIServer interface {
	pps_client.APIServer

	NewPropagater(*txncontext.TransactionContext) txncontext.PpsPropagater
	NewJobStopper(*txncontext.TransactionContext) txncontext.PpsJobStopper
	NewJobFinisher(*txncontext.TransactionContext) txncontext.PpsJobFinisher

	StopJobInTransaction(*txncontext.TransactionContext, *pps_client.StopJobRequest) error
	UpdateJobStateInTransaction(*txncontext.TransactionContext, *pps_client.UpdateJobStateRequest) error
	CreatePipelineInTransaction(context.Context, *txncontext.TransactionContext, *pps_client.CreatePipelineRequest) error
	// InspectPipelineInTransaction returns the pipeline information for a
	// pipeline.  Note that the pipeline name may include ancestry syntax.
	InspectPipelineInTransaction(*txncontext.TransactionContext, *pps.Pipeline) (*pps_client.PipelineInfo, error)
	ActivateAuthInTransaction(*txncontext.TransactionContext, *pps_client.ActivateAuthRequest) (*pps_client.ActivateAuthResponse, error)
}
