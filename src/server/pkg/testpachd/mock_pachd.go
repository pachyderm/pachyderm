package testpachd

import (
	"context"
	"net"
	"reflect"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	version "github.com/pachyderm/pachyderm/src/client/version/versionpb"
)

// linkServers can be used to default a mock server to make calls to a real api
// server. Due to some reflection shenanigans, mockServerPtr must explicitly be
// a pointer to the mock server instance.
func linkServers(mockServerPtr interface{}, realServer interface{}) {
	mockValue := reflect.ValueOf(mockServerPtr).Elem()
	realValue := reflect.ValueOf(realServer)
	mockType := mockValue.Type()
	for i := 0; i < mockType.NumField(); i++ {
		field := mockType.Field(i)
		if field.Name != "api" {
			mock := mockValue.FieldByName(field.Name)
			realMethod := realValue.MethodByName(field.Name)

			// We need a pointer to the mock field to call the right method
			mockPtr := reflect.New(reflect.PtrTo(mock.Type()))
			mockPtrValue := mockPtr.Elem()
			mockPtrValue.Set(mock.Addr())

			useFn := mockPtrValue.MethodByName("Use")
			useFn.Call([]reflect.Value{realMethod})
		}
	}
}

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

type adminServerAPI struct {
	mock *mockAdminServer
}

type mockAdminServer struct {
	api             adminServerAPI
	Extract         mockExtract
	ExtractPipeline mockExtractPipeline
	Restore         mockRestore
	InspectCluster  mockInspectCluster
}

func (api *adminServerAPI) Extract(req *admin.ExtractRequest, serv admin.API_ExtractServer) error {
	if api.mock.Extract.handler != nil {
		return api.mock.Extract.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock: admin.Extract")
}
func (api *adminServerAPI) ExtractPipeline(ctx context.Context, req *admin.ExtractPipelineRequest) (*admin.Op, error) {
	if api.mock.ExtractPipeline.handler != nil {
		return api.mock.ExtractPipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock admin.ExtractPipeline")
}
func (api *adminServerAPI) Restore(serv admin.API_RestoreServer) error {
	if api.mock.Restore.handler != nil {
		return api.mock.Restore.handler(serv)
	}
	return errors.Errorf("unhandled pachd mock admin.Restore")
}
func (api *adminServerAPI) InspectCluster(ctx context.Context, req *types.Empty) (*admin.ClusterInfo, error) {
	if api.mock.InspectCluster.handler != nil {
		return api.mock.InspectCluster.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock admin.InspectCluster")
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
func (mock *mockRevokeAuthToken) Use(cb revokeAuthTokenFunc)       { mock.handler = cb }
func (mock *mockSetGroupsForUser) Use(cb setGroupsForUserFunc)     { mock.handler = cb }
func (mock *mockModifyMembers) Use(cb modifyMembersFunc)           { mock.handler = cb }
func (mock *mockGetGroups) Use(cb getGroupsFunc)                   { mock.handler = cb }
func (mock *mockGetUsers) Use(cb getUsersFunc)                     { mock.handler = cb }
func (mock *mockGetOneTimePassword) Use(cb getOneTimePasswordFunc) { mock.handler = cb }

type authServerAPI struct {
	mock *mockAuthServer
}

type mockAuthServer struct {
	api                authServerAPI
	Activate           mockActivateAuth
	Deactivate         mockDeactivateAuth
	GetConfiguration   mockGetConfiguration
	SetConfiguration   mockSetConfiguration
	GetAdmins          mockGetAdmins
	ModifyAdmins       mockModifyAdmins
	Authenticate       mockAuthenticate
	Authorize          mockAuthorize
	WhoAmI             mockWhoAmI
	GetScope           mockGetScope
	SetScope           mockSetScope
	GetACL             mockGetACL
	SetACL             mockSetACL
	GetAuthToken       mockGetAuthToken
	ExtendAuthToken    mockExtendAuthToken
	RevokeAuthToken    mockRevokeAuthToken
	SetGroupsForUser   mockSetGroupsForUser
	ModifyMembers      mockModifyMembers
	GetGroups          mockGetGroups
	GetUsers           mockGetUsers
	GetOneTimePassword mockGetOneTimePassword
}

func (api *authServerAPI) Activate(ctx context.Context, req *auth.ActivateRequest) (*auth.ActivateResponse, error) {
	if api.mock.Activate.handler != nil {
		return api.mock.Activate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.Activate")
}
func (api *authServerAPI) Deactivate(ctx context.Context, req *auth.DeactivateRequest) (*auth.DeactivateResponse, error) {
	if api.mock.Deactivate.handler != nil {
		return api.mock.Deactivate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.Deactivate")
}
func (api *authServerAPI) GetConfiguration(ctx context.Context, req *auth.GetConfigurationRequest) (*auth.GetConfigurationResponse, error) {
	if api.mock.GetConfiguration.handler != nil {
		return api.mock.GetConfiguration.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetConfiguration")
}
func (api *authServerAPI) SetConfiguration(ctx context.Context, req *auth.SetConfigurationRequest) (*auth.SetConfigurationResponse, error) {
	if api.mock.SetConfiguration.handler != nil {
		return api.mock.SetConfiguration.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.SetConfiguration")
}
func (api *authServerAPI) GetAdmins(ctx context.Context, req *auth.GetAdminsRequest) (*auth.GetAdminsResponse, error) {
	if api.mock.GetAdmins.handler != nil {
		return api.mock.GetAdmins.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetAdmins")
}
func (api *authServerAPI) ModifyAdmins(ctx context.Context, req *auth.ModifyAdminsRequest) (*auth.ModifyAdminsResponse, error) {
	if api.mock.ModifyAdmins.handler != nil {
		return api.mock.ModifyAdmins.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.ModifyAdmins")
}
func (api *authServerAPI) Authenticate(ctx context.Context, req *auth.AuthenticateRequest) (*auth.AuthenticateResponse, error) {
	if api.mock.Authenticate.handler != nil {
		return api.mock.Authenticate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.Authenticate")
}
func (api *authServerAPI) Authorize(ctx context.Context, req *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	if api.mock.Authorize.handler != nil {
		return api.mock.Authorize.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.Authorize")
}
func (api *authServerAPI) WhoAmI(ctx context.Context, req *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error) {
	if api.mock.WhoAmI.handler != nil {
		return api.mock.WhoAmI.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.WhoAmI")
}
func (api *authServerAPI) GetScope(ctx context.Context, req *auth.GetScopeRequest) (*auth.GetScopeResponse, error) {
	if api.mock.GetScope.handler != nil {
		return api.mock.GetScope.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetScope")
}
func (api *authServerAPI) SetScope(ctx context.Context, req *auth.SetScopeRequest) (*auth.SetScopeResponse, error) {
	if api.mock.SetScope.handler != nil {
		return api.mock.SetScope.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.SetScope")
}
func (api *authServerAPI) GetACL(ctx context.Context, req *auth.GetACLRequest) (*auth.GetACLResponse, error) {
	if api.mock.GetACL.handler != nil {
		return api.mock.GetACL.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetACL")
}
func (api *authServerAPI) SetACL(ctx context.Context, req *auth.SetACLRequest) (*auth.SetACLResponse, error) {
	if api.mock.SetACL.handler != nil {
		return api.mock.SetACL.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.SetACL")
}
func (api *authServerAPI) GetAuthToken(ctx context.Context, req *auth.GetAuthTokenRequest) (*auth.GetAuthTokenResponse, error) {
	if api.mock.GetAuthToken.handler != nil {
		return api.mock.GetAuthToken.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetAuthToken")
}
func (api *authServerAPI) ExtendAuthToken(ctx context.Context, req *auth.ExtendAuthTokenRequest) (*auth.ExtendAuthTokenResponse, error) {
	if api.mock.ExtendAuthToken.handler != nil {
		return api.mock.ExtendAuthToken.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.ExtendAuthToken")
}
func (api *authServerAPI) RevokeAuthToken(ctx context.Context, req *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error) {
	if api.mock.RevokeAuthToken.handler != nil {
		return api.mock.RevokeAuthToken.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.RevokeAuthToken")
}
func (api *authServerAPI) SetGroupsForUser(ctx context.Context, req *auth.SetGroupsForUserRequest) (*auth.SetGroupsForUserResponse, error) {
	if api.mock.SetGroupsForUser.handler != nil {
		return api.mock.SetGroupsForUser.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.SetGroupsForUser")
}
func (api *authServerAPI) ModifyMembers(ctx context.Context, req *auth.ModifyMembersRequest) (*auth.ModifyMembersResponse, error) {
	if api.mock.ModifyMembers.handler != nil {
		return api.mock.ModifyMembers.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.ModifyMembers")
}
func (api *authServerAPI) GetGroups(ctx context.Context, req *auth.GetGroupsRequest) (*auth.GetGroupsResponse, error) {
	if api.mock.GetGroups.handler != nil {
		return api.mock.GetGroups.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetGroups")
}
func (api *authServerAPI) GetUsers(ctx context.Context, req *auth.GetUsersRequest) (*auth.GetUsersResponse, error) {
	if api.mock.GetUsers.handler != nil {
		return api.mock.GetUsers.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetUsers")
}
func (api *authServerAPI) GetOneTimePassword(ctx context.Context, req *auth.GetOneTimePasswordRequest) (*auth.GetOneTimePasswordResponse, error) {
	if api.mock.GetOneTimePassword.handler != nil {
		return api.mock.GetOneTimePassword.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock auth.GetOneTimePassword")
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

type enterpriseServerAPI struct {
	mock *mockEnterpriseServer
}

type mockEnterpriseServer struct {
	api        enterpriseServerAPI
	Activate   mockActivateEnterprise
	GetState   mockGetState
	Deactivate mockDeactivateEnterprise
}

func (api *enterpriseServerAPI) Activate(ctx context.Context, req *enterprise.ActivateRequest) (*enterprise.ActivateResponse, error) {
	if api.mock.Activate.handler != nil {
		return api.mock.Activate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.Activate")
}
func (api *enterpriseServerAPI) GetState(ctx context.Context, req *enterprise.GetStateRequest) (*enterprise.GetStateResponse, error) {
	if api.mock.GetState.handler != nil {
		return api.mock.GetState.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.GetState")
}
func (api *enterpriseServerAPI) Deactivate(ctx context.Context, req *enterprise.DeactivateRequest) (*enterprise.DeactivateResponse, error) {
	if api.mock.Deactivate.handler != nil {
		return api.mock.Deactivate.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock enterprise.Deactivate")
}

/* PFS Server Mocks */

type createRepoFunc func(context.Context, *pfs.CreateRepoRequest) (*types.Empty, error)
type inspectRepoFunc func(context.Context, *pfs.InspectRepoRequest) (*pfs.RepoInfo, error)
type listRepoFunc func(context.Context, *pfs.ListRepoRequest) (*pfs.ListRepoResponse, error)
type deleteRepoFunc func(context.Context, *pfs.DeleteRepoRequest) (*types.Empty, error)
type startCommitFunc func(context.Context, *pfs.StartCommitRequest) (*pfs.Commit, error)
type finishCommitFunc func(context.Context, *pfs.FinishCommitRequest) (*types.Empty, error)
type inspectCommitFunc func(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error)
type listCommitFunc func(context.Context, *pfs.ListCommitRequest) (*pfs.CommitInfos, error)
type listCommitStreamFunc func(*pfs.ListCommitRequest, pfs.API_ListCommitStreamServer) error
type deleteCommitFunc func(context.Context, *pfs.DeleteCommitRequest) (*types.Empty, error)
type flushCommitFunc func(*pfs.FlushCommitRequest, pfs.API_FlushCommitServer) error
type subscribeCommitFunc func(*pfs.SubscribeCommitRequest, pfs.API_SubscribeCommitServer) error
type buildCommitFunc func(context.Context, *pfs.BuildCommitRequest) (*pfs.Commit, error)
type createBranchFunc func(context.Context, *pfs.CreateBranchRequest) (*types.Empty, error)
type inspectBranchFunc func(context.Context, *pfs.InspectBranchRequest) (*pfs.BranchInfo, error)
type listBranchFunc func(context.Context, *pfs.ListBranchRequest) (*pfs.BranchInfos, error)
type deleteBranchFunc func(context.Context, *pfs.DeleteBranchRequest) (*types.Empty, error)
type putFileFunc func(pfs.API_PutFileServer) error
type copyFileFunc func(context.Context, *pfs.CopyFileRequest) (*types.Empty, error)
type getFileFunc func(*pfs.GetFileRequest, pfs.API_GetFileServer) error
type inspectFileFunc func(context.Context, *pfs.InspectFileRequest) (*pfs.FileInfo, error)
type listFileFunc func(context.Context, *pfs.ListFileRequest) (*pfs.FileInfos, error)
type listFileStreamFunc func(*pfs.ListFileRequest, pfs.API_ListFileStreamServer) error
type walkFileFunc func(*pfs.WalkFileRequest, pfs.API_WalkFileServer) error
type globFileFunc func(context.Context, *pfs.GlobFileRequest) (*pfs.FileInfos, error)
type globFileStreamFunc func(*pfs.GlobFileRequest, pfs.API_GlobFileStreamServer) error
type diffFileFunc func(context.Context, *pfs.DiffFileRequest) (*pfs.DiffFileResponse, error)
type deleteFileFunc func(context.Context, *pfs.DeleteFileRequest) (*types.Empty, error)
type deleteAllPFSFunc func(context.Context, *types.Empty) (*types.Empty, error)
type fsckFunc func(*pfs.FsckRequest, pfs.API_FsckServer) error
type putTarFunc func(pfs.API_PutTarServer) error
type getTarFunc func(*pfs.GetTarRequest, pfs.API_GetTarServer) error
type getTarConditionalFunc func(pfs.API_GetTarConditionalServer) error
type listFileNSFunc func(*pfs.ListFileRequest, pfs.API_ListFileNSServer) error

type mockCreateRepo struct{ handler createRepoFunc }
type mockInspectRepo struct{ handler inspectRepoFunc }
type mockListRepo struct{ handler listRepoFunc }
type mockDeleteRepo struct{ handler deleteRepoFunc }
type mockStartCommit struct{ handler startCommitFunc }
type mockFinishCommit struct{ handler finishCommitFunc }
type mockInspectCommit struct{ handler inspectCommitFunc }
type mockListCommit struct{ handler listCommitFunc }
type mockListCommitStream struct{ handler listCommitStreamFunc }
type mockDeleteCommit struct{ handler deleteCommitFunc }
type mockFlushCommit struct{ handler flushCommitFunc }
type mockSubscribeCommit struct{ handler subscribeCommitFunc }
type mockBuildCommit struct{ handler buildCommitFunc }
type mockCreateBranch struct{ handler createBranchFunc }
type mockInspectBranch struct{ handler inspectBranchFunc }
type mockListBranch struct{ handler listBranchFunc }
type mockDeleteBranch struct{ handler deleteBranchFunc }
type mockPutFile struct{ handler putFileFunc }
type mockCopyFile struct{ handler copyFileFunc }
type mockGetFile struct{ handler getFileFunc }
type mockInspectFile struct{ handler inspectFileFunc }
type mockListFile struct{ handler listFileFunc }
type mockListFileStream struct{ handler listFileStreamFunc }
type mockWalkFile struct{ handler walkFileFunc }
type mockGlobFile struct{ handler globFileFunc }
type mockGlobFileStream struct{ handler globFileStreamFunc }
type mockDiffFile struct{ handler diffFileFunc }
type mockDeleteFile struct{ handler deleteFileFunc }
type mockDeleteAllPFS struct{ handler deleteAllPFSFunc }
type mockFsck struct{ handler fsckFunc }
type mockPutTar struct{ handler putTarFunc }
type mockGetTar struct{ handler getTarFunc }
type mockGetTarConditional struct{ handler getTarConditionalFunc }
type mockListFileNS struct{ handler listFileNSFunc }

func (mock *mockCreateRepo) Use(cb createRepoFunc)               { mock.handler = cb }
func (mock *mockInspectRepo) Use(cb inspectRepoFunc)             { mock.handler = cb }
func (mock *mockListRepo) Use(cb listRepoFunc)                   { mock.handler = cb }
func (mock *mockDeleteRepo) Use(cb deleteRepoFunc)               { mock.handler = cb }
func (mock *mockStartCommit) Use(cb startCommitFunc)             { mock.handler = cb }
func (mock *mockFinishCommit) Use(cb finishCommitFunc)           { mock.handler = cb }
func (mock *mockInspectCommit) Use(cb inspectCommitFunc)         { mock.handler = cb }
func (mock *mockListCommit) Use(cb listCommitFunc)               { mock.handler = cb }
func (mock *mockListCommitStream) Use(cb listCommitStreamFunc)   { mock.handler = cb }
func (mock *mockDeleteCommit) Use(cb deleteCommitFunc)           { mock.handler = cb }
func (mock *mockFlushCommit) Use(cb flushCommitFunc)             { mock.handler = cb }
func (mock *mockSubscribeCommit) Use(cb subscribeCommitFunc)     { mock.handler = cb }
func (mock *mockBuildCommit) Use(cb buildCommitFunc)             { mock.handler = cb }
func (mock *mockCreateBranch) Use(cb createBranchFunc)           { mock.handler = cb }
func (mock *mockInspectBranch) Use(cb inspectBranchFunc)         { mock.handler = cb }
func (mock *mockListBranch) Use(cb listBranchFunc)               { mock.handler = cb }
func (mock *mockDeleteBranch) Use(cb deleteBranchFunc)           { mock.handler = cb }
func (mock *mockPutFile) Use(cb putFileFunc)                     { mock.handler = cb }
func (mock *mockCopyFile) Use(cb copyFileFunc)                   { mock.handler = cb }
func (mock *mockGetFile) Use(cb getFileFunc)                     { mock.handler = cb }
func (mock *mockInspectFile) Use(cb inspectFileFunc)             { mock.handler = cb }
func (mock *mockListFile) Use(cb listFileFunc)                   { mock.handler = cb }
func (mock *mockListFileStream) Use(cb listFileStreamFunc)       { mock.handler = cb }
func (mock *mockWalkFile) Use(cb walkFileFunc)                   { mock.handler = cb }
func (mock *mockGlobFile) Use(cb globFileFunc)                   { mock.handler = cb }
func (mock *mockGlobFileStream) Use(cb globFileStreamFunc)       { mock.handler = cb }
func (mock *mockDiffFile) Use(cb diffFileFunc)                   { mock.handler = cb }
func (mock *mockDeleteFile) Use(cb deleteFileFunc)               { mock.handler = cb }
func (mock *mockDeleteAllPFS) Use(cb deleteAllPFSFunc)           { mock.handler = cb }
func (mock *mockFsck) Use(cb fsckFunc)                           { mock.handler = cb }
func (mock *mockPutTar) Use(cb putTarFunc)                       { mock.handler = cb }
func (mock *mockGetTar) Use(cb getTarFunc)                       { mock.handler = cb }
func (mock *mockGetTarConditional) Use(cb getTarConditionalFunc) { mock.handler = cb }
func (mock *mockListFileNS) Use(cb listFileNSFunc)               { mock.handler = cb }

type pfsServerAPI struct {
	mock *mockPFSServer
}

type mockPFSServer struct {
	api               pfsServerAPI
	CreateRepo        mockCreateRepo
	InspectRepo       mockInspectRepo
	ListRepo          mockListRepo
	DeleteRepo        mockDeleteRepo
	StartCommit       mockStartCommit
	FinishCommit      mockFinishCommit
	InspectCommit     mockInspectCommit
	ListCommit        mockListCommit
	ListCommitStream  mockListCommitStream
	DeleteCommit      mockDeleteCommit
	FlushCommit       mockFlushCommit
	SubscribeCommit   mockSubscribeCommit
	BuildCommit       mockBuildCommit
	CreateBranch      mockCreateBranch
	InspectBranch     mockInspectBranch
	ListBranch        mockListBranch
	DeleteBranch      mockDeleteBranch
	PutFile           mockPutFile
	CopyFile          mockCopyFile
	GetFile           mockGetFile
	InspectFile       mockInspectFile
	ListFile          mockListFile
	ListFileStream    mockListFileStream
	WalkFile          mockWalkFile
	GlobFile          mockGlobFile
	GlobFileStream    mockGlobFileStream
	DiffFile          mockDiffFile
	DeleteFile        mockDeleteFile
	DeleteAll         mockDeleteAllPFS
	Fsck              mockFsck
	PutTar            mockPutTar
	GetTar            mockGetTar
	GetTarConditional mockGetTarConditional
	ListFileNS        mockListFileNS
}

func (api *pfsServerAPI) CreateRepo(ctx context.Context, req *pfs.CreateRepoRequest) (*types.Empty, error) {
	if api.mock.CreateRepo.handler != nil {
		return api.mock.CreateRepo.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.CreateRepo")
}
func (api *pfsServerAPI) InspectRepo(ctx context.Context, req *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	if api.mock.InspectRepo.handler != nil {
		return api.mock.InspectRepo.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.InspectRepo")
}
func (api *pfsServerAPI) ListRepo(ctx context.Context, req *pfs.ListRepoRequest) (*pfs.ListRepoResponse, error) {
	if api.mock.ListRepo.handler != nil {
		return api.mock.ListRepo.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.ListRepo")
}
func (api *pfsServerAPI) DeleteRepo(ctx context.Context, req *pfs.DeleteRepoRequest) (*types.Empty, error) {
	if api.mock.DeleteRepo.handler != nil {
		return api.mock.DeleteRepo.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DeleteRepo")
}
func (api *pfsServerAPI) StartCommit(ctx context.Context, req *pfs.StartCommitRequest) (*pfs.Commit, error) {
	if api.mock.StartCommit.handler != nil {
		return api.mock.StartCommit.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.StartCommit")
}
func (api *pfsServerAPI) FinishCommit(ctx context.Context, req *pfs.FinishCommitRequest) (*types.Empty, error) {
	if api.mock.FinishCommit.handler != nil {
		return api.mock.FinishCommit.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.FinishCommit")
}
func (api *pfsServerAPI) InspectCommit(ctx context.Context, req *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	if api.mock.InspectCommit.handler != nil {
		return api.mock.InspectCommit.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.InspectCommit")
}
func (api *pfsServerAPI) ListCommit(ctx context.Context, req *pfs.ListCommitRequest) (*pfs.CommitInfos, error) {
	if api.mock.ListCommit.handler != nil {
		return api.mock.ListCommit.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.ListCommit")
}
func (api *pfsServerAPI) ListCommitStream(req *pfs.ListCommitRequest, serv pfs.API_ListCommitStreamServer) error {
	if api.mock.ListCommitStream.handler != nil {
		return api.mock.ListCommitStream.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.ListCommitStream")
}
func (api *pfsServerAPI) DeleteCommit(ctx context.Context, req *pfs.DeleteCommitRequest) (*types.Empty, error) {
	if api.mock.DeleteCommit.handler != nil {
		return api.mock.DeleteCommit.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DeleteCommit")
}
func (api *pfsServerAPI) FlushCommit(req *pfs.FlushCommitRequest, serv pfs.API_FlushCommitServer) error {
	if api.mock.FlushCommit.handler != nil {
		return api.mock.FlushCommit.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.FlushCommit")
}
func (api *pfsServerAPI) SubscribeCommit(req *pfs.SubscribeCommitRequest, serv pfs.API_SubscribeCommitServer) error {
	if api.mock.SubscribeCommit.handler != nil {
		return api.mock.SubscribeCommit.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.SubscribeCommit")
}
func (api *pfsServerAPI) BuildCommit(ctx context.Context, req *pfs.BuildCommitRequest) (*pfs.Commit, error) {
	if api.mock.BuildCommit.handler != nil {
		return api.mock.BuildCommit.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.BuildCommit")
}
func (api *pfsServerAPI) CreateBranch(ctx context.Context, req *pfs.CreateBranchRequest) (*types.Empty, error) {
	if api.mock.CreateBranch.handler != nil {
		return api.mock.CreateBranch.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.CreateBranch")
}
func (api *pfsServerAPI) InspectBranch(ctx context.Context, req *pfs.InspectBranchRequest) (*pfs.BranchInfo, error) {
	if api.mock.InspectBranch.handler != nil {
		return api.mock.InspectBranch.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.InspectBranch")
}
func (api *pfsServerAPI) ListBranch(ctx context.Context, req *pfs.ListBranchRequest) (*pfs.BranchInfos, error) {
	if api.mock.ListBranch.handler != nil {
		return api.mock.ListBranch.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.ListBranch")
}
func (api *pfsServerAPI) DeleteBranch(ctx context.Context, req *pfs.DeleteBranchRequest) (*types.Empty, error) {
	if api.mock.DeleteBranch.handler != nil {
		return api.mock.DeleteBranch.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DeleteBranch")
}
func (api *pfsServerAPI) PutFile(serv pfs.API_PutFileServer) error {
	if api.mock.PutFile.handler != nil {
		return api.mock.PutFile.handler(serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.PutFile")
}
func (api *pfsServerAPI) CopyFile(ctx context.Context, req *pfs.CopyFileRequest) (*types.Empty, error) {
	if api.mock.CopyFile.handler != nil {
		return api.mock.CopyFile.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.CopyFile")
}
func (api *pfsServerAPI) GetFile(req *pfs.GetFileRequest, serv pfs.API_GetFileServer) error {
	if api.mock.GetFile.handler != nil {
		return api.mock.GetFile.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.GetFile")
}
func (api *pfsServerAPI) InspectFile(ctx context.Context, req *pfs.InspectFileRequest) (*pfs.FileInfo, error) {
	if api.mock.InspectFile.handler != nil {
		return api.mock.InspectFile.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.InspectFile")
}
func (api *pfsServerAPI) ListFile(ctx context.Context, req *pfs.ListFileRequest) (*pfs.FileInfos, error) {
	if api.mock.ListFile.handler != nil {
		return api.mock.ListFile.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.ListFile")
}
func (api *pfsServerAPI) ListFileStream(req *pfs.ListFileRequest, serv pfs.API_ListFileStreamServer) error {
	if api.mock.ListFileStream.handler != nil {
		return api.mock.ListFileStream.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.ListFileStream")
}
func (api *pfsServerAPI) WalkFile(req *pfs.WalkFileRequest, serv pfs.API_WalkFileServer) error {
	if api.mock.WalkFile.handler != nil {
		return api.mock.WalkFile.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.WalkFile")
}
func (api *pfsServerAPI) GlobFile(ctx context.Context, req *pfs.GlobFileRequest) (*pfs.FileInfos, error) {
	if api.mock.GlobFile.handler != nil {
		return api.mock.GlobFile.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.GlobFile")
}
func (api *pfsServerAPI) GlobFileStream(req *pfs.GlobFileRequest, serv pfs.API_GlobFileStreamServer) error {
	if api.mock.GlobFileStream.handler != nil {
		return api.mock.GlobFileStream.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.GlobFileStream")
}
func (api *pfsServerAPI) DiffFile(ctx context.Context, req *pfs.DiffFileRequest) (*pfs.DiffFileResponse, error) {
	if api.mock.DiffFile.handler != nil {
		return api.mock.DiffFile.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DiffFile")
}
func (api *pfsServerAPI) DeleteFile(ctx context.Context, req *pfs.DeleteFileRequest) (*types.Empty, error) {
	if api.mock.DeleteFile.handler != nil {
		return api.mock.DeleteFile.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DeleteFile")
}
func (api *pfsServerAPI) DeleteAll(ctx context.Context, req *types.Empty) (*types.Empty, error) {
	if api.mock.DeleteAll.handler != nil {
		return api.mock.DeleteAll.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pfs.DeleteAll")
}
func (api *pfsServerAPI) Fsck(req *pfs.FsckRequest, serv pfs.API_FsckServer) error {
	if api.mock.Fsck.handler != nil {
		return api.mock.Fsck.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.Fsck")
}
func (api *pfsServerAPI) PutTar(serv pfs.API_PutTarServer) error {
	if api.mock.PutTar.handler != nil {
		return api.mock.PutTar.handler(serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.PutTar")
}
func (api *pfsServerAPI) GetTar(req *pfs.GetTarRequest, serv pfs.API_GetTarServer) error {
	if api.mock.GetTar.handler != nil {
		return api.mock.GetTar.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.GetTar")
}
func (api *pfsServerAPI) GetTarConditional(serv pfs.API_GetTarConditionalServer) error {
	if api.mock.GetTarConditional.handler != nil {
		return api.mock.GetTarConditional.handler(serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.GetTarConditional")
}
func (api *pfsServerAPI) ListFileNS(req *pfs.ListFileRequest, serv pfs.API_ListFileNSServer) error {
	if api.mock.ListFileNS.handler != nil {
		return api.mock.ListFileNS.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pfs.ListFileNS")
}

/* PPS Server Mocks */

type createJobFunc func(context.Context, *pps.CreateJobRequest) (*pps.Job, error)
type inspectJobFunc func(context.Context, *pps.InspectJobRequest) (*pps.JobInfo, error)
type listJobFunc func(context.Context, *pps.ListJobRequest) (*pps.JobInfos, error)
type listJobStreamFunc func(*pps.ListJobRequest, pps.API_ListJobStreamServer) error
type flushJobFunc func(*pps.FlushJobRequest, pps.API_FlushJobServer) error
type deleteJobFunc func(context.Context, *pps.DeleteJobRequest) (*types.Empty, error)
type stopJobFunc func(context.Context, *pps.StopJobRequest) (*types.Empty, error)
type updateJobStateFunc func(context.Context, *pps.UpdateJobStateRequest) (*types.Empty, error)
type inspectDatumFunc func(context.Context, *pps.InspectDatumRequest) (*pps.DatumInfo, error)
type listDatumFunc func(context.Context, *pps.ListDatumRequest) (*pps.ListDatumResponse, error)
type listDatumStreamFunc func(*pps.ListDatumRequest, pps.API_ListDatumStreamServer) error
type restartDatumFunc func(context.Context, *pps.RestartDatumRequest) (*types.Empty, error)
type createPipelineFunc func(context.Context, *pps.CreatePipelineRequest) (*types.Empty, error)
type inspectPipelineFunc func(context.Context, *pps.InspectPipelineRequest) (*pps.PipelineInfo, error)
type listPipelineFunc func(context.Context, *pps.ListPipelineRequest) (*pps.PipelineInfos, error)
type deletePipelineFunc func(context.Context, *pps.DeletePipelineRequest) (*types.Empty, error)
type startPipelineFunc func(context.Context, *pps.StartPipelineRequest) (*types.Empty, error)
type stopPipelineFunc func(context.Context, *pps.StopPipelineRequest) (*types.Empty, error)
type runPipelineFunc func(context.Context, *pps.RunPipelineRequest) (*types.Empty, error)
type runCronFunc func(context.Context, *pps.RunCronRequest) (*types.Empty, error)
type createSecretFunc func(context.Context, *pps.CreateSecretRequest) (*types.Empty, error)
type deleteSecretFunc func(context.Context, *pps.DeleteSecretRequest) (*types.Empty, error)
type inspectSecretFunc func(context.Context, *pps.InspectSecretRequest) (*pps.SecretInfo, error)
type listSecretFunc func(context.Context, *types.Empty) (*pps.SecretInfos, error)
type deleteAllPPSFunc func(context.Context, *types.Empty) (*types.Empty, error)
type getLogsFunc func(*pps.GetLogsRequest, pps.API_GetLogsServer) error
type garbageCollectFunc func(context.Context, *pps.GarbageCollectRequest) (*pps.GarbageCollectResponse, error)
type activateAuthPPSFunc func(context.Context, *pps.ActivateAuthRequest) (*pps.ActivateAuthResponse, error)

type mockCreateJob struct{ handler createJobFunc }
type mockInspectJob struct{ handler inspectJobFunc }
type mockListJob struct{ handler listJobFunc }
type mockListJobStream struct{ handler listJobStreamFunc }
type mockFlushJob struct{ handler flushJobFunc }
type mockDeleteJob struct{ handler deleteJobFunc }
type mockStopJob struct{ handler stopJobFunc }
type mockUpdateJobState struct{ handler updateJobStateFunc }
type mockInspectDatum struct{ handler inspectDatumFunc }
type mockListDatum struct{ handler listDatumFunc }
type mockListDatumStream struct{ handler listDatumStreamFunc }
type mockRestartDatum struct{ handler restartDatumFunc }
type mockCreatePipeline struct{ handler createPipelineFunc }
type mockInspectPipeline struct{ handler inspectPipelineFunc }
type mockListPipeline struct{ handler listPipelineFunc }
type mockDeletePipeline struct{ handler deletePipelineFunc }
type mockStartPipeline struct{ handler startPipelineFunc }
type mockStopPipeline struct{ handler stopPipelineFunc }
type mockRunPipeline struct{ handler runPipelineFunc }
type mockRunCron struct{ handler runCronFunc }
type mockCreateSecret struct{ handler createSecretFunc }
type mockDeleteSecret struct{ handler deleteSecretFunc }
type mockInspectSecret struct{ handler inspectSecretFunc }
type mockListSecret struct{ handler listSecretFunc }
type mockDeleteAllPPS struct{ handler deleteAllPPSFunc }
type mockGetLogs struct{ handler getLogsFunc }
type mockGarbageCollect struct{ handler garbageCollectFunc }
type mockActivateAuthPPS struct{ handler activateAuthPPSFunc }

func (mock *mockCreateJob) Use(cb createJobFunc)             { mock.handler = cb }
func (mock *mockInspectJob) Use(cb inspectJobFunc)           { mock.handler = cb }
func (mock *mockListJob) Use(cb listJobFunc)                 { mock.handler = cb }
func (mock *mockListJobStream) Use(cb listJobStreamFunc)     { mock.handler = cb }
func (mock *mockFlushJob) Use(cb flushJobFunc)               { mock.handler = cb }
func (mock *mockDeleteJob) Use(cb deleteJobFunc)             { mock.handler = cb }
func (mock *mockStopJob) Use(cb stopJobFunc)                 { mock.handler = cb }
func (mock *mockUpdateJobState) Use(cb updateJobStateFunc)   { mock.handler = cb }
func (mock *mockInspectDatum) Use(cb inspectDatumFunc)       { mock.handler = cb }
func (mock *mockListDatum) Use(cb listDatumFunc)             { mock.handler = cb }
func (mock *mockListDatumStream) Use(cb listDatumStreamFunc) { mock.handler = cb }
func (mock *mockRestartDatum) Use(cb restartDatumFunc)       { mock.handler = cb }
func (mock *mockCreatePipeline) Use(cb createPipelineFunc)   { mock.handler = cb }
func (mock *mockInspectPipeline) Use(cb inspectPipelineFunc) { mock.handler = cb }
func (mock *mockListPipeline) Use(cb listPipelineFunc)       { mock.handler = cb }
func (mock *mockDeletePipeline) Use(cb deletePipelineFunc)   { mock.handler = cb }
func (mock *mockStartPipeline) Use(cb startPipelineFunc)     { mock.handler = cb }
func (mock *mockStopPipeline) Use(cb stopPipelineFunc)       { mock.handler = cb }
func (mock *mockRunPipeline) Use(cb runPipelineFunc)         { mock.handler = cb }
func (mock *mockRunCron) Use(cb runCronFunc)                 { mock.handler = cb }
func (mock *mockCreateSecret) Use(cb createSecretFunc)       { mock.handler = cb }
func (mock *mockDeleteSecret) Use(cb deleteSecretFunc)       { mock.handler = cb }
func (mock *mockInspectSecret) Use(cb inspectSecretFunc)     { mock.handler = cb }
func (mock *mockListSecret) Use(cb listSecretFunc)           { mock.handler = cb }
func (mock *mockDeleteAllPPS) Use(cb deleteAllPPSFunc)       { mock.handler = cb }
func (mock *mockGetLogs) Use(cb getLogsFunc)                 { mock.handler = cb }
func (mock *mockGarbageCollect) Use(cb garbageCollectFunc)   { mock.handler = cb }
func (mock *mockActivateAuthPPS) Use(cb activateAuthPPSFunc) { mock.handler = cb }

type ppsServerAPI struct {
	mock *mockPPSServer
}

type mockPPSServer struct {
	api             ppsServerAPI
	CreateJob       mockCreateJob
	InspectJob      mockInspectJob
	ListJob         mockListJob
	ListJobStream   mockListJobStream
	FlushJob        mockFlushJob
	DeleteJob       mockDeleteJob
	StopJob         mockStopJob
	UpdateJobState  mockUpdateJobState
	InspectDatum    mockInspectDatum
	ListDatum       mockListDatum
	ListDatumStream mockListDatumStream
	RestartDatum    mockRestartDatum
	CreatePipeline  mockCreatePipeline
	InspectPipeline mockInspectPipeline
	ListPipeline    mockListPipeline
	DeletePipeline  mockDeletePipeline
	StartPipeline   mockStartPipeline
	StopPipeline    mockStopPipeline
	RunPipeline     mockRunPipeline
	RunCron         mockRunCron
	CreateSecret    mockCreateSecret
	DeleteSecret    mockDeleteSecret
	InspectSecret   mockInspectSecret
	ListSecret      mockListSecret
	DeleteAll       mockDeleteAllPPS
	GetLogs         mockGetLogs
	GarbageCollect  mockGarbageCollect
	ActivateAuth    mockActivateAuthPPS
}

func (api *ppsServerAPI) CreateJob(ctx context.Context, req *pps.CreateJobRequest) (*pps.Job, error) {
	if api.mock.CreateJob.handler != nil {
		return api.mock.CreateJob.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.CreateJob")
}
func (api *ppsServerAPI) InspectJob(ctx context.Context, req *pps.InspectJobRequest) (*pps.JobInfo, error) {
	if api.mock.InspectJob.handler != nil {
		return api.mock.InspectJob.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.InspectJob")
}
func (api *ppsServerAPI) ListJob(ctx context.Context, req *pps.ListJobRequest) (*pps.JobInfos, error) {
	if api.mock.ListJob.handler != nil {
		return api.mock.ListJob.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.ListJob")
}
func (api *ppsServerAPI) ListJobStream(req *pps.ListJobRequest, serv pps.API_ListJobStreamServer) error {
	if api.mock.ListJobStream.handler != nil {
		return api.mock.ListJobStream.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pps.ListJobStream")
}
func (api *ppsServerAPI) FlushJob(req *pps.FlushJobRequest, serv pps.API_FlushJobServer) error {
	if api.mock.FlushJob.handler != nil {
		return api.mock.FlushJob.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pps.FlushJob")
}
func (api *ppsServerAPI) DeleteJob(ctx context.Context, req *pps.DeleteJobRequest) (*types.Empty, error) {
	if api.mock.DeleteJob.handler != nil {
		return api.mock.DeleteJob.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.DeleteJob")
}
func (api *ppsServerAPI) UpdateJobState(ctx context.Context, req *pps.UpdateJobStateRequest) (*types.Empty, error) {
	if api.mock.UpdateJobState.handler != nil {
		return api.mock.UpdateJobState.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.UpdateJobState")
}
func (api *ppsServerAPI) StopJob(ctx context.Context, req *pps.StopJobRequest) (*types.Empty, error) {
	if api.mock.StopJob.handler != nil {
		return api.mock.StopJob.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.StopJob")
}
func (api *ppsServerAPI) InspectDatum(ctx context.Context, req *pps.InspectDatumRequest) (*pps.DatumInfo, error) {
	if api.mock.InspectDatum.handler != nil {
		return api.mock.InspectDatum.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.InspectDatum")
}
func (api *ppsServerAPI) ListDatum(ctx context.Context, req *pps.ListDatumRequest) (*pps.ListDatumResponse, error) {
	if api.mock.ListDatum.handler != nil {
		return api.mock.ListDatum.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.ListDatum")
}
func (api *ppsServerAPI) ListDatumStream(req *pps.ListDatumRequest, serv pps.API_ListDatumStreamServer) error {
	if api.mock.ListDatumStream.handler != nil {
		return api.mock.ListDatumStream.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pps.ListDatumStream")
}
func (api *ppsServerAPI) RestartDatum(ctx context.Context, req *pps.RestartDatumRequest) (*types.Empty, error) {
	if api.mock.RestartDatum.handler != nil {
		return api.mock.RestartDatum.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.RestartDatum")
}
func (api *ppsServerAPI) CreatePipeline(ctx context.Context, req *pps.CreatePipelineRequest) (*types.Empty, error) {
	if api.mock.CreatePipeline.handler != nil {
		return api.mock.CreatePipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.CreatePipeline")
}
func (api *ppsServerAPI) InspectPipeline(ctx context.Context, req *pps.InspectPipelineRequest) (*pps.PipelineInfo, error) {
	if api.mock.InspectPipeline.handler != nil {
		return api.mock.InspectPipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.InspectPipeline")
}
func (api *ppsServerAPI) ListPipeline(ctx context.Context, req *pps.ListPipelineRequest) (*pps.PipelineInfos, error) {
	if api.mock.ListPipeline.handler != nil {
		return api.mock.ListPipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.ListPipeline")
}
func (api *ppsServerAPI) DeletePipeline(ctx context.Context, req *pps.DeletePipelineRequest) (*types.Empty, error) {
	if api.mock.DeletePipeline.handler != nil {
		return api.mock.DeletePipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.DeletePipeline")
}
func (api *ppsServerAPI) StartPipeline(ctx context.Context, req *pps.StartPipelineRequest) (*types.Empty, error) {
	if api.mock.StartPipeline.handler != nil {
		return api.mock.StartPipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.StartPipeline")
}
func (api *ppsServerAPI) StopPipeline(ctx context.Context, req *pps.StopPipelineRequest) (*types.Empty, error) {
	if api.mock.StopPipeline.handler != nil {
		return api.mock.StopPipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.StopPipeline")
}
func (api *ppsServerAPI) RunPipeline(ctx context.Context, req *pps.RunPipelineRequest) (*types.Empty, error) {
	if api.mock.RunPipeline.handler != nil {
		return api.mock.RunPipeline.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.RunPipeline")
}
func (api *ppsServerAPI) RunCron(ctx context.Context, req *pps.RunCronRequest) (*types.Empty, error) {
	if api.mock.RunCron.handler != nil {
		return api.mock.RunCron.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.RunCron")
}
func (api *ppsServerAPI) CreateSecret(ctx context.Context, req *pps.CreateSecretRequest) (*types.Empty, error) {
	if api.mock.CreateSecret.handler != nil {
		return api.mock.CreateSecret.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.CreateSecret")
}
func (api *ppsServerAPI) DeleteSecret(ctx context.Context, req *pps.DeleteSecretRequest) (*types.Empty, error) {
	if api.mock.DeleteSecret.handler != nil {
		return api.mock.DeleteSecret.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.DeleteSecret")
}
func (api *ppsServerAPI) InspectSecret(ctx context.Context, req *pps.InspectSecretRequest) (*pps.SecretInfo, error) {
	if api.mock.InspectSecret.handler != nil {
		return api.mock.InspectSecret.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.InspectSecret")
}
func (api *ppsServerAPI) ListSecret(ctx context.Context, in *types.Empty) (*pps.SecretInfos, error) {
	if api.mock.ListSecret.handler != nil {
		return api.mock.ListSecret.handler(ctx, in)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.ListSecret")
}
func (api *ppsServerAPI) DeleteAll(ctx context.Context, req *types.Empty) (*types.Empty, error) {
	if api.mock.DeleteAll.handler != nil {
		return api.mock.DeleteAll.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.DeleteAll")
}
func (api *ppsServerAPI) GetLogs(req *pps.GetLogsRequest, serv pps.API_GetLogsServer) error {
	if api.mock.GetLogs.handler != nil {
		return api.mock.GetLogs.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock pps.GetLogs")
}
func (api *ppsServerAPI) GarbageCollect(ctx context.Context, req *pps.GarbageCollectRequest) (*pps.GarbageCollectResponse, error) {
	if api.mock.GarbageCollect.handler != nil {
		return api.mock.GarbageCollect.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.GarbageCollect")
}
func (api *ppsServerAPI) ActivateAuth(ctx context.Context, req *pps.ActivateAuthRequest) (*pps.ActivateAuthResponse, error) {
	if api.mock.ActivateAuth.handler != nil {
		return api.mock.ActivateAuth.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock pps.ActivateAuth")
}

/* Transaction Server Mocks */

type batchTransactionFunc func(context.Context, *transaction.BatchTransactionRequest) (*transaction.TransactionInfo, error)
type startTransactionFunc func(context.Context, *transaction.StartTransactionRequest) (*transaction.Transaction, error)
type inspectTransactionFunc func(context.Context, *transaction.InspectTransactionRequest) (*transaction.TransactionInfo, error)
type deleteTransactionFunc func(context.Context, *transaction.DeleteTransactionRequest) (*types.Empty, error)
type listTransactionFunc func(context.Context, *transaction.ListTransactionRequest) (*transaction.TransactionInfos, error)
type finishTransactionFunc func(context.Context, *transaction.FinishTransactionRequest) (*transaction.TransactionInfo, error)
type deleteAllTransactionFunc func(context.Context, *transaction.DeleteAllRequest) (*types.Empty, error)

type mockBatchTransaction struct{ handler batchTransactionFunc }
type mockStartTransaction struct{ handler startTransactionFunc }
type mockInspectTransaction struct{ handler inspectTransactionFunc }
type mockDeleteTransaction struct{ handler deleteTransactionFunc }
type mockListTransaction struct{ handler listTransactionFunc }
type mockFinishTransaction struct{ handler finishTransactionFunc }
type mockDeleteAllTransaction struct{ handler deleteAllTransactionFunc }

func (mock *mockBatchTransaction) Use(cb batchTransactionFunc)         { mock.handler = cb }
func (mock *mockStartTransaction) Use(cb startTransactionFunc)         { mock.handler = cb }
func (mock *mockInspectTransaction) Use(cb inspectTransactionFunc)     { mock.handler = cb }
func (mock *mockDeleteTransaction) Use(cb deleteTransactionFunc)       { mock.handler = cb }
func (mock *mockListTransaction) Use(cb listTransactionFunc)           { mock.handler = cb }
func (mock *mockFinishTransaction) Use(cb finishTransactionFunc)       { mock.handler = cb }
func (mock *mockDeleteAllTransaction) Use(cb deleteAllTransactionFunc) { mock.handler = cb }

type transactionServerAPI struct {
	mock *mockTransactionServer
}

type mockTransactionServer struct {
	api                transactionServerAPI
	BatchTransaction   mockBatchTransaction
	StartTransaction   mockStartTransaction
	InspectTransaction mockInspectTransaction
	DeleteTransaction  mockDeleteTransaction
	ListTransaction    mockListTransaction
	FinishTransaction  mockFinishTransaction
	DeleteAll          mockDeleteAllTransaction
}

func (api *transactionServerAPI) BatchTransaction(ctx context.Context, req *transaction.BatchTransactionRequest) (*transaction.TransactionInfo, error) {
	if api.mock.BatchTransaction.handler != nil {
		return api.mock.BatchTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.BatchTransaction")
}
func (api *transactionServerAPI) StartTransaction(ctx context.Context, req *transaction.StartTransactionRequest) (*transaction.Transaction, error) {
	if api.mock.StartTransaction.handler != nil {
		return api.mock.StartTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.StartTransaction")
}
func (api *transactionServerAPI) InspectTransaction(ctx context.Context, req *transaction.InspectTransactionRequest) (*transaction.TransactionInfo, error) {
	if api.mock.InspectTransaction.handler != nil {
		return api.mock.InspectTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.InspectTransaction")
}
func (api *transactionServerAPI) DeleteTransaction(ctx context.Context, req *transaction.DeleteTransactionRequest) (*types.Empty, error) {
	if api.mock.DeleteTransaction.handler != nil {
		return api.mock.DeleteTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.DeleteTransaction")
}
func (api *transactionServerAPI) ListTransaction(ctx context.Context, req *transaction.ListTransactionRequest) (*transaction.TransactionInfos, error) {
	if api.mock.ListTransaction.handler != nil {
		return api.mock.ListTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.ListTransaction")
}
func (api *transactionServerAPI) FinishTransaction(ctx context.Context, req *transaction.FinishTransactionRequest) (*transaction.TransactionInfo, error) {
	if api.mock.FinishTransaction.handler != nil {
		return api.mock.FinishTransaction.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.FinishTransaction")
}
func (api *transactionServerAPI) DeleteAll(ctx context.Context, req *transaction.DeleteAllRequest) (*types.Empty, error) {
	if api.mock.DeleteAll.handler != nil {
		return api.mock.DeleteAll.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock transaction.DeleteAll")
}

/* Version Server Mocks */

type getVersionFunc func(context.Context, *types.Empty) (*version.Version, error)

type mockGetVersion struct{ handler getVersionFunc }

func (mock *mockGetVersion) Use(cb getVersionFunc) { mock.handler = cb }

type versionServerAPI struct {
	mock *mockVersionServer
}

type mockVersionServer struct {
	api        versionServerAPI
	GetVersion mockGetVersion
}

func (api *versionServerAPI) GetVersion(ctx context.Context, req *types.Empty) (*version.Version, error) {
	if api.mock.GetVersion.handler != nil {
		return api.mock.GetVersion.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock version.GetVersion")
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
type putObjDirectFunc func(pfs.ObjectAPI_PutObjDirectServer) error
type getObjDirectFunc func(*pfs.GetObjDirectRequest, pfs.ObjectAPI_GetObjDirectServer) error

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
type mockPutObjDirect struct{ handler putObjDirectFunc }
type mockGetObjDirect struct{ handler getObjDirectFunc }

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
func (mock *mockPutObjDirect) Use(cb putObjDirectFunc)     { mock.handler = cb }
func (mock *mockGetObjDirect) Use(cb getObjDirectFunc)     { mock.handler = cb }

type objectServerAPI struct {
	mock *mockObjectServer
}

type mockObjectServer struct {
	api            objectServerAPI
	PutObject      mockPutObject
	PutObjectSplit mockPutObjectSplit
	PutObjects     mockPutObjects
	CreateObject   mockCreateObject
	GetObject      mockGetObject
	GetObjects     mockGetObjects
	PutBlock       mockPutBlock
	GetBlock       mockGetBlock
	GetBlocks      mockGetBlocks
	ListBlock      mockListBlock
	TagObject      mockTagObject
	InspectObject  mockInspectObject
	CheckObject    mockCheckObject
	ListObjects    mockListObjects
	DeleteObjects  mockDeleteObjects
	GetTag         mockGetTag
	InspectTag     mockInspectTag
	ListTags       mockListTags
	DeleteTags     mockDeleteTags
	Compact        mockCompact
	PutObjDirect   mockPutObjDirect
	GetObjDirect   mockGetObjDirect
}

func (api *objectServerAPI) PutObject(serv pfs.ObjectAPI_PutObjectServer) error {
	if api.mock.PutObject.handler != nil {
		return api.mock.PutObject.handler(serv)
	}
	return errors.Errorf("unhandled pachd mock object.PutObject")
}
func (api *objectServerAPI) PutObjectSplit(serv pfs.ObjectAPI_PutObjectSplitServer) error {
	if api.mock.PutObjectSplit.handler != nil {
		return api.mock.PutObjectSplit.handler(serv)
	}
	return errors.Errorf("unhandled pachd mock object.PutObjectSplit")
}
func (api *objectServerAPI) PutObjects(serv pfs.ObjectAPI_PutObjectsServer) error {
	if api.mock.PutObjects.handler != nil {
		return api.mock.PutObjects.handler(serv)
	}
	return errors.Errorf("unhandled pachd mock object.PutObjects")
}
func (api *objectServerAPI) CreateObject(ctx context.Context, serv *pfs.CreateObjectRequest) (*types.Empty, error) {
	if api.mock.CreateObject.handler != nil {
		return api.mock.CreateObject.handler(ctx, serv)
	}
	return nil, errors.Errorf("unhandled pachd mock object.CreateObject")
}
func (api *objectServerAPI) GetObject(req *pfs.Object, serv pfs.ObjectAPI_GetObjectServer) error {
	if api.mock.GetObject.handler != nil {
		return api.mock.GetObject.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock object.GetObject")
}
func (api *objectServerAPI) GetObjects(req *pfs.GetObjectsRequest, serv pfs.ObjectAPI_GetObjectsServer) error {
	if api.mock.GetObjects.handler != nil {
		return api.mock.GetObjects.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock object.GetObjects")
}
func (api *objectServerAPI) PutBlock(serv pfs.ObjectAPI_PutBlockServer) error {
	if api.mock.PutBlock.handler != nil {
		return api.mock.PutBlock.handler(serv)
	}
	return errors.Errorf("unhandled pachd mock object.PutBlock")
}
func (api *objectServerAPI) GetBlock(req *pfs.GetBlockRequest, serv pfs.ObjectAPI_GetBlockServer) error {
	if api.mock.GetBlock.handler != nil {
		return api.mock.GetBlock.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock object.GetBlock")
}
func (api *objectServerAPI) GetBlocks(req *pfs.GetBlocksRequest, serv pfs.ObjectAPI_GetBlocksServer) error {
	if api.mock.GetBlocks.handler != nil {
		return api.mock.GetBlocks.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock object.GetBlocks")
}
func (api *objectServerAPI) ListBlock(req *pfs.ListBlockRequest, serv pfs.ObjectAPI_ListBlockServer) error {
	if api.mock.ListBlock.handler != nil {
		return api.mock.ListBlock.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock object.ListBlock")
}
func (api *objectServerAPI) TagObject(ctx context.Context, req *pfs.TagObjectRequest) (*types.Empty, error) {
	if api.mock.TagObject.handler != nil {
		return api.mock.TagObject.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock object.TagObject")
}
func (api *objectServerAPI) InspectObject(ctx context.Context, req *pfs.Object) (*pfs.ObjectInfo, error) {
	if api.mock.InspectObject.handler != nil {
		return api.mock.InspectObject.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock object.InspectObject")
}
func (api *objectServerAPI) CheckObject(ctx context.Context, req *pfs.CheckObjectRequest) (*pfs.CheckObjectResponse, error) {
	if api.mock.CheckObject.handler != nil {
		return api.mock.CheckObject.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock object.CheckObject")
}
func (api *objectServerAPI) ListObjects(req *pfs.ListObjectsRequest, serv pfs.ObjectAPI_ListObjectsServer) error {
	if api.mock.ListObjects.handler != nil {
		return api.mock.ListObjects.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock object.ListObjects")
}
func (api *objectServerAPI) DeleteObjects(ctx context.Context, req *pfs.DeleteObjectsRequest) (*pfs.DeleteObjectsResponse, error) {
	if api.mock.DeleteObjects.handler != nil {
		return api.mock.DeleteObjects.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock object.DeleteObjects")
}
func (api *objectServerAPI) GetTag(req *pfs.Tag, serv pfs.ObjectAPI_GetTagServer) error {
	if api.mock.GetTag.handler != nil {
		return api.mock.GetTag.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock object.GetTag")
}
func (api *objectServerAPI) InspectTag(ctx context.Context, req *pfs.Tag) (*pfs.ObjectInfo, error) {
	if api.mock.InspectTag.handler != nil {
		return api.mock.InspectTag.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock object.InspectTag")
}
func (api *objectServerAPI) ListTags(req *pfs.ListTagsRequest, serv pfs.ObjectAPI_ListTagsServer) error {
	if api.mock.ListTags.handler != nil {
		return api.mock.ListTags.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock object.ListTags")
}
func (api *objectServerAPI) DeleteTags(ctx context.Context, req *pfs.DeleteTagsRequest) (*pfs.DeleteTagsResponse, error) {
	if api.mock.DeleteTags.handler != nil {
		return api.mock.DeleteTags.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock object.DeleteTags")
}
func (api *objectServerAPI) Compact(ctx context.Context, req *types.Empty) (*types.Empty, error) {
	if api.mock.Compact.handler != nil {
		return api.mock.Compact.handler(ctx, req)
	}
	return nil, errors.Errorf("unhandled pachd mock object.Compact")
}
func (api *objectServerAPI) PutObjDirect(serv pfs.ObjectAPI_PutObjDirectServer) error {
	if api.mock.PutObjDirect.handler != nil {
		return api.mock.PutObjDirect.handler(serv)
	}
	return errors.Errorf("unhandled pachd mock object.PutObjDirect")
}
func (api *objectServerAPI) GetObjDirect(req *pfs.GetObjDirectRequest, serv pfs.ObjectAPI_GetObjDirectServer) error {
	if api.mock.GetObjDirect.handler != nil {
		return api.mock.GetObjDirect.handler(req, serv)
	}
	return errors.Errorf("unhandled pachd mock object.GetObjDirect")
}

// MockPachd provides an interface for running the interface for a Pachd API
// server locally without any of its dependencies. Tests may mock out specific
// API calls by providing a handler function, and later check information about
// the mocked calls.
type MockPachd struct {
	cancel  context.CancelFunc
	errchan chan error

	Addr net.Addr

	Object      mockObjectServer
	PFS         mockPFSServer
	PPS         mockPPSServer
	Auth        mockAuthServer
	Transaction mockTransactionServer
	Enterprise  mockEnterpriseServer
	Version     mockVersionServer
	Admin       mockAdminServer
}

// NewMockPachd constructs a mock Pachd API server whose behavior can be
// controlled through the MockPachd instance. By default, all API calls will
// error, unless a handler is specified.
func NewMockPachd(ctx context.Context) (*MockPachd, error) {
	mock := &MockPachd{
		errchan: make(chan error),
	}

	ctx, mock.cancel = context.WithCancel(ctx)

	mock.Object.api.mock = &mock.Object
	mock.PFS.api.mock = &mock.PFS
	mock.PPS.api.mock = &mock.PPS
	mock.Auth.api.mock = &mock.Auth
	mock.Transaction.api.mock = &mock.Transaction
	mock.Enterprise.api.mock = &mock.Enterprise
	mock.Version.api.mock = &mock.Version
	mock.Admin.api.mock = &mock.Admin

	server, err := grpcutil.NewServer(ctx, false)
	if err != nil {
		return nil, err
	}

	admin.RegisterAPIServer(server.Server, &mock.Admin.api)
	auth.RegisterAPIServer(server.Server, &mock.Auth.api)
	enterprise.RegisterAPIServer(server.Server, &mock.Enterprise.api)
	pfs.RegisterObjectAPIServer(server.Server, &mock.Object.api)
	pfs.RegisterAPIServer(server.Server, &mock.PFS.api)
	pps.RegisterAPIServer(server.Server, &mock.PPS.api)
	transaction.RegisterAPIServer(server.Server, &mock.Transaction.api)
	version.RegisterAPIServer(server.Server, &mock.Version.api)

	listener, err := server.ListenTCP("localhost", 0)
	if err != nil {
		return nil, err
	}

	go func() {
		mock.errchan <- server.Wait()
		close(mock.errchan)
	}()

	mock.Addr = listener.Addr()

	return mock, nil
}

// Err returns a read-only channel that will receive the first error that occurs
// in the server group (stopping all the servers).
func (mock *MockPachd) Err() <-chan error {
	return mock.errchan
}

// Close will cancel the mock Pachd API server goroutine and return its result
func (mock *MockPachd) Close() error {
	mock.cancel()
	return <-mock.errchan
}
