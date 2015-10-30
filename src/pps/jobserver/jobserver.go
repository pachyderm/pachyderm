package jobserver

import (
	"github.com/pachyderm/pachyderm/src/pkg/container"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/jobserver/run"
	"github.com/pachyderm/pachyderm/src/pps/persist"
)

func NewAPIServer(
	persistAPIClient persist.APIClient,
	containerClient container.Client,
	pfsMountDir string,
) pps.JobAPIServer {
	return newAPIServer(
		persistAPIClient,
		jobserverrun.NewJobRunner(
			persistAPIClient,
			containerClient,
			pfsMountDir,
		),
	)
}
