package server

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	enterpriseclient "github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/keycache"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsconsts"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/version"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	allClusterUsersSubject = "allClusterUsers"

	tokensPrefix       = "/tokens"
	roleBindingsPrefix = "/role-bindings"
	membersPrefix      = "/members"
	groupsPrefix       = "/groups"
	configPrefix       = "/config"
	oidcAuthnPrefix    = "/oidc-authns"

	// defaultSessionTTLSecs is the lifetime of an auth token from Authenticate,
	// and the default lifetime of an auth token from GetAuthToken.
	//
	// Note: if 'defaultSessionTTLSecs' is changed, then the description of
	// '--ttl' in 'pachctl get-auth-token' must also be changed
	defaultSessionTTLSecs = 30 * 24 * 60 * 60 // 30 days

	// ppsUser is a special, unrevokable cluster administrator account used by PPS
	// to create pipeline tokens, close commits, and do other necessary PPS work.
	// It's not possible to authenticate as ppsUser (pps reads the auth token for
	// this user directly from etcd). This string is not secret, but is long and
	// random to avoid collisions with real usernames
	ppsUser = `magic:GZD4jKDGcirJyWQt6HtK4hhRD6faOofP1mng34xNZsI`

	// configKey is a key (in etcd, in the config collection) that maps to the
	// auth configuration. This is the only key in that collection (due to
	// implemenation details of our config library, we can't use an empty key)
	configKey = "config"

	// clusterRoleBindingKey is a key in etcd, in the roleBindings collection,
	// that contains the set of role bindings for the cluster. These are frequently
	// accessed so we cache them.
	clusterRoleBindingKey = "CLUSTER:"

	// GitHookPort is 655
	// Prometheus uses 656

	// OidcPort is the port where OIDC ID Providers can send auth assertions
	OidcPort = 657
)

// DefaultOIDCConfig is the default config for the auth API server
var DefaultOIDCConfig = auth.OIDCConfig{}

// epsilon is small, nonempty protobuf to use as an etcd value (the etcd client
// library can't distinguish between empty values and missing values, even
// though empty values are still stored in etcd)
var epsilon = &types.BoolValue{Value: true}

// APIServer represents an auth api server
type APIServer interface {
	auth.APIServer
	txnenv.AuthTransactionServer
}

// apiServer implements the public interface of the Pachyderm auth system,
// including all RPCs defined in the protobuf spec.
type apiServer struct {
	env        *serviceenv.ServiceEnv
	txnEnv     *txnenv.TransactionEnv
	pachLogger log.Logger

	configCache             *keycache.Cache
	clusterRoleBindingCache *keycache.Cache

	// tokens is a collection of hashedToken -> TokenInfo mappings. These tokens are
	// returned to users by Authenticate()
	tokens col.Collection
	// roleBindings is a collection of resource name -> role binding mappings.
	roleBindings col.Collection
	// members is a collection of username -> groups mappings.
	members col.Collection
	// groups is a collection of group -> usernames mappings.
	groups col.Collection
	// collection containing the auth config (under the key configKey)
	authConfig col.Collection
	// oidcStates  contains the set of OIDC nonces for requests that are in progress
	oidcStates col.Collection

	// This is a cache of the PPS master token. It's set once on startup and then
	// never updated
	ppsToken string

	// public addresses the fact that pachd in full mode initializes two auth
	// servers: one that exposes a public API, possibly over TLS, and one that
	// exposes a private API, for internal services. Only the public-facing auth
	// service should export the SAML ACS and Metadata services, so if public
	// is true and auth is active, this may export those SAML services
	public bool
}

// LogReq is like log.Logger.Log(), but it assumes that it's being called from
// the top level of a GRPC method implementation, and correspondingly extracts
// the method name from the parent stack frame
func (a *apiServer) LogReq(request interface{}) {
	a.pachLogger.Log(request, nil, nil, 0)
}

// LogResp is like log.Logger.Log(). However,
// 1) It assumes that it's being called from a defer() statement in a GRPC
//    method , and correspondingly extracts the method name from the grandparent
//    stack frame
// 2) It logs NotActivatedError at DebugLevel instead of ErrorLevel, as, in most
//    cases, this error is expected, and logging it frequently may confuse users
func (a *apiServer) LogResp(request interface{}, response interface{}, err error, duration time.Duration) {
	if err == nil {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.InfoLevel, 4)
	} else if auth.IsErrNotActivated(err) {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.DebugLevel, 4)
	} else {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.ErrorLevel, 4)
	}
}

// NewAuthServer returns an implementation of auth.APIServer.
func NewAuthServer(
	env *serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	etcdPrefix string,
	public bool,
	requireNoncriticalServers bool,
) (APIServer, error) {

	authConfig := col.NewCollection(
		env.GetEtcdClient(),
		path.Join(etcdPrefix, configKey),
		nil,
		&auth.OIDCConfig{},
		nil,
		nil,
	)
	roleBindings := col.NewCollection(
		env.GetEtcdClient(),
		path.Join(etcdPrefix, roleBindingsPrefix),
		nil,
		&auth.RoleBinding{},
		nil,
		nil,
	)
	s := &apiServer{
		env:        env,
		txnEnv:     txnEnv,
		pachLogger: log.NewLogger("auth.API"),
		tokens: col.NewCollection(
			env.GetEtcdClient(),
			path.Join(etcdPrefix, tokensPrefix),
			nil,
			&auth.TokenInfo{},
			nil,
			nil,
		),
		members: col.NewCollection(
			env.GetEtcdClient(),
			path.Join(etcdPrefix, membersPrefix),
			nil,
			&auth.Groups{},
			nil,
			nil,
		),
		groups: col.NewCollection(
			env.GetEtcdClient(),
			path.Join(etcdPrefix, groupsPrefix),
			nil,
			&auth.Users{},
			nil,
			nil,
		),
		oidcStates: col.NewCollection(
			env.GetEtcdClient(),
			path.Join(oidcAuthnPrefix),
			nil,
			&auth.SessionInfo{},
			nil,
			nil,
		),
		authConfig:              authConfig,
		roleBindings:            roleBindings,
		configCache:             keycache.NewCache(authConfig, configKey, &DefaultOIDCConfig),
		clusterRoleBindingCache: keycache.NewCache(authConfig, configKey, &auth.RoleBinding{}),
		public:                  public,
	}
	go s.retrieveOrGeneratePPSToken()

	if public {
		// start OIDC service (won't respond to anything until config is set)
		go waitForError("OIDC HTTP Server", requireNoncriticalServers, s.serveOIDC)
	}

	// Watch for new auth config options
	go s.configCache.Watch()

	// Watch for changes to the cluster role binding
	go s.clusterRoleBindingCache.Watch()
	return s, nil
}

func waitForError(name string, required bool, cb func() error) {
	if err := cb(); !errors.Is(err, http.ErrServerClosed) {
		if required {
			logrus.Fatalf("error setting up and/or running %v (use --require-critical-servers-only deploy flag to ignore errors from noncritical servers): %v", name, err)
		}
		logrus.Errorf("error setting up and/or running %v: %v", name, err)
	}
}

// isActive returns an error when auth is not enabled. If there are no cluster role bindings auth has not been enabled.
func (a *apiServer) isActive() error {
	bindings, ok := a.clusterRoleBindingCache.Load().(*auth.RoleBinding)
	if !ok {
		return errors.New("cached auth config had unexpected type")
	}

	if bindings.Entries == nil {
		return auth.ErrNotActivated
	}
	return nil
}

// Retrieve the PPS master token, or generate it and put it in etcd.
// TODO This is a hack. It avoids the need to return superuser tokens from
// GetAuthToken (essentially, PPS and Auth communicate through etcd instead of
// an API) but we should define an internal API and use that instead.
func (a *apiServer) retrieveOrGeneratePPSToken() {
	var tokenProto types.StringValue // will contain PPS token
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 60 * time.Second
	b.MaxInterval = 5 * time.Second
	if err := backoff.Retry(func() error {
		if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
			superUserTokenCol := col.NewCollection(a.env.GetEtcdClient(), ppsconsts.PPSTokenKey, nil, &types.StringValue{}, nil, nil).ReadWrite(stm)
			// TODO(msteffen): Don't use an empty key, as it will not be erased by
			// superUserTokenCol.DeleteAll()
			err := superUserTokenCol.Get("", &tokenProto)
			if err == nil {
				return nil
			}
			if col.IsErrNotFound(err) {
				// no existing token yet -- generate token
				token := uuid.NewWithoutDashes()
				tokenProto.Value = token
				if err := superUserTokenCol.Create("", &tokenProto); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		a.ppsToken = tokenProto.Value
		return nil
	}, b); err != nil {
		panic(fmt.Sprintf("couldn't create/retrieve PPS superuser token within 60s of starting up: %v", err))
	}
}

func (a *apiServer) getEnterpriseTokenState(ctx context.Context) (enterpriseclient.State, error) {
	pachClient := a.env.GetPachClient(ctx)
	resp, err := pachClient.Enterprise.GetState(pachClient.Ctx(),
		&enterpriseclient.GetStateRequest{})
	if err != nil {
		return 0, errors.Wrapf(grpcutil.ScrubGRPC(err), "could not get Enterprise status")
	}
	return resp.State, nil
}

// Activate implements the protobuf auth.Activate RPC
func (a *apiServer) Activate(ctx context.Context, req *auth.ActivateRequest) (resp *auth.ActivateResponse, retErr error) {
	pachClient := a.env.GetPachClient(ctx)
	ctx = pachClient.Ctx() // copy auth information
	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.LogResp(nil, nil, retErr, time.Since(start)) }(time.Now())
	// If the cluster's Pachyderm Enterprise token isn't active, the auth system
	// cannot be activated
	state, err := a.getEnterpriseTokenState(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "error confirming Pachyderm Enterprise token")
	}
	if state != enterpriseclient.State_ACTIVE {
		return nil, errors.Errorf("Pachyderm Enterprise is not active in this " +
			"cluster, and the Pachyderm auth API is an Enterprise-level feature")
	}

	// Activating an already activated auth service should fail, because
	// otherwise anyone can just activate the service again and set
	// themselves as an admin. If activation failed in PFS, calling auth.Activate
	// again should work (in this state, the only admin will be 'ppsUser')
	if err := a.isActive(); err != nil {
		return nil, auth.ErrAlreadyActivated
	}

	// If the token hash was in the request, use it and return an empty response.
	// Otherwise generate a new random token.
	pachToken := req.RootToken
	if pachToken == "" {
		pachToken = uuid.NewWithoutDashes()
	}

	// Store a new Pachyderm token (as the caller is authenticating) and
	// initialize the root user as a cluster admin
	if _, err = col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		admins := a.roleBindings.ReadWrite(stm)
		tokens := a.tokens.ReadWrite(stm)
		if err := admins.Put(clusterRoleBindingKey, &auth.RoleBinding{
			Entries: map[string]*auth.Roles{
				auth.RootUser: &auth.Roles{Roles: map[string]bool{auth.ClusterAdminRole: true}},
			},
		}); err != nil {
			return err
		}
		return tokens.Put(
			auth.HashToken(pachToken),
			&auth.TokenInfo{
				Subject: auth.RootUser,
				Source:  auth.TokenInfo_AUTHENTICATE,
			},
		)
	}); err != nil {
		return nil, err
	}

	// wait until the clusterRoleBinding watcher has updated the local cache
	// (changing the activation state), so that Activate() is less likely to
	// race with subsequent calls that expect auth to be activated.
	if err := backoff.Retry(func() error {
		if err := a.isActive(); err != nil {
			return errors.Errorf("auth never activated")
		}
		return nil
	}, backoff.RetryEvery(100*time.Millisecond)); err != nil {
		return nil, err
	}
	return &auth.ActivateResponse{PachToken: pachToken}, nil
}

// Deactivate implements the protobuf auth.Deactivate RPC
func (a *apiServer) Deactivate(ctx context.Context, req *auth.DeactivateRequest) (resp *auth.DeactivateResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		a.roleBindings.ReadWrite(stm).DeleteAll()
		a.tokens.ReadWrite(stm).DeleteAll()
		a.members.ReadWrite(stm).DeleteAll()
		a.groups.ReadWrite(stm).DeleteAll()
		a.authConfig.ReadWrite(stm).DeleteAll()
		return nil
	})
	if err != nil {
		return nil, err
	}

	// wait until the clusterRoleBinding watcher sees the deleted role binding,
	// so that Deactivate() is less likely to race with subsequent calls that
	// expect auth to be deactivated.
	if err := backoff.Retry(func() error {
		if err := a.isActive(); err == nil {
			return errors.Errorf("auth still activated")
		}
		return nil
	}, backoff.RetryEvery(100*time.Millisecond)); err != nil {
		return nil, err
	}
	return &auth.DeactivateResponse{}, nil
}

// expiredEnterpriseCheck enforces that if the cluster's enterprise token is
// expired, users cannot log in. The root token can be used to access the cluster.
func (a *apiServer) expiredEnterpriseCheck(ctx context.Context) error {
	state, err := a.getEnterpriseTokenState(ctx)
	if err != nil {
		return errors.Wrapf(err, "error confirming Pachyderm Enterprise token")
	}

	if state != enterpriseclient.State_ACTIVE {
		return errors.New("Pachyderm Enterprise is not active in this " +
			"cluster (until Pachyderm Enterprise is re-activated or Pachyderm " +
			"auth is deactivated, users cannot log in)")
	}
	return nil
}

// Authenticate implements the protobuf auth.Authenticate RPC
func (a *apiServer) Authenticate(ctx context.Context, req *auth.AuthenticateRequest) (resp *auth.AuthenticateResponse, retErr error) {
	if err := a.isActive(); err != nil {
		return nil, err
	}

	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.LogResp(nil, nil, retErr, time.Since(start)) }(time.Now())

	// If the cluster's enterprise token is expired, login is disabled
	if err := a.expiredEnterpriseCheck(ctx); err != nil {
		return nil, err
	}

	// verify whatever credential the user has presented, and write a new
	// Pachyderm token for the user that their credential belongs to
	var pachToken string
	switch {
	case req.OIDCState != "":
		// Determine caller's Pachyderm/OIDC user info (email)
		email, err := a.OIDCStateToEmail(ctx, req.OIDCState)
		if err != nil {
			return nil, err
		}

		username := auth.UserPrefix + email

		// Generate a new Pachyderm token and write it
		pachToken = uuid.NewWithoutDashes()
		if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
			tokens := a.tokens.ReadWrite(stm)
			return tokens.PutTTL(auth.HashToken(pachToken),
				&auth.TokenInfo{
					Subject: username,
					Source:  auth.TokenInfo_AUTHENTICATE,
				},
				defaultSessionTTLSecs)
		}); err != nil {
			return nil, errors.Wrapf(err, "error storing auth token for user \"%s\"", username)
		}
	case req.IdToken != "":
		// Determine caller's Pachyderm/OIDC user info (email)
		token, claims, err := a.validateIDToken(ctx, req.IdToken)
		if err != nil {
			return nil, err
		}

		username := auth.UserPrefix + claims.Email

		// Sync the user's group membership from the groups claim
		if err := a.syncGroupMembership(ctx, claims); err != nil {
			return nil, err
		}

		// Compute the remaining time before the ID token expires,
		// and limit the pach token to the same expiration time.
		// If the token would be longer-lived than the default pach token,
		// TTL clamp the expiration to the default TTL.
		expirationSecs := int64(time.Until(token.Expiry).Seconds())
		if expirationSecs > defaultSessionTTLSecs {
			expirationSecs = defaultSessionTTLSecs
		}

		// Generate a new Pachyderm token and write it
		pachToken = uuid.NewWithoutDashes()
		if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
			tokens := a.tokens.ReadWrite(stm)
			return tokens.PutTTL(auth.HashToken(pachToken),
				&auth.TokenInfo{
					Subject: username,
					Source:  auth.TokenInfo_AUTHENTICATE,
				},
				expirationSecs)
		}); err != nil {
			return nil, errors.Wrapf(err, "error storing auth token for user \"%s\"", username)
		}
	default:
		return nil, errors.Errorf("unrecognized authentication mechanism (old pachd?)")
	}

	logrus.Info("Authentication checks successful, now returning pachToken")

	// Return new pachyderm token to caller
	return &auth.AuthenticateResponse{
		PachToken: pachToken,
	}, nil
}

func (a *apiServer) getCallerTTL(ctx context.Context) (int64, error) {
	token, err := auth.GetAuthToken(ctx)
	if err != nil {
		return 0, err
	}
	ttl, err := a.tokens.ReadOnly(ctx).TTL(auth.HashToken(token)) // lookup token TTL
	if err != nil {
		return 0, errors.Wrapf(err, "error looking up TTL for token")
	}
	return ttl, nil
}

func resourceKey(r *auth.Resource) string {
	if r.Type == auth.ResourceType_CLUSTER {
		return clusterRoleBindingKey
	}
	return fmt.Sprintf("%s:%s", r.Type, r.Name)
}

// AuthorizeInTransaction is identical to Authorize except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) AuthorizeInTransaction(
	txnCtx *txnenv.TransactionContext,
	req *auth.AuthorizeRequest,
) (resp *auth.AuthorizeResponse, retErr error) {
	if err := a.isActive(); err != nil {
		return nil, err
	}

	callerInfo, err := a.getAuthenticatedUser(txnCtx.ClientContext)
	if err != nil {
		return nil, err
	}

	// If the cluster's enterprise token is expired, only pipelines can authorize
	state, err := a.getEnterpriseTokenState(txnCtx.ClientContext)
	if err != nil {
		return nil, errors.Wrapf(err, "error confirming Pachyderm Enterprise token")
	}
	if state != enterpriseclient.State_ACTIVE &&
		!strings.HasPrefix(callerInfo.Subject, auth.PipelinePrefix) {
		return nil, errors.New("Pachyderm Enterprise is not active in this " +
			"cluster (until Pachyderm Enterprise is re-activated or Pachyderm " +
			"auth is deactivated, only pipelines can perform any operations)")
	}

	request := newAuthorizeRequest(callerInfo.Subject, req.Permissions, a.groupsForUser)

	// Check the permissions at the cluster level
	binding, ok := a.clusterRoleBindingCache.Load().(*auth.RoleBinding)
	if !ok {
		return nil, errors.New("cached cluster role binding had unexpected type")
	}
	if err := request.evaluateRoleBinding(binding); err != nil {
		return nil, err
	}

	// If all the permissions are satisfied by the cached cluster binding don't
	// retrieve the resource bindings
	if request.satisfied() {
		return &auth.AuthorizeResponse{Authorized: true}, nil
	}

	// Get the role bindings for the resource to check
	var roleBinding auth.RoleBinding
	if err := a.roleBindings.ReadWrite(txnCtx.Stm).Get(resourceKey(req.Resource), &roleBinding); err != nil && !col.IsErrNotFound(err) {
		return nil, errors.Wrapf(err, "error getting role bindings for %s \"%s\"", req.Resource.Type, req.Resource.Name)
	}
	if err := request.evaluateRoleBinding(&roleBinding); err != nil {
		return nil, err
	}

	return &auth.AuthorizeResponse{
		Authorized: request.satisfied(),
	}, nil
}

// Authorize implements the protobuf auth.Authorize RPC
func (a *apiServer) Authorize(
	ctx context.Context,
	req *auth.AuthorizeRequest,
) (resp *auth.AuthorizeResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	var response *auth.AuthorizeResponse
	if err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		var err error
		response, err = a.AuthorizeInTransaction(txnCtx, req)
		return err
	}); err != nil {
		return nil, err
	}
	return response, nil
}

// WhoAmI implements the protobuf auth.WhoAmI RPC
func (a *apiServer) WhoAmI(ctx context.Context, req *auth.WhoAmIRequest) (resp *auth.WhoAmIResponse, retErr error) {
	a.pachLogger.LogAtLevelFromDepth(req, nil, nil, 0, logrus.DebugLevel, 2)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.isActive(); err != nil {
		return nil, err
	}

	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// Get TTL of user's token
	ttl := int64(-1) // value returned by etcd for keys w/ no lease (no TTL)
	if callerInfo.Subject != ppsUser {
		token, err := auth.GetAuthToken(ctx)
		if err != nil {
			return nil, err
		}
		ttl, err = a.tokens.ReadOnly(ctx).TTL(auth.HashToken(token)) // lookup token TTL
		if err != nil {
			return nil, errors.Wrapf(err, "error looking up TTL for token")
		}
	}

	// return final result
	return &auth.WhoAmIResponse{
		Username: callerInfo.Subject,
		TTL:      ttl,
	}, nil
}

// hasClusterPermissions returns true if the subject has the specified permissions at the cluster level.
// It returns the list of permissions  that are _not fulfilled_ by the cluster role bindings.
// This hits the cached cluster role binding instead of etcd, so it's fast _unless_ the user derives
// their permissions from a group membership.
func (a *apiServer) hasClusterPermissions(ctx context.Context, subject string, desired []auth.Permission) (bool, error) {
	binding, ok := a.clusterRoleBindingCache.Load().(*auth.RoleBinding)
	if !ok {
		return false, errors.New("cached auth config had unexpected type")
	}

	request := newAuthorizeRequest(subject, desired, a.groupsForUser)
	if err := request.evaluateRoleBinding(binding); err != nil {
		return false, err
	}
	return request.satisfied(), nil
}

// CreateRoleBindingInTransaction is identitical to CreateRoleBinding except that it can run inside
// an existing etcd STM transaction. This is not an RPC.
func (a *apiServer) CreateRoleBindingInTransaction(
	txnCtx *txnenv.TransactionContext,
	req *auth.CreateRoleBindingRequest,
) (*auth.CreateRoleBindingResponse, error) {
	if err := a.isActive(); err != nil {
		return nil, err
	}

	if err := a.checkCanonicalSubject(req.Principal); err != nil {
		return nil, err
	}

	// Check that the role binding does not currently exist
	key := resourceKey(req.Resource)
	roleBindings := a.roleBindings.ReadWrite(txnCtx.Stm)
	var bindings auth.RoleBindings
	if err := roleBindings.Get(key, &bindings); err == nil {
		return nil, fmt.Errorf("role binding already exists for resource %v", req.Resource)
	} else if !col.IsErrNotFound(err) {
		return nil, err
	}

	roleBindings.Entries = map[string]*auth.Roles{
		req.Principal: make(map[string]bool),
	}

	for _, r := range req.Roles {
		roleBindings.Entries[req.Principal][r] = true
	}

	if err := roleBindings.Put(key, &roleBindings); err != nil {
		return nil, err
	}

	return &auth.CreateRoleBindingResponse{}, nil
}

// ModifyRolesInTransaction is identical to ModifyRoles except that it can run inside
// an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) ModifyRolesInTransaction(
	txnCtx *txnenv.TransactionContext,
	req *auth.ModifyRolesRequest,
) (*auth.ModifyRolesResponse, error) {
	if err := a.isActive(); err != nil {
		return nil, err
	}

	if err := a.checkCanonicalSubject(req.Principal); err != nil {
		return nil, err
	}

	callerInfo, err := a.getAuthenticatedUser(txnCtx.ClientContext)
	if err != nil {
		return nil, err
	}

	// ModifyRoles can be called for any type of resource, and the permission required depends on
	// the type of resource.
	var permissions []auth.Permission
	switch req.Resource.Type {
	case auth.ResourceType_CLUSTER:
		permissions := []auth.Permission{auth.Permission_CLUSTER_ADMIN}
	case auth.ResourceType_REPO:
		permissions := []auth.Permission{auth.Permission_REPO_MODIFY_BINDINGS}
	default:
		return nil, fmt.Errorf("unknown resource type %v", req.Resource.Type)
	}

	// Check if the caller is authorized
	authorized, err := a.AuthorizeInTransaction(txnCtx, &auth.AuthorizeRequest{
		Resource:    req.Resource,
		Permissions: permissions,
	})
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, &auth.ErrNotAuthorized{
			Subject:     callerInfo.Subject,
			Resource:    req.Resource,
			Permissions: permissions,
		}
	}

	key := resourceKey(req.Resource)
	roleBindings := a.roleBindings.ReadWrite(txnCtx.Stm)
	var bindings auth.RoleBindings
	if err := roleBindings.Get(key, &bindings); err != nil {
		return nil, err
	}

	if _, ok := bindings[req.Principal]; !ok {
		bindings[req.Principal] = make(map[string]bool)
	}

	for _, r := range req.ToRemove {
		delete(bindings[req.Principal], r)
	}

	for _, r := range req.ToAdd {
		bindings[req.Principal][r] = true
	}

	return &auth.ModifyRolesResponse{}, nil
}

// ModifyRoles implements the protobuf auth.ModifyRoles RPC
func (a *apiServer) ModifyRoles(ctx context.Context, req *auth.ModifyRolesRequest) (resp *auth.ModifyRolesResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	var response *auth.ModifyRolesResponse
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		var err error
		response, err = txn.ModifyRoles(req)
		return err
	}); err != nil {
		return nil, err
	}
	return response, nil
}

// GetRolesInTransaction is identical to GetRoles except that it can run inside
// an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) GetRolesInTransaction(
	txnCtx *txnenv.TransactionContext,
	req *auth.GetRolesRequest,
) (*auth.GetRolesResponse, error) {
	if err := a.isActive(); err != nil {
		return nil, err
	}

	callerInfo, err := a.getAuthenticatedUser(txnCtx.ClientContext)
	if err != nil {
		return nil, err
	}
	callerIsAdmin, err := a.hasClusterRole(txnCtx.ClientContext, callerInfo.Subject, auth.ClusterRole_FS)
	if err != nil {
		return nil, err
	}

	// If the caller is getting another user's scopes, the caller must have
	// READER access to every repo in the request--check this requirement
	targetSubject := callerInfo.Subject
	mustHaveReadAccess := false
	if req.Username != "" {
		if err := a.checkCanonicalSubject(req.Username); err != nil {
			return nil, err
		}
		targetSubject = req.Username
		mustHaveReadAccess = true
	}

	// Read repo ACLs from etcd & compute final result
	// Note: for now, we don't return OWNER if the user is an admin, even though
	// that's their effective access scope for all repos--the caller may want to
	// know what will happen if the user's admin privileges are revoked. Note
	// that pfs.ListRepo overrides this logic—the auth info it returns for a
	// listed repo indicates that a user is OWNER of all repos if they are an
	// admin
	acls := a.acls.ReadWrite(txnCtx.Stm)
	response := new(auth.GetScopeResponse)
	for _, repo := range req.Repos {
		var acl auth.ACL
		if err := acls.Get(repo, &acl); err != nil && !col.IsErrNotFound(err) {
			return nil, err
		}

		// determine if ACL read is authorized
		if mustHaveReadAccess && !callerIsAdmin {
			// Caller is getting another user's scopes. Check if the caller is
			// authorized to view this repo's ACL
			callerScope, err := a.getScope(txnCtx.ClientContext, callerInfo.Subject, &acl)
			if err != nil {
				return nil, err
			}
			if callerScope < auth.Scope_READER {
				return nil, &auth.ErrNotAuthorized{
					Subject:  callerInfo.Subject,
					Repo:     repo,
					Required: auth.Scope_READER,
				}
			}
		}

		// compute target's access scope to this repo
		targetScope, err := a.getScope(txnCtx.ClientContext, targetSubject, &acl)
		if err != nil {
			return nil, err
		}
		response.Scopes = append(response.Scopes, targetScope)
	}

	return response, nil
}

// GetRoles implements the protobuf auth.GetRoles RPC
func (a *apiServer) GetRoles(ctx context.Context, req *auth.GetRolesRequest) (resp *auth.GetRolesResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	var response *auth.GetScopeResponse
	if err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		var err error
		response, err = a.GetScopeInTransaction(txnCtx, req)
		return err
	}); err != nil {
		return nil, err
	}
	return response, nil
}

// GetRoleBindingsInTransaction is identical to GetRoleBindings except that it can run inside
// an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) GetRoleBindingsInTransaction(
	txnCtx *txnenv.TransactionContext,
	req *auth.GetRoleBindingsRequest,
) (*auth.GetRoleBindingsResponse, error) {
	if err := a.isActive(); err != nil {
		return nil, err
	}

	// Validate request
	if req.Repo == "" {
		return nil, errors.Errorf("invalid request: must provide name of repo to get that repo's ACL")
	}

	// Get calling user
	callerInfo, err := a.getAuthenticatedUser(txnCtx.ClientContext)
	if err != nil {
		return nil, err
	}
	if err := a.expiredEnterpriseCheck(txnCtx.ClientContext); err != nil {
		return nil, err
	}

	// Read repo ACL from etcd
	acl := &auth.ACL{}
	if err := a.acls.ReadWrite(txnCtx.Stm).Get(req.Repo, acl); err != nil && !col.IsErrNotFound(err) {
		return nil, err
	}
	response := &auth.GetRoleBindingsResponse{
		Entries: make([]*auth.ACLEntry, 0),
	}
	for user, scope := range acl.Entries {
		response.Entries = append(response.Entries, &auth.ACLEntry{
			Username: user,
			Scope:    scope,
		})
	}
	return response, nil
}

// GetRoleBindings implements the protobuf auth.GetRoleBindings RPC
func (a *apiServer) GetRoleBindings(ctx context.Context, req *auth.GetRoleBindingsRequest) (resp *auth.GetRoleBindingsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	var response *auth.GetRoleBindingsResponse
	if err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		var err error
		response, err = a.GetRoleBindingsInTransaction(txnCtx, req)
		return err
	}); err != nil {
		return nil, err
	}
	return response, nil
}

// SetACLInTransaction is identical to SetACL except that it can run inside
// an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) SetACLInTransaction(
	txnCtx *txnenv.TransactionContext,
	req *auth.SetACLRequest,
) (*auth.SetACLResponse, error) {
	if err := a.isActive(); err != nil {
		return nil, err
	}

	// Validate request
	if req.Repo == "" {
		return nil, errors.Errorf("invalid request: must provide name of repo you want to modify")
	}

	// Get calling user
	callerInfo, err := a.getAuthenticatedUser(txnCtx.ClientContext)
	if err != nil {
		return nil, err
	}
	isAdmin, err := a.hasClusterRole(txnCtx.ClientContext, callerInfo.Subject, auth.ClusterRole_FS)
	if err != nil {
		return nil, err
	}

	// Canonicalize entries in the request (must have canonical request before we
	// can authorize, as we inspect the ACL contents during authorization in the
	// case where we're creating a new repo)
	newACL := new(auth.ACL)
	if len(req.Entries) > 0 {
		newACL.Entries = make(map[string]auth.Scope)
	}
	for _, entry := range req.Entries {
		user, scope := entry.Username, entry.Scope
		if user == ppsUser {
			continue
		}

		if err := a.checkCanonicalSubject(user); err != nil {
			return nil, err
		}
		newACL.Entries[user] = scope
	}

	// Read repo ACL from etcd
	acls := a.acls.ReadWrite(txnCtx.Stm)

	// determine if the caller is authorized to set this repo's ACL
	authorized, err := func() (bool, error) {
		if isAdmin {
			// admins are automatically authorized
			return true, nil
		}

		// Check if the cluster's enterprise token is expired (fail if so)
		state, err := a.getEnterpriseTokenState(txnCtx.ClientContext)
		if err != nil {
			return false, errors.Wrapf(err, "error confirming Pachyderm Enterprise token")
		}
		if state != enterpriseclient.State_ACTIVE {
			return false, errors.Errorf("Pachyderm Enterprise is not active in this " +
				"cluster (only a cluster admin can modify an ACL)")
		}

		// Check if there is an existing ACL, and if the user is on it
		var acl auth.ACL
		if err := acls.Get(req.Repo, &acl); err != nil {
			// ACL not found -- construct empty ACL proto
			acl.Entries = make(map[string]auth.Scope)
		}
		if len(acl.Entries) > 0 {
			// ACL is present; caller must be authorized directly
			scope, err := a.getScope(txnCtx.ClientContext, callerInfo.Subject, &acl)
			if err != nil {
				return false, err
			}
			if scope == auth.Scope_OWNER {
				return true, nil
			}
			return false, nil
		}

		// No ACL -- check if the repo being modified exists
		_, err = txnCtx.Pfs().InspectRepoInTransaction(txnCtx, &pfs.InspectRepoRequest{Repo: &pfs.Repo{Name: req.Repo}})
		if err == nil {
			// Repo exists -- user isn't authorized
			return false, nil
		} else if !strings.HasSuffix(err.Error(), "not found") {
			// Unclear if repo exists -- return error
			return false, errors.Wrapf(err, "could not inspect \"%s\"", req.Repo)
		} else if len(newACL.Entries) == 1 &&
			newACL.Entries[callerInfo.Subject] == auth.Scope_OWNER {
			// Special case: Repo doesn't exist, but user is creating a new Repo, and
			// making themself the owner, e.g. for CreateRepo or CreatePipeline, then
			// the request is authorized
			return true, nil
		}
		return false, err
	}()
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, &auth.ErrNotAuthorized{
			Subject:  callerInfo.Subject,
			Repo:     req.Repo,
			Required: auth.Scope_OWNER,
		}
	}

	// Set new ACL
	if len(newACL.Entries) == 0 {
		err := acls.Delete(req.Repo)
		if err != nil && !col.IsErrNotFound(err) {
			return nil, err
		}
	} else {
		err = acls.Put(req.Repo, newACL)
		if err != nil {
			return nil, errors.Wrapf(err, "could not put new ACL")
		}
	}
	return &auth.SetACLResponse{}, nil
}

// authorizeNewToken is a helper for GetAuthToken and GetOTP that checks if
// the caller ('callerInfo') is authorized to get a Pachyderm token or OTP for
// 'targetUser', or for themselves if 'targetUser' is empty (a convention use by
// both API endpoints). It returns a canonicalized version of 'targetSubject'
// (or the caller if 'targetSubject' is empty) xor an error, e.g. if the caller
// isn't authorized.
//
// The code isn't too long or complex, but centralizing it keeps the two API
// endpoints syncronized
func (a *apiServer) authorizeNewToken(ctx context.Context, callerInfo *auth.TokenInfo, isAdmin bool, targetSubject string) (string, error) {
	if targetSubject == "" {
		// [Simple case] caller wants an implicit OTP for themselves
		// Canonicalization: callerInfo.Subject is already canonical.
		// Authorization: Getting a token/OTP for yourself is always authorized.
		// NOTE: After this point, req.Subject is permitted to be ppsUser (even
		// though we reject ppsUser when set explicitly.)
		targetSubject = callerInfo.Subject
	} else {
		// [Harder case] explicit req.Subject
		if err := a.checkCanonicalSubject(targetSubject); err != nil {
			return "", err
		}

		// Authorization: caller must be admin to get OTP for another user
		if !isAdmin && targetSubject != callerInfo.Subject {
			return "", &auth.ErrNotAuthorized{
				Subject: callerInfo.Subject,
				AdminOp: "get token on behalf of another user",
			}
		}
	}
	return targetSubject, nil
}

// GetAuthToken implements the protobuf auth.GetAuthToken RPC
func (a *apiServer) GetAuthToken(ctx context.Context, req *auth.GetAuthTokenRequest) (resp *auth.GetAuthTokenResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	a.txnEnv.WithWriteContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		resp, retErr = a.GetAuthTokenInTransaction(txnCtx, req)
		return retErr
	})
	return resp, retErr
}

// GetAuthToken implements the protobuf auth.GetAuthToken RPC
func (a *apiServer) GetAuthTokenInTransaction(txnCtx *txnenv.TransactionContext, req *auth.GetAuthTokenRequest) (resp *auth.GetAuthTokenResponse, retErr error) {
	if err := a.isActive(); err != nil {
		// GetAuthToken must work in the partially-activated state so that PPS can
		// get tokens for all existing pipelines during activation
		return nil, err
	}

	// Get current caller and authorize them if req.Subject is set to a different
	// user
	callerInfo, err := a.getAuthenticatedUser(txnCtx.ClientContext)
	if err != nil {
		return nil, err
	}
	isAdmin, err := a.hasClusterRole(txnCtx.ClientContext, callerInfo.Subject, auth.ClusterRole_SUPER)
	if err != nil {
		return nil, err
	}

	if req.TTL < 0 && !isAdmin {
		return nil, errors.Errorf("GetAuthTokenRequest.TTL must be >= 0")
	}

	// check if this request is auhorized
	req.Subject, err = a.authorizeNewToken(txnCtx.ClientContext, callerInfo, isAdmin, req.Subject)
	if err != nil {
		if errors.As(err, &auth.ErrNotAuthorized{}) {
			// return more descriptive error
			return nil, &auth.ErrNotAuthorized{
				Subject: callerInfo.Subject,
				AdminOp: "GetAuthToken on behalf of another user",
			}
		}
		return nil, err
	}

	// TODO: switch the ppsUser to have the `pach:` prefix to simplify this case
	if req.Subject == ppsUser || strings.HasPrefix(req.Subject, auth.PachPrefix) {
		return nil, errors.Errorf("GetAuthTokenRequest.Subject is invalid")
	}

	// Compute TTL for new token that the user will get once OTP is exchanged
	// Note: For Pachyderm <1.10, admin tokens always come with an indefinite
	// session, unless a limit is requested. For Pachyderm >=1.10, Admins always
	// use default TTL (30 days currently), unless they explicitly request an
	// indefinite TTL.
	//
	// TODO(msteffen): Either way, admins can can extend their session by getting
	// a new token for themselves. I don't know if allowing users to do this is
	// bad security practice (it might be). However, preventing admins from
	// extending their own session here wouldn't improve security: admins can
	// currently manufacture a buddy account, promote it to admin, and then extend
	// their session indefinitely with the buddy account. Pachyderm cannot prevent
	// this with the information is currently stores, so preventing admins from
	// extending their session indefinitely would require larger changes to the
	// auth model.
	if !isAdmin {
		// Caller is getting OTP for themselves--use TTL of their current token
		ttl, err := a.getCallerTTL(txnCtx.ClientContext)
		if err != nil {
			return nil, err
		}
		if req.TTL == 0 || req.TTL > ttl {
			req.TTL = ttl
		}
	} else if version.IsAtLeast(1, 10) && (req.TTL == 0 || req.TTL > defaultSessionTTLSecs) {
		// To create a token with no TTL, an admin can call GetAuthToken and set TTL
		// to -1, but the default behavior (TTL == 0) is use the default token
		// lifetime.
		req.TTL = defaultSessionTTLSecs
	}
	tokenInfo := auth.TokenInfo{
		Source:  auth.TokenInfo_GET_TOKEN,
		Subject: req.Subject,
	}

	// generate new token, and write to etcd
	token := uuid.NewWithoutDashes()
	if err := a.tokens.ReadWrite(txnCtx.Stm).PutTTL(auth.HashToken(token), &tokenInfo, req.TTL); err != nil {
		if tokenInfo.Subject != ppsUser {
			return nil, errors.Wrapf(err, "error storing token for user \"%s\"", tokenInfo.Subject)
		}
		return nil, errors.Wrapf(err, "error storing token")
	}
	return &auth.GetAuthTokenResponse{
		Subject: req.Subject,
		Token:   token,
	}, nil
}

// GetOIDCLogin implements the protobuf auth.GetOIDCLogin RPC
func (a *apiServer) GetOIDCLogin(ctx context.Context, req *auth.GetOIDCLoginRequest) (resp *auth.GetOIDCLoginResponse, retErr error) {
	a.LogReq(req)
	// Don't log response to avoid logging OIDC state token
	defer func(start time.Time) { a.LogResp(req, nil, retErr, time.Since(start)) }(time.Now())
	var err error

	authURL, state, err := a.GetOIDCLoginURL(ctx)
	if err != nil {
		return nil, err
	}
	return &auth.GetOIDCLoginResponse{
		LoginURL: authURL,
		State:    state,
	}, nil
}

// ExtendAuthToken implements the protobuf auth.ExtendAuthToken RPC
func (a *apiServer) ExtendAuthToken(ctx context.Context, req *auth.ExtendAuthTokenRequest) (resp *auth.ExtendAuthTokenResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if req.TTL == 0 {
		return nil, errors.Errorf("invalid request: ExtendAuthTokenRequest.TTL must be > 0")
	}

	// Only let people extend tokens by up to 30 days (the equivalent of logging
	// in again)
	if req.TTL > defaultSessionTTLSecs {
		return nil, errors.Errorf("can only extend tokens by at most %d seconds", defaultSessionTTLSecs)
	}

	// The token must already exist. If a token has been revoked, it can't be
	// extended
	if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)

		// Actually look up the request token in the relevant collections
		var tokenInfo auth.TokenInfo
		if err := tokens.Get(auth.HashToken(req.Token), &tokenInfo); err != nil && !col.IsErrNotFound(err) {
			return err
		}
		if tokenInfo.Subject == "" {
			return auth.ErrBadToken
		}

		ttl, err := tokens.TTL(auth.HashToken(req.Token))
		if err != nil {
			return errors.Wrapf(err, "error looking up TTL for token")
		}
		// TODO(msteffen): ttl may be -1 if the token has no TTL. We deliberately do
		// not check this case so that admins can put TTLs on tokens that don't have
		// them (otherwise any attempt to do so would get ErrTooShortTTL), but that
		// decision may be revised
		if req.TTL < ttl {
			return auth.ErrTooShortTTL{
				RequestTTL:  req.TTL,
				ExistingTTL: ttl,
			}
		}
		return tokens.PutTTL(auth.HashToken(req.Token), &tokenInfo, req.TTL)
	}); err != nil {
		return nil, err
	}
	return &auth.ExtendAuthTokenResponse{}, nil
}

// RevokeAuthToken implements the protobuf auth.RevokeAuthToken RPC
func (a *apiServer) RevokeAuthToken(ctx context.Context, req *auth.RevokeAuthTokenRequest) (resp *auth.RevokeAuthTokenResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	a.txnEnv.WithWriteContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		resp, retErr = a.RevokeAuthTokenInTransaction(txnCtx, req)
		return retErr
	})
	return resp, retErr
}

func (a *apiServer) RevokeAuthTokenInTransaction(txnCtx *txnenv.TransactionContext, req *auth.RevokeAuthTokenRequest) (resp *auth.RevokeAuthTokenResponse, retErr error) {
	if err := a.isActive(); err != nil {
		return nil, err
	}
	// Get the caller. Users can revoke their own tokens, and admins can revoke
	// any user's token
	callerInfo, err := a.getAuthenticatedUser(txnCtx.ClientContext)
	if err != nil {
		return nil, err
	}
	isAdmin, err := a.hasClusterRole(txnCtx.ClientContext, callerInfo.Subject, auth.ClusterRole_SUPER)
	if err != nil {
		return nil, err
	}

	tokens := a.tokens.ReadWrite(txnCtx.Stm)
	var tokenInfo auth.TokenInfo
	if err := tokens.Get(auth.HashToken(req.Token), &tokenInfo); err != nil {
		if !col.IsErrNotFound(err) {
			return nil, err
		}
	}
	if !isAdmin && tokenInfo.Subject != callerInfo.Subject {
		return nil, &auth.ErrNotAuthorized{
			Subject: callerInfo.Subject,
			AdminOp: "RevokeAuthToken on another user's token",
		}
	}
	if err := tokens.Delete(auth.HashToken(req.Token)); err != nil {
		return nil, err
	}
	return &auth.RevokeAuthTokenResponse{}, nil
}

// setGroupsForUserInternal is a helper function used by SetGroupsForUser, and
// also by handleSAMLResponse and handleOIDCExchangeInternal (which updates
// group membership information based on signed SAML assertions or JWT claims).
// This does no auth checks, so the caller must do all relevant authorization.
func (a *apiServer) setGroupsForUserInternal(ctx context.Context, subject string, groups []string) error {
	_, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		members := a.members.ReadWrite(stm)

		// Get groups to remove/add user from/to
		var removeGroups auth.Groups
		addGroups := addToSet(nil, groups...)
		if err := members.Get(subject, &removeGroups); err == nil {
			for _, group := range groups {
				if removeGroups.Groups[group] {
					removeGroups.Groups = removeFromSet(removeGroups.Groups, group)
					addGroups = removeFromSet(addGroups, group)
				}
			}
		}

		// Set groups for user
		if err := members.Put(subject, &auth.Groups{
			Groups: addToSet(nil, groups...),
		}); err != nil {
			return err
		}

		// Remove user from previous groups
		groups := a.groups.ReadWrite(stm)
		var membersProto auth.Users
		for group := range removeGroups.Groups {
			if err := groups.Upsert(group, &membersProto, func() error {
				membersProto.Usernames = removeFromSet(membersProto.Usernames, subject)
				return nil
			}); err != nil {
				return err
			}
		}

		// Add user to new groups
		for group := range addGroups {
			if err := groups.Upsert(group, &membersProto, func() error {
				membersProto.Usernames = addToSet(membersProto.Usernames, subject)
				return nil
			}); err != nil {
				return err
			}
		}

		return nil
	})
	return err
}

// SetGroupsForUser implements the protobuf auth.SetGroupsForUser RPC
func (a *apiServer) SetGroupsForUser(ctx context.Context, req *auth.SetGroupsForUserRequest) (resp *auth.SetGroupsForUserResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.checkCanonicalSubject(req.Username); err != nil {
		return nil, err
	}
	// TODO(msteffen): canonicalize group names
	if err := a.setGroupsForUserInternal(ctx, req.Username, req.Groups); err != nil {
		return nil, err
	}
	return &auth.SetGroupsForUserResponse{}, nil
}

// ModifyMembers implements the protobuf auth.ModifyMembers RPC
func (a *apiServer) ModifyMembers(ctx context.Context, req *auth.ModifyMembersRequest) (resp *auth.ModifyMembersResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.checkCanonicalSubjects(req.Add); err != nil {
		return nil, err
	}
	// TODO(bryce) Skip canonicalization if the users can be found.
	if err := a.checkCanonicalSubjects(req.Remove); err != nil {
		return nil, err
	}

	if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		members := a.members.ReadWrite(stm)
		var groupsProto auth.Groups
		for _, username := range req.Add {
			if err := members.Upsert(username, &groupsProto, func() error {
				groupsProto.Groups = addToSet(groupsProto.Groups, req.Group)
				return nil
			}); err != nil {
				return err
			}
		}
		for _, username := range req.Remove {
			if err := members.Upsert(username, &groupsProto, func() error {
				groupsProto.Groups = removeFromSet(groupsProto.Groups, req.Group)
				return nil
			}); err != nil {
				return err
			}
		}

		groups := a.groups.ReadWrite(stm)
		var membersProto auth.Users
		if err := groups.Upsert(req.Group, &membersProto, func() error {
			membersProto.Usernames = addToSet(membersProto.Usernames, req.Add...)
			membersProto.Usernames = removeFromSet(membersProto.Usernames, req.Remove...)
			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &auth.ModifyMembersResponse{}, nil
}

func addToSet(set map[string]bool, elems ...string) map[string]bool {
	if set == nil {
		set = map[string]bool{}
	}

	for _, elem := range elems {
		set[elem] = true
	}
	return set
}

func removeFromSet(set map[string]bool, elems ...string) map[string]bool {
	if set != nil {
		for _, elem := range elems {
			delete(set, elem)
		}
	}

	return set
}

// getGroups is a helper function used primarily by the GRPC API GetGroups, but
// also by Authorize() and isAdmin().
func (a *apiServer) getGroups(ctx context.Context, subject string) ([]string, error) {
	members := a.members.ReadOnly(ctx)
	var groupsProto auth.Groups
	if err := members.Get(subject, &groupsProto); err != nil {
		if col.IsErrNotFound(err) {
			return []string{}, nil
		}
		return nil, err
	}
	return setToList(groupsProto.Groups), nil
}

// GetGroups implements the protobuf auth.GetGroups RPC
func (a *apiServer) GetGroups(ctx context.Context, req *auth.GetGroupsRequest) (resp *auth.GetGroupsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// must be admin or user getting your own groups (default value of
	// req.Username is the caller).
	// Note: hasClusterRole(subject) calls getGroups(subject), so getGroups(subject)
	// must only call hasClusterRole(caller) when caller != subject, or else we'll have
	// infinite recursion
	var target string
	if req.Username != "" && req.Username != callerInfo.Subject {
		isAdmin, err := a.hasClusterRole(ctx, callerInfo.Subject, auth.ClusterRole_SUPER)
		if err != nil {
			return nil, err
		}
		if !isAdmin {
			return nil, &auth.ErrNotAuthorized{
				Subject: callerInfo.Subject,
				AdminOp: "GetGroups",
			}
		}
		if err := a.checkCanonicalSubject(req.Username); err != nil {
			return nil, err
		}
		target = req.Username
	} else {
		target = callerInfo.Subject
	}

	groups, err := a.getGroups(ctx, target)
	if err != nil {
		return nil, err
	}
	return &auth.GetGroupsResponse{Groups: groups}, nil
}

// GetUsers implements the protobuf auth.GetUsers RPC
func (a *apiServer) GetUsers(ctx context.Context, req *auth.GetUsersRequest) (resp *auth.GetUsersResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	// Filter by group
	if req.Group != "" {
		var membersProto auth.Users
		if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
			groups := a.groups.ReadWrite(stm)
			if err := groups.Get(req.Group, &membersProto); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}

		return &auth.GetUsersResponse{Usernames: setToList(membersProto.Usernames)}, nil
	}

	membersCol := a.members.ReadOnly(ctx)
	groups := &auth.Groups{}
	var users []string
	if err := membersCol.List(groups, col.DefaultOptions, func(user string) error {
		users = append(users, user)
		return nil
	}); err != nil {
		return nil, err
	}
	return &auth.GetUsersResponse{Usernames: users}, nil
}

func setToList(set map[string]bool) []string {
	if set == nil {
		return []string{}
	}

	list := []string{}
	for elem := range set {
		list = append(list, elem)
	}
	return list
}

func (a *apiServer) getAuthenticatedUser(ctx context.Context) (*auth.TokenInfo, error) {
	// TODO(msteffen) cache these lookups, especially since users always authorize
	// themselves at the beginning of a request. Don't want to look up the same
	// token -> username entry twice.
	token, err := auth.GetAuthToken(ctx)
	if err != nil {
		return nil, err
	}
	if token == a.ppsToken {
		// TODO(msteffen): This is a hack. The idea is that there is a logical user
		// entry mapping ppsToken to ppsUser. Soon, ppsUser will go away and
		// this check should happen in authorize
		return &auth.TokenInfo{
			Subject: ppsUser,
			Source:  auth.TokenInfo_GET_TOKEN,
		}, nil
	}

	// Lookup the token
	var tokenInfo auth.TokenInfo
	if err := a.tokens.ReadOnly(ctx).Get(auth.HashToken(token), &tokenInfo); err != nil {
		if col.IsErrNotFound(err) {
			return nil, auth.ErrBadToken
		}
		return nil, err
	}
	return &tokenInfo, nil
}

// checkCanonicalSubjects applies checkCanonicalSubject to a list
func (a *apiServer) checkCanonicalSubjects(subjects []string) error {
	for _, subject := range subjects {
		if err := a.checkCanonicalSubject(subject); err != nil {
			return err
		}
	}
	return nil
}

// checkCanonicalSubject returns an error if a subject doesn't have a prefix, or the prefix is
// not recognized
func (a *apiServer) checkCanonicalSubject(subject string) error {
	if subject == allClusterUsersSubject {
		return nil
	}

	colonIdx := strings.Index(subject, ":")
	if colonIdx < 0 {
		return errors.Errorf("subject has no prefix, must be of the form <type>:<name>")
	}
	prefix := subject[:colonIdx]

	// check against fixed prefixes
	prefix += ":" // append ":" to match constants
	switch prefix {
	case auth.PipelinePrefix, auth.RobotPrefix, auth.PachPrefix, auth.UserPrefix:
		break
	default:
		return errors.Errorf("subject has unrecognized prefix: %s", subject[:colonIdx+1])
	}
	return nil
}

// GetConfiguration implements the protobuf auth.GetConfiguration RPC.
func (a *apiServer) GetConfiguration(ctx context.Context, req *auth.GetConfigurationRequest) (resp *auth.GetConfigurationResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	config, ok := a.configCache.Load().(*auth.OIDCConfig)
	if !ok {
		return nil, errors.New("cached auth config had unexpected type")
	}

	return &auth.GetConfigurationResponse{
		Configuration: config,
	}, nil
}

// SetConfiguration implements the protobuf auth.SetConfiguration RPC
func (a *apiServer) SetConfiguration(ctx context.Context, req *auth.SetConfigurationRequest) (resp *auth.SetConfigurationResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	var configToStore *auth.OIDCConfig
	if req.Configuration != nil {
		// Validate new config
		if err := validateOIDCConfig(ctx, req.Configuration); err != nil {
			return nil, err
		}
		configToStore = req.Configuration
	} else {
		configToStore = proto.Clone(&DefaultOIDCConfig).(*auth.OIDCConfig)
	}

	// set the new config
	if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		return a.authConfig.ReadWrite(stm).Put(configKey, configToStore)
	}); err != nil {
		return nil, err
	}

	// block until the watcher observes the write
	if err := backoff.Retry(func() error {
		record, ok := a.configCache.Load().(*auth.OIDCConfig)
		if !ok {
			return errors.Errorf("could not retrieve auth config from cache")
		}
		if !proto.Equal(record, configToStore) {
			return errors.Errorf("config in cache was not updated")
		}
		return nil
	}, backoff.RetryEvery(time.Second)); err != nil {
		return nil, err
	}

	return &auth.SetConfigurationResponse{}, nil
}

func (a *apiServer) ExtractAuthTokens(ctx context.Context, req *auth.ExtractAuthTokensRequest) (resp *auth.ExtractAuthTokensResponse, retErr error) {
	// We don't want to actually log the request/response since they contain
	// credentials.
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(nil, nil, retErr, time.Since(start)) }(time.Now())
	if err := a.isActive(); err != nil {
		return nil, err
	}

	extracted := make([]*auth.HashedAuthToken, 0)

	tokens := a.tokens.ReadOnly(ctx)
	var val auth.TokenInfo
	if err := tokens.List(&val, col.DefaultOptions, func(hash string) error {
		// Only extract robot tokens
		if !strings.HasPrefix(val.Subject, auth.RobotPrefix) {
			return nil
		}

		ttl, err := tokens.TTL(hash)
		if err != nil {
			return err
		}

		token := &auth.HashedAuthToken{
			HashedToken: hash,
			TokenInfo: &auth.TokenInfo{
				Subject: val.Subject,
				Source:  val.Source,
			},
		}
		if ttl != -1 {
			expiration, err := types.TimestampProto(time.Now().Add(time.Second * time.Duration(ttl)))
			if err != nil {
				return err
			}
			token.Expiration = expiration
		}
		extracted = append(extracted, token)
		return nil
	}); err != nil {
		return nil, err
	}

	return &auth.ExtractAuthTokensResponse{Tokens: extracted}, nil
}

func (a *apiServer) RestoreAuthToken(ctx context.Context, req *auth.RestoreAuthTokenRequest) (resp *auth.RestoreAuthTokenResponse, retErr error) {
	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.LogResp(nil, nil, retErr, time.Since(start)) }(time.Now())

	var ttl int64
	if req.Token.Expiration != nil {
		ts, err := types.TimestampFromProto(req.Token.Expiration)
		if err != nil {
			return nil, err
		}
		ttl = int64(time.Until(ts).Seconds())
		if ttl < 0 {
			return nil, auth.ErrExpiredToken
		}
	}

	// Check whether the token hash already exists - we don't want to replace an existing token
	if _, err := col.NewSTM(ctx, a.env.GetEtcdClient(), func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)
		var existing auth.TokenInfo
		err := tokens.Get(req.Token.HashedToken, &existing)
		if err == nil {
			return errors.New("cannot overwrite existing token with same hash")
		} else if err != nil && !col.IsErrNotFound(err) {
			return err
		}

		return tokens.PutTTL(req.Token.HashedToken,
			req.Token.TokenInfo,
			ttl)
	}); err != nil {
		return nil, errors.Wrapf(err, "error restoring auth token")
	}
	return &auth.RestoreAuthTokenResponse{}, nil
}
