package pps

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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

var ErrInvalidPipelineSpec = errors.New("invalid pipeline spec")

// PipelineSpec creates a pipeline spec from a CreatePipelineRequest.
// If the request has a DetailsJSON field, it unmarshals it; otherwise, it fills
// the spec from the request.
func (req *CreatePipelineRequest) PipelineSpec() (*PipelineSpec, error) {
	if js := req.GetDetailsJson(); js != "" {
		var ps = new(PipelineSpec)
		if err := protojson.Unmarshal([]byte(js), ps); err != nil {
			return ps, errors.Wrap(ErrInvalidPipelineSpec, err.Error())
		}
		return ps, nil
	}
	return &PipelineSpec{
		Pipeline:                req.Pipeline,
		TfJob:                   req.TfJob,
		Transform:               req.Transform,
		ParallelismSpec:         req.ParallelismSpec,
		Egress:                  req.Egress,
		OutputBranch:            req.OutputBranch,
		S3Out:                   req.S3Out,
		ResourceRequests:        req.ResourceRequests,
		ResourceLimits:          req.ResourceLimits,
		SidecarResourceLimits:   req.SidecarResourceLimits,
		Input:                   req.Input,
		Description:             req.Description,
		Service:                 req.Service,
		Spout:                   req.Spout,
		DatumSetSpec:            req.DatumSetSpec,
		DatumTimeout:            req.DatumTimeout,
		JobTimeout:              req.JobTimeout,
		Salt:                    req.Salt,
		DatumTries:              req.DatumTries,
		SchedulingSpec:          req.SchedulingSpec,
		PodSpec:                 req.PodSpec,
		PodPatch:                req.PodPatch,
		SpecCommit:              req.SpecCommit,
		Metadata:                req.Metadata,
		ReprocessSpec:           req.ReprocessSpec,
		Autoscaling:             req.Autoscaling,
		Tolerations:             req.Tolerations,
		SidecarResourceRequests: req.SidecarResourceRequests,
	}, nil

}

// ValidateJSONSpec checks that a JSON spec is a valid spec.  Currently it only
// checks that it unmarshals cleanly.
//
// TODO(CORE-1809): add static semantic checks
func ValidateJSONPipelineSpec(spec string) error {
	var ps PipelineSpec
	if err := protojson.Unmarshal([]byte(spec), &ps); err != nil {
		return errors.Wrap(ErrInvalidPipelineSpec, err.Error())
	}
	return nil
}
