package transform

import (
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
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
	// wrap SubscribeJob() in a retry to mitigate database connection flakiness.
	return backoff.RetryUntilCancel(driver.PachClient().Ctx(),
		func() error {
			err := driver.PachClient().SubscribeJob(
				driver.PipelineInfo().Pipeline.Project.GetName(),
				driver.PipelineInfo().Pipeline.Name,
				true,
				func(jobInfo *pps.JobInfo) error {
					if jobInfo.PipelineVersion != driver.PipelineInfo().Version {
						// Skip this job - we should be shut down soon, but don't error out in the meantime
						return nil
					}
					if jobInfo.State == pps.JobState_JOB_FINISHING {
						return nil
					}
					return reg.startJob(proto.Clone(jobInfo).(*pps.JobInfo))
				},
			)
			if errutil.IsDatabaseDisconnect(err) {
				logger.Logf("retry SubscribeJob() in transform.Run(); err: %v", err)
				return backoff.ErrContinue
			}
			return err
		}, backoff.RetryEvery(time.Second), nil)
}
