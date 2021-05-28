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
	return fmt.Sprintf("pipeline job %v has already finished", e.Job.ID)
}

var (
	jobFinishedRe = regexp.MustCompile("pipeline job [^ ]+ has already finished")
)

// IsJobFinishedErr returns true if 'err' has an error message that matches ErrJobFinished
func IsJobFinishedErr(err error) bool {
	if err == nil {
		return false
	}
	return jobFinishedRe.MatchString(grpcutil.ScrubGRPC(err).Error())
}
