package ppsutil

import (
	"encoding/json"
	"fmt"
	"time"
	"unicode"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cronutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsfile"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	v1 "k8s.io/api/core/v1"
)

const (
	maxPipelineNameLength = 51

	// dnsLabelLimit is the maximum length of a ReplicationController
	// or Service name.
	dnsLabelLimit = 63

	// DefaultDatumTries is the default number of times a datum will be tried
	// before we give up and consider the job failed.
	DefaultDatumTries = 3
)

func ValidatePipelineRequest(request *pps.CreatePipelineRequest) error {
	if request.Pipeline == nil {
		return errors.New("invalid pipeline spec: request.Pipeline cannot be nil")
	}
	if request.Pipeline.Name == "" {
		return errors.New("invalid pipeline spec: request.Pipeline.Name cannot be empty")
	}
	if err := ancestry.ValidateName(request.Pipeline.Name); err != nil {
		return errors.Wrapf(err, "invalid pipeline name")
	}
	if len(request.Pipeline.Name) > maxPipelineNameLength {
		return errors.Errorf("pipeline name %q is %d characters longer than the %d max",
			request.Pipeline.Name, len(request.Pipeline.Name)-maxPipelineNameLength, maxPipelineNameLength)
	}
	// TODO(CORE-1489): Remove dependency of name length on Kubernetes
	// resource naming convention.
	if k8sName := PipelineRcName(&pps.PipelineInfo{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: request.Pipeline.GetProject().GetName()}, Name: request.Pipeline.Name}, Version: 99}); len(k8sName) > dnsLabelLimit {
		return errors.Errorf("Kubernetes name %q is %d characters longer than the %d max", k8sName, len(k8sName)-dnsLabelLimit, dnsLabelLimit)
	}
	// TODO(msteffen) eventually TFJob and Transform will be alternatives, but
	// currently TFJob isn't supported
	if request.TFJob != nil {
		return errors.New("embedding TFJobs in pipelines is not supported yet")
	}
	if request.S3Out && ((request.Service != nil) || (request.Spout != nil)) {
		return errors.New("s3 output is not supported in spouts or services")
	}
	if request.Transform == nil {
		return errors.Errorf("pipeline must specify a transform")
	}
	if request.ReprocessSpec != "" &&
		request.ReprocessSpec != client.ReprocessSpecUntilSuccess &&
		request.ReprocessSpec != client.ReprocessSpecEveryJob {
		return errors.Errorf("invalid pipeline spec: ReprocessSpec must be one of '%s' or '%s'",
			client.ReprocessSpecUntilSuccess, client.ReprocessSpecEveryJob)
	}
	if request.Spout != nil && request.Autoscaling {
		return errors.Errorf("autoscaling can't be used with spouts (spouts aren't triggered externally)")
	}
	var tolErrs error
	for i, t := range request.GetTolerations() {
		if _, err := transformToleration(t); err != nil {
			errors.JoinInto(&tolErrs, errors.Errorf("toleration %d/%d: %v", i+1, len(request.GetTolerations()), err))
		}
	}
	if tolErrs != nil {
		return tolErrs
	}
	return nil
}

// setPipelineDefaults sets the default values for a pipeline info
func setPipelineDefaults(pipelineInfo *pps.PipelineInfo) error {
	if pipelineInfo.Details.Transform.Image == "" {
		pipelineInfo.Details.Transform.Image = DefaultUserImage
	}
	setInputDefaults(pipelineInfo.Pipeline.Name, pipelineInfo.Details.Input)
	if pipelineInfo.Details.OutputBranch == "" {
		// Output branches default to master
		pipelineInfo.Details.OutputBranch = "master"
	}
	if pipelineInfo.Details.DatumTries == 0 {
		pipelineInfo.Details.DatumTries = DefaultDatumTries
	}
	if pipelineInfo.Details.Service != nil {
		if pipelineInfo.Details.Service.Type == "" {
			pipelineInfo.Details.Service.Type = string(v1.ServiceTypeNodePort)
		}
	}
	if pipelineInfo.Details.Spout != nil && pipelineInfo.Details.Spout.Service != nil && pipelineInfo.Details.Spout.Service.Type == "" {
		pipelineInfo.Details.Spout.Service.Type = string(v1.ServiceTypeNodePort)
	}
	if pipelineInfo.Details.ReprocessSpec == "" {
		pipelineInfo.Details.ReprocessSpec = client.ReprocessSpecUntilSuccess
	}
	return nil
}

func setInputDefaults(pipelineName string, input *pps.Input) {
	pps.SortInput(input)
	now := time.Now()
	nCreatedBranches := make(map[string]int)
	pps.VisitInput(input, func(input *pps.Input) error { //nolint:errcheck
		if input.Pfs != nil {
			if input.Pfs.Branch == "" {
				if input.Pfs.Trigger != nil {
					// We start counting trigger branches at 1
					nCreatedBranches[input.Pfs.Repo]++
					input.Pfs.Branch = fmt.Sprintf("%s-trigger-%d", pipelineName, nCreatedBranches[input.Pfs.Repo])
					if input.Pfs.Trigger.Branch == "" {
						input.Pfs.Trigger.Branch = "master"
					}
				} else {
					input.Pfs.Branch = "master"
				}
			}
			if input.Pfs.Name == "" {
				input.Pfs.Name = input.Pfs.Repo
			}
			if input.Pfs.RepoType == "" {
				input.Pfs.RepoType = pfs.UserRepoType
			}
			if input.Pfs.Glob != "" {
				input.Pfs.Glob = pfsfile.CleanPath(input.Pfs.Glob)
			}
		}
		if input.Cron != nil {
			if input.Cron.Start == nil {
				start, _ := types.TimestampProto(now)
				input.Cron.Start = start
			}
			if input.Cron.Repo == "" {
				input.Cron.Repo = fmt.Sprintf("%s_%s", pipelineName, input.Cron.Name)
			}
		}
		return nil
	})
}

func validateTransform(transform *pps.Transform) error {
	if transform == nil {
		return errors.Errorf("pipeline must specify a transform")
	}
	if transform.Image == "" {
		return errors.Errorf("pipeline transform must contain an image")
	}
	return nil
}

func validatePipeline(pipelineInfo *pps.PipelineInfo) error {
	if pipelineInfo.Pipeline == nil {
		return errors.New("invalid pipeline spec: Pipeline field cannot be nil")
	}
	if pipelineInfo.Pipeline.Name == "" {
		return errors.New("invalid pipeline spec: Pipeline.Name cannot be empty")
	}
	if err := ancestry.ValidateName(pipelineInfo.Pipeline.Name); err != nil {
		return errors.Wrapf(err, "invalid pipeline name")
	}
	first := rune(pipelineInfo.Pipeline.Name[0])
	if !unicode.IsLetter(first) && !unicode.IsDigit(first) {
		return errors.Errorf("pipeline names must start with an alphanumeric character")
	}
	if len(pipelineInfo.Pipeline.Name) > 63 {
		return errors.Errorf("pipeline name is %d characters long, but must have at most 63: %q",
			len(pipelineInfo.Pipeline.Name), pipelineInfo.Pipeline.Name)
	}
	if err := validateTransform(pipelineInfo.Details.Transform); err != nil {
		return errors.Wrapf(err, "invalid transform")
	}
	if err := validateInput(pipelineInfo.Pipeline, pipelineInfo.Details.Input); err != nil {
		return err
	}
	if err := validateEgress(pipelineInfo.Pipeline.Name, pipelineInfo.Details.Egress); err != nil {
		return err
	}
	if pipelineInfo.Details.ParallelismSpec != nil {
		if pipelineInfo.Details.Service != nil && pipelineInfo.Details.ParallelismSpec.Constant != 1 {
			return errors.New("services can only be run with a constant parallelism of 1")
		}
	}
	if pipelineInfo.Details.OutputBranch == "" {
		return errors.New("pipeline needs to specify an output branch")
	}
	if pipelineInfo.Details.JobTimeout != nil {
		_, err := types.DurationFromProto(pipelineInfo.Details.JobTimeout)
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
	if pipelineInfo.Details.DatumTimeout != nil {
		_, err := types.DurationFromProto(pipelineInfo.Details.DatumTimeout)
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
	if pipelineInfo.Details.PodSpec != "" && !json.Valid([]byte(pipelineInfo.Details.PodSpec)) {
		return errors.Errorf("malformed PodSpec")
	}
	if pipelineInfo.Details.PodPatch != "" && !json.Valid([]byte(pipelineInfo.Details.PodPatch)) {
		return errors.Errorf("malformed PodPatch")
	}
	if pipelineInfo.Details.Service != nil {
		validServiceTypes := map[v1.ServiceType]bool{
			v1.ServiceTypeClusterIP:    true,
			v1.ServiceTypeLoadBalancer: true,
			v1.ServiceTypeNodePort:     true,
		}

		if !validServiceTypes[v1.ServiceType(pipelineInfo.Details.Service.Type)] {
			return errors.Errorf("the following service type %s is not allowed", pipelineInfo.Details.Service.Type)
		}
	}
	if pipelineInfo.Details.Spout != nil {
		if pipelineInfo.Details.Spout.Service == nil && pipelineInfo.Details.Input != nil {
			return errors.Errorf("spout pipelines (without a service) must not have an input")
		}
	}
	return nil
}

func validateEgress(pipelineName string, egress *pps.Egress) error {
	if egress == nil {
		return nil
	}
	return pfsserver.ValidateSQLDatabaseEgress(egress.GetSqlDatabase())
}

func validateInput(pipeline *pps.Pipeline, input *pps.Input) error {
	if err := validateNames(make(map[string]bool), input); err != nil {
		return err
	}
	return pps.VisitInput(input, func(input *pps.Input) error {
		set := false
		if input.Pfs != nil {
			if input.Pfs.Project == "" {
				input.Pfs.Project = pipeline.Project.GetName()
			}
			set = true
			switch {
			case input.Pfs.Repo == "":
				return errors.Errorf("input must specify a repo")
			case input.Pfs.Repo == "out" && input.Pfs.Name == "":
				return errors.Errorf("inputs based on repos named \"out\" must have " +
					"'name' set, as pachyderm already creates /pfs/out to collect " +
					"job output")
			case input.Pfs.Branch == "":
				return errors.Errorf("input must specify a branch")
			case !input.Pfs.S3 && len(input.Pfs.Glob) == 0:
				return errors.Errorf("input must specify a glob")
			case input.Pfs.S3 && input.Pfs.Glob != "/":
				return errors.Errorf("inputs that set 's3' to 'true' must also set " +
					"'glob', to \"/\", as the S3 gateway is only able to expose data " +
					"at the commit level")
			case input.Pfs.S3 && input.Pfs.Lazy:
				return errors.Errorf("input cannot specify both 's3' and 'lazy', as " +
					"'s3' requires input data to be accessed via Pachyderm's S3 " +
					"gateway rather than the file system")
			case input.Pfs.S3 && input.Pfs.EmptyFiles:
				return errors.Errorf("input cannot specify both 's3' and " +
					"'empty_files', as 's3' requires input data to be accessed via " +
					"Pachyderm's S3 gateway rather than the file system")
			case input.Pfs.Commit != "":
				return errors.Errorf("input cannot come from a commit; use a branch with head pointing to the commit")
			}
		}
		if input.Cross != nil {
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
		}
		if input.Join != nil {
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
			if ContainsS3Inputs(input) {
				// The best datum semantics for s3 inputs embedded in join expressions
				// are not yet clear, and we see no use case for them yet, so block
				// them until we know how they should work
				return errors.Errorf("S3 inputs in join expressions are not supported")
			}
		}
		if input.Group != nil {
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
			if ContainsS3Inputs(input) {
				// See above for "joins"; block s3 inputs in group expressions until
				// we know how they should work
				return errors.Errorf("S3 inputs in group expressions are not supported")
			}
		}
		if input.Union != nil {
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
			if ContainsS3Inputs(input) {
				// See above for "joins"; block s3 inputs in union expressions until
				// we know how they should work
				return errors.Errorf("S3 inputs in union expressions are not supported")
			}
		}
		if input.Cron != nil {
			if input.Cron.Project == "" {
				input.Cron.Project = pipeline.Project.GetName()
			}
			if set {
				return errors.Errorf("multiple input types set")
			}
			set = true
			if _, err := cronutil.ParseCronExpression(input.Cron.Spec); err != nil {
				return errors.Wrapf(err, "error parsing cron-spec")
			}
		}
		if !set {
			return errors.Errorf("no input set")
		}
		return nil
	})
}

func validateNames(names map[string]bool, input *pps.Input) error {
	switch {
	case input == nil:
		return nil // spouts can have nil input
	case input.Pfs != nil:
		if err := validateName(input.Pfs.Name); err != nil {
			return err
		}
		if names[input.Pfs.Name] {
			return errors.Errorf(`name "%s" was used more than once`, input.Pfs.Name)
		}
		names[input.Pfs.Name] = true
	case input.Cron != nil:
		if err := validateName(input.Cron.Name); err != nil {
			return err
		}
		if names[input.Cron.Name] {
			return errors.Errorf(`name "%s" was used more than once`, input.Cron.Name)
		}
		names[input.Cron.Name] = true
	case input.Union != nil:
		for _, input := range input.Union {
			namesCopy := make(map[string]bool)
			merge(names, namesCopy)
			if err := validateNames(namesCopy, input); err != nil {
				return err
			}
			// we defer this because subinputs of a union input are allowed to
			// have conflicting names but other inputs that are, for example,
			// crossed with this union cannot conflict with any of the names it
			// might present
			defer merge(namesCopy, names)
		}
	case input.Cross != nil:
		for _, input := range input.Cross {
			if err := validateNames(names, input); err != nil {
				return err
			}
		}
	case input.Join != nil:
		for _, input := range input.Join {
			if err := validateNames(names, input); err != nil {
				return err
			}
		}
	case input.Group != nil:
		for _, input := range input.Group {
			if err := validateNames(names, input); err != nil {
				return err
			}
		}
	}
	return nil
}

func merge(from, to map[string]bool) {
	for s := range from {
		to[s] = true
	}
}

func validateName(name string) error {
	if name == "" {
		return errors.Errorf("input must specify a name")
	}
	switch name {
	case common.OutputPrefix, common.EnvFileName:
		return errors.Errorf("input cannot be named %v", name)
	}
	return nil
}

func now() *types.Timestamp {
	t, err := types.TimestampProto(time.Now())
	if err != nil {
		panic(err)
	}
	return t
}

func PipelineInfoFromCreatePipelineRequest(request *pps.CreatePipelineRequest, oldPipelineInfo *pps.PipelineInfo) (*pps.PipelineInfo, error) {
	if err := ValidatePipelineRequest(request); err != nil {
		return nil, err
	}
	// Reprocess overrides the salt in the request
	if request.Salt == "" || request.Reprocess {
		request.Salt = uuid.NewWithoutDashes()
	}

	pipelineInfo := &pps.PipelineInfo{
		Pipeline: request.Pipeline,
		Version:  1,
		Details: &pps.PipelineInfo_Details{
			Transform:               request.Transform,
			TFJob:                   request.TFJob,
			ParallelismSpec:         request.ParallelismSpec,
			Input:                   request.Input,
			OutputBranch:            request.OutputBranch,
			Egress:                  request.Egress,
			CreatedAt:               now(),
			ResourceRequests:        request.ResourceRequests,
			ResourceLimits:          request.ResourceLimits,
			SidecarResourceLimits:   request.SidecarResourceLimits,
			SidecarResourceRequests: request.SidecarResourceRequests,
			Description:             request.Description,
			Salt:                    request.Salt,
			Service:                 request.Service,
			Spout:                   request.Spout,
			DatumSetSpec:            request.DatumSetSpec,
			DatumTimeout:            request.DatumTimeout,
			JobTimeout:              request.JobTimeout,
			DatumTries:              request.DatumTries,
			SchedulingSpec:          request.SchedulingSpec,
			PodSpec:                 request.PodSpec,
			PodPatch:                request.PodPatch,
			S3Out:                   request.S3Out,
			Metadata:                request.Metadata,
			ReprocessSpec:           request.ReprocessSpec,
			Autoscaling:             request.Autoscaling,
			Tolerations:             request.Tolerations,
		},
	}

	if err := setPipelineDefaults(pipelineInfo); err != nil {
		return nil, err
	}
	// Validate final PipelineInfo (now that defaults have been populated)
	if err := validatePipeline(pipelineInfo); err != nil {
		return nil, err
	}

	if oldPipelineInfo != nil {
		// Modify pipelineInfo (increment Version, and *preserve Stopped* so
		// that updating a pipeline doesn't restart it)
		pipelineInfo.Version = oldPipelineInfo.Version + 1
		if oldPipelineInfo.Stopped {
			pipelineInfo.Stopped = true
		}
		if !request.Reprocess {
			pipelineInfo.Details.Salt = oldPipelineInfo.Details.Salt
		}
	}

	return pipelineInfo, nil
}

func ValidatePipelineInfo(pipelineInfo *pps.PipelineInfo) error {
	if pipelineInfo.Pipeline == nil {
		return errors.New("invalid pipeline spec: Pipeline field cannot be nil")
	}
	if pipelineInfo.Pipeline.Name == "" {
		return errors.New("invalid pipeline spec: Pipeline.Name cannot be empty")
	}
	if err := ancestry.ValidateName(pipelineInfo.Pipeline.Name); err != nil {
		return errors.Wrapf(err, "invalid pipeline name")
	}
	first := rune(pipelineInfo.Pipeline.Name[0])
	if !unicode.IsLetter(first) && !unicode.IsDigit(first) {
		return errors.Errorf("pipeline names must start with an alphanumeric character")
	}
	if len(pipelineInfo.Pipeline.Name) > 63 {
		return errors.Errorf("pipeline name is %d characters long, but must have at most 63: %q",
			len(pipelineInfo.Pipeline.Name), pipelineInfo.Pipeline.Name)
	}
	if err := validateTransform(pipelineInfo.Details.Transform); err != nil {
		return errors.Wrapf(err, "invalid transform")
	}
	if err := validateInput(pipelineInfo.Pipeline, pipelineInfo.Details.Input); err != nil {
		return err
	}
	if err := validateEgress(pipelineInfo.Pipeline.Name, pipelineInfo.Details.Egress); err != nil {
		return err
	}
	if pipelineInfo.Details.ParallelismSpec != nil {
		if pipelineInfo.Details.Service != nil && pipelineInfo.Details.ParallelismSpec.Constant != 1 {
			return errors.New("services can only be run with a constant parallelism of 1")
		}
	}
	if pipelineInfo.Details.OutputBranch == "" {
		return errors.New("pipeline needs to specify an output branch")
	}
	if pipelineInfo.Details.JobTimeout != nil {
		_, err := types.DurationFromProto(pipelineInfo.Details.JobTimeout)
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
	if pipelineInfo.Details.DatumTimeout != nil {
		_, err := types.DurationFromProto(pipelineInfo.Details.DatumTimeout)
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
	if pipelineInfo.Details.PodSpec != "" && !json.Valid([]byte(pipelineInfo.Details.PodSpec)) {
		return errors.Errorf("malformed PodSpec")
	}
	if pipelineInfo.Details.PodPatch != "" && !json.Valid([]byte(pipelineInfo.Details.PodPatch)) {
		return errors.Errorf("malformed PodPatch")
	}
	if pipelineInfo.Details.Service != nil {
		validServiceTypes := map[v1.ServiceType]bool{
			v1.ServiceTypeClusterIP:    true,
			v1.ServiceTypeLoadBalancer: true,
			v1.ServiceTypeNodePort:     true,
		}

		if !validServiceTypes[v1.ServiceType(pipelineInfo.Details.Service.Type)] {
			return errors.Errorf("the following service type %s is not allowed", pipelineInfo.Details.Service.Type)
		}
	}
	if pipelineInfo.Details.Spout != nil {
		if pipelineInfo.Details.Spout.Service == nil && pipelineInfo.Details.Input != nil {
			return errors.Errorf("spout pipelines (without a service) must not have an input")
		}
	}
	return nil
}
