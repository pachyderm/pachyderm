package snapshot

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	snapshotserver "github.com/pachyderm/pachyderm/v2/src/snapshot"
)

type Env struct {
	DB    *pachsql.DB
	store *fileset.Storage
}

func NewAPIServer(env Env) snapshotserver.APIServer {
	return &apiServer{
		env: env,
	}
}
