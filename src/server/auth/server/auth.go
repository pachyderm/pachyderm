package auth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"path"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/metadata"

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

	publicKey = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAoaPoEfv5RcVUbCuWNnOB
WtLHzcyQSe4SbtGGQom/X27iq/7s8dcebSsCd2cwYoyKihEQ5OlaghrhcxTTV5AN
39O6S0YnWjt/+4PWQQP3NpcEhqWj8RLPJtYq+JNrqlyjxBlca7vDcFSTa6iCqXay
iVD2OyTbWrD6KZ/YTSmSY8mY2qdYvHyp3Ue5ueH3rSkKRUjo4Jyjf59PntZD884P
yb9kC+weh/1KlbDQ4aV0U9p6DSBkW7dinOQj7a1/ikDoA9Nebnrkb1FF9Hr2+utO
We4e4yOViDzAP9hhQiBhOVR0F6wJF5i+NfuLit4tk5ViboogEZqIyuakTD6abSFg
UPqBTDDG0UsVqjnU5ysJ1DKQqALnOrxEKZoVXtH80/m7kgmeY3VDHCFt+WCSdaSq
1w8SoIpJAZPJpKlDjMxe+NqsX2qUODQ2KNkqfEqFtyUNZzfS9o9pEg/KJzDuDclM
oMQr1BG8vc3msX4UiGQPkohznwlCSGWf62IkSS6P8hQRCBKGRS5yGjmT3J+/chZw
Je46y8zNLV7t2pOL6UemdmDjTaMCt0YBc1FmG2eUipAWcHJWEHgQm2Yz6QjtBgvt
jFqnYeiDwdxU7CQD3oF9H+uVHqz8Jmmf9BxY9PhlMSUGPUsTpZ717ysL0UrBhQhW
xYp8vpeQ3by9WxPBE/WrxN8CAwEAAQ==
-----END PUBLIC KEY-----
`
)

type apiServer struct {
	protorpclog.Logger
	etcdClient *etcd.Client

	// 'activated' stores a timestamp that is effectively a cache of whether the
	// auth service has been activated. If 'activated' is 1/1/1, then the auth
	// service has been activated. Otherwise, return "NotActivatedError" until the
	// timestamp in 'activated' passes and then re-check etc to see if it has been
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

	s := &apiServer{
		Logger:     protorpclog.NewLogger("auth.API"),
		etcdClient: etcdClient,
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
	}
	s.activated.Store(time.Now().Add(1 * time.Minute))
	return s, nil
}

func (a *apiServer) Activate(ctx context.Context, req *authclient.ActivateRequest) (resp *authclient.ActivateResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())

	activated, err := a.isActivated(ctx)
	if err != nil {
		return nil, err
	}
	// Activating an already activated auth service should fail, because
	// otherwise anyone can just activate the service again and set
	// themselves as an admin.
	if activated {
		return nil, fmt.Errorf("already activated")
	}

	// Validate the activation code
	if err := validateActivationCode(req.ActivationCode); err != nil {
		return nil, fmt.Errorf("error validating activation code: %v", err)
	}

	// Initialize admins
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
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

	a.activated.Store(time.Time{})
	return &authclient.ActivateResponse{}, nil
}

func (a *apiServer) isActivated(ctx context.Context) (bool, error) {
	t := a.activated.Load().(time.Time)
	if t.IsZero() {
		return true, nil
	}
	if t.After(time.Now()) {
		return false, nil
	}

	adminCount, err := a.admins.ReadOnly(ctx).Count()
	if err != nil {
		return false, fmt.Errorf("error checking if auth service is activated")
	}

	// If there are admins, then the auth service has been activated
	if adminCount > 0 {
		a.activated.Store(time.Time{})
		return true, nil
	}
	a.activated.Store(time.Now().Add(1 * time.Minute))
	return false, nil
}

func (a *apiServer) Authenticate(ctx context.Context, req *authclient.AuthenticateRequest) (resp *authclient.AuthenticateResponse, retErr error) {
	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.Log(nil, nil, retErr, time.Since(start)) }(time.Now())
	activated, err := a.isActivated(ctx)
	if err != nil {
		return nil, err
	}
	if !activated {
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
	activated, err := a.isActivated(ctx)
	if err != nil {
		return nil, err
	}
	if !activated {
		return nil, authclient.NotActivatedError{}
	}

	user, err := a.getAuthenticatedUser(ctx)
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
		Authorized: req.Scope <= acl.Entries[user.Username],
	}, nil
}

func (a *apiServer) WhoAmI(ctx context.Context, req *authclient.WhoAmIRequest) (resp *authclient.WhoAmIResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	activated, err := a.isActivated(ctx)
	if err != nil {
		return nil, err
	}
	if !activated {
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
	if req.Repo == nil || req.Repo.Name == "" {
		return fmt.Errorf("invalid request: must set repo")
	}
	return nil
}

func (a *apiServer) SetScope(ctx context.Context, req *authclient.SetScopeRequest) (resp *authclient.SetScopeResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	activated, err := a.isActivated(ctx)
	if err != nil {
		return nil, err
	}
	if !activated {
		return nil, authclient.NotActivatedError{}
	}

	if err := validateSetScopeRequest(req); err != nil {
		return nil, err
	}
	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		acls := a.acls.ReadWrite(stm)

		var acl authclient.ACL
		if err := acls.Get(req.Repo.Name, &acl); err != nil {
			// Might be creating a new ACL. Check that 'req' sets the caller to be an owner
			// TODO(msteffen): check that the repo exists?
			if req.Username != user.Username || req.Scope != authclient.Scope_OWNER {
				return fmt.Errorf("ACL not found for repo %v", req.Repo.Name)
			}
			acl.Entries = make(map[string]authclient.Scope)
		}
		if len(acl.Entries) > 0 && !user.Admin &&
			acl.Entries[user.Username] != authclient.Scope_OWNER {
			return fmt.Errorf("user %v is not authorized to update ACL for repo %v", user, req.Repo.Name)
		}

		if req.Scope != authclient.Scope_NONE {
			acl.Entries[req.Username] = req.Scope
		} else {
			delete(acl.Entries, req.Username)
		}

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
	activated, err := a.isActivated(ctx)
	if err != nil {
		return nil, err
	}
	if !activated {
		return nil, authclient.NotActivatedError{}
	}

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) GetACL(ctx context.Context, req *authclient.GetACLRequest) (resp *authclient.GetACLResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	activated, err := a.isActivated(ctx)
	if err != nil {
		return nil, err
	}
	if !activated {
		return nil, authclient.NotActivatedError{}
	}

	return nil, fmt.Errorf("TODO")
}

func (a *apiServer) GetCapability(ctx context.Context, req *authclient.GetCapabilityRequest) (resp *authclient.GetCapabilityResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	activated, err := a.isActivated(ctx)
	if err != nil {
		return nil, err
	}
	if !activated {
		return nil, authclient.NotActivatedError{}
	}

	user, err := a.getAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}

	capability := uuid.NewWithoutDashes()
	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		tokens := a.tokens.ReadWrite(stm)
		// Capabilities are forver; they don't expire.
		return tokens.Put(hashToken(capability), &authclient.User{
			Username: user.Username,
			Admin:    user.Admin,
		})
	})
	if err != nil {
		return nil, fmt.Errorf("error storing capability for user %v: %v", user.Username, err)
	}

	return &authclient.GetCapabilityResponse{
		Capability: capability,
	}, nil
}

func (a *apiServer) RevokeAuthToken(ctx context.Context, req *authclient.RevokeAuthTokenRequest) (resp *authclient.RevokeAuthTokenResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	activated, err := a.isActivated(ctx)
	if err != nil {
		return nil, err
	}
	if !activated {
		return nil, authclient.NotActivatedError{}
	}

	// Even though anyone can revoke anyone's auth token, we still want
	// the user to be authenticated.
	if _, err = a.getAuthenticatedUser(ctx); err != nil {
		return nil, err
	}

	_, err = col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
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
		return nil, fmt.Errorf("error revoking token: %v", err)
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
		return nil, fmt.Errorf("error getting token: %v", err)
	}

	return &user, nil
}

type activationCode struct {
	Token     string
	Signature string
}

type token struct {
	Expiry string
}

// validateActivationCode checks the validity of an activation code
func validateActivationCode(code string) error {
	// Parse the public key.  If these steps fail, something is seriously
	// wrong and we should crash the service by panicking.
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		panic("failed to pem decode public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(fmt.Sprintf("failed to parse DER encoded public key: %+v", err))
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		panic("public key isn't an RSA key")
	}

	// Decode the base64-encoded activation code
	decodedActivationCode, err := base64.StdEncoding.DecodeString(code)
	if err != nil {
		return fmt.Errorf("activation code is not base64 encoded")
	}
	activationCode := &activationCode{}
	if err := json.Unmarshal(decodedActivationCode, &activationCode); err != nil {
		return fmt.Errorf("activation code is not valid JSON")
	}

	// Decode the signature
	decodedSignature, err := base64.StdEncoding.DecodeString(activationCode.Signature)
	if err != nil {
		return fmt.Errorf("signature is not base64 encoded")
	}

	// Compute the sha256 checksum of the token
	hashedToken := sha256.Sum256([]byte(activationCode.Token))

	// Verify that the signature is valid
	if err := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hashedToken[:], decodedSignature); err != nil {
		return fmt.Errorf("invalid signature in activation code")
	}

	// Unmarshal the token
	token := token{}
	if err := json.Unmarshal([]byte(activationCode.Token), &token); err != nil {
		return fmt.Errorf("token is not valid JSON")
	}

	// Parse the expiry
	expiry, err := time.Parse(time.RFC3339, token.Expiry)
	if err != nil {
		return fmt.Errorf("expiry is not valid ISO 8601 string")
	}

	// Check that the activation code has not expired
	if time.Now().After(expiry) {
		return fmt.Errorf("the activation code has expired")
	}

	return nil
}
