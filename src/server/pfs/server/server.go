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
	taskSource := env.TaskService.NewSource(StorageTaskNamespace)
	go compactionWorker(env.BackgroundContext, taskSource, a.driver.storage) //nolint:errcheck
	return newValidatedAPIServer(a, env.AuthServer), nil
}

func NewSidecarAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	if !env.PachwEnabled {
		taskSource := env.TaskService.NewSource(StorageTaskNamespace)
		go compactionWorker(env.BackgroundContext, taskSource, a.driver.storage) //nolint:errcheck
	}
	return newValidatedAPIServer(a, env.AuthServer), nil
}

// NewPachwAPIServer is used for Pachw Mode. Pachw worker pool instances to process background tasks from the task service.
func NewPachwAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	go func() { pfsload.Worker(env.GetPachClient(env.BackgroundContext), env.TaskService) }() //nolint:errcheck
	taskSource := env.TaskService.NewSource(StorageTaskNamespace)
	go compactionWorker(env.BackgroundContext, taskSource, a.driver.storage) //nolint:errcheck
	return newValidatedAPIServer(a, env.AuthServer), nil
}
