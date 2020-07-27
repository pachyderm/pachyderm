package pps

import (
	"fmt"
	"regexp"

	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

// ErrJobFinished represents a finished job error.
type ErrJobFinished struct {
	Job *pps.Job
}

func (e ErrJobFinished) Error() string {
	return fmt.Sprintf("job %v has already finished", e.Job.ID)
}

var (
	jobFinishedRe = regexp.MustCompile("job [^ ]+ has already finished")
)

// IsJobFinishedErr returns true if 'err' has an error message that matches ErrJobFinished
func IsJobFinishedErr(err error) bool {
	if err == nil {
		return false
	}
	return jobFinishedRe.MatchString(grpcutil.ScrubGRPC(err).Error())
}
