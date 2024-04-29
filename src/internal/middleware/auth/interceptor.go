package auth

import (
	"context"

	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"

	"google.golang.org/grpc"
)

// authHandlers is a mapping of RPCs to authorization levels required to access them.
// This interceptor fails closed - whenever a new RPC is added, it's disabled
// until some authentication is added. Some RPCs may do additional auth checks
// beyond what's required by this interceptor.
var authHandlers = map[string]authHandler{
	//
	// Admin API
	//

	// Allow InspectCluster to succeed before a user logs in
	"/admin_v2.API/InspectCluster": unauthenticated,

	//
	// Auth API
	//

	// Activate only has an effect when auth is not enabled
	// Authenticate, Authorize and WhoAmI check auth status themselves
	// GetOIDCLogin is necessary to authenticate
	"/auth_v2.API/Activate":     unauthenticated,
	"/auth_v2.API/Authenticate": unauthenticated,
	"/auth_v2.API/Authorize":    unauthenticated,
	"/auth_v2.API/WhoAmI":       unauthenticated,
	"/auth_v2.API/GetOIDCLogin": unauthenticated,

	// TODO: restrict GetClusterRoleBinding to cluster admins?
	"/auth_v2.API/GetRoleBinding":        authenticated,
	"/auth_v2.API/ModifyRoleBinding":     authenticated,
	"/auth_v2.API/RevokeAuthToken":       authenticated,
	"/auth_v2.API/GetGroups":             authenticated,
	"/auth_v2.API/GetPermissions":        authenticated,
	"/auth_v2.API/GetRolesForPermission": authenticated,

	"/auth_v2.API/GetGroupsForPrincipal":      clusterPermissions(auth.Permission_CLUSTER_AUTH_GET_GROUPS),
	"/auth_v2.API/GetPermissionsForPrincipal": clusterPermissions(auth.Permission_CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL),
	"/auth_v2.API/GetConfiguration":           clusterPermissions(auth.Permission_CLUSTER_AUTH_GET_CONFIG),
	"/auth_v2.API/SetConfiguration":           clusterPermissions(auth.Permission_CLUSTER_AUTH_SET_CONFIG),
	"/auth_v2.API/GetRobotToken":              clusterPermissions(auth.Permission_CLUSTER_AUTH_GET_ROBOT_TOKEN),
	"/auth_v2.API/SetGroupsForUser":           clusterPermissions(auth.Permission_CLUSTER_AUTH_MODIFY_GROUP_MEMBERS),
	"/auth_v2.API/ModifyMembers":              clusterPermissions(auth.Permission_CLUSTER_AUTH_MODIFY_GROUP_MEMBERS),
	"/auth_v2.API/GetUsers":                   clusterPermissions(auth.Permission_CLUSTER_AUTH_GET_GROUP_USERS),
	"/auth_v2.API/ExtractAuthTokens":          clusterPermissions(auth.Permission_CLUSTER_AUTH_EXTRACT_TOKENS),
	"/auth_v2.API/RestoreAuthToken":           clusterPermissions(auth.Permission_CLUSTER_AUTH_RESTORE_TOKEN),
	"/auth_v2.API/Deactivate":                 clusterPermissions(auth.Permission_CLUSTER_AUTH_DEACTIVATE),
	"/auth_v2.API/DeleteExpiredAuthTokens":    clusterPermissions(auth.Permission_CLUSTER_AUTH_DELETE_EXPIRED_TOKENS),
	"/auth_v2.API/RevokeAuthTokensForUser":    clusterPermissions(auth.Permission_CLUSTER_AUTH_REVOKE_USER_TOKENS),
	"/auth_v2.API/RotateRootToken":            clusterPermissions(auth.Permission_CLUSTER_AUTH_ROTATE_ROOT_TOKEN),

	//
	// Debug API
	//

	"/debug_v2.Debug/Profile":               authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_DEBUG_DUMP)),
	"/debug_v2.Debug/Binary":                authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_DEBUG_DUMP)),
	"/debug_v2.Debug/Dump":                  authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_DEBUG_DUMP)),
	"/debug_v2.Debug/GetDumpV2Template":     authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_DEBUG_DUMP)),
	"/debug_v2.Debug/DumpV2":                authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_DEBUG_DUMP)),
	"/debug_v2.Debug/SetLogLevel":           authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_DEBUG_DUMP)),
	"/debug_v2.Debug/Trace":                 authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_DEBUG_DUMP)),
	"/debug_v2.Debug/RunPFSLoadTest":        authDisabledOr(authenticated),
	"/debug_v2.Debug/RunPFSLoadTestDefault": authDisabledOr(authenticated),

	//
	// Enterprise API
	//

	"/enterprise_v2.API/GetState":          unauthenticated,
	"/enterprise_v2.API/Activate":          authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_ENTERPRISE_ACTIVATE)),
	"/enterprise_v2.API/GetActivationCode": authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_ENTERPRISE_GET_CODE)),
	"/enterprise_v2.API/Deactivate":        authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_ENTERPRISE_DEACTIVATE)),
	"/enterprise_v2.API/Heartbeat":         authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_ENTERPRISE_HEARTBEAT)),
	"/enterprise_v2.API/Pause":             authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_ENTERPRISE_PAUSE)),
	"/enterprise_v2.API/Unpause":           authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_ENTERPRISE_PAUSE)),
	"/enterprise_v2.API/PauseStatus":       authDisabledOr(authenticated),

	//
	// Health API
	//
	"/grpc.health.v1.Health/Check": unauthenticated,
	"/grpc.health.v1.Health/Watch": unauthenticated,

	//
	// Identity API
	//
	"/identity_v2.API/SetIdentityServerConfig": clusterPermissions(auth.Permission_CLUSTER_IDENTITY_SET_CONFIG),
	"/identity_v2.API/GetIdentityServerConfig": clusterPermissions(auth.Permission_CLUSTER_IDENTITY_GET_CONFIG),
	"/identity_v2.API/CreateIDPConnector":      clusterPermissions(auth.Permission_CLUSTER_IDENTITY_CREATE_IDP),
	"/identity_v2.API/UpdateIDPConnector":      clusterPermissions(auth.Permission_CLUSTER_IDENTITY_UPDATE_IDP),
	"/identity_v2.API/ListIDPConnectors":       clusterPermissions(auth.Permission_CLUSTER_IDENTITY_LIST_IDPS),
	"/identity_v2.API/GetIDPConnector":         clusterPermissions(auth.Permission_CLUSTER_IDENTITY_GET_IDP),
	"/identity_v2.API/DeleteIDPConnector":      clusterPermissions(auth.Permission_CLUSTER_IDENTITY_DELETE_IDP),
	"/identity_v2.API/CreateOIDCClient":        clusterPermissions(auth.Permission_CLUSTER_IDENTITY_CREATE_OIDC_CLIENT),
	"/identity_v2.API/UpdateOIDCClient":        clusterPermissions(auth.Permission_CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT),
	"/identity_v2.API/GetOIDCClient":           clusterPermissions(auth.Permission_CLUSTER_IDENTITY_GET_OIDC_CLIENT),
	"/identity_v2.API/ListOIDCClients":         clusterPermissions(auth.Permission_CLUSTER_IDENTITY_LIST_OIDC_CLIENTS),
	"/identity_v2.API/DeleteOIDCClient":        clusterPermissions(auth.Permission_CLUSTER_IDENTITY_DELETE_OIDC_CLIENT),
	"/identity_v2.API/DeleteAll":               clusterPermissions(auth.Permission_CLUSTER_DELETE_ALL),

	//
	// License API
	//
	"/license_v2.API/Activate":          authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_LICENSE_ACTIVATE)),
	"/license_v2.API/GetActivationCode": authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_LICENSE_GET_CODE)),
	"/license_v2.API/AddCluster":        authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_LICENSE_ADD_CLUSTER)),
	"/license_v2.API/UpdateCluster":     authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_LICENSE_UPDATE_CLUSTER)),
	"/license_v2.API/DeleteCluster":     authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_LICENSE_DELETE_CLUSTER)),
	"/license_v2.API/ListClusters":      authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_LICENSE_LIST_CLUSTERS)),
	"/license_v2.API/DeleteAll":         authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_DELETE_ALL)),
	// Heartbeat relies on the shared secret generated at cluster registration-time
	"/license_v2.API/Heartbeat":        unauthenticated,
	"/license_v2.API/ListUserClusters": authDisabledOr(authenticated),

	//
	// PFS API
	//

	// TODO: Add methods to handle repo permissions
	"/pfs_v2.API/ActivateAuth":         clusterPermissions(auth.Permission_CLUSTER_AUTH_ACTIVATE),
	"/pfs_v2.API/CreateRepo":           authDisabledOr(authenticated),
	"/pfs_v2.API/InspectRepo":          authDisabledOr(authenticated),
	"/pfs_v2.API/ListRepo":             authDisabledOr(authenticated),
	"/pfs_v2.API/DeleteRepo":           authDisabledOr(authenticated),
	"/pfs_v2.API/DeleteRepos":          authDisabledOr(authenticated),
	"/pfs_v2.API/StartCommit":          authDisabledOr(authenticated),
	"/pfs_v2.API/FinishCommit":         authDisabledOr(authenticated),
	"/pfs_v2.API/InspectCommit":        authDisabledOr(authenticated),
	"/pfs_v2.API/ListCommit":           authDisabledOr(authenticated),
	"/pfs_v2.API/SubscribeCommit":      authDisabledOr(authenticated),
	"/pfs_v2.API/ClearCommit":          authDisabledOr(authenticated),
	"/pfs_v2.API/SquashCommit":         authDisabledOr(clusterPermissions(auth.Permission_REPO_DELETE_COMMIT)),
	"/pfs_v2.API/DropCommit":           authDisabledOr(clusterPermissions(auth.Permission_REPO_DELETE_COMMIT)),
	"/pfs_v2.API/InspectCommitSet":     authDisabledOr(authenticated),
	"/pfs_v2.API/ListCommitSet":        authDisabledOr(authenticated),
	"/pfs_v2.API/SquashCommitSet":      authDisabledOr(clusterPermissions(auth.Permission_REPO_DELETE_COMMIT)),
	"/pfs_v2.API/DropCommitSet":        authDisabledOr(clusterPermissions(auth.Permission_REPO_DELETE_COMMIT)),
	"/pfs_v2.API/FindCommits":          authDisabledOr(authenticated),
	"/pfs_v2.API/WalkCommitProvenance": authDisabledOr(authenticated),
	"/pfs_v2.API/WalkCommitSubvenance": authDisabledOr(authenticated),
	"/pfs_v2.API/CreateBranch":         authDisabledOr(authenticated),
	"/pfs_v2.API/InspectBranch":        authDisabledOr(authenticated),
	"/pfs_v2.API/ListBranch":           authDisabledOr(authenticated),
	"/pfs_v2.API/DeleteBranch":         authDisabledOr(authenticated),
	"/pfs_v2.API/WalkBranchProvenance": authDisabledOr(authenticated),
	"/pfs_v2.API/WalkBranchSubvenance": authDisabledOr(authenticated),
	"/pfs_v2.API/CreateProject":        authDisabledOr(clusterPermissions(auth.Permission_PROJECT_CREATE)),
	"/pfs_v2.API/InspectProject":       authDisabledOr(authenticated),
	"/pfs_v2.API/InspectProjectV2":     authDisabledOr(authenticated),
	"/pfs_v2.API/ListProject":          authDisabledOr(authenticated),
	"/pfs_v2.API/DeleteProject":        authDisabledOr(authenticated),
	"/pfs_v2.API/ModifyFile":           authDisabledOr(authenticated),
	"/pfs_v2.API/GetFile":              authDisabledOr(authenticated),
	// TODO: GetFileTAR is unauthenticated for performance reasons. Normal authentication
	// will be applied internally when a commit is used. When a file set id is used, we lean
	// on the capability based authentication of file sets.
	"/pfs_v2.API/GetFileTAR":     unauthenticated,
	"/pfs_v2.API/InspectFile":    authDisabledOr(authenticated),
	"/pfs_v2.API/ListFile":       authDisabledOr(authenticated),
	"/pfs_v2.API/WalkFile":       authDisabledOr(authenticated),
	"/pfs_v2.API/GlobFile":       authDisabledOr(authenticated),
	"/pfs_v2.API/DiffFile":       authDisabledOr(authenticated),
	"/pfs_v2.API/DeleteAll":      authDisabledOr(authenticated),
	"/pfs_v2.API/Fsck":           authDisabledOr(authenticated),
	"/pfs_v2.API/CreateFileSet":  authDisabledOr(authenticated),
	"/pfs_v2.API/GetFileSet":     authDisabledOr(authenticated),
	"/pfs_v2.API/AddFileSet":     authDisabledOr(authenticated),
	"/pfs_v2.API/RenewFileSet":   authDisabledOr(authenticated),
	"/pfs_v2.API/ComposeFileSet": authDisabledOr(authenticated),
	"/pfs_v2.API/ShardFileSet":   authDisabledOr(authenticated),
	"/pfs_v2.API/CheckStorage":   authDisabledOr(authenticated),
	"/pfs_v2.API/PutCache":       authDisabledOr(authenticated),
	"/pfs_v2.API/GetCache":       authDisabledOr(authenticated),
	"/pfs_v2.API/ClearCache":     authDisabledOr(authenticated),
	"/pfs_v2.API/ListTask":       authDisabledOr(authenticated),
	"/pfs_v2.API/Egress":         authDisabledOr(authenticated),
	"/pfs_v2.API/ReposSummary":   authDisabledOr(authenticated),

	//
	// Storage API
	//

	"/storage.Fileset/CreateFileset":  authDisabledOr(authenticated),
	"/storage.Fileset/ReadFileset":    authDisabledOr(authenticated),
	"/storage.Fileset/RenewFileset":   authDisabledOr(authenticated),
	"/storage.Fileset/ComposeFileset": authDisabledOr(authenticated),
	"/storage.Fileset/ShardFileset":   authDisabledOr(authenticated),

	//
	// PPS API
	//

	// TODO: Add per-repo permissions checks for these
	// TODO: split GetLogs into master and not-master and add check for pipeline permissions
	"/pps_v2.API/InspectJob":       authDisabledOr(authenticated),
	"/pps_v2.API/ListJob":          authDisabledOr(authenticated),
	"/pps_v2.API/ListJobStream":    authDisabledOr(authenticated),
	"/pps_v2.API/SubscribeJob":     authDisabledOr(authenticated),
	"/pps_v2.API/DeleteJob":        authDisabledOr(authenticated),
	"/pps_v2.API/StopJob":          authDisabledOr(authenticated),
	"/pps_v2.API/InspectJobSet":    authDisabledOr(authenticated),
	"/pps_v2.API/ListJobSet":       authDisabledOr(authenticated),
	"/pps_v2.API/InspectDatum":     authDisabledOr(authenticated),
	"/pps_v2.API/ListDatum":        authDisabledOr(authenticated),
	"/pps_v2.API/ListDatumStream":  authDisabledOr(authenticated),
	"/pps_v2.API/CreateDatum":      authDisabledOr(authenticated),
	"/pps_v2.API/RestartDatum":     authDisabledOr(authenticated),
	"/pps_v2.API/CreatePipeline":   authDisabledOr(authenticated),
	"/pps_v2.API/CreatePipelineV2": authDisabledOr(authenticated),
	"/pps_v2.API/RerunPipeline":    authDisabledOr(authenticated),
	"/pps_v2.API/InspectPipeline":  authDisabledOr(authenticated),
	"/pps_v2.API/DeletePipeline":   authDisabledOr(authenticated),
	"/pps_v2.API/DeletePipelines":  authDisabledOr(authenticated),
	"/pps_v2.API/StartPipeline":    authDisabledOr(authenticated),
	"/pps_v2.API/StopPipeline":     authDisabledOr(authenticated),
	"/pps_v2.API/RunPipeline":      authDisabledOr(authenticated),
	"/pps_v2.API/CheckStatus":      authDisabledOr(authenticated),
	"/pps_v2.API/RunCron":          authDisabledOr(authenticated),
	"/pps_v2.API/GetLogs":          authDisabledOr(authenticated),
	"/logs.API/GetLogs":            authDisabledOr(authenticated),
	"/pps_v2.API/GarbageCollect":   authDisabledOr(authenticated),
	"/pps_v2.API/UpdateJobState":   authDisabledOr(authenticated),
	"/pps_v2.API/ListPipeline":     authDisabledOr(authenticated),
	"/pps_v2.API/ActivateAuth":     clusterPermissions(auth.Permission_CLUSTER_AUTH_ACTIVATE),
	"/pps_v2.API/DeleteAll":        authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_DELETE_ALL)),

	"/pps_v2.API/CreateSecret":       authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_CREATE_SECRET)),
	"/pps_v2.API/ListSecret":         authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_LIST_SECRETS)),
	"/pps_v2.API/DeleteSecret":       authDisabledOr(clusterPermissions(auth.Permission_SECRET_DELETE)),
	"/pps_v2.API/InspectSecret":      authDisabledOr(clusterPermissions(auth.Permission_SECRET_INSPECT)),
	"/pps_v2.API/RunLoadTest":        authDisabledOr(authenticated),
	"/pps_v2.API/RunLoadTestDefault": authDisabledOr(authenticated),
	"/pps_v2.API/RenderTemplate":     authDisabledOr(authenticated),
	"/pps_v2.API/ListTask":           authDisabledOr(authenticated),
	"/pps_v2.API/GetKubeEvents":      authDisabledOr(authenticated),
	"/pps_v2.API/QueryLoki":          authDisabledOr(authenticated),
	"/pps_v2.API/GetClusterDefaults": authDisabledOr(authenticated),
	"/pps_v2.API/SetClusterDefaults": authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_SET_DEFAULTS)),
	"/pps_v2.API/GetProjectDefaults": authDisabledOr(authenticated),
	"/pps_v2.API/SetProjectDefaults": authDisabledOr(authenticated),
	"/pps_v2.API/PipelinesSummary":   authDisabledOr(authenticated),

	//
	// TransactionAPI
	//

	"/transaction_v2.API/BatchTransaction":   authDisabledOr(authenticated),
	"/transaction_v2.API/StartTransaction":   authDisabledOr(authenticated),
	"/transaction_v2.API/InspectTransaction": authDisabledOr(authenticated),
	"/transaction_v2.API/DeleteTransaction":  authDisabledOr(authenticated),
	"/transaction_v2.API/ListTransaction":    authDisabledOr(authenticated),
	"/transaction_v2.API/FinishTransaction":  authDisabledOr(authenticated),
	"/transaction_v2.API/DeleteAll":          authDisabledOr(clusterPermissions(auth.Permission_CLUSTER_DELETE_ALL)),

	//
	// Version API
	//

	"/versionpb_v2.API/GetVersion": unauthenticated,

	//
	// Proxy API
	//

	// TODO: Only the pachd sidecar instances should be able to use this endpoint.
	"/proxy.API/Listen": unauthenticated,

	//
	// Metadata API
	//
	"/metadata.API/EditMetadata": authDisabledOr(authenticated),
}

// NewInterceptor instantiates a new Interceptor
func NewInterceptor(getAuthServer func() authserver.APIServer) *Interceptor {
	return &Interceptor{
		getAuthServer: getAuthServer,
	}
}

// we use ServerStreamWrapper to set the stream's Context with added values
type ServerStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (s ServerStreamWrapper) Context() context.Context {
	return s.ctx
}

// Interceptor checks the authentication metadata in unary and streaming RPCs
// and prevents unknown or unauthorized calls.
type Interceptor struct {
	getAuthServer func() authserver.APIServer
}

// InterceptUnary applies authentication rules to unary RPCs
func (i *Interceptor) InterceptUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	a, ok := authHandlers[info.FullMethod]
	if !ok {
		log.DPanic(ctx, "no auth function defined")
		return nil, errors.Errorf("no auth function for %q, this is a bug", info.FullMethod)
	}

	username, err := a(pctx.Child(ctx, "checkAuth"), i.getAuthServer(), info.FullMethod)
	if err != nil {
		log.Info(ctx, "denied unary call", zap.Error(err), zap.String("user", username))
		return nil, err
	}

	if username != "" {
		ctx = setWhoAmI(ctx, username)
	}

	return handler(ctx, req)
}

// InterceptStream applies authentication rules to streaming RPCs
func (i *Interceptor) InterceptStream(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()
	a, ok := authHandlers[info.FullMethod]
	if !ok {
		log.DPanic(ctx, "no auth function defined")
		return errors.Errorf("no auth function for %q, this is a bug", info.FullMethod)
	}

	username, err := a(pctx.Child(ctx, "checkAuth"), i.getAuthServer(), info.FullMethod)
	if err != nil {
		log.Info(ctx, "denied streaming call", zap.Error(err), zap.String("user", username))
		return err
	}

	if username != "" {
		newCtx := setWhoAmI(ctx, username)
		stream = ServerStreamWrapper{stream, newCtx}
	}
	return handler(srv, stream)
}
