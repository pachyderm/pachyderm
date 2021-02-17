package server

import (
	"github.com/pachyderm/pachyderm/v2/src/catalog"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
)

func NewCatalogServer(env *serviceenv.ServiceEnv) (catalog.APIServer, error) {
	return newAPIServer(env)
}
