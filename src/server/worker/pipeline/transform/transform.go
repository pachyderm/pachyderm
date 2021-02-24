package transform

import (
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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
	return forEachCommit(driver, func(commitInfo *pfs.CommitInfo, metaCommit *pfs.Commit) error {
		return reg.startJob(commitInfo, metaCommit)
	})
}

func getStatsCommit(commitInfo *pfs.CommitInfo) *pfs.Commit {
	for _, commitRange := range commitInfo.Subvenance {
		if commitRange.Lower.Repo.Name == commitInfo.Commit.Repo.Name {
			return commitRange.Lower
		}
	}
	return nil
}

func forEachCommit(driver driver.Driver, cb func(*pfs.CommitInfo, *pfs.Commit) error) error {
	pachClient := driver.PachClient()
	pi := driver.PipelineInfo()
	// TODO: Readd subscribe on spec commit provenance. Current code simplifies correctness in terms
	// of commits being closed / jobs being finished.
	return pachClient.SubscribeCommit(
		pi.Pipeline.Name,
		"",
		nil,
		"",
		pfs.CommitState_READY,
		func(ci *pfs.CommitInfo) error {
			return cb(ci, getStatsCommit(ci))
		},
	)
}
