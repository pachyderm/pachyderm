package auth

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// authFn returns an error if the request is not permitted
type authFn func(i *Interceptor, ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) error

var authFns = map[string]authFn{
	// Admin API
	// Allow InspectCluster to succeed before a user logs in
	"/admin.API/InspectCluster": unauthenticated,

	// Auth API
	// Activate only has an effect when auth is not enabled
	"/auth.API/Activate":                 unauthenticated,
	"/auth.API/Deactivate":               adminOnly,
	"/auth.API/GetConfiguration":         adminOnly,
	"/auth.API/SetConfiguration":         adminOnly,
	"/auth.API/GetAdmins":                adminOnly,
	"/auth.API/ModifyAdmins":             adminOnly,
	"/auth.API/GetClusterRoleBindings":   authenticated,
	"/auth.API/ModifyClusterRoleBinding": adminOnly,
	"/auth.API/Authenticate":             unauthenticated,
	"/auth.API/Authorize":                unauthenticated,
	"/auth.API/WhoAmI":                   unauthenticated,
	// TOOD: split GetScope for self and others
	"/auth.API/GetScope": authenticated,
	"/auth.API/SetScope": authenticated,
	"/auth.API/GetACL":   authenticated,
	"/auth.API/SetACL":   authenticated,
	// GetOIDCLogin is necessary to authenticate
	"/auth.API/GetOIDCLogin": unauthenticated,
	// TOOD: split GetAuthToken for self and others
	"/auth.API/GetAuthToken":    authenticated,
	"/auth.API/ExtendAuthToken": adminOnly,
	// TODO: split RevokeAuthToken for self and others
	"/auth.API/RevokeAuthToken":  authenticated,
	"/auth.API/SetGroupsForUser": adminOnly,
	"/auth.API/ModifyMembers":    adminOnly,
	// TODO: split GetGroups for self and others
	"/auth.API/GetGroups":          authenticated,
	"/auth.API/GetUsers":           adminOnly,
	"/auth.API/GetOneTimePassword": adminOnly,
	"/auth.API/ExtractAuthTokens":  adminOnly,
	"/auth.API/RestoreAuthToken":   adminOnly,

	// Debug API
	"/debug.API/Profile": adminOnly,
	"/debug.API/Binary":  adminOnly,
	"/debug.API/Dump":    adminOnly,

	// Enterprise API
	"/enterprise.API/Activate":          unauthenticated,
	"/enterprise.API/GetState":          unauthenticated,
	"/enterprise.API/GetActivationCode": adminOnly,
	"/enterprise.API/Deactivate":        adminOnly,

	// Health API
	"/health.API/Health": unauthenticated,

	// PFS API
	"/pfs.API/CreateRepo":       authenticated,
	"/pfs.API/InspectRepo":      authenticated,
	"/pfs.API/ListRepo":         authenticated,
	"/pfs.API/DeleteRepo":       authenticated,
	"/pfs.API/StartCommit":      authenticated,
	"/pfs.API/FinishCommit":     authenticated,
	"/pfs.API/InspectCommit":    authenticated,
	"/pfs.API/ListCommit":       authenticated,
	"/pfs.API/ListCommitStream": authenticated,
	"/pfs.API/DeleteCommit":     authenticated,
	"/pfs.API/FlushCommit":      authenticated,
	"/pfs.API/SubscribeCommit":  authenticated,
	"/pfs.API/BuildCommit":      authenticated,
	"/pfs.API/CreateBranch":     authenticated,
	"/pfs.API/InspectBranch":    authenticated,
	"/pfs.API/ListBranch":       authenticated,
	"/pfs.API/DeleteBranch":     authenticated,
	"/pfs.API/PutFile":          authenticated,
	"/pfs.API/CopyFile":         authenticated,
	"/pfs.API/GetFile":          authenticated,
	"/pfs.API/InspectFile":      authenticated,
	"/pfs.API/ListFile":         authenticated,
	"/pfs.API/ListFileStream":   authenticated,
	"/pfs.API/WalkFile":         authenticated,
	"/pfs.API/GlobFile":         authenticated,
	"/pfs.API/GlobFileStream":   authenticated,
	"/pfs.API/DiffFile":         authenticated,
	"/pfs.API/DeleteFile":       authenticated,
	"/pfs.API/DeleteAll":        authenticated,
	"/pfs.API/Fsck":             authenticated,
	"/pfs.API/FileOperationV2":  authenticated,
	"/pfs.API/GetTarV2":         authenticated,
	"/pfs.API/DiffFileV2":       authenticated,
	"/pfs.API/CreateTmpFileSet": authenticated,
	"/pfs.API/RenewTmpFileSet":  authenticated,
	"/pfs.API/ClearCommitV2":    authenticated,

	// Object API - unauthenticated, internal only
	"/pfs.ObjectAPI/PutObject":       unauthenticated,
	"/pfs.ObjectAPI/PutObjectSplit":  unauthenticated,
	"/pfs.ObjectAPI/PutObjects":      unauthenticated,
	"/pfs.ObjectAPI/GetObject":       unauthenticated,
	"/pfs.ObjectAPI/GetObjects":      unauthenticated,
	"/pfs.ObjectAPI/PutBlock":        unauthenticated,
	"/pfs.ObjectAPI/GetBlock":        unauthenticated,
	"/pfs.ObjectAPI/GetBlocks":       unauthenticated,
	"/pfs.ObjectAPI/ListBlock":       unauthenticated,
	"/pfs.ObjectAPI/TagObject":       unauthenticated,
	"/pfs.ObjectAPI/InspectObject":   unauthenticated,
	"/pfs.ObjectAPI/CheckObject":     unauthenticated,
	"/pfs.ObjectAPI/ListObjects":     unauthenticated,
	"/pfs.ObjectAPI/DeleteObjects":   unauthenticated,
	"/pfs.ObjectAPI/GetTag":          unauthenticated,
	"/pfs.ObjectAPI/InspectTag":      unauthenticated,
	"/pfs.ObjectAPI/ListTags":        unauthenticated,
	"/pfs.ObjectAPI/DeleteTags":      unauthenticated,
	"/pfs.ObjectAPI/Compact":         unauthenticated,
	"/pfs.ObjectAPI/PutObjDirect":    unauthenticated,
	"/pfs.ObjectAPI/GetObjDirect":    unauthenticated,
	"/pfs.ObjectAPI/DeleteObjDirect": unauthenticated,

	// PPS API
	// TODO: PPS still checks repo-level auth in the methods themselves
	"/pps.API/CreateJob":       authenticated,
	"/pps.API/InspectJob":      authenticated,
	"/pps.API/ListJob":         authenticated,
	"/pps.API/ListJobStream":   authenticated,
	"/pps.API/FlushJob":        authenticated,
	"/pps.API/DeleteJob":       authenticated,
	"/pps.API/StopJob":         authenticated,
	"/pps.API/InspectDatum":    authenticated,
	"/pps.API/ListDatum":       authenticated,
	"/pps.API/ListDatumStream": authenticated,
	"/pps.API/RestartDatum":    authenticated,
	"/pps.API/CreatePipeline":  authenticated,
	"/pps.API/InspectPipeline": authenticated,
	"/pps.API/ListPipeline":    authenticated,
	"/pps.API/DeletePipeline":  authenticated,
	"/pps.API/StartPipeline":   authenticated,
	"/pps.API/StopPipeline":    authenticated,
	"/pps.API/RunPipeline":     authenticated,
	"/pps.API/RunCron":         authenticated,
	"/pps.API/CreateSecret":    authenticated,
	"/pps.API/DeleteSecret":    authenticated,
	"/pps.API/ListSecret":      authenticated,
	"/pps.API/InspectSecret":   authenticated,
	"/pps.API/DeleteAll":       adminOnly,
	// TODO: split GetLogs into master and not-master and add check for pipeline permissions
	"/pps.API/GetLogs":        authenticated,
	"/pps.API/GarbageCollect": authenticated,
	"/pps.API/ActivateAuth":   authenticated,
	"/pps.API/UpdateJobState": authenticated,

	// TransactionAPI
	"/transaction.API/BatchTransaction":   authenticated,
	"/transaction.API/StartTransaction":   authenticated,
	"/transaction.API/InspectTransaction": authenticated,
	"/transaction.API/DeleteTransaction":  authenticated,
	"/transaction.API/ListTransaction":    authenticated,
	"/transaction.API/FinishTransaction":  authenticated,
	"/transaction.API/DeleteAll":          adminOnly,

	// Version API
	"/versionpb.API/GetVersion": authenticated,
}

// unauthenticated allows all RPCs to succeed, even if there is no auth metadata
func unauthenticated(i *Interceptor, ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) error {
	return nil
}

// authenticated allows all RPCs to succeed as long as the user has a valid auth token
func authenticated(i *Interceptor, ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) error {
	pachClient := i.env.GetPachClient(ctx)
	_, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{})
	if err != nil && !auth.IsErrNotActivated(err) {
		return err
	}
	return nil
}

// adminOnly allows an RPC to succeed only if the user has cluster admin status
func adminOnly(i *Interceptor, ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) error {
	pachClient := i.env.GetPachClient(ctx)
	me, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{})
	if err != nil {
		if auth.IsErrNotActivated(err) {
			return nil
		}
		return err
	}

	if me.ClusterRoles != nil {
		for _, s := range me.ClusterRoles.Roles {
			if s == auth.ClusterRole_SUPER {
				return nil
			}
		}
	}

	return &auth.ErrNotAuthorized{
		Subject: me.Username,
		AdminOp: info.FullMethod,
	}
}

// Interceptor applies auth rules to incoming RPCs
type Interceptor struct {
	env *serviceenv.ServiceEnv
}

// NewInterceptor instantiates a new authInterceptor
func NewInterceptor(env *serviceenv.ServiceEnv) *Interceptor {
	return &Interceptor{
		env: env,
	}
}

// InterceptUnary applies authentication rules to unary RPCs
func (i *Interceptor) InterceptUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	a, ok := authFns[info.FullMethod]
	if !ok {
		logrus.Errorf("no auth function for %q\n", info.FullMethod)
		return nil, fmt.Errorf("no auth function for %q, this is a bug", info.FullMethod)
	}

	if err := a(i, ctx, info, req); err != nil {
		logrus.Errorf("denied unary call %q\n", info.FullMethod)
		return nil, err
	}

	logrus.Debugf("allowed unary call %q\n", info.FullMethod)
	return handler(ctx, req)
}
