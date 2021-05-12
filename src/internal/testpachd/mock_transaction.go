package testpachd

import (
	"fmt"

	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// This code can all go away if we ever get the ability to run a PPS server without external dependencies
type updateJobStateInTransactionFunc func(*txnenv.TransactionContext, *pps.UpdateJobStateRequest) error

type mockUpdateJobStateInTransaction struct {
	handler updateJobStateInTransactionFunc
}

func (mock *mockUpdateJobStateInTransaction) Use(cb updateJobStateInTransactionFunc) {
	mock.handler = cb
}

type createPipelineInTransactionFunc func(*txnenv.TransactionContext, *pps.CreatePipelineRequest, string, *pfs.Commit) error

type mockCreatePipelineInTransaction struct {
	handler createPipelineInTransactionFunc
}

func (mock *mockCreatePipelineInTransaction) Use(cb createPipelineInTransactionFunc) {
	mock.handler = cb
}

type preparePipelineSpecFilesetFunc func(*txnenv.TransactionContext, *pps.CreatePipelineRequest) (string, *pfs.Commit, error)

type mockPreparePipelineSpecFileset struct {
	handler preparePipelineSpecFilesetFunc
}

func (mock *mockPreparePipelineSpecFileset) Use(cb preparePipelineSpecFilesetFunc) {
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
	PreparePipelineSpecFileset  mockPreparePipelineSpecFileset
}

func (api *ppsTransactionAPI) UpdateJobStateInTransaction(txnCtx *txnenv.TransactionContext, req *pps.UpdateJobStateRequest) error {
	if api.mock.UpdateJobStateInTransaction.handler != nil {
		return api.mock.UpdateJobStateInTransaction.handler(txnCtx, req)
	}
	return fmt.Errorf("unhandled pachd mock: pps.UpdateJobStateInTransaction")
}

func (api *ppsTransactionAPI) CreatePipelineInTransaction(txnCtx *txnenv.TransactionContext, req *pps.CreatePipelineRequest, filesetID string, prevSpecCommit *pfs.Commit) error {
	if api.mock.CreatePipelineInTransaction.handler != nil {
		return api.mock.CreatePipelineInTransaction.handler(txnCtx, req, filesetID, prevSpecCommit)
	}
	return fmt.Errorf("unhandled pachd mock: pps.CreatePipelineInTransaction")
}

func (api *ppsTransactionAPI) PreparePipelineSpecFileset(txnCtx *txnenv.TransactionContext, req *pps.CreatePipelineRequest) (string, *pfs.Commit, error) {
	if api.mock.PreparePipelineSpecFileset.handler != nil {
		return api.mock.PreparePipelineSpecFileset.handler(txnCtx, req)
	}
	return "", nil, fmt.Errorf("unhandled pachd mock: pps.CreatePipelineInTransaction")
}

// NewMockPPSTransactionServer instantiates a MockPPSTransactionServer
func NewMockPPSTransactionServer() *MockPPSTransactionServer {
	result := &MockPPSTransactionServer{}
	result.api.mock = result
	return result
}
