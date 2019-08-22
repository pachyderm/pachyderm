package spawner

func (s *Spawner) RunSpout(pachClient *client.APIClient) error {
	ctx := pachClient.Ctx()

	var dir string

	logger := logs.NewMasterLogger(a.pipelineInfo).WithJob("spout")
	puller := filesync.NewPuller()

	if err := a.unlinkData(nil); err != nil {
		return fmt.Errorf("unlinkData: %v", err)
	}
	// If this is our second time through the loop cleanup the old data.
	// TODO: this won't get hit
	if dir != "" {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("os.RemoveAll: %v", err)
		}
	}
	dir, err := a.downloadData(pachClient, logger, nil, puller, &pps.ProcessStats{}, nil)
	if err != nil {
		return err
	}
	if err := a.linkData(nil, dir); err != nil {
		return fmt.Errorf("linkData: %v", err)
	}

	err = a.runService(ctx, logger)
	if err != nil {
		logger.Logf("error from runService: %+v", err)
	}
	return nil
}

func (s *Spawner) receiveSpout(ctx context.Context, logger logs.TaggedLogger) error {
	return backoff.RetryNotify(func() error {
		repo := a.pipelineInfo.Pipeline.Name
		for {
			// this extra closure is so that we can scope the defer
			if err := func() (retErr error) {
				// open a read connection to the /pfs/out named pipe
				out, err := os.Open("/pfs/out")
				if err != nil {
					return err
				}
				// and close it at the end of each loop
				defer func() {
					if err := out.Close(); err != nil && retErr == nil {
						// this lets us pass the error through if Close fails
						retErr = err
					}
				}()
				outTar := tar.NewReader(out)

				// start commit
				commit, err := a.pachClient.PfsAPIClient.StartCommit(a.pachClient.Ctx(), &pfs.StartCommitRequest{
					Parent:     client.NewCommit(repo, ""),
					Branch:     a.pipelineInfo.OutputBranch,
					Provenance: []*pfs.CommitProvenance{client.NewCommitProvenance(ppsconsts.SpecRepo, a.pipelineInfo.OutputBranch, a.pipelineInfo.SpecCommit.ID)},
				})
				if err != nil {
					return err
				}

				// finish the commit even if there was an issue
				defer func() {
					if err := a.pachClient.FinishCommit(repo, commit.ID); err != nil && retErr == nil {
						// this lets us pass the error through if FinishCommit fails
						retErr = err
					}
				}()
				// this loops through all the files in the tar that we've read from /pfs/out
				for {
					fileHeader, err := outTar.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						return err
					}
					// put files into pachyderm
					if a.pipelineInfo.Spout.Overwrite {
						_, err = a.pachClient.PutFileOverwrite(repo, commit.ID, fileHeader.Name, outTar, 0)
						if err != nil {
							return err
						}
					} else {
						_, err = a.pachClient.PutFile(repo, commit.ID, fileHeader.Name, outTar)
						if err != nil {
							return err
						}
					}
				}
				return nil
			}(); err != nil {
				return err
			}
		}
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		select {
		case <-ctx.Done():
			return err
		default:
			logger.Logf("error running spout: %+v, retrying in: %+v", err, d)
			return nil
		}
	})
}
