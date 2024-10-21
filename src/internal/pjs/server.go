package pjs

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	pjsserver "github.com/pachyderm/pachyderm/v2/src/pjs"
)

// Env defines the PJS API server's dependencies. optional fields ought to go somewhere else.
type Env struct {
	// DB is the postgres database client.
	DB *pachsql.DB
	// GetPermissionser is a subset of Pachyderm Auth server's gRPC service.
	GetPermissionser GetPermissionser
	// GetAuthToken doesn't go through an RPC, otherwise it would make more sense to combine it
	// with the GetPermissionser interface. Using a closure here makes it easier to mock in tests.
	GetAuthToken func(context.Context) (string, error)
	Storage      *storage.Server
}

func NewAPIServer(env Env) pjsserver.APIServer {
	return &apiServer{
		env:          env,
		pollInterval: 5 * time.Second,
	}
}

// GetPermissionser is an interface that currently exposes the Pachyderm Auth server's GetPermissions RPC.
// If PJS needs more GetPermissionser server RPCs, this interface should be changed and renamed.
type GetPermissionser interface {
	GetPermissions(ctx context.Context, req *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error)
}
