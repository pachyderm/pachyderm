package pps

import (
	"bytes"
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	pfs "github.com/pachyderm/pachyderm/v2/src/pfs"
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

// ProjectNameField is not properly documented.
//
// TODO: document.
//
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

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *PipelinePicker) UnmarshalText(b []byte) error {
	parts := bytes.SplitN(b, []byte{'/'}, 2)
	var project, repo []byte
	switch len(parts) {
	case 0:
		return errors.New("invalid repo picker: empty")
	case 1:
		project, repo = nil, parts[0]
	case 2:
		project, repo = parts[0], parts[1]
	default:
		return errors.New("invalid repo picker: too many slashes")
	}
	rnp := &PipelinePicker_PipelineName{
		Project: &pfs.ProjectPicker{},
		Name:    string(repo),
	}
	p.Picker = &PipelinePicker_Name{
		Name: rnp,
	}
	if project != nil {
		if err := rnp.Project.UnmarshalText(project); err != nil {
			return errors.Wrapf(err, "unmarshal project %s", project)
		}
	}
	if err := p.ValidateAll(); err != nil {
		return err
	}
	return nil
}
