package auth

import (
	"crypto/sha256"
	"fmt"
	"path"
	"sync/atomic"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/google/go-github/github"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/pachyderm/pachyderm/src/client"
	authclient "github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

const (
	tokensPrefix = "/auth/tokens"
	aclsPrefix   = "/auth/acls"
	adminsPrefix = "/auth/admins"

	defaultTokenTTLSecs = 24 * 60 * 60
	authnToken          = "authn-token"
)

type apiServer struct {
	protorpclog.Logger
	etcdClient *etcd.Client

	// This atomic variable stores a boolean flag that indicates
	// whether the auth service has been activated.
	activated atomic.Value

	// tokens is a collection of hashedToken -> User mappings.
	tokens col.Collection
	// acls is a collection of repoName -> ACL mappings.
	acls col.Collection
	// admins is a collection of username -> User mappings.
	admins col.Collection
}

// NewAuthServer returns an implementation of auth.APIServer.
func NewAuthServer(etcdAddress string, etcdPrefix string) (authclient.APIServer, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return nil, fmt.Errorf("error constructing etcdClient: %v", err)
	}

	activated := atomic.Value{}
	activated.Store(false)
	return &apiServer{
		Logger:     protorpclog.NewLogger("auth.API"),
		etcdClient: etcdClient,
		activated:  activated,
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
			&authclient.User{},
			nil,
		),
	}, nil
}

func (a *apiServer) Activate(ctx context.Context, req *authclient.ActivateRequest) (resp *authclient.ActivateResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())

	// Activating an already activated auth service should fail, because
	// otherwise anyone can just activate the service again and set
	// themselves as an admin.
	if a.activated.Load().(bool) {
		return nil, fmt.Errorf("already activated")
	}

	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		admins := a.admins.ReadWrite(stm)
		for _, admin := range req.Admins {
			admins.Put(admin, &authclient.User{
				Username: admin,
				Admin:    true,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	a.activated.Store(true)
	return &authclient.ActivateResponse{}, nil
}

func (a *apiServer) Authenticate(ctx context.Context, req *authclient.AuthenticateRequest) (resp *authclient.AuthenticateResponse, retErr error) {
	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.Log(nil, nil, retErr, time.Since(start)) }(time.Now())
	if !a.activated.Load().(bool) {
		return nil, authclient.NotActivatedError{}
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: req.GithubToken,
		},
	)
	tc := oauth2.NewClient(ctx, ts)

	gclient := github.NewClient(tc)

	// Passing the empty string gets us the authenticated user
	user, _, err := gclient.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("error getting the authenticated user: %v", err)
	}

	username := user.GetName()

	// Check if the user is an admin.  If they are, authenticate them as
	// an admin.
	var u authclient.User
	var admin bool
	if err := a.admins.ReadOnly(ctx).Get(username, &u); err != nil {
		if _, ok := err.(col.ErrNotFound); !ok {
			return nil, fmt.Errorf("error checking if user %v is an admin: %v", username, err)
		}
	} else {
		admin = true
	}

	pachToken := uuid.NewWithoutDashes()

	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)
		return tokens.PutTTL(hashToken(pachToken), &authclient.User{
			Username: username,
			Admin:    admin,
		}, defaultTokenTTLSecs)
	})
	if err != nil {
		return nil, fmt.Errorf("error storing auth token for user %v: %v", username, err)
	}

	return &authclient.AuthenticateResponse{
		PachToken: pachToken,
	}, nil
}

func (a *apiServer) Authorize(ctx context.Context, req *authclient.AuthorizeRequest) (resp *authclient.AuthorizeResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.activated.Load().(bool) {
		return nil, authclient.NotActivatedError{}
	}

	user, err := a.getAuthorizedUser(ctx)
	if err != nil {
		return nil, err
	}

	if user.Admin {
		// admins are always authorized
		return &authclient.AuthorizeResponse{
			Authorized: true,
		}, nil
	}

	var acl authclient.ACL
	if err := a.acls.ReadOnly(ctx).Get(req.Repo.Name, &acl); err != nil {
		if _, ok := err.(col.ErrNotFound); ok {
			return nil, fmt.Errorf("ACL not found for repo %v", req.Repo.Name)
		}
		return nil, fmt.Errorf("error getting ACL for repo %v: %v", req.Repo.Name, err)
	}

	return &authclient.AuthorizeResponse{
		Authorized: req.Scope == acl.Entries[user.Username],
	}, nil
}

func (a *apiServer) SetScope(ctx context.Context, req *authclient.SetScopeRequest) (resp *authclient.SetScopeResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.activated.Load().(bool) {
		return nil, authclient.NotActivatedError{}
	}

	user, err := a.getAuthorizedUser(ctx)
	if err != nil {
		return nil, err
	}

	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		acls := a.acls.ReadWrite(stm)

		var acl authclient.ACL
		if err := acls.Get(req.Repo.Name, &acl); err != nil {
			return fmt.Errorf("ACL not found for repo %v", req.Repo.Name)
		}

		if acl.Entries[user.Username] != authclient.Scope_OWNER {
			return fmt.Errorf("user %v is not authorized to update ACL for repo %v", user, req.Repo.Name)
		}

		acl.Entries[req.Username] = req.Scope
		acls.Put(req.Repo.Name, &acl)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &authclient.SetScopeResponse{}, nil
}

func (a *apiServer) GetScope(ctx context.Context, req *authclient.GetScopeRequest) (resp *authclient.GetScopeResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.activated.Load().(bool) {
		return nil, authclient.NotActivatedError{}
	}

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) GetACL(ctx context.Context, req *authclient.GetACLRequest) (resp *authclient.GetACLResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	if !a.activated.Load().(bool) {
		return nil, authclient.NotActivatedError{}
	}

	return nil, fmt.Errorf("TODO")
}

// hashToken converts a token to a cryptographic hash.
// We don't want to store tokens verbatim in the database, as then whoever
// that has access to the database has access to all tokens.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}

func (a *apiServer) getAuthorizedUser(ctx context.Context) (*authclient.User, error) {
	token := ctx.Value(authnToken)
	if token == nil {
		return nil, fmt.Errorf("auth token not found in context")
	}

	tokenStr, ok := token.(string)
	if !ok {
		return nil, fmt.Errorf("auth token found in context is malformed")
	}

	var user authclient.User
	if err := a.tokens.ReadOnly(ctx).Get(hashToken(tokenStr), &user); err != nil {
		if _, ok := err.(col.ErrNotFound); ok {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("error getting token: %v", err)
	}

	return &user, nil
}
