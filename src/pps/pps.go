package pps

import (
	fmt "fmt"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func (j *Job) String() string {
	return fmt.Sprintf("%s@%s", j.Pipeline.Name, j.ID)
}

func (p *Pipeline) String() string {
	if p == nil {
		return ""
	}
	projectName, pipelineName := p.Project.GetName(), p.Name
	if projectName == "" {
		return pipelineName
	}
	return projectName + "/" + pipelineName
}

// EnsurePipelineProject ensures that a pipeline’s repo is valid.  It does nothing
// if pipeline is nil.
func EnsurePipelineProject(p *Pipeline) {
	if p == nil {
		return
	}
	if p.Project.GetName() == "" {
		p.Project = &pfs.Project{Name: pfs.DefaultProjectName}
	}
}
