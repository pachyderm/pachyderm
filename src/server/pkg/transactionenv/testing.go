package transactionenv

import (
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
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
func (mats *MockAuthTransactionServer) AuthorizeInTransaction(*TransactionContext, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	return nil, unimplementedError("AuthTransactionServer.AuthorizeInTransaction")
}

// GetScopeInTransaction always errors
func (mats *MockAuthTransactionServer) GetScopeInTransaction(*TransactionContext, *auth.GetScopeRequest) (*auth.GetScopeResponse, error) {
	return nil, unimplementedError("AuthTransactionServer.GetScopeInTransaction")
}

// SetScopeInTransaction always errors
func (mats *MockAuthTransactionServer) SetScopeInTransaction(*TransactionContext, *auth.SetScopeRequest) (*auth.SetScopeResponse, error) {
	return nil, unimplementedError("AuthTransactionServer.SetScopeInTransaction")
}

// GetACLInTransaction always errors
func (mats *MockAuthTransactionServer) GetACLInTransaction(*TransactionContext, *auth.GetACLRequest) (*auth.GetACLResponse, error) {
	return nil, unimplementedError("AuthTransactionServer.GetACLInTransaction")
}

// SetACLInTransaction always errors
func (mats *MockAuthTransactionServer) SetACLInTransaction(*TransactionContext, *auth.SetACLRequest) (*auth.SetACLResponse, error) {
	return nil, unimplementedError("AuthTransactionServer.SetACLInTransaction")
}

// MockPfsTransactionServer is a simple mock that can be used to satisfy the
// PfsTransactionServer interface
type MockPfsTransactionServer struct{}

// NewMockPfsTransactionServer instantiates a MockPfsTransactionServer
func NewMockPfsTransactionServer() *MockPfsTransactionServer {
	return &MockPfsTransactionServer{}
}

// NewPropagater returns a MockPfsPropagater
func (mpts *MockPfsTransactionServer) NewPropagater(col.STM) PfsPropagater {
	return NewMockPfsPropagater()
}

// CreateRepoInTransaction always errors
func (mpts *MockPfsTransactionServer) CreateRepoInTransaction(*TransactionContext, *pfs.CreateRepoRequest) error {
	return unimplementedError("PfsTransactionServer.CreateRepoInTransaction")
}

// InspectRepoInTransaction always errors
func (mpts *MockPfsTransactionServer) InspectRepoInTransaction(*TransactionContext, *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	return nil, unimplementedError("PfsTransactionServer.InspectRepoInTransaction")
}

// DeleteRepoInTransaction always errors
func (mpts *MockPfsTransactionServer) DeleteRepoInTransaction(*TransactionContext, *pfs.DeleteRepoRequest) error {
	return unimplementedError("PfsTransactionServer.DeleteRepoInTransaction")
}

// StartCommitInTransaction always errors
func (mpts *MockPfsTransactionServer) StartCommitInTransaction(*TransactionContext, *pfs.StartCommitRequest, *pfs.Commit) (*pfs.Commit, error) {
	return nil, unimplementedError("PfsTransactionServer.StartCommitInTransaction")
}

// FinishCommitInTransaction always errors
func (mpts *MockPfsTransactionServer) FinishCommitInTransaction(*TransactionContext, *pfs.FinishCommitRequest) error {
	return unimplementedError("PfsTransactionServer.FinishCommitInTransaction")
}

// DeleteCommitInTransaction always errors
func (mpts *MockPfsTransactionServer) DeleteCommitInTransaction(*TransactionContext, *pfs.DeleteCommitRequest) error {
	return unimplementedError("PfsTransactionServer.DeleteCommitInTransaction")
}

// CreateBranchInTransaction always errors
func (mpts *MockPfsTransactionServer) CreateBranchInTransaction(*TransactionContext, *pfs.CreateBranchRequest) error {
	return unimplementedError("PfsTransactionServer.CreateBranchInTransaction")
}

// DeleteBranchInTransaction always errors
func (mpts *MockPfsTransactionServer) DeleteBranchInTransaction(*TransactionContext, *pfs.DeleteBranchRequest) error {
	return unimplementedError("PfsTransactionServer.DeleteBranchInTransaction")
}

// MockPpsTransactionServer is a simple mock that can be used to satisfy the
// PpsTransactionServer interface
type MockPpsTransactionServer struct{}

// NewMockPpsTransactionServer instantiates a MockPpsTransactionServer
func NewMockPpsTransactionServer() *MockPpsTransactionServer {
	return &MockPpsTransactionServer{}
}

// UpdateJobStateInTransaction always errors
func (mpts *MockPpsTransactionServer) UpdateJobStateInTransaction(*TransactionContext, *pps.UpdateJobStateRequest) error {
	return unimplementedError("PpsTransactionServer.UpdateJobStateInTransaction")
}
