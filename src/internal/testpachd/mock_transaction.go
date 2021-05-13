package testpachd

import (
	"fmt"

	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// This code can all go away if we ever get the ability to run a PPS server without external dependencies
type stopJobInTransactionFunc func(*txnenv.TransactionContext, *pps.StopJobRequest) error

type mockStopJobInTransaction struct {
	handler stopJobInTransactionFunc
}

func (mock *mockStopJobInTransaction) Use(cb stopJobInTransactionFunc) {
	mock.handler = cb
}

type updateJobStateInTransactionFunc func(*txnenv.TransactionContext, *pps.UpdateJobStateRequest) error

type mockUpdateJobStateInTransaction struct {
	handler updateJobStateInTransactionFunc
}

func (mock *mockUpdateJobStateInTransaction) Use(cb updateJobStateInTransactionFunc) {
	mock.handler = cb
}

type createPipelineInTransactionFunc func(*txnenv.TransactionContext, *pps.CreatePipelineRequest, *string, **pfs.Commit) error

type mockCreatePipelineInTransaction struct {
	handler createPipelineInTransactionFunc
}

func (mock *mockCreatePipelineInTransaction) Use(cb createPipelineInTransactionFunc) {
	mock.handler = cb
}

type ppsTransactionAPI struct {
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

func (api *ppsTransactionAPI) StopJobInTransaction(txnCtx *txnenv.TransactionContext, req *pps.StopJobRequest) error {
	if api.mock.StopJobInTransaction.handler != nil {
		return api.mock.StopJobInTransaction.handler(txnCtx, req)
	}
	return fmt.Errorf("unhandled pachd mock: pps.StopJobInTransaction")
}

func (api *ppsTransactionAPI) UpdateJobStateInTransaction(txnCtx *txnenv.TransactionContext, req *pps.UpdateJobStateRequest) error {
	if api.mock.UpdateJobStateInTransaction.handler != nil {
		return api.mock.UpdateJobStateInTransaction.handler(txnCtx, req)
	}
	return fmt.Errorf("unhandled pachd mock: pps.UpdateJobStateInTransaction")
}

func (api *ppsTransactionAPI) CreatePipelineInTransaction(txnCtx *txnenv.TransactionContext, req *pps.CreatePipelineRequest, filesetID *string, prevSpecCommit **pfs.Commit) error {
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
