package testpachd

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/transactionenv/txncontext"
)

// This code can all go away if we ever get the ability to run a PPS server without external dependencies
type updateJobStateInTransactionFunc func(*txncontext.TransactionContext, *pps.UpdateJobStateRequest) error

type mockUpdateJobStateInTransaction struct {
	handler updateJobStateInTransactionFunc
}

func (mock *mockUpdateJobStateInTransaction) Use(cb updateJobStateInTransactionFunc) {
	mock.handler = cb
}

// This code can all go away if we ever get the ability to run a PPS server without external dependencies
type createPipelineInTransactionFunc func(*txncontext.TransactionContext, *pps.CreatePipelineRequest, **pfs.Commit) error

type mockCreatePipelineInTransaction struct {
	handler createPipelineInTransactionFunc
}

func (mock *mockCreatePipelineInTransaction) Use(cb createPipelineInTransactionFunc) {
	mock.handler = cb
}

// This code can all go away if we ever get the ability to run a PPS server without external dependencies
type listPipelineNoAuthFunc func(context.Context, *pps.ListPipelineRequest) (*pps.PipelineInfos, error)

type mockListPipelineNoAuth struct {
	handler listPipelineNoAuthFunc
}

func (mock *mockListPipelineNoAuth) Use(cb listPipelineNoAuthFunc) {
	mock.handler = cb
}

type ppsTransactionAPI struct {
	mock *MockPPSTransactionServer
}

// MockPPSTransactionServer provides a mocking interface for overriding PPS
// behavior inside transactions.
type MockPPSTransactionServer struct {
	api                         ppsTransactionAPI
	UpdateJobStateInTransaction mockUpdateJobStateInTransaction
	CreatePipelineInTransaction mockCreatePipelineInTransaction
	ListPipelineNoAuth          mockListPipelineNoAuth
}

func (api *ppsTransactionAPI) UpdateJobStateInTransaction(txnCtx *txncontext.TransactionContext, req *pps.UpdateJobStateRequest) error {
	if api.mock.UpdateJobStateInTransaction.handler != nil {
		return api.mock.UpdateJobStateInTransaction.handler(txnCtx, req)
	}
	return fmt.Errorf("unhandled pachd mock: pps.UpdateJobStateInTransaction")
}

func (api *ppsTransactionAPI) CreatePipelineInTransaction(txnCtx *txncontext.TransactionContext, req *pps.CreatePipelineRequest, specCommit **pfs.Commit) error {
	if api.mock.UpdateJobStateInTransaction.handler != nil {
		return api.mock.CreatePipelineInTransaction.handler(txnCtx, req, specCommit)
	}
	return fmt.Errorf("unhandled pachd mock: pps.CreatePipelineInTransaction")
}

func (api *ppsTransactionAPI) ListPipelineNoAuth(ctx context.Context, req *pps.ListPipelineRequest) (*pps.PipelineInfos, error) {
	if api.mock.ListPipelineNoAuth.handler != nil {
		return api.mock.ListPipelineNoAuth.handler(ctx, req)
	}
	return nil, fmt.Errorf("unhandled pachd mock: pps.ListPipelineNoAuth")
}

// NewMockPPSTransactionServer instantiates a MockPPSTransactionServer
func NewMockPPSTransactionServer() *MockPPSTransactionServer {
	result := &MockPPSTransactionServer{}
	result.api.mock = result
	return result
}
