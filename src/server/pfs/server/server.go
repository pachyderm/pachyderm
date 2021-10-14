package server

import (
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	go a.driver.master(env.BackgroundContext)
	return newValidatedAPIServer(a, env.AuthServer), nil
}

func NewSidecarAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env.AuthServer), nil
}
