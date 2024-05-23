package pps

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
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

func GetAlerts(pi *PipelineInfo) []string {
	var alerts []string
	zero := &timestamppb.Timestamp{}
	if pi.Details.WorkersStartedAt.AsTime() != zero.AsTime() &&
		time.Since(pi.Details.WorkersStartedAt.AsTime()) > pi.Details.MaximumExpectedUptime.AsDuration() {
		exceededTime := time.Since(pi.Details.WorkersStartedAt.AsTime()) - pi.Details.MaximumExpectedUptime.AsDuration()
		alerts = append(alerts, fmt.Sprintf("%s/%s: Started running at %v, exceeded maximum uptime by %s", pi.Pipeline.Project.GetName(), pi.Pipeline.Name, pi.Details.WorkersStartedAt.AsTime(), exceededTime.String()))
	}
	return alerts
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
