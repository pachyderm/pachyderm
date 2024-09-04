package pjs

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	pjsserver "github.com/pachyderm/pachyderm/v2/src/pjs"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
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
	// GetStorageClient follows a common pattern within the Pachyderm code base for getting a client to another server
	// that implements a gRPC service. Using a client is better than having a pointer to the server directly, which
	// could lead to tight coupling of services that ought to be independent.
	GetStorageClient func(ctx context.Context) storage.FilesetClient
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
