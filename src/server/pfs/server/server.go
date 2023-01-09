package server

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsload"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	go a.driver.master(pctx.Child(env.BackgroundContext, "master"))
	go a.driver.URLWorker(pctx.Child(env.BackgroundContext, "urlWorker"))
	go func() {
		pfsload.Worker(env.GetPachClient(pctx.Child(env.BackgroundContext, "pfsload")), env.TaskService) //nolint:errcheck
	}()
	taskSource := env.TaskService.NewSource(StorageTaskNamespace)
	go compactionWorker(pctx.Child(env.BackgroundContext, "compactionWorker"), taskSource, a.driver.storage) //nolint:errcheck
	return newValidatedAPIServer(a, env.AuthServer), nil
}

func NewSidecarAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	if env.PachwInSidecar {
		taskSource := env.TaskService.NewSource(StorageTaskNamespace)
		go compactionWorker(env.BackgroundContext, taskSource, a.driver.storage) //nolint:errcheck
	}
	return newValidatedAPIServer(a, env.AuthServer), nil
}

// NewPachwAPIServer is used when running pachd in Pachw Mode.
// In Pachw Mode, a pachd instance processes storage and URl related tasks via the task service.
func NewPachwAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	go a.driver.URLWorker(env.BackgroundContext)
	go func() { pfsload.Worker(env.GetPachClient(env.BackgroundContext), env.TaskService) }() //nolint:errcheck
	taskSource := env.TaskService.NewSource(StorageTaskNamespace)
	go compactionWorker(env.BackgroundContext, taskSource, a.driver.storage) //nolint:errcheck
	return newValidatedAPIServer(a, env.AuthServer), nil
}
