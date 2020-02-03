package pipeline

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

// RunUserCode will run the pipeline's user code until canceled by the context
// - used for services and spouts. Unlike how the transform worker runs user
// code, this does not set environment variables or collect stats.
func RunUserCode(
	ctx context.Context,
	driver driver.Driver,
	logger logs.TaggedLogger,
	dir string,
) error {
	return backoff.RetryUntilCancel(ctx, func() error {
		// TODO: shouldn't this set up env like the worker does?
		// TODO: what about the user error handling code?
		return driver.RunUserCode(logger, nil, dir, &pps.ProcessStats{}, nil)
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		logger.Logf("error in RunUserCode: %+v, retrying in: %+v", err, d)
		return nil
	})
}

// ReceiveSpout is used by both services and spouts if a spout is defined on the
// pipeline. ctx is separate from pachClient because services may call this, and
// they use a cancel function that affects the context but not the pachClient
// (so metadata updates can still be made while unwinding).
func ReceiveSpout(
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
				commit, err := pachClient.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
					Parent:     client.NewCommit(repo, ""),
					Branch:     pipelineInfo.OutputBranch,
					Provenance: []*pfs.CommitProvenance{client.NewCommitProvenance(ppsconsts.SpecRepo, repo, pipelineInfo.SpecCommit.ID)},
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
					if pipelineInfo.Spout.Marker != "" && strings.HasPrefix(path.Clean(fileHeader.Name), pipelineInfo.Spout.Marker) {
						// we'll check that this is the latest version of the spout, and then commit to it
						// we need to do this atomically because we otherwise might hit a subtle race condition

						// check to see if this spout is the latest version of this spout by seeing if its spec commit has any children
						spec, err := pachClient.InspectCommit(ppsconsts.SpecRepo, pipelineInfo.SpecCommit.ID)
						if err != nil && !errutil.IsNotFoundError(err) {
							return err
						}
						if spec != nil && len(spec.ChildCommits) != 0 {
							return fmt.Errorf("outdated spout, now shutting down")
						}
						_, err = pachClient.PutFileOverwrite(repo, ppsconsts.SpoutMarkerBranch, fileHeader.Name, outTar, 0)
						if err != nil {
							return err
						}

					} else if pipelineInfo.Spout.Overwrite {
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
