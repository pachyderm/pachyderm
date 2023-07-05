package pps

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/pfs"

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

type PipelineSpec struct {
	Pipeline *Pipeline `json:"pipeline,omitempty"`
	// tf_job encodes a Kubeflow TFJob spec. Pachyderm uses this to create TFJobs
	// when running in a kubernetes cluster on which kubeflow has been installed.
	// Exactly one of 'tf_job' and 'transform' should be set
	TFJob           *TFJob           `json:"tf_job,omitempty"`
	Transform       *Transform       `json:"transform,omitempty"`
	ParallelismSpec *ParallelismSpec `json:"parallelism_spec,omitempty"`
	Egress          *Egress          `json:"egress,omitempty"`
	Update          bool             `json:"update,omitempty"`
	OutputBranch    string           `json:"output_branch,omitempty"`
	// s3_out, if set, requires a pipeline's user to write to its output repo
	// via Pachyderm's s3 gateway (if set, workers will serve Pachyderm's s3
	// gateway API at http://<pipeline>-s3.<namespace>/<job id>.out/my/file).
	// In this mode /pfs_v2/out won't be walked or uploaded, and the s3 gateway
	// service in the workers will allow writes to the job's output commit
	S3Out                 bool          `json:"s3_out,omitempty"`
	ResourceRequests      *ResourceSpec `json:"resource_requests,omitempty"`
	ResourceLimits        *ResourceSpec `json:"resource_limits,omitempty"`
	SidecarResourceLimits *ResourceSpec `json:"sidecar_resource_limits,omitempty"`
	Input                 *Input        `json:"input,omitempty"`
	Description           string        `json:"description,omitempty"`
	// Reprocess forces the pipeline to reprocess all datums.
	// It only has meaning if Update is true
	Reprocess               bool            `json:"reprocess,omitempty"`
	Service                 *Service        `json:"service,omitempty"`
	Spout                   *Spout          `json:"spout,omitempty"`
	DatumSetSpec            *DatumSetSpec   `json:"datum_set_spec,omitempty"`
	DatumTimeout            time.Duration   `json:"datum_timeout,omitempty"`
	JobTimeout              time.Duration   `json:"job_timeout,omitempty"`
	Salt                    string          `json:"salt,omitempty"`
	DatumTries              int64           `json:"datum_tries,omitempty"`
	SchedulingSpec          *SchedulingSpec `json:"scheduling_spec,omitempty"`
	PodSpec                 string          `json:"pod_spec,omitempty"`
	PodPatch                string          `json:"pod_patch,omitempty"`
	SpecCommit              *pfs.Commit     `json:"spec_commit,omitempty"`
	Metadata                *Metadata       `json:"metadata,omitempty"`
	ReprocessSpec           string          `json:"reprocess_spec,omitempty"`
	Autoscaling             bool            `json:"autoscaling,omitempty"`
	Tolerations             []*Toleration   `json:"tolerations,omitempty"`
	SidecarResourceRequests *ResourceSpec   `json:"sidecar_resource_requests,omitempty"`
}

var ErrInvalidPipelineSpec = errors.New("invalid pipeline spec")

// ValidateJSONSpec checks that a JSON spec is a valid spec.  Currently it only
// checks that it unmarshals cleanly and has no unknown fields.
//
// TODO(CORE-1809): add static semantic checks
func ValidateJSONPipelineSpec(spec string) error {
	var d = json.NewDecoder(strings.NewReader(spec))

	d.DisallowUnknownFields()
	if err := d.Decode(&PipelineSpec{}); err != nil {
		return errors.Wrap(ErrInvalidPipelineSpec, err.Error())
	}
	return nil
}
