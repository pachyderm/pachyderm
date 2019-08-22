package spawner

func (s *Spawner) RunService(pachClient *client.APIClient) error {
	ctx := pachClient.Ctx()
	commitIter, err := pachClient.SubscribeCommit(a.pipelineInfo.Pipeline.Name, a.pipelineInfo.OutputBranch, "", pfs.CommitState_READY)
	if err != nil {
		return err
	}
	defer commitIter.Close()
	var serviceCtx context.Context
	var serviceCancel func()
	var dir string
	for {
		commitInfo, err := commitIter.Next()
		if err != nil {
			return err
		}
		if commitInfo.Finished != nil {
			continue
		}

		// Create a job document matching the service's output commit
		jobInput := ppsutil.JobInput(a.pipelineInfo, commitInfo)
		job, err := pachClient.CreateJob(a.pipelineInfo.Pipeline.Name, commitInfo.Commit)
		if err != nil {
			return err
		}
		df, err := NewDatumFactory(pachClient, jobInput)
		if err != nil {
			return err
		}
		if df.Len() != 1 {
			return fmt.Errorf("services must have a single datum")
		}
		data := df.Datum(0)
		logger := logs.NewMasterLogger(a.pipelineInfo).WithJob(job.ID).WithData(data)
		puller := filesync.NewPuller()
		// If this is our second time through the loop cleanup the old data.
		if dir != "" {
			if err := a.unlinkData(data); err != nil {
				return fmt.Errorf("unlinkData: %v", err)
			}
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("os.RemoveAll: %v", err)
			}
		}
		dir, err = a.downloadData(pachClient, logger, data, puller, &pps.ProcessStats{}, nil)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(client.PPSInputPrefix, 0666); err != nil {
			return err
		}
		if err := a.linkData(data, dir); err != nil {
			return fmt.Errorf("linkData: %v", err)
		}
		if serviceCancel != nil {
			serviceCancel()
		}
		serviceCtx, serviceCancel = context.WithCancel(ctx)
		defer serviceCancel() // make go vet happy: infinite loop obviates 'defer'
		go func() {
			serviceCtx := serviceCtx
			if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
				jobs := a.jobs.ReadWrite(stm)
				jobPtr := &pps.EtcdJobInfo{}
				if err := jobs.Get(job.ID, jobPtr); err != nil {
					return err
				}
				return ppsutil.UpdateJobState(a.pipelines.ReadWrite(stm), a.jobs.ReadWrite(stm), jobPtr, pps.JobState_JOB_RUNNING, "")
			}); err != nil {
				logger.Logf("error updating job state: %+v", err)
			}
			err := a.runService(serviceCtx, logger)
			if err != nil {
				logger.Logf("error from runService: %+v", err)
			}
			select {
			case <-serviceCtx.Done():
				if _, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
					jobs := a.jobs.ReadWrite(stm)
					jobPtr := &pps.EtcdJobInfo{}
					if err := jobs.Get(job.ID, jobPtr); err != nil {
						return err
					}
					return ppsutil.UpdateJobState(a.pipelines.ReadWrite(stm), a.jobs.ReadWrite(stm), jobPtr, pps.JobState_JOB_SUCCESS, "")
				}); err != nil {
					logger.Logf("error updating job progress: %+v", err)
				}
				if err := pachClient.FinishCommit(commitInfo.Commit.Repo.Name, commitInfo.Commit.ID); err != nil {
					logger.Logf("could not finish output commit: %v", err)
				}
			default:
			}
		}()
	}
}
