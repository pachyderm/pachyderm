# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [admin/admin.proto](#admin_admin-proto)
    - [ClusterInfo](#admin_v2-ClusterInfo)
    - [InspectClusterRequest](#admin_v2-InspectClusterRequest)
    - [WebResource](#admin_v2-WebResource)
  
    - [API](#admin_v2-API)
  
- [auth/auth.proto](#auth_auth-proto)
    - [ActivateRequest](#auth_v2-ActivateRequest)
    - [ActivateResponse](#auth_v2-ActivateResponse)
    - [AuthenticateRequest](#auth_v2-AuthenticateRequest)
    - [AuthenticateResponse](#auth_v2-AuthenticateResponse)
    - [AuthorizeRequest](#auth_v2-AuthorizeRequest)
    - [AuthorizeResponse](#auth_v2-AuthorizeResponse)
    - [DeactivateRequest](#auth_v2-DeactivateRequest)
    - [DeactivateResponse](#auth_v2-DeactivateResponse)
    - [DeleteExpiredAuthTokensRequest](#auth_v2-DeleteExpiredAuthTokensRequest)
    - [DeleteExpiredAuthTokensResponse](#auth_v2-DeleteExpiredAuthTokensResponse)
    - [ExtractAuthTokensRequest](#auth_v2-ExtractAuthTokensRequest)
    - [ExtractAuthTokensResponse](#auth_v2-ExtractAuthTokensResponse)
    - [GetConfigurationRequest](#auth_v2-GetConfigurationRequest)
    - [GetConfigurationResponse](#auth_v2-GetConfigurationResponse)
    - [GetGroupsForPrincipalRequest](#auth_v2-GetGroupsForPrincipalRequest)
    - [GetGroupsRequest](#auth_v2-GetGroupsRequest)
    - [GetGroupsResponse](#auth_v2-GetGroupsResponse)
    - [GetOIDCLoginRequest](#auth_v2-GetOIDCLoginRequest)
    - [GetOIDCLoginResponse](#auth_v2-GetOIDCLoginResponse)
    - [GetPermissionsForPrincipalRequest](#auth_v2-GetPermissionsForPrincipalRequest)
    - [GetPermissionsRequest](#auth_v2-GetPermissionsRequest)
    - [GetPermissionsResponse](#auth_v2-GetPermissionsResponse)
    - [GetRobotTokenRequest](#auth_v2-GetRobotTokenRequest)
    - [GetRobotTokenResponse](#auth_v2-GetRobotTokenResponse)
    - [GetRoleBindingRequest](#auth_v2-GetRoleBindingRequest)
    - [GetRoleBindingResponse](#auth_v2-GetRoleBindingResponse)
    - [GetRolesForPermissionRequest](#auth_v2-GetRolesForPermissionRequest)
    - [GetRolesForPermissionResponse](#auth_v2-GetRolesForPermissionResponse)
    - [GetUsersRequest](#auth_v2-GetUsersRequest)
    - [GetUsersResponse](#auth_v2-GetUsersResponse)
    - [Groups](#auth_v2-Groups)
    - [Groups.GroupsEntry](#auth_v2-Groups-GroupsEntry)
    - [ModifyMembersRequest](#auth_v2-ModifyMembersRequest)
    - [ModifyMembersResponse](#auth_v2-ModifyMembersResponse)
    - [ModifyRoleBindingRequest](#auth_v2-ModifyRoleBindingRequest)
    - [ModifyRoleBindingResponse](#auth_v2-ModifyRoleBindingResponse)
    - [OIDCConfig](#auth_v2-OIDCConfig)
    - [Resource](#auth_v2-Resource)
    - [RestoreAuthTokenRequest](#auth_v2-RestoreAuthTokenRequest)
    - [RestoreAuthTokenResponse](#auth_v2-RestoreAuthTokenResponse)
    - [RevokeAuthTokenRequest](#auth_v2-RevokeAuthTokenRequest)
    - [RevokeAuthTokenResponse](#auth_v2-RevokeAuthTokenResponse)
    - [RevokeAuthTokensForUserRequest](#auth_v2-RevokeAuthTokensForUserRequest)
    - [RevokeAuthTokensForUserResponse](#auth_v2-RevokeAuthTokensForUserResponse)
    - [Role](#auth_v2-Role)
    - [RoleBinding](#auth_v2-RoleBinding)
    - [RoleBinding.EntriesEntry](#auth_v2-RoleBinding-EntriesEntry)
    - [Roles](#auth_v2-Roles)
    - [Roles.RolesEntry](#auth_v2-Roles-RolesEntry)
    - [RotateRootTokenRequest](#auth_v2-RotateRootTokenRequest)
    - [RotateRootTokenResponse](#auth_v2-RotateRootTokenResponse)
    - [SessionInfo](#auth_v2-SessionInfo)
    - [SetConfigurationRequest](#auth_v2-SetConfigurationRequest)
    - [SetConfigurationResponse](#auth_v2-SetConfigurationResponse)
    - [SetGroupsForUserRequest](#auth_v2-SetGroupsForUserRequest)
    - [SetGroupsForUserResponse](#auth_v2-SetGroupsForUserResponse)
    - [TokenInfo](#auth_v2-TokenInfo)
    - [Users](#auth_v2-Users)
    - [Users.UsernamesEntry](#auth_v2-Users-UsernamesEntry)
    - [WhoAmIRequest](#auth_v2-WhoAmIRequest)
    - [WhoAmIResponse](#auth_v2-WhoAmIResponse)
  
    - [Permission](#auth_v2-Permission)
    - [ResourceType](#auth_v2-ResourceType)
  
    - [API](#auth_v2-API)
  
- [debug/debug.proto](#debug_debug-proto)
    - [App](#debug_v2-App)
    - [BinaryRequest](#debug_v2-BinaryRequest)
    - [DumpChunk](#debug_v2-DumpChunk)
    - [DumpContent](#debug_v2-DumpContent)
    - [DumpProgress](#debug_v2-DumpProgress)
    - [DumpRequest](#debug_v2-DumpRequest)
    - [DumpV2Request](#debug_v2-DumpV2Request)
    - [DumpV2Request.Defaults](#debug_v2-DumpV2Request-Defaults)
    - [Filter](#debug_v2-Filter)
    - [GetDumpV2TemplateRequest](#debug_v2-GetDumpV2TemplateRequest)
    - [GetDumpV2TemplateResponse](#debug_v2-GetDumpV2TemplateResponse)
    - [Pipeline](#debug_v2-Pipeline)
    - [Pod](#debug_v2-Pod)
    - [Profile](#debug_v2-Profile)
    - [ProfileRequest](#debug_v2-ProfileRequest)
    - [RunPFSLoadTestRequest](#debug_v2-RunPFSLoadTestRequest)
    - [RunPFSLoadTestResponse](#debug_v2-RunPFSLoadTestResponse)
    - [SetLogLevelRequest](#debug_v2-SetLogLevelRequest)
    - [SetLogLevelResponse](#debug_v2-SetLogLevelResponse)
    - [Starlark](#debug_v2-Starlark)
    - [StarlarkLiteral](#debug_v2-StarlarkLiteral)
    - [System](#debug_v2-System)
    - [Worker](#debug_v2-Worker)
  
    - [SetLogLevelRequest.LogLevel](#debug_v2-SetLogLevelRequest-LogLevel)
  
    - [Debug](#debug_v2-Debug)
  
- [enterprise/enterprise.proto](#enterprise_enterprise-proto)
    - [ActivateRequest](#enterprise_v2-ActivateRequest)
    - [ActivateResponse](#enterprise_v2-ActivateResponse)
    - [DeactivateRequest](#enterprise_v2-DeactivateRequest)
    - [DeactivateResponse](#enterprise_v2-DeactivateResponse)
    - [EnterpriseConfig](#enterprise_v2-EnterpriseConfig)
    - [EnterpriseRecord](#enterprise_v2-EnterpriseRecord)
    - [GetActivationCodeRequest](#enterprise_v2-GetActivationCodeRequest)
    - [GetActivationCodeResponse](#enterprise_v2-GetActivationCodeResponse)
    - [GetStateRequest](#enterprise_v2-GetStateRequest)
    - [GetStateResponse](#enterprise_v2-GetStateResponse)
    - [HeartbeatRequest](#enterprise_v2-HeartbeatRequest)
    - [HeartbeatResponse](#enterprise_v2-HeartbeatResponse)
    - [LicenseRecord](#enterprise_v2-LicenseRecord)
    - [PauseRequest](#enterprise_v2-PauseRequest)
    - [PauseResponse](#enterprise_v2-PauseResponse)
    - [PauseStatusRequest](#enterprise_v2-PauseStatusRequest)
    - [PauseStatusResponse](#enterprise_v2-PauseStatusResponse)
    - [TokenInfo](#enterprise_v2-TokenInfo)
    - [UnpauseRequest](#enterprise_v2-UnpauseRequest)
    - [UnpauseResponse](#enterprise_v2-UnpauseResponse)
  
    - [PauseStatusResponse.PauseStatus](#enterprise_v2-PauseStatusResponse-PauseStatus)
    - [State](#enterprise_v2-State)
  
    - [API](#enterprise_v2-API)
  
- [identity/identity.proto](#identity_identity-proto)
    - [CreateIDPConnectorRequest](#identity_v2-CreateIDPConnectorRequest)
    - [CreateIDPConnectorResponse](#identity_v2-CreateIDPConnectorResponse)
    - [CreateOIDCClientRequest](#identity_v2-CreateOIDCClientRequest)
    - [CreateOIDCClientResponse](#identity_v2-CreateOIDCClientResponse)
    - [DeleteAllRequest](#identity_v2-DeleteAllRequest)
    - [DeleteAllResponse](#identity_v2-DeleteAllResponse)
    - [DeleteIDPConnectorRequest](#identity_v2-DeleteIDPConnectorRequest)
    - [DeleteIDPConnectorResponse](#identity_v2-DeleteIDPConnectorResponse)
    - [DeleteOIDCClientRequest](#identity_v2-DeleteOIDCClientRequest)
    - [DeleteOIDCClientResponse](#identity_v2-DeleteOIDCClientResponse)
    - [GetIDPConnectorRequest](#identity_v2-GetIDPConnectorRequest)
    - [GetIDPConnectorResponse](#identity_v2-GetIDPConnectorResponse)
    - [GetIdentityServerConfigRequest](#identity_v2-GetIdentityServerConfigRequest)
    - [GetIdentityServerConfigResponse](#identity_v2-GetIdentityServerConfigResponse)
    - [GetOIDCClientRequest](#identity_v2-GetOIDCClientRequest)
    - [GetOIDCClientResponse](#identity_v2-GetOIDCClientResponse)
    - [IDPConnector](#identity_v2-IDPConnector)
    - [IdentityServerConfig](#identity_v2-IdentityServerConfig)
    - [ListIDPConnectorsRequest](#identity_v2-ListIDPConnectorsRequest)
    - [ListIDPConnectorsResponse](#identity_v2-ListIDPConnectorsResponse)
    - [ListOIDCClientsRequest](#identity_v2-ListOIDCClientsRequest)
    - [ListOIDCClientsResponse](#identity_v2-ListOIDCClientsResponse)
    - [OIDCClient](#identity_v2-OIDCClient)
    - [SetIdentityServerConfigRequest](#identity_v2-SetIdentityServerConfigRequest)
    - [SetIdentityServerConfigResponse](#identity_v2-SetIdentityServerConfigResponse)
    - [UpdateIDPConnectorRequest](#identity_v2-UpdateIDPConnectorRequest)
    - [UpdateIDPConnectorResponse](#identity_v2-UpdateIDPConnectorResponse)
    - [UpdateOIDCClientRequest](#identity_v2-UpdateOIDCClientRequest)
    - [UpdateOIDCClientResponse](#identity_v2-UpdateOIDCClientResponse)
    - [User](#identity_v2-User)
  
    - [API](#identity_v2-API)
  
- [internal/clusterstate/v2.5.0/commit_info.proto](#internal_clusterstate_v2-5-0_commit_info-proto)
    - [CommitInfo](#v2_5_0-CommitInfo)
  
- [internal/collection/test.proto](#internal_collection_test-proto)
    - [TestItem](#common-TestItem)
  
- [internal/config/config.proto](#internal_config_config-proto)
    - [Config](#config_v2-Config)
    - [ConfigV1](#config_v2-ConfigV1)
    - [ConfigV2](#config_v2-ConfigV2)
    - [ConfigV2.ContextsEntry](#config_v2-ConfigV2-ContextsEntry)
    - [Context](#config_v2-Context)
    - [Context.PortForwardersEntry](#config_v2-Context-PortForwardersEntry)
  
    - [ContextSource](#config_v2-ContextSource)
  
- [internal/metrics/metrics.proto](#internal_metrics_metrics-proto)
    - [Metrics](#metrics-Metrics)
  
- [internal/pfsload/pfsload.proto](#internal_pfsload_pfsload-proto)
    - [CommitSpec](#pfsload-CommitSpec)
    - [FileSourceSpec](#pfsload-FileSourceSpec)
    - [FrequencySpec](#pfsload-FrequencySpec)
    - [ModificationSpec](#pfsload-ModificationSpec)
    - [PutFileSpec](#pfsload-PutFileSpec)
    - [PutFileTask](#pfsload-PutFileTask)
    - [PutFileTaskResult](#pfsload-PutFileTaskResult)
    - [RandomDirectorySpec](#pfsload-RandomDirectorySpec)
    - [RandomFileSourceSpec](#pfsload-RandomFileSourceSpec)
    - [SizeSpec](#pfsload-SizeSpec)
    - [State](#pfsload-State)
    - [State.Commit](#pfsload-State-Commit)
    - [ValidatorSpec](#pfsload-ValidatorSpec)
  
- [internal/ppsdb/ppsdb.proto](#internal_ppsdb_ppsdb-proto)
    - [ClusterDefaultsWrapper](#pps_v2-ClusterDefaultsWrapper)
    - [ProjectDefaultsWrapper](#pps_v2-ProjectDefaultsWrapper)
  
- [internal/ppsload/ppsload.proto](#internal_ppsload_ppsload-proto)
    - [State](#ppsload-State)
  
- [internal/storage/chunk/chunk.proto](#internal_storage_chunk_chunk-proto)
    - [DataRef](#chunk-DataRef)
    - [Ref](#chunk-Ref)
  
    - [CompressionAlgo](#chunk-CompressionAlgo)
    - [EncryptionAlgo](#chunk-EncryptionAlgo)
  
- [internal/storage/fileset/fileset.proto](#internal_storage_fileset_fileset-proto)
    - [Composite](#fileset-Composite)
    - [Metadata](#fileset-Metadata)
    - [Primitive](#fileset-Primitive)
    - [TestCacheValue](#fileset-TestCacheValue)
  
- [internal/storage/fileset/index/index.proto](#internal_storage_fileset_index_index-proto)
    - [File](#index-File)
    - [Index](#index-Index)
    - [Range](#index-Range)
  
- [internal/task/task.proto](#internal_task_task-proto)
    - [Claim](#task-Claim)
    - [Group](#task-Group)
    - [Task](#task-Task)
    - [TestTask](#task-TestTask)
  
    - [State](#task-State)
  
- [internal/tracing/extended/extended_trace.proto](#internal_tracing_extended_extended_trace-proto)
    - [TraceProto](#extended-TraceProto)
    - [TraceProto.SerializedTraceEntry](#extended-TraceProto-SerializedTraceEntry)
  
- [license/license.proto](#license_license-proto)
    - [ActivateRequest](#license_v2-ActivateRequest)
    - [ActivateResponse](#license_v2-ActivateResponse)
    - [AddClusterRequest](#license_v2-AddClusterRequest)
    - [AddClusterResponse](#license_v2-AddClusterResponse)
    - [ClusterStatus](#license_v2-ClusterStatus)
    - [DeactivateRequest](#license_v2-DeactivateRequest)
    - [DeactivateResponse](#license_v2-DeactivateResponse)
    - [DeleteAllRequest](#license_v2-DeleteAllRequest)
    - [DeleteAllResponse](#license_v2-DeleteAllResponse)
    - [DeleteClusterRequest](#license_v2-DeleteClusterRequest)
    - [DeleteClusterResponse](#license_v2-DeleteClusterResponse)
    - [GetActivationCodeRequest](#license_v2-GetActivationCodeRequest)
    - [GetActivationCodeResponse](#license_v2-GetActivationCodeResponse)
    - [HeartbeatRequest](#license_v2-HeartbeatRequest)
    - [HeartbeatResponse](#license_v2-HeartbeatResponse)
    - [ListClustersRequest](#license_v2-ListClustersRequest)
    - [ListClustersResponse](#license_v2-ListClustersResponse)
    - [ListUserClustersRequest](#license_v2-ListUserClustersRequest)
    - [ListUserClustersResponse](#license_v2-ListUserClustersResponse)
    - [UpdateClusterRequest](#license_v2-UpdateClusterRequest)
    - [UpdateClusterResponse](#license_v2-UpdateClusterResponse)
    - [UserClusterInfo](#license_v2-UserClusterInfo)
  
    - [API](#license_v2-API)
  
- [logs/logs.proto](#logs_logs-proto)
    - [AdminLogQuery](#logs-AdminLogQuery)
    - [GetLogsRequest](#logs-GetLogsRequest)
    - [GetLogsResponse](#logs-GetLogsResponse)
    - [LogFilter](#logs-LogFilter)
    - [LogMessage](#logs-LogMessage)
    - [LogQuery](#logs-LogQuery)
    - [PagingHint](#logs-PagingHint)
    - [ParsedJSONLogMessage](#logs-ParsedJSONLogMessage)
    - [PipelineJobLogQuery](#logs-PipelineJobLogQuery)
    - [PipelineLogQuery](#logs-PipelineLogQuery)
    - [PodContainer](#logs-PodContainer)
    - [RegexLogFilter](#logs-RegexLogFilter)
    - [TimeRangeLogFilter](#logs-TimeRangeLogFilter)
    - [UserLogQuery](#logs-UserLogQuery)
    - [VerbatimLogMessage](#logs-VerbatimLogMessage)
  
    - [LogFormat](#logs-LogFormat)
    - [LogLevel](#logs-LogLevel)
  
    - [API](#logs-API)
  
- [pfs/pfs.proto](#pfs_pfs-proto)
    - [ActivateAuthRequest](#pfs_v2-ActivateAuthRequest)
    - [ActivateAuthResponse](#pfs_v2-ActivateAuthResponse)
    - [AddFile](#pfs_v2-AddFile)
    - [AddFile.URLSource](#pfs_v2-AddFile-URLSource)
    - [AddFileSetRequest](#pfs_v2-AddFileSetRequest)
    - [AuthInfo](#pfs_v2-AuthInfo)
    - [Branch](#pfs_v2-Branch)
    - [BranchInfo](#pfs_v2-BranchInfo)
    - [CheckStorageRequest](#pfs_v2-CheckStorageRequest)
    - [CheckStorageResponse](#pfs_v2-CheckStorageResponse)
    - [ClearCacheRequest](#pfs_v2-ClearCacheRequest)
    - [ClearCommitRequest](#pfs_v2-ClearCommitRequest)
    - [Commit](#pfs_v2-Commit)
    - [CommitInfo](#pfs_v2-CommitInfo)
    - [CommitInfo.Details](#pfs_v2-CommitInfo-Details)
    - [CommitOrigin](#pfs_v2-CommitOrigin)
    - [CommitSet](#pfs_v2-CommitSet)
    - [CommitSetInfo](#pfs_v2-CommitSetInfo)
    - [ComposeFileSetRequest](#pfs_v2-ComposeFileSetRequest)
    - [CopyFile](#pfs_v2-CopyFile)
    - [CreateBranchRequest](#pfs_v2-CreateBranchRequest)
    - [CreateFileSetResponse](#pfs_v2-CreateFileSetResponse)
    - [CreateProjectRequest](#pfs_v2-CreateProjectRequest)
    - [CreateRepoRequest](#pfs_v2-CreateRepoRequest)
    - [DeleteBranchRequest](#pfs_v2-DeleteBranchRequest)
    - [DeleteFile](#pfs_v2-DeleteFile)
    - [DeleteProjectRequest](#pfs_v2-DeleteProjectRequest)
    - [DeleteRepoRequest](#pfs_v2-DeleteRepoRequest)
    - [DeleteRepoResponse](#pfs_v2-DeleteRepoResponse)
    - [DeleteReposRequest](#pfs_v2-DeleteReposRequest)
    - [DeleteReposResponse](#pfs_v2-DeleteReposResponse)
    - [DiffFileRequest](#pfs_v2-DiffFileRequest)
    - [DiffFileResponse](#pfs_v2-DiffFileResponse)
    - [DropCommitRequest](#pfs_v2-DropCommitRequest)
    - [DropCommitResponse](#pfs_v2-DropCommitResponse)
    - [DropCommitSetRequest](#pfs_v2-DropCommitSetRequest)
    - [EgressRequest](#pfs_v2-EgressRequest)
    - [EgressResponse](#pfs_v2-EgressResponse)
    - [EgressResponse.ObjectStorageResult](#pfs_v2-EgressResponse-ObjectStorageResult)
    - [EgressResponse.SQLDatabaseResult](#pfs_v2-EgressResponse-SQLDatabaseResult)
    - [EgressResponse.SQLDatabaseResult.RowsWrittenEntry](#pfs_v2-EgressResponse-SQLDatabaseResult-RowsWrittenEntry)
    - [File](#pfs_v2-File)
    - [FileInfo](#pfs_v2-FileInfo)
    - [FindCommitsRequest](#pfs_v2-FindCommitsRequest)
    - [FindCommitsResponse](#pfs_v2-FindCommitsResponse)
    - [FinishCommitRequest](#pfs_v2-FinishCommitRequest)
    - [FsckRequest](#pfs_v2-FsckRequest)
    - [FsckResponse](#pfs_v2-FsckResponse)
    - [GetCacheRequest](#pfs_v2-GetCacheRequest)
    - [GetCacheResponse](#pfs_v2-GetCacheResponse)
    - [GetFileRequest](#pfs_v2-GetFileRequest)
    - [GetFileSetRequest](#pfs_v2-GetFileSetRequest)
    - [GlobFileRequest](#pfs_v2-GlobFileRequest)
    - [InspectBranchRequest](#pfs_v2-InspectBranchRequest)
    - [InspectCommitRequest](#pfs_v2-InspectCommitRequest)
    - [InspectCommitSetRequest](#pfs_v2-InspectCommitSetRequest)
    - [InspectFileRequest](#pfs_v2-InspectFileRequest)
    - [InspectProjectRequest](#pfs_v2-InspectProjectRequest)
    - [InspectProjectV2Request](#pfs_v2-InspectProjectV2Request)
    - [InspectProjectV2Response](#pfs_v2-InspectProjectV2Response)
    - [InspectRepoRequest](#pfs_v2-InspectRepoRequest)
    - [ListBranchRequest](#pfs_v2-ListBranchRequest)
    - [ListCommitRequest](#pfs_v2-ListCommitRequest)
    - [ListCommitSetRequest](#pfs_v2-ListCommitSetRequest)
    - [ListFileRequest](#pfs_v2-ListFileRequest)
    - [ListProjectRequest](#pfs_v2-ListProjectRequest)
    - [ListRepoRequest](#pfs_v2-ListRepoRequest)
    - [ModifyFileRequest](#pfs_v2-ModifyFileRequest)
    - [ObjectStorageEgress](#pfs_v2-ObjectStorageEgress)
    - [PathRange](#pfs_v2-PathRange)
    - [Project](#pfs_v2-Project)
    - [ProjectInfo](#pfs_v2-ProjectInfo)
    - [PutCacheRequest](#pfs_v2-PutCacheRequest)
    - [RenewFileSetRequest](#pfs_v2-RenewFileSetRequest)
    - [Repo](#pfs_v2-Repo)
    - [RepoInfo](#pfs_v2-RepoInfo)
    - [RepoInfo.Details](#pfs_v2-RepoInfo-Details)
    - [RepoPage](#pfs_v2-RepoPage)
    - [SQLDatabaseEgress](#pfs_v2-SQLDatabaseEgress)
    - [SQLDatabaseEgress.FileFormat](#pfs_v2-SQLDatabaseEgress-FileFormat)
    - [SQLDatabaseEgress.Secret](#pfs_v2-SQLDatabaseEgress-Secret)
    - [ShardFileSetRequest](#pfs_v2-ShardFileSetRequest)
    - [ShardFileSetResponse](#pfs_v2-ShardFileSetResponse)
    - [SquashCommitRequest](#pfs_v2-SquashCommitRequest)
    - [SquashCommitResponse](#pfs_v2-SquashCommitResponse)
    - [SquashCommitSetRequest](#pfs_v2-SquashCommitSetRequest)
    - [StartCommitRequest](#pfs_v2-StartCommitRequest)
    - [SubscribeCommitRequest](#pfs_v2-SubscribeCommitRequest)
    - [Trigger](#pfs_v2-Trigger)
    - [WalkFileRequest](#pfs_v2-WalkFileRequest)
  
    - [CommitState](#pfs_v2-CommitState)
    - [Delimiter](#pfs_v2-Delimiter)
    - [FileType](#pfs_v2-FileType)
    - [GetFileSetRequest.FileSetType](#pfs_v2-GetFileSetRequest-FileSetType)
    - [OriginKind](#pfs_v2-OriginKind)
    - [RepoPage.Ordering](#pfs_v2-RepoPage-Ordering)
    - [SQLDatabaseEgress.FileFormat.Type](#pfs_v2-SQLDatabaseEgress-FileFormat-Type)
  
    - [API](#pfs_v2-API)
  
- [pjs/pjs.proto](#pjs_pjs-proto)
    - [CancelJobRequest](#pjs-CancelJobRequest)
    - [CancelJobResponse](#pjs-CancelJobResponse)
    - [CreateJobRequest](#pjs-CreateJobRequest)
    - [CreateJobResponse](#pjs-CreateJobResponse)
    - [DeleteJobRequest](#pjs-DeleteJobRequest)
    - [DeleteJobResponse](#pjs-DeleteJobResponse)
    - [InspectJobRequest](#pjs-InspectJobRequest)
    - [InspectJobResponse](#pjs-InspectJobResponse)
    - [InspectQueueRequest](#pjs-InspectQueueRequest)
    - [InspectQueueResponse](#pjs-InspectQueueResponse)
    - [Job](#pjs-Job)
    - [JobInfo](#pjs-JobInfo)
    - [JobInfoDetails](#pjs-JobInfoDetails)
    - [ListJobRequest](#pjs-ListJobRequest)
    - [ListJobResponse](#pjs-ListJobResponse)
    - [ListQueueRequest](#pjs-ListQueueRequest)
    - [ListQueueResponse](#pjs-ListQueueResponse)
    - [ProcessQueueRequest](#pjs-ProcessQueueRequest)
    - [ProcessQueueResponse](#pjs-ProcessQueueResponse)
    - [Queue](#pjs-Queue)
    - [QueueElement](#pjs-QueueElement)
    - [QueueInfo](#pjs-QueueInfo)
    - [QueueInfoDetails](#pjs-QueueInfoDetails)
    - [WalkJobRequest](#pjs-WalkJobRequest)
  
    - [JobErrorCode](#pjs-JobErrorCode)
    - [JobState](#pjs-JobState)
  
    - [API](#pjs-API)
  
- [pps/pps.proto](#pps_pps-proto)
    - [ActivateAuthRequest](#pps_v2-ActivateAuthRequest)
    - [ActivateAuthResponse](#pps_v2-ActivateAuthResponse)
    - [Aggregate](#pps_v2-Aggregate)
    - [AggregateProcessStats](#pps_v2-AggregateProcessStats)
    - [CheckStatusRequest](#pps_v2-CheckStatusRequest)
    - [CheckStatusResponse](#pps_v2-CheckStatusResponse)
    - [ClusterDefaults](#pps_v2-ClusterDefaults)
    - [CreateDatumRequest](#pps_v2-CreateDatumRequest)
    - [CreatePipelineRequest](#pps_v2-CreatePipelineRequest)
    - [CreatePipelineTransaction](#pps_v2-CreatePipelineTransaction)
    - [CreatePipelineV2Request](#pps_v2-CreatePipelineV2Request)
    - [CreatePipelineV2Response](#pps_v2-CreatePipelineV2Response)
    - [CreateSecretRequest](#pps_v2-CreateSecretRequest)
    - [CronInput](#pps_v2-CronInput)
    - [Datum](#pps_v2-Datum)
    - [DatumInfo](#pps_v2-DatumInfo)
    - [DatumSetSpec](#pps_v2-DatumSetSpec)
    - [DatumStatus](#pps_v2-DatumStatus)
    - [DeleteJobRequest](#pps_v2-DeleteJobRequest)
    - [DeletePipelineRequest](#pps_v2-DeletePipelineRequest)
    - [DeletePipelinesRequest](#pps_v2-DeletePipelinesRequest)
    - [DeletePipelinesResponse](#pps_v2-DeletePipelinesResponse)
    - [DeleteSecretRequest](#pps_v2-DeleteSecretRequest)
    - [Determined](#pps_v2-Determined)
    - [Egress](#pps_v2-Egress)
    - [GPUSpec](#pps_v2-GPUSpec)
    - [GetClusterDefaultsRequest](#pps_v2-GetClusterDefaultsRequest)
    - [GetClusterDefaultsResponse](#pps_v2-GetClusterDefaultsResponse)
    - [GetLogsRequest](#pps_v2-GetLogsRequest)
    - [GetProjectDefaultsRequest](#pps_v2-GetProjectDefaultsRequest)
    - [GetProjectDefaultsResponse](#pps_v2-GetProjectDefaultsResponse)
    - [Input](#pps_v2-Input)
    - [InputFile](#pps_v2-InputFile)
    - [InspectDatumRequest](#pps_v2-InspectDatumRequest)
    - [InspectJobRequest](#pps_v2-InspectJobRequest)
    - [InspectJobSetRequest](#pps_v2-InspectJobSetRequest)
    - [InspectPipelineRequest](#pps_v2-InspectPipelineRequest)
    - [InspectSecretRequest](#pps_v2-InspectSecretRequest)
    - [Job](#pps_v2-Job)
    - [JobInfo](#pps_v2-JobInfo)
    - [JobInfo.Details](#pps_v2-JobInfo-Details)
    - [JobInput](#pps_v2-JobInput)
    - [JobSet](#pps_v2-JobSet)
    - [JobSetInfo](#pps_v2-JobSetInfo)
    - [ListDatumRequest](#pps_v2-ListDatumRequest)
    - [ListDatumRequest.Filter](#pps_v2-ListDatumRequest-Filter)
    - [ListJobRequest](#pps_v2-ListJobRequest)
    - [ListJobSetRequest](#pps_v2-ListJobSetRequest)
    - [ListPipelineRequest](#pps_v2-ListPipelineRequest)
    - [LogMessage](#pps_v2-LogMessage)
    - [LokiLogMessage](#pps_v2-LokiLogMessage)
    - [LokiRequest](#pps_v2-LokiRequest)
    - [Metadata](#pps_v2-Metadata)
    - [Metadata.AnnotationsEntry](#pps_v2-Metadata-AnnotationsEntry)
    - [Metadata.LabelsEntry](#pps_v2-Metadata-LabelsEntry)
    - [PFSInput](#pps_v2-PFSInput)
    - [ParallelismSpec](#pps_v2-ParallelismSpec)
    - [Pipeline](#pps_v2-Pipeline)
    - [PipelineInfo](#pps_v2-PipelineInfo)
    - [PipelineInfo.Details](#pps_v2-PipelineInfo-Details)
    - [PipelineInfos](#pps_v2-PipelineInfos)
    - [ProcessStats](#pps_v2-ProcessStats)
    - [ProjectDefaults](#pps_v2-ProjectDefaults)
    - [RenderTemplateRequest](#pps_v2-RenderTemplateRequest)
    - [RenderTemplateRequest.ArgsEntry](#pps_v2-RenderTemplateRequest-ArgsEntry)
    - [RenderTemplateResponse](#pps_v2-RenderTemplateResponse)
    - [RerunPipelineRequest](#pps_v2-RerunPipelineRequest)
    - [ResourceSpec](#pps_v2-ResourceSpec)
    - [RestartDatumRequest](#pps_v2-RestartDatumRequest)
    - [RunCronRequest](#pps_v2-RunCronRequest)
    - [RunLoadTestRequest](#pps_v2-RunLoadTestRequest)
    - [RunLoadTestResponse](#pps_v2-RunLoadTestResponse)
    - [RunPipelineRequest](#pps_v2-RunPipelineRequest)
    - [SchedulingSpec](#pps_v2-SchedulingSpec)
    - [SchedulingSpec.NodeSelectorEntry](#pps_v2-SchedulingSpec-NodeSelectorEntry)
    - [Secret](#pps_v2-Secret)
    - [SecretInfo](#pps_v2-SecretInfo)
    - [SecretInfos](#pps_v2-SecretInfos)
    - [SecretMount](#pps_v2-SecretMount)
    - [Service](#pps_v2-Service)
    - [SetClusterDefaultsRequest](#pps_v2-SetClusterDefaultsRequest)
    - [SetClusterDefaultsResponse](#pps_v2-SetClusterDefaultsResponse)
    - [SetProjectDefaultsRequest](#pps_v2-SetProjectDefaultsRequest)
    - [SetProjectDefaultsResponse](#pps_v2-SetProjectDefaultsResponse)
    - [Spout](#pps_v2-Spout)
    - [StartPipelineRequest](#pps_v2-StartPipelineRequest)
    - [StopJobRequest](#pps_v2-StopJobRequest)
    - [StopPipelineRequest](#pps_v2-StopPipelineRequest)
    - [SubscribeJobRequest](#pps_v2-SubscribeJobRequest)
    - [TFJob](#pps_v2-TFJob)
    - [Toleration](#pps_v2-Toleration)
    - [Transform](#pps_v2-Transform)
    - [Transform.EnvEntry](#pps_v2-Transform-EnvEntry)
    - [UpdateJobStateRequest](#pps_v2-UpdateJobStateRequest)
    - [Worker](#pps_v2-Worker)
    - [WorkerStatus](#pps_v2-WorkerStatus)
  
    - [DatumState](#pps_v2-DatumState)
    - [JobState](#pps_v2-JobState)
    - [PipelineInfo.PipelineType](#pps_v2-PipelineInfo-PipelineType)
    - [PipelineState](#pps_v2-PipelineState)
    - [TaintEffect](#pps_v2-TaintEffect)
    - [TolerationOperator](#pps_v2-TolerationOperator)
    - [WorkerState](#pps_v2-WorkerState)
  
    - [API](#pps_v2-API)
  
- [protoextensions/json-schema-options.proto](#protoextensions_json-schema-options-proto)
    - [EnumOptions](#protoc-gen-jsonschema-EnumOptions)
    - [FieldOptions](#protoc-gen-jsonschema-FieldOptions)
    - [FileOptions](#protoc-gen-jsonschema-FileOptions)
    - [MessageOptions](#protoc-gen-jsonschema-MessageOptions)
  
    - [File-level Extensions](#protoextensions_json-schema-options-proto-extensions)
    - [File-level Extensions](#protoextensions_json-schema-options-proto-extensions)
    - [File-level Extensions](#protoextensions_json-schema-options-proto-extensions)
    - [File-level Extensions](#protoextensions_json-schema-options-proto-extensions)
  
- [protoextensions/log.proto](#protoextensions_log-proto)
    - [File-level Extensions](#protoextensions_log-proto-extensions)
    - [File-level Extensions](#protoextensions_log-proto-extensions)
  
- [protoextensions/validate.proto](#protoextensions_validate-proto)
    - [AnyRules](#validate-AnyRules)
    - [BoolRules](#validate-BoolRules)
    - [BytesRules](#validate-BytesRules)
    - [DoubleRules](#validate-DoubleRules)
    - [DurationRules](#validate-DurationRules)
    - [EnumRules](#validate-EnumRules)
    - [FieldRules](#validate-FieldRules)
    - [Fixed32Rules](#validate-Fixed32Rules)
    - [Fixed64Rules](#validate-Fixed64Rules)
    - [FloatRules](#validate-FloatRules)
    - [Int32Rules](#validate-Int32Rules)
    - [Int64Rules](#validate-Int64Rules)
    - [MapRules](#validate-MapRules)
    - [MessageRules](#validate-MessageRules)
    - [RepeatedRules](#validate-RepeatedRules)
    - [SFixed32Rules](#validate-SFixed32Rules)
    - [SFixed64Rules](#validate-SFixed64Rules)
    - [SInt32Rules](#validate-SInt32Rules)
    - [SInt64Rules](#validate-SInt64Rules)
    - [StringRules](#validate-StringRules)
    - [TimestampRules](#validate-TimestampRules)
    - [UInt32Rules](#validate-UInt32Rules)
    - [UInt64Rules](#validate-UInt64Rules)
  
    - [KnownRegex](#validate-KnownRegex)
  
    - [File-level Extensions](#protoextensions_validate-proto-extensions)
    - [File-level Extensions](#protoextensions_validate-proto-extensions)
    - [File-level Extensions](#protoextensions_validate-proto-extensions)
    - [File-level Extensions](#protoextensions_validate-proto-extensions)
  
- [proxy/proxy.proto](#proxy_proxy-proto)
    - [ListenRequest](#proxy-ListenRequest)
    - [ListenResponse](#proxy-ListenResponse)
  
    - [API](#proxy-API)
  
- [server/pfs/server/pfsserver.proto](#server_pfs_server_pfsserver-proto)
    - [CompactTask](#pfsserver-CompactTask)
    - [CompactTaskResult](#pfsserver-CompactTaskResult)
    - [ConcatTask](#pfsserver-ConcatTask)
    - [ConcatTaskResult](#pfsserver-ConcatTaskResult)
    - [GetFileURLTask](#pfsserver-GetFileURLTask)
    - [GetFileURLTaskResult](#pfsserver-GetFileURLTaskResult)
    - [PathRange](#pfsserver-PathRange)
    - [PutFileURLTask](#pfsserver-PutFileURLTask)
    - [PutFileURLTaskResult](#pfsserver-PutFileURLTaskResult)
    - [ShardTask](#pfsserver-ShardTask)
    - [ShardTaskResult](#pfsserver-ShardTaskResult)
    - [ValidateTask](#pfsserver-ValidateTask)
    - [ValidateTaskResult](#pfsserver-ValidateTaskResult)
  
- [server/worker/common/common.proto](#server_worker_common_common-proto)
    - [Input](#common-Input)
  
- [server/worker/datum/datum.proto](#server_worker_datum_datum-proto)
    - [ComposeTask](#datum-ComposeTask)
    - [ComposeTaskResult](#datum-ComposeTaskResult)
    - [CrossTask](#datum-CrossTask)
    - [CrossTaskResult](#datum-CrossTaskResult)
    - [KeyTask](#datum-KeyTask)
    - [KeyTaskResult](#datum-KeyTaskResult)
    - [MergeTask](#datum-MergeTask)
    - [MergeTaskResult](#datum-MergeTaskResult)
    - [Meta](#datum-Meta)
    - [PFSTask](#datum-PFSTask)
    - [PFSTaskResult](#datum-PFSTaskResult)
    - [SetSpec](#datum-SetSpec)
    - [Stats](#datum-Stats)
  
    - [KeyTask.Type](#datum-KeyTask-Type)
    - [MergeTask.Type](#datum-MergeTask-Type)
    - [State](#datum-State)
  
- [server/worker/pipeline/transform/transform.proto](#server_worker_pipeline_transform_transform-proto)
    - [CreateDatumSetsTask](#pachyderm-worker-pipeline-transform-CreateDatumSetsTask)
    - [CreateDatumSetsTaskResult](#pachyderm-worker-pipeline-transform-CreateDatumSetsTaskResult)
    - [CreateParallelDatumsTask](#pachyderm-worker-pipeline-transform-CreateParallelDatumsTask)
    - [CreateParallelDatumsTaskResult](#pachyderm-worker-pipeline-transform-CreateParallelDatumsTaskResult)
    - [CreateSerialDatumsTask](#pachyderm-worker-pipeline-transform-CreateSerialDatumsTask)
    - [CreateSerialDatumsTaskResult](#pachyderm-worker-pipeline-transform-CreateSerialDatumsTaskResult)
    - [DatumSetTask](#pachyderm-worker-pipeline-transform-DatumSetTask)
    - [DatumSetTaskResult](#pachyderm-worker-pipeline-transform-DatumSetTaskResult)
  
- [task/task.proto](#task_task-proto)
    - [Group](#taskapi-Group)
    - [ListTaskRequest](#taskapi-ListTaskRequest)
    - [TaskInfo](#taskapi-TaskInfo)
  
    - [State](#taskapi-State)
  
- [transaction/transaction.proto](#transaction_transaction-proto)
    - [BatchTransactionRequest](#transaction_v2-BatchTransactionRequest)
    - [DeleteAllRequest](#transaction_v2-DeleteAllRequest)
    - [DeleteTransactionRequest](#transaction_v2-DeleteTransactionRequest)
    - [FinishTransactionRequest](#transaction_v2-FinishTransactionRequest)
    - [InspectTransactionRequest](#transaction_v2-InspectTransactionRequest)
    - [ListTransactionRequest](#transaction_v2-ListTransactionRequest)
    - [StartTransactionRequest](#transaction_v2-StartTransactionRequest)
    - [Transaction](#transaction_v2-Transaction)
    - [TransactionInfo](#transaction_v2-TransactionInfo)
    - [TransactionInfos](#transaction_v2-TransactionInfos)
    - [TransactionRequest](#transaction_v2-TransactionRequest)
    - [TransactionResponse](#transaction_v2-TransactionResponse)
  
    - [API](#transaction_v2-API)
  
- [version/versionpb/version.proto](#version_versionpb_version-proto)
    - [Version](#versionpb_v2-Version)
  
    - [API](#versionpb_v2-API)
  
- [worker/worker.proto](#worker_worker-proto)
    - [CancelRequest](#pachyderm-worker-CancelRequest)
    - [CancelResponse](#pachyderm-worker-CancelResponse)
    - [NextDatumRequest](#pachyderm-worker-NextDatumRequest)
    - [NextDatumResponse](#pachyderm-worker-NextDatumResponse)
  
    - [Worker](#pachyderm-worker-Worker)
  
- [Scalar Value Types](#scalar-value-types)



<a name="admin_admin-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## admin/admin.proto



<a name="admin_v2-ClusterInfo"></a>

### ClusterInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| deployment_id | [string](#string) |  |  |
| warnings_ok | [bool](#bool) |  | True if the server is capable of generating warnings. |
| warnings | [string](#string) | repeated | Warnings about the client configuration. |
| proxy_host | [string](#string) |  | The configured public URL of Pachyderm. |
| proxy_tls | [bool](#bool) |  | True if Pachyderm is served over TLS (HTTPS). |
| paused | [bool](#bool) |  | True if this pachd is in &#34;paused&#34; mode. |
| web_resources | [WebResource](#admin_v2-WebResource) |  | Any HTTP links that the client might want to be aware of. |






<a name="admin_v2-InspectClusterRequest"></a>

### InspectClusterRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client_version | [versionpb_v2.Version](#versionpb_v2-Version) |  | The version of the client that&#39;s connecting; used by the server to warn about too-old (or too-new!) clients. |
| current_project | [pfs_v2.Project](#pfs_v2-Project) |  | If CurrentProject is set, then InspectCluster will return an error if the project does not exist. |






<a name="admin_v2-WebResource"></a>

### WebResource
WebResource contains URL prefixes of common HTTP functions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| archive_download_base_url | [string](#string) |  | The base URL of the archive server; append a filename to this. Empty if the archive server is not exposed. |
| create_pipeline_request_json_schema_url | [string](#string) |  | Where to find the CreatePipelineRequest JSON schema; if this server is not accessible via a URL, then a link to Github is provided based on the baked-in version of the server. |





 

 

 


<a name="admin_v2-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| InspectCluster | [InspectClusterRequest](#admin_v2-InspectClusterRequest) | [ClusterInfo](#admin_v2-ClusterInfo) |  |

 



<a name="auth_auth-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## auth/auth.proto



<a name="auth_v2-ActivateRequest"></a>

### ActivateRequest
ActivateRequest enables authentication on the cluster. It issues an auth token
with no expiration for the irrevocable admin user `pach:root`.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| root_token | [string](#string) |  | If set, this token is used as the root user login token. Otherwise the root token is randomly generated and returned in the response. |






<a name="auth_v2-ActivateResponse"></a>

### ActivateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pach_token | [string](#string) |  | pach_token authenticates the caller with Pachyderm (if you want to perform Pachyderm operations after auth has been activated as themselves, you must present this token along with your regular request) |






<a name="auth_v2-AuthenticateRequest"></a>

### AuthenticateRequest
Exactly one of &#39;id_token&#39; or &#39;one_time_password&#39; must be set:


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| oidc_state | [string](#string) |  | This is the session state that Pachyderm creates in order to keep track of information related to the current OIDC session. |
| id_token | [string](#string) |  | This is an ID Token issued by the OIDC provider. |






<a name="auth_v2-AuthenticateResponse"></a>

### AuthenticateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pach_token | [string](#string) |  | pach_token authenticates the caller with Pachyderm (if you want to perform Pachyderm operations after auth has been activated as themselves, you must present this token along with your regular request) |






<a name="auth_v2-AuthorizeRequest"></a>

### AuthorizeRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [Resource](#auth_v2-Resource) |  |  |
| permissions | [Permission](#auth_v2-Permission) | repeated | permissions are the operations the caller is attempting to perform |






<a name="auth_v2-AuthorizeResponse"></a>

### AuthorizeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| authorized | [bool](#bool) |  | authorized is true if the caller has the require permissions |
| satisfied | [Permission](#auth_v2-Permission) | repeated | satisfied is the set of permission that the principal has |
| missing | [Permission](#auth_v2-Permission) | repeated | missing is the set of permissions that the principal lacks |
| principal | [string](#string) |  | principal is the principal the request was evaluated for |






<a name="auth_v2-DeactivateRequest"></a>

### DeactivateRequest







<a name="auth_v2-DeactivateResponse"></a>

### DeactivateResponse







<a name="auth_v2-DeleteExpiredAuthTokensRequest"></a>

### DeleteExpiredAuthTokensRequest







<a name="auth_v2-DeleteExpiredAuthTokensResponse"></a>

### DeleteExpiredAuthTokensResponse







<a name="auth_v2-ExtractAuthTokensRequest"></a>

### ExtractAuthTokensRequest
ExtractAuthTokens returns all the hashed robot tokens that have been issued.
User tokens are not extracted as they can be recreated by logging in.






<a name="auth_v2-ExtractAuthTokensResponse"></a>

### ExtractAuthTokensResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tokens | [TokenInfo](#auth_v2-TokenInfo) | repeated |  |






<a name="auth_v2-GetConfigurationRequest"></a>

### GetConfigurationRequest







<a name="auth_v2-GetConfigurationResponse"></a>

### GetConfigurationResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| configuration | [OIDCConfig](#auth_v2-OIDCConfig) |  |  |






<a name="auth_v2-GetGroupsForPrincipalRequest"></a>

### GetGroupsForPrincipalRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| principal | [string](#string) |  |  |






<a name="auth_v2-GetGroupsRequest"></a>

### GetGroupsRequest







<a name="auth_v2-GetGroupsResponse"></a>

### GetGroupsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| groups | [string](#string) | repeated |  |






<a name="auth_v2-GetOIDCLoginRequest"></a>

### GetOIDCLoginRequest







<a name="auth_v2-GetOIDCLoginResponse"></a>

### GetOIDCLoginResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| login_url | [string](#string) |  | The login URL generated for the OIDC object |
| state | [string](#string) |  |  |






<a name="auth_v2-GetPermissionsForPrincipalRequest"></a>

### GetPermissionsForPrincipalRequest
GetPermissionsForPrincipal evaluates an arbitrary principal&#39;s permissions
on a resource


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [Resource](#auth_v2-Resource) |  |  |
| principal | [string](#string) |  |  |






<a name="auth_v2-GetPermissionsRequest"></a>

### GetPermissionsRequest
GetPermissions evaluates the current user&#39;s permissions on a resource


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [Resource](#auth_v2-Resource) |  |  |






<a name="auth_v2-GetPermissionsResponse"></a>

### GetPermissionsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| permissions | [Permission](#auth_v2-Permission) | repeated | permissions is the set of permissions the principal has |
| roles | [string](#string) | repeated | roles is the set of roles the principal has |






<a name="auth_v2-GetRobotTokenRequest"></a>

### GetRobotTokenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| robot | [string](#string) |  | The returned token will allow the caller to access resources as this robot user |
| ttl | [int64](#int64) |  | ttl indicates the requested (approximate) remaining lifetime of this token, in seconds |






<a name="auth_v2-GetRobotTokenResponse"></a>

### GetRobotTokenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| token | [string](#string) |  | A new auth token for the requested robot |






<a name="auth_v2-GetRoleBindingRequest"></a>

### GetRoleBindingRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [Resource](#auth_v2-Resource) |  |  |






<a name="auth_v2-GetRoleBindingResponse"></a>

### GetRoleBindingResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| binding | [RoleBinding](#auth_v2-RoleBinding) |  |  |






<a name="auth_v2-GetRolesForPermissionRequest"></a>

### GetRolesForPermissionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| permission | [Permission](#auth_v2-Permission) |  |  |






<a name="auth_v2-GetRolesForPermissionResponse"></a>

### GetRolesForPermissionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| roles | [Role](#auth_v2-Role) | repeated |  |






<a name="auth_v2-GetUsersRequest"></a>

### GetUsersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [string](#string) |  |  |






<a name="auth_v2-GetUsersResponse"></a>

### GetUsersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| usernames | [string](#string) | repeated |  |






<a name="auth_v2-Groups"></a>

### Groups



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| groups | [Groups.GroupsEntry](#auth_v2-Groups-GroupsEntry) | repeated |  |






<a name="auth_v2-Groups-GroupsEntry"></a>

### Groups.GroupsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bool](#bool) |  |  |






<a name="auth_v2-ModifyMembersRequest"></a>

### ModifyMembersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [string](#string) |  |  |
| add | [string](#string) | repeated |  |
| remove | [string](#string) | repeated |  |






<a name="auth_v2-ModifyMembersResponse"></a>

### ModifyMembersResponse







<a name="auth_v2-ModifyRoleBindingRequest"></a>

### ModifyRoleBindingRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| resource | [Resource](#auth_v2-Resource) |  | resource is the resource to modify the role bindings on |
| principal | [string](#string) |  | principal is the principal to modify the roles binding for |
| roles | [string](#string) | repeated | roles is the set of roles for principal - an empty list removes all role bindings |






<a name="auth_v2-ModifyRoleBindingResponse"></a>

### ModifyRoleBindingResponse







<a name="auth_v2-OIDCConfig"></a>

### OIDCConfig
Configure Pachyderm&#39;s auth system with an OIDC provider


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| issuer | [string](#string) |  |  |
| client_id | [string](#string) |  |  |
| client_secret | [string](#string) |  |  |
| redirect_uri | [string](#string) |  |  |
| scopes | [string](#string) | repeated |  |
| require_email_verified | [bool](#bool) |  |  |
| localhost_issuer | [bool](#bool) |  | localhost_issuer ignores the contents of the issuer claim and makes all OIDC requests to the embedded OIDC provider. This is necessary to support some network configurations like Minikube. |
| user_accessible_issuer_host | [string](#string) |  | user_accessible_issuer_host can be set to override the host used in the OAuth2 authorization URL in case the OIDC issuer isn&#39;t accessible outside the cluster. This requires a fully formed URL with scheme of either http or https. This is necessary to support some configurations like Minikube. |






<a name="auth_v2-Resource"></a>

### Resource
Resource represents any resource that has role-bindings in the system


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [ResourceType](#auth_v2-ResourceType) |  |  |
| name | [string](#string) |  |  |






<a name="auth_v2-RestoreAuthTokenRequest"></a>

### RestoreAuthTokenRequest
RestoreAuthToken inserts a hashed token that has previously been extracted.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| token | [TokenInfo](#auth_v2-TokenInfo) |  |  |






<a name="auth_v2-RestoreAuthTokenResponse"></a>

### RestoreAuthTokenResponse







<a name="auth_v2-RevokeAuthTokenRequest"></a>

### RevokeAuthTokenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| token | [string](#string) |  |  |






<a name="auth_v2-RevokeAuthTokenResponse"></a>

### RevokeAuthTokenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | [int64](#int64) |  |  |






<a name="auth_v2-RevokeAuthTokensForUserRequest"></a>

### RevokeAuthTokensForUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| username | [string](#string) |  |  |






<a name="auth_v2-RevokeAuthTokensForUserResponse"></a>

### RevokeAuthTokensForUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | [int64](#int64) |  |  |






<a name="auth_v2-Role"></a>

### Role



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| permissions | [Permission](#auth_v2-Permission) | repeated |  |
| can_be_bound_to | [ResourceType](#auth_v2-ResourceType) | repeated | Resources this role can be bound to. For example, you can&#39;t apply clusterAdmin to a repo, so REPO would not be listed here. |
| returned_for | [ResourceType](#auth_v2-ResourceType) | repeated | Resources this role is returned for. For example, a principal might have clusterAdmin permissions on the cluster, and this is what allows them to write to a repo. So, clusterAdmin is returned for the repo, even though it cannot be bound to a repo. |






<a name="auth_v2-RoleBinding"></a>

### RoleBinding
RoleBinding represents the set of roles principals have on a given Resource


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| entries | [RoleBinding.EntriesEntry](#auth_v2-RoleBinding-EntriesEntry) | repeated | principal -&gt; roles. All principal names include the structured prefix indicating their type. |






<a name="auth_v2-RoleBinding-EntriesEntry"></a>

### RoleBinding.EntriesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Roles](#auth_v2-Roles) |  |  |






<a name="auth_v2-Roles"></a>

### Roles
Roles represents the set of roles a principal has


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| roles | [Roles.RolesEntry](#auth_v2-Roles-RolesEntry) | repeated |  |






<a name="auth_v2-Roles-RolesEntry"></a>

### Roles.RolesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bool](#bool) |  |  |






<a name="auth_v2-RotateRootTokenRequest"></a>

### RotateRootTokenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| root_token | [string](#string) |  | root_token is used as the new root token value. If it&#39;s unset, then a token will be auto-generated. |






<a name="auth_v2-RotateRootTokenResponse"></a>

### RotateRootTokenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| root_token | [string](#string) |  |  |






<a name="auth_v2-SessionInfo"></a>

### SessionInfo
SessionInfo stores information associated with one OIDC authentication
session (i.e. a single instance of a single user logging in). Sessions are
short-lived and stored in the &#39;oidc-authns&#39; collection, keyed by the OIDC
&#39;state&#39; token (30-character CSPRNG-generated string). &#39;GetOIDCLogin&#39;
generates and inserts entries, then /authorization-code/callback retrieves
an access token from the ID provider and uses it to retrive the caller&#39;s
email and store it in &#39;email&#39;, and finally Authorize() returns a Pachyderm
token identified with that email address as a subject in Pachyderm.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| nonce | [string](#string) |  | nonce is used by /authorization-code/callback to validate session continuity with the IdP after a user has arrived there from GetOIDCLogin(). This is a 30-character CSPRNG-generated string. |
| email | [string](#string) |  | email contains the email adddress associated with a user in their OIDC ID provider. Currently users are identified with their email address rather than their OIDC subject identifier to make switching between OIDC ID providers easier for users, and to make user identities more easily comprehensible in Pachyderm. The OIDC spec doesn&#39;t require that users&#39; emails be present or unique, but we think this will be preferable in practice. |
| conversion_err | [bool](#bool) |  | conversion_err indicates whether an error was encountered while exchanging an auth code for an access token, or while obtaining a user&#39;s email (in /authorization-code/callback). Storing the error state here allows any sibling calls to Authenticate() (i.e. using the same OIDC state token) to notify their caller that an error has occurred. We avoid passing the caller any details of the error (which are logged by Pachyderm) to avoid giving information to a user who has network access to Pachyderm but not an account in the OIDC provider. |






<a name="auth_v2-SetConfigurationRequest"></a>

### SetConfigurationRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| configuration | [OIDCConfig](#auth_v2-OIDCConfig) |  |  |






<a name="auth_v2-SetConfigurationResponse"></a>

### SetConfigurationResponse







<a name="auth_v2-SetGroupsForUserRequest"></a>

### SetGroupsForUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| username | [string](#string) |  |  |
| groups | [string](#string) | repeated |  |






<a name="auth_v2-SetGroupsForUserResponse"></a>

### SetGroupsForUserResponse







<a name="auth_v2-TokenInfo"></a>

### TokenInfo
TokenInfo is the &#39;value&#39; of an auth token &#39;key&#39; in the &#39;tokens&#39; collection


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| subject | [string](#string) |  | Subject (i.e. Pachyderm account) that a given token authorizes. See the note at the top of the doc for an explanation of subject structure. |
| expiration | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| hashed_token | [string](#string) |  |  |






<a name="auth_v2-Users"></a>

### Users



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| usernames | [Users.UsernamesEntry](#auth_v2-Users-UsernamesEntry) | repeated |  |






<a name="auth_v2-Users-UsernamesEntry"></a>

### Users.UsernamesEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [bool](#bool) |  |  |






<a name="auth_v2-WhoAmIRequest"></a>

### WhoAmIRequest







<a name="auth_v2-WhoAmIResponse"></a>

### WhoAmIResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| username | [string](#string) |  |  |
| expiration | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |





 


<a name="auth_v2-Permission"></a>

### Permission
Permission represents the ability to perform a given operation on a Resource

| Name | Number | Description |
| ---- | ------ | ----------- |
| PERMISSION_UNKNOWN | 0 |  |
| CLUSTER_MODIFY_BINDINGS | 100 |  |
| CLUSTER_GET_BINDINGS | 101 |  |
| CLUSTER_GET_PACHD_LOGS | 148 |  |
| CLUSTER_GET_LOKI_LOGS | 150 |  |
| CLUSTER_AUTH_ACTIVATE | 102 |  |
| CLUSTER_AUTH_DEACTIVATE | 103 |  |
| CLUSTER_AUTH_GET_CONFIG | 104 |  |
| CLUSTER_AUTH_SET_CONFIG | 105 |  |
| CLUSTER_AUTH_GET_ROBOT_TOKEN | 139 |  |
| CLUSTER_AUTH_MODIFY_GROUP_MEMBERS | 109 |  |
| CLUSTER_AUTH_GET_GROUPS | 110 |  |
| CLUSTER_AUTH_GET_GROUP_USERS | 111 |  |
| CLUSTER_AUTH_EXTRACT_TOKENS | 112 |  |
| CLUSTER_AUTH_RESTORE_TOKEN | 113 |  |
| CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL | 141 |  |
| CLUSTER_AUTH_DELETE_EXPIRED_TOKENS | 140 |  |
| CLUSTER_AUTH_REVOKE_USER_TOKENS | 142 |  |
| CLUSTER_AUTH_ROTATE_ROOT_TOKEN | 147 |  |
| CLUSTER_ENTERPRISE_ACTIVATE | 114 |  |
| CLUSTER_ENTERPRISE_HEARTBEAT | 115 |  |
| CLUSTER_ENTERPRISE_GET_CODE | 116 |  |
| CLUSTER_ENTERPRISE_DEACTIVATE | 117 |  |
| CLUSTER_ENTERPRISE_PAUSE | 149 |  |
| CLUSTER_IDENTITY_SET_CONFIG | 118 |  |
| CLUSTER_IDENTITY_GET_CONFIG | 119 |  |
| CLUSTER_IDENTITY_CREATE_IDP | 120 |  |
| CLUSTER_IDENTITY_UPDATE_IDP | 121 |  |
| CLUSTER_IDENTITY_LIST_IDPS | 122 |  |
| CLUSTER_IDENTITY_GET_IDP | 123 |  |
| CLUSTER_IDENTITY_DELETE_IDP | 124 |  |
| CLUSTER_IDENTITY_CREATE_OIDC_CLIENT | 125 |  |
| CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT | 126 |  |
| CLUSTER_IDENTITY_LIST_OIDC_CLIENTS | 127 |  |
| CLUSTER_IDENTITY_GET_OIDC_CLIENT | 128 |  |
| CLUSTER_IDENTITY_DELETE_OIDC_CLIENT | 129 |  |
| CLUSTER_DEBUG_DUMP | 131 |  |
| CLUSTER_LICENSE_ACTIVATE | 132 |  |
| CLUSTER_LICENSE_GET_CODE | 133 |  |
| CLUSTER_LICENSE_ADD_CLUSTER | 134 |  |
| CLUSTER_LICENSE_UPDATE_CLUSTER | 135 |  |
| CLUSTER_LICENSE_DELETE_CLUSTER | 136 |  |
| CLUSTER_LICENSE_LIST_CLUSTERS | 137 |  |
| CLUSTER_CREATE_SECRET | 143 | TODO(actgardner): Make k8s secrets into nouns and add an Update RPC |
| CLUSTER_LIST_SECRETS | 144 |  |
| SECRET_DELETE | 145 |  |
| SECRET_INSPECT | 146 |  |
| CLUSTER_DELETE_ALL | 138 |  |
| REPO_READ | 200 |  |
| REPO_WRITE | 201 |  |
| REPO_MODIFY_BINDINGS | 202 |  |
| REPO_DELETE | 203 |  |
| REPO_INSPECT_COMMIT | 204 |  |
| REPO_LIST_COMMIT | 205 |  |
| REPO_DELETE_COMMIT | 206 |  |
| REPO_CREATE_BRANCH | 207 |  |
| REPO_LIST_BRANCH | 208 |  |
| REPO_DELETE_BRANCH | 209 |  |
| REPO_INSPECT_FILE | 210 |  |
| REPO_LIST_FILE | 211 |  |
| REPO_ADD_PIPELINE_READER | 212 |  |
| REPO_REMOVE_PIPELINE_READER | 213 |  |
| REPO_ADD_PIPELINE_WRITER | 214 |  |
| PIPELINE_LIST_JOB | 301 |  |
| CLUSTER_SET_DEFAULTS | 302 | CLUSTER_SET_DEFAULTS is part of PPS. |
| PROJECT_SET_DEFAULTS | 303 | PROJECT_SET_DEFAULTS is part of PPS. |
| PROJECT_CREATE | 400 |  |
| PROJECT_DELETE | 401 |  |
| PROJECT_LIST_REPO | 402 |  |
| PROJECT_CREATE_REPO | 403 |  |
| PROJECT_MODIFY_BINDINGS | 404 |  |



<a name="auth_v2-ResourceType"></a>

### ResourceType
ResourceType represents the type of a Resource

| Name | Number | Description |
| ---- | ------ | ----------- |
| RESOURCE_TYPE_UNKNOWN | 0 |  |
| CLUSTER | 1 |  |
| REPO | 2 |  |
| SPEC_REPO | 3 |  |
| PROJECT | 4 |  |


 

 


<a name="auth_v2-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Activate | [ActivateRequest](#auth_v2-ActivateRequest) | [ActivateResponse](#auth_v2-ActivateResponse) | Activate/Deactivate the auth API. &#39;Activate&#39; sets an initial set of admins for the Pachyderm cluster, and &#39;Deactivate&#39; removes all ACLs, tokens, and admins from the Pachyderm cluster, making all data publicly accessable |
| Deactivate | [DeactivateRequest](#auth_v2-DeactivateRequest) | [DeactivateResponse](#auth_v2-DeactivateResponse) |  |
| GetConfiguration | [GetConfigurationRequest](#auth_v2-GetConfigurationRequest) | [GetConfigurationResponse](#auth_v2-GetConfigurationResponse) |  |
| SetConfiguration | [SetConfigurationRequest](#auth_v2-SetConfigurationRequest) | [SetConfigurationResponse](#auth_v2-SetConfigurationResponse) |  |
| Authenticate | [AuthenticateRequest](#auth_v2-AuthenticateRequest) | [AuthenticateResponse](#auth_v2-AuthenticateResponse) |  |
| Authorize | [AuthorizeRequest](#auth_v2-AuthorizeRequest) | [AuthorizeResponse](#auth_v2-AuthorizeResponse) |  |
| GetPermissions | [GetPermissionsRequest](#auth_v2-GetPermissionsRequest) | [GetPermissionsResponse](#auth_v2-GetPermissionsResponse) |  |
| GetPermissionsForPrincipal | [GetPermissionsForPrincipalRequest](#auth_v2-GetPermissionsForPrincipalRequest) | [GetPermissionsResponse](#auth_v2-GetPermissionsResponse) |  |
| WhoAmI | [WhoAmIRequest](#auth_v2-WhoAmIRequest) | [WhoAmIResponse](#auth_v2-WhoAmIResponse) |  |
| GetRolesForPermission | [GetRolesForPermissionRequest](#auth_v2-GetRolesForPermissionRequest) | [GetRolesForPermissionResponse](#auth_v2-GetRolesForPermissionResponse) |  |
| ModifyRoleBinding | [ModifyRoleBindingRequest](#auth_v2-ModifyRoleBindingRequest) | [ModifyRoleBindingResponse](#auth_v2-ModifyRoleBindingResponse) |  |
| GetRoleBinding | [GetRoleBindingRequest](#auth_v2-GetRoleBindingRequest) | [GetRoleBindingResponse](#auth_v2-GetRoleBindingResponse) |  |
| GetOIDCLogin | [GetOIDCLoginRequest](#auth_v2-GetOIDCLoginRequest) | [GetOIDCLoginResponse](#auth_v2-GetOIDCLoginResponse) |  |
| GetRobotToken | [GetRobotTokenRequest](#auth_v2-GetRobotTokenRequest) | [GetRobotTokenResponse](#auth_v2-GetRobotTokenResponse) |  |
| RevokeAuthToken | [RevokeAuthTokenRequest](#auth_v2-RevokeAuthTokenRequest) | [RevokeAuthTokenResponse](#auth_v2-RevokeAuthTokenResponse) |  |
| RevokeAuthTokensForUser | [RevokeAuthTokensForUserRequest](#auth_v2-RevokeAuthTokensForUserRequest) | [RevokeAuthTokensForUserResponse](#auth_v2-RevokeAuthTokensForUserResponse) |  |
| SetGroupsForUser | [SetGroupsForUserRequest](#auth_v2-SetGroupsForUserRequest) | [SetGroupsForUserResponse](#auth_v2-SetGroupsForUserResponse) |  |
| ModifyMembers | [ModifyMembersRequest](#auth_v2-ModifyMembersRequest) | [ModifyMembersResponse](#auth_v2-ModifyMembersResponse) |  |
| GetGroups | [GetGroupsRequest](#auth_v2-GetGroupsRequest) | [GetGroupsResponse](#auth_v2-GetGroupsResponse) |  |
| GetGroupsForPrincipal | [GetGroupsForPrincipalRequest](#auth_v2-GetGroupsForPrincipalRequest) | [GetGroupsResponse](#auth_v2-GetGroupsResponse) |  |
| GetUsers | [GetUsersRequest](#auth_v2-GetUsersRequest) | [GetUsersResponse](#auth_v2-GetUsersResponse) |  |
| ExtractAuthTokens | [ExtractAuthTokensRequest](#auth_v2-ExtractAuthTokensRequest) | [ExtractAuthTokensResponse](#auth_v2-ExtractAuthTokensResponse) |  |
| RestoreAuthToken | [RestoreAuthTokenRequest](#auth_v2-RestoreAuthTokenRequest) | [RestoreAuthTokenResponse](#auth_v2-RestoreAuthTokenResponse) |  |
| DeleteExpiredAuthTokens | [DeleteExpiredAuthTokensRequest](#auth_v2-DeleteExpiredAuthTokensRequest) | [DeleteExpiredAuthTokensResponse](#auth_v2-DeleteExpiredAuthTokensResponse) |  |
| RotateRootToken | [RotateRootTokenRequest](#auth_v2-RotateRootTokenRequest) | [RotateRootTokenResponse](#auth_v2-RotateRootTokenResponse) |  |

 



<a name="debug_debug-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## debug/debug.proto



<a name="debug_v2-App"></a>

### App



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| pods | [Pod](#debug_v2-Pod) | repeated |  |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| pipeline | [Pipeline](#debug_v2-Pipeline) |  |  |






<a name="debug_v2-BinaryRequest"></a>

### BinaryRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filter | [Filter](#debug_v2-Filter) |  |  |






<a name="debug_v2-DumpChunk"></a>

### DumpChunk



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| content | [DumpContent](#debug_v2-DumpContent) |  |  |
| progress | [DumpProgress](#debug_v2-DumpProgress) |  |  |






<a name="debug_v2-DumpContent"></a>

### DumpContent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| content | [bytes](#bytes) |  |  |






<a name="debug_v2-DumpProgress"></a>

### DumpProgress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| task | [string](#string) |  |  |
| total | [int64](#int64) |  |  |
| progress | [int64](#int64) |  |  |






<a name="debug_v2-DumpRequest"></a>

### DumpRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filter | [Filter](#debug_v2-Filter) |  |  |
| limit | [int64](#int64) |  | Limit sets the limit for the number of commits / jobs that are returned for each repo / pipeline in the dump. |






<a name="debug_v2-DumpV2Request"></a>

### DumpV2Request



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| system | [System](#debug_v2-System) |  | Which system-level information to include in the debug dump. |
| pipelines | [Pipeline](#debug_v2-Pipeline) | repeated | Which pipelines to fetch information about and include in the debug dump. |
| input_repos | [bool](#bool) |  | If true, fetch information about non-output repos. |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  | How long to run the dump for. |
| defaults | [DumpV2Request.Defaults](#debug_v2-DumpV2Request-Defaults) |  | Which defaults to include in the debug dump. |
| starlark_scripts | [Starlark](#debug_v2-Starlark) | repeated | A list of Starlark scripts to run. |






<a name="debug_v2-DumpV2Request-Defaults"></a>

### DumpV2Request.Defaults



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_defaults | [bool](#bool) |  | If true, include the cluster defaults. |






<a name="debug_v2-Filter"></a>

### Filter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pachd | [bool](#bool) |  |  |
| pipeline | [pps_v2.Pipeline](#pps_v2-Pipeline) |  |  |
| worker | [Worker](#debug_v2-Worker) |  |  |
| database | [bool](#bool) |  |  |






<a name="debug_v2-GetDumpV2TemplateRequest"></a>

### GetDumpV2TemplateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filters | [string](#string) | repeated |  |






<a name="debug_v2-GetDumpV2TemplateResponse"></a>

### GetDumpV2TemplateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| request | [DumpV2Request](#debug_v2-DumpV2Request) |  |  |






<a name="debug_v2-Pipeline"></a>

### Pipeline



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  |  |
| name | [string](#string) |  |  |






<a name="debug_v2-Pod"></a>

### Pod



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| ip | [string](#string) |  |  |
| containers | [string](#string) | repeated |  |






<a name="debug_v2-Profile"></a>

### Profile



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| duration | [google.protobuf.Duration](#google-protobuf-Duration) |  | only meaningful if name == &#34;cpu&#34; |






<a name="debug_v2-ProfileRequest"></a>

### ProfileRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| profile | [Profile](#debug_v2-Profile) |  |  |
| filter | [Filter](#debug_v2-Filter) |  |  |






<a name="debug_v2-RunPFSLoadTestRequest"></a>

### RunPFSLoadTestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| spec | [string](#string) |  |  |
| branch | [pfs_v2.Branch](#pfs_v2-Branch) |  |  |
| seed | [int64](#int64) |  |  |
| state_id | [string](#string) |  |  |






<a name="debug_v2-RunPFSLoadTestResponse"></a>

### RunPFSLoadTestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| spec | [string](#string) |  |  |
| branch | [pfs_v2.Branch](#pfs_v2-Branch) |  |  |
| seed | [int64](#int64) |  |  |
| error | [string](#string) |  |  |
| duration | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| state_id | [string](#string) |  |  |






<a name="debug_v2-SetLogLevelRequest"></a>

### SetLogLevelRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pachyderm | [SetLogLevelRequest.LogLevel](#debug_v2-SetLogLevelRequest-LogLevel) |  |  |
| grpc | [SetLogLevelRequest.LogLevel](#debug_v2-SetLogLevelRequest-LogLevel) |  |  |
| duration | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| recurse | [bool](#bool) |  |  |






<a name="debug_v2-SetLogLevelResponse"></a>

### SetLogLevelResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| affected_pods | [string](#string) | repeated |  |
| errored_pods | [string](#string) | repeated |  |






<a name="debug_v2-Starlark"></a>

### Starlark
Starlark controls the running of a Starlark script.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| builtin | [string](#string) |  | One built into the pachd binary. |
| literal | [StarlarkLiteral](#debug_v2-StarlarkLiteral) |  | Or a script supplied in this request. |
| timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  | How long to allow the script to run for. If unset, defaults to 1 minute. |






<a name="debug_v2-StarlarkLiteral"></a>

### StarlarkLiteral
StarlarkLiteral is a custom Starlark script.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | The name of the script; used for debug messages and to control where the output goes. |
| program_text | [string](#string) |  | The text of the &#34;debugdump&#34; personality Starlark program. |






<a name="debug_v2-System"></a>

### System



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| helm | [bool](#bool) |  |  |
| database | [bool](#bool) |  |  |
| version | [bool](#bool) |  |  |
| describes | [App](#debug_v2-App) | repeated |  |
| logs | [App](#debug_v2-App) | repeated |  |
| loki_logs | [App](#debug_v2-App) | repeated |  |
| binaries | [App](#debug_v2-App) | repeated |  |
| profiles | [App](#debug_v2-App) | repeated |  |






<a name="debug_v2-Worker"></a>

### Worker



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pod | [string](#string) |  |  |
| redirected | [bool](#bool) |  |  |





 


<a name="debug_v2-SetLogLevelRequest-LogLevel"></a>

### SetLogLevelRequest.LogLevel


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| DEBUG | 1 |  |
| INFO | 2 |  |
| ERROR | 3 |  |
| OFF | 4 | Only GRPC logs can be turned off. |


 

 


<a name="debug_v2-Debug"></a>

### Debug


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Profile | [ProfileRequest](#debug_v2-ProfileRequest) | [.google.protobuf.BytesValue](#google-protobuf-BytesValue) stream |  |
| Binary | [BinaryRequest](#debug_v2-BinaryRequest) | [.google.protobuf.BytesValue](#google-protobuf-BytesValue) stream |  |
| Dump | [DumpRequest](#debug_v2-DumpRequest) | [.google.protobuf.BytesValue](#google-protobuf-BytesValue) stream |  |
| SetLogLevel | [SetLogLevelRequest](#debug_v2-SetLogLevelRequest) | [SetLogLevelResponse](#debug_v2-SetLogLevelResponse) |  |
| GetDumpV2Template | [GetDumpV2TemplateRequest](#debug_v2-GetDumpV2TemplateRequest) | [GetDumpV2TemplateResponse](#debug_v2-GetDumpV2TemplateResponse) |  |
| DumpV2 | [DumpV2Request](#debug_v2-DumpV2Request) | [DumpChunk](#debug_v2-DumpChunk) stream |  |
| RunPFSLoadTest | [RunPFSLoadTestRequest](#debug_v2-RunPFSLoadTestRequest) | [RunPFSLoadTestResponse](#debug_v2-RunPFSLoadTestResponse) | RunLoadTest runs a load test. |
| RunPFSLoadTestDefault | [.google.protobuf.Empty](#google-protobuf-Empty) | [RunPFSLoadTestResponse](#debug_v2-RunPFSLoadTestResponse) | RunLoadTestDefault runs the default load tests. |

 



<a name="enterprise_enterprise-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## enterprise/enterprise.proto



<a name="enterprise_v2-ActivateRequest"></a>

### ActivateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| license_server | [string](#string) |  |  |
| id | [string](#string) |  |  |
| secret | [string](#string) |  |  |






<a name="enterprise_v2-ActivateResponse"></a>

### ActivateResponse







<a name="enterprise_v2-DeactivateRequest"></a>

### DeactivateRequest







<a name="enterprise_v2-DeactivateResponse"></a>

### DeactivateResponse







<a name="enterprise_v2-EnterpriseConfig"></a>

### EnterpriseConfig
EnterpriseConfig is the configuration we store for heartbeating
to the license server.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| license_server | [string](#string) |  | license_server is the address of the grpc license service |
| id | [string](#string) |  | id is the unique identifier for this pachd, which is registered with the license service |
| secret | [string](#string) |  | secret is a shared secret between this pachd and the license service |






<a name="enterprise_v2-EnterpriseRecord"></a>

### EnterpriseRecord
EnterpriseRecord is a protobuf we cache in etcd to store the
enterprise status.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| license | [LicenseRecord](#enterprise_v2-LicenseRecord) |  | license is the cached LicenseRecord retrieved from the most recent heartbeat to the license server. |
| last_heartbeat | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | last_heartbeat is the timestamp of the last successful heartbeat to the license server |
| heartbeat_failed | [bool](#bool) |  | heartbeat_failed is set if the license is still valid, but the pachd is no longer registered with an enterprise server. This is the same as the expired state, where auth is locked but not disabled. |






<a name="enterprise_v2-GetActivationCodeRequest"></a>

### GetActivationCodeRequest







<a name="enterprise_v2-GetActivationCodeResponse"></a>

### GetActivationCodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [State](#enterprise_v2-State) |  |  |
| info | [TokenInfo](#enterprise_v2-TokenInfo) |  |  |
| activation_code | [string](#string) |  |  |






<a name="enterprise_v2-GetStateRequest"></a>

### GetStateRequest







<a name="enterprise_v2-GetStateResponse"></a>

### GetStateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [State](#enterprise_v2-State) |  |  |
| info | [TokenInfo](#enterprise_v2-TokenInfo) |  |  |
| activation_code | [string](#string) |  | activation_code will always be an empty string, call GetEnterpriseCode to get the activation code |






<a name="enterprise_v2-HeartbeatRequest"></a>

### HeartbeatRequest
Heartbeat in the enterprise service just triggers a heartbeat for
testing purposes. The RPC used to communicate with the license
service is defined in the license service.






<a name="enterprise_v2-HeartbeatResponse"></a>

### HeartbeatResponse







<a name="enterprise_v2-LicenseRecord"></a>

### LicenseRecord
LicenseRecord is the record we store in etcd for a Pachyderm enterprise
token that has been provided to a Pachyderm license server


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| activation_code | [string](#string) |  |  |
| expires | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="enterprise_v2-PauseRequest"></a>

### PauseRequest







<a name="enterprise_v2-PauseResponse"></a>

### PauseResponse







<a name="enterprise_v2-PauseStatusRequest"></a>

### PauseStatusRequest







<a name="enterprise_v2-PauseStatusResponse"></a>

### PauseStatusResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [PauseStatusResponse.PauseStatus](#enterprise_v2-PauseStatusResponse-PauseStatus) |  |  |






<a name="enterprise_v2-TokenInfo"></a>

### TokenInfo
TokenInfo contains information about the currently active enterprise token


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| expires | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | expires indicates when the current token expires (unset if there is no current token) |






<a name="enterprise_v2-UnpauseRequest"></a>

### UnpauseRequest







<a name="enterprise_v2-UnpauseResponse"></a>

### UnpauseResponse






 


<a name="enterprise_v2-PauseStatusResponse-PauseStatus"></a>

### PauseStatusResponse.PauseStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNPAUSED | 0 |  |
| PARTIALLY_PAUSED | 1 |  |
| PAUSED | 2 |  |



<a name="enterprise_v2-State"></a>

### State


| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 |  |
| ACTIVE | 1 |  |
| EXPIRED | 2 |  |
| HEARTBEAT_FAILED | 3 |  |


 

 


<a name="enterprise_v2-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Activate | [ActivateRequest](#enterprise_v2-ActivateRequest) | [ActivateResponse](#enterprise_v2-ActivateResponse) | Provide a Pachyderm enterprise token, enabling Pachyderm enterprise features, such as the Pachyderm Dashboard and Auth system |
| GetState | [GetStateRequest](#enterprise_v2-GetStateRequest) | [GetStateResponse](#enterprise_v2-GetStateResponse) |  |
| GetActivationCode | [GetActivationCodeRequest](#enterprise_v2-GetActivationCodeRequest) | [GetActivationCodeResponse](#enterprise_v2-GetActivationCodeResponse) |  |
| Heartbeat | [HeartbeatRequest](#enterprise_v2-HeartbeatRequest) | [HeartbeatResponse](#enterprise_v2-HeartbeatResponse) | Heartbeat is used in testing to trigger a heartbeat on demand. Normally this happens on a timer. |
| Deactivate | [DeactivateRequest](#enterprise_v2-DeactivateRequest) | [DeactivateResponse](#enterprise_v2-DeactivateResponse) | Deactivate removes a cluster&#39;s enterprise activation token and sets its enterprise state to NONE. |
| Pause | [PauseRequest](#enterprise_v2-PauseRequest) | [PauseResponse](#enterprise_v2-PauseResponse) | Pause pauses the cluster. |
| Unpause | [UnpauseRequest](#enterprise_v2-UnpauseRequest) | [UnpauseResponse](#enterprise_v2-UnpauseResponse) | Unpause unpauses the cluser. |
| PauseStatus | [PauseStatusRequest](#enterprise_v2-PauseStatusRequest) | [PauseStatusResponse](#enterprise_v2-PauseStatusResponse) |  |

 



<a name="identity_identity-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## identity/identity.proto



<a name="identity_v2-CreateIDPConnectorRequest"></a>

### CreateIDPConnectorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connector | [IDPConnector](#identity_v2-IDPConnector) |  |  |






<a name="identity_v2-CreateIDPConnectorResponse"></a>

### CreateIDPConnectorResponse







<a name="identity_v2-CreateOIDCClientRequest"></a>

### CreateOIDCClientRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client | [OIDCClient](#identity_v2-OIDCClient) |  |  |






<a name="identity_v2-CreateOIDCClientResponse"></a>

### CreateOIDCClientResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client | [OIDCClient](#identity_v2-OIDCClient) |  |  |






<a name="identity_v2-DeleteAllRequest"></a>

### DeleteAllRequest







<a name="identity_v2-DeleteAllResponse"></a>

### DeleteAllResponse







<a name="identity_v2-DeleteIDPConnectorRequest"></a>

### DeleteIDPConnectorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="identity_v2-DeleteIDPConnectorResponse"></a>

### DeleteIDPConnectorResponse







<a name="identity_v2-DeleteOIDCClientRequest"></a>

### DeleteOIDCClientRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="identity_v2-DeleteOIDCClientResponse"></a>

### DeleteOIDCClientResponse







<a name="identity_v2-GetIDPConnectorRequest"></a>

### GetIDPConnectorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="identity_v2-GetIDPConnectorResponse"></a>

### GetIDPConnectorResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connector | [IDPConnector](#identity_v2-IDPConnector) |  |  |






<a name="identity_v2-GetIdentityServerConfigRequest"></a>

### GetIdentityServerConfigRequest







<a name="identity_v2-GetIdentityServerConfigResponse"></a>

### GetIdentityServerConfigResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| config | [IdentityServerConfig](#identity_v2-IdentityServerConfig) |  |  |






<a name="identity_v2-GetOIDCClientRequest"></a>

### GetOIDCClientRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="identity_v2-GetOIDCClientResponse"></a>

### GetOIDCClientResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client | [OIDCClient](#identity_v2-OIDCClient) |  |  |






<a name="identity_v2-IDPConnector"></a>

### IDPConnector
IDPConnector represents a connection to an identity provider


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | ID is the unique identifier for this connector. |
| name | [string](#string) |  | Name is the human-readable identifier for this connector, which will be shown to end users when they&#39;re authenticating. |
| type | [string](#string) |  | Type is the type of the IDP ex. `saml`, `oidc`, `github`. |
| configVersion | [int64](#int64) |  | ConfigVersion must be incremented every time a connector is updated, to avoid concurrent updates conflicting. |
| jsonConfig | [string](#string) |  | This is left for backwards compatibility, but we want users to use the config defined below. |
| config | [google.protobuf.Struct](#google-protobuf-Struct) |  | Config is the configuration for the upstream IDP, which varies based on the type. We make the assumption that this is either yaml or JSON. |






<a name="identity_v2-IdentityServerConfig"></a>

### IdentityServerConfig
IdentityServerConfig is the configuration for the identity web server.
When the configuration is changed the web server is reloaded automatically.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| issuer | [string](#string) |  |  |
| id_token_expiry | [string](#string) |  |  |
| rotation_token_expiry | [string](#string) |  |  |






<a name="identity_v2-ListIDPConnectorsRequest"></a>

### ListIDPConnectorsRequest







<a name="identity_v2-ListIDPConnectorsResponse"></a>

### ListIDPConnectorsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connectors | [IDPConnector](#identity_v2-IDPConnector) | repeated |  |






<a name="identity_v2-ListOIDCClientsRequest"></a>

### ListOIDCClientsRequest







<a name="identity_v2-ListOIDCClientsResponse"></a>

### ListOIDCClientsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clients | [OIDCClient](#identity_v2-OIDCClient) | repeated |  |






<a name="identity_v2-OIDCClient"></a>

### OIDCClient



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| redirect_uris | [string](#string) | repeated |  |
| trusted_peers | [string](#string) | repeated |  |
| name | [string](#string) |  |  |
| secret | [string](#string) |  |  |






<a name="identity_v2-SetIdentityServerConfigRequest"></a>

### SetIdentityServerConfigRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| config | [IdentityServerConfig](#identity_v2-IdentityServerConfig) |  |  |






<a name="identity_v2-SetIdentityServerConfigResponse"></a>

### SetIdentityServerConfigResponse







<a name="identity_v2-UpdateIDPConnectorRequest"></a>

### UpdateIDPConnectorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| connector | [IDPConnector](#identity_v2-IDPConnector) |  |  |






<a name="identity_v2-UpdateIDPConnectorResponse"></a>

### UpdateIDPConnectorResponse







<a name="identity_v2-UpdateOIDCClientRequest"></a>

### UpdateOIDCClientRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| client | [OIDCClient](#identity_v2-OIDCClient) |  |  |






<a name="identity_v2-UpdateOIDCClientResponse"></a>

### UpdateOIDCClientResponse







<a name="identity_v2-User"></a>

### User
User represents an IDP user that has authenticated via OIDC


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| email | [string](#string) |  |  |
| last_authenticated | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |





 

 

 


<a name="identity_v2-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| SetIdentityServerConfig | [SetIdentityServerConfigRequest](#identity_v2-SetIdentityServerConfigRequest) | [SetIdentityServerConfigResponse](#identity_v2-SetIdentityServerConfigResponse) |  |
| GetIdentityServerConfig | [GetIdentityServerConfigRequest](#identity_v2-GetIdentityServerConfigRequest) | [GetIdentityServerConfigResponse](#identity_v2-GetIdentityServerConfigResponse) |  |
| CreateIDPConnector | [CreateIDPConnectorRequest](#identity_v2-CreateIDPConnectorRequest) | [CreateIDPConnectorResponse](#identity_v2-CreateIDPConnectorResponse) |  |
| UpdateIDPConnector | [UpdateIDPConnectorRequest](#identity_v2-UpdateIDPConnectorRequest) | [UpdateIDPConnectorResponse](#identity_v2-UpdateIDPConnectorResponse) |  |
| ListIDPConnectors | [ListIDPConnectorsRequest](#identity_v2-ListIDPConnectorsRequest) | [ListIDPConnectorsResponse](#identity_v2-ListIDPConnectorsResponse) |  |
| GetIDPConnector | [GetIDPConnectorRequest](#identity_v2-GetIDPConnectorRequest) | [GetIDPConnectorResponse](#identity_v2-GetIDPConnectorResponse) |  |
| DeleteIDPConnector | [DeleteIDPConnectorRequest](#identity_v2-DeleteIDPConnectorRequest) | [DeleteIDPConnectorResponse](#identity_v2-DeleteIDPConnectorResponse) |  |
| CreateOIDCClient | [CreateOIDCClientRequest](#identity_v2-CreateOIDCClientRequest) | [CreateOIDCClientResponse](#identity_v2-CreateOIDCClientResponse) |  |
| UpdateOIDCClient | [UpdateOIDCClientRequest](#identity_v2-UpdateOIDCClientRequest) | [UpdateOIDCClientResponse](#identity_v2-UpdateOIDCClientResponse) |  |
| GetOIDCClient | [GetOIDCClientRequest](#identity_v2-GetOIDCClientRequest) | [GetOIDCClientResponse](#identity_v2-GetOIDCClientResponse) |  |
| ListOIDCClients | [ListOIDCClientsRequest](#identity_v2-ListOIDCClientsRequest) | [ListOIDCClientsResponse](#identity_v2-ListOIDCClientsResponse) |  |
| DeleteOIDCClient | [DeleteOIDCClientRequest](#identity_v2-DeleteOIDCClientRequest) | [DeleteOIDCClientResponse](#identity_v2-DeleteOIDCClientResponse) |  |
| DeleteAll | [DeleteAllRequest](#identity_v2-DeleteAllRequest) | [DeleteAllResponse](#identity_v2-DeleteAllResponse) |  |

 



<a name="internal_clusterstate_v2-5-0_commit_info-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/clusterstate/v2.5.0/commit_info.proto



<a name="v2_5_0-CommitInfo"></a>

### CommitInfo
CommitInfo is the main data structure representing a commit in etcd


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [pfs_v2.Commit](#pfs_v2-Commit) |  |  |
| origin | [pfs_v2.CommitOrigin](#pfs_v2-CommitOrigin) |  |  |
| description | [string](#string) |  | description is a user-provided script describing this commit |
| parent_commit | [pfs_v2.Commit](#pfs_v2-Commit) |  |  |
| child_commits | [pfs_v2.Commit](#pfs_v2-Commit) | repeated |  |
| started | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| finishing | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| finished | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| direct_provenance | [pfs_v2.Branch](#pfs_v2-Branch) | repeated |  |
| error | [string](#string) |  |  |
| size_bytes_upper_bound | [int64](#int64) |  |  |
| details | [pfs_v2.CommitInfo.Details](#pfs_v2-CommitInfo-Details) |  |  |





 

 

 

 



<a name="internal_collection_test-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/collection/test.proto



<a name="common-TestItem"></a>

### TestItem



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| value | [string](#string) |  |  |
| data | [string](#string) |  |  |





 

 

 

 



<a name="internal_config_config-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/config/config.proto



<a name="config_v2-Config"></a>

### Config
Config specifies the pachyderm config that is read and interpreted by the
pachctl command-line tool. Right now, this is stored at
$HOME/.pachyderm/config.

Different versions of the pachyderm config are specified as subfields of this
message (this allows us to make significant changes to the config structure
without breaking existing users by defining a new config version).

These structures are stored in a JSON format, so it should be safe to modify
fields as long as compatibility is ensured with previous versions.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| v1 | [ConfigV1](#config_v2-ConfigV1) |  | Configuration options. Exactly one of these fields should be set (depending on which version of the config is being used) |
| v2 | [ConfigV2](#config_v2-ConfigV2) |  |  |






<a name="config_v2-ConfigV1"></a>

### ConfigV1
ConfigV1 specifies v1 of the pachyderm config (June 30 2017 - June 2019)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pachd_address | [string](#string) |  | A host:port pointing pachd at a pachyderm cluster. |
| server_cas | [string](#string) |  | Trusted root certificates (overrides installed certificates), formatted as base64-encoded PEM |
| session_token | [string](#string) |  | A secret token identifying the current pachctl user within their pachyderm cluster. This is included in all RPCs sent by pachctl, and used to determine if pachctl actions are authorized. |
| active_transaction | [string](#string) |  | The currently active transaction for batching together pachctl commands. This can be set or cleared via many of the `pachctl * transaction` commands. This is the ID of the transaction object stored in the pachyderm etcd. |






<a name="config_v2-ConfigV2"></a>

### ConfigV2
ConfigV2 specifies v2 of the pachyderm config (June 2019 - present)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| active_context | [string](#string) |  |  |
| active_enterprise_context | [string](#string) |  |  |
| contexts | [ConfigV2.ContextsEntry](#config_v2-ConfigV2-ContextsEntry) | repeated |  |
| metrics | [bool](#bool) |  |  |
| max_shell_completions | [int64](#int64) |  |  |






<a name="config_v2-ConfigV2-ContextsEntry"></a>

### ConfigV2.ContextsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [Context](#config_v2-Context) |  |  |






<a name="config_v2-Context"></a>

### Context



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| source | [ContextSource](#config_v2-ContextSource) |  | Where this context came from |
| pachd_address | [string](#string) |  | The hostname or IP address pointing pachd at a pachyderm cluster. |
| server_cas | [string](#string) |  | Trusted root certificates (overrides installed certificates), formatted as base64-encoded PEM. |
| session_token | [string](#string) |  | A secret token identifying the current pachctl user within their pachyderm cluster. This is included in all RPCs sent by pachctl, and used to determine if pachctl actions are authorized. |
| active_transaction | [string](#string) |  | The currently active transaction for batching together pachctl commands. This can be set or cleared via many of the `pachctl * transaction` commands. This is the ID of the transaction object stored in the pachyderm etcd. |
| cluster_name | [string](#string) |  | The k8s cluster name - used to construct a k8s context. |
| auth_info | [string](#string) |  | The k8s auth info - used to construct a k8s context. |
| namespace | [string](#string) |  | The k8s namespace - used to construct a k8s context. |
| port_forwarders | [Context.PortForwardersEntry](#config_v2-Context-PortForwardersEntry) | repeated | A mapping of service -&gt; port number, when port forwarding is running for this context. |
| cluster_deployment_id | [string](#string) |  | A unique ID for the cluster deployment. At client initialization time, we ensure this is the same as what the cluster reports back, to prevent us from connecting to the wrong cluster. |
| enterprise_server | [bool](#bool) |  | A boolean that records whether the context points at an enterprise server. If false, the context points at a stand-alone pachd. |
| project | [string](#string) |  | The current project. |






<a name="config_v2-Context-PortForwardersEntry"></a>

### Context.PortForwardersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [uint32](#uint32) |  |  |





 


<a name="config_v2-ContextSource"></a>

### ContextSource


| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 |  |
| CONFIG_V1 | 1 |  |
| HUB | 2 |  |
| IMPORTED | 3 |  |


 

 

 



<a name="internal_metrics_metrics-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/metrics/metrics.proto



<a name="metrics-Metrics"></a>

### Metrics



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_id | [string](#string) |  |  |
| pod_id | [string](#string) |  |  |
| nodes | [int64](#int64) |  |  |
| version | [string](#string) |  |  |
| repos | [int64](#int64) |  | Number of repos |
| commits | [int64](#int64) |  | Number of commits -- not used |
| files | [int64](#int64) |  | Number of files -- not used |
| bytes | [uint64](#uint64) |  | Number of bytes in all repos |
| jobs | [int64](#int64) |  | Number of jobs |
| pipelines | [int64](#int64) |  | Number of pipelines in the cluster -- not the same as DAG |
| archived_commits | [int64](#int64) |  | Number of archived commit -- not used |
| cancelled_commits | [int64](#int64) |  | Number of cancelled commits -- not used |
| activation_code | [string](#string) |  | Activation code |
| max_branches | [uint64](#uint64) |  | Max branches in across all the repos |
| pps_spout | [int64](#int64) |  | Number of spout pipelines |
| pps_spout_service | [int64](#int64) |  | Number of spout services |
| cfg_egress | [int64](#int64) |  | Number of pipelines with Egress configured |
| cfg_standby | [int64](#int64) |  | Number of pipelines with Standby congigured |
| cfg_s3gateway | [int64](#int64) |  | Number of pipelines with S3 Gateway configured |
| cfg_services | [int64](#int64) |  | Number of pipelines with services configured |
| cfg_errcmd | [int64](#int64) |  | Number of pipelines with error cmd set |
| cfg_tfjob | [int64](#int64) |  | Number of pipelines with TFJobs configured |
| input_group | [int64](#int64) |  | Number of pipelines with group inputs |
| input_join | [int64](#int64) |  | Number of pipelines with join inputs |
| input_cross | [int64](#int64) |  | Number of pipelines with cross inputs |
| input_union | [int64](#int64) |  | Number of pipelines with union inputs |
| input_cron | [int64](#int64) |  | Number of pipelines with cron inputs |
| input_git | [int64](#int64) |  | Number of pipelines with git inputs |
| input_pfs | [int64](#int64) |  | Number of pfs inputs |
| input_commit | [int64](#int64) |  | Number of pfs inputs with commits |
| input_join_on | [int64](#int64) |  | Number of pfs inputs with join_on |
| input_outer_join | [int64](#int64) |  | Number of pipelines with outer joins |
| input_lazy | [int64](#int64) |  | Number of pipelines with lazy set |
| input_empty_files | [int64](#int64) |  | Number of pipelines with empty files set |
| input_s3 | [int64](#int64) |  | Number of pipelines with S3 input |
| input_trigger | [int64](#int64) |  | Number of pipelines with triggers set |
| resource_cpu_req | [float](#float) |  | Total CPU request across all pipelines |
| resource_cpu_req_max | [float](#float) |  | Max CPU resource requests set |
| resource_mem_req | [string](#string) |  | Sting of memory requests set across all pipelines |
| resource_gpu_req | [int64](#int64) |  | Total GPU requests across all pipelines |
| resource_gpu_req_max | [int64](#int64) |  | Max GPU request across all pipelines |
| resource_disk_req | [string](#string) |  | String of disk requests set across all pipelines |
| resource_cpu_limit | [float](#float) |  | Total CPU limits set across all pipelines |
| resource_cpu_limit_max | [float](#float) |  | Max CPU limit set |
| resource_mem_limit | [string](#string) |  | String of memory limits set |
| resource_gpu_limit | [int64](#int64) |  | Number of pipelines with |
| resource_gpu_limit_max | [int64](#int64) |  | Max GPU limit set |
| resource_disk_limit | [string](#string) |  | String of disk limits set across all pipelines |
| max_parallelism | [uint64](#uint64) |  | Max parallelism set |
| min_parallelism | [uint64](#uint64) |  | Min parallelism set |
| num_parallelism | [uint64](#uint64) |  | Number of pipelines with parallelism set |
| enterprise_failures | [int64](#int64) |  | Number of times a command has failed due to an enterprise check |
| pipeline_with_alerts | [bool](#bool) |  | pipelines with alerts indicator |





 

 

 

 



<a name="internal_pfsload_pfsload-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/pfsload/pfsload.proto



<a name="pfsload-CommitSpec"></a>

### CommitSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| count | [int64](#int64) |  |  |
| modifications | [ModificationSpec](#pfsload-ModificationSpec) | repeated |  |
| file_sources | [FileSourceSpec](#pfsload-FileSourceSpec) | repeated |  |
| validator | [ValidatorSpec](#pfsload-ValidatorSpec) |  |  |






<a name="pfsload-FileSourceSpec"></a>

### FileSourceSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| random | [RandomFileSourceSpec](#pfsload-RandomFileSourceSpec) |  |  |






<a name="pfsload-FrequencySpec"></a>

### FrequencySpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| count | [int64](#int64) |  |  |
| prob | [int64](#int64) |  |  |






<a name="pfsload-ModificationSpec"></a>

### ModificationSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| count | [int64](#int64) |  |  |
| put_file | [PutFileSpec](#pfsload-PutFileSpec) |  |  |






<a name="pfsload-PutFileSpec"></a>

### PutFileSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| count | [int64](#int64) |  |  |
| source | [string](#string) |  |  |






<a name="pfsload-PutFileTask"></a>

### PutFileTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| count | [int64](#int64) |  |  |
| file_source | [FileSourceSpec](#pfsload-FileSourceSpec) |  |  |
| seed | [int64](#int64) |  |  |
| auth_token | [string](#string) |  |  |






<a name="pfsload-PutFileTaskResult"></a>

### PutFileTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |
| hash | [bytes](#bytes) |  |  |






<a name="pfsload-RandomDirectorySpec"></a>

### RandomDirectorySpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| depth | [SizeSpec](#pfsload-SizeSpec) |  |  |
| run | [int64](#int64) |  |  |






<a name="pfsload-RandomFileSourceSpec"></a>

### RandomFileSourceSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| directory | [RandomDirectorySpec](#pfsload-RandomDirectorySpec) |  |  |
| sizes | [SizeSpec](#pfsload-SizeSpec) | repeated |  |
| increment_path | [bool](#bool) |  |  |






<a name="pfsload-SizeSpec"></a>

### SizeSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| min_size | [int64](#int64) |  |  |
| max_size | [int64](#int64) |  |  |
| prob | [int64](#int64) |  |  |






<a name="pfsload-State"></a>

### State



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commits | [State.Commit](#pfsload-State-Commit) | repeated |  |






<a name="pfsload-State-Commit"></a>

### State.Commit



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [pfs_v2.Commit](#pfs_v2-Commit) |  |  |
| hash | [bytes](#bytes) |  |  |






<a name="pfsload-ValidatorSpec"></a>

### ValidatorSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| frequency | [FrequencySpec](#pfsload-FrequencySpec) |  |  |





 

 

 

 



<a name="internal_ppsdb_ppsdb-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/ppsdb/ppsdb.proto



<a name="pps_v2-ClusterDefaultsWrapper"></a>

### ClusterDefaultsWrapper



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| json | [string](#string) |  |  |






<a name="pps_v2-ProjectDefaultsWrapper"></a>

### ProjectDefaultsWrapper



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| json | [string](#string) |  |  |





 

 

 

 



<a name="internal_ppsload_ppsload-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/ppsload/ppsload.proto



<a name="ppsload-State"></a>

### State



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| branch | [pfs_v2.Branch](#pfs_v2-Branch) |  |  |
| pfs_state_id | [string](#string) |  |  |





 

 

 

 



<a name="internal_storage_chunk_chunk-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/storage/chunk/chunk.proto



<a name="chunk-DataRef"></a>

### DataRef
DataRef is a reference to data within a chunk.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ref | [Ref](#chunk-Ref) |  | The chunk the referenced data is located in. |
| hash | [bytes](#bytes) |  | The hash of the data being referenced. |
| offset_bytes | [int64](#int64) |  | The offset and size used for accessing the data within the chunk. |
| size_bytes | [int64](#int64) |  |  |






<a name="chunk-Ref"></a>

### Ref



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [bytes](#bytes) |  |  |
| size_bytes | [int64](#int64) |  |  |
| edge | [bool](#bool) |  |  |
| dek | [bytes](#bytes) |  |  |
| encryption_algo | [EncryptionAlgo](#chunk-EncryptionAlgo) |  |  |
| compression_algo | [CompressionAlgo](#chunk-CompressionAlgo) |  |  |





 


<a name="chunk-CompressionAlgo"></a>

### CompressionAlgo


| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 |  |
| GZIP_BEST_SPEED | 1 |  |



<a name="chunk-EncryptionAlgo"></a>

### EncryptionAlgo


| Name | Number | Description |
| ---- | ------ | ----------- |
| ENCRYPTION_ALGO_UNKNOWN | 0 |  |
| CHACHA20 | 1 |  |


 

 

 



<a name="internal_storage_fileset_fileset-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/storage/fileset/fileset.proto



<a name="fileset-Composite"></a>

### Composite



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| layers | [string](#string) | repeated |  |






<a name="fileset-Metadata"></a>

### Metadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| primitive | [Primitive](#fileset-Primitive) |  |  |
| composite | [Composite](#fileset-Composite) |  |  |






<a name="fileset-Primitive"></a>

### Primitive



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deletive | [index.Index](#index-Index) |  |  |
| additive | [index.Index](#index-Index) |  |  |
| size_bytes | [int64](#int64) |  |  |






<a name="fileset-TestCacheValue"></a>

### TestCacheValue



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |





 

 

 

 



<a name="internal_storage_fileset_index_index-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/storage/fileset/index/index.proto



<a name="index-File"></a>

### File



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| datum | [string](#string) |  |  |
| data_refs | [chunk.DataRef](#chunk-DataRef) | repeated |  |






<a name="index-Index"></a>

### Index
Index stores an index to and metadata about a range of files or a file.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |
| range | [Range](#index-Range) |  | NOTE: range and file are mutually exclusive. |
| file | [File](#index-File) |  |  |
| num_files | [int64](#int64) |  | NOTE: num_files and size_bytes did not exist in older versions of 2.x, so they will not be set. |
| size_bytes | [int64](#int64) |  |  |






<a name="index-Range"></a>

### Range



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| offset | [int64](#int64) |  |  |
| last_path | [string](#string) |  |  |
| chunk_ref | [chunk.DataRef](#chunk-DataRef) |  |  |





 

 

 

 



<a name="internal_task_task-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/task/task.proto



<a name="task-Claim"></a>

### Claim







<a name="task-Group"></a>

### Group







<a name="task-Task"></a>

### Task
TODO: Consider splitting this up into separate structures for each state in a oneof.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| state | [State](#task-State) |  |  |
| input | [google.protobuf.Any](#google-protobuf-Any) |  |  |
| output | [google.protobuf.Any](#google-protobuf-Any) |  |  |
| reason | [string](#string) |  |  |
| index | [int64](#int64) |  |  |






<a name="task-TestTask"></a>

### TestTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |





 


<a name="task-State"></a>

### State


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATE_UNKNOWN | 0 |  |
| RUNNING | 1 |  |
| SUCCESS | 2 |  |
| FAILURE | 3 |  |


 

 

 



<a name="internal_tracing_extended_extended_trace-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## internal/tracing/extended/extended_trace.proto



<a name="extended-TraceProto"></a>

### TraceProto
TraceProto contains information identifying a Jaeger trace. It&#39;s used to
propagate traces that follow the lifetime of a long operation (e.g. creating
a pipeline or running a job), and which live longer than any single RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| serialized_trace | [TraceProto.SerializedTraceEntry](#extended-TraceProto-SerializedTraceEntry) | repeated | serialized_trace contains the info identifying a trace in Jaeger (a (trace ID, span ID, sampled) tuple, basically) |
| project | [string](#string) |  |  |
| pipeline | [string](#string) |  | pipeline specifies the target pipeline of this trace; this would be set for a trace created by &#39;pachctl create-pipeline&#39; or &#39;pachctl update-pipeline&#39; and would include the kubernetes RPCs involved in creating a pipeline |






<a name="extended-TraceProto-SerializedTraceEntry"></a>

### TraceProto.SerializedTraceEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 

 

 

 



<a name="license_license-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## license/license.proto



<a name="license_v2-ActivateRequest"></a>

### ActivateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| activation_code | [string](#string) |  | activation_code is a Pachyderm enterprise activation code. New users can obtain trial activation codes |
| expires | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | expires is a timestamp indicating when this activation code will expire. This should not generally be set (it&#39;s primarily used for testing), and is only applied if it&#39;s earlier than the signed expiration time in &#39;activation_code&#39;. |






<a name="license_v2-ActivateResponse"></a>

### ActivateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [enterprise_v2.TokenInfo](#enterprise_v2-TokenInfo) |  |  |






<a name="license_v2-AddClusterRequest"></a>

### AddClusterRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | id is the unique, immutable identifier for this cluster |
| address | [string](#string) |  | address is the public GPRC address where the cluster can be reached |
| secret | [string](#string) |  | If set, secret specifies the shared secret this cluster will use to authenticate to the license server. Otherwise a secret will be generated and returned in the response. |
| user_address | [string](#string) |  | The pachd address for users to connect to. |
| cluster_deployment_id | [string](#string) |  | The deployment ID value generated by each Cluster |
| enterprise_server | [bool](#bool) |  | This field indicates whether the address points to an enterprise server |






<a name="license_v2-AddClusterResponse"></a>

### AddClusterResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secret | [string](#string) |  |  |






<a name="license_v2-ClusterStatus"></a>

### ClusterStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| address | [string](#string) |  |  |
| version | [string](#string) |  |  |
| auth_enabled | [bool](#bool) |  |  |
| client_id | [string](#string) |  |  |
| last_heartbeat | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="license_v2-DeactivateRequest"></a>

### DeactivateRequest







<a name="license_v2-DeactivateResponse"></a>

### DeactivateResponse







<a name="license_v2-DeleteAllRequest"></a>

### DeleteAllRequest







<a name="license_v2-DeleteAllResponse"></a>

### DeleteAllResponse







<a name="license_v2-DeleteClusterRequest"></a>

### DeleteClusterRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="license_v2-DeleteClusterResponse"></a>

### DeleteClusterResponse







<a name="license_v2-GetActivationCodeRequest"></a>

### GetActivationCodeRequest







<a name="license_v2-GetActivationCodeResponse"></a>

### GetActivationCodeResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [enterprise_v2.State](#enterprise_v2-State) |  |  |
| info | [enterprise_v2.TokenInfo](#enterprise_v2-TokenInfo) |  |  |
| activation_code | [string](#string) |  |  |






<a name="license_v2-HeartbeatRequest"></a>

### HeartbeatRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| secret | [string](#string) |  |  |
| version | [string](#string) |  |  |
| auth_enabled | [bool](#bool) |  |  |
| client_id | [string](#string) |  |  |






<a name="license_v2-HeartbeatResponse"></a>

### HeartbeatResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| license | [enterprise_v2.LicenseRecord](#enterprise_v2-LicenseRecord) |  |  |






<a name="license_v2-ListClustersRequest"></a>

### ListClustersRequest







<a name="license_v2-ListClustersResponse"></a>

### ListClustersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clusters | [ClusterStatus](#license_v2-ClusterStatus) | repeated |  |






<a name="license_v2-ListUserClustersRequest"></a>

### ListUserClustersRequest







<a name="license_v2-ListUserClustersResponse"></a>

### ListUserClustersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| clusters | [UserClusterInfo](#license_v2-UserClusterInfo) | repeated |  |






<a name="license_v2-UpdateClusterRequest"></a>

### UpdateClusterRequest
Note: Updates of the enterprise-server field are not allowed. In the worst case, a user can recreate their cluster if they need the field updated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| address | [string](#string) |  |  |
| user_address | [string](#string) |  |  |
| cluster_deployment_id | [string](#string) |  |  |
| secret | [string](#string) |  |  |






<a name="license_v2-UpdateClusterResponse"></a>

### UpdateClusterResponse







<a name="license_v2-UserClusterInfo"></a>

### UserClusterInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| cluster_deployment_id | [string](#string) |  |  |
| address | [string](#string) |  |  |
| enterprise_server | [bool](#bool) |  |  |





 

 

 


<a name="license_v2-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Activate | [ActivateRequest](#license_v2-ActivateRequest) | [ActivateResponse](#license_v2-ActivateResponse) | Activate enables the license service by setting the enterprise activation code to serve. |
| GetActivationCode | [GetActivationCodeRequest](#license_v2-GetActivationCodeRequest) | [GetActivationCodeResponse](#license_v2-GetActivationCodeResponse) |  |
| DeleteAll | [DeleteAllRequest](#license_v2-DeleteAllRequest) | [DeleteAllResponse](#license_v2-DeleteAllResponse) | DeleteAll deactivates the server and removes all data. |
| AddCluster | [AddClusterRequest](#license_v2-AddClusterRequest) | [AddClusterResponse](#license_v2-AddClusterResponse) | CRUD operations for the pachds registered with this server. |
| DeleteCluster | [DeleteClusterRequest](#license_v2-DeleteClusterRequest) | [DeleteClusterResponse](#license_v2-DeleteClusterResponse) |  |
| ListClusters | [ListClustersRequest](#license_v2-ListClustersRequest) | [ListClustersResponse](#license_v2-ListClustersResponse) |  |
| UpdateCluster | [UpdateClusterRequest](#license_v2-UpdateClusterRequest) | [UpdateClusterResponse](#license_v2-UpdateClusterResponse) |  |
| Heartbeat | [HeartbeatRequest](#license_v2-HeartbeatRequest) | [HeartbeatResponse](#license_v2-HeartbeatResponse) | Heartbeat is the RPC registered pachds make to the license server to communicate their status and fetch updates. |
| ListUserClusters | [ListUserClustersRequest](#license_v2-ListUserClustersRequest) | [ListUserClustersResponse](#license_v2-ListUserClustersResponse) | Lists all clusters available to user |

 



<a name="logs_logs-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## logs/logs.proto



<a name="logs-AdminLogQuery"></a>

### AdminLogQuery



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| logql | [string](#string) |  | Arbitrary LogQL query |
| pod | [string](#string) |  | A pod&#39;s logs (all containers) |
| pod_container | [PodContainer](#logs-PodContainer) |  | One container |
| app | [string](#string) |  | One &#34;app&#34; (logql -&gt; {app=X}) |
| master | [PipelineLogQuery](#logs-PipelineLogQuery) |  | All master worker lines from a pipeline |
| storage | [PipelineLogQuery](#logs-PipelineLogQuery) |  | All storage container lines from a pipeline |
| user | [UserLogQuery](#logs-UserLogQuery) |  | All worker lines from a pipeline/job |






<a name="logs-GetLogsRequest"></a>

### GetLogsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| query | [LogQuery](#logs-LogQuery) |  |  |
| filter | [LogFilter](#logs-LogFilter) |  |  |
| tail | [bool](#bool) |  |  |
| want_paging_hint | [bool](#bool) |  |  |
| log_format | [LogFormat](#logs-LogFormat) |  |  |






<a name="logs-GetLogsResponse"></a>

### GetLogsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| paging_hint | [PagingHint](#logs-PagingHint) |  |  |
| log | [LogMessage](#logs-LogMessage) |  |  |






<a name="logs-LogFilter"></a>

### LogFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| time_range | [TimeRangeLogFilter](#logs-TimeRangeLogFilter) |  |  |
| limit | [uint64](#uint64) |  |  |
| regex | [RegexLogFilter](#logs-RegexLogFilter) |  |  |
| level | [LogLevel](#logs-LogLevel) |  | Minimum log level to return; worker will always run at level debug, but setting INFO here restores original behavior |






<a name="logs-LogMessage"></a>

### LogMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| verbatim | [VerbatimLogMessage](#logs-VerbatimLogMessage) |  |  |
| json | [ParsedJSONLogMessage](#logs-ParsedJSONLogMessage) |  |  |
| pps_log_message | [pps_v2.LogMessage](#pps_v2-LogMessage) |  |  |






<a name="logs-LogQuery"></a>

### LogQuery



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [UserLogQuery](#logs-UserLogQuery) |  |  |
| admin | [AdminLogQuery](#logs-AdminLogQuery) |  |  |






<a name="logs-PagingHint"></a>

### PagingHint



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| older | [GetLogsRequest](#logs-GetLogsRequest) |  |  |
| newer | [GetLogsRequest](#logs-GetLogsRequest) |  |  |






<a name="logs-ParsedJSONLogMessage"></a>

### ParsedJSONLogMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| verbatim | [VerbatimLogMessage](#logs-VerbatimLogMessage) |  | The verbatim line from Loki |
| object | [google.protobuf.Struct](#google-protobuf-Struct) |  | A raw JSON parse of the entire line |
| native_timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | If a parseable timestamp was found in `fields` |
| pps_log_message | [pps_v2.LogMessage](#pps_v2-LogMessage) |  | For code that wants to filter on pipeline/job/etc |






<a name="logs-PipelineJobLogQuery"></a>

### PipelineJobLogQuery



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [PipelineLogQuery](#logs-PipelineLogQuery) |  |  |
| job | [string](#string) |  |  |






<a name="logs-PipelineLogQuery"></a>

### PipelineLogQuery



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  |  |
| pipeline | [string](#string) |  |  |






<a name="logs-PodContainer"></a>

### PodContainer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pod | [string](#string) |  |  |
| container | [string](#string) |  |  |






<a name="logs-RegexLogFilter"></a>

### RegexLogFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pattern | [string](#string) |  |  |
| negate | [bool](#bool) |  |  |






<a name="logs-TimeRangeLogFilter"></a>

### TimeRangeLogFilter



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Can be null |
| until | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Can be null |






<a name="logs-UserLogQuery"></a>

### UserLogQuery
Only returns &#34;user&#34; logs


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  | All pipelines in the project |
| pipeline | [PipelineLogQuery](#logs-PipelineLogQuery) |  | One pipeline in a project |
| datum | [string](#string) |  | One datum. |
| job | [string](#string) |  | One job, across pipelines and projects |
| pipeline_job | [PipelineJobLogQuery](#logs-PipelineJobLogQuery) |  | One job in one pipeline |






<a name="logs-VerbatimLogMessage"></a>

### VerbatimLogMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| line | [bytes](#bytes) |  |  |
| timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |





 


<a name="logs-LogFormat"></a>

### LogFormat


| Name | Number | Description |
| ---- | ------ | ----------- |
| LOG_FORMAT_UNKNOWN | 0 | error |
| LOG_FORMAT_VERBATIM_WITH_TIMESTAMP | 1 |  |
| LOG_FORMAT_PARSED_JSON | 2 |  |
| LOG_FORMAT_PPS_LOGMESSAGE | 3 |  |



<a name="logs-LogLevel"></a>

### LogLevel


| Name | Number | Description |
| ---- | ------ | ----------- |
| LOG_LEVEL_DEBUG | 0 |  |
| LOG_LEVEL_INFO | 1 |  |
| LOG_LEVEL_ERROR | 2 |  |


 

 


<a name="logs-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetLogs | [GetLogsRequest](#logs-GetLogsRequest) | [GetLogsResponse](#logs-GetLogsResponse) stream |  |

 



<a name="pfs_pfs-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## pfs/pfs.proto



<a name="pfs_v2-ActivateAuthRequest"></a>

### ActivateAuthRequest







<a name="pfs_v2-ActivateAuthResponse"></a>

### ActivateAuthResponse







<a name="pfs_v2-AddFile"></a>

### AddFile



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |
| datum | [string](#string) |  |  |
| raw | [google.protobuf.BytesValue](#google-protobuf-BytesValue) |  |  |
| url | [AddFile.URLSource](#pfs_v2-AddFile-URLSource) |  |  |






<a name="pfs_v2-AddFile-URLSource"></a>

### AddFile.URLSource



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| URL | [string](#string) |  |  |
| recursive | [bool](#bool) |  |  |
| concurrency | [uint32](#uint32) |  |  |






<a name="pfs_v2-AddFileSetRequest"></a>

### AddFileSetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |
| file_set_id | [string](#string) |  |  |






<a name="pfs_v2-AuthInfo"></a>

### AuthInfo
AuthInfo includes the caller&#39;s access scope for a resource, and is returned
by services like ListRepo, InspectRepo, and ListProject, but is not persisted in the database.
It&#39;s used by the Pachyderm dashboard to render repo access appropriately.
To set a user&#39;s auth scope for a resource, use the Pachyderm Auth API (in src/auth/auth.proto)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| permissions | [auth_v2.Permission](#auth_v2-Permission) | repeated | The callers access level to the relevant resource. These are very granular permissions - for the end user it makes sense to show them the roles they have instead. |
| roles | [string](#string) | repeated | The caller&#39;s roles on the relevant resource. This includes inherited roles from the cluster, project, group membership, etc. |






<a name="pfs_v2-Branch"></a>

### Branch



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repo | [Repo](#pfs_v2-Repo) |  |  |
| name | [string](#string) |  |  |






<a name="pfs_v2-BranchInfo"></a>

### BranchInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| branch | [Branch](#pfs_v2-Branch) |  |  |
| head | [Commit](#pfs_v2-Commit) |  |  |
| provenance | [Branch](#pfs_v2-Branch) | repeated |  |
| subvenance | [Branch](#pfs_v2-Branch) | repeated |  |
| direct_provenance | [Branch](#pfs_v2-Branch) | repeated |  |
| trigger | [Trigger](#pfs_v2-Trigger) |  |  |






<a name="pfs_v2-CheckStorageRequest"></a>

### CheckStorageRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| read_chunk_data | [bool](#bool) |  |  |
| chunk_begin | [bytes](#bytes) |  |  |
| chunk_end | [bytes](#bytes) |  |  |






<a name="pfs_v2-CheckStorageResponse"></a>

### CheckStorageResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| chunk_object_count | [int64](#int64) |  |  |






<a name="pfs_v2-ClearCacheRequest"></a>

### ClearCacheRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tag_prefix | [string](#string) |  |  |






<a name="pfs_v2-ClearCommitRequest"></a>

### ClearCommitRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |






<a name="pfs_v2-Commit"></a>

### Commit
Commit is a reference to a commit (e.g. the collection of branches and the
collection of currently-open commits in etcd are collections of Commit
protos)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repo | [Repo](#pfs_v2-Repo) |  |  |
| id | [string](#string) |  |  |
| branch | [Branch](#pfs_v2-Branch) |  | only used by the client |






<a name="pfs_v2-CommitInfo"></a>

### CommitInfo
CommitInfo is the main data structure representing a commit in etcd


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |
| origin | [CommitOrigin](#pfs_v2-CommitOrigin) |  |  |
| description | [string](#string) |  | description is a user-provided script describing this commit |
| parent_commit | [Commit](#pfs_v2-Commit) |  |  |
| child_commits | [Commit](#pfs_v2-Commit) | repeated |  |
| started | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| finishing | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| finished | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| direct_provenance | [Commit](#pfs_v2-Commit) | repeated |  |
| error | [string](#string) |  |  |
| size_bytes_upper_bound | [int64](#int64) |  |  |
| details | [CommitInfo.Details](#pfs_v2-CommitInfo-Details) |  |  |






<a name="pfs_v2-CommitInfo-Details"></a>

### CommitInfo.Details
Details are only provided when explicitly requested


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size_bytes | [int64](#int64) |  |  |
| compacting_time | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| validating_time | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="pfs_v2-CommitOrigin"></a>

### CommitOrigin



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| kind | [OriginKind](#pfs_v2-OriginKind) |  |  |






<a name="pfs_v2-CommitSet"></a>

### CommitSet



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="pfs_v2-CommitSetInfo"></a>

### CommitSetInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit_set | [CommitSet](#pfs_v2-CommitSet) |  |  |
| commits | [CommitInfo](#pfs_v2-CommitInfo) | repeated |  |






<a name="pfs_v2-ComposeFileSetRequest"></a>

### ComposeFileSetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_ids | [string](#string) | repeated |  |
| ttl_seconds | [int64](#int64) |  |  |
| compact | [bool](#bool) |  |  |






<a name="pfs_v2-CopyFile"></a>

### CopyFile



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dst | [string](#string) |  |  |
| datum | [string](#string) |  |  |
| src | [File](#pfs_v2-File) |  |  |
| append | [bool](#bool) |  |  |






<a name="pfs_v2-CreateBranchRequest"></a>

### CreateBranchRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| head | [Commit](#pfs_v2-Commit) |  |  |
| branch | [Branch](#pfs_v2-Branch) |  |  |
| provenance | [Branch](#pfs_v2-Branch) | repeated |  |
| trigger | [Trigger](#pfs_v2-Trigger) |  |  |
| new_commit_set | [bool](#bool) |  | overrides the default behavior of using the same CommitSet as &#39;head&#39; |






<a name="pfs_v2-CreateFileSetResponse"></a>

### CreateFileSetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |






<a name="pfs_v2-CreateProjectRequest"></a>

### CreateProjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [Project](#pfs_v2-Project) |  |  |
| description | [string](#string) |  |  |
| update | [bool](#bool) |  |  |






<a name="pfs_v2-CreateRepoRequest"></a>

### CreateRepoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repo | [Repo](#pfs_v2-Repo) |  |  |
| description | [string](#string) |  |  |
| update | [bool](#bool) |  |  |






<a name="pfs_v2-DeleteBranchRequest"></a>

### DeleteBranchRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| branch | [Branch](#pfs_v2-Branch) |  |  |
| force | [bool](#bool) |  |  |






<a name="pfs_v2-DeleteFile"></a>

### DeleteFile



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  |  |
| datum | [string](#string) |  |  |






<a name="pfs_v2-DeleteProjectRequest"></a>

### DeleteProjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [Project](#pfs_v2-Project) |  |  |
| force | [bool](#bool) |  |  |






<a name="pfs_v2-DeleteRepoRequest"></a>

### DeleteRepoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repo | [Repo](#pfs_v2-Repo) |  |  |
| force | [bool](#bool) |  |  |






<a name="pfs_v2-DeleteRepoResponse"></a>

### DeleteRepoResponse
DeleteRepoResponse returns the repos that were deleted by a DeleteRepo call.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deleted | [bool](#bool) |  | The repos that were deleted, perhaps none. |






<a name="pfs_v2-DeleteReposRequest"></a>

### DeleteReposRequest
DeleteReposRequest is used to delete more than one repo at once.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| projects | [Project](#pfs_v2-Project) | repeated | All repos in each project will be deleted if the caller has permission. |
| force | [bool](#bool) |  |  |
| all | [bool](#bool) |  | If all is set, then all repos in all projects will be deleted if the caller has permission. |






<a name="pfs_v2-DeleteReposResponse"></a>

### DeleteReposResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repos | [Repo](#pfs_v2-Repo) | repeated |  |






<a name="pfs_v2-DiffFileRequest"></a>

### DiffFileRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| new_file | [File](#pfs_v2-File) |  |  |
| old_file | [File](#pfs_v2-File) |  | OldFile may be left nil in which case the same path in the parent of NewFile&#39;s commit will be used. |
| shallow | [bool](#bool) |  |  |






<a name="pfs_v2-DiffFileResponse"></a>

### DiffFileResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| new_file | [FileInfo](#pfs_v2-FileInfo) |  |  |
| old_file | [FileInfo](#pfs_v2-FileInfo) |  |  |






<a name="pfs_v2-DropCommitRequest"></a>

### DropCommitRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |
| recursive | [bool](#bool) |  | Setting recursive to true indicates that the drop should be applied recursively to subvenant commits. If recursive is set to false and the provided commit has subvenant commits, the drop will fail. |






<a name="pfs_v2-DropCommitResponse"></a>

### DropCommitResponse







<a name="pfs_v2-DropCommitSetRequest"></a>

### DropCommitSetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit_set | [CommitSet](#pfs_v2-CommitSet) |  |  |






<a name="pfs_v2-EgressRequest"></a>

### EgressRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |
| object_storage | [ObjectStorageEgress](#pfs_v2-ObjectStorageEgress) |  |  |
| sql_database | [SQLDatabaseEgress](#pfs_v2-SQLDatabaseEgress) |  |  |






<a name="pfs_v2-EgressResponse"></a>

### EgressResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| object_storage | [EgressResponse.ObjectStorageResult](#pfs_v2-EgressResponse-ObjectStorageResult) |  |  |
| sql_database | [EgressResponse.SQLDatabaseResult](#pfs_v2-EgressResponse-SQLDatabaseResult) |  |  |






<a name="pfs_v2-EgressResponse-ObjectStorageResult"></a>

### EgressResponse.ObjectStorageResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bytes_written | [int64](#int64) |  |  |






<a name="pfs_v2-EgressResponse-SQLDatabaseResult"></a>

### EgressResponse.SQLDatabaseResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rows_written | [EgressResponse.SQLDatabaseResult.RowsWrittenEntry](#pfs_v2-EgressResponse-SQLDatabaseResult-RowsWrittenEntry) | repeated |  |






<a name="pfs_v2-EgressResponse-SQLDatabaseResult-RowsWrittenEntry"></a>

### EgressResponse.SQLDatabaseResult.RowsWrittenEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [int64](#int64) |  |  |






<a name="pfs_v2-File"></a>

### File



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |
| path | [string](#string) |  |  |
| datum | [string](#string) |  |  |






<a name="pfs_v2-FileInfo"></a>

### FileInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file | [File](#pfs_v2-File) |  |  |
| file_type | [FileType](#pfs_v2-FileType) |  |  |
| committed | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| size_bytes | [int64](#int64) |  |  |
| hash | [bytes](#bytes) |  |  |






<a name="pfs_v2-FindCommitsRequest"></a>

### FindCommitsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start | [Commit](#pfs_v2-Commit) |  |  |
| file_path | [string](#string) |  |  |
| limit | [uint32](#uint32) |  | a limit of 0 means there is no upper bound on the limit. |






<a name="pfs_v2-FindCommitsResponse"></a>

### FindCommitsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| found_commit | [Commit](#pfs_v2-Commit) |  |  |
| last_searched_commit | [Commit](#pfs_v2-Commit) |  |  |
| commits_searched | [uint32](#uint32) |  |  |






<a name="pfs_v2-FinishCommitRequest"></a>

### FinishCommitRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |
| description | [string](#string) |  | description is a user-provided string describing this commit. Setting this will overwrite the description set in StartCommit |
| error | [string](#string) |  |  |
| force | [bool](#bool) |  |  |






<a name="pfs_v2-FsckRequest"></a>

### FsckRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fix | [bool](#bool) |  |  |
| zombie_target | [Commit](#pfs_v2-Commit) |  |  |
| zombie_all | [bool](#bool) |  | run zombie data detection against all pipelines |






<a name="pfs_v2-FsckResponse"></a>

### FsckResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| fix | [string](#string) |  |  |
| error | [string](#string) |  |  |






<a name="pfs_v2-GetCacheRequest"></a>

### GetCacheRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |






<a name="pfs_v2-GetCacheResponse"></a>

### GetCacheResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | [google.protobuf.Any](#google-protobuf-Any) |  |  |






<a name="pfs_v2-GetFileRequest"></a>

### GetFileRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file | [File](#pfs_v2-File) |  |  |
| URL | [string](#string) |  |  |
| offset | [int64](#int64) |  |  |
| path_range | [PathRange](#pfs_v2-PathRange) |  | TODO: int64 size_bytes = 3; |






<a name="pfs_v2-GetFileSetRequest"></a>

### GetFileSetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |
| type | [GetFileSetRequest.FileSetType](#pfs_v2-GetFileSetRequest-FileSetType) |  |  |






<a name="pfs_v2-GlobFileRequest"></a>

### GlobFileRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |
| pattern | [string](#string) |  |  |
| path_range | [PathRange](#pfs_v2-PathRange) |  |  |






<a name="pfs_v2-InspectBranchRequest"></a>

### InspectBranchRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| branch | [Branch](#pfs_v2-Branch) |  |  |






<a name="pfs_v2-InspectCommitRequest"></a>

### InspectCommitRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |
| wait | [CommitState](#pfs_v2-CommitState) |  | Wait causes inspect commit to wait until the commit is in the desired state. |






<a name="pfs_v2-InspectCommitSetRequest"></a>

### InspectCommitSetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit_set | [CommitSet](#pfs_v2-CommitSet) |  |  |
| wait | [bool](#bool) |  | When true, wait until all commits in the set are finished |






<a name="pfs_v2-InspectFileRequest"></a>

### InspectFileRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file | [File](#pfs_v2-File) |  |  |






<a name="pfs_v2-InspectProjectRequest"></a>

### InspectProjectRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [Project](#pfs_v2-Project) |  |  |






<a name="pfs_v2-InspectProjectV2Request"></a>

### InspectProjectV2Request



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [Project](#pfs_v2-Project) |  |  |






<a name="pfs_v2-InspectProjectV2Response"></a>

### InspectProjectV2Response



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [ProjectInfo](#pfs_v2-ProjectInfo) |  |  |
| defaults_json | [string](#string) |  |  |






<a name="pfs_v2-InspectRepoRequest"></a>

### InspectRepoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repo | [Repo](#pfs_v2-Repo) |  |  |






<a name="pfs_v2-ListBranchRequest"></a>

### ListBranchRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repo | [Repo](#pfs_v2-Repo) |  |  |
| reverse | [bool](#bool) |  | Returns branches oldest to newest |






<a name="pfs_v2-ListCommitRequest"></a>

### ListCommitRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repo | [Repo](#pfs_v2-Repo) |  |  |
| from | [Commit](#pfs_v2-Commit) |  |  |
| to | [Commit](#pfs_v2-Commit) |  |  |
| number | [int64](#int64) |  |  |
| reverse | [bool](#bool) |  | Return commits oldest to newest |
| all | [bool](#bool) |  | Return commits of all kinds (without this, aliases are excluded) |
| origin_kind | [OriginKind](#pfs_v2-OriginKind) |  | Return only commits of this kind (mutually exclusive with all) |
| started_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Return commits started before this time |






<a name="pfs_v2-ListCommitSetRequest"></a>

### ListCommitSetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [Project](#pfs_v2-Project) |  |  |






<a name="pfs_v2-ListFileRequest"></a>

### ListFileRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file | [File](#pfs_v2-File) |  | File is the parent directory of the files we want to list. This sets the repo, the commit/branch, and path prefix of files we&#39;re interested in If the &#34;path&#34; field is omitted, a list of files at the top level of the repo is returned |
| paginationMarker | [File](#pfs_v2-File) |  | Marker for pagination. If set, the files that come after the marker in lexicographical order will be returned. If reverse is also set, the files that come before the marker in lexicographical order will be returned. |
| number | [int64](#int64) |  | Number of files to return |
| reverse | [bool](#bool) |  | If true, return files in reverse order |






<a name="pfs_v2-ListProjectRequest"></a>

### ListProjectRequest







<a name="pfs_v2-ListRepoRequest"></a>

### ListRepoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  | Type is the type of (system) repo that should be returned. An empty string requests all repos. |
| projects | [Project](#pfs_v2-Project) | repeated | Filters out repos whos project isn&#39;t represented. An empty list of projects doesn&#39;t filter repos by their project. |
| page | [RepoPage](#pfs_v2-RepoPage) |  | Specifies which page of repos should be returned. If page isn&#39;t specified, a single page containing all the relevant repos is returned. |






<a name="pfs_v2-ModifyFileRequest"></a>

### ModifyFileRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| set_commit | [Commit](#pfs_v2-Commit) |  |  |
| add_file | [AddFile](#pfs_v2-AddFile) |  |  |
| delete_file | [DeleteFile](#pfs_v2-DeleteFile) |  |  |
| copy_file | [CopyFile](#pfs_v2-CopyFile) |  |  |






<a name="pfs_v2-ObjectStorageEgress"></a>

### ObjectStorageEgress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | [string](#string) |  |  |






<a name="pfs_v2-PathRange"></a>

### PathRange



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| lower | [string](#string) |  |  |
| upper | [string](#string) |  |  |






<a name="pfs_v2-Project"></a>

### Project



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name="pfs_v2-ProjectInfo"></a>

### ProjectInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [Project](#pfs_v2-Project) |  |  |
| description | [string](#string) |  |  |
| auth_info | [AuthInfo](#pfs_v2-AuthInfo) |  |  |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="pfs_v2-PutCacheRequest"></a>

### PutCacheRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [google.protobuf.Any](#google-protobuf-Any) |  |  |
| file_set_ids | [string](#string) | repeated |  |
| tag | [string](#string) |  |  |






<a name="pfs_v2-RenewFileSetRequest"></a>

### RenewFileSetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |
| ttl_seconds | [int64](#int64) |  |  |






<a name="pfs_v2-Repo"></a>

### Repo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| type | [string](#string) |  |  |
| project | [Project](#pfs_v2-Project) |  |  |






<a name="pfs_v2-RepoInfo"></a>

### RepoInfo
RepoInfo is the main data structure representing a Repo in etcd


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repo | [Repo](#pfs_v2-Repo) |  |  |
| created | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| size_bytes_upper_bound | [int64](#int64) |  |  |
| description | [string](#string) |  |  |
| branches | [Branch](#pfs_v2-Branch) | repeated |  |
| auth_info | [AuthInfo](#pfs_v2-AuthInfo) |  | Set by ListRepo and InspectRepo if Pachyderm&#39;s auth system is active, but not stored in etcd. To set a user&#39;s auth scope for a repo, use the Pachyderm Auth API (in src/client/auth/auth.proto) |
| details | [RepoInfo.Details](#pfs_v2-RepoInfo-Details) |  |  |






<a name="pfs_v2-RepoInfo-Details"></a>

### RepoInfo.Details
Details are only provided when explicitly requested


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| size_bytes | [int64](#int64) |  |  |






<a name="pfs_v2-RepoPage"></a>

### RepoPage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order | [RepoPage.Ordering](#pfs_v2-RepoPage-Ordering) |  |  |
| page_size | [int64](#int64) |  |  |
| page_index | [int64](#int64) |  |  |






<a name="pfs_v2-SQLDatabaseEgress"></a>

### SQLDatabaseEgress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | [string](#string) |  |  |
| file_format | [SQLDatabaseEgress.FileFormat](#pfs_v2-SQLDatabaseEgress-FileFormat) |  |  |
| secret | [SQLDatabaseEgress.Secret](#pfs_v2-SQLDatabaseEgress-Secret) |  |  |






<a name="pfs_v2-SQLDatabaseEgress-FileFormat"></a>

### SQLDatabaseEgress.FileFormat



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [SQLDatabaseEgress.FileFormat.Type](#pfs_v2-SQLDatabaseEgress-FileFormat-Type) |  |  |
| columns | [string](#string) | repeated |  |






<a name="pfs_v2-SQLDatabaseEgress-Secret"></a>

### SQLDatabaseEgress.Secret



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| key | [string](#string) |  |  |






<a name="pfs_v2-ShardFileSetRequest"></a>

### ShardFileSetRequest
If both num_files and size_bytes are set, shards are created
based on whichever threshold is surpassed first. If a shard
configuration field (num_files, size_bytes) is unset, the
storage&#39;s default value is used.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |
| num_files | [int64](#int64) |  | Number of files targeted in each shard |
| size_bytes | [int64](#int64) |  | Size (in bytes) targeted for each shard |






<a name="pfs_v2-ShardFileSetResponse"></a>

### ShardFileSetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| shards | [PathRange](#pfs_v2-PathRange) | repeated |  |






<a name="pfs_v2-SquashCommitRequest"></a>

### SquashCommitRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [Commit](#pfs_v2-Commit) |  |  |
| recursive | [bool](#bool) |  | Setting recursive to true indicates that the squash should be applied recursively to subvenant commits. If recursive is set to false and the provided commit has subvenant commits, the squash will fail. |






<a name="pfs_v2-SquashCommitResponse"></a>

### SquashCommitResponse







<a name="pfs_v2-SquashCommitSetRequest"></a>

### SquashCommitSetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit_set | [CommitSet](#pfs_v2-CommitSet) |  |  |






<a name="pfs_v2-StartCommitRequest"></a>

### StartCommitRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| parent | [Commit](#pfs_v2-Commit) |  | parent may be empty in which case the commit that Branch points to will be used as the parent. If the branch does not exist, the commit will have no parent. |
| description | [string](#string) |  | description is a user-provided string describing this commit |
| branch | [Branch](#pfs_v2-Branch) |  |  |






<a name="pfs_v2-SubscribeCommitRequest"></a>

### SubscribeCommitRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| repo | [Repo](#pfs_v2-Repo) |  |  |
| branch | [string](#string) |  |  |
| from | [Commit](#pfs_v2-Commit) |  | only commits created since this commit are returned |
| state | [CommitState](#pfs_v2-CommitState) |  | Don&#39;t return commits until they&#39;re in (at least) the desired state. |
| all | [bool](#bool) |  | Return commits of all kinds (without this, aliases are excluded) |
| origin_kind | [OriginKind](#pfs_v2-OriginKind) |  | Return only commits of this kind (mutually exclusive with all) |






<a name="pfs_v2-Trigger"></a>

### Trigger
Trigger defines the conditions under which a head is moved, and to which
branch it is moved.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| branch | [string](#string) |  | Which branch this trigger refers to |
| all | [bool](#bool) |  | All indicates that all conditions must be satisfied before the trigger happens, otherwise any conditions being satisfied will trigger it. |
| rate_limit_spec | [string](#string) |  | Triggers if the rate limit spec (cron expression) has been satisfied since the last trigger. |
| size | [string](#string) |  | Triggers if there&#39;s been `size` new data added since the last trigger. |
| commits | [int64](#int64) |  | Triggers if there&#39;s been `commits` new commits added since the last trigger. |
| cron_spec | [string](#string) |  | Creates a background process which fires the trigger on the schedule provided by the cron spec. This condition is mutually exclusive with respect to the others, so setting this will result with the trigger only firing based on the cron schedule. |






<a name="pfs_v2-WalkFileRequest"></a>

### WalkFileRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file | [File](#pfs_v2-File) |  |  |
| paginationMarker | [File](#pfs_v2-File) |  | Marker for pagination. If set, the files that come after the marker in lexicographical order will be returned. If reverse is also set, the files that come before the marker in lexicographical order will be returned. |
| number | [int64](#int64) |  | Number of files to return |
| reverse | [bool](#bool) |  | If true, return files in reverse order |





 


<a name="pfs_v2-CommitState"></a>

### CommitState
CommitState describes the states a commit can be in.
The states are increasingly specific, i.e. a commit that is FINISHED also counts as STARTED.

| Name | Number | Description |
| ---- | ------ | ----------- |
| COMMIT_STATE_UNKNOWN | 0 |  |
| STARTED | 1 | The commit has been started, all commits satisfy this state. |
| READY | 2 | The commit has been started, and all of its provenant commits have been finished. |
| FINISHING | 3 | The commit is in the process of being finished. |
| FINISHED | 4 | The commit has been finished. |



<a name="pfs_v2-Delimiter"></a>

### Delimiter


| Name | Number | Description |
| ---- | ------ | ----------- |
| NONE | 0 |  |
| JSON | 1 |  |
| LINE | 2 |  |
| SQL | 3 |  |
| CSV | 4 |  |



<a name="pfs_v2-FileType"></a>

### FileType


| Name | Number | Description |
| ---- | ------ | ----------- |
| RESERVED | 0 |  |
| FILE | 1 |  |
| DIR | 2 |  |



<a name="pfs_v2-GetFileSetRequest-FileSetType"></a>

### GetFileSetRequest.FileSetType


| Name | Number | Description |
| ---- | ------ | ----------- |
| TOTAL | 0 |  |
| DIFF | 1 |  |



<a name="pfs_v2-OriginKind"></a>

### OriginKind
These are the different places where a commit may be originated from

| Name | Number | Description |
| ---- | ------ | ----------- |
| ORIGIN_KIND_UNKNOWN | 0 |  |
| USER | 1 |  |
| AUTO | 2 |  |
| FSCK | 3 |  |



<a name="pfs_v2-RepoPage-Ordering"></a>

### RepoPage.Ordering


| Name | Number | Description |
| ---- | ------ | ----------- |
| PROJECT_REPO | 0 |  |



<a name="pfs_v2-SQLDatabaseEgress-FileFormat-Type"></a>

### SQLDatabaseEgress.FileFormat.Type


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| CSV | 1 |  |
| JSON | 2 |  |
| PARQUET | 3 |  |


 

 


<a name="pfs_v2-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateRepo | [CreateRepoRequest](#pfs_v2-CreateRepoRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | CreateRepo creates a new repo. |
| InspectRepo | [InspectRepoRequest](#pfs_v2-InspectRepoRequest) | [RepoInfo](#pfs_v2-RepoInfo) | InspectRepo returns info about a repo. |
| ListRepo | [ListRepoRequest](#pfs_v2-ListRepoRequest) | [RepoInfo](#pfs_v2-RepoInfo) stream | ListRepo returns info about all repos. |
| DeleteRepo | [DeleteRepoRequest](#pfs_v2-DeleteRepoRequest) | [DeleteRepoResponse](#pfs_v2-DeleteRepoResponse) | DeleteRepo deletes a repo. |
| DeleteRepos | [DeleteReposRequest](#pfs_v2-DeleteReposRequest) | [DeleteReposResponse](#pfs_v2-DeleteReposResponse) | DeleteRepos deletes more than one repo at once. It attempts to delete every repo matching the DeleteReposRequest. When deleting all repos matching a project, any repos not deletable by the caller will remain, and the project will not be empty; this is not an error. The returned DeleteReposResponse will contain a list of all actually-deleted repos. |
| StartCommit | [StartCommitRequest](#pfs_v2-StartCommitRequest) | [Commit](#pfs_v2-Commit) | StartCommit creates a new write commit from a parent commit. |
| FinishCommit | [FinishCommitRequest](#pfs_v2-FinishCommitRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | FinishCommit turns a write commit into a read commit. |
| ClearCommit | [ClearCommitRequest](#pfs_v2-ClearCommitRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | ClearCommit removes all data from the commit. |
| InspectCommit | [InspectCommitRequest](#pfs_v2-InspectCommitRequest) | [CommitInfo](#pfs_v2-CommitInfo) | InspectCommit returns the info about a commit. |
| ListCommit | [ListCommitRequest](#pfs_v2-ListCommitRequest) | [CommitInfo](#pfs_v2-CommitInfo) stream | ListCommit returns info about all commits. |
| SubscribeCommit | [SubscribeCommitRequest](#pfs_v2-SubscribeCommitRequest) | [CommitInfo](#pfs_v2-CommitInfo) stream | SubscribeCommit subscribes for new commits on a given branch. |
| SquashCommit | [SquashCommitRequest](#pfs_v2-SquashCommitRequest) | [SquashCommitResponse](#pfs_v2-SquashCommitResponse) | SquashCommit squashes the provided commit into its children. |
| DropCommit | [DropCommitRequest](#pfs_v2-DropCommitRequest) | [DropCommitResponse](#pfs_v2-DropCommitResponse) | DropCommit drops the provided commit. |
| InspectCommitSet | [InspectCommitSetRequest](#pfs_v2-InspectCommitSetRequest) | [CommitInfo](#pfs_v2-CommitInfo) stream | InspectCommitSet returns the info about a CommitSet. |
| ListCommitSet | [ListCommitSetRequest](#pfs_v2-ListCommitSetRequest) | [CommitSetInfo](#pfs_v2-CommitSetInfo) stream | ListCommitSet returns info about all CommitSets. |
| SquashCommitSet | [SquashCommitSetRequest](#pfs_v2-SquashCommitSetRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | SquashCommitSet squashes the commits of a CommitSet into their children. Deprecated: Use SquashCommit instead. |
| DropCommitSet | [DropCommitSetRequest](#pfs_v2-DropCommitSetRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | DropCommitSet drops the commits of a CommitSet and all data included in the commits. Deprecated: Use DropCommit instead. |
| FindCommits | [FindCommitsRequest](#pfs_v2-FindCommitsRequest) | [FindCommitsResponse](#pfs_v2-FindCommitsResponse) stream | FindCommits searches for commits that reference a supplied file being modified in a branch. |
| CreateBranch | [CreateBranchRequest](#pfs_v2-CreateBranchRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | CreateBranch creates a new branch. |
| InspectBranch | [InspectBranchRequest](#pfs_v2-InspectBranchRequest) | [BranchInfo](#pfs_v2-BranchInfo) | InspectBranch returns info about a branch. |
| ListBranch | [ListBranchRequest](#pfs_v2-ListBranchRequest) | [BranchInfo](#pfs_v2-BranchInfo) stream | ListBranch returns info about the heads of branches. |
| DeleteBranch | [DeleteBranchRequest](#pfs_v2-DeleteBranchRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | DeleteBranch deletes a branch; note that the commits still exist. |
| ModifyFile | [ModifyFileRequest](#pfs_v2-ModifyFileRequest) stream | [.google.protobuf.Empty](#google-protobuf-Empty) | ModifyFile performs modifications on a set of files. |
| GetFile | [GetFileRequest](#pfs_v2-GetFileRequest) | [.google.protobuf.BytesValue](#google-protobuf-BytesValue) stream | GetFile returns the contents of a single file |
| GetFileTAR | [GetFileRequest](#pfs_v2-GetFileRequest) | [.google.protobuf.BytesValue](#google-protobuf-BytesValue) stream | GetFileTAR returns a TAR stream of the contents matched by the request |
| InspectFile | [InspectFileRequest](#pfs_v2-InspectFileRequest) | [FileInfo](#pfs_v2-FileInfo) | InspectFile returns info about a file. |
| ListFile | [ListFileRequest](#pfs_v2-ListFileRequest) | [FileInfo](#pfs_v2-FileInfo) stream | ListFile returns info about all files. |
| WalkFile | [WalkFileRequest](#pfs_v2-WalkFileRequest) | [FileInfo](#pfs_v2-FileInfo) stream | WalkFile walks over all the files under a directory, including children of children. |
| GlobFile | [GlobFileRequest](#pfs_v2-GlobFileRequest) | [FileInfo](#pfs_v2-FileInfo) stream | GlobFile returns info about all files. |
| DiffFile | [DiffFileRequest](#pfs_v2-DiffFileRequest) | [DiffFileResponse](#pfs_v2-DiffFileResponse) stream | DiffFile returns the differences between 2 paths at 2 commits. |
| ActivateAuth | [ActivateAuthRequest](#pfs_v2-ActivateAuthRequest) | [ActivateAuthResponse](#pfs_v2-ActivateAuthResponse) | ActivateAuth creates a role binding for all existing repos |
| DeleteAll | [.google.protobuf.Empty](#google-protobuf-Empty) | [.google.protobuf.Empty](#google-protobuf-Empty) | DeleteAll deletes everything. |
| Fsck | [FsckRequest](#pfs_v2-FsckRequest) | [FsckResponse](#pfs_v2-FsckResponse) stream | Fsck does a file system consistency check for pfs. |
| CreateFileSet | [ModifyFileRequest](#pfs_v2-ModifyFileRequest) stream | [CreateFileSetResponse](#pfs_v2-CreateFileSetResponse) | FileSet API CreateFileSet creates a new file set. |
| GetFileSet | [GetFileSetRequest](#pfs_v2-GetFileSetRequest) | [CreateFileSetResponse](#pfs_v2-CreateFileSetResponse) | GetFileSet returns a file set with the data from a commit |
| AddFileSet | [AddFileSetRequest](#pfs_v2-AddFileSetRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | AddFileSet associates a file set with a commit |
| RenewFileSet | [RenewFileSetRequest](#pfs_v2-RenewFileSetRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | RenewFileSet prevents a file set from being deleted for a set amount of time. |
| ComposeFileSet | [ComposeFileSetRequest](#pfs_v2-ComposeFileSetRequest) | [CreateFileSetResponse](#pfs_v2-CreateFileSetResponse) | ComposeFileSet composes a file set from a list of file sets. |
| ShardFileSet | [ShardFileSetRequest](#pfs_v2-ShardFileSetRequest) | [ShardFileSetResponse](#pfs_v2-ShardFileSetResponse) |  |
| CheckStorage | [CheckStorageRequest](#pfs_v2-CheckStorageRequest) | [CheckStorageResponse](#pfs_v2-CheckStorageResponse) | CheckStorage runs integrity checks for the storage layer. |
| PutCache | [PutCacheRequest](#pfs_v2-PutCacheRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| GetCache | [GetCacheRequest](#pfs_v2-GetCacheRequest) | [GetCacheResponse](#pfs_v2-GetCacheResponse) |  |
| ClearCache | [ClearCacheRequest](#pfs_v2-ClearCacheRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| ListTask | [.taskapi.ListTaskRequest](#taskapi-ListTaskRequest) | [.taskapi.TaskInfo](#taskapi-TaskInfo) stream | ListTask lists PFS tasks |
| Egress | [EgressRequest](#pfs_v2-EgressRequest) | [EgressResponse](#pfs_v2-EgressResponse) | Egress writes data from a commit to an external system |
| CreateProject | [CreateProjectRequest](#pfs_v2-CreateProjectRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | Project API CreateProject creates a new project. |
| InspectProject | [InspectProjectRequest](#pfs_v2-InspectProjectRequest) | [ProjectInfo](#pfs_v2-ProjectInfo) | InspectProject returns info about a project. |
| InspectProjectV2 | [InspectProjectV2Request](#pfs_v2-InspectProjectV2Request) | [InspectProjectV2Response](#pfs_v2-InspectProjectV2Response) | InspectProjectV2 returns info about and defaults for a project. |
| ListProject | [ListProjectRequest](#pfs_v2-ListProjectRequest) | [ProjectInfo](#pfs_v2-ProjectInfo) stream | ListProject returns info about all projects. |
| DeleteProject | [DeleteProjectRequest](#pfs_v2-DeleteProjectRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | DeleteProject deletes a project. |

 



<a name="pjs_pjs-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## pjs/pjs.proto



<a name="pjs-CancelJobRequest"></a>

### CancelJobRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| context | [string](#string) |  | context is a bearer token used when calling from within a running Job. |
| job | [Job](#pjs-Job) |  |  |






<a name="pjs-CancelJobResponse"></a>

### CancelJobResponse







<a name="pjs-CreateJobRequest"></a>

### CreateJobRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| context | [string](#string) |  | context is a bearer token used when calling from within a running Job. |
| spec | [google.protobuf.Any](#google-protobuf-Any) |  |  |
| input | [QueueElement](#pjs-QueueElement) |  |  |
| cache_read | [bool](#bool) |  |  |
| cache_write | [bool](#bool) |  |  |






<a name="pjs-CreateJobResponse"></a>

### CreateJobResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [Job](#pjs-Job) |  |  |






<a name="pjs-DeleteJobRequest"></a>

### DeleteJobRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| context | [string](#string) |  | context is a bearer token used when calling from within a running Job. |
| job | [Job](#pjs-Job) |  |  |






<a name="pjs-DeleteJobResponse"></a>

### DeleteJobResponse







<a name="pjs-InspectJobRequest"></a>

### InspectJobRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| context | [string](#string) |  | context is a bearer token used when calling from within a running Job. |
| job | [Job](#pjs-Job) |  | job is the job to start walking from. If unset the context Job is assumed. |






<a name="pjs-InspectJobResponse"></a>

### InspectJobResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| details | [JobInfoDetails](#pjs-JobInfoDetails) |  |  |






<a name="pjs-InspectQueueRequest"></a>

### InspectQueueRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| queue | [Queue](#pjs-Queue) |  |  |






<a name="pjs-InspectQueueResponse"></a>

### InspectQueueResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| details | [QueueInfoDetails](#pjs-QueueInfoDetails) |  |  |






<a name="pjs-Job"></a>

### Job
Job uniquely identifies a Job
Job will be nil to indicate no Job, or an unset Job.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  |  |






<a name="pjs-JobInfo"></a>

### JobInfo
JobInfo describes a Job


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [Job](#pjs-Job) |  | Job is the Job&#39;s identity |
| parent_job | [Job](#pjs-Job) |  | parent_job is the Job&#39;s parent if it exists. |
| state | [JobState](#pjs-JobState) |  | state is the Job&#39;s state. See JobState for a description of the possible states. |
| spec | [google.protobuf.Any](#google-protobuf-Any) |  | spec is the code specification for the Job. |
| input | [QueueElement](#pjs-QueueElement) |  | input is the input data for the Job. |
| output | [QueueElement](#pjs-QueueElement) |  | output is produced by a successfully completing Job |
| error | [JobErrorCode](#pjs-JobErrorCode) |  | error is set when the Job is unable to complete successfully |






<a name="pjs-JobInfoDetails"></a>

### JobInfoDetails
JobInfoDetails is more detailed information about a Job.
It contains a superset of the information in JobInfo


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job_info | [JobInfo](#pjs-JobInfo) |  |  |






<a name="pjs-ListJobRequest"></a>

### ListJobRequest
TODO:
- Filter
- Paginate


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| context | [string](#string) |  | context is a bearer token used when calling from within a running Job. |
| job | [Job](#pjs-Job) |  | job is the job to start listing at. If nil, then the listing starts at the first job in the natural ordering. |






<a name="pjs-ListJobResponse"></a>

### ListJobResponse
ListJobResponse lists information about Jobs
ID will always be set.
Info and Details may not be set depending on how much information was requested.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [Job](#pjs-Job) |  |  |
| info | [JobInfo](#pjs-JobInfo) |  |  |
| details | [JobInfoDetails](#pjs-JobInfoDetails) |  |  |






<a name="pjs-ListQueueRequest"></a>

### ListQueueRequest
TODO:
- Filter
- Paginate






<a name="pjs-ListQueueResponse"></a>

### ListQueueResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [Queue](#pjs-Queue) |  |  |
| info | [QueueInfo](#pjs-QueueInfo) |  |  |
| details | [QueueInfoDetails](#pjs-QueueInfoDetails) |  |  |






<a name="pjs-ProcessQueueRequest"></a>

### ProcessQueueRequest
Queue Messages
ProcessQueueRequest is the client -&gt; server message for the bi-di ProcessQueue RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| queue | [Queue](#pjs-Queue) |  | queue is set to start processing from a Queue. |
| output | [QueueElement](#pjs-QueueElement) |  | output is set by the client to complete the Job successfully. |
| failed | [bool](#bool) |  | failed is set by the client to fail the Job. The Job will transition to state DONE with code FAILED. |






<a name="pjs-ProcessQueueResponse"></a>

### ProcessQueueResponse
ProcessQueueResposne is the server -&gt; client message for the bi-di ProcessQueue RPC.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| context | [string](#string) |  | context is a bearer token used to act on behalf of the Job in other RPCs. The server issues this token to the client, and the client should use it when performing Job RPCs. |
| input | [QueueElement](#pjs-QueueElement) |  | input is the input data for a Job. The server sends this to ask the client to compute the output. |






<a name="pjs-Queue"></a>

### Queue
Queue uniquely identifies a Queue
Queue will be nil to identify no Queue, or to indicate unset.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [bytes](#bytes) |  |  |






<a name="pjs-QueueElement"></a>

### QueueElement
QueueElement is a single element in a Queue.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [bytes](#bytes) |  | data is opaque data used as the input and output of Jobs |
| filesets | [string](#string) | repeated | filesets is a list of Fileset handles, used to associate Filesets with the input and output of Jobs. Any of the filesets referenced here will be persisted for as long as this element is in a Queue. New handles, pointing to equivalent Filesets, are minted whenever they cross the API boundary. |






<a name="pjs-QueueInfo"></a>

### QueueInfo
QueueInfo describes a Queue


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| queue | [Queue](#pjs-Queue) |  | queue is the Queue&#39;s identity |
| spec | [google.protobuf.Any](#google-protobuf-Any) |  | spec specifies the code to be run to process the Queue. |






<a name="pjs-QueueInfoDetails"></a>

### QueueInfoDetails
QueueInfoDetails contains detailed information about a Queue, which may be more expensive to get.
It contains a superset of the information in QueueInfo.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| queue_info | [QueueInfo](#pjs-QueueInfo) |  |  |
| size | [int64](#int64) |  | size is the number of elements queued. |






<a name="pjs-WalkJobRequest"></a>

### WalkJobRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| context | [string](#string) |  | context is a bearer token used when calling from within a running Job. |
| job | [Job](#pjs-Job) |  | job is the job to start walking from. If unset, the context Job is assumed. |





 


<a name="pjs-JobErrorCode"></a>

### JobErrorCode


| Name | Number | Description |
| ---- | ------ | ----------- |
| JobErrorCode_UNSPECIFIED | 0 | UNSPECIFIED means the job error code is unspecified. |
| FAILED | 1 | FAILED means that the worker processing the job indicated that it failed. |
| DISCONNECTED | 2 | DISCONNECTED means the worker processing the job disconnected. |
| CANCELED | 3 | CANCELED means the job was canceled. |



<a name="pjs-JobState"></a>

### JobState


| Name | Number | Description |
| ---- | ------ | ----------- |
| JobState_UNSPECIFIED | 0 | UNSPECIFIED means the job state is unspecified. |
| QUEUED | 1 | QUEUED means the job is currently in a queue. A QUEUED job will not have any descendants. |
| PROCESSING | 2 | PROCESSING means the job is currently being processed by a worker. |
| DONE | 3 | DONE means the job, and all of its descendants, are done. |


 

 


<a name="pjs-API"></a>

### API
Job API

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateJob | [CreateJobRequest](#pjs-CreateJobRequest) | [CreateJobResponse](#pjs-CreateJobResponse) | CreateJob creates a new job. Child jobs can be created by setting the context field to the appropriate parent job context. |
| CancelJob | [CancelJobRequest](#pjs-CancelJobRequest) | [CancelJobResponse](#pjs-CancelJobResponse) | CancelJob cancels a job. Canceling a job transitions all of the associated QUEUED and PROCESSING jobs to the DONE state and sets their error codes to CANCELED. This will terminate all ongoing processing associated with the job. Nothing will be deleted. A job can only be canceled with the parent job context. |
| DeleteJob | [DeleteJobRequest](#pjs-DeleteJobRequest) | [DeleteJobResponse](#pjs-DeleteJobResponse) | DeleteJob deletes a job. DeleteJob first cancels the job, then deletes all of the metadata and filesets associated with the job. A job can only be deleted with the parent job context. |
| ListJob | [ListJobRequest](#pjs-ListJobRequest) | [ListJobResponse](#pjs-ListJobResponse) stream | ListJob returns a list of jobs and information about each job. The jobs returned in the list are the child jobs of the provided job. If no job is provided, the list is the child jobs of the provided job context. The provided job must be associated with the provided job context or a descendant of the job associated with the provided job context. |
| WalkJob | [WalkJobRequest](#pjs-WalkJobRequest) | [ListJobResponse](#pjs-ListJobResponse) stream | WalkJob returns a list of jobs in a hierarchy and information about each job. Walking a job traverses the job hierarchy rooted at the provided job. The provided job must be associated with the provided job context or a descendant of the job associated with the provided job context. |
| InspectJob | [InspectJobRequest](#pjs-InspectJobRequest) | [InspectJobResponse](#pjs-InspectJobResponse) | InspectJob returns detailed information about a job. |
| ProcessQueue | [ProcessQueueRequest](#pjs-ProcessQueueRequest) stream | [ProcessQueueResponse](#pjs-ProcessQueueResponse) stream | ProcessQueue should be called by workers to process jobs in a queue. The protocol is as follows: Worker sends an initial request with the queue id. For each job: Server sends a response with a job context and the associated queue element. Worker processes the job. Worker sends a request with the job output or indicates that the job failed. This RPC should generally be run indefinitely. Workers will be scaled based on demand, so the expectation is that they should be processing queues while they are up. This RPC will be canceled by the server if the current job is canceled. Workers should generally retry the RPC when disconnects occur. |
| ListQueue | [ListQueueRequest](#pjs-ListQueueRequest) | [ListQueueResponse](#pjs-ListQueueResponse) stream | ListQueue returns a list of queues and information about each queue. |
| InspectQueue | [InspectQueueRequest](#pjs-InspectQueueRequest) | [InspectQueueResponse](#pjs-InspectQueueResponse) | InspectQueue returns detailed information about a queue. |

 



<a name="pps_pps-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## pps/pps.proto



<a name="pps_v2-ActivateAuthRequest"></a>

### ActivateAuthRequest







<a name="pps_v2-ActivateAuthResponse"></a>

### ActivateAuthResponse







<a name="pps_v2-Aggregate"></a>

### Aggregate



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| count | [int64](#int64) |  |  |
| mean | [double](#double) |  |  |
| stddev | [double](#double) |  |  |
| fifth_percentile | [double](#double) |  |  |
| ninety_fifth_percentile | [double](#double) |  |  |






<a name="pps_v2-AggregateProcessStats"></a>

### AggregateProcessStats



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| download_time | [Aggregate](#pps_v2-Aggregate) |  |  |
| process_time | [Aggregate](#pps_v2-Aggregate) |  |  |
| upload_time | [Aggregate](#pps_v2-Aggregate) |  |  |
| download_bytes | [Aggregate](#pps_v2-Aggregate) |  |  |
| upload_bytes | [Aggregate](#pps_v2-Aggregate) |  |  |






<a name="pps_v2-CheckStatusRequest"></a>

### CheckStatusRequest
Request to check the status of pipelines within a project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| all | [bool](#bool) |  | boolean field indicating status of all project pipelines. |
| project | [pfs_v2.Project](#pfs_v2-Project) |  | project field |






<a name="pps_v2-CheckStatusResponse"></a>

### CheckStatusResponse
Response for check status request. Provides alerts if any.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [pfs_v2.Project](#pfs_v2-Project) |  | project field |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  | pipeline field |
| alerts | [string](#string) | repeated | alert indicators |






<a name="pps_v2-ClusterDefaults"></a>

### ClusterDefaults



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| create_pipeline_request | [CreatePipelineRequest](#pps_v2-CreatePipelineRequest) |  | CreatePipelineRequest contains the default JSON CreatePipelineRequest into which pipeline specs are merged to form the effective spec used to create a pipeline. |






<a name="pps_v2-CreateDatumRequest"></a>

### CreateDatumRequest
Emits a stream of datums as they are created from the given input. Client
must cancel the stream when it no longer wants to receive datums.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| input | [Input](#pps_v2-Input) |  | Input is the input to list datums from. The datums listed are the ones that would be run if a pipeline was created with the provided input. The input field is only required for the first request. The server ignores subsequent requests&#39; input field. |
| number | [int64](#int64) |  | Number of datums to return in next response |






<a name="pps_v2-CreatePipelineRequest"></a>

### CreatePipelineRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |
| tf_job | [TFJob](#pps_v2-TFJob) |  | tf_job encodes a Kubeflow TFJob spec. Pachyderm uses this to create TFJobs when running in a kubernetes cluster on which kubeflow has been installed. Exactly one of &#39;tf_job&#39; and &#39;transform&#39; should be set |
| transform | [Transform](#pps_v2-Transform) |  |  |
| parallelism_spec | [ParallelismSpec](#pps_v2-ParallelismSpec) |  |  |
| egress | [Egress](#pps_v2-Egress) |  |  |
| update | [bool](#bool) |  |  |
| output_branch | [string](#string) |  |  |
| s3_out | [bool](#bool) |  | s3_out, if set, requires a pipeline&#39;s user to write to its output repo via Pachyderm&#39;s s3 gateway (if set, workers will serve Pachyderm&#39;s s3 gateway API at http://&lt;pipeline&gt;-s3.&lt;namespace&gt;/&lt;job id&gt;.out/my/file). In this mode /pfs_v2/out won&#39;t be walked or uploaded, and the s3 gateway service in the workers will allow writes to the job&#39;s output commit |
| resource_requests | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| resource_limits | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| sidecar_resource_limits | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| input | [Input](#pps_v2-Input) |  |  |
| description | [string](#string) |  |  |
| reprocess | [bool](#bool) |  | Reprocess forces the pipeline to reprocess all datums. It only has meaning if Update is true |
| service | [Service](#pps_v2-Service) |  |  |
| spout | [Spout](#pps_v2-Spout) |  |  |
| datum_set_spec | [DatumSetSpec](#pps_v2-DatumSetSpec) |  |  |
| datum_timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| job_timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| salt | [string](#string) |  |  |
| datum_tries | [int64](#int64) |  |  |
| scheduling_spec | [SchedulingSpec](#pps_v2-SchedulingSpec) |  |  |
| pod_spec | [string](#string) |  | deprecated, use pod_patch below |
| pod_patch | [string](#string) |  | a json patch will be applied to the pipeline&#39;s pod_spec before it&#39;s created; |
| spec_commit | [pfs_v2.Commit](#pfs_v2-Commit) |  |  |
| metadata | [Metadata](#pps_v2-Metadata) |  |  |
| reprocess_spec | [string](#string) |  |  |
| autoscaling | [bool](#bool) |  |  |
| tolerations | [Toleration](#pps_v2-Toleration) | repeated |  |
| sidecar_resource_requests | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| dry_run | [bool](#bool) |  |  |
| determined | [Determined](#pps_v2-Determined) |  |  |
| maximum_expected_uptime | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |






<a name="pps_v2-CreatePipelineTransaction"></a>

### CreatePipelineTransaction



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| create_pipeline_request | [CreatePipelineRequest](#pps_v2-CreatePipelineRequest) |  |  |
| user_json | [string](#string) |  | the JSON the user originally submitted |
| effective_json | [string](#string) |  | the effective spec: the result of merging the user JSON into the cluster defaults |






<a name="pps_v2-CreatePipelineV2Request"></a>

### CreatePipelineV2Request



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| create_pipeline_request_json | [string](#string) |  | a JSON-encoded CreatePipelineRequest |
| dry_run | [bool](#bool) |  |  |
| update | [bool](#bool) |  |  |
| reprocess | [bool](#bool) |  |  |






<a name="pps_v2-CreatePipelineV2Response"></a>

### CreatePipelineV2Response



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| effective_create_pipeline_request_json | [string](#string) |  |  |






<a name="pps_v2-CreateSecretRequest"></a>

### CreateSecretRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file | [bytes](#bytes) |  |  |






<a name="pps_v2-CronInput"></a>

### CronInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| project | [string](#string) |  |  |
| repo | [string](#string) |  |  |
| commit | [string](#string) |  |  |
| spec | [string](#string) |  |  |
| overwrite | [bool](#bool) |  | Overwrite, if true, will expose a single datum that gets overwritten each tick. If false, it will create a new datum for each tick. |
| start | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="pps_v2-Datum"></a>

### Datum



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [Job](#pps_v2-Job) |  | ID is the hash computed from all the files |
| id | [string](#string) |  |  |






<a name="pps_v2-DatumInfo"></a>

### DatumInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| datum | [Datum](#pps_v2-Datum) |  |  |
| state | [DatumState](#pps_v2-DatumState) |  |  |
| stats | [ProcessStats](#pps_v2-ProcessStats) |  |  |
| pfs_state | [pfs_v2.File](#pfs_v2-File) |  |  |
| data | [pfs_v2.FileInfo](#pfs_v2-FileInfo) | repeated |  |
| image_id | [string](#string) |  |  |






<a name="pps_v2-DatumSetSpec"></a>

### DatumSetSpec
DatumSetSpec specifies how a pipeline should split its datums into datum sets.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | [int64](#int64) |  | number, if nonzero, specifies that each datum set should contain `number` datums. Datum sets may contain fewer if the total number of datums don&#39;t divide evenly. |
| size_bytes | [int64](#int64) |  | size_bytes, if nonzero, specifies a target size for each datum set. Datum sets may be larger or smaller than size_bytes, but will usually be pretty close to size_bytes in size. |
| per_worker | [int64](#int64) |  | per_worker, if nonzero, specifies how many datum sets should be created for each worker. It can&#39;t be set with number or size_bytes. |






<a name="pps_v2-DatumStatus"></a>

### DatumStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| started | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | Started is the time processing on the current datum began. |
| data | [InputFile](#pps_v2-InputFile) | repeated |  |






<a name="pps_v2-DeleteJobRequest"></a>

### DeleteJobRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [Job](#pps_v2-Job) |  |  |






<a name="pps_v2-DeletePipelineRequest"></a>

### DeletePipelineRequest
Delete a pipeline.  If the deprecated all member is true, then delete all
pipelines in the default project.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |
| all | [bool](#bool) |  | **Deprecated.** Deprecated. |
| force | [bool](#bool) |  |  |
| keep_repo | [bool](#bool) |  |  |
| must_exist | [bool](#bool) |  | If true, an error will be returned if the pipeline doesn&#39;t exist. |






<a name="pps_v2-DeletePipelinesRequest"></a>

### DeletePipelinesRequest
Delete more than one pipeline.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| projects | [pfs_v2.Project](#pfs_v2-Project) | repeated | All pipelines in each project will be deleted if the caller has permission. |
| force | [bool](#bool) |  |  |
| keep_repo | [bool](#bool) |  |  |
| all | [bool](#bool) |  | If set, all pipelines in all projects will be deleted if the caller has permission. |






<a name="pps_v2-DeletePipelinesResponse"></a>

### DeletePipelinesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipelines | [Pipeline](#pps_v2-Pipeline) | repeated |  |






<a name="pps_v2-DeleteSecretRequest"></a>

### DeleteSecretRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secret | [Secret](#pps_v2-Secret) |  |  |






<a name="pps_v2-Determined"></a>

### Determined



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workspaces | [string](#string) | repeated |  |






<a name="pps_v2-Egress"></a>

### Egress



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| URL | [string](#string) |  |  |
| object_storage | [pfs_v2.ObjectStorageEgress](#pfs_v2-ObjectStorageEgress) |  |  |
| sql_database | [pfs_v2.SQLDatabaseEgress](#pfs_v2-SQLDatabaseEgress) |  |  |






<a name="pps_v2-GPUSpec"></a>

### GPUSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  | The type of GPU (nvidia.com/gpu or amd.com/gpu for example). |
| number | [int64](#int64) |  | The number of GPUs to request. |






<a name="pps_v2-GetClusterDefaultsRequest"></a>

### GetClusterDefaultsRequest







<a name="pps_v2-GetClusterDefaultsResponse"></a>

### GetClusterDefaultsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_defaults_json | [string](#string) |  | A JSON-encoded ClusterDefaults message, this is the verbatim input passed to SetClusterDefaults. |






<a name="pps_v2-GetLogsRequest"></a>

### GetLogsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  | The pipeline from which we want to get logs (required if the job in &#39;job&#39; was created as part of a pipeline. To get logs from a non-orphan job without the pipeline that created it, you need to use ElasticSearch). |
| job | [Job](#pps_v2-Job) |  | The job from which we want to get logs. |
| data_filters | [string](#string) | repeated | Names of input files from which we want processing logs. This may contain multiple files, to query pipelines that contain multiple inputs. Each filter may be an absolute path of a file within a pps repo, or it may be a hash for that file (to search for files at specific versions) |
| datum | [Datum](#pps_v2-Datum) |  |  |
| master | [bool](#bool) |  | If true get logs from the master process |
| follow | [bool](#bool) |  | Continue to follow new logs as they become available. |
| tail | [int64](#int64) |  | If nonzero, the number of lines from the end of the logs to return. Note: tail applies per container, so you will get tail * &lt;number of pods&gt; total lines back. |
| use_loki_backend | [bool](#bool) |  | UseLokiBackend causes the logs request to go through the loki backend rather than through kubernetes. This behavior can also be achieved by setting the LOKI_LOGGING feature flag. |
| since | [google.protobuf.Duration](#google-protobuf-Duration) |  | Since specifies how far in the past to return logs from. It defaults to 24 hours. |






<a name="pps_v2-GetProjectDefaultsRequest"></a>

### GetProjectDefaultsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [pfs_v2.Project](#pfs_v2-Project) |  |  |






<a name="pps_v2-GetProjectDefaultsResponse"></a>

### GetProjectDefaultsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_defaults_json | [string](#string) |  | A JSON-encoded ProjectDefaults message, this is the verbatim input passed to SetProjectDefaults. |






<a name="pps_v2-Input"></a>

### Input



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pfs | [PFSInput](#pps_v2-PFSInput) |  |  |
| join | [Input](#pps_v2-Input) | repeated |  |
| group | [Input](#pps_v2-Input) | repeated |  |
| cross | [Input](#pps_v2-Input) | repeated |  |
| union | [Input](#pps_v2-Input) | repeated |  |
| cron | [CronInput](#pps_v2-CronInput) |  |  |






<a name="pps_v2-InputFile"></a>

### InputFile



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| path | [string](#string) |  | This file&#39;s absolute path within its pfs repo. |
| hash | [bytes](#bytes) |  | This file&#39;s hash |






<a name="pps_v2-InspectDatumRequest"></a>

### InspectDatumRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| datum | [Datum](#pps_v2-Datum) |  |  |






<a name="pps_v2-InspectJobRequest"></a>

### InspectJobRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [Job](#pps_v2-Job) |  | Callers should set either Job or OutputCommit, not both. |
| wait | [bool](#bool) |  | wait until state is either FAILURE or SUCCESS |
| details | [bool](#bool) |  |  |






<a name="pps_v2-InspectJobSetRequest"></a>

### InspectJobSetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job_set | [JobSet](#pps_v2-JobSet) |  |  |
| wait | [bool](#bool) |  | When true, wait until all jobs in the set are finished |
| details | [bool](#bool) |  |  |






<a name="pps_v2-InspectPipelineRequest"></a>

### InspectPipelineRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |
| details | [bool](#bool) |  | When true, return PipelineInfos with the details field, which requires loading the pipeline spec from PFS. |






<a name="pps_v2-InspectSecretRequest"></a>

### InspectSecretRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secret | [Secret](#pps_v2-Secret) |  |  |






<a name="pps_v2-Job"></a>

### Job



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |
| id | [string](#string) |  |  |






<a name="pps_v2-JobInfo"></a>

### JobInfo
JobInfo is the data stored in the database regarding a given job.  The
&#39;details&#39; field contains more information about the job which is expensive to
fetch, requiring querying workers or loading the pipeline spec from object
storage.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [Job](#pps_v2-Job) |  |  |
| pipeline_version | [uint64](#uint64) |  |  |
| output_commit | [pfs_v2.Commit](#pfs_v2-Commit) |  |  |
| restart | [uint64](#uint64) |  | Job restart count (e.g. due to datum failure) |
| data_processed | [int64](#int64) |  | Counts of how many times we processed or skipped a datum |
| data_skipped | [int64](#int64) |  |  |
| data_total | [int64](#int64) |  |  |
| data_failed | [int64](#int64) |  |  |
| data_recovered | [int64](#int64) |  |  |
| stats | [ProcessStats](#pps_v2-ProcessStats) |  | Download/process/upload time and download/upload bytes |
| state | [JobState](#pps_v2-JobState) |  |  |
| reason | [string](#string) |  | reason explains why the job is in the current state |
| created | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| started | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| finished | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| details | [JobInfo.Details](#pps_v2-JobInfo-Details) |  |  |
| auth_token | [string](#string) |  |  |






<a name="pps_v2-JobInfo-Details"></a>

### JobInfo.Details



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| transform | [Transform](#pps_v2-Transform) |  |  |
| parallelism_spec | [ParallelismSpec](#pps_v2-ParallelismSpec) |  |  |
| egress | [Egress](#pps_v2-Egress) |  |  |
| service | [Service](#pps_v2-Service) |  |  |
| spout | [Spout](#pps_v2-Spout) |  |  |
| worker_status | [WorkerStatus](#pps_v2-WorkerStatus) | repeated |  |
| resource_requests | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| resource_limits | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| sidecar_resource_limits | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| input | [Input](#pps_v2-Input) |  |  |
| salt | [string](#string) |  |  |
| datum_set_spec | [DatumSetSpec](#pps_v2-DatumSetSpec) |  |  |
| datum_timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| job_timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| datum_tries | [int64](#int64) |  |  |
| scheduling_spec | [SchedulingSpec](#pps_v2-SchedulingSpec) |  |  |
| pod_spec | [string](#string) |  |  |
| pod_patch | [string](#string) |  |  |
| sidecar_resource_requests | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |






<a name="pps_v2-JobInput"></a>

### JobInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| commit | [pfs_v2.Commit](#pfs_v2-Commit) |  |  |
| glob | [string](#string) |  |  |
| lazy | [bool](#bool) |  |  |






<a name="pps_v2-JobSet"></a>

### JobSet



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="pps_v2-JobSetInfo"></a>

### JobSetInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job_set | [JobSet](#pps_v2-JobSet) |  |  |
| jobs | [JobInfo](#pps_v2-JobInfo) | repeated |  |






<a name="pps_v2-ListDatumRequest"></a>

### ListDatumRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [Job](#pps_v2-Job) |  | Job and Input are two different ways to specify the datums you want. Only one can be set. Job is the job to list datums from. |
| input | [Input](#pps_v2-Input) |  | Input is the input to list datums from. The datums listed are the ones that would be run if a pipeline was created with the provided input. |
| filter | [ListDatumRequest.Filter](#pps_v2-ListDatumRequest-Filter) |  |  |
| paginationMarker | [string](#string) |  | datum id to start from. we do not include this datum in the response |
| number | [int64](#int64) |  | Number of datums to return |
| reverse | [bool](#bool) |  | If true, return datums in reverse order |






<a name="pps_v2-ListDatumRequest-Filter"></a>

### ListDatumRequest.Filter
Filter restricts returned DatumInfo messages to those which match
all of the filtered attributes.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state | [DatumState](#pps_v2-DatumState) | repeated | Must match one of the given states. |






<a name="pps_v2-ListJobRequest"></a>

### ListJobRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| projects | [pfs_v2.Project](#pfs_v2-Project) | repeated | A list of projects to filter jobs on, nil means don&#39;t filter. |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  | nil means all pipelines |
| input_commit | [pfs_v2.Commit](#pfs_v2-Commit) | repeated | nil means all inputs |
| history | [int64](#int64) |  | History indicates return jobs from historical versions of pipelines semantics are: 0: Return jobs from the current version of the pipeline or pipelines. 1: Return the above and jobs from the next most recent version 2: etc. -1: Return jobs from all historical versions. |
| details | [bool](#bool) |  | Details indicates whether the result should include all pipeline details in each JobInfo, or limited information including name and status, but excluding information in the pipeline spec. Leaving this &#34;false&#34; can make the call significantly faster in clusters with a large number of pipelines and jobs. Note that if &#39;input_commit&#39; is set, this field is coerced to &#34;true&#34; |
| jqFilter | [string](#string) |  | A jq program string for additional result filtering |
| paginationMarker | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | timestamp that is pagination marker |
| number | [int64](#int64) |  | number of results to return |
| reverse | [bool](#bool) |  | flag to indicated if results should be returned in reverse order |






<a name="pps_v2-ListJobSetRequest"></a>

### ListJobSetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| details | [bool](#bool) |  |  |
| projects | [pfs_v2.Project](#pfs_v2-Project) | repeated | A list of projects to filter jobs on, nil means don&#39;t filter. |
| paginationMarker | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | we return job sets created before or after this time based on the reverse flag |
| number | [int64](#int64) |  | number of results to return |
| reverse | [bool](#bool) |  | if true, return results in reverse order |
| jqFilter | [string](#string) |  | A jq program string for additional result filtering |






<a name="pps_v2-ListPipelineRequest"></a>

### ListPipelineRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  | If non-nil, only return info about a single pipeline, this is redundant with InspectPipeline unless history is non-zero. |
| history | [int64](#int64) |  | History indicates how many historical versions you want returned. Its semantics are: 0: Return the current version of the pipeline or pipelines. 1: Return the above and the next most recent version 2: etc. -1: Return all historical versions. |
| details | [bool](#bool) |  | **Deprecated.** Deprecated: Details are always returned. |
| jqFilter | [string](#string) |  | A jq program string for additional result filtering |
| commit_set | [pfs_v2.CommitSet](#pfs_v2-CommitSet) |  | If non-nil, will return all the pipeline infos at this commit set |
| projects | [pfs_v2.Project](#pfs_v2-Project) | repeated | Projects to filter on. Empty list means no filter, so return all pipelines. |






<a name="pps_v2-LogMessage"></a>

### LogMessage
LogMessage is a log line from a PPS worker, annotated with metadata
indicating when and why the line was logged.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_name | [string](#string) |  | The job and pipeline for which a PFS file is being processed (if the job is an orphan job, pipeline name and ID will be unset) |
| pipeline_name | [string](#string) |  |  |
| job_id | [string](#string) |  |  |
| worker_id | [string](#string) |  |  |
| datum_id | [string](#string) |  |  |
| master | [bool](#bool) |  |  |
| data | [InputFile](#pps_v2-InputFile) | repeated | The PFS files being processed (one per pipeline/job input) |
| user | [bool](#bool) |  | User is true if log message comes from the users code. |
| ts | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The message logged, and the time at which it was logged |
| message | [string](#string) |  |  |






<a name="pps_v2-LokiLogMessage"></a>

### LokiLogMessage



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [string](#string) |  |  |






<a name="pps_v2-LokiRequest"></a>

### LokiRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| since | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| query | [string](#string) |  |  |






<a name="pps_v2-Metadata"></a>

### Metadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| annotations | [Metadata.AnnotationsEntry](#pps_v2-Metadata-AnnotationsEntry) | repeated |  |
| labels | [Metadata.LabelsEntry](#pps_v2-Metadata-LabelsEntry) | repeated |  |






<a name="pps_v2-Metadata-AnnotationsEntry"></a>

### Metadata.AnnotationsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="pps_v2-Metadata-LabelsEntry"></a>

### Metadata.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="pps_v2-PFSInput"></a>

### PFSInput



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [string](#string) |  |  |
| name | [string](#string) |  |  |
| repo | [string](#string) |  |  |
| repo_type | [string](#string) |  |  |
| branch | [string](#string) |  |  |
| commit | [string](#string) |  |  |
| glob | [string](#string) |  |  |
| join_on | [string](#string) |  |  |
| outer_join | [bool](#bool) |  |  |
| group_by | [string](#string) |  |  |
| lazy | [bool](#bool) |  |  |
| empty_files | [bool](#bool) |  | EmptyFiles, if true, will cause files from this PFS input to be presented as empty files. This is useful in shuffle pipelines where you want to read the names of files and reorganize them using symlinks. |
| s3 | [bool](#bool) |  | S3, if true, will cause the worker to NOT download or link files from this input into the /pfs_v2 directory. Instead, an instance of our S3 gateway service will run on each of the sidecars, and data can be retrieved from this input by querying http://&lt;pipeline&gt;-s3.&lt;namespace&gt;/&lt;job id&gt;.&lt;input&gt;/my/file |
| trigger | [pfs_v2.Trigger](#pfs_v2-Trigger) |  | Trigger defines when this input is processed by the pipeline, if it&#39;s nil the input is processed anytime something is committed to the input branch. |






<a name="pps_v2-ParallelismSpec"></a>

### ParallelismSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| constant | [uint64](#uint64) |  | Starts the pipeline/job with a &#39;constant&#39; workers, unless &#39;constant&#39; is zero. If &#39;constant&#39; is zero (which is the zero value of ParallelismSpec), then Pachyderm will choose the number of workers that is started, (currently it chooses the number of workers in the cluster) |






<a name="pps_v2-Pipeline"></a>

### Pipeline



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [pfs_v2.Project](#pfs_v2-Project) |  |  |
| name | [string](#string) |  |  |






<a name="pps_v2-PipelineInfo"></a>

### PipelineInfo
PipelineInfo is proto for each pipeline that Pachd stores in the
database. It tracks the state of the pipeline, and points to its metadata in
PFS (and, by pointing to a PFS commit, de facto tracks the pipeline&#39;s
version).  Any information about the pipeline _not_ stored in the database is
in the Details object, which requires fetching the spec from PFS or other
potentially expensive operations.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |
| version | [uint64](#uint64) |  |  |
| spec_commit | [pfs_v2.Commit](#pfs_v2-Commit) |  | The first spec commit for this version of the pipeline |
| stopped | [bool](#bool) |  |  |
| state | [PipelineState](#pps_v2-PipelineState) |  | state indicates the current state of the pipeline |
| reason | [string](#string) |  | reason includes any error messages associated with a failed pipeline |
| last_job_state | [JobState](#pps_v2-JobState) |  | last_job_state indicates the state of the most recently created job |
| parallelism | [uint64](#uint64) |  | parallelism tracks the literal number of workers that this pipeline should run. |
| type | [PipelineInfo.PipelineType](#pps_v2-PipelineInfo-PipelineType) |  |  |
| auth_token | [string](#string) |  |  |
| details | [PipelineInfo.Details](#pps_v2-PipelineInfo-Details) |  |  |
| user_spec_json | [string](#string) |  | The user-submitted pipeline spec in JSON format. |
| effective_spec_json | [string](#string) |  | The effective spec used to create the pipeline. Created by merging the user spec into the cluster defaults. |






<a name="pps_v2-PipelineInfo-Details"></a>

### PipelineInfo.Details



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| transform | [Transform](#pps_v2-Transform) |  |  |
| tf_job | [TFJob](#pps_v2-TFJob) |  | tf_job encodes a Kubeflow TFJob spec. Pachyderm uses this to create TFJobs when running in a kubernetes cluster on which kubeflow has been installed. Exactly one of &#39;tf_job&#39; and &#39;transform&#39; should be set |
| parallelism_spec | [ParallelismSpec](#pps_v2-ParallelismSpec) |  |  |
| egress | [Egress](#pps_v2-Egress) |  |  |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| recent_error | [string](#string) |  |  |
| workers_requested | [int64](#int64) |  |  |
| workers_available | [int64](#int64) |  |  |
| output_branch | [string](#string) |  |  |
| resource_requests | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| resource_limits | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| sidecar_resource_limits | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| input | [Input](#pps_v2-Input) |  |  |
| description | [string](#string) |  |  |
| salt | [string](#string) |  |  |
| reason | [string](#string) |  |  |
| service | [Service](#pps_v2-Service) |  |  |
| spout | [Spout](#pps_v2-Spout) |  |  |
| datum_set_spec | [DatumSetSpec](#pps_v2-DatumSetSpec) |  |  |
| datum_timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| job_timeout | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| datum_tries | [int64](#int64) |  |  |
| scheduling_spec | [SchedulingSpec](#pps_v2-SchedulingSpec) |  |  |
| pod_spec | [string](#string) |  |  |
| pod_patch | [string](#string) |  |  |
| s3_out | [bool](#bool) |  |  |
| metadata | [Metadata](#pps_v2-Metadata) |  |  |
| reprocess_spec | [string](#string) |  |  |
| unclaimed_tasks | [int64](#int64) |  |  |
| worker_rc | [string](#string) |  |  |
| autoscaling | [bool](#bool) |  |  |
| tolerations | [Toleration](#pps_v2-Toleration) | repeated |  |
| sidecar_resource_requests | [ResourceSpec](#pps_v2-ResourceSpec) |  |  |
| determined | [Determined](#pps_v2-Determined) |  |  |
| maximum_expected_uptime | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| workers_started_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="pps_v2-PipelineInfos"></a>

### PipelineInfos



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline_info | [PipelineInfo](#pps_v2-PipelineInfo) | repeated |  |






<a name="pps_v2-ProcessStats"></a>

### ProcessStats



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| download_time | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| process_time | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| upload_time | [google.protobuf.Duration](#google-protobuf-Duration) |  |  |
| download_bytes | [int64](#int64) |  |  |
| upload_bytes | [int64](#int64) |  |  |






<a name="pps_v2-ProjectDefaults"></a>

### ProjectDefaults



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| create_pipeline_request | [CreatePipelineRequest](#pps_v2-CreatePipelineRequest) |  |  |






<a name="pps_v2-RenderTemplateRequest"></a>

### RenderTemplateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| template | [string](#string) |  |  |
| args | [RenderTemplateRequest.ArgsEntry](#pps_v2-RenderTemplateRequest-ArgsEntry) | repeated |  |






<a name="pps_v2-RenderTemplateRequest-ArgsEntry"></a>

### RenderTemplateRequest.ArgsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="pps_v2-RenderTemplateResponse"></a>

### RenderTemplateResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| json | [string](#string) |  |  |
| specs | [CreatePipelineRequest](#pps_v2-CreatePipelineRequest) | repeated |  |






<a name="pps_v2-RerunPipelineRequest"></a>

### RerunPipelineRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |
| reprocess | [bool](#bool) |  | Reprocess forces the pipeline to reprocess all datums. |






<a name="pps_v2-ResourceSpec"></a>

### ResourceSpec
ResourceSpec describes the amount of resources that pipeline pods should
request from kubernetes, for scheduling.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cpu | [float](#float) |  | The number of CPUs each worker needs (partial values are allowed, and encouraged) |
| memory | [string](#string) |  | The amount of memory each worker needs (in bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc). |
| gpu | [GPUSpec](#pps_v2-GPUSpec) |  | The spec for GPU resources. |
| disk | [string](#string) |  | The amount of ephemeral storage each worker needs (in bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc). |






<a name="pps_v2-RestartDatumRequest"></a>

### RestartDatumRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [Job](#pps_v2-Job) |  |  |
| data_filters | [string](#string) | repeated |  |






<a name="pps_v2-RunCronRequest"></a>

### RunCronRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |






<a name="pps_v2-RunLoadTestRequest"></a>

### RunLoadTestRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dag_spec | [string](#string) |  |  |
| load_spec | [string](#string) |  |  |
| seed | [int64](#int64) |  |  |
| parallelism | [int64](#int64) |  |  |
| pod_patch | [string](#string) |  |  |
| state_id | [string](#string) |  |  |






<a name="pps_v2-RunLoadTestResponse"></a>

### RunLoadTestResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| error | [string](#string) |  |  |
| state_id | [string](#string) |  |  |






<a name="pps_v2-RunPipelineRequest"></a>

### RunPipelineRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |
| provenance | [pfs_v2.Commit](#pfs_v2-Commit) | repeated |  |
| job_id | [string](#string) |  |  |






<a name="pps_v2-SchedulingSpec"></a>

### SchedulingSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| node_selector | [SchedulingSpec.NodeSelectorEntry](#pps_v2-SchedulingSpec-NodeSelectorEntry) | repeated |  |
| priority_class_name | [string](#string) |  |  |






<a name="pps_v2-SchedulingSpec-NodeSelectorEntry"></a>

### SchedulingSpec.NodeSelectorEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="pps_v2-Secret"></a>

### Secret



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name="pps_v2-SecretInfo"></a>

### SecretInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secret | [Secret](#pps_v2-Secret) |  |  |
| type | [string](#string) |  |  |
| creation_timestamp | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |






<a name="pps_v2-SecretInfos"></a>

### SecretInfos



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| secret_info | [SecretInfo](#pps_v2-SecretInfo) | repeated |  |






<a name="pps_v2-SecretMount"></a>

### SecretMount



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name must be the name of the secret in kubernetes. |
| key | [string](#string) |  | Key of the secret to load into env_var, this field only has meaning if EnvVar != &#34;&#34;. |
| mount_path | [string](#string) |  |  |
| env_var | [string](#string) |  |  |






<a name="pps_v2-Service"></a>

### Service



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| internal_port | [int32](#int32) |  |  |
| external_port | [int32](#int32) |  |  |
| ip | [string](#string) |  |  |
| type | [string](#string) |  |  |






<a name="pps_v2-SetClusterDefaultsRequest"></a>

### SetClusterDefaultsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| regenerate | [bool](#bool) |  |  |
| reprocess | [bool](#bool) |  | must be false if regenerate is false |
| dry_run | [bool](#bool) |  |  |
| cluster_defaults_json | [string](#string) |  | A JSON-encoded ClusterDefaults message, this will be stored verbatim. |






<a name="pps_v2-SetClusterDefaultsResponse"></a>

### SetClusterDefaultsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| affected_pipelines | [Pipeline](#pps_v2-Pipeline) | repeated |  |






<a name="pps_v2-SetProjectDefaultsRequest"></a>

### SetProjectDefaultsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project | [pfs_v2.Project](#pfs_v2-Project) |  |  |
| regenerate | [bool](#bool) |  |  |
| reprocess | [bool](#bool) |  | must be false if regenerate is false |
| dry_run | [bool](#bool) |  |  |
| project_defaults_json | [string](#string) |  | A JSON-encoded ProjectDefaults message, this will be stored verbatim. |






<a name="pps_v2-SetProjectDefaultsResponse"></a>

### SetProjectDefaultsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| affected_pipelines | [Pipeline](#pps_v2-Pipeline) | repeated |  |






<a name="pps_v2-Spout"></a>

### Spout



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service | [Service](#pps_v2-Service) |  |  |






<a name="pps_v2-StartPipelineRequest"></a>

### StartPipelineRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |






<a name="pps_v2-StopJobRequest"></a>

### StopJobRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [Job](#pps_v2-Job) |  |  |
| reason | [string](#string) |  |  |






<a name="pps_v2-StopPipelineRequest"></a>

### StopPipelineRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |
| must_exist | [bool](#bool) |  | If true, an error will be returned if the pipeline doesn&#39;t exist. |






<a name="pps_v2-SubscribeJobRequest"></a>

### SubscribeJobRequest
Streams open jobs until canceled


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pipeline | [Pipeline](#pps_v2-Pipeline) |  |  |
| details | [bool](#bool) |  | Same as ListJobRequest.Details |






<a name="pps_v2-TFJob"></a>

### TFJob



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tf_job | [string](#string) |  | tf_job is a serialized Kubeflow TFJob spec. Pachyderm sends this directly to a kubernetes cluster on which kubeflow has been installed, instead of creating a pipeline ReplicationController as it normally would. |






<a name="pps_v2-Toleration"></a>

### Toleration
Toleration is a Kubernetes toleration.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  | key is the taint key that the toleration applies to. Empty means match all taint keys. |
| operator | [TolerationOperator](#pps_v2-TolerationOperator) |  | operator represents a key&#39;s relationship to the value. |
| value | [string](#string) |  | value is the taint value the toleration matches to. |
| effect | [TaintEffect](#pps_v2-TaintEffect) |  | effect indicates the taint effect to match. Empty means match all taint effects. |
| toleration_seconds | [google.protobuf.Int64Value](#google-protobuf-Int64Value) |  | toleration_seconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. If not set, tolerate the taint forever. |






<a name="pps_v2-Transform"></a>

### Transform



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| image | [string](#string) |  |  |
| cmd | [string](#string) | repeated |  |
| err_cmd | [string](#string) | repeated |  |
| env | [Transform.EnvEntry](#pps_v2-Transform-EnvEntry) | repeated |  |
| secrets | [SecretMount](#pps_v2-SecretMount) | repeated |  |
| image_pull_secrets | [string](#string) | repeated |  |
| stdin | [string](#string) | repeated |  |
| err_stdin | [string](#string) | repeated |  |
| accept_return_code | [int64](#int64) | repeated |  |
| debug | [bool](#bool) |  |  |
| user | [string](#string) |  |  |
| working_dir | [string](#string) |  |  |
| dockerfile | [string](#string) |  |  |
| memory_volume | [bool](#bool) |  |  |
| datum_batching | [bool](#bool) |  |  |






<a name="pps_v2-Transform-EnvEntry"></a>

### Transform.EnvEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="pps_v2-UpdateJobStateRequest"></a>

### UpdateJobStateRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [Job](#pps_v2-Job) |  |  |
| state | [JobState](#pps_v2-JobState) |  |  |
| reason | [string](#string) |  |  |
| restart | [uint64](#uint64) |  |  |
| data_processed | [int64](#int64) |  |  |
| data_skipped | [int64](#int64) |  |  |
| data_failed | [int64](#int64) |  |  |
| data_recovered | [int64](#int64) |  |  |
| data_total | [int64](#int64) |  |  |
| stats | [ProcessStats](#pps_v2-ProcessStats) |  |  |






<a name="pps_v2-Worker"></a>

### Worker



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| state | [WorkerState](#pps_v2-WorkerState) |  |  |






<a name="pps_v2-WorkerStatus"></a>

### WorkerStatus



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| worker_id | [string](#string) |  |  |
| job_id | [string](#string) |  |  |
| datum_status | [DatumStatus](#pps_v2-DatumStatus) |  |  |





 


<a name="pps_v2-DatumState"></a>

### DatumState


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 | or not part of a job |
| FAILED | 1 |  |
| SUCCESS | 2 |  |
| SKIPPED | 3 |  |
| STARTING | 4 |  |
| RECOVERED | 5 |  |



<a name="pps_v2-JobState"></a>

### JobState


| Name | Number | Description |
| ---- | ------ | ----------- |
| JOB_STATE_UNKNOWN | 0 |  |
| JOB_CREATED | 1 |  |
| JOB_STARTING | 2 |  |
| JOB_RUNNING | 3 |  |
| JOB_FAILURE | 4 |  |
| JOB_SUCCESS | 5 |  |
| JOB_KILLED | 6 |  |
| JOB_EGRESSING | 7 |  |
| JOB_FINISHING | 8 |  |
| JOB_UNRUNNABLE | 9 |  |



<a name="pps_v2-PipelineInfo-PipelineType"></a>

### PipelineInfo.PipelineType
The pipeline type is stored here so that we can internally know the type of
the pipeline without loading the spec from PFS.

| Name | Number | Description |
| ---- | ------ | ----------- |
| PIPELINT_TYPE_UNKNOWN | 0 |  |
| PIPELINE_TYPE_TRANSFORM | 1 |  |
| PIPELINE_TYPE_SPOUT | 2 |  |
| PIPELINE_TYPE_SERVICE | 3 |  |



<a name="pps_v2-PipelineState"></a>

### PipelineState


| Name | Number | Description |
| ---- | ------ | ----------- |
| PIPELINE_STATE_UNKNOWN | 0 |  |
| PIPELINE_STARTING | 1 | There is a PipelineInfo &#43; spec commit, but no RC This happens when a pipeline has been created but not yet picked up by a PPS server. |
| PIPELINE_RUNNING | 2 | A pipeline has a spec commit and a service &#43; RC This is the normal state of a pipeline. |
| PIPELINE_RESTARTING | 3 | Equivalent to STARTING (there is a PipelineInfo &#43; commit, but no RC) After some error caused runPipeline to exit, but before the pipeline is re-run. This is when the exponential backoff is in effect. |
| PIPELINE_FAILURE | 4 | The pipeline has encountered unrecoverable errors and is no longer being retried. It won&#39;t leave this state until the pipeline is updated. |
| PIPELINE_PAUSED | 5 | The pipeline has been explicitly paused by the user (the pipeline spec&#39;s Stopped field should be true if the pipeline is in this state) |
| PIPELINE_STANDBY | 6 | The pipeline is fully functional, but there are no commits to process. |
| PIPELINE_CRASHING | 7 | The pipeline&#39;s workers are crashing, or failing to come up, this may resolve itself, the pipeline may make progress while in this state if the problem is only being experienced by some workers. |



<a name="pps_v2-TaintEffect"></a>

### TaintEffect
TaintEffect is an effect that can be matched by a toleration.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ALL_EFFECTS | 0 | Empty matches all effects. |
| NO_SCHEDULE | 1 | &#34;NoSchedule&#34; |
| PREFER_NO_SCHEDULE | 2 | &#34;PreferNoSchedule&#34; |
| NO_EXECUTE | 3 | &#34;NoExecute&#34; |



<a name="pps_v2-TolerationOperator"></a>

### TolerationOperator
TolerationOperator relates a Toleration&#39;s key to its value.

| Name | Number | Description |
| ---- | ------ | ----------- |
| EMPTY | 0 | K8s doesn&#39;t have this, but it&#39;s possible to represent something similar. |
| EXISTS | 1 | &#34;Exists&#34; |
| EQUAL | 2 | &#34;Equal&#34; |



<a name="pps_v2-WorkerState"></a>

### WorkerState


| Name | Number | Description |
| ---- | ------ | ----------- |
| WORKER_STATE_UNKNOWN | 0 |  |
| POD_RUNNING | 1 |  |
| POD_SUCCESS | 2 |  |
| POD_FAILED | 3 |  |


 

 


<a name="pps_v2-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| InspectJob | [InspectJobRequest](#pps_v2-InspectJobRequest) | [JobInfo](#pps_v2-JobInfo) |  |
| InspectJobSet | [InspectJobSetRequest](#pps_v2-InspectJobSetRequest) | [JobInfo](#pps_v2-JobInfo) stream |  |
| ListJob | [ListJobRequest](#pps_v2-ListJobRequest) | [JobInfo](#pps_v2-JobInfo) stream | ListJob returns information about current and past Pachyderm jobs. |
| ListJobSet | [ListJobSetRequest](#pps_v2-ListJobSetRequest) | [JobSetInfo](#pps_v2-JobSetInfo) stream |  |
| SubscribeJob | [SubscribeJobRequest](#pps_v2-SubscribeJobRequest) | [JobInfo](#pps_v2-JobInfo) stream |  |
| DeleteJob | [DeleteJobRequest](#pps_v2-DeleteJobRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| StopJob | [StopJobRequest](#pps_v2-StopJobRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| InspectDatum | [InspectDatumRequest](#pps_v2-InspectDatumRequest) | [DatumInfo](#pps_v2-DatumInfo) |  |
| ListDatum | [ListDatumRequest](#pps_v2-ListDatumRequest) | [DatumInfo](#pps_v2-DatumInfo) stream | ListDatum returns information about each datum fed to a Pachyderm job |
| CreateDatum | [CreateDatumRequest](#pps_v2-CreateDatumRequest) stream | [DatumInfo](#pps_v2-DatumInfo) stream |  |
| RestartDatum | [RestartDatumRequest](#pps_v2-RestartDatumRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| RerunPipeline | [RerunPipelineRequest](#pps_v2-RerunPipelineRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| CreatePipeline | [CreatePipelineRequest](#pps_v2-CreatePipelineRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| CreatePipelineV2 | [CreatePipelineV2Request](#pps_v2-CreatePipelineV2Request) | [CreatePipelineV2Response](#pps_v2-CreatePipelineV2Response) |  |
| InspectPipeline | [InspectPipelineRequest](#pps_v2-InspectPipelineRequest) | [PipelineInfo](#pps_v2-PipelineInfo) |  |
| ListPipeline | [ListPipelineRequest](#pps_v2-ListPipelineRequest) | [PipelineInfo](#pps_v2-PipelineInfo) stream |  |
| DeletePipeline | [DeletePipelineRequest](#pps_v2-DeletePipelineRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| DeletePipelines | [DeletePipelinesRequest](#pps_v2-DeletePipelinesRequest) | [DeletePipelinesResponse](#pps_v2-DeletePipelinesResponse) |  |
| StartPipeline | [StartPipelineRequest](#pps_v2-StartPipelineRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| StopPipeline | [StopPipelineRequest](#pps_v2-StopPipelineRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| RunPipeline | [RunPipelineRequest](#pps_v2-RunPipelineRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| RunCron | [RunCronRequest](#pps_v2-RunCronRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| CheckStatus | [CheckStatusRequest](#pps_v2-CheckStatusRequest) | [CheckStatusResponse](#pps_v2-CheckStatusResponse) stream | Check Status returns the status of pipelines within a project. |
| CreateSecret | [CreateSecretRequest](#pps_v2-CreateSecretRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| DeleteSecret | [DeleteSecretRequest](#pps_v2-DeleteSecretRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| ListSecret | [.google.protobuf.Empty](#google-protobuf-Empty) | [SecretInfos](#pps_v2-SecretInfos) |  |
| InspectSecret | [InspectSecretRequest](#pps_v2-InspectSecretRequest) | [SecretInfo](#pps_v2-SecretInfo) |  |
| DeleteAll | [.google.protobuf.Empty](#google-protobuf-Empty) | [.google.protobuf.Empty](#google-protobuf-Empty) | DeleteAll deletes everything |
| GetLogs | [GetLogsRequest](#pps_v2-GetLogsRequest) | [LogMessage](#pps_v2-LogMessage) stream |  |
| ActivateAuth | [ActivateAuthRequest](#pps_v2-ActivateAuthRequest) | [ActivateAuthResponse](#pps_v2-ActivateAuthResponse) | An internal call that causes PPS to put itself into an auth-enabled state (all pipeline have tokens, correct permissions, etcd) |
| UpdateJobState | [UpdateJobStateRequest](#pps_v2-UpdateJobStateRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | An internal call used to move a job from one state to another |
| RunLoadTest | [RunLoadTestRequest](#pps_v2-RunLoadTestRequest) | [RunLoadTestResponse](#pps_v2-RunLoadTestResponse) | RunLoadTest runs a load test. |
| RunLoadTestDefault | [.google.protobuf.Empty](#google-protobuf-Empty) | [RunLoadTestResponse](#pps_v2-RunLoadTestResponse) | RunLoadTestDefault runs the default load test. |
| RenderTemplate | [RenderTemplateRequest](#pps_v2-RenderTemplateRequest) | [RenderTemplateResponse](#pps_v2-RenderTemplateResponse) | RenderTemplate renders the provided template and arguments into a list of Pipeline specicifications |
| ListTask | [.taskapi.ListTaskRequest](#taskapi-ListTaskRequest) | [.taskapi.TaskInfo](#taskapi-TaskInfo) stream | ListTask lists PPS tasks |
| GetKubeEvents | [LokiRequest](#pps_v2-LokiRequest) | [LokiLogMessage](#pps_v2-LokiLogMessage) stream | GetKubeEvents returns a stream of kubernetes events |
| QueryLoki | [LokiRequest](#pps_v2-LokiRequest) | [LokiLogMessage](#pps_v2-LokiLogMessage) stream | QueryLoki returns a stream of loki log messages given a query string |
| GetClusterDefaults | [GetClusterDefaultsRequest](#pps_v2-GetClusterDefaultsRequest) | [GetClusterDefaultsResponse](#pps_v2-GetClusterDefaultsResponse) | GetClusterDefaults returns the current cluster defaults. |
| SetClusterDefaults | [SetClusterDefaultsRequest](#pps_v2-SetClusterDefaultsRequest) | [SetClusterDefaultsResponse](#pps_v2-SetClusterDefaultsResponse) | SetClusterDefaults returns the current cluster defaults. |
| GetProjectDefaults | [GetProjectDefaultsRequest](#pps_v2-GetProjectDefaultsRequest) | [GetProjectDefaultsResponse](#pps_v2-GetProjectDefaultsResponse) | GetProjectDefaults returns the defaults for a particular project. |
| SetProjectDefaults | [SetProjectDefaultsRequest](#pps_v2-SetProjectDefaultsRequest) | [SetProjectDefaultsResponse](#pps_v2-SetProjectDefaultsResponse) | SetProjectDefaults sets the defaults for a particular project. |

 



<a name="protoextensions_json-schema-options-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## protoextensions/json-schema-options.proto



<a name="protoc-gen-jsonschema-EnumOptions"></a>

### EnumOptions
Custom EnumOptions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| enums_as_constants | [bool](#bool) |  | Enums tagged with this will have be encoded to use constants instead of simple types (supports value annotations): |
| enums_as_strings_only | [bool](#bool) |  | Enums tagged with this will only provide string values as options (not their numerical equivalents): |
| enums_trim_prefix | [bool](#bool) |  | Enums tagged with this will have enum name prefix removed from values: |
| ignore | [bool](#bool) |  | Enums tagged with this will not be processed |






<a name="protoc-gen-jsonschema-FieldOptions"></a>

### FieldOptions
Custom FieldOptions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ignore | [bool](#bool) |  | Fields tagged with this will be omitted from generated schemas |
| required | [bool](#bool) |  | Fields tagged with this will be marked as &#34;required&#34; in generated schemas |
| min_length | [int32](#int32) |  | Fields tagged with this will constrain strings using the &#34;minLength&#34; keyword in generated schemas |
| max_length | [int32](#int32) |  | Fields tagged with this will constrain strings using the &#34;maxLength&#34; keyword in generated schemas |
| pattern | [string](#string) |  | Fields tagged with this will constrain strings using the &#34;pattern&#34; keyword in generated schemas |






<a name="protoc-gen-jsonschema-FileOptions"></a>

### FileOptions
Custom FileOptions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ignore | [bool](#bool) |  | Files tagged with this will not be processed |
| extension | [string](#string) |  | Override the default file extension for schemas generated from this file |






<a name="protoc-gen-jsonschema-MessageOptions"></a>

### MessageOptions
Custom MessageOptions


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ignore | [bool](#bool) |  | Messages tagged with this will not be processed |
| all_fields_required | [bool](#bool) |  | Messages tagged with this will have all fields marked as &#34;required&#34;: |
| allow_null_values | [bool](#bool) |  | Messages tagged with this will additionally accept null values for all properties: |
| disallow_additional_properties | [bool](#bool) |  | Messages tagged with this will have all fields marked as not allowing additional properties: |
| enums_as_constants | [bool](#bool) |  | Messages tagged with this will have all nested enums encoded to use constants instead of simple types (supports value annotations): |





 

 


<a name="protoextensions_json-schema-options-proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| enum_options | EnumOptions | .google.protobuf.EnumOptions | 1128 |  |
| field_options | FieldOptions | .google.protobuf.FieldOptions | 1125 |  |
| file_options | FileOptions | .google.protobuf.FileOptions | 1126 |  |
| message_options | MessageOptions | .google.protobuf.MessageOptions | 1127 |  |

 

 



<a name="protoextensions_log-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## protoextensions/log.proto


 

 


<a name="protoextensions_log-proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| half | bool | .google.protobuf.FieldOptions | 50002 |  |
| mask | bool | .google.protobuf.FieldOptions | 50001 |  |

 

 



<a name="protoextensions_validate-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## protoextensions/validate.proto



<a name="validate-AnyRules"></a>

### AnyRules
AnyRules describe constraints applied exclusively to the
`google.protobuf.Any` well-known type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| required | [bool](#bool) | optional | Required specifies that this field must be set |
| in | [string](#string) | repeated | In specifies that this field&#39;s `type_url` must be equal to one of the specified values. |
| not_in | [string](#string) | repeated | NotIn specifies that this field&#39;s `type_url` must not be equal to any of the specified values. |






<a name="validate-BoolRules"></a>

### BoolRules
BoolRules describes the constraints applied to `bool` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [bool](#bool) | optional | Const specifies that this field must be exactly the specified value |






<a name="validate-BytesRules"></a>

### BytesRules
BytesRules describe the constraints applied to `bytes` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [bytes](#bytes) | optional | Const specifies that this field must be exactly the specified value |
| len | [uint64](#uint64) | optional | Len specifies that this field must be the specified number of bytes |
| min_len | [uint64](#uint64) | optional | MinLen specifies that this field must be the specified number of bytes at a minimum |
| max_len | [uint64](#uint64) | optional | MaxLen specifies that this field must be the specified number of bytes at a maximum |
| pattern | [string](#string) | optional | Pattern specifes that this field must match against the specified regular expression (RE2 syntax). The included expression should elide any delimiters. |
| prefix | [bytes](#bytes) | optional | Prefix specifies that this field must have the specified bytes at the beginning of the string. |
| suffix | [bytes](#bytes) | optional | Suffix specifies that this field must have the specified bytes at the end of the string. |
| contains | [bytes](#bytes) | optional | Contains specifies that this field must have the specified bytes anywhere in the string. |
| in | [bytes](#bytes) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [bytes](#bytes) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ip | [bool](#bool) | optional | Ip specifies that the field must be a valid IP (v4 or v6) address in byte format |
| ipv4 | [bool](#bool) | optional | Ipv4 specifies that the field must be a valid IPv4 address in byte format |
| ipv6 | [bool](#bool) | optional | Ipv6 specifies that the field must be a valid IPv6 address in byte format |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-DoubleRules"></a>

### DoubleRules
DoubleRules describes the constraints applied to `double` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [double](#double) | optional | Const specifies that this field must be exactly the specified value |
| lt | [double](#double) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [double](#double) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [double](#double) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [double](#double) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [double](#double) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [double](#double) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-DurationRules"></a>

### DurationRules
DurationRules describe the constraints applied exclusively to the
`google.protobuf.Duration` well-known type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| required | [bool](#bool) | optional | Required specifies that this field must be set |
| const | [google.protobuf.Duration](#google-protobuf-Duration) | optional | Const specifies that this field must be exactly the specified value |
| lt | [google.protobuf.Duration](#google-protobuf-Duration) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [google.protobuf.Duration](#google-protobuf-Duration) | optional | Lt specifies that this field must be less than the specified value, inclusive |
| gt | [google.protobuf.Duration](#google-protobuf-Duration) | optional | Gt specifies that this field must be greater than the specified value, exclusive |
| gte | [google.protobuf.Duration](#google-protobuf-Duration) | optional | Gte specifies that this field must be greater than the specified value, inclusive |
| in | [google.protobuf.Duration](#google-protobuf-Duration) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [google.protobuf.Duration](#google-protobuf-Duration) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |






<a name="validate-EnumRules"></a>

### EnumRules
EnumRules describe the constraints applied to enum values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [int32](#int32) | optional | Const specifies that this field must be exactly the specified value |
| defined_only | [bool](#bool) | optional | DefinedOnly specifies that this field must be only one of the defined values for this enum, failing on any undefined value. |
| in | [int32](#int32) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [int32](#int32) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |






<a name="validate-FieldRules"></a>

### FieldRules
FieldRules encapsulates the rules for each type of field. Depending on the
field, the correct set should be used to ensure proper validations.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| message | [MessageRules](#validate-MessageRules) | optional |  |
| float | [FloatRules](#validate-FloatRules) | optional | Scalar Field Types |
| double | [DoubleRules](#validate-DoubleRules) | optional |  |
| int32 | [Int32Rules](#validate-Int32Rules) | optional |  |
| int64 | [Int64Rules](#validate-Int64Rules) | optional |  |
| uint32 | [UInt32Rules](#validate-UInt32Rules) | optional |  |
| uint64 | [UInt64Rules](#validate-UInt64Rules) | optional |  |
| sint32 | [SInt32Rules](#validate-SInt32Rules) | optional |  |
| sint64 | [SInt64Rules](#validate-SInt64Rules) | optional |  |
| fixed32 | [Fixed32Rules](#validate-Fixed32Rules) | optional |  |
| fixed64 | [Fixed64Rules](#validate-Fixed64Rules) | optional |  |
| sfixed32 | [SFixed32Rules](#validate-SFixed32Rules) | optional |  |
| sfixed64 | [SFixed64Rules](#validate-SFixed64Rules) | optional |  |
| bool | [BoolRules](#validate-BoolRules) | optional |  |
| string | [StringRules](#validate-StringRules) | optional |  |
| bytes | [BytesRules](#validate-BytesRules) | optional |  |
| enum | [EnumRules](#validate-EnumRules) | optional | Complex Field Types |
| repeated | [RepeatedRules](#validate-RepeatedRules) | optional |  |
| map | [MapRules](#validate-MapRules) | optional |  |
| any | [AnyRules](#validate-AnyRules) | optional | Well-Known Field Types |
| duration | [DurationRules](#validate-DurationRules) | optional |  |
| timestamp | [TimestampRules](#validate-TimestampRules) | optional |  |






<a name="validate-Fixed32Rules"></a>

### Fixed32Rules
Fixed32Rules describes the constraints applied to `fixed32` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [fixed32](#fixed32) | optional | Const specifies that this field must be exactly the specified value |
| lt | [fixed32](#fixed32) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [fixed32](#fixed32) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [fixed32](#fixed32) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [fixed32](#fixed32) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [fixed32](#fixed32) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [fixed32](#fixed32) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-Fixed64Rules"></a>

### Fixed64Rules
Fixed64Rules describes the constraints applied to `fixed64` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [fixed64](#fixed64) | optional | Const specifies that this field must be exactly the specified value |
| lt | [fixed64](#fixed64) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [fixed64](#fixed64) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [fixed64](#fixed64) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [fixed64](#fixed64) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [fixed64](#fixed64) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [fixed64](#fixed64) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-FloatRules"></a>

### FloatRules
FloatRules describes the constraints applied to `float` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [float](#float) | optional | Const specifies that this field must be exactly the specified value |
| lt | [float](#float) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [float](#float) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [float](#float) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [float](#float) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [float](#float) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [float](#float) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-Int32Rules"></a>

### Int32Rules
Int32Rules describes the constraints applied to `int32` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [int32](#int32) | optional | Const specifies that this field must be exactly the specified value |
| lt | [int32](#int32) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [int32](#int32) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [int32](#int32) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [int32](#int32) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [int32](#int32) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [int32](#int32) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-Int64Rules"></a>

### Int64Rules
Int64Rules describes the constraints applied to `int64` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [int64](#int64) | optional | Const specifies that this field must be exactly the specified value |
| lt | [int64](#int64) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [int64](#int64) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [int64](#int64) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [int64](#int64) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [int64](#int64) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [int64](#int64) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-MapRules"></a>

### MapRules
MapRules describe the constraints applied to `map` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| min_pairs | [uint64](#uint64) | optional | MinPairs specifies that this field must have the specified number of KVs at a minimum |
| max_pairs | [uint64](#uint64) | optional | MaxPairs specifies that this field must have the specified number of KVs at a maximum |
| no_sparse | [bool](#bool) | optional | NoSparse specifies values in this field cannot be unset. This only applies to map&#39;s with message value types. |
| keys | [FieldRules](#validate-FieldRules) | optional | Keys specifies the constraints to be applied to each key in the field. |
| values | [FieldRules](#validate-FieldRules) | optional | Values specifies the constraints to be applied to the value of each key in the field. Message values will still have their validations evaluated unless skip is specified here. |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-MessageRules"></a>

### MessageRules
MessageRules describe the constraints applied to embedded message values.
For message-type fields, validation is performed recursively.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| skip | [bool](#bool) | optional | Skip specifies that the validation rules of this field should not be evaluated |
| required | [bool](#bool) | optional | Required specifies that this field must be set |






<a name="validate-RepeatedRules"></a>

### RepeatedRules
RepeatedRules describe the constraints applied to `repeated` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| min_items | [uint64](#uint64) | optional | MinItems specifies that this field must have the specified number of items at a minimum |
| max_items | [uint64](#uint64) | optional | MaxItems specifies that this field must have the specified number of items at a maximum |
| unique | [bool](#bool) | optional | Unique specifies that all elements in this field must be unique. This contraint is only applicable to scalar and enum types (messages are not supported). |
| items | [FieldRules](#validate-FieldRules) | optional | Items specifies the contraints to be applied to each item in the field. Repeated message fields will still execute validation against each item unless skip is specified here. |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-SFixed32Rules"></a>

### SFixed32Rules
SFixed32Rules describes the constraints applied to `sfixed32` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [sfixed32](#sfixed32) | optional | Const specifies that this field must be exactly the specified value |
| lt | [sfixed32](#sfixed32) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [sfixed32](#sfixed32) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [sfixed32](#sfixed32) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [sfixed32](#sfixed32) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [sfixed32](#sfixed32) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [sfixed32](#sfixed32) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-SFixed64Rules"></a>

### SFixed64Rules
SFixed64Rules describes the constraints applied to `sfixed64` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [sfixed64](#sfixed64) | optional | Const specifies that this field must be exactly the specified value |
| lt | [sfixed64](#sfixed64) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [sfixed64](#sfixed64) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [sfixed64](#sfixed64) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [sfixed64](#sfixed64) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [sfixed64](#sfixed64) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [sfixed64](#sfixed64) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-SInt32Rules"></a>

### SInt32Rules
SInt32Rules describes the constraints applied to `sint32` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [sint32](#sint32) | optional | Const specifies that this field must be exactly the specified value |
| lt | [sint32](#sint32) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [sint32](#sint32) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [sint32](#sint32) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [sint32](#sint32) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [sint32](#sint32) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [sint32](#sint32) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-SInt64Rules"></a>

### SInt64Rules
SInt64Rules describes the constraints applied to `sint64` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [sint64](#sint64) | optional | Const specifies that this field must be exactly the specified value |
| lt | [sint64](#sint64) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [sint64](#sint64) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [sint64](#sint64) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [sint64](#sint64) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [sint64](#sint64) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [sint64](#sint64) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-StringRules"></a>

### StringRules
StringRules describe the constraints applied to `string` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [string](#string) | optional | Const specifies that this field must be exactly the specified value |
| len | [uint64](#uint64) | optional | Len specifies that this field must be the specified number of characters (Unicode code points). Note that the number of characters may differ from the number of bytes in the string. |
| min_len | [uint64](#uint64) | optional | MinLen specifies that this field must be the specified number of characters (Unicode code points) at a minimum. Note that the number of characters may differ from the number of bytes in the string. |
| max_len | [uint64](#uint64) | optional | MaxLen specifies that this field must be the specified number of characters (Unicode code points) at a maximum. Note that the number of characters may differ from the number of bytes in the string. |
| len_bytes | [uint64](#uint64) | optional | LenBytes specifies that this field must be the specified number of bytes |
| min_bytes | [uint64](#uint64) | optional | MinBytes specifies that this field must be the specified number of bytes at a minimum |
| max_bytes | [uint64](#uint64) | optional | MaxBytes specifies that this field must be the specified number of bytes at a maximum |
| pattern | [string](#string) | optional | Pattern specifes that this field must match against the specified regular expression (RE2 syntax). The included expression should elide any delimiters. |
| prefix | [string](#string) | optional | Prefix specifies that this field must have the specified substring at the beginning of the string. |
| suffix | [string](#string) | optional | Suffix specifies that this field must have the specified substring at the end of the string. |
| contains | [string](#string) | optional | Contains specifies that this field must have the specified substring anywhere in the string. |
| not_contains | [string](#string) | optional | NotContains specifies that this field cannot have the specified substring anywhere in the string. |
| in | [string](#string) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [string](#string) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| email | [bool](#bool) | optional | Email specifies that the field must be a valid email address as defined by RFC 5322 |
| hostname | [bool](#bool) | optional | Hostname specifies that the field must be a valid hostname as defined by RFC 1034. This constraint does not support internationalized domain names (IDNs). |
| ip | [bool](#bool) | optional | Ip specifies that the field must be a valid IP (v4 or v6) address. Valid IPv6 addresses should not include surrounding square brackets. |
| ipv4 | [bool](#bool) | optional | Ipv4 specifies that the field must be a valid IPv4 address. |
| ipv6 | [bool](#bool) | optional | Ipv6 specifies that the field must be a valid IPv6 address. Valid IPv6 addresses should not include surrounding square brackets. |
| uri | [bool](#bool) | optional | Uri specifies that the field must be a valid, absolute URI as defined by RFC 3986 |
| uri_ref | [bool](#bool) | optional | UriRef specifies that the field must be a valid URI as defined by RFC 3986 and may be relative or absolute. |
| address | [bool](#bool) | optional | Address specifies that the field must be either a valid hostname as defined by RFC 1034 (which does not support internationalized domain names or IDNs), or it can be a valid IP (v4 or v6). |
| uuid | [bool](#bool) | optional | Uuid specifies that the field must be a valid UUID as defined by RFC 4122 |
| well_known_regex | [KnownRegex](#validate-KnownRegex) | optional | WellKnownRegex specifies a common well known pattern defined as a regex. |
| strict | [bool](#bool) | optional | This applies to regexes HTTP_HEADER_NAME and HTTP_HEADER_VALUE to enable strict header validation. By default, this is true, and HTTP header validations are RFC-compliant. Setting to false will enable a looser validations that only disallows \r\n\0 characters, which can be used to bypass header matching rules. Default: true |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-TimestampRules"></a>

### TimestampRules
TimestampRules describe the constraints applied exclusively to the
`google.protobuf.Timestamp` well-known type


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| required | [bool](#bool) | optional | Required specifies that this field must be set |
| const | [google.protobuf.Timestamp](#google-protobuf-Timestamp) | optional | Const specifies that this field must be exactly the specified value |
| lt | [google.protobuf.Timestamp](#google-protobuf-Timestamp) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [google.protobuf.Timestamp](#google-protobuf-Timestamp) | optional | Lte specifies that this field must be less than the specified value, inclusive |
| gt | [google.protobuf.Timestamp](#google-protobuf-Timestamp) | optional | Gt specifies that this field must be greater than the specified value, exclusive |
| gte | [google.protobuf.Timestamp](#google-protobuf-Timestamp) | optional | Gte specifies that this field must be greater than the specified value, inclusive |
| lt_now | [bool](#bool) | optional | LtNow specifies that this must be less than the current time. LtNow can only be used with the Within rule. |
| gt_now | [bool](#bool) | optional | GtNow specifies that this must be greater than the current time. GtNow can only be used with the Within rule. |
| within | [google.protobuf.Duration](#google-protobuf-Duration) | optional | Within specifies that this field must be within this duration of the current time. This constraint can be used alone or with the LtNow and GtNow rules. |






<a name="validate-UInt32Rules"></a>

### UInt32Rules
UInt32Rules describes the constraints applied to `uint32` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [uint32](#uint32) | optional | Const specifies that this field must be exactly the specified value |
| lt | [uint32](#uint32) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [uint32](#uint32) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [uint32](#uint32) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [uint32](#uint32) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [uint32](#uint32) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [uint32](#uint32) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |






<a name="validate-UInt64Rules"></a>

### UInt64Rules
UInt64Rules describes the constraints applied to `uint64` values


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| const | [uint64](#uint64) | optional | Const specifies that this field must be exactly the specified value |
| lt | [uint64](#uint64) | optional | Lt specifies that this field must be less than the specified value, exclusive |
| lte | [uint64](#uint64) | optional | Lte specifies that this field must be less than or equal to the specified value, inclusive |
| gt | [uint64](#uint64) | optional | Gt specifies that this field must be greater than the specified value, exclusive. If the value of Gt is larger than a specified Lt or Lte, the range is reversed. |
| gte | [uint64](#uint64) | optional | Gte specifies that this field must be greater than or equal to the specified value, inclusive. If the value of Gte is larger than a specified Lt or Lte, the range is reversed. |
| in | [uint64](#uint64) | repeated | In specifies that this field must be equal to one of the specified values |
| not_in | [uint64](#uint64) | repeated | NotIn specifies that this field cannot be equal to one of the specified values |
| ignore_empty | [bool](#bool) | optional | IgnoreEmpty specifies that the validation rules of this field should be evaluated only if the field is not empty |





 


<a name="validate-KnownRegex"></a>

### KnownRegex
WellKnownRegex contain some well-known patterns.

| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| HTTP_HEADER_NAME | 1 | HTTP header name as defined by RFC 7230. |
| HTTP_HEADER_VALUE | 2 | HTTP header value as defined by RFC 7230. |


 


<a name="protoextensions_validate-proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| rules | FieldRules | .google.protobuf.FieldOptions | 1071 | Rules specify the validations to be performed on this field. By default, no validation is performed against a field. |
| disabled | bool | .google.protobuf.MessageOptions | 1071 | Disabled nullifies any validation rules for this message, including any message fields associated with it that do support validation. |
| ignored | bool | .google.protobuf.MessageOptions | 1072 | Ignore skips generation of validation methods for this message. |
| required | bool | .google.protobuf.OneofOptions | 1071 | Required ensures that exactly one the field options in a oneof is set; validation fails if no fields in the oneof are set. |

 

 



<a name="proxy_proxy-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## proxy/proxy.proto



<a name="proxy-ListenRequest"></a>

### ListenRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| channel | [string](#string) |  |  |






<a name="proxy-ListenResponse"></a>

### ListenResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| extra | [string](#string) |  |  |





 

 

 


<a name="proxy-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Listen | [ListenRequest](#proxy-ListenRequest) | [ListenResponse](#proxy-ListenResponse) stream | Listen streams database events. It signals that it is internally set up by sending an initial empty ListenResponse. |

 



<a name="server_pfs_server_pfsserver-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## server/pfs/server/pfsserver.proto



<a name="pfsserver-CompactTask"></a>

### CompactTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inputs | [string](#string) | repeated |  |
| path_range | [PathRange](#pfsserver-PathRange) |  |  |






<a name="pfsserver-CompactTaskResult"></a>

### CompactTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="pfsserver-ConcatTask"></a>

### ConcatTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inputs | [string](#string) | repeated |  |






<a name="pfsserver-ConcatTaskResult"></a>

### ConcatTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="pfsserver-GetFileURLTask"></a>

### GetFileURLTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| URL | [string](#string) |  |  |
| Fileset | [string](#string) |  |  |
| path_range | [pfs_v2.PathRange](#pfs_v2-PathRange) |  |  |






<a name="pfsserver-GetFileURLTaskResult"></a>

### GetFileURLTaskResult







<a name="pfsserver-PathRange"></a>

### PathRange



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| lower | [string](#string) |  |  |
| upper | [string](#string) |  |  |






<a name="pfsserver-PutFileURLTask"></a>

### PutFileURLTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| dst | [string](#string) |  |  |
| datum | [string](#string) |  |  |
| URL | [string](#string) |  |  |
| paths | [string](#string) | repeated |  |
| start_offset | [int64](#int64) |  |  |
| end_offset | [int64](#int64) |  |  |






<a name="pfsserver-PutFileURLTaskResult"></a>

### PutFileURLTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="pfsserver-ShardTask"></a>

### ShardTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| inputs | [string](#string) | repeated |  |
| path_range | [PathRange](#pfsserver-PathRange) |  |  |






<a name="pfsserver-ShardTaskResult"></a>

### ShardTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| compact_tasks | [CompactTask](#pfsserver-CompactTask) | repeated |  |






<a name="pfsserver-ValidateTask"></a>

### ValidateTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| path_range | [PathRange](#pfsserver-PathRange) |  |  |






<a name="pfsserver-ValidateTaskResult"></a>

### ValidateTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| first | [index.Index](#index-Index) |  |  |
| last | [index.Index](#index-Index) |  |  |
| error | [string](#string) |  |  |
| size_bytes | [int64](#int64) |  |  |





 

 

 

 



<a name="server_worker_common_common-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## server/worker/common/common.proto



<a name="common-Input"></a>

### Input



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_info | [pfs_v2.FileInfo](#pfs_v2-FileInfo) |  |  |
| name | [string](#string) |  |  |
| join_on | [string](#string) |  |  |
| outer_join | [bool](#bool) |  |  |
| group_by | [string](#string) |  |  |
| lazy | [bool](#bool) |  |  |
| branch | [string](#string) |  |  |
| empty_files | [bool](#bool) |  |  |
| s3 | [bool](#bool) |  | If set, workers won&#39;t create an input directory for this input |





 

 

 

 



<a name="server_worker_datum_datum-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## server/worker/datum/datum.proto



<a name="datum-ComposeTask"></a>

### ComposeTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_ids | [string](#string) | repeated |  |
| auth_token | [string](#string) |  |  |






<a name="datum-ComposeTaskResult"></a>

### ComposeTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |






<a name="datum-CrossTask"></a>

### CrossTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_ids | [string](#string) | repeated |  |
| base_file_set_index | [int64](#int64) |  |  |
| base_file_set_path_range | [pfs_v2.PathRange](#pfs_v2-PathRange) |  |  |
| base_index | [int64](#int64) |  |  |
| auth_token | [string](#string) |  |  |






<a name="datum-CrossTaskResult"></a>

### CrossTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |






<a name="datum-KeyTask"></a>

### KeyTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |
| path_range | [pfs_v2.PathRange](#pfs_v2-PathRange) |  |  |
| type | [KeyTask.Type](#datum-KeyTask-Type) |  |  |
| auth_token | [string](#string) |  |  |






<a name="datum-KeyTaskResult"></a>

### KeyTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |






<a name="datum-MergeTask"></a>

### MergeTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_ids | [string](#string) | repeated |  |
| path_range | [pfs_v2.PathRange](#pfs_v2-PathRange) |  |  |
| type | [MergeTask.Type](#datum-MergeTask-Type) |  |  |
| auth_token | [string](#string) |  |  |






<a name="datum-MergeTaskResult"></a>

### MergeTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |






<a name="datum-Meta"></a>

### Meta



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [pps_v2.Job](#pps_v2-Job) |  |  |
| inputs | [common.Input](#common-Input) | repeated |  |
| hash | [string](#string) |  |  |
| state | [State](#datum-State) |  |  |
| reason | [string](#string) |  |  |
| stats | [pps_v2.ProcessStats](#pps_v2-ProcessStats) |  |  |
| index | [int64](#int64) |  |  |
| image_id | [string](#string) |  |  |






<a name="datum-PFSTask"></a>

### PFSTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| input | [pps_v2.PFSInput](#pps_v2-PFSInput) |  |  |
| path_range | [pfs_v2.PathRange](#pfs_v2-PathRange) |  |  |
| base_index | [int64](#int64) |  |  |
| auth_token | [string](#string) |  |  |






<a name="datum-PFSTaskResult"></a>

### PFSTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |






<a name="datum-SetSpec"></a>

### SetSpec



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | [int64](#int64) |  |  |
| size_bytes | [int64](#int64) |  |  |






<a name="datum-Stats"></a>

### Stats



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| process_stats | [pps_v2.ProcessStats](#pps_v2-ProcessStats) |  |  |
| processed | [int64](#int64) |  |  |
| skipped | [int64](#int64) |  |  |
| total | [int64](#int64) |  |  |
| failed | [int64](#int64) |  |  |
| recovered | [int64](#int64) |  |  |
| failed_id | [string](#string) |  |  |





 


<a name="datum-KeyTask-Type"></a>

### KeyTask.Type


| Name | Number | Description |
| ---- | ------ | ----------- |
| JOIN | 0 |  |
| GROUP | 1 |  |



<a name="datum-MergeTask-Type"></a>

### MergeTask.Type


| Name | Number | Description |
| ---- | ------ | ----------- |
| JOIN | 0 |  |
| GROUP | 1 |  |



<a name="datum-State"></a>

### State


| Name | Number | Description |
| ---- | ------ | ----------- |
| PROCESSED | 0 |  |
| FAILED | 1 |  |
| RECOVERED | 2 |  |


 

 

 



<a name="server_worker_pipeline_transform_transform-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## server/worker/pipeline/transform/transform.proto



<a name="pachyderm-worker-pipeline-transform-CreateDatumSetsTask"></a>

### CreateDatumSetsTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |
| path_range | [pfs_v2.PathRange](#pfs_v2-PathRange) |  |  |
| set_spec | [datum.SetSpec](#datum-SetSpec) |  |  |
| auth_token | [string](#string) |  |  |






<a name="pachyderm-worker-pipeline-transform-CreateDatumSetsTaskResult"></a>

### CreateDatumSetsTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| datum_sets | [pfs_v2.PathRange](#pfs_v2-PathRange) | repeated |  |






<a name="pachyderm-worker-pipeline-transform-CreateParallelDatumsTask"></a>

### CreateParallelDatumsTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [pps_v2.Job](#pps_v2-Job) |  |  |
| salt | [string](#string) |  |  |
| file_set_id | [string](#string) |  |  |
| base_file_set_id | [string](#string) |  |  |
| path_range | [pfs_v2.PathRange](#pfs_v2-PathRange) |  |  |
| auth_token | [string](#string) |  |  |






<a name="pachyderm-worker-pipeline-transform-CreateParallelDatumsTaskResult"></a>

### CreateParallelDatumsTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |
| stats | [datum.Stats](#datum-Stats) |  |  |






<a name="pachyderm-worker-pipeline-transform-CreateSerialDatumsTask"></a>

### CreateSerialDatumsTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [pps_v2.Job](#pps_v2-Job) |  |  |
| salt | [string](#string) |  |  |
| file_set_id | [string](#string) |  |  |
| base_meta_commit | [pfs_v2.Commit](#pfs_v2-Commit) |  |  |
| no_skip | [bool](#bool) |  |  |
| path_range | [pfs_v2.PathRange](#pfs_v2-PathRange) |  |  |
| auth_token | [string](#string) |  |  |






<a name="pachyderm-worker-pipeline-transform-CreateSerialDatumsTaskResult"></a>

### CreateSerialDatumsTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_set_id | [string](#string) |  |  |
| output_delete_file_set_id | [string](#string) |  |  |
| meta_delete_file_set_id | [string](#string) |  |  |
| stats | [datum.Stats](#datum-Stats) |  |  |






<a name="pachyderm-worker-pipeline-transform-DatumSetTask"></a>

### DatumSetTask



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job | [pps_v2.Job](#pps_v2-Job) |  |  |
| file_set_id | [string](#string) |  |  |
| path_range | [pfs_v2.PathRange](#pfs_v2-PathRange) |  |  |
| output_commit | [pfs_v2.Commit](#pfs_v2-Commit) |  |  |






<a name="pachyderm-worker-pipeline-transform-DatumSetTaskResult"></a>

### DatumSetTaskResult



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| output_file_set_id | [string](#string) |  |  |
| meta_file_set_id | [string](#string) |  |  |
| stats | [datum.Stats](#datum-Stats) |  |  |





 

 

 

 



<a name="task_task-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## task/task.proto



<a name="taskapi-Group"></a>

### Group



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| namespace | [string](#string) |  |  |
| group | [string](#string) |  |  |






<a name="taskapi-ListTaskRequest"></a>

### ListTaskRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| group | [Group](#taskapi-Group) |  |  |






<a name="taskapi-TaskInfo"></a>

### TaskInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| group | [Group](#taskapi-Group) |  |  |
| state | [State](#taskapi-State) |  |  |
| reason | [string](#string) |  |  |
| input_type | [string](#string) |  |  |
| input_data | [string](#string) |  |  |





 


<a name="taskapi-State"></a>

### State


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 |  |
| RUNNING | 1 |  |
| SUCCESS | 2 |  |
| FAILURE | 3 |  |
| CLAIMED | 4 | not a real state used by task logic |


 

 

 



<a name="transaction_transaction-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## transaction/transaction.proto



<a name="transaction_v2-BatchTransactionRequest"></a>

### BatchTransactionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| requests | [TransactionRequest](#transaction_v2-TransactionRequest) | repeated |  |






<a name="transaction_v2-DeleteAllRequest"></a>

### DeleteAllRequest







<a name="transaction_v2-DeleteTransactionRequest"></a>

### DeleteTransactionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| transaction | [Transaction](#transaction_v2-Transaction) |  |  |






<a name="transaction_v2-FinishTransactionRequest"></a>

### FinishTransactionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| transaction | [Transaction](#transaction_v2-Transaction) |  |  |






<a name="transaction_v2-InspectTransactionRequest"></a>

### InspectTransactionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| transaction | [Transaction](#transaction_v2-Transaction) |  |  |






<a name="transaction_v2-ListTransactionRequest"></a>

### ListTransactionRequest







<a name="transaction_v2-StartTransactionRequest"></a>

### StartTransactionRequest







<a name="transaction_v2-Transaction"></a>

### Transaction



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |






<a name="transaction_v2-TransactionInfo"></a>

### TransactionInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| transaction | [Transaction](#transaction_v2-Transaction) |  |  |
| requests | [TransactionRequest](#transaction_v2-TransactionRequest) | repeated |  |
| responses | [TransactionResponse](#transaction_v2-TransactionResponse) | repeated |  |
| started | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| version | [uint64](#uint64) |  |  |






<a name="transaction_v2-TransactionInfos"></a>

### TransactionInfos



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| transaction_info | [TransactionInfo](#transaction_v2-TransactionInfo) | repeated |  |






<a name="transaction_v2-TransactionRequest"></a>

### TransactionRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| create_repo | [pfs_v2.CreateRepoRequest](#pfs_v2-CreateRepoRequest) |  | Exactly one of these fields should be set |
| delete_repo | [pfs_v2.DeleteRepoRequest](#pfs_v2-DeleteRepoRequest) |  |  |
| start_commit | [pfs_v2.StartCommitRequest](#pfs_v2-StartCommitRequest) |  |  |
| finish_commit | [pfs_v2.FinishCommitRequest](#pfs_v2-FinishCommitRequest) |  |  |
| squash_commit_set | [pfs_v2.SquashCommitSetRequest](#pfs_v2-SquashCommitSetRequest) |  |  |
| create_branch | [pfs_v2.CreateBranchRequest](#pfs_v2-CreateBranchRequest) |  |  |
| delete_branch | [pfs_v2.DeleteBranchRequest](#pfs_v2-DeleteBranchRequest) |  |  |
| update_job_state | [pps_v2.UpdateJobStateRequest](#pps_v2-UpdateJobStateRequest) |  |  |
| stop_job | [pps_v2.StopJobRequest](#pps_v2-StopJobRequest) |  |  |
| create_pipeline_v2 | [pps_v2.CreatePipelineTransaction](#pps_v2-CreatePipelineTransaction) |  |  |






<a name="transaction_v2-TransactionResponse"></a>

### TransactionResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| commit | [pfs_v2.Commit](#pfs_v2-Commit) |  | At most, one of these fields should be set (most responses are empty)

Only used for StartCommit - any way we can deterministically provide this before finishing the transaction? |





 

 

 


<a name="transaction_v2-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| BatchTransaction | [BatchTransactionRequest](#transaction_v2-BatchTransactionRequest) | [TransactionInfo](#transaction_v2-TransactionInfo) | Transaction rpcs |
| StartTransaction | [StartTransactionRequest](#transaction_v2-StartTransactionRequest) | [Transaction](#transaction_v2-Transaction) |  |
| InspectTransaction | [InspectTransactionRequest](#transaction_v2-InspectTransactionRequest) | [TransactionInfo](#transaction_v2-TransactionInfo) |  |
| DeleteTransaction | [DeleteTransactionRequest](#transaction_v2-DeleteTransactionRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |
| ListTransaction | [ListTransactionRequest](#transaction_v2-ListTransactionRequest) | [TransactionInfos](#transaction_v2-TransactionInfos) |  |
| FinishTransaction | [FinishTransactionRequest](#transaction_v2-FinishTransactionRequest) | [TransactionInfo](#transaction_v2-TransactionInfo) |  |
| DeleteAll | [DeleteAllRequest](#transaction_v2-DeleteAllRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) |  |

 



<a name="version_versionpb_version-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## version/versionpb/version.proto



<a name="versionpb_v2-Version"></a>

### Version



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| major | [uint32](#uint32) |  |  |
| minor | [uint32](#uint32) |  |  |
| micro | [uint32](#uint32) |  |  |
| additional | [string](#string) |  |  |
| git_commit | [string](#string) |  |  |
| git_tree_modified | [string](#string) |  |  |
| build_date | [string](#string) |  |  |
| go_version | [string](#string) |  |  |
| platform | [string](#string) |  |  |





 

 

 


<a name="versionpb_v2-API"></a>

### API


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetVersion | [.google.protobuf.Empty](#google-protobuf-Empty) | [Version](#versionpb_v2-Version) |  |

 



<a name="worker_worker-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## worker/worker.proto



<a name="pachyderm-worker-CancelRequest"></a>

### CancelRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| job_id | [string](#string) |  |  |
| data_filters | [string](#string) | repeated |  |






<a name="pachyderm-worker-CancelResponse"></a>

### CancelResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | [bool](#bool) |  |  |






<a name="pachyderm-worker-NextDatumRequest"></a>

### NextDatumRequest
Error indicates that the processing of the current datum errored.
Datum error semantics with datum batching enabled are similar to datum error
semantics without datum batching enabled in that the datum may be retried,
recovered, or result with a job failure.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| error | [string](#string) |  |  |






<a name="pachyderm-worker-NextDatumResponse"></a>

### NextDatumResponse
Env is a list of environment variables that should be set for the processing
of the next datum.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| env | [string](#string) | repeated |  |





 

 

 


<a name="pachyderm-worker-Worker"></a>

### Worker


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Status | [.google.protobuf.Empty](#google-protobuf-Empty) | [.pps_v2.WorkerStatus](#pps_v2-WorkerStatus) |  |
| Cancel | [CancelRequest](#pachyderm-worker-CancelRequest) | [CancelResponse](#pachyderm-worker-CancelResponse) |  |
| NextDatum | [NextDatumRequest](#pachyderm-worker-NextDatumRequest) | [NextDatumResponse](#pachyderm-worker-NextDatumResponse) | NextDatum should only be called by user code running in a pipeline with datum batching enabled. NextDatum will signal to the worker code that the user code is ready to proceed to the next datum. This generally means setting up the next datum&#39;s filesystem state and updating internal metadata similarly to datum processing in a normal pipeline. NextDatum is a synchronous operation, so user code should expect to block on this until the next datum is set up for processing. User code should generally be migratable to datum batching by wrapping it in a loop that calls next datum. |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

