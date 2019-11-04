package testutil

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	version "github.com/pachyderm/pachyderm/src/client/version/versionpb"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

/* Admin Server Mocks */

type extractFunc func(*admin.ExtractRequest, admin.API_ExtractServer) error
type extractPipelineFunc func(context.Context, *admin.ExtractPipelineRequest) (*admin.Op, error)
type restoreFunc func(admin.API_RestoreServer) error
type inspectClusterFunc func(context.Context, *types.Empty) (*admin.ClusterInfo, error)

type mockExtract struct{ handler extractFunc }
type mockExtractPipeline struct{ handler extractPipelineFunc }
type mockRestore struct{ handler restoreFunc }
type mockInspectCluster struct{ handler inspectClusterFunc }

func (mock *mockExtract) Use(cb extractFunc)                 { mock.handler = cb }
func (mock *mockExtractPipeline) Use(cb extractPipelineFunc) { mock.handler = cb }
func (mock *mockRestore) Use(cb restoreFunc)                 { mock.handler = cb }
func (mock *mockInspectCluster) Use(cb inspectClusterFunc)   { mock.handler = cb }

type mockAdminServer struct {
	MockExtract         mockExtract
	MockExtractPipeline mockExtractPipeline
	MockRestore         mockRestore
	MockInspectCluster  mockInspectCluster
}

func (mock *mockAdminServer) Extract(req *admin.ExtractRequest, serv admin.API_ExtractServer) error {
	if mock.MockExtract.handler != nil {
		return mock.MockExtract.handler(req, serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockAdminServer) ExtractPipeline(ctx context.Context, req *admin.ExtractPipelineRequest) (*admin.Op, error) {
	if mock.MockExtractPipeline.handler != nil {
		return mock.MockExtractPipeline.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAdminServer) Restore(serv admin.API_RestoreServer) error {
	if mock.MockRestore.handler != nil {
		return mock.MockRestore.handler(serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockAdminServer) InspectCluster(ctx context.Context, req *types.Empty) (*admin.ClusterInfo, error) {
	if mock.MockInspectCluster.handler != nil {
		return mock.MockInspectCluster.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}

/* Auth Server Mocks */

type activateAuthFunc func(context.Context, *auth.ActivateRequest) (*auth.ActivateResponse, error)
type deactivateAuthFunc func(context.Context, *auth.DeactivateRequest) (*auth.DeactivateResponse, error)
type getConfigurationFunc func(context.Context, *auth.GetConfigurationRequest) (*auth.GetConfigurationResponse, error)
type setConfigurationFunc func(context.Context, *auth.SetConfigurationRequest) (*auth.SetConfigurationResponse, error)
type getAdminsFunc func(context.Context, *auth.GetAdminsRequest) (*auth.GetAdminsResponse, error)
type modifyAdminsFunc func(context.Context, *auth.ModifyAdminsRequest) (*auth.ModifyAdminsResponse, error)
type authenticateFunc func(context.Context, *auth.AuthenticateRequest) (*auth.AuthenticateResponse, error)
type authorizeFunc func(context.Context, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error)
type whoAmIFunc func(context.Context, *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error)
type getScopeFunc func(context.Context, *auth.GetScopeRequest) (*auth.GetScopeResponse, error)
type setScopeFunc func(context.Context, *auth.SetScopeRequest) (*auth.SetScopeResponse, error)
type getACLFunc func(context.Context, *auth.GetACLRequest) (*auth.GetACLResponse, error)
type setACLFunc func(context.Context, *auth.SetACLRequest) (*auth.SetACLResponse, error)
type getAuthTokenFunc func(context.Context, *auth.GetAuthTokenRequest) (*auth.GetAuthTokenResponse, error)
type extendAuthTokenFunc func(context.Context, *auth.ExtendAuthTokenRequest) (*auth.ExtendAuthTokenResponse, error)
type revokeAuthTokenFunc func(context.Context, *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error)
type setGroupsForUserFunc func(context.Context, *auth.SetGroupsForUserRequest) (*auth.SetGroupsForUserResponse, error)
type modifyMembersFunc func(context.Context, *auth.ModifyMembersRequest) (*auth.ModifyMembersResponse, error)
type getGroupsFunc func(context.Context, *auth.GetGroupsRequest) (*auth.GetGroupsResponse, error)
type getUsersFunc func(context.Context, *auth.GetUsersRequest) (*auth.GetUsersResponse, error)
type getOneTimePasswordFunc func(context.Context, *auth.GetOneTimePasswordRequest) (*auth.GetOneTimePasswordResponse, error)

type mockActivateAuth struct{ handler activateAuthFunc }
type mockDeactivateAuth struct{ handler deactivateAuthFunc }
type mockGetConfiguration struct{ handler getConfigurationFunc }
type mockSetConfiguration struct{ handler setConfigurationFunc }
type mockGetAdmins struct{ handler getAdminsFunc }
type mockModifyAdmins struct{ handler modifyAdminsFunc }
type mockAuthenticate struct{ handler authenticateFunc }
type mockAuthorize struct{ handler authorizeFunc }
type mockWhoAmI struct{ handler whoAmIFunc }
type mockGetScope struct{ handler getScopeFunc }
type mockSetScope struct{ handler setScopeFunc }
type mockGetACL struct{ handler getACLFunc }
type mockSetACL struct{ handler setACLFunc }
type mockGetAuthToken struct{ handler getAuthTokenFunc }
type mockExtendAuthToken struct{ handler extendAuthTokenFunc }
type mockRevokeAuthToken struct{ handler revokeAuthTokenFunc }
type mockSetGroupsForUser struct{ handler setGroupsForUserFunc }
type mockModifyMembers struct{ handler modifyMembersFunc }
type mockGetGroups struct{ handler getGroupsFunc }
type mockGetUsers struct{ handler getUsersFunc }
type mockGetOneTimePassword struct{ handler getOneTimePasswordFunc }

func (mock *mockActivateAuth) Use(cb activateAuthFunc)             { mock.handler = cb }
func (mock *mockDeactivateAuth) Use(cb deactivateAuthFunc)         { mock.handler = cb }
func (mock *mockGetConfiguration) Use(cb getConfigurationFunc)     { mock.handler = cb }
func (mock *mockSetConfiguration) Use(cb setConfigurationFunc)     { mock.handler = cb }
func (mock *mockGetAdmins) Use(cb getAdminsFunc)                   { mock.handler = cb }
func (mock *mockModifyAdmins) Use(cb modifyAdminsFunc)             { mock.handler = cb }
func (mock *mockAuthenticate) Use(cb authenticateFunc)             { mock.handler = cb }
func (mock *mockAuthorize) Use(cb authorizeFunc)                   { mock.handler = cb }
func (mock *mockWhoAmI) Use(cb whoAmIFunc)                         { mock.handler = cb }
func (mock *mockGetScope) Use(cb getScopeFunc)                     { mock.handler = cb }
func (mock *mockSetScope) Use(cb setScopeFunc)                     { mock.handler = cb }
func (mock *mockGetACL) Use(cb getACLFunc)                         { mock.handler = cb }
func (mock *mockSetACL) Use(cb setACLFunc)                         { mock.handler = cb }
func (mock *mockGetAuthToken) Use(cb getAuthTokenFunc)             { mock.handler = cb }
func (mock *mockExtendAuthToken) Use(cb extendAuthTokenFunc)       { mock.handler = cb }
func (mock *mockSetGroupsForUser) Use(cb setGroupsForUserFunc)     { mock.handler = cb }
func (mock *mockModifyMembers) Use(cb modifyMembersFunc)           { mock.handler = cb }
func (mock *mockGetGroups) Use(cb getGroupsFunc)                   { mock.handler = cb }
func (mock *mockGetUsers) Use(cb getUsersFunc)                     { mock.handler = cb }
func (mock *mockGetOneTimePassword) Use(cb getOneTimePasswordFunc) { mock.handler = cb }

type mockAuthServer struct {
	MockActivate           mockActivateAuth
	MockDeactivate         mockDeactivateAuth
	MockGetConfiguration   mockGetConfiguration
	MockSetConfiguration   mockSetConfiguration
	MockGetAdmins          mockGetAdmins
	MockModifyAdmins       mockModifyAdmins
	MockAuthenticate       mockAuthenticate
	MockAuthorize          mockAuthorize
	MockWhoAmI             mockWhoAmI
	MockGetScope           mockGetScope
	MockSetScope           mockSetScope
	MockGetACL             mockGetACL
	MockSetACL             mockSetACL
	MockGetAuthToken       mockGetAuthToken
	MockExtendAuthToken    mockExtendAuthToken
	MockRevokeAuthToken    mockRevokeAuthToken
	MockSetGroupsForUser   mockSetGroupsForUser
	MockModifyMembers      mockModifyMembers
	MockGetGroups          mockGetGroups
	MockGetUsers           mockGetUsers
	MockGetOneTimePassword mockGetOneTimePassword
}

func (mock *mockAuthServer) Activate(ctx context.Context, req *auth.ActivateRequest) (*auth.ActivateResponse, error) {
	if mock.MockActivate.handler != nil {
		return mock.MockActivate.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) Deactivate(ctx context.Context, req *auth.DeactivateRequest) (*auth.DeactivateResponse, error) {
	if mock.MockDeactivate.handler != nil {
		return mock.MockDeactivate.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetConfiguration(ctx context.Context, req *auth.GetConfigurationRequest) (*auth.GetConfigurationResponse, error) {
	if mock.MockGetConfiguration.handler != nil {
		return mock.MockGetConfiguration.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) SetConfiguration(ctx context.Context, req *auth.SetConfigurationRequest) (*auth.SetConfigurationResponse, error) {
	if mock.MockSetConfiguration.handler != nil {
		return mock.MockSetConfiguration.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetAdmins(ctx context.Context, req *auth.GetAdminsRequest) (*auth.GetAdminsResponse, error) {
	if mock.MockGetAdmins.handler != nil {
		return mock.MockGetAdmins.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) ModifyAdmins(ctx context.Context, req *auth.ModifyAdminsRequest) (*auth.ModifyAdminsResponse, error) {
	if mock.MockModifyAdmins.handler != nil {
		return mock.MockModifyAdmins.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) Authenticate(ctx context.Context, req *auth.AuthenticateRequest) (*auth.AuthenticateResponse, error) {
	if mock.MockAuthenticate.handler != nil {
		return mock.MockAuthenticate.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	if mock.MockAuthorize.handler != nil {
		return mock.MockAuthorize.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) WhoAmI(ctx context.Context, req *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error) {
	if mock.MockWhoAmI.handler != nil {
		return mock.MockWhoAmI.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetScope(ctx context.Context, req *auth.GetScopeRequest) (*auth.GetScopeResponse, error) {
	if mock.MockGetScope.handler != nil {
		return mock.MockGetScope.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) SetScope(ctx context.Context, req *auth.SetScopeRequest) (*auth.SetScopeResponse, error) {
	if mock.MockSetScope.handler != nil {
		return mock.MockSetScope.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetACL(ctx context.Context, req *auth.GetACLRequest) (*auth.GetACLResponse, error) {
	if mock.MockGetACL.handler != nil {
		return mock.MockGetACL.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) SetACL(ctx context.Context, req *auth.SetACLRequest) (*auth.SetACLResponse, error) {
	if mock.MockSetACL.handler != nil {
		return mock.MockSetACL.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetAuthToken(ctx context.Context, req *auth.GetAuthTokenRequest) (*auth.GetAuthTokenResponse, error) {
	if mock.MockGetAuthToken.handler != nil {
		return mock.MockGetAuthToken.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) ExtendAuthToken(ctx context.Context, req *auth.ExtendAuthTokenRequest) (*auth.ExtendAuthTokenResponse, error) {
	if mock.MockExtendAuthToken.handler != nil {
		return mock.MockExtendAuthToken.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) RevokeAuthToken(ctx context.Context, req *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error) {
	if mock.MockRevokeAuthToken.handler != nil {
		return mock.MockRevokeAuthToken.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) SetGroupsForUser(ctx context.Context, req *auth.SetGroupsForUserRequest) (*auth.SetGroupsForUserResponse, error) {
	if mock.MockSetGroupsForUser.handler != nil {
		return mock.MockSetGroupsForUser.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) ModifyMembers(ctx context.Context, req *auth.ModifyMembersRequest) (*auth.ModifyMembersResponse, error) {
	if mock.MockModifyMembers.handler != nil {
		return mock.MockModifyMembers.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetGroups(ctx context.Context, req *auth.GetGroupsRequest) (*auth.GetGroupsResponse, error) {
	if mock.MockGetGroups.handler != nil {
		return mock.MockGetGroups.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetUsers(ctx context.Context, req *auth.GetUsersRequest) (*auth.GetUsersResponse, error) {
	if mock.MockGetUsers.handler != nil {
		return mock.MockGetUsers.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetOneTimePassword(ctx context.Context, req *auth.GetOneTimePasswordRequest) (*auth.GetOneTimePasswordResponse, error) {
	if mock.MockGetOneTimePassword.handler != nil {
		return mock.MockGetOneTimePassword.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}

/* Enterprise Server Mocks */

type activateEnterpriseFunc func(context.Context, *enterprise.ActivateRequest) (*enterprise.ActivateResponse, error)
type getStateFunc func(context.Context, *enterprise.GetStateRequest) (*enterprise.GetStateResponse, error)
type deactivateEnterpriseFunc func(context.Context, *enterprise.DeactivateRequest) (*enterprise.DeactivateResponse, error)

type mockActivateEnterprise struct{ handler activateEnterpriseFunc }
type mockGetState struct{ handler getStateFunc }
type mockDeactivateEnterprise struct{ handler deactivateEnterpriseFunc }

func (mock *mockActivateEnterprise) Use(cb activateEnterpriseFunc)     { mock.handler = cb }
func (mock *mockGetState) Use(cb getStateFunc)                         { mock.handler = cb }
func (mock *mockDeactivateEnterprise) Use(cb deactivateEnterpriseFunc) { mock.handler = cb }

type mockEnterpriseServer struct {
	MockActivate   mockActivateEnterprise
	MockGetState   mockGetState
	MockDeactivate mockDeactivateEnterprise
}

func (mock *mockEnterpriseServer) Activate(ctx context.Context, req *enterprise.ActivateRequest) (*enterprise.ActivateResponse, error) {
	if mock.MockActivate.handler != nil {
		return mock.MockActivate.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockEnterpriseServer) GetState(ctx context.Context, req *enterprise.GetStateRequest) (*enterprise.GetStateResponse, error) {
	if mock.MockGetState.handler != nil {
		return mock.MockGetState.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockEnterpriseServer) Deactivate(ctx context.Context, req *enterprise.DeactivateRequest) (*enterprise.DeactivateResponse, error) {
	if mock.MockDeactivate.handler != nil {
		return mock.MockDeactivate.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}

/* PFS Server Mocks */

type mockPfsServer struct{}

func (mock *mockPfsServer) CreateRepo(context.Context, *pfs.CreateRepoRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) InspectRepo(context.Context, *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListRepo(context.Context, *pfs.ListRepoRequest) (*pfs.ListRepoResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DeleteRepo(context.Context, *pfs.DeleteRepoRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) StartCommit(context.Context, *pfs.StartCommitRequest) (*pfs.Commit, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) FinishCommit(context.Context, *pfs.FinishCommitRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) InspectCommit(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListCommit(context.Context, *pfs.ListCommitRequest) (*pfs.CommitInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListCommitStream(*pfs.ListCommitRequest, pfs.API_ListCommitStreamServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DeleteCommit(context.Context, *pfs.DeleteCommitRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) FlushCommit(*pfs.FlushCommitRequest, pfs.API_FlushCommitServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) SubscribeCommit(*pfs.SubscribeCommitRequest, pfs.API_SubscribeCommitServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) BuildCommit(context.Context, *pfs.BuildCommitRequest) (*pfs.Commit, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) CreateBranch(context.Context, *pfs.CreateBranchRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) InspectBranch(context.Context, *pfs.InspectBranchRequest) (*pfs.BranchInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListBranch(context.Context, *pfs.ListBranchRequest) (*pfs.BranchInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DeleteBranch(context.Context, *pfs.DeleteBranchRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) PutFile(pfs.API_PutFileServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) CopyFile(context.Context, *pfs.CopyFileRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) GetFile(*pfs.GetFileRequest, pfs.API_GetFileServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) InspectFile(context.Context, *pfs.InspectFileRequest) (*pfs.FileInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListFile(context.Context, *pfs.ListFileRequest) (*pfs.FileInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListFileStream(*pfs.ListFileRequest, pfs.API_ListFileStreamServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) WalkFile(*pfs.WalkFileRequest, pfs.API_WalkFileServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) GlobFile(context.Context, *pfs.GlobFileRequest) (*pfs.FileInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) GlobFileStream(*pfs.GlobFileRequest, pfs.API_GlobFileStreamServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DiffFile(context.Context, *pfs.DiffFileRequest) (*pfs.DiffFileResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DeleteFile(context.Context, *pfs.DeleteFileRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DeleteAll(context.Context, *types.Empty) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) Fsck(*pfs.FsckRequest, pfs.API_FsckServer) error {
	return fmt.Errorf("Mock")
}

/* PPS Server Mocks */

type mockPpsServer struct{}

func (mock *mockPpsServer) CreateJob(context.Context, *pps.CreateJobRequest) (*pps.Job, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) InspectJob(context.Context, *pps.InspectJobRequest) (*pps.JobInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ListJob(context.Context, *pps.ListJobRequest) (*pps.JobInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ListJobStream(*pps.ListJobRequest, pps.API_ListJobStreamServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPpsServer) FlushJob(*pps.FlushJobRequest, pps.API_FlushJobServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPpsServer) DeleteJob(context.Context, *pps.DeleteJobRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) StopJob(context.Context, *pps.StopJobRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) InspectDatum(context.Context, *pps.InspectDatumRequest) (*pps.DatumInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ListDatum(context.Context, *pps.ListDatumRequest) (*pps.ListDatumResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ListDatumStream(*pps.ListDatumRequest, pps.API_ListDatumStreamServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPpsServer) RestartDatum(context.Context, *pps.RestartDatumRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) CreatePipeline(context.Context, *pps.CreatePipelineRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) InspectPipeline(context.Context, *pps.InspectPipelineRequest) (*pps.PipelineInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ListPipeline(context.Context, *pps.ListPipelineRequest) (*pps.PipelineInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) DeletePipeline(context.Context, *pps.DeletePipelineRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) StartPipeline(context.Context, *pps.StartPipelineRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) StopPipeline(context.Context, *pps.StopPipelineRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) RunPipeline(context.Context, *pps.RunPipelineRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) DeleteAll(context.Context, *types.Empty) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) GetLogs(*pps.GetLogsRequest, pps.API_GetLogsServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPpsServer) GarbageCollect(context.Context, *pps.GarbageCollectRequest) (*pps.GarbageCollectResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ActivateAuth(context.Context, *pps.ActivateAuthRequest) (*pps.ActivateAuthResponse, error) {
	return nil, fmt.Errorf("Mock")
}

/* Transaction Server Mocks */

type startTransactionFunc func(context.Context, *transaction.StartTransactionRequest) (*transaction.Transaction, error)
type inspectTransactionFunc func(context.Context, *transaction.InspectTransactionRequest) (*transaction.TransactionInfo, error)
type deleteTransactionFunc func(context.Context, *transaction.DeleteTransactionRequest) (*types.Empty, error)
type listTransactionFunc func(context.Context, *transaction.ListTransactionRequest) (*transaction.TransactionInfos, error)
type finishTransactionFunc func(context.Context, *transaction.FinishTransactionRequest) (*transaction.TransactionInfo, error)
type deleteAllTransactionFunc func(context.Context, *transaction.DeleteAllRequest) (*types.Empty, error)

type mockStartTransaction struct{ handler startTransactionFunc }
type mockInspectTransaction struct{ handler inspectTransactionFunc }
type mockDeleteTransaction struct{ handler deleteTransactionFunc }
type mockListTransaction struct{ handler listTransactionFunc }
type mockFinishTransaction struct{ handler finishTransactionFunc }
type mockDeleteAllTransaction struct{ handler deleteAllTransactionFunc }

func (mock *mockStartTransaction) Use(cb startTransactionFunc)         { mock.handler = cb }
func (mock *mockInspectTransaction) Use(cb inspectTransactionFunc)     { mock.handler = cb }
func (mock *mockDeleteTransaction) Use(cb deleteTransactionFunc)       { mock.handler = cb }
func (mock *mockListTransaction) Use(cb listTransactionFunc)           { mock.handler = cb }
func (mock *mockFinishTransaction) Use(cb finishTransactionFunc)       { mock.handler = cb }
func (mock *mockDeleteAllTransaction) Use(cb deleteAllTransactionFunc) { mock.handler = cb }

type mockTransactionServer struct {
	MockStartTransaction   mockStartTransaction
	MockInspectTransaction mockInspectTransaction
	MockDeleteTransaction  mockDeleteTransaction
	MockListTransaction    mockListTransaction
	MockFinishTransaction  mockFinishTransaction
	MockDeleteAll          mockDeleteAllTransaction
}

func (mock *mockTransactionServer) StartTransaction(ctx context.Context, req *transaction.StartTransactionRequest) (*transaction.Transaction, error) {
	if mock.MockStartTransaction.handler != nil {
		return mock.MockStartTransaction.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockTransactionServer) InspectTransaction(ctx context.Context, req *transaction.InspectTransactionRequest) (*transaction.TransactionInfo, error) {
	if mock.MockInspectTransaction.handler != nil {
		return mock.MockInspectTransaction.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockTransactionServer) DeleteTransaction(ctx context.Context, req *transaction.DeleteTransactionRequest) (*types.Empty, error) {
	if mock.MockDeleteTransaction.handler != nil {
		return mock.MockDeleteTransaction.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockTransactionServer) ListTransaction(ctx context.Context, req *transaction.ListTransactionRequest) (*transaction.TransactionInfos, error) {
	if mock.MockListTransaction.handler != nil {
		return mock.MockListTransaction.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockTransactionServer) FinishTransaction(ctx context.Context, req *transaction.FinishTransactionRequest) (*transaction.TransactionInfo, error) {
	if mock.MockFinishTransaction.handler != nil {
		return mock.MockFinishTransaction.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockTransactionServer) DeleteAll(ctx context.Context, req *transaction.DeleteAllRequest) (*types.Empty, error) {
	if mock.MockDeleteAll.handler != nil {
		return mock.MockDeleteAll.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}

/* Version Server Mocks */

type getVersionFunc func(context.Context, *types.Empty) (*version.Version, error)

type mockGetVersion struct{ handler getVersionFunc }

func (mock *mockGetVersion) Use(cb getVersionFunc) { mock.handler = cb }

type mockVersionServer struct {
	MockGetVersion mockGetVersion
}

func (mock *mockVersionServer) GetVersion(ctx context.Context, req *types.Empty) (*version.Version, error) {
	if mock.MockGetVersion.handler != nil {
		return mock.MockGetVersion.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}

/* Object Server Mocks */

type putObjectFunc func(pfs.ObjectAPI_PutObjectServer) error
type putObjectSplitFunc func(pfs.ObjectAPI_PutObjectSplitServer) error
type putObjectsFunc func(pfs.ObjectAPI_PutObjectsServer) error
type createObjectFunc func(context.Context, *pfs.CreateObjectRequest) (*types.Empty, error)
type getObjectFunc func(*pfs.Object, pfs.ObjectAPI_GetObjectServer) error
type getObjectsFunc func(*pfs.GetObjectsRequest, pfs.ObjectAPI_GetObjectsServer) error
type putBlockFunc func(pfs.ObjectAPI_PutBlockServer) error
type getBlockFunc func(*pfs.GetBlockRequest, pfs.ObjectAPI_GetBlockServer) error
type getBlocksFunc func(*pfs.GetBlocksRequest, pfs.ObjectAPI_GetBlocksServer) error
type listBlockFunc func(*pfs.ListBlockRequest, pfs.ObjectAPI_ListBlockServer) error
type tagObjectFunc func(context.Context, *pfs.TagObjectRequest) (*types.Empty, error)
type inspectObjectFunc func(context.Context, *pfs.Object) (*pfs.ObjectInfo, error)
type checkObjectFunc func(context.Context, *pfs.CheckObjectRequest) (*pfs.CheckObjectResponse, error)
type listObjectsFunc func(*pfs.ListObjectsRequest, pfs.ObjectAPI_ListObjectsServer) error
type deleteObjectsFunc func(context.Context, *pfs.DeleteObjectsRequest) (*pfs.DeleteObjectsResponse, error)
type getTagFunc func(*pfs.Tag, pfs.ObjectAPI_GetTagServer) error
type inspectTagFunc func(context.Context, *pfs.Tag) (*pfs.ObjectInfo, error)
type listTagsFunc func(*pfs.ListTagsRequest, pfs.ObjectAPI_ListTagsServer) error
type deleteTagsFunc func(context.Context, *pfs.DeleteTagsRequest) (*pfs.DeleteTagsResponse, error)
type compactFunc func(context.Context, *types.Empty) (*types.Empty, error)

type mockPutObject struct{ handler putObjectFunc }
type mockPutObjectSplit struct{ handler putObjectSplitFunc }
type mockPutObjects struct{ handler putObjectsFunc }
type mockCreateObject struct{ handler createObjectFunc }
type mockGetObject struct{ handler getObjectFunc }
type mockGetObjects struct{ handler getObjectsFunc }
type mockPutBlock struct{ handler putBlockFunc }
type mockGetBlock struct{ handler getBlockFunc }
type mockGetBlocks struct{ handler getBlocksFunc }
type mockListBlock struct{ handler listBlockFunc }
type mockTagObject struct{ handler tagObjectFunc }
type mockInspectObject struct{ handler inspectObjectFunc }
type mockCheckObject struct{ handler checkObjectFunc }
type mockListObjects struct{ handler listObjectsFunc }
type mockDeleteObjects struct{ handler deleteObjectsFunc }
type mockGetTag struct{ handler getTagFunc }
type mockInspectTag struct{ handler inspectTagFunc }
type mockListTags struct{ handler listTagsFunc }
type mockDeleteTags struct{ handler deleteTagsFunc }
type mockCompact struct{ handler compactFunc }

func (mock *mockPutObject) Use(cb putObjectFunc)           { mock.handler = cb }
func (mock *mockPutObjectSplit) Use(cb putObjectSplitFunc) { mock.handler = cb }
func (mock *mockPutObjects) Use(cb putObjectsFunc)         { mock.handler = cb }
func (mock *mockCreateObject) Use(cb createObjectFunc)     { mock.handler = cb }
func (mock *mockGetObject) Use(cb getObjectFunc)           { mock.handler = cb }
func (mock *mockGetObjects) Use(cb getObjectsFunc)         { mock.handler = cb }
func (mock *mockPutBlock) Use(cb putBlockFunc)             { mock.handler = cb }
func (mock *mockGetBlock) Use(cb getBlockFunc)             { mock.handler = cb }
func (mock *mockGetBlocks) Use(cb getBlocksFunc)           { mock.handler = cb }
func (mock *mockListBlock) Use(cb listBlockFunc)           { mock.handler = cb }
func (mock *mockTagObject) Use(cb tagObjectFunc)           { mock.handler = cb }
func (mock *mockInspectObject) Use(cb inspectObjectFunc)   { mock.handler = cb }
func (mock *mockCheckObject) Use(cb checkObjectFunc)       { mock.handler = cb }
func (mock *mockListObjects) Use(cb listObjectsFunc)       { mock.handler = cb }
func (mock *mockDeleteObjects) Use(cb deleteObjectsFunc)   { mock.handler = cb }
func (mock *mockGetTag) Use(cb getTagFunc)                 { mock.handler = cb }
func (mock *mockInspectTag) Use(cb inspectTagFunc)         { mock.handler = cb }
func (mock *mockListTags) Use(cb listTagsFunc)             { mock.handler = cb }
func (mock *mockDeleteTags) Use(cb deleteTagsFunc)         { mock.handler = cb }
func (mock *mockCompact) Use(cb compactFunc)               { mock.handler = cb }

type mockObjectServer struct {
	MockPutObject      mockPutObject
	MockPutObjectSplit mockPutObjectSplit
	MockPutObjects     mockPutObjects
	MockCreateObject   mockCreateObject
	MockGetObject      mockGetObject
	MockGetObjects     mockGetObjects
	MockPutBlock       mockPutBlock
	MockGetBlock       mockGetBlock
	MockGetBlocks      mockGetBlocks
	MockListBlock      mockListBlock
	MockTagObject      mockTagObject
	MockInspectObject  mockInspectObject
	MockCheckObject    mockCheckObject
	MockListObjects    mockListObjects
	MockDeleteObjects  mockDeleteObjects
	MockGetTag         mockGetTag
	MockInspectTag     mockInspectTag
	MockListTags       mockListTags
	MockDeleteTags     mockDeleteTags
	MockCompact        mockCompact
}

func (mock *mockObjectServer) PutObject(serv pfs.ObjectAPI_PutObjectServer) error {
	if mock.MockPutObject.handler != nil {
		return mock.MockPutObject.handler(serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) PutObjectSplit(serv pfs.ObjectAPI_PutObjectSplitServer) error {
	if mock.MockPutObjectSplit.handler != nil {
		return mock.MockPutObjectSplit.handler(serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) PutObjects(serv pfs.ObjectAPI_PutObjectsServer) error {
	if mock.MockPutObjects.handler != nil {
		return mock.MockPutObjects.handler(serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) CreateObject(ctx context.Context, serv *pfs.CreateObjectRequest) (*types.Empty, error) {
	if mock.MockCreateObject.handler != nil {
		return mock.MockCreateObject.handler(ctx, serv)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) GetObject(req *pfs.Object, serv pfs.ObjectAPI_GetObjectServer) error {
	if mock.MockGetObject.handler != nil {
		return mock.MockGetObject.handler(req, serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) GetObjects(req *pfs.GetObjectsRequest, serv pfs.ObjectAPI_GetObjectsServer) error {
	if mock.MockGetObjects.handler != nil {
		return mock.MockGetObjects.handler(req, serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) PutBlock(serv pfs.ObjectAPI_PutBlockServer) error {
	if mock.MockPutBlock.handler != nil {
		return mock.MockPutBlock.handler(serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) GetBlock(req *pfs.GetBlockRequest, serv pfs.ObjectAPI_GetBlockServer) error {
	if mock.MockGetBlock.handler != nil {
		return mock.MockGetBlock.handler(req, serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) GetBlocks(req *pfs.GetBlocksRequest, serv pfs.ObjectAPI_GetBlocksServer) error {
	if mock.MockGetBlocks.handler != nil {
		return mock.MockGetBlocks.handler(req, serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) ListBlock(req *pfs.ListBlockRequest, serv pfs.ObjectAPI_ListBlockServer) error {
	if mock.MockListBlock.handler != nil {
		return mock.MockListBlock.handler(req, serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) TagObject(ctx context.Context, req *pfs.TagObjectRequest) (*types.Empty, error) {
	if mock.MockTagObject.handler != nil {
		return mock.MockTagObject.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) InspectObject(ctx context.Context, req *pfs.Object) (*pfs.ObjectInfo, error) {
	if mock.MockInspectObject.handler != nil {
		return mock.MockInspectObject.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) CheckObject(ctx context.Context, req *pfs.CheckObjectRequest) (*pfs.CheckObjectResponse, error) {
	if mock.MockCheckObject.handler != nil {
		return mock.MockCheckObject.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) ListObjects(req *pfs.ListObjectsRequest, serv pfs.ObjectAPI_ListObjectsServer) error {
	if mock.MockListObjects.handler != nil {
		return mock.MockListObjects.handler(req, serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) DeleteObjects(ctx context.Context, req *pfs.DeleteObjectsRequest) (*pfs.DeleteObjectsResponse, error) {
	if mock.MockDeleteObjects.handler != nil {
		return mock.MockDeleteObjects.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) GetTag(req *pfs.Tag, serv pfs.ObjectAPI_GetTagServer) error {
	if mock.MockGetTag.handler != nil {
		return mock.MockGetTag.handler(req, serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) InspectTag(ctx context.Context, req *pfs.Tag) (*pfs.ObjectInfo, error) {
	if mock.MockInspectTag.handler != nil {
		return mock.MockInspectTag.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) ListTags(req *pfs.ListTagsRequest, serv pfs.ObjectAPI_ListTagsServer) error {
	if mock.MockListTags.handler != nil {
		return mock.MockListTags.handler(req, serv)
	}
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) DeleteTags(ctx context.Context, req *pfs.DeleteTagsRequest) (*pfs.DeleteTagsResponse, error) {
	if mock.MockDeleteTags.handler != nil {
		return mock.MockDeleteTags.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) Compact(ctx context.Context, req *types.Empty) (*types.Empty, error) {
	if mock.MockCompact.handler != nil {
		return mock.MockCompact.handler(ctx, req)
	}
	return nil, fmt.Errorf("Mock")
}

// PachdMock provides an interface for running the interface for a Pachd API
// server locally without any of its dependencies. Tests may mock out specific
// API calls by providing a handler function, and later check information about
// the mocked calls.
type PachdMock struct {
	cancel context.CancelFunc
	eg     *errgroup.Group

	ObjectAPI      mockObjectServer
	PfsAPI         mockPfsServer
	PpsAPI         mockPpsServer
	AuthAPI        mockAuthServer
	TransactionAPI mockTransactionServer
	EnterpriseAPI  mockEnterpriseServer
	VersionAPI     mockVersionServer
	AdminAPI       mockAdminServer
}

// NewPachdMock constructs a mock Pachd API server whose behavior can be
// controlled through the PachdMock instance. By default, all API calls will
// error, unless a handler is specified.
func NewPachdMock(port uint16) *PachdMock {
	mock := &PachdMock{}

	ctx := context.Background()
	ctx, mock.cancel = context.WithCancel(ctx)
	mock.eg, ctx = errgroup.WithContext(ctx)

	mock.eg.Go(func() error {
		err := grpcutil.Serve(
			grpcutil.ServerOptions{
				Port:       port,
				MaxMsgSize: grpcutil.MaxMsgSize,
				RegisterFunc: func(s *grpc.Server) error {
					admin.RegisterAPIServer(s, &mock.AdminAPI)
					auth.RegisterAPIServer(s, &mock.AuthAPI)
					enterprise.RegisterAPIServer(s, &mock.EnterpriseAPI)
					pfs.RegisterObjectAPIServer(s, &mock.ObjectAPI)
					pfs.RegisterAPIServer(s, &mock.PfsAPI)
					pps.RegisterAPIServer(s, &mock.PpsAPI)
					transaction.RegisterAPIServer(s, &mock.TransactionAPI)
					version.RegisterAPIServer(s, &mock.VersionAPI)
					return nil
				},
			},
		)
		if err != nil {
			log.Printf("error starting grpc server %v\n", err)
		}
		return err
	})

	return mock
}

// Close will cancel the mock Pachd API server goroutine and return its result
func (mock *PachdMock) Close() error {
	mock.cancel()
	return mock.eg.Wait()
}
