package pps

import (
	"fmt"

	. "github.com/pachyderm/pachyderm/src/client/pfs"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
)

func JobRepo(job *ppsclient.Job) *Repo {
	return &Repo{Name: fmt.Sprintf("job-%s", job.ID)}
}

func PipelineRepo(pipeline *ppsclient.Pipeline) *Repo {
	return &Repo{Name: pipeline.Name}
}
