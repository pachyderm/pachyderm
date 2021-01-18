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
	"/auth.API/GetAuthToken": unauthenticated,
	"/auth.API/Deactivate":   unauthenticated,

	// TODO: restrict GetClusterRoleBindings to cluster admins?
	// TODO: split GetScope for self and others
	// TODO: split GetAuthToken for self and others
	// TODO: split RevokeAuthToken for self and others
	// TODO: split GetGroups for self and others
	"/auth.API/GetAdmins":              authenticated,
	"/auth.API/GetClusterRoleBindings": authenticated,
	"/auth.API/GetConfiguration":       authenticated,
	"/auth.API/GetScope":               authenticated,
	"/auth.API/SetScope":               authenticated,
	"/auth.API/GetACL":                 authenticated,
	"/auth.API/SetACL":                 authenticated,
	"/auth.API/RevokeAuthToken":        authenticated,
	"/auth.API/GetGroups":              authenticated,
	"/auth.API/GetOneTimePassword":     authenticated,

	// Deactivate can be called when the cluster is partially activated,
	// but the rest of the API is prohibited
	"/auth.API/SetConfiguration":         admin,
	"/auth.API/ModifyAdmins":             admin,
	"/auth.API/ModifyClusterRoleBinding": admin,
	"/auth.API/ExtendAuthToken":          admin,
	"/auth.API/SetGroupsForUser":         admin,
	"/auth.API/ModifyMembers":            admin,
	"/auth.API/GetUsers":                 admin,
	"/auth.API/ExtractAuthTokens":        admin,
	"/auth.API/RestoreAuthToken":         admin,

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
	// PFS API
	//

	// TODO: Add methods to handle repo permissions
	"/pfs.API/CreateRepo":      authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/InspectRepo":     authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/ListRepo":        authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/DeleteRepo":      authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/StartCommit":     authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/FinishCommit":    authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/InspectCommit":   authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/ListCommit":      authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/DeleteCommit":    authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/FlushCommit":     authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/SubscribeCommit": authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/ClearCommit":     authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/CreateBranch":    authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/InspectBranch":   authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/ListBranch":      authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/DeleteBranch":    authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/ModifyFile":      authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/CopyFile":        authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/GetFile":         authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/InspectFile":     authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/ListFile":        authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/WalkFile":        authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/GlobFile":        authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/DiffFile":        authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/DeleteAll":       authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/Fsck":            authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/CreateFileset":   authDisabledOr(authPartialAnd(authenticated)),
	"/pfs.API/RenewFileset":    authDisabledOr(authPartialAnd(authenticated)),

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
	"/pps.API/CreateJob":       authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/InspectJob":      authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/ListJob":         authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/ListJobStream":   authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/FlushJob":        authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/DeleteJob":       authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/StopJob":         authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/InspectDatum":    authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/ListDatum":       authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/ListDatumStream": authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/RestartDatum":    authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/CreatePipeline":  authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/InspectPipeline": authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/DeletePipeline":  authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/StartPipeline":   authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/StopPipeline":    authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/RunPipeline":     authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/RunCron":         authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/CreateSecret":    authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/DeleteSecret":    authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/ListSecret":      authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/InspectSecret":   authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/GetLogs":         authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/GarbageCollect":  authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/UpdateJobState":  authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/ListPipeline":    authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/ActivateAuth":    authDisabledOr(authPartialAnd(authenticated)),
	"/pps.API/DeleteAll":       authDisabledOr(authPartialAnd(admin)),

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

	"/versionpb.API/GetVersion": authDisabledOr(authenticated),
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
