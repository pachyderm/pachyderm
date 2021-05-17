package transactionenv

import (
	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/context"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func unimplementedError(name string) error {
	return errors.Errorf("%s is not implemented in this mock", name)
}

// MockPfsPropagater is a simple mock that can be used to satisfy the
// PfsPropagater interface
type MockPfsPropagater struct{}

// NewMockPfsPropagater instantiates a MockPfsPropagater
func NewMockPfsPropagater() *MockPfsPropagater {
	return &MockPfsPropagater{}
}

// PropagateCommit always errors
func (mpp *MockPfsPropagater) PropagateCommit(branch *pfs.Branch, isNewCommit bool) error {
	return unimplementedError("PfsPropagater.PropagateCommit")
}

// Run always errors
func (mpp *MockPfsPropagater) Run() error {
	return unimplementedError("PfsPropagater.Run")
}

// MockAuthTransactionServer is a simple mock that can be used to satisfy the
// AuthTransactionServer interface
type MockAuthTransactionServer struct{}

// NewMockAuthTransactionServer instantiates a MockAuthTransactionServer
func NewMockAuthTransactionServer() *MockAuthTransactionServer {
	return &MockAuthTransactionServer{}
}

// AuthorizeInTransaction always errors
func (mats *MockAuthTransactionServer) AuthorizeInTransaction(*txncontext.TransactionContext, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	return nil, unimplementedError("AuthTransactionServer.AuthorizeInTransaction")
}

// ModifyRoleBindingInTransaction always errors
func (mats *MockAuthTransactionServer) ModifyRoleBindingInTransaction(*txncontext.TransactionContext, *auth.ModifyRoleBindingRequest) (*auth.ModifyRoleBindingResponse, error) {
	return nil, unimplementedError("AuthTransactionServer.ModifyRoleBindingInTransaction")
}

// GetRoleBindingInTransaction always errors
func (mats *MockAuthTransactionServer) GetRoleBindingInTransaction(*txncontext.TransactionContext, *auth.GetRoleBindingRequest) (*auth.GetRoleBindingResponse, error) {
	return nil, unimplementedError("AuthTransactionServer.GetRoleBindingInTransaction")
}

// MockPfsTransactionServer is a simple mock that can be used to satisfy the
// PfsTransactionServer interface
type MockPfsTransactionServer struct{}

// NewMockPfsTransactionServer instantiates a MockPfsTransactionServer
func NewMockPfsTransactionServer() *MockPfsTransactionServer {
	return &MockPfsTransactionServer{}
}

// NewPropagater returns a MockPfsPropagater
func (mpts *MockPfsTransactionServer) NewPropagater(*sqlx.Tx) txncontext.PfsPropagater {
	return NewMockPfsPropagater()
}

// CreateRepoInTransaction always errors
func (mpts *MockPfsTransactionServer) CreateRepoInTransaction(*txncontext.TransactionContext, *pfs.CreateRepoRequest) error {
	return unimplementedError("PfsTransactionServer.CreateRepoInTransaction")
}

// InspectRepoInTransaction always errors
func (mpts *MockPfsTransactionServer) InspectRepoInTransaction(*txncontext.TransactionContext, *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	return nil, unimplementedError("PfsTransactionServer.InspectRepoInTransaction")
}

// DeleteRepoInTransaction always errors
func (mpts *MockPfsTransactionServer) DeleteRepoInTransaction(*txncontext.TransactionContext, *pfs.DeleteRepoRequest) error {
	return unimplementedError("PfsTransactionServer.DeleteRepoInTransaction")
}

// StartCommitInTransaction always errors
func (mpts *MockPfsTransactionServer) StartCommitInTransaction(*txncontext.TransactionContext, *pfs.StartCommitRequest, *pfs.Commit) (*pfs.Commit, error) {
	return nil, unimplementedError("PfsTransactionServer.StartCommitInTransaction")
}

// FinishCommitInTransaction always errors
func (mpts *MockPfsTransactionServer) FinishCommitInTransaction(*txncontext.TransactionContext, *pfs.FinishCommitRequest) error {
	return unimplementedError("PfsTransactionServer.FinishCommitInTransaction")
}

// SquashCommitInTransaction always errors
func (mpts *MockPfsTransactionServer) SquashCommitInTransaction(*txncontext.TransactionContext, *pfs.SquashCommitRequest) error {
	return unimplementedError("PfsTransactionServer.SquashCommitInTransaction")
}

// CreateBranchInTransaction always errors
func (mpts *MockPfsTransactionServer) CreateBranchInTransaction(*txncontext.TransactionContext, *pfs.CreateBranchRequest) error {
	return unimplementedError("PfsTransactionServer.CreateBranchInTransaction")
}

// DeleteBranchInTransaction always errors
func (mpts *MockPfsTransactionServer) DeleteBranchInTransaction(*txncontext.TransactionContext, *pfs.DeleteBranchRequest) error {
	return unimplementedError("PfsTransactionServer.DeleteBranchInTransaction")
}

// AddFilesetInTransaction always errors
func (mpts *MockPfsTransactionServer) AddFilesetInTransaction(*TransactionContext, *pfs.AddFilesetRequest) error {
	return unimplementedError("PfsTransactionServer.AddFilesetInTransaction")
}

// MockPpsTransactionServer is a simple mock that can be used to satisfy the
// PpsTransactionServer interface
type MockPpsTransactionServer struct{}

// NewMockPpsTransactionServer instantiates a MockPpsTransactionServer
func NewMockPpsTransactionServer() *MockPpsTransactionServer {
	return &MockPpsTransactionServer{}
}

// StopPipelineJobInTransaction always errors
func (mpts *MockPpsTransactionServer) StopPipelineJobInTransaction(*txncontext.TransactionContext, *pps.StopPipelineJobRequest) error {
	return unimplementedError("PpsTransactionServer.StopPipelineJobInTransaction")
}

// UpdatePipelineJobStateInTransaction always errors
func (mpts *MockPpsTransactionServer) UpdatePipelineJobStateInTransaction(*txncontext.TransactionContext, *pps.UpdatePipelineJobStateRequest) error {
	return unimplementedError("PpsTransactionServer.UpdatePipelineJobStateInTransaction")
}

// CreatePipelineInTransaction always errors
func (mpts *MockPpsTransactionServer) CreatePipelineInTransaction(*txncontext.TransactionContext, *pps.CreatePipelineRequest, *string, **pfs.Commit) error {
	return unimplementedError("PpsTransactionServer.CreatePipelineInTransaction")
}
