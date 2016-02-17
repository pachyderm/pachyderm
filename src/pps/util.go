package pps

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/pfs"
)

func JobRepo(job *Job) *pfs.Repo {
	return &pfs.Repo{Name: fmt.Sprintf("job-%s", job.Id)}
}

func PipelineRepo(pipeline *Pipeline) *pfs.Repo {
	return &pfs.Repo{Name: pipeline.Name}
}
