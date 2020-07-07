package transform

import (
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

func RunV2(driver driver.Driver, logger logs.TaggedLogger) error {
	reg, err := newRegistryV2(driver, logger)
	if err != nil {
		return err
	}
	logger.Logf("transform spawner started")
	return forEachCommitV2(driver, func(commitInfo *pfs.CommitInfo, metaCommit *pfs.Commit) error {
		return reg.startJob(commitInfo, metaCommit)
	})
}

func forEachCommitV2(driver driver.Driver, cb func(*pfs.CommitInfo, *pfs.Commit) error) error {
	pachClient := driver.PachClient()
	pi := driver.PipelineInfo()
	return pachClient.SubscribeCommitF(
		pi.Pipeline.Name,
		"",
		client.NewCommitProvenance(ppsconsts.SpecRepo, pi.Pipeline.Name, pi.SpecCommit.ID),
		"",
		pfs.CommitState_READY,
		func(ci *pfs.CommitInfo) error {
			metaCommit := getStatsCommit(pi, ci)
			// TODO: ensure ci and statsCommit are in a consistent state
			if ci.Finished == nil {
				// Inspect the commit and check again if it has been finished (it may have
				// been closed since it was queued, e.g. by StopPipeline or StopJob)
				if ci, err := pachClient.InspectCommit(ci.Commit.Repo.Name, ci.Commit.ID); err != nil {
					return err
				} else if ci.Finished == nil {
					return cb(ci, metaCommit)
				}
				// TODO: recovery for inconsistent output commit / job state?
			}
			return nil
		},
	)
}
