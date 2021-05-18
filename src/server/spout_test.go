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
		tu.DeleteAll(t)
		defer tu.DeleteAll(t)
		c := tu.GetAuthenticatedPachClient(t, auth.RootUser)

		dataRepo := tu.UniqueString("TestSpoutAuth_data")
		require.NoError(t, c.CreateRepo(dataRepo))

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

		// get 5 succesive commits, and ensure that the file size increases each time
		// since the spout should be appending to that file on each commit
		countBreakFunc := newCountBreakFunc(5)
		var prevLength uint64
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", nil, "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
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
		err = c.SquashCommit(pipeline, "master", "")
		require.NoError(t, err)

		// finally, let's make sure that the provenance is in a consistent state after running the spout test
		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
			if resp.Error != "" {
				return errors.New(resp.Error)
			}
			return nil
		}))
	})
	t.Run("SpoutAuthEnabledAfter", func(t *testing.T) {
		tu.DeleteAll(t)
		c := tu.GetPachClient(t)

		dataRepo := tu.UniqueString("TestSpoutAuthEnabledAfter_data")
		require.NoError(t, c.CreateRepo(dataRepo))

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

		// get 5 succesive commits
		countBreakFunc := newCountBreakFunc(5)
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", nil, "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				files, err := c.ListFileAll(ci.Commit, "")
				require.NoError(t, err)
				require.Equal(t, 1, len(files))
				return nil
			})
		}))

		// now let's authenticate, and make sure the spout fails due to a lack of authorization
		c = tu.GetAuthenticatedPachClient(t, auth.RootUser)
		defer tu.DeleteAll(t)

		// make sure we can delete commits
		err = c.SquashCommit(pipeline, "master", "")
		require.NoError(t, err)

		// now let's update the pipeline and make sure it works again
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
				Update:    true,
				Reprocess: true,         // to ensure subscribe commit will only read commits since the update
				Spout:     &pps.Spout{}, // this needs to be non-nil to make it a spout
			})
		require.NoError(t, err)

		// get 5 succesive commits
		countBreakFunc = newCountBreakFunc(5)
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", nil, "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				files, err := c.ListFileAll(ci.Commit, "")
				require.NoError(t, err)
				require.Equal(t, 1, len(files))
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
	})

	testSpout(t, true)
}

func testSpout(t *testing.T, usePachctl bool) {
	tu.DeleteAll(t)
	defer tu.DeleteAll(t)
	c := tu.GetPachClient(t)

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
		dataRepo := tu.UniqueString("TestSpoutBasic_data")
		require.NoError(t, c.CreateRepo(dataRepo))

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
		// get 5 succesive commits, and ensure that the file size increases each time
		// since the spout should be appending to that file on each commit
		countBreakFunc := newCountBreakFunc(5)
		var prevLength uint64
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", nil, "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
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
		err = c.SquashCommit(pipeline, "master", "")
		require.NoError(t, err)

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

		// we should have one job between pipeline and downstreamPipeline
		pipelineJobInfos, err := c.FlushPipelineJobAll([]*pfs.Commit{client.NewCommit(pipeline, "master", "")}, []string{downstreamPipeline})
		require.NoError(t, err)
		require.Equal(t, 1, len(pipelineJobInfos))

		// check that the spec commit for the pipeline has the correct subvenance -
		// there should be one entry for the output commit in the spout pipeline,
		// and one for the propagated commit in the downstream pipeline
		commitInfo, err := c.InspectCommit("__spec__", pipeline, "")
		require.NoError(t, err)
		require.Equal(t, 3, len(commitInfo.Subvenance))

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
		dataRepo := tu.UniqueString("TestSpoutOverwrite_data")
		require.NoError(t, c.CreateRepo(dataRepo))

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

		// if the overwrite flag is enabled, then the spout will overwrite the file on each commit
		// so the commits should have files that stay the same size
		countBreakFunc := newCountBreakFunc(5)
		var count int
		var prevLength uint64
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", nil, "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				files, err := c.ListFileAll(ci.Commit, "")
				require.NoError(t, err)
				require.Equal(t, 1, len(files))

				fileLength := files[0].SizeBytes
				if count > 0 && fileLength != prevLength {
					t.Errorf("File length was expected to stay the same. Prev: %v, Cur: %v", prevLength, fileLength)
				}
				prevLength = fileLength
				count++
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
		dataRepo := tu.UniqueString("TestSpoutProvenance_data")
		require.NoError(t, c.CreateRepo(dataRepo))

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

		// get some commits
		pipelineInfo, err := c.InspectPipeline(pipeline)
		require.NoError(t, err)
		countBreakFunc := newCountBreakFunc(3)
		// and we want to make sure that these commits all have the same provenance
		provenanceID := ""
		var count int
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "", pipelineInfo.SpecCommit.NewProvenance(), "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				require.Equal(t, 1, len(ci.Provenance))
				provenance := ci.Provenance[0].Commit
				if count == 0 {
					// set first one
					provenanceID = provenance.ID
				} else {
					require.Equal(t, provenanceID, provenance.ID)
				}
				count++
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
				Spout:     &pps.Spout{},
				Update:    true,
				Reprocess: true,
			})
		require.NoError(t, err)

		pipelineInfo, err = c.InspectPipeline(pipeline)
		require.NoError(t, err)
		countBreakFunc = newCountBreakFunc(3)
		count = 0
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "", pipelineInfo.SpecCommit.NewProvenance(), "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
			return countBreakFunc(func() error {
				require.Equal(t, 1, len(ci.Provenance))
				provenance := ci.Provenance[0].Commit
				if count == 0 {
					// this time, we expect our commits to have different provenance from the commits earlier
					require.NotEqual(t, provenanceID, provenance.ID)
					provenanceID = provenance.ID
				} else {
					// but they should still have the same provenance as each other
					require.Equal(t, provenanceID, provenance.ID)
				}
				count++
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
		dataRepo := tu.UniqueString("TestSpoutService_data")
		require.NoError(t, c.CreateRepo(dataRepo))

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
				Transform: &pps.Transform{
					Image: "pachyderm/ubuntuplusnetcat:latest",
					Cmd:   []string{"sh"},
					Stdin: []string{netcatCommand},
				},
				ParallelismSpec: &pps.ParallelismSpec{
					Constant: 1,
				},
				Input:  client.NewPFSInput(dataRepo, "/"),
				Update: false,
				Spout: &pps.Spout{
					Service: &pps.Service{
						InternalPort: 8000,
						ExternalPort: 31800,
					},
				},
			})
		require.NoError(t, err)
		time.Sleep(20 * time.Second)

		host := c.GetAddress().Host
		serviceAddr := net.JoinHostPort(host, "31800")

		// Write a tar stream with a single file to
		// the tcp connection of the pipeline service's
		// external port.
		backoff.Retry(func() error {
			raddr, err := net.ResolveTCPAddr("tcp", serviceAddr)
			if err != nil {
				return err
			}

			conn, err := net.DialTCP("tcp", nil, raddr)
			if err != nil {
				return err
			}
			tarwriter := tar.NewWriter(conn)
			defer tarwriter.Close()
			headerinfo := &tar.Header{
				Name: "file1",
				Size: int64(len("foo")),
			}

			err = tarwriter.WriteHeader(headerinfo)
			if err != nil {
				return err
			}

			_, err = tarwriter.Write([]byte("foo"))
			if err != nil {
				return err
			}
			return nil
		}, backoff.NewTestingBackOff())
		require.NoError(t, c.SubscribeCommit(client.NewRepo(pipeline), "master", nil, "", pfs.CommitState_FINISHED, func(ci *pfs.CommitInfo) error {
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
}
