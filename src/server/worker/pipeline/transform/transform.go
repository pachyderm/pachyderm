package transform

import (
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

// synchronously start job for commit - this will do the initial task creation
// waits until the number of running jobs
// asynchronously wait for datums and then merge
// semaphore/pool to control how many

func getStatsCommit(pipelineInfo *pps.PipelineInfo, commitInfo *pfs.CommitInfo) *pfs.Commit {
	for _, commitRange := range commitInfo.Subvenance {
		if commitRange.Lower.Repo.Name == pipelineInfo.Pipeline.Name && commitRange.Upper.Repo.Name == pipelineInfo.Pipeline.Name {
			return commitRange.Lower
		}
	}
	return nil
}

// forEachCommit listens for each READY output commit in the pipeline, and calls
// the given callback once for each such commit, synchronously.
func forEachCommit(
	driver driver.Driver,
	cb func(*pfs.CommitInfo, *pfs.Commit) error,
) error {
	pachClient := driver.PachClient()
	pi := driver.PipelineInfo()

	return pachClient.SubscribeCommitF(
		pi.Pipeline.Name,
		"",
		client.NewCommitProvenance(ppsconsts.SpecRepo, pi.Pipeline.Name, pi.SpecCommit.ID),
		"",
		pfs.CommitState_READY,
		func(ci *pfs.CommitInfo) error {
			statsCommit := getStatsCommit(pi, ci)
			// TODO: ensure ci and statsCommit are in a consistent state
			if ci.Finished == nil {
				// Inspect the commit and check again if it has been finished (it may have
				// been closed since it was queued, e.g. by StopPipeline or StopJob)
				if ci, err := pachClient.InspectCommit(ci.Commit.Repo.Name, ci.Commit.ID); err != nil {
					return err
				} else if ci.Finished == nil {
					return cb(ci, statsCommit)
				} else {
					// Make sure that the job has been correctly finished as the commit has.
					ji, err := pachClient.InspectJobOutputCommit(ci.Commit.Repo.Name, ci.Commit.ID, true)
					if err != nil {
						return err
					}
					if !ppsutil.IsTerminal(ji.State) {
						if ci.Trees == nil && ci.Tree == nil {
							ji.State = pps.JobState_JOB_KILLED
							ji.Reason = "output commit is finished without data, but job state has not been updated"
						} else {
							ji.State = pps.JobState_JOB_SUCCESS
						}

						if err := finishJob(pi, pachClient, ji, ji.State, ji.Reason, nil, nil, 0, nil, 0); err != nil {
							return errors.Wrap(err, "could not update job with finished output commit")
						}
					}
				}
			}
			return nil
		},
	)
}

// Run will run a transform pipeline until the driver is canceled.
func Run(driver driver.Driver, logger logs.TaggedLogger) error {
	reg, err := newRegistry(logger, driver)
	if err != nil {
		return err
	}

	logger.Logf("transform spawner started")

	// TODO: goroutine linearly waiting on jobs in the registry and cleaning up
	// after them, bubbling up errors, canceling

	return forEachCommit(driver, func(commitInfo *pfs.CommitInfo, statsCommit *pfs.Commit) error {
		return reg.startJob(commitInfo, statsCommit)
	})
}
