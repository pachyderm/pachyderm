package spawner

func RunService(...) error {
	// TODO: we need to cancel the serviceCtx when the next commit is ready
	forEachCommit(pachClient, s.pipelineInfo, func(commitInfo *pfs.CommitInfo) error {
		// Create a job document matching the service's output commit
		jobInput := ppsutil.JobInput(a.pipelineInfo, commitInfo)
		job, err := pachClient.CreateJob(a.pipelineInfo.Pipeline.Name, commitInfo.Commit)
		if err != nil {
			return err
		}
		logger := logger.WithJob(job.ID)

		df, err := NewDatumFactory(pachClient, jobInput)
		if err != nil {
			return err
		}
		if df.Len() != 1 {
			return fmt.Errorf("services must have a single datum")
		}
		data := df.Datum(0)
    logger = logger.WithData(data)

		// TODO: reduce downtime between commits?
		utils.WithProvisionedNode(pachClient, data, logger, func() error {
			serviceCtx, serviceCancel = context.WithCancel(ctx)
			defer serviceCancel()

			// TODO: this used to be in a goroutine, probably to keep the service
			// running until the next commit was ready to be served
			if err := utils.UpdateJobState(ctx, job.ID, pps.JobState_JOB_RUNNING); err != nil {
				logger.Logf("error updating job state: %+v", err)
			}
			if err := a.runServiceUserCode(serviceCtx, logger); err != nil {
				logger.Logf("error from runService: %+v", err)
			}
			select {
			case <-serviceCtx.Done():
				if err := utils.UpdateJobState(ctx, job.ID, pps.JobState_JOB_SUCCESS); err != nil {
					logger.Logf("error updating job progress: %+v", err)
				}
				if err := pachClient.FinishCommit(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID); err != nil {
					logger.Logf("could not finish output commit: %v", err)
				}
			default:
			}
		})
	})
}
