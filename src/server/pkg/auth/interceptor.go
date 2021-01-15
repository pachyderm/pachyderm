package auth

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// authHandlers is a mapping of RPCs to authorization levels required to access them.
// This interceptor fails closed - whenever a new RPC is added, it's disabled
// until some authentication is added. Some RPCs may do additional auth checks
// beyond what's required by this interceptor.
var authHandlers = map[string]authHandlerFn{
	//
	// Admin API
	//

	// Allow InspectCluster to succeed before a user logs in
	"/admin.API/InspectCluster": unauthenticated,

	//
	// Auth API
	//

	// Activate only has an effect when auth is not enabled
	// Authenticate, Authorize and WhoAmI check auth status themselves
	// GetOIDCLogin is necessary to authenticate
	"/auth.API/Activate":     unauthenticated,
	"/auth.API/Authenticate": unauthenticated,
	"/auth.API/Authorize":    unauthenticated,
	"/auth.API/WhoAmI":       unauthenticated,
	"/auth.API/GetOIDCLogin": unauthenticated,

	// TODO: restrict GetClusterRoleBindings to cluster admins?
	// TODO: split GetScope for self and others
	// TODO: split GetAuthToken for self and others
	// TODO: split RevokeAuthToken for self and others
	// TODO: split GetGroups for self and others
	"/auth.API/GetAdmins":              requireAuthEnabled(authenticated),
	"/auth.API/GetClusterRoleBindings": requireAuthEnabled(authenticated),
	"/auth.API/GetConfiguration":       requireAuthEnabled(authenticated),
	"/auth.API/GetScope":               requireAuthEnabled(authenticated),
	"/auth.API/SetScope":               requireAuthEnabled(authenticated),
	"/auth.API/GetACL":                 requireAuthEnabled(authenticated),
	"/auth.API/SetACL":                 requireAuthEnabled(authenticated),
	"/auth.API/GetAuthToken":           requireAuthEnabled(authenticated),
	"/auth.API/RevokeAuthToken":        requireAuthEnabled(authenticated),
	"/auth.API/GetGroups":              requireAuthEnabled(authenticated),
	"/auth.API/GetOneTimePassword":     requireAuthEnabled(authenticated),

	// Deactivate can be called when the cluster is partially activated,
	// but the rest of the API is prohibited
	"/auth.API/Deactivate":               adminOnly,
	"/auth.API/SetConfiguration":         requireAuthEnabled(adminOnly),
	"/auth.API/ModifyAdmins":             requireAuthEnabled(adminOnly),
	"/auth.API/ModifyClusterRoleBinding": requireAuthEnabled(adminOnly),
	"/auth.API/ExtendAuthToken":          requireAuthEnabled(adminOnly),
	"/auth.API/SetGroupsForUser":         requireAuthEnabled(adminOnly),
	"/auth.API/ModifyMembers":            requireAuthEnabled(adminOnly),
	"/auth.API/GetUsers":                 requireAuthEnabled(adminOnly),
	"/auth.API/ExtractAuthTokens":        requireAuthEnabled(adminOnly),
	"/auth.API/RestoreAuthToken":         requireAuthEnabled(adminOnly),

	//
	// Debug API
	//

	"/debug.Debug/Profile": adminOnly,
	"/debug.Debug/Binary":  adminOnly,
	"/debug.Debug/Dump":    adminOnly,

	//
	// Enterprise API
	//

	"/enterprise.API/Activate":          unauthenticated,
	"/enterprise.API/GetState":          unauthenticated,
	"/enterprise.API/GetActivationCode": adminOnly,
	"/enterprise.API/Deactivate":        adminOnly,

	//
	// Health API
	//
	"/health.Health/Health": unauthenticated,

	//
	// Identity API
	//
	"/identity.API/SetIdentityServerConfig": requireAuthEnabled(adminOnly),
	"/identity.API/GetIdentityServerConfig": requireAuthEnabled(adminOnly),
	"/identity.API/CreateIDPConnector":      requireAuthEnabled(adminOnly),
	"/identity.API/UpdateIDPConnector":      requireAuthEnabled(adminOnly),
	"/identity.API/ListIDPConnectors":       requireAuthEnabled(adminOnly),
	"/identity.API/GetIDPConnector":         requireAuthEnabled(adminOnly),
	"/identity.API/DeleteIDPConnector":      requireAuthEnabled(adminOnly),
	"/identity.API/CreateOIDCClient":        requireAuthEnabled(adminOnly),
	"/identity.API/UpdateOIDCClient":        requireAuthEnabled(adminOnly),
	"/identity.API/GetOIDCClient":           requireAuthEnabled(adminOnly),
	"/identity.API/ListOIDCClients":         requireAuthEnabled(adminOnly),
	"/identity.API/DeleteOIDCClient":        requireAuthEnabled(adminOnly),
	"/identity.API/DeleteAll":               requireAuthEnabled(adminOnly),

	//
	// PFS API
	//

	// TODO: Add methods to handle repo permissions
	"/pfs.API/CreateRepo":      authenticated,
	"/pfs.API/InspectRepo":     authenticated,
	"/pfs.API/ListRepo":        authenticated,
	"/pfs.API/DeleteRepo":      authenticated,
	"/pfs.API/StartCommit":     authenticated,
	"/pfs.API/FinishCommit":    authenticated,
	"/pfs.API/InspectCommit":   authenticated,
	"/pfs.API/ListCommit":      authenticated,
	"/pfs.API/DeleteCommit":    authenticated,
	"/pfs.API/FlushCommit":     authenticated,
	"/pfs.API/SubscribeCommit": authenticated,
	"/pfs.API/ClearCommit":     authenticated,
	"/pfs.API/CreateBranch":    authenticated,
	"/pfs.API/InspectBranch":   authenticated,
	"/pfs.API/ListBranch":      authenticated,
	"/pfs.API/DeleteBranch":    authenticated,
	"/pfs.API/ModifyFile":      authenticated,
	"/pfs.API/CopyFile":        authenticated,
	"/pfs.API/GetFile":         authenticated,
	"/pfs.API/InspectFile":     authenticated,
	"/pfs.API/ListFile":        authenticated,
	"/pfs.API/WalkFile":        authenticated,
	"/pfs.API/GlobFile":        authenticated,
	"/pfs.API/DiffFile":        authenticated,
	"/pfs.API/DeleteAll":       authenticated,
	"/pfs.API/Fsck":            authenticated,
	"/pfs.API/CreateFileset":   authenticated,
	"/pfs.API/RenewFileset":    authenticated,

	//
	// Object API
	//

	// Object API is unauthenticated and only for internal use
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

	//
	// PPS API
	//

	// TODO: Add per-repo permissions checks for these
	// TODO: split GetLogs into master and not-master and add check for pipeline permissions
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
	"/pps.API/GetLogs":         authenticated,
	"/pps.API/GarbageCollect":  authenticated,
	"/pps.API/ActivateAuth":    authenticated,
	"/pps.API/UpdateJobState":  authenticated,
	"/pps.API/DeleteAll":       adminOnly,

	//
	// TransactionAPI
	//

	"/transaction.API/BatchTransaction":   authenticated,
	"/transaction.API/StartTransaction":   authenticated,
	"/transaction.API/InspectTransaction": authenticated,
	"/transaction.API/DeleteTransaction":  authenticated,
	"/transaction.API/ListTransaction":    authenticated,
	"/transaction.API/FinishTransaction":  authenticated,
	"/transaction.API/DeleteAll":          adminOnly,

	//
	// Version API
	//

	"/versionpb.API/GetVersion": authenticated,
}

// NewInterceptor instantiates a new Interceptor
func NewInterceptor(env *serviceenv.ServiceEnv) *Interceptor {
	return &Interceptor{
		env: env,
	}
}

// Interceptor checks the authentication metadata in unary and streaming RPCs
// and prevents unknown or unauthorized calls.
type Interceptor struct {
	env *serviceenv.ServiceEnv
}

// InterceptUnary applies authentication rules to unary RPCs
func (i *Interceptor) InterceptUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	a, ok := authHandlers[info.FullMethod]
	if !ok {
		logrus.Errorf("no auth function for %q\n", info.FullMethod)
		return nil, fmt.Errorf("no auth function for %q, this is a bug", info.FullMethod)
	}

	if err := a(i).unary(ctx, info, req); err != nil {
		logrus.Errorf("denied unary call %q\n", info.FullMethod)
		return nil, err
	}
	return handler(ctx, req)
}

// InterceptStream applies authentication rules to streaming RPCs
func (i *Interceptor) InterceptStream(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	a, ok := authHandlers[info.FullMethod]
	if !ok {
		logrus.Errorf("no auth function for %q\n", info.FullMethod)
		return fmt.Errorf("no auth function for %q, this is a bug", info.FullMethod)
	}

	if err := a(i).stream(stream.Context(), info, nil); err != nil {
		logrus.Errorf("denied streaming call %q\n", info.FullMethod)
		return err
	}
	return handler(srv, stream)
}
