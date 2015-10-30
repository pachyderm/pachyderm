package jobserverrun

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/container"
	"github.com/pachyderm/pachyderm/src/pps/persist"
)

const (
	InputMountDir  = "/pfs/in"
	OutputMountDir = "/pfs/out"
)

type JobRunner interface {
	Start(*persist.JobInfo) error
}

type TestJobRunner interface {
	JobRunner
	GetJobIDToPersistJobInfo() map[string]*persist.JobInfo
}

type JobRunnerOptions struct {
	RemoveContainers bool
}

func NewJobRunner(
	pfsAPIClient pfs.APIClient,
	persistAPIClient persist.APIClient,
	containerClient container.Client,
	pfsMountDir string,
	options JobRunnerOptions,
) JobRunner {
	return newJobRunner(
		pfsAPIClient,
		persistAPIClient,
		containerClient,
		pfsMountDir,
		options,
	)
}

func NewTestJobRunner() TestJobRunner {
	return newTestJobRunner()
}
