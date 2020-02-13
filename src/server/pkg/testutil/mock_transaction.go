package testutil

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pps"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
)

// This code can all go away if we ever get the ability to run a PPS server without external dependencies
type updateJobStateInTransactionFunc func(*txnenv.TransactionContext, *pps.UpdateJobStateRequest) error

type mockUpdateJobStateInTransaction struct {
	handler updateJobStateInTransactionFunc
}

func (mock *mockUpdateJobStateInTransaction) Use(cb updateJobStateInTransactionFunc) {
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
}

func (api *ppsTransactionAPI) UpdateJobStateInTransaction(txnCtx *txnenv.TransactionContext, req *pps.UpdateJobStateRequest) error {
	if api.mock.UpdateJobStateInTransaction.handler != nil {
		return api.mock.UpdateJobStateInTransaction.handler(txnCtx, req)
	}
	return fmt.Errorf("unhandled pachd mock: pps.UpdateJobStateInTransaction")
}

// NewMockPPSTransactionServer instantiates a MockPPSTransactionServer
func NewMockPPSTransactionServer() *MockPPSTransactionServer {
	result := &MockPPSTransactionServer{}
	result.api.mock = result
	return result
}
