package auth

import (
	"context"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"

	"github.com/sirupsen/logrus"
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
	"/admin.API/InspectCluster": unauthenticated,

	//
	// Auth API
	//

	// Activate only has an effect when auth is not enabled
	// Authenticate, Authorize and WhoAmI check auth status themselves
	// GetOIDCLogin is necessary to authenticate
	// GetAuthToken and Deactivate are necessary in the partial activation state
	// and have their own auth logic
	"/auth.API/Activate":     unauthenticated,
	"/auth.API/Authenticate": unauthenticated,
	"/auth.API/Authorize":    unauthenticated,
	"/auth.API/WhoAmI":       unauthenticated,
	"/auth.API/GetOIDCLogin": unauthenticated,

	// TODO: restrict GetClusterRoleBinding to cluster admins?
	// TODO: split GetScope for self and others
	// TODO: split GetAuthToken for self and others
	// TODO: split RevokeAuthToken for self and others
	// TODO: split GetGroups for self and others
	"/auth.API/GetConfiguration":  authenticated,
	"/auth.API/CreateRoleBinding": authenticated,
	"/auth.API/GetRoleBinding":    authenticated,
	"/auth.API/ModifyRoleBinding": authenticated,
	"/auth.API/DeleteRoleBinding": authenticated,
	"/auth.API/RevokeAuthToken":   authenticated,
	"/auth.API/GetGroups":         authenticated,
	"/auth.API/GetAuthToken":      admin,
	"/auth.API/SetConfiguration":  admin,
	"/auth.API/ExtendAuthToken":   admin,
	"/auth.API/SetGroupsForUser":  admin,
	"/auth.API/ModifyMembers":     admin,
	"/auth.API/GetUsers":          admin,
	"/auth.API/ExtractAuthTokens": admin,
	"/auth.API/RestoreAuthToken":  admin,
	"/auth.API/Deactivate":        admin,

	//
	// Debug API
	//

	"/debug.Debug/Profile": authDisabledOr(admin),
	"/debug.Debug/Binary":  authDisabledOr(admin),
	"/debug.Debug/Dump":    authDisabledOr(admin),

	//
	// Enterprise API
	//

	"/enterprise.API/Activate":          unauthenticated,
	"/enterprise.API/GetState":          unauthenticated,
	"/enterprise.API/GetActivationCode": authDisabledOr(admin),
	"/enterprise.API/Deactivate":        authDisabledOr(admin),
	"/enterprise.API/Heartbeat":         authDisabledOr(admin),

	//
	// Health API
	//
	"/health.Health/Health": unauthenticated,

	//
	// Identity API
	//
	"/identity.API/SetIdentityServerConfig": admin,
	"/identity.API/GetIdentityServerConfig": admin,
	"/identity.API/CreateIDPConnector":      admin,
	"/identity.API/UpdateIDPConnector":      admin,
	"/identity.API/ListIDPConnectors":       admin,
	"/identity.API/GetIDPConnector":         admin,
	"/identity.API/DeleteIDPConnector":      admin,
	"/identity.API/CreateOIDCClient":        admin,
	"/identity.API/UpdateOIDCClient":        admin,
	"/identity.API/GetOIDCClient":           admin,
	"/identity.API/ListOIDCClients":         admin,
	"/identity.API/DeleteOIDCClient":        admin,
	"/identity.API/DeleteAll":               admin,

	//
	// License API
	//
	"/license.API/Activate":          authDisabledOr(admin),
	"/license.API/GetActivationCode": authDisabledOr(admin),
	"/license.API/Deactivate":        authDisabledOr(admin),
	"/license.API/AddCluster":        authDisabledOr(admin),
	"/license.API/UpdateCluster":     authDisabledOr(admin),
	"/license.API/DeleteCluster":     authDisabledOr(admin),
	"/license.API/ListClusters":      authDisabledOr(admin),
	"/license.API/DeleteAll":         authDisabledOr(admin),
	// Heartbeat relies on the shared secret generated at cluster registration-time
	"/license.API/Heartbeat": unauthenticated,

	//
	// PFS API
	//

	// TODO: Add methods to handle repo permissions
	"/pfs.API/ActivateAuth":    admin,
	"/pfs.API/CreateRepo":      authDisabledOr(authenticated),
	"/pfs.API/InspectRepo":     authDisabledOr(authenticated),
	"/pfs.API/ListRepo":        authDisabledOr(authenticated),
	"/pfs.API/DeleteRepo":      authDisabledOr(authenticated),
	"/pfs.API/StartCommit":     authDisabledOr(authenticated),
	"/pfs.API/FinishCommit":    authDisabledOr(authenticated),
	"/pfs.API/InspectCommit":   authDisabledOr(authenticated),
	"/pfs.API/ListCommit":      authDisabledOr(authenticated),
	"/pfs.API/SquashCommit":    authDisabledOr(authenticated),
	"/pfs.API/FlushCommit":     authDisabledOr(authenticated),
	"/pfs.API/SubscribeCommit": authDisabledOr(authenticated),
	"/pfs.API/ClearCommit":     authDisabledOr(authenticated),
	"/pfs.API/CreateBranch":    authDisabledOr(authenticated),
	"/pfs.API/InspectBranch":   authDisabledOr(authenticated),
	"/pfs.API/ListBranch":      authDisabledOr(authenticated),
	"/pfs.API/DeleteBranch":    authDisabledOr(authenticated),
	"/pfs.API/ModifyFile":      authDisabledOr(authenticated),
	"/pfs.API/CopyFile":        authDisabledOr(authenticated),
	"/pfs.API/GetFile":         authDisabledOr(authenticated),
	"/pfs.API/InspectFile":     authDisabledOr(authenticated),
	"/pfs.API/ListFile":        authDisabledOr(authenticated),
	"/pfs.API/WalkFile":        authDisabledOr(authenticated),
	"/pfs.API/GlobFile":        authDisabledOr(authenticated),
	"/pfs.API/DiffFile":        authDisabledOr(authenticated),
	"/pfs.API/DeleteAll":       authDisabledOr(authenticated),
	"/pfs.API/Fsck":            authDisabledOr(authenticated),
	"/pfs.API/CreateFileset":   authDisabledOr(authenticated),
	"/pfs.API/RenewFileset":    authDisabledOr(authenticated),

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
	"/pps.API/CreateJob":       authDisabledOr(authenticated),
	"/pps.API/InspectJob":      authDisabledOr(authenticated),
	"/pps.API/ListJob":         authDisabledOr(authenticated),
	"/pps.API/ListJobStream":   authDisabledOr(authenticated),
	"/pps.API/FlushJob":        authDisabledOr(authenticated),
	"/pps.API/DeleteJob":       authDisabledOr(authenticated),
	"/pps.API/StopJob":         authDisabledOr(authenticated),
	"/pps.API/InspectDatum":    authDisabledOr(authenticated),
	"/pps.API/ListDatum":       authDisabledOr(authenticated),
	"/pps.API/ListDatumStream": authDisabledOr(authenticated),
	"/pps.API/RestartDatum":    authDisabledOr(authenticated),
	"/pps.API/CreatePipeline":  authDisabledOr(authenticated),
	"/pps.API/InspectPipeline": authDisabledOr(authenticated),
	"/pps.API/DeletePipeline":  authDisabledOr(authenticated),
	"/pps.API/StartPipeline":   authDisabledOr(authenticated),
	"/pps.API/StopPipeline":    authDisabledOr(authenticated),
	"/pps.API/RunPipeline":     authDisabledOr(authenticated),
	"/pps.API/RunCron":         authDisabledOr(authenticated),
	"/pps.API/CreateSecret":    authDisabledOr(authenticated),
	"/pps.API/DeleteSecret":    authDisabledOr(authenticated),
	"/pps.API/ListSecret":      authDisabledOr(authenticated),
	"/pps.API/InspectSecret":   authDisabledOr(authenticated),
	"/pps.API/GetLogs":         authDisabledOr(authenticated),
	"/pps.API/GarbageCollect":  authDisabledOr(authenticated),
	"/pps.API/UpdateJobState":  authDisabledOr(authenticated),
	"/pps.API/ListPipeline":    authDisabledOr(authenticated),
	"/pps.API/ActivateAuth":    admin,
	"/pps.API/DeleteAll":       authDisabledOr(admin),

	//
	// TransactionAPI
	//

	"/transaction.API/BatchTransaction":   authDisabledOr(authenticated),
	"/transaction.API/StartTransaction":   authDisabledOr(authenticated),
	"/transaction.API/InspectTransaction": authDisabledOr(authenticated),
	"/transaction.API/DeleteTransaction":  authDisabledOr(authenticated),
	"/transaction.API/ListTransaction":    authDisabledOr(authenticated),
	"/transaction.API/FinishTransaction":  authDisabledOr(authenticated),
	"/transaction.API/DeleteAll":          authDisabledOr(admin),

	//
	// Version API
	//

	"/versionpb.API/GetVersion": unauthenticated,
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
	pachClient := i.env.GetPachClient(ctx)
	a, ok := authHandlers[info.FullMethod]
	if !ok {
		logrus.Errorf("no auth function for %q\n", info.FullMethod)
		return nil, fmt.Errorf("no auth function for %q, this is a bug", info.FullMethod)
	}

	if err := a(pachClient, info.FullMethod); err != nil {
		logrus.Errorf("denied unary call %q\n", info.FullMethod)
		return nil, err
	}
	return handler(ctx, req)
}

// InterceptStream applies authentication rules to streaming RPCs
func (i *Interceptor) InterceptStream(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	pachClient := i.env.GetPachClient(stream.Context())
	a, ok := authHandlers[info.FullMethod]
	if !ok {
		logrus.Errorf("no auth function for %q\n", info.FullMethod)
		return fmt.Errorf("no auth function for %q, this is a bug", info.FullMethod)
	}

	if err := a(pachClient, info.FullMethod); err != nil {
		logrus.Errorf("denied streaming call %q\n", info.FullMethod)
		return err
	}
	return handler(srv, stream)
}
