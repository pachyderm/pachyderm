package pps

import (
	"fmt"

	"go.uber.org/zap"
)

func (j *Job) String() string {
	return fmt.Sprintf("%s@%s", j.Pipeline, j.Id)
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

// The following are zap fields for log lines that can be parsed into a LogMessage.  The field name
// must be equal to the `protobuf:"...,json="` name.
func ProjectNameField(name string) zap.Field {
	return zap.String("projectName", name)
}

func PipelineNameField(name string) zap.Field {
	return zap.String("pipelineName", name)
}

func JobIDField(id string) zap.Field {
	return zap.String("jobId", id)
}

func WorkerIDField(id string) zap.Field {
	return zap.String("workerId", id)
}

func DatumIDField(id string) zap.Field {
	return zap.String("datumId", id)
}

func MasterField(master bool) zap.Field {
	return zap.Bool("master", master)
}

func UserField(user bool) zap.Field {
	return zap.Bool("user", user)
}

func DataField(data []*InputFile) zap.Field {
	if len(data) == 0 {
		return zap.Skip()
	}
	return zap.Objects("data", data)
}
