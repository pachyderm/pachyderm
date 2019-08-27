package spawner

type serviceJob {
	commitInfo *pfs.CommitInfo
	cancel func()
}

// Runs the given callback with the latest commit for the pipeline.  The given
// context will be canceled if a newer commit is ready.
func forLatestCommit(
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	cb func(context.Context, *pfs.CommitInfo) error,
) error {
	ctx := pachClient.Context()

	commitIter, err := pachClient.SubscribeCommit(pipelineInfo.Pipeline.Name, pipelineInfo.OutputBranch, "", pfs.CommitState_READY)
	if err != nil {
		return err
	}

	var serviceCtx context.Context
	var serviceCancel func()

	cancel := func() {
		if serviceCancel != nil {
			serviceCancel()
		}
	}

	defer cancel()

	for {
		if commitInfo, err := commitIter.Next(); err != nil {
			return err
		} else if commitInfo.Finished == nil {
			cancel()
			serviceCtx, serviceCancel = context.WithCancel(ctx)
			go func() {
				cb(serviceCtx, serviceCancel)
			}
		}
	}
}

func RunService(pachClient *client.APIClient, pipelineInfo *pps.PipelineInfo, utils common.Utils, logger logs.TemplateLogger) error {
	ctx := pachClient.Context()

	// The serviceCtx is only used for canceling user code (due to a new output
	// commit being ready)
	return forLatestCommit(pachClient, pipelineInfo, func(serviceCtx context.Context, commitInfo *pfs.CommitInfo) error {
		// Create a job document matching the service's output commit
		jobInput := ppsutil.JobInput(pipelineInfo, commitInfo)
		job, err := pachClient.CreateJob(pipelineInfo.Pipeline.Name, commitInfo.Commit)
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

		utils.WithProvisionedNode(pachClient, data, logger, func() error {
			if err := utils.UpdateJobState(ctx, job.ID, pps.JobState_JOB_RUNNING); err != nil {
				logger.Logf("error updating job state: %+v", err)
			}
			if err := runServiceUserCode(serviceCtx, logger); err != nil {
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
