package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsload"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	go a.driver.master(env.BackgroundContext)
	go a.driver.URLWorker(env.BackgroundContext)
	go func() { pfsload.Worker(env.GetPachClient(env.BackgroundContext), env.TaskService) }() //nolint:errcheck
	return newValidatedAPIServer(a, env.AuthServer), nil
}

func NewSidecarAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env.AuthServer), nil
}
