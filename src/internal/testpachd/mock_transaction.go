package testpachd

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// This code can all go away if we ever get the ability to run a PPS server without external dependencies
type stopJobInTransactionFunc func(*txncontext.TransactionContext, *pps.StopJobRequest) error

type mockStopJobInTransaction struct {
	handler stopJobInTransactionFunc
}

func (mock *mockStopJobInTransaction) Use(cb stopJobInTransactionFunc) {
	mock.handler = cb
}

type updateJobStateInTransactionFunc func(*txncontext.TransactionContext, *pps.UpdateJobStateRequest) error

type mockUpdateJobStateInTransaction struct {
	handler updateJobStateInTransactionFunc
}

func (mock *mockUpdateJobStateInTransaction) Use(cb updateJobStateInTransactionFunc) {
	mock.handler = cb
}

type createPipelineInTransactionFunc func(*txncontext.TransactionContext, *pps.CreatePipelineRequest, *string, **pfs.Commit) error

type mockCreatePipelineInTransaction struct {
	handler createPipelineInTransactionFunc
}

func (mock *mockCreatePipelineInTransaction) Use(cb createPipelineInTransactionFunc) {
	mock.handler = cb
}

type ppsTransactionAPI struct {
	ppsServerAPI
	mock *MockPPSTransactionServer
}

// MockPPSTransactionServer provides a mocking interface for overriding PPS
// behavior inside transactions.
type MockPPSTransactionServer struct {
	api                         ppsTransactionAPI
	StopJobInTransaction        mockStopJobInTransaction
	UpdateJobStateInTransaction mockUpdateJobStateInTransaction
	CreatePipelineInTransaction mockCreatePipelineInTransaction
}

func (api *ppsTransactionAPI) StopJobInTransaction(txnCtx *txncontext.TransactionContext, req *pps.StopJobRequest) error {
	if api.mock.StopJobInTransaction.handler != nil {
		return api.mock.StopJobInTransaction.handler(txnCtx, req)
	}
	return fmt.Errorf("unhandled pachd mock: pps.StopJobInTransaction")
}

func (api *ppsTransactionAPI) UpdateJobStateInTransaction(txnCtx *txncontext.TransactionContext, req *pps.UpdateJobStateRequest) error {
	if api.mock.UpdateJobStateInTransaction.handler != nil {
		return api.mock.UpdateJobStateInTransaction.handler(txnCtx, req)
	}
	return fmt.Errorf("unhandled pachd mock: pps.UpdateJobStateInTransaction")
}

func (api *ppsTransactionAPI) CreatePipelineInTransaction(txnCtx *txncontext.TransactionContext, req *pps.CreatePipelineRequest, filesetID *string, prevSpecCommit **pfs.Commit) error {
	if api.mock.CreatePipelineInTransaction.handler != nil {
		return api.mock.CreatePipelineInTransaction.handler(txnCtx, req, filesetID, prevSpecCommit)
	}
	return fmt.Errorf("unhandled pachd mock: pps.CreatePipelineInTransaction")
}

// NewMockPPSTransactionServer instantiates a MockPPSTransactionServer
func NewMockPPSTransactionServer() *MockPPSTransactionServer {
	result := &MockPPSTransactionServer{}
	result.api.mock = result
	return result
}
