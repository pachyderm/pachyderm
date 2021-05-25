package pps

import (
	"fmt"
	"regexp"

	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// ErrPipelineJobFinished represents a finished job error.
type ErrPipelineJobFinished struct {
	PipelineJob *pps.PipelineJob
}

func (e ErrPipelineJobFinished) Error() string {
	return fmt.Sprintf("pipeline job %v has already finished", e.PipelineJob.ID)
}

var (
	pipelineJobFinishedRe = regexp.MustCompile("pipeline job [^ ]+ has already finished")
)

// IsPipelineJobFinishedErr returns true if 'err' has an error message that matches ErrPipelineJobFinished
func IsPipelineJobFinishedErr(err error) bool {
	if err == nil {
		return false
	}
	return pipelineJobFinishedRe.MatchString(grpcutil.ScrubGRPC(err).Error())
}
