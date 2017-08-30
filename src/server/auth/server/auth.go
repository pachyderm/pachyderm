package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"
	"github.com/google/go-github/github"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/pachyderm/pachyderm/src/client"
	authclient "github.com/pachyderm/pachyderm/src/client/auth"
	enterpriseclient "github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
	"google.golang.org/grpc"
)

const (
	// DisableAuthenticationEnvVar specifies an environment variable that, if set, causes
	// Pachyderm authentication to ignore github and authmatically generate a
	// pachyderm token for any username in the AuthenticateRequest.GithubToken field
	DisableAuthenticationEnvVar = "PACHYDERM_AUTHENTICATION_DISABLED_FOR_TESTING"

	tokensPrefix = "/tokens"
	aclsPrefix   = "/acls"
	adminsPrefix = "/admins"

	defaultTokenTTLSecs = 14 * 24 * 60 * 60 // two weeks

	// magicUser is a special, unrevokable cluster administrator. It's not
	// possible to log in as magicUser, but pipelines with no owner are run as
	// magicUser when auth is activated.
	magicUser = `GZD4jKDGcirJyWQt6HtK4hhRD6faOofP1mng34xNZsI`
)

// epsilon is small, nonempty protobuf to use as an etcd value (the etcd client
// library can't distinguish between empty values and missing values, even
// though empty values are still stored in etcd)
var epsilon = &types.BoolValue{Value: true}

// authEnterpriseClient contains address and connection info needed to check
// the cluster's enterprise status (auth is an enterprise feature, and cannot
// be used if the cluster doesn't have enterprise features enabled)
type enterpriseClient struct {
	clientOnce sync.Once
	address    string
	conn       *grpc.ClientConn
	clientErr  error
	client     enterpriseclient.APIClient
}

type apiServer struct {
	pachLogger log.Logger
	enterprise enterpriseClient // for checking enterprise token status
	etcdClient *etcd.Client

	// admins is a cache of the current cluster administrators
	adminMu    sync.Mutex
	adminCache map[string]struct{}

	// tokens is a collection of hashedToken -> User mappings.
	tokens col.Collection
	// acls is a collection of repoName -> ACL mappings.
	acls col.Collection
	// admins is a collection of username -> Empty mappings (keys indicate which
	// github users are cluster admins)
	admins col.Collection
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
	} else if authclient.IsNotActivatedError(err) {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.DebugLevel, 4)
	} else {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.ErrorLevel, 4)
	}
}

// NewAuthServer returns an implementation of authclient.APIServer.
func NewAuthServer(pachdAddress string, etcdAddress string, etcdPrefix string) (authclient.APIServer, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return nil, fmt.Errorf("error constructing etcdClient: %s", err.Error())
	}

	s := &apiServer{
		pachLogger: log.NewLogger("authclient.API"),
		etcdClient: etcdClient,
		enterprise: enterpriseClient{
			address: pachdAddress,
		},
		adminCache: make(map[string]struct{}),
		tokens: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, tokensPrefix),
			nil,
			&authclient.User{},
			nil,
		),
		acls: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, aclsPrefix),
			nil,
			&authclient.ACL{},
			nil,
		),
		admins: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, adminsPrefix),
			nil,
			&types.BoolValue{}, // typeof(epsilon) == types.BoolValue; epsilon is the only value
			nil,
		),
	}
	go s.getEnterpriseTokenState() // initialize connection to enterprise API
	go s.watchAdmins(path.Join(etcdPrefix, adminsPrefix))
	return s, nil
}

func (a *apiServer) watchAdmins(fullAdminPrefix string) {
	backoff.RetryNotify(func() error {
		// Watch for the addition/removal of new admins. Note that this will return
		// any existing admins, so if the auth service is already activated, it will
		// stay activated.
		watcher, err := a.admins.ReadOnly(context.Background()).Watch()
		if err != nil {
			return err
		}
		defer watcher.Close()
		// The auth service is activated if we have admins, and not
		// activated otherwise.
		for {
			ev, ok := <-watcher.Watch()
			if !ok {
				return errors.New("admin watch closed unexpectedly")
			}

			if err := func() error {
				// Lock a.adminMu in case we need to modify a.adminCache
				a.adminMu.Lock()
				defer a.adminMu.Unlock()

				// Parse event data and potentially update adminCache
				var key string
				var boolProto types.BoolValue
				ev.Unmarshal(&key, &boolProto)
				username := strings.TrimPrefix(key, fullAdminPrefix+"/")
				switch ev.Type {
				case watch.EventPut:
					a.adminCache[username] = struct{}{}
				case watch.EventDelete:
					delete(a.adminCache, username)
				case watch.EventError:
					return ev.Err
				}
				return nil // unlock mu
			}(); err != nil {
				return err
			}
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logrus.Printf("error from activation check: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *apiServer) getEnterpriseTokenState() (enterpriseclient.State, error) {
	a.enterprise.clientOnce.Do(func() {
		a.enterprise.conn, a.enterprise.clientErr = grpc.Dial(a.enterprise.address, client.PachDialOptions()...)
		a.enterprise.client = enterpriseclient.NewAPIClient(a.enterprise.conn)
	})
	if a.enterprise.clientErr != nil {
		return 0, a.enterprise.clientErr
	}
	resp, err := a.enterprise.client.GetState(context.Background(),
		&enterpriseclient.GetStateRequest{})
	if err != nil {
		return 0, err
	}
	return resp.State, nil
}

func (a *apiServer) Activate(ctx context.Context, req *authclient.ActivateRequest) (resp *authclient.ActivateResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return nil, fmt.Errorf("error confirming Pachyderm Enterprise token: %s", err.Error())
	}
	if state != enterpriseclient.State_ACTIVE {
		return nil, fmt.Errorf("Pachyderm Enterprise is not active in this cluster")
	}

	// Activating an already activated auth service should fail, because
	// otherwise anyone can just activate the service again and set
	// themselves as an admin.
	if a.isActivated() {
		return nil, fmt.Errorf("already activated")
	}

	// Initialize admins (watchAdmins() above will see the write)
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		admins := a.admins.ReadWrite(stm)
		for _, user := range req.Admins {
			admins.Put(user, epsilon)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &authclient.ActivateResponse{}, nil
}

func (a *apiServer) Deactivate(ctx context.Context, req *authclient.DeactivateRequest) (resp *authclient.DeactivateResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}

	// Get calling user. The user must be a cluster admin to disable auth for the
	// cluster
	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if !a.isAdmin(user.Username) {
		return nil, fmt.Errorf("must be an admin to disable cluster auth")
	}
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		a.acls.ReadWrite(stm).DeleteAll()
		a.tokens.ReadWrite(stm).DeleteAll()
		a.admins.ReadWrite(stm).DeleteAll()  // watchAdmins() will see the write
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &authclient.DeactivateResponse{}, nil
}

func (a *apiServer) isActivated() bool {
	a.adminMu.Lock()
	defer a.adminMu.Unlock()
	return len(a.adminCache) > 0
}

// AccessTokenToUsername takes a OAuth access token issued by GitHub and uses
// it discover the username of the user who obtained the code. This is how
// Pachyderm currently implements authorization in a production cluster
func AccessTokenToUsername(ctx context.Context, token string) (string, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	gclient := github.NewClient(tc)

	// Passing the empty string gets us the authenticated user
	user, _, err := gclient.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("error getting the authenticated user: %s", err.Error())
	}
	return user.GetName(), nil
}

func (a *apiServer) GetAdmins(ctx context.Context, req *authclient.GetAdminsRequest) (resp *authclient.GetAdminsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}

	// Get calling user. There is no auth check to see the list of cluster admins,
	// other than that the user must log in. Otherwise how will users know who to
	// ask for admin privileges? Requiring the user to be logged in mitigates
	// phishing
	_, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	a.adminMu.Lock()
	defer a.adminMu.Unlock()
	resp = &authclient.GetAdminsResponse{
		Admins: make([]string, 0, len(a.adminCache)),
	}
	for admin := range a.adminCache {
		resp.Admins = append(resp.Admins, admin)
	}
	return resp, nil
}

func (a *apiServer) validateModifyAdminsRequest(req *authclient.ModifyAdminsRequest) error {
	// Check to make sure that req doesn't remove all cluster admins
	m := make(map[string]struct{})
	// copy existing admins into m
	func() {
		a.adminMu.Lock()
		defer a.adminMu.Unlock()
		for u := range a.adminCache {
			m[u] = struct{}{}
		}
	}()
	for _, u := range req.Add {
		m[u] = struct{}{}
	}
	for _, u := range req.Remove {
		delete(m, u)
	}
	if len(m) == 0 {
		return fmt.Errorf("invalid request: cannot remove all cluster administrators while auth is active, to avoid unfixable cluster states")
	}
	return nil
}

func (a *apiServer) ModifyAdmins(ctx context.Context, req *authclient.ModifyAdminsRequest) (resp *authclient.ModifyAdminsResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}

	// Get calling user. The user must be an admin to change the list of admins
	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	if !a.isAdmin(user.Username) {
		return nil, fmt.Errorf("must be an admin to modify set of cluster admins")
	}
	if err := a.validateModifyAdminsRequest(req); err != nil {
		return nil, err
	}

	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		admins := a.admins.ReadWrite(stm)
		// Update "admins" list (watchAdmins() will update admins cache)
		for _, user := range req.Add {
			admins.Put(user, epsilon)
		}
		for _, user := range req.Remove {
			admins.Delete(user)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &authclient.ModifyAdminsResponse{}, nil
}

func (a *apiServer) Authenticate(ctx context.Context, req *authclient.AuthenticateRequest) (resp *authclient.AuthenticateResponse, retErr error) {
	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.LogResp(nil, nil, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}
	if req.GithubUsername == magicUser {
		return nil, fmt.Errorf("invalid user")
	}

	var username string
	if os.Getenv(DisableAuthenticationEnvVar) == "true" {
		// Test mode--the caller automatically authenticates as whoever is requested
		username = req.GithubUsername
	} else {
		// Prod mode--send access code to GitHub to discover authenticating user
		var err error
		username, err = AccessTokenToUsername(ctx, req.GithubToken)
		if err != nil {
			return nil, err
		}
		if req.GithubUsername != "" && req.GithubUsername != username {
			return nil, fmt.Errorf("attempted to authenticate as %s, but Github " +
				"token did not originate from that account")
		}
	}

	pachToken := uuid.NewWithoutDashes()
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)
		return tokens.PutTTL(hashToken(pachToken), &authclient.User{username},
			defaultTokenTTLSecs)
	})
	if err != nil {
		return nil, fmt.Errorf("error storing auth token for user %v: %s", username, err.Error())
	}

	return &authclient.AuthenticateResponse{
		PachToken: pachToken,
	}, nil
}

func (a *apiServer) Authorize(ctx context.Context, req *authclient.AuthorizeRequest) (resp *authclient.AuthorizeResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}

	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	if a.isAdmin(user.Username) {
		// admins are always authorized
		return &authclient.AuthorizeResponse{
			Authorized: true,
		}, nil
	}

	var acl authclient.ACL
	if err := a.acls.ReadOnly(ctx).Get(req.Repo, &acl); err != nil {
		if _, ok := err.(col.ErrNotFound); ok {
			// Return a special error instead of a generic "not authorized" response
			// in case a consistency error has occurred and an admin needs to fix the
			// repo--explaining that the ACL is gone will help debug
			return nil, fmt.Errorf("ACL not found for repo %v", req.Repo)
		}
		return nil, fmt.Errorf("error getting ACL for repo %v: %s", req.Repo, err.Error())
	}

	return &authclient.AuthorizeResponse{
		Authorized: req.Scope <= acl.Entries[user.Username],
	}, nil
}

func (a *apiServer) WhoAmI(ctx context.Context, req *authclient.WhoAmIRequest) (resp *authclient.WhoAmIResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}

	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	return &authclient.WhoAmIResponse{
		Username: user.Username,
	}, nil
}

func validateSetScopeRequest(req *authclient.SetScopeRequest) error {
	if req.Username == "" {
		return fmt.Errorf("invalid request: must set username")
	}
	if req.Repo == "" {
		return fmt.Errorf("invalid request: must set repo")
	}
	return nil
}

func (a *apiServer) isAdmin(user string) bool {
	if user == magicUser {
		return true
	}
	a.adminMu.Lock()
	defer a.adminMu.Unlock()
	_, ok := a.adminCache[user]
	return ok
}

func (a *apiServer) SetScope(ctx context.Context, req *authclient.SetScopeRequest) (resp *authclient.SetScopeResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}

	if err := validateSetScopeRequest(req); err != nil {
		return nil, err
	}
	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	admin := a.isAdmin(user.Username) // Check if the caller is an admin

	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		acls := a.acls.ReadWrite(stm)
		var acl authclient.ACL
		if err := acls.Get(req.Repo, &acl); err != nil {
			// TODO(msteffen): ACL not found; check that the repo exists?
			acl.Entries = make(map[string]authclient.Scope)
		}
		switch {
		case req.Username == user.Username && req.Scope == authclient.Scope_OWNER:
			// Special case: creating a new ACL. This is allowed
			// TODO(msteffen): remove this case and inline this into CreateRepo
		case admin:
			// Admins can fix empty ACLs
		case acl.Entries[user.Username] == authclient.Scope_OWNER:
			// user is an owner, and is authorized to modify the ACL
		case len(acl.Entries) == 0:
			// warn user that there is no ACL for this repo
			return fmt.Errorf("ACL not found for repo %v", req.Repo)
		default:
			return &authclient.NotAuthorizedError{
				Repo:     req.Repo,
				Required: authclient.Scope_OWNER,
			}
		}
		if req.Scope != authclient.Scope_NONE {
			acl.Entries[req.Username] = req.Scope
		} else {
			delete(acl.Entries, req.Username)
		}
		acls.Put(req.Repo, &acl)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &authclient.SetScopeResponse{}, nil
}

func (a *apiServer) GetScope(ctx context.Context, req *authclient.GetScopeRequest) (resp *authclient.GetScopeResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}

	var username string
	if req.Username != "" {
		username = req.Username
	} else {
		user, err := a.getAuthenticatedUser(ctx)
		if err != nil {
			return nil, err
		}
		username = user.Username
	}

	// For now, we don't return OWNER if the user is an admin, even though that's
	// their effective access scope for all repos--the caller may want to know
	// what will happen if the user's admin privileges are revoked

	// Read repo ACL from etcd
	resp = new(authclient.GetScopeResponse)
	for _, repo := range req.Repos {
		var acl authclient.ACL
		err := a.acls.ReadOnly(ctx).Get(repo, &acl)
		if err != nil || acl.Entries == nil {
			// ACL not found. User has no scope
			resp.Scopes = append(resp.Scopes, authclient.Scope_NONE)
		} else {
			resp.Scopes = append(resp.Scopes, acl.Entries[username])
		}
	}
	return resp, nil
}

func (a *apiServer) GetACL(ctx context.Context, req *authclient.GetACLRequest) (resp *authclient.GetACLResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}

	// Validate request
	if req.Repo == "" {
		return nil, fmt.Errorf("invalid request: must provide name of repo to get that repo's ACL")
	}

	// Get calling user
	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// Read repo ACL from etcd
	resp = &authclient.GetACLResponse{
		ACL: &authclient.ACL{},
	}
	if err = a.acls.ReadOnly(ctx).Get(req.Repo, resp.ACL); err != nil {
		if _, ok := err.(col.ErrNotFound); !ok {
			return nil, err
		}
		// else: ACL not found. No error, just return an empty ACL
	}
	// For now, require READER access to read repo metadata (commits, and ACLs)
	if !a.isAdmin(user.Username) && resp.ACL.Entries[user.Username] < authclient.Scope_READER {
		return nil, &authclient.NotAuthorizedError{
			Repo:     req.Repo,
			Required: authclient.Scope_READER,
		}
	}
	return resp, nil
}

func (a *apiServer) SetACL(ctx context.Context, req *authclient.SetACLRequest) (resp *authclient.SetACLResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}

	// Validate request
	if req.Repo == "" {
		return nil, fmt.Errorf("invalid request: must provide name of repo you want to modify")
	}

	// Get calling user
	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// Read repo ACL from etcd
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		acls := a.acls.ReadWrite(stm)

		// Require OWNER access to modify repo ACL
		var acl authclient.ACL
		acls.Get(req.Repo, &acl)
		if !a.isAdmin(user.Username) && acl.Entries[user.Username] < authclient.Scope_OWNER {
			return &authclient.NotAuthorizedError{
				Repo:     req.Repo,
				Required: authclient.Scope_OWNER,
			}
		}

		// Set new ACL
		if req.NewACL == nil || len(req.NewACL.Entries) == 0 {
			return acls.Delete(req.Repo)
		}
		return acls.Put(req.Repo, req.NewACL)
	})
	if err != nil {
		return nil, fmt.Errorf("could not put new ACL: %s", err.Error())
	}
	return &authclient.SetACLResponse{}, nil
}

func (a *apiServer) GetCapability(ctx context.Context, req *authclient.GetCapabilityRequest) (resp *authclient.GetCapabilityResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	var user *authclient.User
	if !a.isActivated() {
		// If auth service is not activated, we want to return a capability
		// that's able to access any repo.  That way, when we create a
		// pipeline, we can assign it with a capability that would allow
		// it to access any repo after the auth service has been activated.
		user = &authclient.User{magicUser}
	} else {
		var err error
		user, err = a.getAuthenticatedUser(ctx)
		if err != nil {
			return nil, err
		}
	}

	capability := uuid.NewWithoutDashes()
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)
		// Capabilities are forever; they don't expire.
		return tokens.Put(hashToken(capability), user)
	})
	if err != nil {
		return nil, fmt.Errorf("error storing capability for user %v: %s", user.Username, err.Error())
	}

	return &authclient.GetCapabilityResponse{
		Capability: capability,
	}, nil
}

func (a *apiServer) RevokeAuthToken(ctx context.Context, req *authclient.RevokeAuthTokenRequest) (resp *authclient.RevokeAuthTokenResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.isActivated() {
		return nil, authclient.NotActivatedError{}
	}

	// Even though anyone can revoke anyone's auth token, we still want
	// the user to be authenticated.
	if _, err := a.getAuthenticatedUser(ctx); err != nil {
		return nil, err
	}

	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)
		// Capabilities are forver; they don't expire.
		if err := tokens.Delete(hashToken(req.Token)); err != nil {
			// We ignore NotFound errors, since it's ok to revoke a
			// nonexistent token.
			if _, ok := err.(col.ErrNotFound); !ok {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error revoking token: %s", err.Error())
	}

	return &authclient.RevokeAuthTokenResponse{}, nil
}

// hashToken converts a token to a cryptographic hash.
// We don't want to store tokens verbatim in the database, as then whoever
// that has access to the database has access to all tokens.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}

func (a *apiServer) getAuthenticatedUser(ctx context.Context) (*authclient.User, error) {
	// TODO(msteffen) cache these lookups, especially since users always authorize
	// themselves at the beginning of a request. Don't want to look up the same
	// token -> username entry twice.
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no authentication metadata found in context")
	}
	if len(md[authclient.ContextTokenKey]) != 1 {
		return nil, fmt.Errorf("auth token not found in context")
	}
	token := md[authclient.ContextTokenKey][0]

	var user authclient.User
	if err := a.tokens.ReadOnly(ctx).Get(hashToken(token), &user); err != nil {
		if _, ok := err.(col.ErrNotFound); ok {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("error getting token: %s", err.Error())
	}

	return &user, nil
}
