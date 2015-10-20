package jobserver // import "go.pachyderm.com/pachyderm/src/pps/jobserver"

import (
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
)

func NewAPIServer(
	persistAPIClient persist.APIClient,
	containerClient container.Client,
) pps.JobAPIServer {
	return newLogAPIServer(newAPIServer(persistAPIClient, containerClient))
}
