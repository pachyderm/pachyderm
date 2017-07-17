package auth

import (
	"fmt"
	"path"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/google/go-github/github"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/pachyderm/pachyderm/src/client"
	authclient "github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
)

const (
	authEtcdPrefix      = "auth_tokens"
	defaultTokenTTLSecs = 24 * 60 * 60
)

type apiServer struct {
	protorpclog.Logger
	etcdClient *etcd.Client
	etcdPrefix string
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

	return &apiServer{
		Logger:     protorpclog.NewLogger("auth.API"),
		etcdClient: etcdClient,
		etcdPrefix: path.Join(etcdPrefix, authEtcdPrefix),
	}, nil
}

func (a *apiServer) Authenticate(ctx context.Context, req *authclient.AuthenticateRequest) (resp *authclient.AuthenticateResponse, retErr error) {
	// We don't want to actually log the request/response since they contain
	// credentials.
	defer func(start time.Time) { a.Log(nil, nil, retErr, time.Since(start)) }(time.Now())

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
	pachToken := uuid.NewWithoutDashes()

	lease, err := a.etcdClient.Grant(ctx, defaultTokenTTLSecs)
	if err != nil {
		return nil, fmt.Errorf("error granting token TTL: %v", err)
	}

	_, err = a.etcdClient.Put(ctx, path.Join(a.etcdPrefix, pachToken), username, etcd.WithLease(lease.ID))
	if err != nil {
		return nil, fmt.Errorf("error storing the auth token: %v", err)
	}

	return &authclient.AuthenticateResponse{
		PachToken: pachToken,
	}, nil
}
