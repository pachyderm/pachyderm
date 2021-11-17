package pps

import (
	"fmt"
	"regexp"

	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// ErrJobFinished represents a finished job error.
type ErrJobFinished struct {
	Job *pps.Job
}

func (e ErrJobFinished) Error() string {
	return fmt.Sprintf("job %v has already finished", e.Job)
}

type ErrPipelineNotFound struct {
	Pipeline *pps.Pipeline
}

func (e ErrPipelineNotFound) Error() string {
	return fmt.Sprintf("pipeline %q not found", e.Pipeline.Name)
}

var (
	jobFinishedRe      = regexp.MustCompile("job [^ ]+ has already finished")
	pipelineNotFoundRe = regexp.MustCompile("pipeline [^ ]+ not found")
)

// IsJobFinishedErr returns true if 'err' has an error message that matches ErrJobFinished
func IsJobFinishedErr(err error) bool {
	if err == nil {
		return false
	}
	return jobFinishedRe.MatchString(grpcutil.ScrubGRPC(err).Error())
}

// IsPipelineNotFoundErr returns true if 'err' has an error message that matches ErrJobFinished
func IsPipelineNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return pipelineNotFoundRe.MatchString(err.Error())
}
