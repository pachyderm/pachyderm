package spawner

func runSpout(
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	logger logs.TaggedLogger,
	utils common.Utils,
) error {
	logger = logger.WithJob("spout")

	return utils.WithProvisionedNode(pachClient, nil, logger, func() error {
		eg, serviceCtx := errgroup.Group{}.WithContext(pachClient.Context())
		eg.Go(runUserCode(serviceCtx, logger))
		eg.Go(receiveSpout(serviceCtx, pachClient, pipelineInfo, logger))
		return eg.Wait()
	})
}

// ctx is separate from pachClient because services may call this, and they use
// a cancel function that affects the context but not the pachClient (so metadata
// updates can still be made while unwinding).
func receiveSpout(
	ctx context.Context,
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	logger logs.TaggedLogger,
) error {
	return runUntil(ctx, "receiveSpout", func() error {
		repo := pipelineInfo.Pipeline.Name
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
				commit, err := pachClient.PfsAPIClient.StartCommit(pachClient.Ctx(), &pfs.StartCommitRequest{
					Parent:     client.NewCommit(repo, ""),
					Branch:     pipelineInfo.OutputBranch,
					Provenance: []*pfs.CommitProvenance{client.NewCommitProvenance(ppsconsts.SpecRepo, pipelineInfo.OutputBranch, pipelineInfo.SpecCommit.ID)},
				})
				if err != nil {
					return err
				}

				// finish the commit even if there was an issue
				defer func() {
					if err := pachClient.FinishCommit(repo, commit.ID); err != nil && retErr == nil {
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
					if pipelineInfo.Spout.Overwrite {
						_, err = pachClient.PutFileOverwrite(repo, commit.ID, fileHeader.Name, outTar, 0)
						if err != nil {
							return err
						}
					} else {
						_, err = pachClient.PutFile(repo, commit.ID, fileHeader.Name, outTar)
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
	})
}
