package pps

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// APIServer is the internal interface for other services to call this one.
// This includes all the public RPC methods and additional internal-only methods for use within pachd.
// These methods *do not* check that a user is authorized unless otherwise noted.
type APIServer interface {
	pps.APIServer

	NewPropagater(*txncontext.TransactionContext) txncontext.PpsPropagater
	NewJobStopper(*txncontext.TransactionContext) txncontext.PpsJobStopper
	NewJobFinisher(*txncontext.TransactionContext) txncontext.PpsJobFinisher

	StopJobInTransaction(context.Context, *txncontext.TransactionContext, *pps.StopJobRequest) error
	UpdateJobStateInTransaction(context.Context, *txncontext.TransactionContext, *pps.UpdateJobStateRequest) error
	CreatePipelineInTransaction(context.Context, *txncontext.TransactionContext, *pps.CreatePipelineTransaction) error
	// InspectPipelineInTransaction returns the pipeline information for a
	// pipeline.  Note that the pipeline name may include ancestry syntax.
	InspectPipelineInTransaction(context.Context, *txncontext.TransactionContext, *pps.Pipeline) (*pps.PipelineInfo, error)
	ActivateAuthInTransaction(context.Context, *txncontext.TransactionContext, *pps.ActivateAuthRequest) (*pps.ActivateAuthResponse, error)
	CreateDetPipelineSideEffects(context.Context, *pps.Pipeline, []string) error
}
