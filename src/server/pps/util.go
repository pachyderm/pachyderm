package pps

import (
	"fmt"

	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
)

func JobRepo(job *ppsclient.Job) *pfsserver.Repo {
	return &pfsserver.Repo{Name: fmt.Sprintf("job-%s", job.ID)}
}

func PipelineRepo(pipeline *ppsclient.Pipeline) *pfsserver.Repo {
	return &pfsserver.Repo{Name: pipeline.Name}
}
