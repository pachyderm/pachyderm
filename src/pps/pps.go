package pps

import fmt "fmt"

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
