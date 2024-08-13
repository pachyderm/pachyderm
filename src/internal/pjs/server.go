package pjs

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	pjsserver "github.com/pachyderm/pachyderm/v2/src/pjs"
)

type Env struct {
	DB *pachsql.DB
}

func NewAPIServer(env Env) pjsserver.APIServer {
	return newAPIServer(env)
}
