package jobserver

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/container"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/jobserver/run"
	"github.com/pachyderm/pachyderm/src/pps/persist"
)

func NewAPIServer(
	pfsAPIClient pfs.APIClient,
	persistAPIClient persist.APIClient,
	containerClient container.Client,
	pfsMountDir string,
	options jobserverrun.JobRunnerOptions,
) pps.JobAPIServer {
	return newAPIServer(
		persistAPIClient,
		jobserverrun.NewJobRunner(
			pfsAPIClient,
			persistAPIClient,
			containerClient,
			pfsMountDir,
			options,
		),
	)
}
