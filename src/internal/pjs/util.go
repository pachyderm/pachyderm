package pjs

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"google.golang.org/grpc"
)

type ClientOptions func(env *Env)

func NewTestClient(t testing.TB, db *pachsql.DB, storage *storage.Server, opts ...ClientOptions) pjs.APIClient {
	env := Env{
		DB:               db,
		GetPermissionser: &testPermitter{mode: permitterAllow},
		GetAuthToken:     func(ctx context.Context) (string, error) { return tu.RootToken, nil },
		Storage:          storage,
	}
	for _, opt := range opts {
		opt(&env)
	}
	srv := NewAPIServer(env)
	gc := grpcutil.NewTestClient(t, func(s *grpc.Server) {
		pjs.RegisterAPIServer(s, srv)
	})
	return pjs.NewAPIClient(gc)
}

type permitterEnum int

const (
	permitterAllow permitterEnum = iota
	permitterDeny
)

type testPermitter struct {
	override *auth.GetPermissionsResponse
	mode     permitterEnum
}

func (p *testPermitter) GetPermissions(ctx context.Context, req *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error) {
	switch {
	case p.override != nil:
		return p.override, nil
	case p.mode == permitterAllow:
		return &auth.GetPermissionsResponse{
			Permissions: []auth.Permission{auth.Permission_JOB_SKIP_CTX},
			Roles:       []string{auth.ClusterAdminRole},
		}, nil
	case p.mode == permitterDeny:
		return &auth.GetPermissionsResponse{}, nil
	default:
		return nil, nil
	}
}
