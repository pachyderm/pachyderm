package testpachd

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/context"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// This code can all go away if we ever get the ability to run a PPS server without external dependencies
type stopPipelineJobInTransactionFunc func(*txncontext.TransactionContext, *pps.StopPipelineJobRequest) error

type mockStopPipelineJobInTransaction struct {
	handler stopPipelineJobInTransactionFunc
}

func (mock *mockStopPipelineJobInTransaction) Use(cb stopPipelineJobInTransactionFunc) {
	mock.handler = cb
}

type updatePipelineJobStateInTransactionFunc func(*txncontext.TransactionContext, *pps.UpdatePipelineJobStateRequest) error

type mockUpdatePipelineJobStateInTransaction struct {
	handler updatePipelineJobStateInTransactionFunc
}

func (mock *mockUpdatePipelineJobStateInTransaction) Use(cb updatePipelineJobStateInTransactionFunc) {
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
	api                                 ppsTransactionAPI
	StopPipelineJobInTransaction        mockStopPipelineJobInTransaction
	UpdatePipelineJobStateInTransaction mockUpdatePipelineJobStateInTransaction
	CreatePipelineInTransaction         mockCreatePipelineInTransaction
}

func (api *ppsTransactionAPI) StopPipelineJobInTransaction(txnCtx *txncontext.TransactionContext, req *pps.StopPipelineJobRequest) error {
	if api.mock.StopPipelineJobInTransaction.handler != nil {
		return api.mock.StopPipelineJobInTransaction.handler(txnCtx, req)
	}
	return fmt.Errorf("unhandled pachd mock: pps.StopPipelineJobInTransaction")
}

func (api *ppsTransactionAPI) UpdatePipelineJobStateInTransaction(txnCtx *txncontext.TransactionContext, req *pps.UpdatePipelineJobStateRequest) error {
	if api.mock.UpdatePipelineJobStateInTransaction.handler != nil {
		return api.mock.UpdatePipelineJobStateInTransaction.handler(txnCtx, req)
	}
	return fmt.Errorf("unhandled pachd mock: pps.UpdatePipelineJobStateInTransaction")
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
