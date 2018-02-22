package server

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
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
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	authclient "github.com/pachyderm/pachyderm/src/client/auth"
	enterpriseclient "github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/watch"
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

	// githubPrefix is a prefix we prepend to Users in the 'tokens' collection
	// and ACL Entries (i.e. all usernames that have been verified with GitHub)
	// to indicate that they're GitHub usernames. Right now, all users are GitHub
	// usernames, but someday they may be groups or LDAP users
	githubPrefix = "github:"
)

// epsilon is small, nonempty protobuf to use as an etcd value (the etcd client
// library can't distinguish between empty values and missing values, even
// though empty values are still stored in etcd)
var epsilon = &types.BoolValue{Value: true}

type apiServer struct {
	pachLogger log.Logger
	etcdClient *etcd.Client

	address        string            // address of a Pachd server
	pachClient     *client.APIClient // pachd client
	pachClientOnce sync.Once         // used to initialize pachClient
	clientErr      error             // set if initializing pachClient fails

	adminCache map[string]struct{} // cache of current cluster admins
	adminMu    sync.Mutex          // synchronize ontrol access to adminCache

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

func (a *apiServer) getPachClient() (*client.APIClient, error) {
	a.pachClientOnce.Do(func() {
		a.pachClient, a.clientErr = client.NewFromAddress(a.address)
	})
	if a.clientErr != nil {
		return nil, a.clientErr
	}
	return a.pachClient, nil
}

// NewAuthServer returns an implementation of authclient.APIServer.
func NewAuthServer(pachdAddress string, etcdAddress string, etcdPrefix string) (authclient.APIServer, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return nil, fmt.Errorf("error constructing etcdClient: %v", err)
	}

	s := &apiServer{
		pachLogger: log.NewLogger("authclient.API"),
		etcdClient: etcdClient,
		address:    pachdAddress,
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
	go s.getPachClient() // initialize connection to Pachd
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
	pachClient, err := a.getPachClient()
	if err != nil {
		return 0, fmt.Errorf("could not get Pachd client to determine Enterprise status: %s", err)
	}
	resp, err := pachClient.Enterprise.GetState(context.Background(),
		&enterpriseclient.GetStateRequest{})
	if err != nil {
		return 0, fmt.Errorf("could not get Enterprise status: %v", grpcutil.ScrubGRPC(err))
	}
	return resp.State, nil
}

func (a *apiServer) Activate(ctx context.Context, req *authclient.ActivateRequest) (resp *authclient.ActivateResponse, retErr error) {
	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.LogResp(nil, nil, retErr, time.Since(start)) }(time.Now())
	if req.GithubUsername == magicUser {
		return nil, fmt.Errorf("invalid user")
	}

	// If the cluster's Pachyderm Enterprise token isn't active, the auth system
	// cannot be activated
	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return nil, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
	}
	if state != enterpriseclient.State_ACTIVE {
		return nil, fmt.Errorf("Pachyderm Enterprise is not active in this " +
			"cluster, and the Pachyderm auth API is an Enterprise-level feature")
	}

	// Activating an already activated auth service should fail, because
	// otherwise anyone can just activate the service again and set
	// themselves as an admin.
	if a.isActivated() {
		return nil, fmt.Errorf("already activated")
	}

	// Determine caller's Pachyderm/GitHub username
	username, err := GitHubTokenToUsername(ctx, req.GithubUsername, req.GithubToken)
	if err != nil {
		return nil, err
	}

	// Generate a new Pachyderm token (as the caller is authenticating) and
	// initialize admins (watchAdmins() above will see the write)
	pachToken := uuid.NewWithoutDashes()
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		admins := a.admins.ReadWrite(stm)
		tokens := a.tokens.ReadWrite(stm)
		if err := admins.Put(username, epsilon); err != nil {
			return err
		}
		return tokens.PutTTL(
			hashToken(pachToken),
			&authclient.User{Username: username, Type: authclient.User_GITHUB},
			defaultTokenTTLSecs,
		)
	})
	if err != nil {
		return nil, err
	}
	return &authclient.ActivateResponse{PachToken: pachToken}, nil
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
		return nil, errors.New("not authorized to deactivate auth, must be a cluster admin")
	}
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		a.acls.ReadWrite(stm).DeleteAll()
		a.tokens.ReadWrite(stm).DeleteAll()
		a.admins.ReadWrite(stm).DeleteAll() // watchAdmins() will see the write
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

// GitHubTokenToUsername takes a OAuth access token issued by GitHub and uses
// it discover the username of the user who obtained the code (or verify that
// the code belongs to githubUsername). This is how Pachyderm currently
// implements authorization in a production cluster
func GitHubTokenToUsername(ctx context.Context, githubUsername string, token string) (string, error) {
	if os.Getenv(DisableAuthenticationEnvVar) == "true" && githubUsername != "" {
		// Test mode--the caller automatically authenticates as whoever is requested
		return githubPrefix + githubUsername, nil
	}

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
		return "", fmt.Errorf("error getting the authenticated user: %v", err)
	}
	verifiedUsername := user.GetLogin()
	if githubUsername != "" && githubUsername != verifiedUsername {
		return "", fmt.Errorf("attempted to authenticate as %s, but Github "+
			"token did not originate from that account", githubUsername)
	}
	return githubPrefix + verifiedUsername, nil
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
		resp.Admins = append(resp.Admins, strings.TrimPrefix(admin, githubPrefix))
	}
	return resp, nil
}

func (a *apiServer) validateModifyAdminsRequest(add []string, remove []string) error {
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
	for _, u := range add {
		m[u] = struct{}{}
	}
	for _, u := range remove {
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
		return nil, errors.New("not authorized to modify cluster admins, must be a cluster admin")
	}

	// Canonicalize GitHub usernames in request (must canonicalize before we can
	// validate, so we know who is actually being added/removed & can confirm
	// that not all admins are being removed)
	eg := &errgroup.Group{}
	canonicalizedToAdd := make([]string, len(req.Add))
	for i, user := range req.Add {
		eg.Go(func() error {
			i, user := i, user
			u, err := canonicalizeGitHubUsername(ctx, user)
			if err != nil {
				return err
			}
			canonicalizedToAdd[i] = u
			return nil
		})
	}
	canonicalizedToRemove := make([]string, len(req.Remove))
	for i, user := range req.Remove {
		eg.Go(func() error {
			i, user := i, user
			u, err := canonicalizeGitHubUsername(ctx, user)
			if err != nil {
				return err
			}
			canonicalizedToRemove[i] = u
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	err = a.validateModifyAdminsRequest(canonicalizedToAdd, canonicalizedToRemove)
	if err != nil {
		return nil, err
	}

	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		admins := a.admins.ReadWrite(stm)
		// Update "admins" list (watchAdmins() will update admins cache)
		for _, user := range canonicalizedToAdd {
			admins.Put(user, epsilon)
		}
		for _, user := range canonicalizedToRemove {
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

	// Determine caller's Pachyderm/GitHub username
	username, err := GitHubTokenToUsername(ctx, req.GithubUsername, req.GithubToken)
	if err != nil {
		return nil, err
	}

	// If the cluster's enterprise token is expired, only admins may log in
	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return nil, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
	}
	if state != enterpriseclient.State_ACTIVE && !a.isAdmin(username) {
		return nil, errors.New("Pachyderm Enterprise is not active in this " +
			"cluster (until Pachyderm Enterprise is re-activated or Pachyderm " +
			"auth is deactivated, only cluster admins can perform any operations)")
	}

	// Generate a new Pachyderm token and return it
	pachToken := uuid.NewWithoutDashes()
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)
		return tokens.PutTTL(hashToken(pachToken),
			&authclient.User{
				Username: username,
				Type:     authclient.User_GITHUB,
			},
			defaultTokenTTLSecs)
	})
	if err != nil {
		return nil, fmt.Errorf("error storing auth token for user \"%s\": %v", username, err)
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

	// admins are always authorized
	if a.isAdmin(user.Username) {
		return &authclient.AuthorizeResponse{Authorized: true}, nil
	}

	// If the cluster's enterprise token is expired, only admins and pipelines may
	// authorize (and admins are already handled)
	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return nil, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
	}
	if state != enterpriseclient.State_ACTIVE && user.Type != authclient.User_PIPELINE {
		return nil, fmt.Errorf("Pachyderm Enterprise is not active in this " +
			"cluster (only a cluster admin can authorize)")
	}

	// Get ACL to check
	var acl authclient.ACL
	if err := a.acls.ReadOnly(ctx).Get(req.Repo, &acl); err != nil && !col.IsErrNotFound(err) {
		return nil, fmt.Errorf("error getting ACL for repo \"%s\": %v", req.Repo, err)
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
		Username: strings.TrimPrefix(user.Username, githubPrefix),
		IsAdmin:  a.isAdmin(user.Username),
	}, nil
}

func validateSetScopeRequest(ctx context.Context, req *authclient.SetScopeRequest) error {
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

	if err := validateSetScopeRequest(ctx, req); err != nil {
		return nil, err
	}
	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		acls := a.acls.ReadWrite(stm)
		var acl authclient.ACL
		if err := acls.Get(req.Repo, &acl); err != nil {
			if !col.IsErrNotFound(err) {
				return err
			}
			// TODO(msteffen): ACL not found; check that the repo exists?
			acl.Entries = make(map[string]authclient.Scope)
		}
		authorized, err := func() (bool, error) {
			if a.isAdmin(user.Username) {
				// admins are automatically authorized
				return true, nil
			}

			// Check if the cluster's enterprise token is expired (fail if so)
			state, err := a.getEnterpriseTokenState()
			if err != nil {
				return false, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
			}
			if state != enterpriseclient.State_ACTIVE {
				return false, fmt.Errorf("Pachyderm Enterprise is not active in this " +
					"cluster (only a cluster admin can set a scope)")
			}

			// Check if the user is on the ACL directly
			if acl.Entries[user.Username] == authclient.Scope_OWNER {
				return true, nil
			}
			return false, nil
		}()
		if err != nil {
			return err
		}
		if !authorized {
			return &authclient.NotAuthorizedError{
				Repo:     req.Repo,
				Required: authclient.Scope_OWNER,
			}
		}

		// Scope change is authorized. Make the change
		u, err := canonicalizeGitHubUsername(ctx, req.Username)
		if err != nil {
			return err
		}
		if req.Scope != authclient.Scope_NONE {
			acl.Entries[u] = req.Scope
		} else {
			delete(acl.Entries, u)
		}
		if len(acl.Entries) == 0 {
			return acls.Delete(req.Repo)
		}
		return acls.Put(req.Repo, &acl)
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

	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	// Check if the cluster's enterprise token is expired (fail if so)
	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return nil, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
	}
	if state != enterpriseclient.State_ACTIVE && !a.isAdmin(user.Username) {
		return nil, fmt.Errorf("Pachyderm Enterprise is not active in this " +
			"cluster (only a cluster admin can perform any operations)")
	}

	// For now, we don't return OWNER if the user is an admin, even though that's
	// their effective access scope for all repos--the caller may want to know
	// what will happen if the user's admin privileges are revoked

	// Read repo ACL from etcd
	acls := a.acls.ReadOnly(ctx)
	resp = new(authclient.GetScopeResponse)

	for _, repo := range req.Repos {
		var acl authclient.ACL
		if err := acls.Get(repo, &acl); err != nil && !col.IsErrNotFound(err) {
			return nil, err
		}
		if req.Username == "" {
			resp.Scopes = append(resp.Scopes, acl.Entries[user.Username])
		} else {
			if !a.isAdmin(user.Username) && acl.Entries[user.Username] < authclient.Scope_READER {
				return nil, &authclient.NotAuthorizedError{
					Repo:     repo,
					Required: authclient.Scope_READER,
				}
			}
			u, err := canonicalizeGitHubUsername(ctx, req.Username)
			if err != nil {
				return nil, err
			}
			resp.Scopes = append(resp.Scopes, acl.Entries[u])
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

	// Check if the cluster's enterprise token is expired (fail if so)
	state, err := a.getEnterpriseTokenState()
	if err != nil {
		return nil, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
	}
	if state != enterpriseclient.State_ACTIVE && !a.isAdmin(user.Username) {
		return nil, fmt.Errorf("Pachyderm Enterprise is not active in this " +
			"cluster (only a cluster admin can perform any operations)")
	}

	// Read repo ACL from etcd
	acl := &authclient.ACL{}
	if err = a.acls.ReadOnly(ctx).Get(req.Repo, acl); err != nil && !col.IsErrNotFound(err) {
		return nil, err
	}
	resp = &authclient.GetACLResponse{
		Entries: make([]*authclient.ACLEntry, 0),
	}
	for user, scope := range acl.Entries {
		resp.Entries = append(resp.Entries, &authclient.ACLEntry{
			Username: strings.TrimPrefix(user, githubPrefix),
			Scope:    scope,
		})
	}
	// For now, no access is require to read a repo's ACL
	// https://github.com/pachyderm/pachyderm/issues/2353
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

	// Canonicalize entries in the request (must have canonical request before we
	// can authorize, as we inspect the ACL contents in authorization)
	eg := &errgroup.Group{}
	var aclMu sync.Mutex
	newACL := new(authclient.ACL)
	if len(req.Entries) > 0 {
		newACL.Entries = make(map[string]authclient.Scope)
	}
	for _, entry := range req.Entries {
		user, scope := entry.Username, entry.Scope
		eg.Go(func() error {
			user, scope := user, scope
			u, err := canonicalizeGitHubUsername(ctx, user)
			if err != nil {
				return err
			}
			aclMu.Lock()
			defer aclMu.Unlock()
			newACL.Entries[u] = scope
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Read repo ACL from etcd
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		acls := a.acls.ReadWrite(stm)

		// determine if the caller is authorized to set this repo's ACL
		authorized, err := func() (bool, error) {
			if a.isAdmin(user.Username) {
				// admins are automatically authorized
				return true, nil
			}

			// Check if the cluster's enterprise token is expired (fail if so)
			state, err := a.getEnterpriseTokenState()
			if err != nil {
				return false, fmt.Errorf("error confirming Pachyderm Enterprise token: %v", err)
			}
			if state != enterpriseclient.State_ACTIVE {
				return false, fmt.Errorf("Pachyderm Enterprise is not active in this " +
					"cluster (only a cluster admin can modify an ACL)")
			}

			// Check if there is an existing ACL, and if the user is on it
			var acl authclient.ACL
			if err := acls.Get(req.Repo, &acl); err != nil {
				// ACL not found -- construct empty ACL proto
				acl.Entries = make(map[string]authclient.Scope)
			}
			if len(acl.Entries) > 0 {
				// ACL is present; caller must be authorized directly
				if acl.Entries[user.Username] == authclient.Scope_OWNER {
					return true, nil
				}
				return false, nil
			}

			// No ACL -- check if the repo being modified exists
			pachClient, err := a.getPachClient()
			if err != nil {
				return false, fmt.Errorf("could not check if repo \"%s\" exists: %v", req.Repo, err)
			}
			_, err = pachClient.InspectRepo(req.Repo)
			err = grpcutil.ScrubGRPC(err)
			if err == nil {
				// Repo exists -- user isn't authorized
				return false, nil
			} else if !strings.HasSuffix(err.Error(), "not found") {
				// Unclear if repo exists -- return error
				return false, fmt.Errorf("could not inspect \"%s\": %v", req.Repo, err)
			} else if len(newACL.Entries) == 1 &&
				newACL.Entries[user.Username] == authclient.Scope_OWNER {
				// Special case: Repo doesn't exist, but user is creating a new Repo, and
				// making themself the owner, e.g. for CreateRepo or CreatePipeline, then
				// the request is authorized
				return true, nil
			}
			return false, err
		}()
		if err != nil {
			return err
		}
		if !authorized {
			return &authclient.NotAuthorizedError{
				Repo:     req.Repo,
				Required: authclient.Scope_OWNER,
			}
		}

		// Set new ACL
		if len(newACL.Entries) == 0 {
			return acls.Delete(req.Repo)
		}
		return acls.Put(req.Repo, newACL)
	})
	if err != nil {
		return nil, fmt.Errorf("could not put new ACL: %v", err)
	}
	return &authclient.SetACLResponse{}, nil
}

func (a *apiServer) GetCapability(ctx context.Context, req *authclient.GetCapabilityRequest) (resp *authclient.GetCapabilityResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	// Generate User that the capability token will point to
	var user *authclient.User
	if !a.isActivated() {
		// If auth service is not activated, we want to return a capability
		// that's able to access any repo.  That way, when we create a
		// pipeline, we can assign it with a capability that would allow
		// it to access any repo after the auth service has been activated.
		user = &authclient.User{Username: magicUser}
	} else {
		var err error
		user, err = a.getAuthenticatedUser(ctx)
		if err != nil {
			return nil, err
		}
	}
	// currently, GetCapability is only called by CreatePipeline
	// TODO(msteffen): Only expose this inside the cluster
	user.Type = authclient.User_PIPELINE

	capability := uuid.NewWithoutDashes()
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)
		// Capabilities are forever; they don't expire.
		return tokens.Put(hashToken(capability), user)
	})
	if err != nil {
		return nil, fmt.Errorf("error storing capability for user \"%s\": %v", user.Username, err)
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

	if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)
		user := authclient.User{}
		if err := tokens.Get(hashToken(req.Token), &user); err != nil && !col.IsErrNotFound(err) {
			return err
		}
		if user.Type != authclient.User_PIPELINE {
			return fmt.Errorf("cannot revoke a non-pipeline auth token")
		}
		return tokens.Delete(hashToken(req.Token))
	}); err != nil {
		return nil, err
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
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no authentication metadata found in context")
	}
	if len(md[authclient.ContextTokenKey]) != 1 {
		return nil, authclient.NotSignedInError{}
	}
	token := md[authclient.ContextTokenKey][0]

	var user authclient.User
	if err := a.tokens.ReadOnly(ctx).Get(hashToken(token), &user); err != nil {
		if col.IsErrNotFound(err) {
			// This error message string is matched in the UI. If edited,
			// it also needs to be updated in the UI code
			return nil, fmt.Errorf("provided auth token is corrupted or has expired (try logging in again)")
		}
		return nil, fmt.Errorf("error getting token: %v", err)
	}
	return &user, nil
}

// canonicalizeGitHubUsername corrects 'username' for case errors by looking
// up the corresponding user's GitHub profile and extracting their login ID
// from that
func canonicalizeGitHubUsername(ctx context.Context, username string) (string, error) {
	if os.Getenv(DisableAuthenticationEnvVar) == "true" {
		// authentication is off -- username might not even be real
		return githubPrefix + username, nil
	}
	gclient := github.NewClient(http.DefaultClient)
	u, _, err := gclient.Users.Get(ctx, strings.ToLower(username))
	if err != nil {
		return "", fmt.Errorf("error canonicalizing \"%s\": %v", username, err)
	}
	return githubPrefix + u.GetLogin(), nil
}
