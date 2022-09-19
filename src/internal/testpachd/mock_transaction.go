package testpachd

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

type newPropagaterFunc func(*txncontext.TransactionContext) txncontext.PpsPropagater

type mockNewPropagater struct {
	handler newPropagaterFunc
}

func (mock *mockNewPropagater) Use(cb newPropagaterFunc) {
	mock.handler = cb
}

type newJobStopperFunc func(*txncontext.TransactionContext) txncontext.PpsJobStopper

type mockNewJobStopper struct {
	handler newJobStopperFunc
}

func (mock *mockNewJobStopper) Use(cb newJobStopperFunc) {
	mock.handler = cb
}

type newJobFinisherFunc func(*txncontext.TransactionContext) txncontext.PpsJobFinisher

type mockNewJobFinisher struct {
	handler newJobFinisherFunc
}

func (mock *mockNewJobFinisher) Use(cb newJobFinisherFunc) {
	mock.handler = cb
}

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

type createPipelineInTransactionFunc func(*txncontext.TransactionContext, *pps.CreatePipelineRequest) error

type mockCreatePipelineInTransaction struct {
	handler createPipelineInTransactionFunc
}

func (mock *mockCreatePipelineInTransaction) Use(cb createPipelineInTransactionFunc) {
	mock.handler = cb
}

type inspectPipelineInTransactionFunc func(*txncontext.TransactionContext, string) (*pps.PipelineInfo, error)

type mockInspectPipelineInTransaction struct {
	handler inspectPipelineInTransactionFunc
}

func (mock *mockInspectPipelineInTransaction) Use(cb inspectPipelineInTransactionFunc) {
	mock.handler = cb
}

type activateAuthInTransactionFunc func(*txncontext.TransactionContext, *pps.ActivateAuthRequest) (*pps.ActivateAuthResponse, error)

type mockActivateAuthInTransaction struct {
	handler activateAuthInTransactionFunc
}

func (mock *mockActivateAuthInTransaction) Use(cb activateAuthInTransactionFunc) {
	mock.handler = cb
}

type ppsTransactionAPI struct {
	ppsServerAPI
	mock *MockPPSTransactionServer
}

// MockPPSTransactionServer provides a mocking interface for overriding PPS
// behavior inside transactions.
type MockPPSTransactionServer struct {
	Api                          ppsTransactionAPI
	NewPropagater                mockNewPropagater
	NewJobStopper                mockNewJobStopper
	NewJobFinisher               mockNewJobFinisher
	StopJobInTransaction         mockStopJobInTransaction
	UpdateJobStateInTransaction  mockUpdateJobStateInTransaction
	CreatePipelineInTransaction  mockCreatePipelineInTransaction
	InspectPipelineInTransaction mockInspectPipelineInTransaction
	ActivateAuthInTransaction    mockActivateAuthInTransaction
}

type MockPPSPropagater struct{}

func (mpp *MockPPSPropagater) PropagateJobs() {}
func (mpp *MockPPSPropagater) Run() error     { return nil }

func (api *ppsTransactionAPI) NewPropagater(txnCtx *txncontext.TransactionContext) txncontext.PpsPropagater {
	if api.mock.NewPropagater.handler != nil {
		return api.mock.NewPropagater.handler(txnCtx)
	}
	return &MockPPSPropagater{}
}

type MockPPSJobStopper struct{}

func (mpp *MockPPSJobStopper) StopJobs(*pfs.CommitSet) {}
func (mpp *MockPPSJobStopper) Run() error              { return nil }

func (api *ppsTransactionAPI) NewJobStopper(txnCtx *txncontext.TransactionContext) txncontext.PpsJobStopper {
	if api.mock.NewJobStopper.handler != nil {
		return api.mock.NewJobStopper.handler(txnCtx)
	}
	return &MockPPSJobStopper{}
}

type MockPPSJobFinisher struct{}

func (mpp *MockPPSJobFinisher) FinishJob(*pfs.CommitInfo) {}
func (mpp *MockPPSJobFinisher) Run() error                { return nil }

func (api *ppsTransactionAPI) NewJobFinisher(txnCtx *txncontext.TransactionContext) txncontext.PpsJobFinisher {
	if api.mock.NewJobFinisher.handler != nil {
		return api.mock.NewJobFinisher.handler(txnCtx)
	}
	return &MockPPSJobFinisher{}
}

func (api *ppsTransactionAPI) StopJobInTransaction(txnCtx *txncontext.TransactionContext, req *pps.StopJobRequest) error {
	if api.mock.StopJobInTransaction.handler != nil {
		return api.mock.StopJobInTransaction.handler(txnCtx, req)
	}
	return errors.Errorf("unhandled pachd mock: pps.StopJobInTransaction")
}

func (api *ppsTransactionAPI) UpdateJobStateInTransaction(txnCtx *txncontext.TransactionContext, req *pps.UpdateJobStateRequest) error {
	if api.mock.UpdateJobStateInTransaction.handler != nil {
		return api.mock.UpdateJobStateInTransaction.handler(txnCtx, req)
	}
	return errors.Errorf("unhandled pachd mock: pps.UpdateJobStateInTransaction")
}

func (api *ppsTransactionAPI) CreatePipelineInTransaction(txnCtx *txncontext.TransactionContext, req *pps.CreatePipelineRequest) error {
	if api.mock.CreatePipelineInTransaction.handler != nil {
		return api.mock.CreatePipelineInTransaction.handler(txnCtx, req)
	}
	return errors.Errorf("unhandled pachd mock: pps.CreatePipelineInTransaction")
}

func (api *ppsTransactionAPI) InspectPipelineInTransaction(txnCtx *txncontext.TransactionContext, pipeline string) (*pps.PipelineInfo, error) {
	if api.mock.InspectPipelineInTransaction.handler != nil {
		return api.mock.InspectPipelineInTransaction.handler(txnCtx, pipeline)
	}
	return nil, errors.Errorf("unhandled pachd mock: pps.InspectPipelineInTransaction")
}

func (api *ppsTransactionAPI) ActivateAuthInTransaction(txnCtx *txncontext.TransactionContext, req *pps.ActivateAuthRequest) (*pps.ActivateAuthResponse, error) {
	if api.mock.ActivateAuthInTransaction.handler != nil {
		return api.mock.ActivateAuthInTransaction.handler(txnCtx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock: pps.AcivateAuthInTransaction")
}

// NewMockPPSTransactionServer instantiates a MockPPSTransactionServer
func NewMockPPSTransactionServer() *MockPPSTransactionServer {
	result := &MockPPSTransactionServer{}
	result.Api.mock = result
	return result
}
