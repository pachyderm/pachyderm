package pipelineserver // import "go.pachyderm.com/pachyderm/src/pps/pipelineserver"

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
)

type APIServer interface {
	pps.PipelineAPIServer
	Start() error
}

func NewAPIServer(
	pfsAPIClient pfs.APIClient,
	jobAPIClient pps.JobAPIClient,
	persistAPIClient persist.APIClient,
) APIServer {
	return newAPIServer(
		pfsAPIClient,
		jobAPIClient,
		persistAPIClient,
	)
}
