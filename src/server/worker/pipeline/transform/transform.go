/*
Package transform drives standard Pachyderm pipelines, as opposed to spouts or services.

Transforms make heavy use of the [task.Service] to distribute work among available pods, both when running the actual pipeline code
and beforehand when laying out the work to be done.

# Invariants

# Job Life Cycle

The high-level life cycle of a job is as follows:
 1. A new job is noticed by the subscribe call in [transform.Run], calling registry.startJob and moves the job to [pps.JobState_JOB_STARTING]
 2. registry.processJobStarting waits for the job's input commits to be finished, and marks the job [pps.JobState_JOB_UNRUNNABLE] if any failed.
    Otherwise, the job moves to [pps.JobState_JOB_RUNNING]

Any top-level errors beyond this point will reset the state to here, as any progress may have been invalidated by another job failing.
Intermediate results are still stored in the [task.Cache], allowing efficient catch-up if nothing changed.
 3. pendingJob.load clears the meta and output commits and the job [pps.ProcessStats], and selects baseMetaCommit.
    The base meta commit is the basis for incrementality in Pachyderm, the last known good output state for this pipeline which can be used to skip work.
 4. Background goroutines are spawned to kill the job if it times out, or if the output commit is deleted.

Most of the remaining steps take place in registry.processJobRunning inside the registry.processJob loop
 5. The job's full list of datums is computed and written to a file set in the standard meta commit format (see [datum.NewFileSetIterator]).
    For now, the result is only used to determine the total number of datums, but it will be picked up from the cache multiple times in the next steps.
 6. The list of parallel datums is computed and written to a file set.
    Parallel datums are datums which are present in this job, but not in the job corresponding to baseMetaCommit.
    These typically correspond to new files having been added to the input, and can be determined even if the base job/commit isn't done yet.
 7. registry.processDatums shards the file set of parallel datums into individual [transform.DatumSet] tasks using the job's [pps.DatumSetSpec].
    Then it sends them out through the task service to be completed by workers, accumulating stats and adding file sets to the output and meta commits as results come in.
    After all datums are processed, if any datum failed, the job moves to [pps.JobState_JOB_FAILURE] and the rest of the steps are skipped.
 8. We wait for the base job to complete.
 9. The lists of serial and deleted datums are computed.
    Serial datums are those present in both the current job and the base job, but that need to be recomputed either because of the [pps.PipelineInfo_Details.ReprocessSpec] or because the underlying file contents changed.
    Unchanged datums present in both jobs can just be skipped.
    Deleted datums are those present in the base job but not the current job.
    Because Pachyderm's file system is diff based, the information about both serial and deleted datums must be explicitly deleted from the output.
    The meta commit stores the output files associated with every datum, so we iterate through and remove them with a [datum.Deleter].
 10. The serial datums are sent through registry.processDatums, just like the parallel datums.
 11. If the [pps.PipelineInfo_Details] has [pps.Egress], the job moves to [pps.JobState_JOB_EGRESSING] and egress takes place.
 12. Otherwise, or if egress succeeds, the job moves to [pps.JobState_JOB_FINISHING] and the output and meta commit move to [pfs.CommitState_FINISHING].
    This is not a terminal state, but just signals that the results are ready for compaction.

The job is now done from the perspective of the transform code, but there's still more to do.
 13. The PFS master notices the commit state changes and starts compacting and finishing the commits.
    It's still possible for the job to fail due to a file system validation error.
 14. If validation of the output commit succeeds, the job moves to [pps.JobState_JOB_SUCCESS]
*/
package transform

import (
	"github.com/gogo/protobuf/proto"

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

	return driver.PachClient().SubscribeJob(
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
}
