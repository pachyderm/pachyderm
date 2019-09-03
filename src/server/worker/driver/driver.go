

// lookupDockerUser looks up users given the argument to a Dockerfile USER directive.
// According to Docker's docs this directive looks like:
// USER <user>[:<group>] or
// USER <UID>[:<GID>]
func lookupDockerUser(userArg string) (_ *user.User, retErr error) {
	userParts := strings.Split(userArg, ":")
	userOrUID := userParts[0]
	groupOrGID := ""
	if len(userParts) > 1 {
		groupOrGID = userParts[1]
	}
	passwd, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := passwd.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	scanner := bufio.NewScanner(passwd)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if parts[0] == userOrUID || parts[2] == userOrUID {
			result := &user.User{
				Username: parts[0],
				Uid:      parts[2],
				Gid:      parts[3],
				Name:     parts[4],
				HomeDir:  parts[5],
			}
			if groupOrGID != "" {
				if parts[0] == userOrUID {
					// groupOrGid is a group
					group, err := lookupGroup(groupOrGID)
					if err != nil {
						return nil, err
					}
					result.Gid = group.Gid
				} else {
					// groupOrGid is a gid
					result.Gid = groupOrGID
				}
			}
			return result, nil
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return nil, fmt.Errorf("user %s not found", userArg)
}

func lookupGroup(group string) (_ *user.Group, retErr error) {
	groupFile, err := os.Open("/etc/group")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := groupFile.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	scanner := bufio.NewScanner(groupFile)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if parts[0] == group {
			return &user.Group{
				Gid:  parts[2],
				Name: parts[0],
			}, nil
		}
	}
	return nil, fmt.Errorf("group %s not found", group)
}

func (a *APIServer) chunks(jobID string) col.Collection {
	return col.NewCollection(a.etcdClient, path.Join(a.etcdPrefix, chunkPrefix, jobID), nil, &ChunkState{}, nil, nil)
}

func (a *APIServer) merges(jobID string) col.Collection {
	return col.NewCollection(a.etcdClient, path.Join(a.etcdPrefix, mergePrefix, jobID), nil, &MergeState{}, nil, nil)
}

func (a *APIServer) downloadData(pachClient *client.APIClient, logger logs.TaggedLogger, inputs []*common.Input, puller *filesync.Puller, stats *pps.ProcessStats, statsTree *hashtree.Ordered) (_ string, retErr error) {
	defer a.reportDownloadTimeStats(time.Now(), stats, logger)
	logger.Logf("starting to download data")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored downloading data after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished downloading data after %v", time.Since(start))
		}
	}(time.Now())
	dir := filepath.Join(client.PPSInputPrefix, client.PPSScratchSpace, uuid.NewWithoutDashes())
	// Create output directory (currently /pfs/out)
	outPath := filepath.Join(dir, "out")
	if a.pipelineInfo.Spout != nil {
		// Spouts need to create a named pipe at /pfs/out
		if err := os.MkdirAll(filepath.Dir(outPath), 0700); err != nil {
			return "", fmt.Errorf("mkdirall :%v", err)
		}
		if err := syscall.Mkfifo(outPath, 0666); err != nil {
			return "", fmt.Errorf("mkfifo :%v", err)
		}
	} else {
		if err := os.MkdirAll(outPath, 0777); err != nil {
			return "", err
		}
	}
	for _, input := range inputs {
		if input.GitURL != "" {
			if err := a.downloadGitData(pachClient, dir, input); err != nil {
				return "", err
			}
			continue
		}
		file := input.FileInfo.File
		root := filepath.Join(dir, input.Name, file.Path)
		var statsRoot string
		if statsTree != nil {
			statsRoot = path.Join(input.Name, file.Path)
			parent, _ := path.Split(statsRoot)
			statsTree.MkdirAll(parent)
		}
		if err := puller.Pull(pachClient, root, file.Commit.Repo.Name, file.Commit.ID, file.Path, input.Lazy, input.EmptyFiles, concurrency, statsTree, statsRoot); err != nil {
			return "", err
		}
	}
	return dir, nil
}

func (a *APIServer) downloadGitData(pachClient *client.APIClient, dir string, input *common.Input) error {
	file := input.FileInfo.File
	pachydermRepoName := input.Name
	var rawJSON bytes.Buffer
	err := pachClient.GetFile(pachydermRepoName, file.Commit.ID, "commit.json", 0, 0, &rawJSON)
	if err != nil {
		return err
	}
	var payload github.PushPayload
	err = json.Unmarshal(rawJSON.Bytes(), &payload)
	if err != nil {
		return err
	}
	sha := payload.After
	// Clone checks out a reference, not a SHA
	r, err := git.PlainClone(
		filepath.Join(dir, pachydermRepoName),
		false,
		&git.CloneOptions{
			URL:           payload.Repository.CloneURL,
			SingleBranch:  true,
			ReferenceName: gitPlumbing.ReferenceName(payload.Ref),
		},
	)
	if err != nil {
		return fmt.Errorf("error cloning repo %v at SHA %v from URL %v: %v", pachydermRepoName, sha, input.GitURL, err)
	}
	hash := gitPlumbing.NewHash(sha)
	wt, err := r.Worktree()
	if err != nil {
		return err
	}
	err = wt.Checkout(&git.CheckoutOptions{Hash: hash})
	if err != nil {
		return fmt.Errorf("error checking out SHA %v from repo %v: %v", sha, pachydermRepoName, err)
	}
	return nil
}

func (a *APIServer) linkData(inputs []*common.Input, dir string) error {
	// Make sure that previously symlinked outputs are removed.
	if err := a.unlinkData(inputs); err != nil {
		return err
	}
	for _, input := range inputs {
		src := filepath.Join(dir, input.Name)
		dst := filepath.Join(client.PPSInputPrefix, input.Name)
		if err := os.Symlink(src, dst); err != nil {
			return err
		}
	}
	return os.Symlink(filepath.Join(dir, "out"), filepath.Join(client.PPSInputPrefix, "out"))
}

func (a *APIServer) unlinkData(inputs []*common.Input) error {
	dirs, err := ioutil.ReadDir(client.PPSInputPrefix)
	if err != nil {
		return fmt.Errorf("ioutil.ReadDir: %v", err)
	}
	for _, d := range dirs {
		if d.Name() == client.PPSScratchSpace {
			continue // don't delete scratch space
		}
		if err := os.RemoveAll(filepath.Join(client.PPSInputPrefix, d.Name())); err != nil {
			return err
		}
	}
	return nil
}

// Run user code and return the combined output of stdout and stderr.
func (a *APIServer) runUserCode(ctx context.Context, logger logs.TaggedLogger, environ []string, stats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {
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
	if isDone(ctx) {
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

// Run user error code and return the combined output of stdout and stderr.
func (a *APIServer) runUserErrorHandlingCode(ctx context.Context, logger logs.TaggedLogger, environ []string, stats *pps.ProcessStats, rawDatumTimeout *types.Duration) (retErr error) {
	logger.Logf("beginning to run user error handling code")
	defer func(start time.Time) {
		if retErr != nil {
			logger.Logf("errored running user error handling code after %v: %v", time.Since(start), retErr)
		} else {
			logger.Logf("finished running user error handling code after %v", time.Since(start))
		}
	}(time.Now())

	cmd := exec.CommandContext(ctx, a.pipelineInfo.Transform.ErrCmd[0], a.pipelineInfo.Transform.ErrCmd[1:]...)
	if a.pipelineInfo.Transform.ErrStdin != nil {
		cmd.Stdin = strings.NewReader(strings.Join(a.pipelineInfo.Transform.ErrStdin, "\n") + "\n")
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
	if isDone(ctx) {
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

func (a *APIServer) updateJobState(ctx context.Context, info *pps.JobInfo, statsCommit *pfs.Commit, state pps.JobState, reason string) error {
	_, err := col.NewSTM(ctx, a.etcdClient, func(stm col.STM) error {
		jobs := a.jobs.ReadWrite(stm)
		jobID := info.Job.ID
		jobPtr := &pps.EtcdJobInfo{}
		if err := jobs.Get(jobID, jobPtr); err != nil {
			return err
		}
		if jobPtr.StatsCommit == nil {
			jobPtr.StatsCommit = statsCommit
		}
		return ppsutil.UpdateJobState(a.pipelines.ReadWrite(stm), a.jobs.ReadWrite(stm), jobPtr, state, reason)
	})
	return err
}

// deleteJob is identical to updateJobState, except that jobPtr points to a job
// that should be deleted rather than marked failed. Jobs may be deleted if
// their output commit is deleted.
func (a *APIServer) deleteJob(stm col.STM, jobPtr *pps.EtcdJobInfo) error {
	pipelinePtr := &pps.EtcdPipelineInfo{}
	if err := a.pipelines.ReadWrite(stm).Update(jobPtr.Pipeline.Name, pipelinePtr, func() error {
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
	return a.jobs.ReadWrite(stm).Delete(jobPtr.Job.ID)
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

func (a *APIServer) reportUploadStats(start time.Time, stats *pps.ProcessStats, logger logs.TaggedLogger) {
	duration := time.Since(start)
	stats.UploadTime = types.DurationProto(duration)
	if a.exportStats {
		if hist, err := datumUploadTime.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get histogram w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			hist.Observe(duration.Seconds())
		}
		if counter, err := datumUploadSecondsCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(duration.Seconds())
		}
		if hist, err := datumUploadSize.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get histogram w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			hist.Observe(float64(stats.UploadBytes))
		}
		if counter, err := datumUploadBytesCount.GetMetricWithLabelValues(a.pipelineInfo.ID, a.jobID); err != nil {
			logger.Logf("failed to get counter w labels: pipeline (%v) job (%v) with error %v", a.pipelineInfo.ID, a.jobID, err)
		} else {
			counter.Add(float64(stats.UploadBytes))
		}
	}
}

func (a *APIServer) reportDownloadSizeStats(downSize float64, logger logs.TaggedLogger) {
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

func (a *APIServer) reportDownloadTimeStats(start time.Time, stats *pps.ProcessStats, logger logs.TaggedLogger) {
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

