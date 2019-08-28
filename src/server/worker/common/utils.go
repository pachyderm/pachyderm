package common

const (
	jobTerminalError = fmt.Errf("job is already in a terminal state, its state cannot be changed")
)

type Utils interface {
	Jobs
	Pipelines
	Chunks(jobID string) col.Collection
	Merges(jobID string) col.Collection

	WithProvisionedNode(...) error
	RunUserCode(...) error
	RunUserErrorHandlingCode(...) error

	DeleteJob
	UpdateJobState
	IsJobTerminalError(error) bool
}

type utils struct {
	etcdClient
	etcdPrefix
	jobs
	pipelines
}

func NewUtils(...) Utils {
	return &utils{
	}
}

func (u *utils) Chunks(jobID string) col.Collection {
	return col.NewCollection(u.etcdClient, path.Join(u.etcdPrefix, chunkPrefix, jobID), nil, &ChunkState{}, nil, nil)
}

func (u *utils) Merges(jobID string) col.Collection {
	return col.NewCollection(u.etcdClient, path.Join(u.etcdPrefix, mergePrefix, jobID), nil, &MergeState{}, nil, nil)
}

// Create puller *
// downloadData *
// defer removeAll *
// defer puller.Cleanup
// lock runMutex
// lock status mutex, set some dumb values
// mkdir
// linkData *
// defer unlinkData
// walk pfs and chown
// run user code and user error handling code
// puller.Cleanup
// report stats
// uploadOutput

func (u *utils) WithProvisionedNode(
	ctx context.Context,
	logger logs.TemplateLogger,
	cb func(*pps.ProcessStats) error) (stats *pps.ProcessStats, retErr error) {
	puller := filesync.NewPuller()
	stats := &pps.ProcessStats{}

	// Download input data
	var err error
	dir, err = a.downloadData(pachClient, logger, data, puller, subStats, inputTree)
	// We run these cleanup functions no matter what, so that if
	// downloadData partially succeeded, we still clean up the resources.
	defer func() {
		if err := os.RemoveAll(dir); err != nil && retErr == nil {
			retErr = err
		}
	}()
	// It's important that we run puller.CleanUp before os.RemoveAll,
	// because otherwise puller.Cleanup might try tp open pipes that have
	// been deleted.
	defer func() {
		if _, err := puller.CleanUp(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if err != nil {
		return fmt.Errorf("error downloadData: %v", err)
	}
	if err := os.MkdirAll(client.PPSInputPrefix, 0777); err != nil {
		return err
	}
	if err := a.linkData(data, dir); err != nil {
		return fmt.Errorf("error linkData: %v", err)
	}
	defer func() {
		if err := a.unlinkData(data); err != nil && retErr == nil {
			retErr = fmt.Errorf("error unlinkData: %v", err)
		}
	}()
	// If the pipeline spec set a custom user to execute the
	// process, make sure `/pfs` and its content are owned by it
	if a.uid != nil && a.gid != nil {
		filepath.Walk("/pfs", func(name string, info os.FileInfo, err error) error {
			if err == nil {
				err = os.Chown(name, int(*a.uid), int(*a.gid))
			}
			return err
		})
	}

	if err := cb(stats); err != nil {
		return err
	}

	// CleanUp is idempotent so we can call it however many times we want.
	// The reason we are calling it here is that the puller could've
	// encountered an error as it was lazily loading files, in which case
	// the output might be invalid since as far as the user's code is
	// concerned, they might've just seen an empty or partially completed
	// file.
	downSize, err := puller.CleanUp()
	if err != nil {
		logger.Logf("puller encountered an error while cleaning up: %+v", err)
		return err
	}

	atomic.AddUint64(&subStats.DownloadBytes, uint64(downSize))
	a.reportDownloadSizeStats(float64(downSize), logger)
}

func (a *APIServer) reportUserCodeStats(logger logs.TaggedLogger) {
	if a.exportStats {
		if counter, err := datumCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID, "started"); err != nil {
			logger.Logf("failed to get histogram w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(1)
		}
	}
}

func (a *APIServer) reportDeferredUserCodeStats(err error, start time.Time, stats *pps.ProcessStats, logger logs.TaggedLogger) {
	duration := time.Since(start)
	stats.ProcessTime = types.DurationProto(duration)
	if a.exportStats {
		state := "errored"
		if err == nil {
			state = "finished"
		}
		if counter, err := datumCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID, state); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) state (%v) with error %v", a.pipelineInfo.ID, a.jobID, state, err)
		} else {
			counter.Add(1)
		}
		if hist, err := datumProcTime.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID, state); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) state (%v) with error %v", a.pipelineInfo.ID, a.jobID, state, err)
		} else {
			hist.Observe(duration.Seconds())
		}
		if counter, err := datumProcSecondsCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(duration.Seconds())
		}
	}
}

// Run user code and return the combined output of stdout and stderr.
func (a *APIServer) RunUserCode(ctx context.Context, logger logs.TaggedLogger, environ []string, stats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {
	a.reportUserCodeStats(logger)
	defer func(start time.Time) { a.reportDeferredUserCodeStats(retErr, start, stats, logger) }(time.Now())
	logger.Logf("beginning to run user code")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored running user code after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished running user code after %v", time.Since(start))
		}
	}(time.Now())
	if rawDatumTimeout != nil {
		datumTimeout, err := types.DurationFromProto(rawDatumTimeout)
		if err != nil {
			return err
		}
		datumTimeoutCtx, cancel := context.WithTimeout(ctx, datumTimeout)
		defer cancel()
		ctx = datumTimeoutCtx
	}

	// Run user code
	cmd := exec.CommandContext(ctx, a.pipelineInfo.Transform.Cmd[0], a.pipelineInfo.Transform.Cmd[1:]...)
	if a.pipelineInfo.Transform.Stdin != nil {
		cmd.Stdin = strings.NewReader(strings.Join(a.pipelineInfo.Transform.Stdin, "\n") + "\n")
	}
	cmd.Stdout = logger.WithUserCode()
	cmd.Stderr = logger.WithUserCode()
	cmd.Env = environ
	if a.uid != nil && a.gid != nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: *a.uid,
				Gid: *a.gid,
			},
		}
	}
	cmd.Dir = a.pipelineInfo.Transform.WorkingDir
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error cmd.Start: %v", err)
	}
	// A context w a deadline will successfully cancel/kill
	// the running process (minus zombies)
	state, err := cmd.Process.Wait()
	if err != nil {
		return fmt.Errorf("error cmd.Wait: %v", err)
	}
	if common.isDone(ctx) {
		if err = ctx.Err(); err != nil {
			return err
		}
	}

	// Because of this issue: https://github.com/golang/go/issues/18874
	// We forked os/exec so that we can call just the part of cmd.Wait() that
	// happens after blocking on the process. Unfortunately calling
	// cmd.Process.Wait() then cmd.Wait() will produce an error. So instead we
	// close the IO using this helper
	err = cmd.WaitIO(state, err)
	// We ignore broken pipe errors, these occur very occasionally if a user
	// specifies Stdin but their process doesn't actually read everything from
	// Stdin. This is a fairly common thing to do, bash by default ignores
	// broken pipe errors.
	if err != nil && !strings.Contains(err.Error(), "broken pipe") {
		// (if err is an acceptable return code, don't return err)
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				for _, returnCode := range a.pipelineInfo.Transform.AcceptReturnCode {
					if int(returnCode) == status.ExitStatus() {
						return nil
					}
				}
			}
		}
		return fmt.Errorf("error cmd.WaitIO: %v", err)
	}
	return nil
}

func (u *utils) UpdateEtcd(ctx context.Context, cbs func(col.STM) error...) error {
	_, err := col.NewSTM(ctx, u.etcdClient, func(stm col.STM) error {
		for cb := range cbs {
			if err := cb(stm); err != nil {
				return err
			}
		}
	})
	return err
}

func (u *utils) UpdateJob(jobID string, cb func(*pps.EtcdJobInfo) error) func(col.STM) error {
	return func(stm col.STM) error {
		jobs := u.jobs.ReadWrite(stm)
		jobPtr := &pps.EtcdJobInfo{}
		if err := jobs.Get(jobID, jobPtr); err != nil {
			return err
		}
		// TODO: move this out
		if jobPtr.StatsCommit == nil {
			jobPtr.StatsCommit = statsCommit
		}
		if ppsutil.IsTerminal(jobPtr.State) {
			return jobTerminalError
		}

		return ppsutil.UpdateJobState(u.pipelines.ReadWrite(stm), u.jobs.ReadWrite(stm), jobPtr, state, reason)
	}
}

func (u *utils) UpdateJobState(ctx context.Context, jobID string, statsCommit *pfs.Commit, state pps.JobState, reason string) error {
	_, err := col.NewSTM(ctx, u.etcdClient, func(stm col.STM) error {
		jobs := u.jobs.ReadWrite(stm)
		jobPtr := &pps.EtcdJobInfo{}
		if err := jobs.Get(jobID, jobPtr); err != nil {
			return err
		}
		// TODO: move this out
		if jobPtr.StatsCommit == nil {
			jobPtr.StatsCommit = statsCommit
		}
		if ppsutil.IsTerminal(jobPtr.State) {
			return jobTerminalError
		}

		return ppsutil.UpdateJobState(u.pipelines.ReadWrite(stm), u.jobs.ReadWrite(stm), jobPtr, state, reason)
	})
	return err
}

func (u *utils) IsJobTerminalError(err error) bool {
	return err == jobTerminalError
}

// DeleteJob is identical to updateJobState, except that jobPtr points to a job
// that should be deleted rather than marked failed. Jobs may be deleted if
// their output commit is deleted.
func (u *utils) DeleteJob(stm col.STM, jobPtr *pps.EtcdJobInfo) error {
	_, err := col.NewSTM(ctx, u.etcdClient, func(stm col.STM) error {
		pipelinePtr := &pps.EtcdPipelineInfo{}
		if err := u.pipelines.ReadWrite(stm).Update(jobPtr.Pipeline.Name, pipelinePtr, func() error {
			if pipelinePtr.JobCounts == nil {
				pipelinePtr.JobCounts = make(map[int32]int32)
			}
			if pipelinePtr.JobCounts[int32(jobPtr.State)] != 0 {
				pipelinePtr.JobCounts[int32(jobPtr.State)]--
			}
			return nil
		}); err != nil {
			return err
		}
		return u.jobs.ReadWrite(stm).Delete(jobPtr.Job.ID)
	})
	return err
}

func (u *utils) reportDownloadSizeStats(downSize float64, logger logs.TaggedLogger) {
	if a.exportStats {
		if hist, err := datumDownloadSize.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get histogram w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			hist.Observe(downSize)
		}
		if counter, err := datumDownloadBytesCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(downSize)
		}
	}
}

func (u *utils) reportDownloadTimeStats(start time.Time, stats *pps.ProcessStats, logger logs.TaggedLogger) {
	duration := time.Since(start)
	stats.DownloadTime = types.DurationProto(duration)
	if a.exportStats {
		if hist, err := datumDownloadTime.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get histogram w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			hist.Observe(duration.Seconds())
		}
		if counter, err := datumDownloadSecondsCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(duration.Seconds())
		}
	}
}
