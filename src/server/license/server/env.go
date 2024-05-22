package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/server/enterprise"
)

// Env is the dependencies required for a license API Server
type Env struct {
	DB               *pachsql.DB
	Listener         collection.PostgresListener
	Config           *pachconfig.Configuration
	EnterpriseServer enterprise.APIServer
}
