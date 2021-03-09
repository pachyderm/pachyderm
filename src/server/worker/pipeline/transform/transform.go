package transform

import (
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsconsts"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Run will run a transform pipeline until the driver is canceled.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	reg, err := newRegistry(driver, logger)
	if err != nil {
		return err
	}
	logger.Logf("transform spawner started")
	return forEachCommit(driver, logger, reg.startJob)
}

func forEachCommit(driver driver.Driver, logger logs.TaggedLogger, cb func(*pps.JobInfo, *pfs.CommitInfo) error) error {
	pachClient := driver.PachClient()
	pipelineInfo := driver.PipelineInfo()
	return pachClient.SubscribeCommit(
		pipelineInfo.Pipeline.Name,
		"",
		client.NewCommitProvenance(ppsconsts.SpecRepo, pipelineInfo.Pipeline.Name, pipelineInfo.SpecCommit.ID),
		"",
		pfs.CommitState_READY,
		func(commitInfo *pfs.CommitInfo) error {
			err := func() error {
				jobInfo, err := pachClient.InspectJobOutputCommit(pipelineInfo.Pipeline.Name, commitInfo.Commit.ID, false)
				if err != nil {
					// TODO: It would be better for this to be a structured error.
					if strings.Contains(err.Error(), "not found") {
						job, err := pachClient.CreateJob(pipelineInfo.Pipeline.Name, commitInfo.Commit, ppsutil.GetStatsCommit(commitInfo))
						if err != nil {
							return err
						}
						jobInfo, err = pachClient.InspectJob(job.ID, false)
						if err != nil {
							return err
						}
						logger.Logf("created new job %q for output commit %q", jobInfo.Job.ID, jobInfo.OutputCommit.ID)
						return cb(jobInfo, commitInfo)
					}
					return err
				}
				logger.Logf("found existing job %q for output commit %q", jobInfo.Job.ID, commitInfo.Commit.ID)
				return cb(jobInfo, commitInfo)
			}()
			// TODO: Figure out how to clean up jobs after deleted commit. Just cleaning up here is not a good solution because
			// we are not guaranteed to hit this code path after a deletion.
			if pfsserver.IsCommitFinishedErr(err) || pfsserver.IsCommitNotFoundErr(err) || pfsserver.IsCommitDeletedErr(err) || ppsserver.IsJobFinishedErr(err) {
				return nil
			}
			return err
		},
	)
}
