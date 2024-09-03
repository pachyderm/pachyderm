package pjs

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"

	pjsserver "github.com/pachyderm/pachyderm/v2/src/pjs"
)

type Env struct {
	DB               *pachsql.DB
	GetPermissionser GetPermissionser
}

func NewAPIServer(env Env) pjsserver.APIServer {
	return newAPIServer(env)
}

// GetPermissionser is an interface that currently exposes the Pachyderm Auth server's GetPermissions RPC.
// If PJS needs more GetPermissionser server RPCs, this interface should be changed and renamed.
type GetPermissionser interface {
	GetPermissions(ctx context.Context, req *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error)
}
