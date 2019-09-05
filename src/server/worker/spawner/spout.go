package spawner

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

func runSpout(
	pachClient *client.APIClient,
	pipelineInfo *pps.PipelineInfo,
	logger logs.TaggedLogger,
	driver driver.Driver,
) error {
	logger = logger.WithJob("spout")

	// TODO: do something with stats?
	_, err := driver.WithData(pachClient.Ctx(), nil, nil, logger, func(*pps.ProcessStats) error {
		eg, serviceCtx := errgroup.WithContext(pachClient.Ctx())
		eg.Go(func() error { return runUserCode(serviceCtx, driver, logger) })
		eg.Go(func() error { return receiveSpout(serviceCtx, pachClient, pipelineInfo, logger) })
		return eg.Wait()
	})
	return err
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
	return backoff.RetryUntilCancel(ctx, func() error {
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
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("error in receiveSpout: %+v, retrying in: %+v", err, d)
		return nil
	})
}
