package jobserverrun

import (
	"github.com/pachyderm/pachyderm/src/pkg/container"
	"github.com/pachyderm/pachyderm/src/pps/persist"
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
