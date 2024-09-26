// Package server implement the auth service gRPC server.
package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	enterpriseclient "github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/authdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/keycache"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	internalauth "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	authiface "github.com/pachyderm/pachyderm/v2/src/server/auth"
)

const (
	// oidc etcd object prefix
	oidcAuthnPrefix = "/oidc-authns"

	// configKey is a key (in etcd, in the config collection) that maps to the
	// auth configuration. This is the only key in that collection (due to
	// implemenation details of our config library, we can't use an empty key)
	configKey = "config"

	// the length of interval between expired auth token cleanups
	cleanupIntervalHours = 24
)

// DefaultOIDCConfig is the default config for the auth API server
var DefaultOIDCConfig = auth.OIDCConfig{}

type APIServer = *apiServer

var _ authiface.APIServer = (APIServer)(nil)

// apiServer implements the public interface of the Pachyderm auth system,
// including all RPCs defined in the protobuf spec.
type apiServer struct {
	auth.UnsafeAPIServer

	env Env

	configCache             *keycache.Cache
	clusterRoleBindingCache *keycache.Cache

	// roleBindings is a collection of resource name -> role binding mappings.
	roleBindings col.PostgresCollection
	// members is a collection of username -> groups mappings.
	members col.PostgresCollection
	// groups is a collection of group -> usernames mappings.
	groups col.PostgresCollection
	// collection containing the auth config (under the key configKey)
	authConfig col.PostgresCollection
	// oidcStates  contains the set of OIDC nonces for requests that are in progress
	oidcStates col.EtcdCollection

	// public addresses the fact that pachd in full mode initializes two auth
	// servers: one that exposes a public API, possibly over TLS, and one that
	// exposes a private API, for internal services. Only one of these can launch
	// the OIDC callback web server.
	public bool

	// watchesEnabled controls whether we cache the auth config and cluster role bindings
	// in the auth service, or whether we look them up each time. Watches are expensive in
	// postgres, so we can't afford to have each sidecar run watches. Pipelines always have
	// direct access to a repo anyways, so the cluster role bindings don't affect their access,
	// and the OIDC server doesn't run in the sidecar so the config doesn't matter.
	watchesEnabled bool
}

// NewAuthServer returns an implementation of auth.APIServer.
func NewAuthServer(env Env, public, requireNoncriticalServers, watchesEnabled bool) (*apiServer, error) {
	oidcStates := col.NewEtcdCollection(
		env.EtcdClient,
		path.Join(oidcAuthnPrefix),
		nil,
		&auth.SessionInfo{},
		nil,
		nil,
	)
	s := &apiServer{
		env:            env,
		authConfig:     authdb.AuthConfigCollection(env.DB, env.Listener),
		roleBindings:   authdb.RoleBindingCollection(env.DB, env.Listener),
		members:        authdb.MembersCollection(env.DB, env.Listener),
		groups:         authdb.GroupsCollection(env.DB, env.Listener),
		oidcStates:     oidcStates,
		public:         public,
		watchesEnabled: watchesEnabled,
	}

	if public {
		// start OIDC service (won't respond to anything until config is set)
		go waitForError("OIDC HTTP Server", requireNoncriticalServers, s.serveOIDC)
	}

	if watchesEnabled {
		s.configCache = keycache.NewCache(s.authConfig.ReadOnly(), configKey, &DefaultOIDCConfig)
		s.clusterRoleBindingCache = keycache.NewCache(s.roleBindings.ReadOnly(), auth.ClusterRoleBindingKey, &auth.RoleBinding{})

		// Watch for new auth config options
		go s.configCache.Watch(env.BackgroundContext)

		// Watch for changes to the cluster role binding
		go s.clusterRoleBindingCache.Watch(env.BackgroundContext)
	}

	if err := s.deleteExpiredTokensRoutine(env.BackgroundContext); err != nil {
		return nil, err
	}
	return s, nil
}

// ActivationScope is an additional service to activate auth for.
type ActivationScope int

const (
	ActivationScopePFS ActivationScope = iota // Activate auth for PFS.
	ActivationScopePPS                        // Activate auth for PPS.
)

// String implements fmt.Stringer.
func (s ActivationScope) String() string {
	switch s { //exhaustive:enforce
	case ActivationScopePFS:
		return "PFS"
	case ActivationScopePPS:
		return "PPS"
	}
	panic(fmt.Sprintf("invalid auth scope %d", int(s)))
}

// ActivateAuthEverywhere activates auth, and PPS and PFS auth, in a single transaction.  It is only
// exposed publicly for testing; do not call it from outside this package.
func (a *apiServer) ActivateAuthEverywhere(ctx context.Context, scopes []ActivationScope, rootToken string) error {
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(ctx context.Context, txCtx *txncontext.TransactionContext) error {
		log.Debug(ctx, "attempting to activate auth")
		if _, err := a.activateInTransaction(ctx, txCtx, &auth.ActivateRequest{
			RootToken: rootToken,
		}); err != nil {
			if !errors.Is(err, auth.ErrAlreadyActivated) {
				return errors.Wrap(err, "activate auth")
			}
			log.Debug(ctx, "auth already active; rotating root token")
			_, err := a.rotateRootTokenInTransaction(ctx, txCtx,
				&auth.RotateRootTokenRequest{
					RootToken: rootToken,
				})
			return errors.Wrap(err, "rotate root token")
		}
		for _, s := range scopes {
			switch s { //exhaustive:enforce
			case ActivationScopePFS:
				log.Debug(ctx, "attempting to activate PFS auth")
				if _, err := a.env.GetPfsServer().ActivateAuthInTransaction(ctx, txCtx, &pfs.ActivateAuthRequest{}); err != nil {
					return errors.Wrap(err, "activate auth for pfs")
				}
			case ActivationScopePPS:
				log.Debug(ctx, "attempting to activate PPS auth")
				if _, err := a.env.GetPpsServer().ActivateAuthInTransaction(ctx, txCtx, &pps.ActivateAuthRequest{}); err != nil {
					return errors.Wrap(err, "activate auth for pps")
				}
			}
		}
		return nil
	}); err != nil {
		return errors.Wrapf(err, "activate auth everywhere (%v)", scopes)
	}
	log.Info(ctx, "activated auth ok", zap.Stringers("scopes", scopes))
	return nil
}

func (a *apiServer) EnvBootstrap(ctx context.Context) error {
	if !a.env.Config.ActivateAuth {
		return nil
	}
	log.Info(ctx, "Started to configure auth server via environment")
	ctx = internalauth.AsInternalUser(ctx, authdb.InternalUser)
	// handle auth activation
	if rootToken := a.env.Config.AuthRootToken; rootToken != "" {
		var scopes []ActivationScope
		if a.env.Config.PachdSpecificConfiguration != nil {
			// If this is pachd and not the enterprise server, activate auth for PFS and
			// PPS.
			scopes = []ActivationScope{ActivationScopePFS, ActivationScopePPS}
		}
		if err := a.ActivateAuthEverywhere(ctx, scopes, rootToken); err != nil {
			return errors.Wrapf(err, "activate auth via environment")
		}
	}

	if err := func() error {
		// handle oidc clients & this cluster's auth config
		if a.env.Config.AuthConfig != "" && a.env.Config.IdentityClients != "" {
			log.Info(ctx, "attempting to add or update oidc clients")
			var config auth.OIDCConfig
			var clients []*identity.OIDCClient
			if err := yaml.Unmarshal([]byte(a.env.Config.AuthConfig), &config); err != nil {
				return errors.Wrapf(err, "unmarshal auth config: %q", a.env.Config.AuthConfig)
			}
			config.ClientSecret = a.env.Config.AuthClientSecret
			if err := yaml.Unmarshal([]byte(a.env.Config.IdentityClients), &clients); err != nil {
				return errors.Wrapf(err, "unmarshal identity clients: %q", a.env.Config.IdentityClients)
			}
			if a.env.Config.IdentityAdditionalClients != "" {
				var extras []*identity.OIDCClient
				if err := yaml.Unmarshal([]byte(a.env.Config.IdentityAdditionalClients), &extras); err != nil {
					return errors.Wrapf(err, "unmarshal extra identity clients: %q", a.env.Config.IdentityAdditionalClients)
				}
				clients = append(clients, extras...)
			}
			for _, c := range clients {
				log.Info(ctx, "adding oidc client", zap.String("id", c.Id), zap.Bool("via_enterprise_server", a.env.Config.EnterpriseMember))
				if c.Id == config.ClientId { // c represents pachd
					c.Secret = config.ClientSecret
					if a.env.Config.TrustedPeers != "" {
						var tps []string
						if err := yaml.Unmarshal([]byte(a.env.Config.TrustedPeers), &tps); err != nil {
							return errors.Wrapf(err, "unmarshal trusted peers: %q", a.env.Config.TrustedPeers)
						}
						c.TrustedPeers = append(c.TrustedPeers, tps...)
						log.Info(ctx, "adding additional pachd trusted peers configured via environment", zap.Strings("trusted_peers", c.TrustedPeers), zap.String("id", c.Id))

					}
				}
				if c.Id == a.env.Config.ConsoleOAuthID {
					c.Secret = a.env.Config.ConsoleOAuthSecret
				}
				if c.Id == a.env.Config.DeterminedOAuthID {
					c.Secret = a.env.Config.DeterminedOAuthSecret
				}
				if !a.env.Config.EnterpriseMember {
					if _, err := a.env.GetIdentityServer().CreateOIDCClient(ctx, &identity.CreateOIDCClientRequest{Client: c}); err != nil {
						if !identity.IsErrAlreadyExists(err) {
							return errors.Wrapf(err, "create oidc client %q", c.Name)
						}
						// recreate the client because updating the client secret is not supported by the dex API
						if _, err := a.env.GetIdentityServer().DeleteOIDCClient(ctx, &identity.DeleteOIDCClientRequest{Id: c.Id}); err != nil {
							return errors.Wrapf(err, "delete oidc client %q", c.Name)
						}
						if _, err := a.env.GetIdentityServer().CreateOIDCClient(ctx, &identity.CreateOIDCClientRequest{Client: c}); err != nil {
							return errors.Wrapf(err, "update oidc client %q", c.Name)
						}
					}
				} else {
					ec, err := client.NewFromURI(ctx, a.env.Config.EnterpriseServerAddress)
					if err != nil {
						return errors.Wrapf(err, "connect to enterprise server")
					}
					ec.SetAuthToken(a.env.Config.EnterpriseServerToken)
					if _, err = ec.IdentityAPIClient.CreateOIDCClient(ec.Ctx(), &identity.CreateOIDCClientRequest{Client: c}); err != nil {
						if !identity.IsErrAlreadyExists(err) {
							return errors.Wrapf(err, "create oidc client %q", c.Name)
						}
						// recreate the client because updating the client secret is not supported by the dex API
						if _, err := ec.IdentityAPIClient.DeleteOIDCClient(ec.Ctx(), &identity.DeleteOIDCClientRequest{Id: c.Id}); err != nil {
							return errors.Wrapf(err, "delete oidc client %q", c.Name)
						}
						if _, err := ec.IdentityAPIClient.CreateOIDCClient(ec.Ctx(), &identity.CreateOIDCClientRequest{Client: c}); err != nil {
							return errors.Wrapf(err, "update oidc client %q", c.Name)
						}
					}
				}
			}
			log.Info(ctx, "setting auth configuration", log.Proto("config", &config))
			if _, err := a.SetConfiguration(ctx, &auth.SetConfigurationRequest{Configuration: &config}); err != nil {
				return err
			}
		}
		// cluster role bindings
		if a.env.Config.AuthClusterRoleBindings != "" {
			log.Info(ctx, "setting up cluster role bindings")
			var roleBinding map[string][]string
			if err := yaml.Unmarshal([]byte(a.env.Config.AuthClusterRoleBindings), &roleBinding); err != nil {
				return errors.Wrapf(err, "unmarshal auth cluster role bindings: %q", a.env.Config.AuthClusterRoleBindings)
			}
			existing, err := a.GetRoleBinding(ctx, &auth.GetRoleBindingRequest{
				Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER},
			})
			if err != nil {
				return errors.Wrapf(err, "get cluster role bindings")
			}
			for p := range existing.Binding.Entries {
				// `pach:` user role bindings cannot be modified
				if strings.HasPrefix(p, auth.PachPrefix) || strings.HasPrefix(p, auth.InternalPrefix) {
					log.Info(ctx, "skipping role binding because pach: role bindings cannot be modified", zap.String("principal", p))
					continue
				}
				if _, ok := roleBinding[p]; !ok {
					rsc := &auth.Resource{Type: auth.ResourceType_CLUSTER}
					log.Debug(ctx, "unsetting existing role binding", log.Proto("resource", rsc), zap.String("principal", p))
					if _, err := a.ModifyRoleBinding(ctx, &auth.ModifyRoleBindingRequest{
						Resource:  rsc,
						Principal: p,
					}); err != nil {
						return errors.Wrapf(err, "unset principal cluster role bindings for principal %q", p)
					}
				}
			}
			for p, r := range roleBinding {
				rsc := &auth.Resource{Type: auth.ResourceType_CLUSTER}
				log.Info(ctx, "modifying existing role binding", log.Proto("resource", rsc), zap.String("principal", p), zap.Strings("roles", r))
				if _, err := a.ModifyRoleBinding(ctx, &auth.ModifyRoleBindingRequest{
					Resource:  rsc,
					Principal: p,
					Roles:     r,
				}); err != nil {
					return errors.Wrapf(err, "modify cluster role bindings")
				}
			}
		}
		return nil
	}(); err != nil {
		return errors.Wrapf(errors.EnsureStack(err), "configure the auth server via environment")
	}
	log.Info(ctx, "Successfully configured auth server via environment")
	return nil
}

func waitForError(name string, required bool, cb func() error) {
	if err := cb(); !errors.Is(err, http.ErrServerClosed) {
		if required {
			log.Error(pctx.TODO(), "error setting up and/or running server (use --require-critical-servers-only deploy flag to ignore errors from noncritical servers)", zap.String("server", name), zap.Error(err))
			os.Exit(20)
		}
		log.Error(pctx.TODO(), "error setting up and/or running server", zap.String("server", name), zap.Error(err))
	}
}

// getClusterRoleBinding attempts to get the current cluster role bindings,
// and returns an error if auth is not activated. This can require hitting
// postgres if watches are not enabled (in the worker sidecar).
func (a *apiServer) getClusterRoleBinding(ctx context.Context) (*auth.RoleBinding, error) {
	if a.watchesEnabled {
		bindings, ok := a.clusterRoleBindingCache.Load().(*auth.RoleBinding)
		if !ok {
			return nil, errors.New("cached cluster binding had unexpected type")
		}

		if len(bindings.Entries) == 0 {
			return nil, auth.ErrNotActivated
		}
		return bindings, nil
	}

	var binding auth.RoleBinding
	if err := a.roleBindings.ReadOnly().Get(ctx, auth.ClusterRoleBindingKey, &binding); err != nil {
		if col.IsErrNotFound(err) {
			return nil, auth.ErrNotActivated
		}
		return nil, errors.EnsureStack(err)
	}
	return &binding, nil
}

func (a *apiServer) isActive(ctx context.Context) error {
	_, err := a.getClusterRoleBinding(ctx)
	return err
}

func (a *apiServer) isActiveInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
	_, err := a.getClusterRoleBindingInTransaction(ctx, txnCtx)
	return err
}

// getClusterRoleBinding attempts to get the current cluster role bindings,
// and returns an error if auth is not activated. This can require hitting
// postgres if watches are not enabled (in the worker sidecar).
func (a *apiServer) getClusterRoleBindingInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext) (*auth.RoleBinding, error) {
	if a.watchesEnabled && !txnCtx.AuthBeingActivated.Load() {
		bindings, ok := a.clusterRoleBindingCache.Load().(*auth.RoleBinding)
		if !ok {
			return nil, errors.New("cached cluster binding had unexpected type")
		}

		if bindings.Entries == nil {
			return nil, auth.ErrNotActivated
		}
		return bindings, nil
	}

	var binding auth.RoleBinding
	if err := a.roleBindings.ReadWrite(txnCtx.SqlTx).Get(ctx, auth.ClusterRoleBindingKey, &binding); err != nil {
		if col.IsErrNotFound(err) {
			return nil, auth.ErrNotActivated
		}
		return nil, errors.EnsureStack(err)
	}
	return &binding, nil
}

func (a *apiServer) getEnterpriseTokenState(ctx context.Context) (enterpriseclient.State, error) {
	if a.env.GetEnterpriseServer() == nil {
		return 0, errors.New("Enterprise Server not yet initialized")
	}
	resp, err := a.env.GetEnterpriseServer().GetState(ctx, &enterpriseclient.GetStateRequest{})
	if err != nil {
		return 0, errors.Wrapf(grpcutil.ScrubGRPC(err), "could not get Enterprise status")
	}
	return resp.State, nil
}

// expiredEnterpriseCheck enforces that if the cluster's enterprise token is
// expired, users cannot log in. The root token can be used to access the cluster.
func (a *apiServer) expiredEnterpriseCheck(ctx context.Context, username string) error {
	state, err := a.getEnterpriseTokenState(ctx)
	if err != nil {
		return errors.Wrapf(err, "error confirming Pachyderm Enterprise token")
	}

	if state == enterpriseclient.State_ACTIVE {
		return nil
	}

	if err = a.userHasExpiredClusterAccessCheck(ctx, username); err != nil {
		return errors.Wrapf(err, "Pachyderm Enterprise is not active in this "+
			"cluster (until Pachyderm Enterprise is re-activated or Pachyderm "+
			"auth is deactivated, only cluster admins can perform any operations)")
	}
	return nil
}

func (a *apiServer) userHasExpiredClusterAccessCheck(ctx context.Context, username string) error {
	// Root User, PPS Master, and any Pipeline keep cluster access
	if username == auth.RootUser || strings.HasPrefix(username, auth.PipelinePrefix) || strings.HasPrefix(username, auth.InternalPrefix) {
		return nil
	}

	// Any User with the Cluster Admin Role keeps cluster access
	isAdmin, err := a.hasClusterRole(ctx, username, auth.ClusterAdminRole)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.Errorf("user: %v, is not priviledged to operate while Enterprise License is disabled", username)
	}
	return nil
}

func (a *apiServer) hasClusterRole(ctx context.Context, username string, role string) (bool, error) {
	bindings, err := a.getClusterRoleBinding(ctx)
	if err != nil {
		return false, err
	}
	if roles, ok := bindings.Entries[username]; ok {
		for r := range roles.Roles {
			if r == role {
				return true, nil
			}
		}
	}
	return false, nil
}

// Activate implements the protobuf auth.Activate RPC
func (a *apiServer) Activate(ctx context.Context, req *auth.ActivateRequest) (*auth.ActivateResponse, error) {
	var resp *auth.ActivateResponse
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(ctx context.Context, txCtx *txncontext.TransactionContext) error {
		var err error
		resp, err = a.activateInTransaction(ctx, txCtx, req)
		return err
	}); err != nil {
		return nil, err
	}
	// wait until the clusterRoleBinding watcher has updated the local cache
	// (changing the activation state), so that Activate() is less likely to
	// race with subsequent calls that expect auth to be activated.
	if err := backoff.RetryUntilCancel(ctx, func() error {
		if err := a.isActive(ctx); err != nil {
			return errors.Errorf("auth activation hasn't fully propagated")
		}
		return nil
	}, backoff.RetryEvery(100*time.Millisecond), nil); err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *apiServer) activateInTransaction(ctx context.Context, txCtx *txncontext.TransactionContext, req *auth.ActivateRequest) (*auth.ActivateResponse, error) {
	// TODO(acohen4) 2.3: disable RPC if a.env.Config.AuthRootToken != ""
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
	// themselves as an admin.
	if err := a.isActiveInTransaction(ctx, txCtx); err == nil {
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
	roleBindings := a.roleBindings.ReadWrite(txCtx.SqlTx)
	if err := roleBindings.Put(ctx, auth.ClusterRoleBindingKey, &auth.RoleBinding{
		Entries: map[string]*auth.Roles{
			auth.RootUser:               {Roles: map[string]bool{auth.ClusterAdminRole: true}},
			authdb.InternalUser:         {Roles: map[string]bool{auth.ClusterAdminRole: true}},
			auth.AllClusterUsersSubject: {Roles: map[string]bool{auth.ProjectCreatorRole: true}},
		},
	}); err != nil {
		return nil, errors.Wrap(err, "add cluster role binding")
	}
	if err := a.insertAuthTokenNoTTLInTransaction(ctx, txCtx, auth.HashToken(pachToken), auth.RootUser); err != nil {
		return nil, errors.Wrap(err, "insert root token")
	}
	txCtx.AuthBeingActivated.Store(true)
	return &auth.ActivateResponse{PachToken: pachToken}, nil
}

// RotateRootToken implements the protobuf auth.RotateRootToken RPC
func (a *apiServer) RotateRootToken(ctx context.Context, req *auth.RotateRootTokenRequest) (*auth.RotateRootTokenResponse, error) {
	if a.env.Config.AuthRootToken != "" {
		return nil, errors.New("RotateRootToken() is disabled when the root token is configured in the environment")
	}
	var resp *auth.RotateRootTokenResponse
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(ctx context.Context, txCtx *txncontext.TransactionContext) error {
		var err error
		resp, err = a.rotateRootTokenInTransaction(ctx, txCtx, req)
		return err
	}); err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *apiServer) rotateRootTokenInTransaction(ctx context.Context, txCtx *txncontext.TransactionContext, req *auth.RotateRootTokenRequest) (resp *auth.RotateRootTokenResponse, retErr error) {
	var rootToken string
	// First revoke root's existing auth token
	if _, err := a.deleteAuthTokensForSubjectInTransaction(ctx, txCtx.SqlTx, auth.RootUser); err != nil {
		return nil, err
	}
	// If the new token is in the request, use it.
	// Otherwise generate a new random token.
	rootToken = req.RootToken
	if rootToken == "" {
		rootToken = uuid.NewWithoutDashes()
	}
	if err := a.insertAuthTokenNoTTLInTransaction(ctx, txCtx, auth.HashToken(rootToken), auth.RootUser); err != nil {
		return nil, err
	}

	return &auth.RotateRootTokenResponse{RootToken: rootToken}, nil
}

// Deactivate implements the protobuf auth.Deactivate RPC
func (a *apiServer) Deactivate(ctx context.Context, req *auth.DeactivateRequest) (resp *auth.DeactivateResponse, retErr error) {
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		if err := a.roleBindings.ReadWrite(sqlTx).DeleteAll(ctx); err != nil {
			return errors.EnsureStack(err)
		}
		if err := a.deleteAllAuthTokens(ctx, sqlTx); err != nil {
			return errors.EnsureStack(err)
		}
		if err := a.members.ReadWrite(sqlTx).DeleteAll(ctx); err != nil {
			return errors.EnsureStack(err)
		}
		if err := a.groups.ReadWrite(sqlTx).DeleteAll(ctx); err != nil {
			return errors.EnsureStack(err)
		}
		if err := a.authConfig.ReadWrite(sqlTx).DeleteAll(ctx); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// wait until the clusterRoleBinding watcher sees the deleted role binding,
	// so that Deactivate() is less likely to race with subsequent calls that
	// expect auth to be deactivated.
	if err := backoff.Retry(func() error {
		if err := a.isActive(ctx); err == nil {
			return errors.Errorf("auth still activated")
		}
		return nil
	}, backoff.RetryEvery(100*time.Millisecond)); err != nil {
		return nil, err
	}
	return &auth.DeactivateResponse{}, nil
}

// Authenticate implements the protobuf auth.Authenticate RPC
func (a *apiServer) Authenticate(ctx context.Context, req *auth.AuthenticateRequest) (resp *auth.AuthenticateResponse, retErr error) {
	if err := a.isActive(ctx); err != nil {
		return nil, err
	}

	// verify whatever credential the user has presented, and write a new
	// Pachyderm token for the user that their credential belongs to
	var pachToken string
	switch {
	case req.OidcState != "":
		// Determine caller's Pachyderm/OIDC user info (email)
		email, err := a.OIDCStateToEmail(ctx, req.OidcState)
		if err != nil {
			return nil, err
		}

		username := auth.UserPrefix + email

		if err := a.expiredEnterpriseCheck(ctx, username); err != nil {
			return nil, err
		}

		// Generate a new Pachyderm token and write it
		t, err := a.generateAndInsertAuthToken(ctx, username, int64(60*a.env.Config.SessionDurationMinutes))
		if err != nil {
			return nil, errors.Wrapf(err, "error storing auth token for user \"%s\"", username)
		}
		pachToken = t

	case req.IdToken != "":
		// Determine caller's Pachyderm/OIDC user info (email)
		token, claims, err := a.validateIDToken(ctx, req.IdToken)
		if err != nil {
			return nil, err
		}

		username := auth.UserPrefix + claims.Email

		if err := a.expiredEnterpriseCheck(ctx, username); err != nil {
			return nil, err
		}

		// Sync the user's group membership from the groups claim
		if err := a.syncGroupMembership(ctx, claims); err != nil {
			return nil, err
		}

		// Compute the remaining time before the ID token expires,
		// and limit the pach token to the same expiration time.
		// If the token would be longer-lived than the default pach token,
		// TTL clamp the expiration to the default TTL.
		expirationSecs := int64(time.Until(token.Expiry).Seconds())
		if expirationSecs > int64(60*a.env.Config.SessionDurationMinutes) {
			expirationSecs = int64(60 * a.env.Config.SessionDurationMinutes)
		}

		t, err := a.generateAndInsertAuthToken(ctx, username, expirationSecs)
		if err != nil {
			return nil, errors.Wrapf(err, "error storing auth token for user \"%s\"", username)
		}
		pachToken = t

	default:
		return nil, errors.Errorf("unrecognized authentication mechanism (old pachd?)")
	}

	log.Info(ctx, "Authentication checks successful, now returning pachToken")

	// Return new pachyderm token to caller
	return &auth.AuthenticateResponse{
		PachToken: pachToken,
	}, nil
}

func (a *apiServer) evaluateRoleBindingInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, principal string, resource *auth.Resource, permissions map[auth.Permission]bool) (*authorizeRequest, error) {
	request := newAuthorizeRequest(principal, permissions, a.getGroupsInTransaction)

	// Special-case making spec repos world-readable, because the alternative breaks reading pipelines.
	// TOOD: 2.0 - should we make this a user-configurable cluster binding instead of hard-coding it?
	if resource.Type == auth.ResourceType_SPEC_REPO {
		if err := request.evaluateRoleBinding(ctx, txnCtx, &auth.RoleBinding{
			Entries: map[string]*auth.Roles{
				auth.AllClusterUsersSubject: {
					Roles: map[string]bool{
						auth.RepoReaderRole: true,
					},
				},
			},
		}); err != nil {
			return nil, err
		}

		// If the user only requested reader access, we can return early.
		// Otherwise we just treat this like a request about the associated user repo.
		if request.isSatisfied() {
			return request, nil
		}
		resource.Type = auth.ResourceType_REPO
	}

	// Check the permissions at the cluster level
	binding, err := a.getClusterRoleBindingInTransaction(ctx, txnCtx)
	if err != nil {
		return nil, err
	}

	if err := request.evaluateRoleBinding(ctx, txnCtx, binding); err != nil {
		return nil, err
	}

	// If all the permissions are satisfied by the cached cluster binding don't
	// retrieve the resource bindings. If the resource in question is the whole
	// cluster we should also exit early
	if request.isSatisfied() || resource.Type == auth.ResourceType_CLUSTER {
		return request, nil
	}

	// if resource is a repo, then we should check project level permissions as well
	var roleBinding auth.RoleBinding
	if resource.Type == auth.ResourceType_REPO {
		repo, err := authRepoResourceToRepo(resource)
		if err != nil {
			return nil, err
		}
		projectResource := &auth.Resource{Type: auth.ResourceType_PROJECT, Name: repo.Project.Name}
		projectKey := authdb.ResourceKey(projectResource)
		if err := a.roleBindings.ReadWrite(txnCtx.SqlTx).Get(ctx, projectKey, &roleBinding); err != nil {
			if col.IsErrNotFound(err) {
				return nil, &auth.ErrNoRoleBinding{Resource: projectResource}
			}
			return nil, errors.Wrapf(err, "error getting role bindings for %s", projectKey)
		}
		if err := request.evaluateRoleBinding(ctx, txnCtx, &roleBinding); err != nil {
			return nil, err
		}
		if request.isSatisfied() {
			return request, nil
		}
	}

	// Get the role bindings for the resource to check
	if err := a.roleBindings.ReadWrite(txnCtx.SqlTx).Get(ctx, authdb.ResourceKey(resource), &roleBinding); err != nil {
		if col.IsErrNotFound(err) {
			return nil, &auth.ErrNoRoleBinding{
				Resource: resource,
			}
		}
		return nil, errors.Wrapf(err, "error getting role bindings for %s \"%s\"", resource.Type, resource.Name)
	}
	if err := request.evaluateRoleBinding(ctx, txnCtx, &roleBinding); err != nil {
		return nil, err
	}
	return request, nil
}

// AuthorizeInTransaction is identical to Authorize except that it can run in a `pachsql.Tx`.
func (a *apiServer) AuthorizeInTransaction(
	ctx context.Context,
	txnCtx *txncontext.TransactionContext,
	req *auth.AuthorizeRequest,
) (resp *auth.AuthorizeResponse, retErr error) {
	me, err := txnCtx.WhoAmI()
	if err != nil {
		return nil, err
	}

	permissions := make(map[auth.Permission]bool)
	for _, p := range req.Permissions {
		permissions[p] = true
	}

	request, err := a.evaluateRoleBindingInTransaction(ctx, txnCtx, me.Username, req.Resource, permissions)
	if err != nil {
		return nil, err
	}

	return &auth.AuthorizeResponse{
		Principal:  me.Username,
		Authorized: request.isSatisfied(),
		Missing:    request.missing(),
		Satisfied:  request.satisfiedPermissions,
	}, nil
}

// Authorize implements the protobuf auth.Authorize RPC
func (a *apiServer) Authorize(
	ctx context.Context,
	req *auth.AuthorizeRequest,
) (resp *auth.AuthorizeResponse, retErr error) {
	var response *auth.AuthorizeResponse
	if err := a.env.TxnEnv.WithReadContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		var err error
		response, err = a.AuthorizeInTransaction(ctx, txnCtx, req)
		return err
	}); err != nil {
		return nil, err
	}
	return response, nil
}

func (a *apiServer) GetPermissionsForPrincipal(ctx context.Context, req *auth.GetPermissionsForPrincipalRequest) (*auth.GetPermissionsResponse, error) {
	permissions := make(map[auth.Permission]bool)
	for p := range auth.Permission_name {
		permissions[auth.Permission(p)] = true
	}
	var request *authorizeRequest
	if err := a.env.TxnEnv.WithReadContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		var err error
		request, err = a.evaluateRoleBindingInTransaction(ctx, txnCtx, req.Principal, req.Resource, permissions)
		return err
	}); err != nil {
		return nil, errors.Wrap(err, "cannot evaluate role binding")
	}
	return &auth.GetPermissionsResponse{
		Roles:       request.rolesForResourceType(req.Resource.Type),
		Permissions: request.satisfiedForResourceType(req.Resource.Type),
	}, nil
}

// GetPermissions implements the protobuf auth.GetPermissions RPC
func (a *apiServer) GetPermissions(ctx context.Context, req *auth.GetPermissionsRequest) (resp *auth.GetPermissionsResponse, retErr error) {
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get authenticated user")
	}

	resp, err = a.GetPermissionsForPrincipal(ctx, &auth.GetPermissionsForPrincipalRequest{Principal: callerInfo.Subject, Resource: req.Resource})
	if err != nil {
		return nil, errors.Wrap(err, "cannot get permissions for principal")
	}
	return resp, nil
}

func (a *apiServer) GetPermissionsInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, req *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error) {
	callerInfo, err := txnCtx.WhoAmI()
	if err != nil {
		return nil, err
	}

	return a.getPermissionsForPrincipalInTransaction(ctx, txnCtx, &auth.GetPermissionsForPrincipalRequest{Principal: callerInfo.Username, Resource: req.Resource})
}

func (a *apiServer) getPermissionsForPrincipalInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, req *auth.GetPermissionsForPrincipalRequest) (*auth.GetPermissionsResponse, error) {
	permissions := make(map[auth.Permission]bool)
	for p := range auth.Permission_name {
		permissions[auth.Permission(p)] = true
	}

	request, err := a.evaluateRoleBindingInTransaction(ctx, txnCtx, req.Principal, req.Resource, permissions)
	if err != nil {
		return nil, err
	}

	return &auth.GetPermissionsResponse{
		Roles:       request.rolesForResourceType(req.Resource.Type),
		Permissions: request.satisfiedForResourceType(req.Resource.Type),
	}, nil
}

// WhoAmI implements the protobuf auth.WhoAmI RPC
func (a *apiServer) WhoAmI(ctx context.Context, req *auth.WhoAmIRequest) (resp *auth.WhoAmIResponse, retErr error) {
	if err := a.isActive(ctx); err != nil {
		return nil, err
	}

	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	return &auth.WhoAmIResponse{
		Username:   callerInfo.Subject,
		Expiration: callerInfo.Expiration,
	}, nil
}

// DeleteRoleBindingInTransaction is used to remove role bindings for resources when they're deleted in other services.
// It doesn't do any auth checks itself - the calling method should ensure the user is allowed to delete this resource.
// This is not an RPC, this is only called in-process.
func (a *apiServer) DeleteRoleBindingInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, resource *auth.Resource) error {
	if err := a.isActiveInTransaction(ctx, txnCtx); err != nil {
		return err
	}

	if resource.Type == auth.ResourceType_CLUSTER {
		return errors.Errorf("cannot delete cluster role binding")
	}

	key := authdb.ResourceKey(resource)
	roleBindings := a.roleBindings.ReadWrite(txnCtx.SqlTx)
	if err := roleBindings.Delete(ctx, key); err != nil {
		return errors.EnsureStack(err)
	}

	return nil
}

// rolesFromRoleSlice converts a slice of strings into *auth.Roles,
// validating that each role name is valid.
func rolesFromRoleSlice(rs []string) (*auth.Roles, error) {
	if len(rs) == 0 {
		return nil, nil
	}

	for _, r := range rs {
		if _, err := getRole(r); err != nil {
			return nil, err
		}
	}

	roles := &auth.Roles{Roles: make(map[string]bool)}
	for _, r := range rs {
		roles.Roles[r] = true
	}
	return roles, nil
}

// CreateRoleBindingInTransaction is an internal-only API to create a role binding for a new resource.
// It doesn't do any authorization checks itself - the calling method should ensure the user is allowed
// to create the resource. This is not an RPC.
func (a *apiServer) CreateRoleBindingInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, principal string, roleSlice []string, resource *auth.Resource) error {
	bindings := &auth.RoleBinding{
		Entries: make(map[string]*auth.Roles),
	}

	if len(roleSlice) != 0 {
		// Check that the subject is in canonical form (`<type>:<principal>`).
		if err := a.checkCanonicalSubject(principal); err != nil {
			return err
		}

		roles, err := rolesFromRoleSlice(roleSlice)
		if err != nil {
			return err
		}

		bindings.Entries[principal] = roles
	}

	// Call Create, this will raise an error if the role binding already exists.
	key := authdb.ResourceKey(resource)
	roleBindings := a.roleBindings.ReadWrite(txnCtx.SqlTx)
	if err := roleBindings.Create(ctx, key, bindings); err != nil {
		return errors.EnsureStack(err)
	}

	return nil
}

// AddPipelineReaderToRepoInTransaction gives a pipeline access to read data from the specified source repo.
// This is distinct from ModifyRoleBinding because AddPipelineReader is a less expansive permission
// that is included in the repoReader role, versus being able to modify all role bindings which is
// part of repoOwner. This method is for internal use and is not exposed as an RPC.
func (a *apiServer) AddPipelineReaderToRepoInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, sourceRepo *pfs.Repo, pipeline *pps.Pipeline) error {
	r := &pfs.Repo{
		Project: sourceRepo.Project,
		Name:    sourceRepo.Name,
		Type:    pfs.UserRepoType,
	}
	if err := a.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, r, auth.Permission_REPO_ADD_PIPELINE_READER); err != nil {
		return err
	}

	return a.setUserRoleBindingInTransaction(ctx, txnCtx, sourceRepo.AuthResource(), auth.PipelinePrefix+pipeline.String(), []string{auth.RepoReaderRole})
}

// AddPipelineWriterToSourceRepoInTransaction gives a pipeline access to write data to the specified source repo.
// The only time a pipeline needs write permission for a source repo is in the case of Cron inputs.
// This is distinct from ModifyRoleBinding because AddPipelineWriter is a less expansive permission
// that is included in the repoWriter role, versus being able to modify all role bindings which is
// part of repoOwner. This method is for internal use and is not exposed as an RPC.
func (a *apiServer) AddPipelineWriterToSourceRepoInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, sourceRepo *pfs.Repo, pipeline *pps.Pipeline) error {
	// Check that the user is allowed to add a pipeline to write to the output repo.
	r := &pfs.Repo{
		Project: sourceRepo.Project,
		Name:    sourceRepo.Name,
		Type:    pfs.UserRepoType,
	}
	if err := a.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, r, auth.Permission_REPO_ADD_PIPELINE_WRITER); err != nil {
		return err
	}
	return a.setUserRoleBindingInTransaction(ctx, txnCtx, sourceRepo.AuthResource(), auth.PipelinePrefix+pipeline.String(), []string{auth.RepoWriterRole})
}

// AddPipelineWriterToRepoInTransaction gives a pipeline access to write to it's own output repo.
// This is distinct from ModifyRoleBinding because AddPipelineWriter is a less expansive permission
// that is included in the repoWriter role, versus being able to modify all role bindings which is
// part of repoOwner. This method is for internal use and is not exposed as an RPC.
func (a *apiServer) AddPipelineWriterToRepoInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, pipeline *pps.Pipeline) error {
	// Check that the user is allowed to add a pipeline to write to the output repo.
	r := &pfs.Repo{
		Project: pipeline.Project,
		Name:    pipeline.Name,
		Type:    pfs.UserRepoType,
	}
	if err := a.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, r, auth.Permission_REPO_ADD_PIPELINE_WRITER); err != nil {
		return err
	}

	return a.setUserRoleBindingInTransaction(ctx, txnCtx, r.AuthResource(), auth.PipelinePrefix+pipeline.String(), []string{auth.RepoWriterRole})
}

// RemovePipelineReaderFromRepo revokes a pipeline's access to read data from the specified source repo.
// This is distinct from ModifyRoleBinding because RemovePipelineReader is a less expansive permission
// that is included in the repoWriter role, versus being able to modify all role bindings which is
// part of repoOwner. This method is for internal use and is not exposed as an RPC.
func (a *apiServer) RemovePipelineReaderFromRepoInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, sourceRepo *pfs.Repo, pipeline *pps.Pipeline) error {
	// Check that the user is allowed to remove input repos from the pipeline repo - this check is on the pipeline itself
	// and not sourceRepo because otherwise users could break piplines they don't have access to by revoking them from the
	// input repo.
	r := &pfs.Repo{
		Project: pipeline.Project,
		Name:    pipeline.Name,
		Type:    pfs.UserRepoType,
	}
	if err := a.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, r, auth.Permission_REPO_REMOVE_PIPELINE_READER); err != nil && !auth.IsErrNoRoleBinding(err) {
		return err
	}

	return a.setUserRoleBindingInTransaction(ctx, txnCtx, sourceRepo.AuthResource(), auth.PipelinePrefix+pipeline.String(), []string{})
}

// ModifyRoleBindingInTransaction is identical to ModifyRoleBinding except that it can run inside
// an existing postgres transaction.  This is not an RPC.
func (a *apiServer) ModifyRoleBindingInTransaction(
	ctx context.Context,
	txnCtx *txncontext.TransactionContext,
	req *auth.ModifyRoleBindingRequest,
) (*auth.ModifyRoleBindingResponse, error) {
	if err := a.isActiveInTransaction(ctx, txnCtx); err != nil {
		return nil, err
	}

	if err := a.checkCanonicalSubject(req.Principal); err != nil {
		return nil, err
	}

	if strings.HasPrefix(req.Principal, auth.PachPrefix) && req.Resource.Type == auth.ResourceType_CLUSTER {
		return nil, errors.Errorf("cannot modify cluster role bindings for pach: users")
	}

	// ModifyRoleBinding can be called for any type of resource,
	// and the permission required depends on the type of resource.
	var permission auth.Permission
	switch req.Resource.Type {
	case auth.ResourceType_CLUSTER:
		permission = auth.Permission_CLUSTER_MODIFY_BINDINGS
	case auth.ResourceType_PROJECT:
		permission = auth.Permission_PROJECT_MODIFY_BINDINGS
	case auth.ResourceType_REPO:
		permission = auth.Permission_REPO_MODIFY_BINDINGS
	default:
		return nil, errors.Errorf("unknown resource type %v", req.Resource.Type)
	}
	if err := a.checkResourceIsAuthorizedInTransaction(ctx, txnCtx, req.Resource, permission); err != nil {
		return nil, err
	}

	if err := a.setUserRoleBindingInTransaction(ctx, txnCtx, req.Resource, req.Principal, req.Roles); err != nil {
		return nil, err
	}

	return &auth.ModifyRoleBindingResponse{}, nil
}

func (a *apiServer) setUserRoleBindingInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, resource *auth.Resource, principal string, roleSlice []string) error {
	roles, err := rolesFromRoleSlice(roleSlice)
	if err != nil {
		return err
	}

	key := authdb.ResourceKey(resource)
	roleBindings := a.roleBindings.ReadWrite(txnCtx.SqlTx)
	var bindings auth.RoleBinding
	if err := roleBindings.Get(ctx, key, &bindings); err != nil {
		if col.IsErrNotFound(err) {
			return &auth.ErrNoRoleBinding{
				Resource: resource,
			}
		}
		return errors.EnsureStack(err)
	}

	if bindings.Entries == nil {
		bindings.Entries = make(map[string]*auth.Roles)
	}

	if len(roleSlice) == 0 {
		delete(bindings.Entries, principal)
	} else {
		bindings.Entries[principal] = roles
	}
	return errors.EnsureStack(roleBindings.Put(ctx, key, &bindings))
}

// ModifyRoleBinding implements the protobuf auth.ModifyRoleBinding RPC
func (a *apiServer) ModifyRoleBinding(ctx context.Context, req *auth.ModifyRoleBindingRequest) (resp *auth.ModifyRoleBindingResponse, retErr error) {
	var response *auth.ModifyRoleBindingResponse
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		var err error
		response, err = txn.ModifyRoleBinding(req)
		return errors.EnsureStack(err)
	}); err != nil {
		return nil, err
	}

	// If the request is not in a transaction, block until the cache is updated
	if req.Resource.Type == auth.ResourceType_CLUSTER {
		expected, err := rolesFromRoleSlice(req.Roles)
		if err != nil {
			return nil, err
		}
		if err := backoff.Retry(func() error {
			bindings, ok := a.clusterRoleBindingCache.Load().(*auth.RoleBinding)
			if !ok {
				return errors.New("cached cluster binding had unexpected type")
			}
			if !proto.Equal(bindings.Entries[req.Principal], expected) {
				return errors.New("waiting for cache to be updated")
			}
			return nil
		}, backoff.RetryEvery(100*time.Millisecond)); err != nil {
			return nil, err
		}
	}

	return response, nil
}

// GetRoleBindingInTransaction is identical to GetRoleBinding except that it can run inside
// an existing postgres transaction.  This is not an RPC.
func (a *apiServer) GetRoleBindingInTransaction(
	ctx context.Context,
	txnCtx *txncontext.TransactionContext,
	req *auth.GetRoleBindingRequest,
) (*auth.GetRoleBindingResponse, error) {
	if err := a.isActiveInTransaction(ctx, txnCtx); err != nil {
		return nil, err
	}

	var roleBindings auth.RoleBinding
	if err := a.roleBindings.ReadWrite(txnCtx.SqlTx).Get(ctx, authdb.ResourceKey(req.Resource), &roleBindings); err != nil && !col.IsErrNotFound(err) {
		return nil, errors.EnsureStack(err)
	}

	if roleBindings.Entries == nil {
		roleBindings.Entries = make(map[string]*auth.Roles)
	}
	return &auth.GetRoleBindingResponse{
		Binding: &roleBindings,
	}, nil
}

// GetRoleBinding implements the protobuf auth.GetRoleBinding RPC
func (a *apiServer) GetRoleBinding(ctx context.Context, req *auth.GetRoleBindingRequest) (resp *auth.GetRoleBindingResponse, retErr error) {
	var response *auth.GetRoleBindingResponse
	if err := a.env.TxnEnv.WithReadContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		var err error
		response, err = a.GetRoleBindingInTransaction(ctx, txnCtx, req)
		return err
	}); err != nil {
		return nil, err
	}
	return response, nil
}

// GetRobotToken implements the protobuf auth.GetRobotToken RPC
func (a *apiServer) GetRobotToken(ctx context.Context, req *auth.GetRobotTokenRequest) (resp *auth.GetRobotTokenResponse, retErr error) {
	// If the user specified a redundant robot: prefix, strip it. Colons are not permitted in robot names.
	subject := strings.TrimPrefix(req.Robot, auth.RobotPrefix)
	if strings.Contains(subject, ":") {
		return nil, errors.New("robot names cannot contain colons (':')")
	}

	subject = auth.RobotPrefix + subject

	// generate new token, and write to postgres
	var token string
	var err error
	if req.Ttl > 0 {
		token, err = a.generateAndInsertAuthToken(ctx, subject, req.Ttl)
	} else {
		token, err = a.generateAndInsertAuthTokenNoTTL(ctx, subject)
	}
	if err != nil {
		return nil, err
	}
	return &auth.GetRobotTokenResponse{
		Token: token,
	}, nil
}

// GetPipelineAuthTokenInTransaction is an internal API used to create a pipeline token for a given pipeline.
// Not an RPC.
func (a *apiServer) GetPipelineAuthTokenInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, pipeline *pps.Pipeline) (string, error) {
	if err := a.isActiveInTransaction(ctx, txnCtx); err != nil {
		return "", err
	}

	token := uuid.NewWithoutDashes()
	if err := a.insertAuthTokenNoTTLInTransaction(ctx, txnCtx, auth.HashToken(token), auth.PipelinePrefix+pipeline.String()); err != nil {
		return "", errors.Wrapf(err, "error storing token")
	} else {
		return token, nil
	}
}

// GetOIDCLogin implements the protobuf auth.GetOIDCLogin RPC
func (a *apiServer) GetOIDCLogin(ctx context.Context, req *auth.GetOIDCLoginRequest) (resp *auth.GetOIDCLoginResponse, retErr error) {
	authURL, state, err := a.GetOIDCLoginURL(ctx)
	if err != nil {
		return nil, err
	}
	return &auth.GetOIDCLoginResponse{
		LoginUrl: authURL,
		State:    state,
	}, nil
}

// RevokeAuthToken implements the protobuf auth.RevokeAuthToken RPC
func (a *apiServer) RevokeAuthToken(ctx context.Context, req *auth.RevokeAuthTokenRequest) (resp *auth.RevokeAuthTokenResponse, retErr error) {
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		resp, retErr = a.RevokeAuthTokenInTransaction(ctx, txnCtx, req)
		return retErr
	}); err != nil {
		return nil, err
	}
	return resp, retErr
}

func (a *apiServer) RevokeAuthTokenInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, req *auth.RevokeAuthTokenRequest) (resp *auth.RevokeAuthTokenResponse, retErr error) {
	if err := a.isActiveInTransaction(ctx, txnCtx); err != nil {
		return nil, err
	}

	n, err := a.deleteAuthToken(txnCtx.SqlTx, auth.HashToken(req.Token))
	if err != nil {
		return nil, err
	}
	return &auth.RevokeAuthTokenResponse{Number: n}, nil
}

// setGroupsForUserInternal is a helper function used by SetGroupsForUser, and
// also by handleOIDCExchangeInternal (which updates group membership information
// based on signed JWT claims). This does no auth checks, so the caller must do all
// relevant authorization.
func (a *apiServer) setGroupsForUserInternal(ctx context.Context, subject string, groups []string) error {
	return dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		members := a.members.ReadWrite(sqlTx)

		// Get groups to remove/add user from/to
		var removeGroups auth.Groups
		addGroups := addToSet(nil, groups...)
		if err := members.Get(ctx, subject, &removeGroups); err == nil {
			for _, group := range groups {
				if removeGroups.Groups[group] {
					removeGroups.Groups = removeFromSet(removeGroups.Groups, group)
					addGroups = removeFromSet(addGroups, group)
				}
			}
		}

		// Set groups for user
		if err := members.Put(ctx, subject, &auth.Groups{
			Groups: addToSet(nil, groups...),
		}); err != nil {
			return errors.EnsureStack(err)
		}

		// Remove user from previous groups
		groups := a.groups.ReadWrite(sqlTx)
		var membersProto auth.Users
		for group := range removeGroups.Groups {
			if err := groups.Upsert(ctx, group, &membersProto, func() error {
				membersProto.Usernames = removeFromSet(membersProto.Usernames, subject)
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
		}

		// Add user to new groups
		for group := range addGroups {
			if err := groups.Upsert(ctx, group, &membersProto, func() error {
				membersProto.Usernames = addToSet(membersProto.Usernames, subject)
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
		}

		return nil
	})
}

// SetGroupsForUser implements the protobuf auth.SetGroupsForUser RPC
func (a *apiServer) SetGroupsForUser(ctx context.Context, req *auth.SetGroupsForUserRequest) (resp *auth.SetGroupsForUserResponse, retErr error) {
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
	if err := a.checkCanonicalSubjects(req.Add); err != nil {
		return nil, err
	}
	// TODO(bryce) Skip canonicalization if the users can be found.
	if err := a.checkCanonicalSubjects(req.Remove); err != nil {
		return nil, err
	}

	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		members := a.members.ReadWrite(sqlTx)
		var groupsProto auth.Groups
		for _, username := range req.Add {
			if err := members.Upsert(ctx, username, &groupsProto, func() error {
				groupsProto.Groups = addToSet(groupsProto.Groups, req.Group)
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
		}
		for _, username := range req.Remove {
			if err := members.Upsert(ctx, username, &groupsProto, func() error {
				groupsProto.Groups = removeFromSet(groupsProto.Groups, req.Group)
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
		}

		groups := a.groups.ReadWrite(sqlTx)
		var membersProto auth.Users
		if err := groups.Upsert(ctx, req.Group, &membersProto, func() error {
			membersProto.Usernames = addToSet(membersProto.Usernames, req.Add...)
			membersProto.Usernames = removeFromSet(membersProto.Usernames, req.Remove...)
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
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
	members := a.members.ReadOnly()
	var groupsProto auth.Groups
	if err := members.Get(ctx, subject, &groupsProto); err != nil {
		if col.IsErrNotFound(err) {
			return []string{}, nil
		}
		return nil, errors.EnsureStack(err)
	}
	return setToList(groupsProto.Groups), nil
}

// getGroups is a helper function used primarily by the GRPC API GetGroups, but
// also by Authorize() and isAdmin().
func (a *apiServer) getGroupsInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, subject string) ([]string, error) {
	members := a.members.ReadWrite(txnCtx.SqlTx)
	var groupsProto auth.Groups
	if err := members.Get(ctx, subject, &groupsProto); err != nil {
		if col.IsErrNotFound(err) {
			return []string{}, nil
		}
		return nil, errors.EnsureStack(err)
	}
	return setToList(groupsProto.Groups), nil
}

// GetGroups implements the protobuf auth.GetGroups RPC
func (a *apiServer) GetGroups(ctx context.Context, req *auth.GetGroupsRequest) (resp *auth.GetGroupsResponse, retErr error) {
	callerInfo, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	groups, err := a.getGroups(ctx, callerInfo.Subject)
	if err != nil {
		return nil, err
	}
	return &auth.GetGroupsResponse{Groups: groups}, nil
}

// GetGroupsForPrincipal implements the protobuf auth.GetGroupsForPrincipal RPC
func (a *apiServer) GetGroupsForPrincipal(ctx context.Context, req *auth.GetGroupsForPrincipalRequest) (resp *auth.GetGroupsResponse, retErr error) {
	groups, err := a.getGroups(ctx, req.Principal)
	if err != nil {
		return nil, err
	}
	return &auth.GetGroupsResponse{Groups: groups}, nil
}

// GetUsers implements the protobuf auth.GetUsers RPC
func (a *apiServer) GetUsers(ctx context.Context, req *auth.GetUsersRequest) (resp *auth.GetUsersResponse, retErr error) {
	// Filter by group
	if req.Group != "" {
		var membersProto auth.Users
		if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
			groups := a.groups.ReadWrite(sqlTx)
			if err := groups.Get(ctx, req.Group, &membersProto); err != nil {
				return errors.EnsureStack(err)
			}
			return nil
		}); err != nil {
			return nil, err
		}

		return &auth.GetUsersResponse{Usernames: setToList(membersProto.Usernames)}, nil
	}

	membersCol := a.members.ReadOnly()
	groups := &auth.Groups{}
	var users []string
	if err := membersCol.List(ctx, groups, col.DefaultOptions(), func(user string) error {
		users = append(users, user)
		return nil
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &auth.GetUsersResponse{Usernames: users}, nil
}

// GetRolesForPermission implements the protobuf auth.GetRolesForPermission RPC
func (a *apiServer) GetRolesForPermission(ctx context.Context, req *auth.GetRolesForPermissionRequest) (resp *auth.GetRolesForPermissionResponse, retErr error) {
	return &auth.GetRolesForPermissionResponse{Roles: rolesForPermission(req.Permission)}, nil
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
	if err := a.isActive(ctx); err != nil {
		return nil, err
	}

	// try to lookup pre-computed subject
	if subject := internalauth.GetWhoAmI(ctx); subject != "" {
		return &auth.TokenInfo{
			Subject: subject,
		}, nil
	}

	// otherwise, we need a token
	token, err := auth.GetAuthToken(ctx)
	if err != nil {
		return nil, err
	}

	tokenInfo, lookupErr := a.lookupAuthTokenInfo(ctx, auth.HashToken(token))
	if lookupErr != nil {
		if col.IsErrNotFound(lookupErr) {
			return nil, auth.ErrBadToken
		}
		return nil, lookupErr
	}

	if err := a.expiredEnterpriseCheck(ctx, tokenInfo.Subject); err != nil {
		return nil, err
	}

	// verify token hasn't expired
	if tokenInfo.Expiration != nil {
		t := tokenInfo.Expiration.AsTime()
		if time.Now().After(t) {
			return nil, auth.ErrExpiredToken
		}
	}
	return tokenInfo, nil
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
	if subject == auth.AllClusterUsersSubject {
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
	case auth.PipelinePrefix, auth.InternalPrefix, auth.RobotPrefix, auth.PachPrefix, auth.UserPrefix, auth.GroupPrefix:
		break
	default:
		return errors.Errorf("subject has unrecognized prefix: %s", subject[:colonIdx+1])
	}
	return nil
}

// GetConfiguration implements the protobuf auth.GetConfiguration RPC.
func (a *apiServer) GetConfiguration(ctx context.Context, req *auth.GetConfigurationRequest) (resp *auth.GetConfigurationResponse, retErr error) {
	if err := a.isActive(ctx); err != nil {
		return nil, err
	}

	if !a.watchesEnabled {
		return nil, errors.New("watches are not enabled, unable to get current config")
	}

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
	if !a.watchesEnabled {
		return nil, errors.New("watches are not enabled, unable to set config")
	}

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
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		return errors.EnsureStack(a.authConfig.ReadWrite(sqlTx).Put(ctx, configKey, configToStore))
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
	if err := a.isActive(ctx); err != nil {
		return nil, err
	}

	extracted, err := a.listRobotTokens(ctx)
	if err != nil {
		return nil, err
	}
	return &auth.ExtractAuthTokensResponse{Tokens: extracted}, nil
}

func (a *apiServer) RestoreAuthToken(ctx context.Context, req *auth.RestoreAuthTokenRequest) (resp *auth.RestoreAuthTokenResponse, retErr error) {
	var ttl int64
	if req.Token.Expiration != nil {
		t := req.Token.Expiration.AsTime()
		ttl = int64(time.Until(t).Seconds())
		if ttl < 0 {
			return nil, auth.ErrExpiredToken
		}
	}

	if err := func() error {
		if ttl > 0 {
			return a.insertAuthToken(ctx, req.Token.HashedToken, req.Token.Subject, ttl)
		} else {
			return a.insertAuthTokenNoTTL(ctx, req.Token.HashedToken, req.Token.Subject)
		}
	}(); err != nil {
		return nil, errors.Wrapf(err, "error restoring auth token")
	}

	return &auth.RestoreAuthTokenResponse{}, nil
}

// implements the protobuf auth.DeleteExpiredAuthTokens RPC
func (a *apiServer) DeleteExpiredAuthTokens(ctx context.Context, req *auth.DeleteExpiredAuthTokensRequest) (*auth.DeleteExpiredAuthTokensResponse, error) {
	info, err := a.env.DB.Exec(`DELETE FROM auth.auth_tokens WHERE NOW() > expiration`)
	if err != nil {
		return nil, errors.Wrapf(err, "error deleting expired tokens")
	}
	if n, err := info.RowsAffected(); err == nil && n > 0 {
		log.Debug(ctx, "deleted expired auth tokens", zap.Int64("n", n))
	}
	return &auth.DeleteExpiredAuthTokensResponse{}, nil
}

func (a *apiServer) RevokeAuthTokensForUser(ctx context.Context, req *auth.RevokeAuthTokensForUserRequest) (resp *auth.RevokeAuthTokensForUserResponse, retErr error) {
	// Allow revoking auth tokens for pipelines, robots and IDP users,
	// but not the root token or PPS user
	if strings.HasPrefix(req.Username, auth.PachPrefix) {
		return nil, errors.New("cannot revoke tokens for pach: users")
	}
	n, err := a.deleteAuthTokensForSubject(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	return &auth.RevokeAuthTokensForUserResponse{Number: n}, nil
}

func (a *apiServer) deleteExpiredTokensRoutine(ctx context.Context) error {
	ctx = pctx.Child(ctx, "deleteExpiredTokensRoutine")
	if _, err := a.DeleteExpiredAuthTokens(ctx, &auth.DeleteExpiredAuthTokensRequest{}); err != nil {
		return err
	}

	go func(ctx context.Context) {
		ticker := backoff.NewTicker(backoff.NewConstantBackOff(time.Duration(cleanupIntervalHours) * time.Hour))
		defer ticker.Stop()
		for {

			if _, err := a.DeleteExpiredAuthTokens(ctx, &auth.DeleteExpiredAuthTokensRequest{}); err != nil {
				log.Error(ctx, "could not delete expired tokens", zap.Error(err))
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}(ctx)

	return nil
}

type dbTokenInfo struct {
	Subject     string     `db:"subject"`
	Expiration  *time.Time `db:"expiration"`
	HashedToken string     `db:"token_hash"`
}

func (ti dbTokenInfo) ToProto() *auth.TokenInfo {
	return &auth.TokenInfo{
		Subject:     ti.Subject,
		Expiration:  protoutil.MustTimestampFromPointer(ti.Expiration),
		HashedToken: ti.HashedToken,
	}
}

// we interpret an expiration value of NULL as "lives forever".
func (a *apiServer) lookupAuthTokenInfo(ctx context.Context, tokenHash string) (*auth.TokenInfo, error) {
	var tokenInfo dbTokenInfo
	err := a.env.DB.GetContext(ctx, &tokenInfo, `SELECT subject, expiration FROM auth.auth_tokens WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return nil, col.ErrNotFound{Type: "auth_tokens", Key: tokenHash}
	}
	return tokenInfo.ToProto(), nil
}

// we will sometimes have expiration values set in the past, since we only remove those values in the deleteExpiredTokensRoutine() goroutine
func (a *apiServer) listRobotTokens(ctx context.Context) ([]*auth.TokenInfo, error) {
	var robotTokens []dbTokenInfo
	if err := a.env.DB.SelectContext(ctx, &robotTokens,
		`SELECT token_hash, subject, expiration
		FROM auth.auth_tokens
		WHERE subject LIKE $1 || '%'`, auth.RobotPrefix); err != nil {
		return nil, errors.Wrapf(err, "error querying token")
	}
	result := make([]*auth.TokenInfo, len(robotTokens))
	for i, t := range robotTokens {
		result[i] = t.ToProto()
	}
	return result, nil
}

func (a *apiServer) generateAndInsertAuthToken(ctx context.Context, subject string, ttlSeconds int64) (string, error) {
	token := uuid.NewWithoutDashes()
	if err := a.insertAuthToken(ctx, auth.HashToken(token), subject, ttlSeconds); err != nil {
		return "", errors.Wrap(err, "could not generate and insert auth token")
	}
	return token, nil
}

func (a *apiServer) generateAndInsertAuthTokenNoTTL(ctx context.Context, subject string) (string, error) {
	token := uuid.NewWithoutDashes()
	if err := a.insertAuthTokenNoTTL(ctx, auth.HashToken(token), subject); err != nil {
		return "", err
	}

	return token, nil
}

// generates a token, and stores it's hash and supporting data in postgres
func (a *apiServer) insertAuthToken(ctx context.Context, tokenHash string, subject string, ttlSeconds int64) error {
	return dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		if err := authdb.EnsurePrincipal(ctx, tx, subject); err != nil {
			return errors.Wrapf(err, "ensure principal %v", subject)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO auth.auth_tokens (token_hash, subject, expiration)
		VALUES ($1, $2, NOW() + $3 * interval '1 sec')`, tokenHash, subject, ttlSeconds); err != nil {
			if dbutil.IsUniqueViolation(err) {
				return errors.New("cannot overwrite existing token with same hash")
			}
			return errors.Wrapf(err, "error storing token")
		}
		return nil
	})
}

// TODO(acohen4): replace this function with what's implemented in postgres-integration once it lands
func (a *apiServer) insertAuthTokenNoTTL(ctx context.Context, tokenHash string, subject string) error {
	return a.env.TxnEnv.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		err := a.insertAuthTokenNoTTLInTransaction(ctx, txnCtx, tokenHash, subject)
		return err
	})
}

func (a *apiServer) insertAuthTokenNoTTLInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, tokenHash string, subject string) error {
	if err := authdb.EnsurePrincipal(ctx, txnCtx.SqlTx, subject); err != nil {
		return errors.Wrapf(err, "ensure principal %v", subject)
	}
	if _, err := txnCtx.SqlTx.ExecContext(ctx,
		`INSERT INTO auth.auth_tokens (token_hash, subject)
		VALUES ($1, $2)`, tokenHash, subject); err != nil {
		if dbutil.IsUniqueViolation(err) {
			return errors.New("cannot overwrite existing token with same hash")
		}
		return errors.Wrapf(err, "error storing token")
	}
	return nil
}

func (a *apiServer) deleteAllAuthTokens(ctx context.Context, sqlTx *pachsql.Tx) error {
	if _, err := sqlTx.ExecContext(ctx, `DELETE FROM auth.auth_tokens`); err != nil {
		return errors.Wrapf(err, "error deleting all auth tokens")
	}
	return nil
}

func (a *apiServer) deleteAuthToken(sqlTx *pachsql.Tx, tokenHash string) (int64, error) {
	result, err := sqlTx.Exec(`DELETE FROM auth.auth_tokens WHERE token_hash=$1`, tokenHash)
	if err != nil {
		return 0, errors.Wrap(err, "error deleting token")
	}
	n, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "error getting rows affected")
	}
	return n, nil
}

func (a *apiServer) deleteAuthTokensForSubject(ctx context.Context, subject string) (n int64, retErr error) {
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		n, retErr = a.deleteAuthTokensForSubjectInTransaction(ctx, sqlTx, subject)
		return retErr
	}, dbutil.WithIsolationLevel(sql.LevelRepeatableRead)); err != nil {
		return 0, errors.Wrap(err, "error deleting auth token for subject")
	}
	return n, retErr
}

func (a *apiServer) deleteAuthTokensForSubjectInTransaction(ctx context.Context, tx *pachsql.Tx, subject string) (int64, error) {
	result, err := tx.ExecContext(ctx, `DELETE FROM auth.auth_tokens WHERE subject = $1`, subject)
	if err != nil {
		return 0, errors.Wrap(err, "error deleting all auth tokens")
	}
	n, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "error getting rows affected")
	}
	return n, nil
}

func authRepoResourceToRepo(resource *auth.Resource) (*pfs.Repo, error) {
	if resource.Type != auth.ResourceType_REPO {
		return nil, errors.Errorf("%v is not a repo", resource)
	}
	parts := strings.Split(resource.Name, "/")
	if len(parts) != 2 {
		return nil, errors.Errorf("invalid resource name %s", resource.Name)
	}
	return &pfs.Repo{Project: &pfs.Project{Name: parts[0]}, Name: parts[1]}, nil
}
