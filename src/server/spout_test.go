//go:build k8s

package server

import (
	"archive/tar"
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func TestSpoutPachctl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// helper functions for SpoutPachctl
	putFileCommand := func(branch, flags, file string) string {
		return fmt.Sprintf("pachctl put file $PPS_PIPELINE_NAME@%s %s -f %s", branch, flags, file)
	}
	basicPutFile := func(file string) string {
		return putFileCommand("master", "-a", file)
	}

	t.Run("SpoutAuth", func(t *testing.T) {
		c, _ := minikubetestenv.AcquireCluster(t)
		c = tu.AuthenticatedPachClient(t, c, auth.RootUser)

		// create a spout pipeline
		pipeline := tu.UniqueString("pipelinespoutauth")
		_, err := c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"/bin/sh"},
					Stdin: []string{
						"while [ : ]",
						"do",
						"sleep 2",
						"date > date",
						basicPutFile("./date*"),
						"done"},
				},
				Spout: &pps.Spout{}, // this needs to be non-nil to make it a spout
			})
		require.NoError(t, err)

		// get 6 successive commits, and ensure that the file size increases each time
		// since the spout should be appending to that file on each commit
		countBreakFunc := newCountBreakFunc(6)
		var prevLength int64
		count := 0
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				count++
				if count == 1 {
					return nil // Empty head commit
				}
				files, err := c.ListFileAll(ci.Commit, "")
				require.NoError(t, err)
				require.Equal(t, 1, len(files))

				fileLength := files[0].SizeBytes
				if fileLength <= prevLength {
					t.Errorf("File length was expected to increase. Prev: %v, Cur: %v", prevLength, fileLength)
				}
				prevLength = fileLength
				return nil
			})
		}))

		// make sure we can delete commits
		commitInfo, err := c.InspectCommit(pipeline, "master", "")
		require.NoError(t, err)
		require.NoError(t, c.DropCommitSet(commitInfo.Commit.ID))

		// finally, let's make sure that the provenance is in a consistent state after running the spout test
		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
			if resp.Error != "" {
				return errors.New(resp.Error)
			}
			return nil
		}))
	})
	t.Run("SpoutAuthEnabledAfter", func(t *testing.T) {
		c, _ := minikubetestenv.AcquireCluster(t)

		// create a spout pipeline
		pipeline := tu.UniqueString("pipelinespoutauthenabledafter")
		_, err := c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"/bin/sh"},
					Stdin: []string{
						"while [ : ]",
						"do",
						"sleep 0.1",
						"pachctl auth whoami &> whoami",
						basicPutFile("./whoami*"),
						"done"},
				},
				Spout: &pps.Spout{}, // this needs to be non-nil to make it a spout
			})
		require.NoError(t, err)

		// get 6 successive commits
		countBreakFunc := newCountBreakFunc(6)
		count := 0
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				count++
				if count == 1 {
					return nil // Empty head commit
				}
				files, err := c.ListFileAll(ci.Commit, "")
				require.NoError(t, err)
				require.Equal(t, 1, len(files))
				return nil
			})
		}))

		// activate auth, which will generate a new PPS token for the pipeline and trigger
		// the RC to be recreated
		c = tu.AuthenticateClient(t, c, auth.RootUser)

		// Make sure we can delete commits. This needs to be retried because we race between
		// seeing the latest commit in commitInfo and dropping it, and we can't drop a
		// commit other than the most recent.
		require.NoErrorWithinTRetry(t, time.Minute, func() error {
			commitInfo, err := c.InspectCommit(pipeline, "master", "")
			if err != nil {
				return fmt.Errorf("inspect commit: %w", err) //nolint:wrapcheck
			}
			if err := c.DropCommitSet(commitInfo.Commit.ID); err != nil {
				return fmt.Errorf("drop commit set: %w", err) //nolint:wrapcheck
			}
			return nil
		}, "should be able to drop the latest commit")

		// get 6 successive commits
		countBreakFunc = newCountBreakFunc(6)
		count = 0
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				count++
				if count == 1 {
					return nil // Empty head commit
				}
				files, err := c.ListFileAll(ci.Commit, "")
				require.NoError(t, err)
				require.Equal(t, 1, len(files))
				return nil
			})
		}))

		// make sure that the provenance is in a consistent state after running the spout test
		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
			if resp.Error != "" {
				return errors.New(resp.Error)
			}
			return nil
		}))

		// Make sure the pipeline is still running
		pipelineInfo, err := c.InspectPipeline(pipeline, false)
		require.NoError(t, err)

		require.Equal(t, pps.PipelineState_PIPELINE_RUNNING, pipelineInfo.State)
	})

	testSpout(t, true)
}

func testSpout(t *testing.T, usePachctl bool) {
	c, _ := minikubetestenv.AcquireCluster(t)

	putFileCommand := func(branch, flags, file string) string {
		if usePachctl {
			return fmt.Sprintf("pachctl put file $PPS_PIPELINE_NAME@%s %s -f %s", branch, flags, file)
		}
		return fmt.Sprintf("tar -cvf /pfs/out %s", file)
	}

	basicPutFile := func(file string) string {
		return putFileCommand("master", "-a", file)
	}

	t.Run("SpoutBasic", func(t *testing.T) {
		// create a spout pipeline
		pipeline := tu.UniqueString("pipelinespoutbasic")
		_, err := c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"/bin/sh"},
					Stdin: []string{
						"while [ : ]",
						"do",
						"sleep 2",
						"date > date",
						basicPutFile("./date*"),
						"done"},
				},
				Spout: &pps.Spout{}, // this needs to be non-nil to make it a spout
			})
		require.NoError(t, err)
		// get 6 successive commits, and ensure that the file size increases each
		// time since the spout should be appending to that file on each commit
		countBreakFunc := newCountBreakFunc(6)
		var prevLength int64
		count := 0
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				count++
				if count == 1 {
					return nil // Empty head commit
				}
				files, err := c.ListFileAll(ci.Commit, "")
				require.NoError(t, err)
				require.Equal(t, 1, len(files))

				fileLength := files[0].SizeBytes
				if fileLength <= prevLength {
					t.Errorf("File length was expected to increase. Prev: %v, Cur: %v", prevLength, fileLength)
				}
				prevLength = fileLength
				return nil
			})
		}))

		// make sure we can delete commits
		commitInfo, err := c.InspectCommit(pipeline, "master", "")
		require.NoError(t, err)
		require.NoError(t, c.DropCommitSet(commitInfo.Commit.ID))

		// and make sure we can attach a downstream pipeline
		downstreamPipeline := tu.UniqueString("pipelinespoutdownstream")
		require.NoError(t, c.CreatePipeline(
			downstreamPipeline,
			"",
			[]string{"/bin/bash"},
			[]string{"cp " + fmt.Sprintf("/pfs/%s/*", pipeline) + " /pfs/out/"},
			nil,
			client.NewPFSInput(pipeline, "/*"),
			"",
			false,
		))

		// we should have one job on the downstream pipeline
		jobInfos, err := c.ListJob(downstreamPipeline, nil, -1, false)
		require.NoError(t, err)
		require.Equal(t, 1, len(jobInfos))

		// check that the spec commit for the pipeline has the correct subvenance -
		// there should be one entry for the output commit in the spout pipeline,
		// and one for the propagated commit in the downstream pipeline
		// TODO(global ids): check commitset instead
		// commitInfo, err := c.PfsAPIClient.InspectCommit(
		// 	c.Ctx(),
		// 	&pfs.InspectCommitRequest{
		// 		Commit: client.NewSystemRepo(pipeline, pfs.SpecRepoType).NewCommit("master", ""),
		// 		Wait:   pfs.CommitState_STARTED,
		// 	})
		// require.NoError(t, err)
		// require.Equal(t, 3, len(commitInfo.Subvenance))

		// finally, let's make sure that the provenance is in a consistent state after running the spout test
		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
			if resp.Error != "" {
				return errors.New(resp.Error)
			}
			return nil
		}))
		require.NoError(t, c.DeleteAll())
	})

	t.Run("SpoutOverwrite", func(t *testing.T) {
		pipeline := tu.UniqueString("pipelinespoutoverwrite")
		_, err := c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"/bin/sh"},
					Stdin: []string{
						// add extra command to get around issues with put file -o on a new repo
						"date > date",
						basicPutFile("./date*"),
						"while [ : ]",
						"do",
						"sleep 2",
						"date > date",
						putFileCommand("master", "", "./date*"),
						"done"},
				},
				Spout: &pps.Spout{},
			})
		require.NoError(t, err)

		// if the overwrite flag is enabled, then the spout will overwrite the file
		// on each commit so the commits should have files that stay the same size
		countBreakFunc := newCountBreakFunc(6)
		var count int
		var prevLength int64
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				count++
				if count == 1 {
					return nil // Empty head commit
				}
				files, err := c.ListFileAll(ci.Commit, "")
				require.NoError(t, err)
				require.Equal(t, 1, len(files))

				fileLength := files[0].SizeBytes
				if count > 2 && fileLength != prevLength {
					t.Errorf("File length was expected to stay the same. Prev: %v, Cur: %v", prevLength, fileLength)
				}
				prevLength = fileLength
				return nil
			})
		}))
		// finally, let's make sure that the provenance is in a consistent state after running the spout test
		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
			if resp.Error != "" {
				return errors.New(resp.Error)
			}
			return nil
		}))
		require.NoError(t, c.DeleteAll())
	})

	t.Run("SpoutProvenance", func(t *testing.T) {
		// create a pipeline
		pipeline := tu.UniqueString("pipelinespoutprovenance")
		_, err := c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"/bin/sh"},
					Stdin: []string{
						"while [ : ]",
						"do",
						"sleep 2",
						"date > date",
						basicPutFile("./date*"),
						"done"},
				},
				Spout: &pps.Spout{},
			})
		require.NoError(t, err)

		// and we want to make sure that these commits all have provenance on the spec repo
		specBranch := client.NewSystemRepo(pipeline, pfs.SpecRepoType).NewBranch("master")
		countBreakFunc := newCountBreakFunc(3)
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				require.Equal(t, 1, len(ci.DirectProvenance))
				require.Equal(t, specBranch, ci.DirectProvenance[0])
				return nil
			})
		}))

		// now we'll update the pipeline
		_, err = c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"/bin/sh"},
					Stdin: []string{
						"while [ : ]",
						"do",
						"sleep 2",
						"date > date",
						basicPutFile("./date*"),
						"done"},
				},
				Spout:  &pps.Spout{},
				Update: true,
			})
		require.NoError(t, err)

		countBreakFunc = newCountBreakFunc(6)
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				require.Equal(t, 1, len(ci.DirectProvenance))
				require.Equal(t, specBranch, ci.DirectProvenance[0])
				return nil
			})
		}))

		// finally, let's make sure that the provenance is in a consistent state after running the spout test
		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
			if resp.Error != "" {
				return errors.New(resp.Error)
			}
			return nil
		}))
		require.NoError(t, c.DeleteAll())
	})
	t.Run("SpoutService", func(t *testing.T) {
		annotations := map[string]string{"foo": "bar"}

		// Create a pipeline that listens for tcp connections
		// on internal port 8000 and dumps whatever it receives
		// (should be in the form of a tar stream) to /pfs/out.

		var netcatCommand string

		pipeline := tu.UniqueString("pipelinespoutservice")
		if usePachctl {
			netcatCommand = fmt.Sprintf("netcat -l -s 0.0.0.0 -p 8000  | tar -x --to-command 'pachctl put file -a %s@master:$TAR_FILENAME'", pipeline)
		} else {
			netcatCommand = "netcat -l -s 0.0.0.0 -p 8000 >/pfs/out"
		}

		_, err := c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Metadata: &pps.Metadata{
					Annotations: annotations,
				},
				// Note: If this image needs to be changed, change the tag to "local" and run make docker-build-netcat
				// You will need to manually update the image on dockerhub
				Transform: &pps.Transform{
					Image: "pachyderm/ubuntuplusnetcat:latest",
					Cmd:   []string{"sh"},
					Stdin: []string{netcatCommand},
				},
				ParallelismSpec: &pps.ParallelismSpec{
					Constant: 1,
				},
				Update: false,
				Spout: &pps.Spout{
					Service: &pps.Service{
						InternalPort: 8000,
						ExternalPort: 31803,
					},
				},
			})
		require.NoError(t, err)
		time.Sleep(20 * time.Second)

		host := c.GetAddress().Host
		serviceAddr := net.JoinHostPort(host, "31803")

		// Write a tar stream with a single file to
		// the tcp connection of the pipeline service's
		// external port.
		backoff.Retry(func() error { //nolint:errcheck
			raddr, err := net.ResolveTCPAddr("tcp", serviceAddr)
			if err != nil {
				return errors.EnsureStack(err)
			}

			conn, err := net.DialTCP("tcp", nil, raddr)
			if err != nil {
				return errors.EnsureStack(err)
			}
			tarwriter := tar.NewWriter(conn)
			defer func() {
				require.NoError(t, tarwriter.Close())
				require.NoError(t, conn.Close())
			}()
			headerinfo := &tar.Header{
				Name: "file1",
				Size: int64(len("foo")),
			}

			err = tarwriter.WriteHeader(headerinfo)
			if err != nil {
				return errors.EnsureStack(err)
			}

			_, err = tarwriter.Write([]byte("foo"))
			if err != nil {
				return errors.EnsureStack(err)
			}
			return nil
		}, backoff.NewTestingBackOff())

		count := 0
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			count++
			if count == 1 {
				return nil // Empty head commit
			}
			files, err := c.ListFileAll(ci.Commit, "")
			require.NoError(t, err)
			require.Equal(t, 1, len(files))
			// Confirm that a commit is made with the file
			// written to the external port of the pipeline's service.
			var buf bytes.Buffer
			err = c.GetFile(ci.Commit, files[0].File.Path, &buf)
			require.NoError(t, err)
			require.Equal(t, buf.String(), "foo")
			return errutil.ErrBreak
		}))

		// finally, let's make sure that the provenance is in a consistent state after running the spout test
		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
			if resp.Error != "" {
				return errors.New(resp.Error)
			}
			return nil
		}))
		require.NoError(t, c.DeleteAll())
	})

	t.Run("SpoutInputValidation", func(t *testing.T) {
		dataRepo := tu.UniqueString("TestSpoutInputValidation_data")
		require.NoError(t, c.CreateRepo(dataRepo))

		pipeline := tu.UniqueString("pipelinespoutinputvalidation")
		_, err := c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"/bin/sh"},
					Stdin: []string{
						"while [ : ]",
						"do",
						"sleep 2",
						"date > date",
						basicPutFile("./date*"),
						"done"},
				},
				Input: client.NewPFSInput(dataRepo, "/*"),
				Spout: &pps.Spout{},
			})
		require.YesError(t, err)
		// finally, let's make sure that the provenance is in a consistent state after running the spout test
		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
			if resp.Error != "" {
				return errors.New(resp.Error)
			}
			return nil
		}))
		require.NoError(t, c.DeleteAll())
	})

	t.Run("SpoutRestart", func(t *testing.T) {
		pipeline := tu.UniqueString("pipeline")
		_, err := c.PpsAPIClient.CreatePipeline(
			c.Ctx(),
			&pps.CreatePipelineRequest{
				Pipeline: client.NewPipeline(pipeline),
				Transform: &pps.Transform{
					Cmd: []string{"sleep", "infinity"},
				},
				Spout: &pps.Spout{},
			},
		)
		require.NoError(t, err)
		require.NoError(t, backoff.Retry(func() error {
			pi, err := c.InspectPipeline(pipeline, false)
			require.NoError(t, err)
			if pi.State != pps.PipelineState_PIPELINE_RUNNING {
				return errors.Errorf("expected pipeline state: %s, but got: %s", pps.PipelineState_PIPELINE_RUNNING, pi.State)
			}
			return nil
		}, backoff.NewTestingBackOff()))

		// stop and start spout pipeline
		require.NoError(t, c.StopPipeline(pipeline))
		require.NoError(t, backoff.Retry(func() error {
			pi, err := c.InspectPipeline(pipeline, false)
			require.NoError(t, err)
			if pi.State != pps.PipelineState_PIPELINE_PAUSED {
				return errors.Errorf("expected pipeline state: %s, but got: %s", pps.PipelineState_PIPELINE_PAUSED, pi.State)
			}
			return nil
		}, backoff.NewTestingBackOff()))
		require.NoError(t, c.StartPipeline(pipeline))
	})
}
