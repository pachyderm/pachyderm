package pipelineserver

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
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
