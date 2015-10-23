package jobserverrun // import "go.pachyderm.com/pachyderm/src/pps/jobserver/run"

import (
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps/persist"
)

type JobRunner interface {
	Start(*persist.JobInfo) error
}

type TestJobRunner interface {
	JobRunner
	GetJobIDToPersistJobInfo() map[string]*persist.JobInfo
}

func NewJobRunner(
	persistAPIClient persist.APIClient,
	containerClient container.Client,
) JobRunner {
	return newJobRunner(
		persistAPIClient,
		containerClient,
	)
}

func NewTestJobRunner() TestJobRunner {
	return newTestJobRunner()
}
