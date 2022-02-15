// Code generated by pachgen (etc/proto/pachgen). DO NOT EDIT.

package client

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"

	types "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

func unsupportedError(name string) error {
	return errors.Errorf("the '%s' API call is not supported in transactions", name)
}

type unsupportedIdentityBuilderClient struct{}

func (c *unsupportedIdentityBuilderClient) SetIdentityServerConfig(_ context.Context, _ *identity.SetIdentityServerConfigRequest, opts ...grpc.CallOption) (*identity.SetIdentityServerConfigResponse, error) {
	return nil, unsupportedError("SetIdentityServerConfig")
}

func (c *unsupportedIdentityBuilderClient) GetIdentityServerConfig(_ context.Context, _ *identity.GetIdentityServerConfigRequest, opts ...grpc.CallOption) (*identity.GetIdentityServerConfigResponse, error) {
	return nil, unsupportedError("GetIdentityServerConfig")
}

func (c *unsupportedIdentityBuilderClient) CreateIDPConnector(_ context.Context, _ *identity.CreateIDPConnectorRequest, opts ...grpc.CallOption) (*identity.CreateIDPConnectorResponse, error) {
	return nil, unsupportedError("CreateIDPConnector")
}

func (c *unsupportedIdentityBuilderClient) UpdateIDPConnector(_ context.Context, _ *identity.UpdateIDPConnectorRequest, opts ...grpc.CallOption) (*identity.UpdateIDPConnectorResponse, error) {
	return nil, unsupportedError("UpdateIDPConnector")
}

func (c *unsupportedIdentityBuilderClient) ListIDPConnectors(_ context.Context, _ *identity.ListIDPConnectorsRequest, opts ...grpc.CallOption) (*identity.ListIDPConnectorsResponse, error) {
	return nil, unsupportedError("ListIDPConnectors")
}

func (c *unsupportedIdentityBuilderClient) GetIDPConnector(_ context.Context, _ *identity.GetIDPConnectorRequest, opts ...grpc.CallOption) (*identity.GetIDPConnectorResponse, error) {
	return nil, unsupportedError("GetIDPConnector")
}

func (c *unsupportedIdentityBuilderClient) DeleteIDPConnector(_ context.Context, _ *identity.DeleteIDPConnectorRequest, opts ...grpc.CallOption) (*identity.DeleteIDPConnectorResponse, error) {
	return nil, unsupportedError("DeleteIDPConnector")
}

func (c *unsupportedIdentityBuilderClient) CreateOIDCClient(_ context.Context, _ *identity.CreateOIDCClientRequest, opts ...grpc.CallOption) (*identity.CreateOIDCClientResponse, error) {
	return nil, unsupportedError("CreateOIDCClient")
}

func (c *unsupportedIdentityBuilderClient) UpdateOIDCClient(_ context.Context, _ *identity.UpdateOIDCClientRequest, opts ...grpc.CallOption) (*identity.UpdateOIDCClientResponse, error) {
	return nil, unsupportedError("UpdateOIDCClient")
}

func (c *unsupportedIdentityBuilderClient) GetOIDCClient(_ context.Context, _ *identity.GetOIDCClientRequest, opts ...grpc.CallOption) (*identity.GetOIDCClientResponse, error) {
	return nil, unsupportedError("GetOIDCClient")
}

func (c *unsupportedIdentityBuilderClient) ListOIDCClients(_ context.Context, _ *identity.ListOIDCClientsRequest, opts ...grpc.CallOption) (*identity.ListOIDCClientsResponse, error) {
	return nil, unsupportedError("ListOIDCClients")
}

func (c *unsupportedIdentityBuilderClient) DeleteOIDCClient(_ context.Context, _ *identity.DeleteOIDCClientRequest, opts ...grpc.CallOption) (*identity.DeleteOIDCClientResponse, error) {
	return nil, unsupportedError("DeleteOIDCClient")
}

func (c *unsupportedIdentityBuilderClient) DeleteAll(_ context.Context, _ *identity.DeleteAllRequest, opts ...grpc.CallOption) (*identity.DeleteAllResponse, error) {
	return nil, unsupportedError("DeleteAll")
}

type unsupportedAuthBuilderClient struct{}

func (c *unsupportedAuthBuilderClient) Activate(_ context.Context, _ *auth.ActivateRequest, opts ...grpc.CallOption) (*auth.ActivateResponse, error) {
	return nil, unsupportedError("Activate")
}

func (c *unsupportedAuthBuilderClient) Deactivate(_ context.Context, _ *auth.DeactivateRequest, opts ...grpc.CallOption) (*auth.DeactivateResponse, error) {
	return nil, unsupportedError("Deactivate")
}

func (c *unsupportedAuthBuilderClient) GetConfiguration(_ context.Context, _ *auth.GetConfigurationRequest, opts ...grpc.CallOption) (*auth.GetConfigurationResponse, error) {
	return nil, unsupportedError("GetConfiguration")
}

func (c *unsupportedAuthBuilderClient) SetConfiguration(_ context.Context, _ *auth.SetConfigurationRequest, opts ...grpc.CallOption) (*auth.SetConfigurationResponse, error) {
	return nil, unsupportedError("SetConfiguration")
}

func (c *unsupportedAuthBuilderClient) Authenticate(_ context.Context, _ *auth.AuthenticateRequest, opts ...grpc.CallOption) (*auth.AuthenticateResponse, error) {
	return nil, unsupportedError("Authenticate")
}

func (c *unsupportedAuthBuilderClient) Authorize(_ context.Context, _ *auth.AuthorizeRequest, opts ...grpc.CallOption) (*auth.AuthorizeResponse, error) {
	return nil, unsupportedError("Authorize")
}

func (c *unsupportedAuthBuilderClient) GetPermissions(_ context.Context, _ *auth.GetPermissionsRequest, opts ...grpc.CallOption) (*auth.GetPermissionsResponse, error) {
	return nil, unsupportedError("GetPermissions")
}

func (c *unsupportedAuthBuilderClient) GetPermissionsForPrincipal(_ context.Context, _ *auth.GetPermissionsForPrincipalRequest, opts ...grpc.CallOption) (*auth.GetPermissionsResponse, error) {
	return nil, unsupportedError("GetPermissionsForPrincipal")
}

func (c *unsupportedAuthBuilderClient) WhoAmI(_ context.Context, _ *auth.WhoAmIRequest, opts ...grpc.CallOption) (*auth.WhoAmIResponse, error) {
	return nil, unsupportedError("WhoAmI")
}

func (c *unsupportedAuthBuilderClient) GetRolesForPermission(_ context.Context, _ *auth.GetRolesForPermissionRequest, opts ...grpc.CallOption) (*auth.GetRolesForPermissionResponse, error) {
	return nil, unsupportedError("GetRolesForPermission")
}

func (c *unsupportedAuthBuilderClient) ModifyRoleBinding(_ context.Context, _ *auth.ModifyRoleBindingRequest, opts ...grpc.CallOption) (*auth.ModifyRoleBindingResponse, error) {
	return nil, unsupportedError("ModifyRoleBinding")
}

func (c *unsupportedAuthBuilderClient) GetRoleBinding(_ context.Context, _ *auth.GetRoleBindingRequest, opts ...grpc.CallOption) (*auth.GetRoleBindingResponse, error) {
	return nil, unsupportedError("GetRoleBinding")
}

func (c *unsupportedAuthBuilderClient) GetOIDCLogin(_ context.Context, _ *auth.GetOIDCLoginRequest, opts ...grpc.CallOption) (*auth.GetOIDCLoginResponse, error) {
	return nil, unsupportedError("GetOIDCLogin")
}

func (c *unsupportedAuthBuilderClient) GetRobotToken(_ context.Context, _ *auth.GetRobotTokenRequest, opts ...grpc.CallOption) (*auth.GetRobotTokenResponse, error) {
	return nil, unsupportedError("GetRobotToken")
}

func (c *unsupportedAuthBuilderClient) RevokeAuthToken(_ context.Context, _ *auth.RevokeAuthTokenRequest, opts ...grpc.CallOption) (*auth.RevokeAuthTokenResponse, error) {
	return nil, unsupportedError("RevokeAuthToken")
}

func (c *unsupportedAuthBuilderClient) RevokeAuthTokensForUser(_ context.Context, _ *auth.RevokeAuthTokensForUserRequest, opts ...grpc.CallOption) (*auth.RevokeAuthTokensForUserResponse, error) {
	return nil, unsupportedError("RevokeAuthTokensForUser")
}

func (c *unsupportedAuthBuilderClient) SetGroupsForUser(_ context.Context, _ *auth.SetGroupsForUserRequest, opts ...grpc.CallOption) (*auth.SetGroupsForUserResponse, error) {
	return nil, unsupportedError("SetGroupsForUser")
}

func (c *unsupportedAuthBuilderClient) ModifyMembers(_ context.Context, _ *auth.ModifyMembersRequest, opts ...grpc.CallOption) (*auth.ModifyMembersResponse, error) {
	return nil, unsupportedError("ModifyMembers")
}

func (c *unsupportedAuthBuilderClient) GetGroups(_ context.Context, _ *auth.GetGroupsRequest, opts ...grpc.CallOption) (*auth.GetGroupsResponse, error) {
	return nil, unsupportedError("GetGroups")
}

func (c *unsupportedAuthBuilderClient) GetGroupsForPrincipal(_ context.Context, _ *auth.GetGroupsForPrincipalRequest, opts ...grpc.CallOption) (*auth.GetGroupsResponse, error) {
	return nil, unsupportedError("GetGroupsForPrincipal")
}

func (c *unsupportedAuthBuilderClient) GetUsers(_ context.Context, _ *auth.GetUsersRequest, opts ...grpc.CallOption) (*auth.GetUsersResponse, error) {
	return nil, unsupportedError("GetUsers")
}

func (c *unsupportedAuthBuilderClient) ExtractAuthTokens(_ context.Context, _ *auth.ExtractAuthTokensRequest, opts ...grpc.CallOption) (*auth.ExtractAuthTokensResponse, error) {
	return nil, unsupportedError("ExtractAuthTokens")
}

func (c *unsupportedAuthBuilderClient) RestoreAuthToken(_ context.Context, _ *auth.RestoreAuthTokenRequest, opts ...grpc.CallOption) (*auth.RestoreAuthTokenResponse, error) {
	return nil, unsupportedError("RestoreAuthToken")
}

func (c *unsupportedAuthBuilderClient) DeleteExpiredAuthTokens(_ context.Context, _ *auth.DeleteExpiredAuthTokensRequest, opts ...grpc.CallOption) (*auth.DeleteExpiredAuthTokensResponse, error) {
	return nil, unsupportedError("DeleteExpiredAuthTokens")
}

func (c *unsupportedAuthBuilderClient) RotateRootToken(_ context.Context, _ *auth.RotateRootTokenRequest, opts ...grpc.CallOption) (*auth.RotateRootTokenResponse, error) {
	return nil, unsupportedError("RotateRootToken")
}

type unsupportedPfsBuilderClient struct{}

func (c *unsupportedPfsBuilderClient) CreateRepo(_ context.Context, _ *pfs.CreateRepoRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("CreateRepo")
}

func (c *unsupportedPfsBuilderClient) InspectRepo(_ context.Context, _ *pfs.InspectRepoRequest, opts ...grpc.CallOption) (*pfs.RepoInfo, error) {
	return nil, unsupportedError("InspectRepo")
}

func (c *unsupportedPfsBuilderClient) ListRepo(_ context.Context, _ *pfs.ListRepoRequest, opts ...grpc.CallOption) (pfs.API_ListRepoClient, error) {
	return nil, unsupportedError("ListRepo")
}

func (c *unsupportedPfsBuilderClient) DeleteRepo(_ context.Context, _ *pfs.DeleteRepoRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteRepo")
}

func (c *unsupportedPfsBuilderClient) StartCommit(_ context.Context, _ *pfs.StartCommitRequest, opts ...grpc.CallOption) (*pfs.Commit, error) {
	return nil, unsupportedError("StartCommit")
}

func (c *unsupportedPfsBuilderClient) FinishCommit(_ context.Context, _ *pfs.FinishCommitRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("FinishCommit")
}

func (c *unsupportedPfsBuilderClient) ClearCommit(_ context.Context, _ *pfs.ClearCommitRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("ClearCommit")
}

func (c *unsupportedPfsBuilderClient) InspectCommit(_ context.Context, _ *pfs.InspectCommitRequest, opts ...grpc.CallOption) (*pfs.CommitInfo, error) {
	return nil, unsupportedError("InspectCommit")
}

func (c *unsupportedPfsBuilderClient) ListCommit(_ context.Context, _ *pfs.ListCommitRequest, opts ...grpc.CallOption) (pfs.API_ListCommitClient, error) {
	return nil, unsupportedError("ListCommit")
}

func (c *unsupportedPfsBuilderClient) SubscribeCommit(_ context.Context, _ *pfs.SubscribeCommitRequest, opts ...grpc.CallOption) (pfs.API_SubscribeCommitClient, error) {
	return nil, unsupportedError("SubscribeCommit")
}

func (c *unsupportedPfsBuilderClient) InspectCommitSet(_ context.Context, _ *pfs.InspectCommitSetRequest, opts ...grpc.CallOption) (pfs.API_InspectCommitSetClient, error) {
	return nil, unsupportedError("InspectCommitSet")
}

func (c *unsupportedPfsBuilderClient) ListCommitSet(_ context.Context, _ *pfs.ListCommitSetRequest, opts ...grpc.CallOption) (pfs.API_ListCommitSetClient, error) {
	return nil, unsupportedError("ListCommitSet")
}

func (c *unsupportedPfsBuilderClient) SquashCommitSet(_ context.Context, _ *pfs.SquashCommitSetRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("SquashCommitSet")
}

func (c *unsupportedPfsBuilderClient) DropCommitSet(_ context.Context, _ *pfs.DropCommitSetRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DropCommitSet")
}

func (c *unsupportedPfsBuilderClient) CreateBranch(_ context.Context, _ *pfs.CreateBranchRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("CreateBranch")
}

func (c *unsupportedPfsBuilderClient) InspectBranch(_ context.Context, _ *pfs.InspectBranchRequest, opts ...grpc.CallOption) (*pfs.BranchInfo, error) {
	return nil, unsupportedError("InspectBranch")
}

func (c *unsupportedPfsBuilderClient) ListBranch(_ context.Context, _ *pfs.ListBranchRequest, opts ...grpc.CallOption) (pfs.API_ListBranchClient, error) {
	return nil, unsupportedError("ListBranch")
}

func (c *unsupportedPfsBuilderClient) DeleteBranch(_ context.Context, _ *pfs.DeleteBranchRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteBranch")
}

func (c *unsupportedPfsBuilderClient) ModifyFile(_ context.Context, opts ...grpc.CallOption) (pfs.API_ModifyFileClient, error) {
	return nil, unsupportedError("ModifyFile")
}

func (c *unsupportedPfsBuilderClient) GetFile(_ context.Context, _ *pfs.GetFileRequest, opts ...grpc.CallOption) (pfs.API_GetFileClient, error) {
	return nil, unsupportedError("GetFile")
}

func (c *unsupportedPfsBuilderClient) GetFileTAR(_ context.Context, _ *pfs.GetFileRequest, opts ...grpc.CallOption) (pfs.API_GetFileTARClient, error) {
	return nil, unsupportedError("GetFileTAR")
}

func (c *unsupportedPfsBuilderClient) InspectFile(_ context.Context, _ *pfs.InspectFileRequest, opts ...grpc.CallOption) (*pfs.FileInfo, error) {
	return nil, unsupportedError("InspectFile")
}

func (c *unsupportedPfsBuilderClient) ListFile(_ context.Context, _ *pfs.ListFileRequest, opts ...grpc.CallOption) (pfs.API_ListFileClient, error) {
	return nil, unsupportedError("ListFile")
}

func (c *unsupportedPfsBuilderClient) WalkFile(_ context.Context, _ *pfs.WalkFileRequest, opts ...grpc.CallOption) (pfs.API_WalkFileClient, error) {
	return nil, unsupportedError("WalkFile")
}

func (c *unsupportedPfsBuilderClient) GlobFile(_ context.Context, _ *pfs.GlobFileRequest, opts ...grpc.CallOption) (pfs.API_GlobFileClient, error) {
	return nil, unsupportedError("GlobFile")
}

func (c *unsupportedPfsBuilderClient) DiffFile(_ context.Context, _ *pfs.DiffFileRequest, opts ...grpc.CallOption) (pfs.API_DiffFileClient, error) {
	return nil, unsupportedError("DiffFile")
}

func (c *unsupportedPfsBuilderClient) ActivateAuth(_ context.Context, _ *pfs.ActivateAuthRequest, opts ...grpc.CallOption) (*pfs.ActivateAuthResponse, error) {
	return nil, unsupportedError("ActivateAuth")
}

func (c *unsupportedPfsBuilderClient) DeleteAll(_ context.Context, _ *types.Empty, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteAll")
}

func (c *unsupportedPfsBuilderClient) Fsck(_ context.Context, _ *pfs.FsckRequest, opts ...grpc.CallOption) (pfs.API_FsckClient, error) {
	return nil, unsupportedError("Fsck")
}

func (c *unsupportedPfsBuilderClient) CreateFileSet(_ context.Context, opts ...grpc.CallOption) (pfs.API_CreateFileSetClient, error) {
	return nil, unsupportedError("CreateFileSet")
}

func (c *unsupportedPfsBuilderClient) GetFileSet(_ context.Context, _ *pfs.GetFileSetRequest, opts ...grpc.CallOption) (*pfs.CreateFileSetResponse, error) {
	return nil, unsupportedError("GetFileSet")
}

func (c *unsupportedPfsBuilderClient) AddFileSet(_ context.Context, _ *pfs.AddFileSetRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("AddFileSet")
}

func (c *unsupportedPfsBuilderClient) RenewFileSet(_ context.Context, _ *pfs.RenewFileSetRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("RenewFileSet")
}

func (c *unsupportedPfsBuilderClient) ComposeFileSet(_ context.Context, _ *pfs.ComposeFileSetRequest, opts ...grpc.CallOption) (*pfs.CreateFileSetResponse, error) {
	return nil, unsupportedError("ComposeFileSet")
}

func (c *unsupportedPfsBuilderClient) CheckStorage(_ context.Context, _ *pfs.CheckStorageRequest, opts ...grpc.CallOption) (*pfs.CheckStorageResponse, error) {
	return nil, unsupportedError("CheckStorage")
}

func (c *unsupportedPfsBuilderClient) PutCache(_ context.Context, _ *pfs.PutCacheRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("PutCache")
}

func (c *unsupportedPfsBuilderClient) GetCache(_ context.Context, _ *pfs.GetCacheRequest, opts ...grpc.CallOption) (*pfs.GetCacheResponse, error) {
	return nil, unsupportedError("GetCache")
}

func (c *unsupportedPfsBuilderClient) ClearCache(_ context.Context, _ *pfs.ClearCacheRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("ClearCache")
}

func (c *unsupportedPfsBuilderClient) RunLoadTest(_ context.Context, _ *pfs.RunLoadTestRequest, opts ...grpc.CallOption) (*pfs.RunLoadTestResponse, error) {
	return nil, unsupportedError("RunLoadTest")
}

func (c *unsupportedPfsBuilderClient) RunLoadTestDefault(_ context.Context, _ *types.Empty, opts ...grpc.CallOption) (*pfs.RunLoadTestResponse, error) {
	return nil, unsupportedError("RunLoadTestDefault")
}

func (c *unsupportedPfsBuilderClient) ListTask(_ context.Context, _ *taskapi.ListTaskRequest, opts ...grpc.CallOption) (pfs.API_ListTaskClient, error) {
	return nil, unsupportedError("ListTask")
}

type unsupportedPpsBuilderClient struct{}

func (c *unsupportedPpsBuilderClient) InspectJob(_ context.Context, _ *pps.InspectJobRequest, opts ...grpc.CallOption) (*pps.JobInfo, error) {
	return nil, unsupportedError("InspectJob")
}

func (c *unsupportedPpsBuilderClient) InspectJobSet(_ context.Context, _ *pps.InspectJobSetRequest, opts ...grpc.CallOption) (pps.API_InspectJobSetClient, error) {
	return nil, unsupportedError("InspectJobSet")
}

func (c *unsupportedPpsBuilderClient) ListJob(_ context.Context, _ *pps.ListJobRequest, opts ...grpc.CallOption) (pps.API_ListJobClient, error) {
	return nil, unsupportedError("ListJob")
}

func (c *unsupportedPpsBuilderClient) ListJobSet(_ context.Context, _ *pps.ListJobSetRequest, opts ...grpc.CallOption) (pps.API_ListJobSetClient, error) {
	return nil, unsupportedError("ListJobSet")
}

func (c *unsupportedPpsBuilderClient) SubscribeJob(_ context.Context, _ *pps.SubscribeJobRequest, opts ...grpc.CallOption) (pps.API_SubscribeJobClient, error) {
	return nil, unsupportedError("SubscribeJob")
}

func (c *unsupportedPpsBuilderClient) DeleteJob(_ context.Context, _ *pps.DeleteJobRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteJob")
}

func (c *unsupportedPpsBuilderClient) StopJob(_ context.Context, _ *pps.StopJobRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("StopJob")
}

func (c *unsupportedPpsBuilderClient) InspectDatum(_ context.Context, _ *pps.InspectDatumRequest, opts ...grpc.CallOption) (*pps.DatumInfo, error) {
	return nil, unsupportedError("InspectDatum")
}

func (c *unsupportedPpsBuilderClient) ListDatum(_ context.Context, _ *pps.ListDatumRequest, opts ...grpc.CallOption) (pps.API_ListDatumClient, error) {
	return nil, unsupportedError("ListDatum")
}

func (c *unsupportedPpsBuilderClient) RestartDatum(_ context.Context, _ *pps.RestartDatumRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("RestartDatum")
}

func (c *unsupportedPpsBuilderClient) CreatePipeline(_ context.Context, _ *pps.CreatePipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("CreatePipeline")
}

func (c *unsupportedPpsBuilderClient) InspectPipeline(_ context.Context, _ *pps.InspectPipelineRequest, opts ...grpc.CallOption) (*pps.PipelineInfo, error) {
	return nil, unsupportedError("InspectPipeline")
}

func (c *unsupportedPpsBuilderClient) ListPipeline(_ context.Context, _ *pps.ListPipelineRequest, opts ...grpc.CallOption) (pps.API_ListPipelineClient, error) {
	return nil, unsupportedError("ListPipeline")
}

func (c *unsupportedPpsBuilderClient) DeletePipeline(_ context.Context, _ *pps.DeletePipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeletePipeline")
}

func (c *unsupportedPpsBuilderClient) StartPipeline(_ context.Context, _ *pps.StartPipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("StartPipeline")
}

func (c *unsupportedPpsBuilderClient) StopPipeline(_ context.Context, _ *pps.StopPipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("StopPipeline")
}

func (c *unsupportedPpsBuilderClient) RunPipeline(_ context.Context, _ *pps.RunPipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("RunPipeline")
}

func (c *unsupportedPpsBuilderClient) RunCron(_ context.Context, _ *pps.RunCronRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("RunCron")
}

func (c *unsupportedPpsBuilderClient) CreateSecret(_ context.Context, _ *pps.CreateSecretRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("CreateSecret")
}

func (c *unsupportedPpsBuilderClient) DeleteSecret(_ context.Context, _ *pps.DeleteSecretRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteSecret")
}

func (c *unsupportedPpsBuilderClient) ListSecret(_ context.Context, _ *types.Empty, opts ...grpc.CallOption) (*pps.SecretInfos, error) {
	return nil, unsupportedError("ListSecret")
}

func (c *unsupportedPpsBuilderClient) InspectSecret(_ context.Context, _ *pps.InspectSecretRequest, opts ...grpc.CallOption) (*pps.SecretInfo, error) {
	return nil, unsupportedError("InspectSecret")
}

func (c *unsupportedPpsBuilderClient) DeleteAll(_ context.Context, _ *types.Empty, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteAll")
}

func (c *unsupportedPpsBuilderClient) GetLogs(_ context.Context, _ *pps.GetLogsRequest, opts ...grpc.CallOption) (pps.API_GetLogsClient, error) {
	return nil, unsupportedError("GetLogs")
}

func (c *unsupportedPpsBuilderClient) ActivateAuth(_ context.Context, _ *pps.ActivateAuthRequest, opts ...grpc.CallOption) (*pps.ActivateAuthResponse, error) {
	return nil, unsupportedError("ActivateAuth")
}

func (c *unsupportedPpsBuilderClient) UpdateJobState(_ context.Context, _ *pps.UpdateJobStateRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("UpdateJobState")
}

func (c *unsupportedPpsBuilderClient) RunLoadTest(_ context.Context, _ *pfs.RunLoadTestRequest, opts ...grpc.CallOption) (*pfs.RunLoadTestResponse, error) {
	return nil, unsupportedError("RunLoadTest")
}

func (c *unsupportedPpsBuilderClient) RunLoadTestDefault(_ context.Context, _ *types.Empty, opts ...grpc.CallOption) (*pfs.RunLoadTestResponse, error) {
	return nil, unsupportedError("RunLoadTestDefault")
}

func (c *unsupportedPpsBuilderClient) RenderTemplate(_ context.Context, _ *pps.RenderTemplateRequest, opts ...grpc.CallOption) (*pps.RenderTemplateResponse, error) {
	return nil, unsupportedError("RenderTemplate")
}

func (c *unsupportedPpsBuilderClient) ListTask(_ context.Context, _ *taskapi.ListTaskRequest, opts ...grpc.CallOption) (pps.API_ListTaskClient, error) {
	return nil, unsupportedError("ListTask")
}

type unsupportedEnterpriseBuilderClient struct{}

func (c *unsupportedEnterpriseBuilderClient) Activate(_ context.Context, _ *enterprise.ActivateRequest, opts ...grpc.CallOption) (*enterprise.ActivateResponse, error) {
	return nil, unsupportedError("Activate")
}

func (c *unsupportedEnterpriseBuilderClient) GetState(_ context.Context, _ *enterprise.GetStateRequest, opts ...grpc.CallOption) (*enterprise.GetStateResponse, error) {
	return nil, unsupportedError("GetState")
}

func (c *unsupportedEnterpriseBuilderClient) GetActivationCode(_ context.Context, _ *enterprise.GetActivationCodeRequest, opts ...grpc.CallOption) (*enterprise.GetActivationCodeResponse, error) {
	return nil, unsupportedError("GetActivationCode")
}

func (c *unsupportedEnterpriseBuilderClient) Heartbeat(_ context.Context, _ *enterprise.HeartbeatRequest, opts ...grpc.CallOption) (*enterprise.HeartbeatResponse, error) {
	return nil, unsupportedError("Heartbeat")
}

func (c *unsupportedEnterpriseBuilderClient) Deactivate(_ context.Context, _ *enterprise.DeactivateRequest, opts ...grpc.CallOption) (*enterprise.DeactivateResponse, error) {
	return nil, unsupportedError("Deactivate")
}

type unsupportedLicenseBuilderClient struct{}

func (c *unsupportedLicenseBuilderClient) Activate(_ context.Context, _ *license.ActivateRequest, opts ...grpc.CallOption) (*license.ActivateResponse, error) {
	return nil, unsupportedError("Activate")
}

func (c *unsupportedLicenseBuilderClient) GetActivationCode(_ context.Context, _ *license.GetActivationCodeRequest, opts ...grpc.CallOption) (*license.GetActivationCodeResponse, error) {
	return nil, unsupportedError("GetActivationCode")
}

func (c *unsupportedLicenseBuilderClient) DeleteAll(_ context.Context, _ *license.DeleteAllRequest, opts ...grpc.CallOption) (*license.DeleteAllResponse, error) {
	return nil, unsupportedError("DeleteAll")
}

func (c *unsupportedLicenseBuilderClient) AddCluster(_ context.Context, _ *license.AddClusterRequest, opts ...grpc.CallOption) (*license.AddClusterResponse, error) {
	return nil, unsupportedError("AddCluster")
}

func (c *unsupportedLicenseBuilderClient) DeleteCluster(_ context.Context, _ *license.DeleteClusterRequest, opts ...grpc.CallOption) (*license.DeleteClusterResponse, error) {
	return nil, unsupportedError("DeleteCluster")
}

func (c *unsupportedLicenseBuilderClient) ListClusters(_ context.Context, _ *license.ListClustersRequest, opts ...grpc.CallOption) (*license.ListClustersResponse, error) {
	return nil, unsupportedError("ListClusters")
}

func (c *unsupportedLicenseBuilderClient) UpdateCluster(_ context.Context, _ *license.UpdateClusterRequest, opts ...grpc.CallOption) (*license.UpdateClusterResponse, error) {
	return nil, unsupportedError("UpdateCluster")
}

func (c *unsupportedLicenseBuilderClient) Heartbeat(_ context.Context, _ *license.HeartbeatRequest, opts ...grpc.CallOption) (*license.HeartbeatResponse, error) {
	return nil, unsupportedError("Heartbeat")
}

func (c *unsupportedLicenseBuilderClient) ListUserClusters(_ context.Context, _ *license.ListUserClustersRequest, opts ...grpc.CallOption) (*license.ListUserClustersResponse, error) {
	return nil, unsupportedError("ListUserClusters")
}

type unsupportedAdminBuilderClient struct{}

func (c *unsupportedAdminBuilderClient) InspectCluster(_ context.Context, _ *types.Empty, opts ...grpc.CallOption) (*admin.ClusterInfo, error) {
	return nil, unsupportedError("InspectCluster")
}

type unsupportedVersionpbBuilderClient struct{}

func (c *unsupportedVersionpbBuilderClient) GetVersion(_ context.Context, _ *types.Empty, opts ...grpc.CallOption) (*versionpb.Version, error) {
	return nil, unsupportedError("GetVersion")
}

type unsupportedDebugBuilderClient struct{}

func (c *unsupportedDebugBuilderClient) Profile(_ context.Context, _ *debug.ProfileRequest, opts ...grpc.CallOption) (debug.Debug_ProfileClient, error) {
	return nil, unsupportedError("Profile")
}

func (c *unsupportedDebugBuilderClient) Binary(_ context.Context, _ *debug.BinaryRequest, opts ...grpc.CallOption) (debug.Debug_BinaryClient, error) {
	return nil, unsupportedError("Binary")
}

func (c *unsupportedDebugBuilderClient) Dump(_ context.Context, _ *debug.DumpRequest, opts ...grpc.CallOption) (debug.Debug_DumpClient, error) {
	return nil, unsupportedError("Dump")
}

type unsupportedProxyBuilderClient struct{}

func (c *unsupportedProxyBuilderClient) Listen(_ context.Context, _ *proxy.ListenRequest, opts ...grpc.CallOption) (proxy.API_ListenClient, error) {
	return nil, unsupportedError("Listen")
}

type unsupportedTransactionBuilderClient struct{}

func (c *unsupportedTransactionBuilderClient) BatchTransaction(_ context.Context, _ *transaction.BatchTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfo, error) {
	return nil, unsupportedError("BatchTransaction")
}

func (c *unsupportedTransactionBuilderClient) StartTransaction(_ context.Context, _ *transaction.StartTransactionRequest, opts ...grpc.CallOption) (*transaction.Transaction, error) {
	return nil, unsupportedError("StartTransaction")
}

func (c *unsupportedTransactionBuilderClient) InspectTransaction(_ context.Context, _ *transaction.InspectTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfo, error) {
	return nil, unsupportedError("InspectTransaction")
}

func (c *unsupportedTransactionBuilderClient) DeleteTransaction(_ context.Context, _ *transaction.DeleteTransactionRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteTransaction")
}

func (c *unsupportedTransactionBuilderClient) ListTransaction(_ context.Context, _ *transaction.ListTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfos, error) {
	return nil, unsupportedError("ListTransaction")
}

func (c *unsupportedTransactionBuilderClient) FinishTransaction(_ context.Context, _ *transaction.FinishTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfo, error) {
	return nil, unsupportedError("FinishTransaction")
}

func (c *unsupportedTransactionBuilderClient) DeleteAll(_ context.Context, _ *transaction.DeleteAllRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteAll")
}
