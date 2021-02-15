package server

// func TestSpoutPipe(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}
// 	c := tu.GetPachClient(t)

// 	testSpout(t, false) // run shared tests

// 	// pipe-specific tests
// 	t.Run("SpoutRapidOpenClose", func(t *testing.T) {
// 		dataRepo := tu.UniqueString("TestSpoutRapidOpenClose_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		// create a spout pipeline
// 		pipeline := tu.UniqueString("pipelinespoutroc")
// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Image: "spout-test:latest",
// 					Cmd:   []string{"go", "run", "./main.go"},
// 				},
// 				Spout: &pps.Spout{}, // this needs to be non-nil to make it a spout
// 			})
// 		require.NoError(t, err)

// 		// get 10 succesive commits, and ensure that the each file name we expect appears without any skips
// 		iter, err := c.SubscribeCommit(pipeline, "master", nil, "", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		for i := 0; i < 10; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, i+1, len(files))
// 			var buf bytes.Buffer
// 			err = c.GetFile(pipeline, "master", fmt.Sprintf("test%v", i), 0, 0, &buf)
// 			if err != nil {
// 				t.Errorf("Could not get file %v", err)
// 			}
// 		}

// 		// finally, let's make sure that the provenance is in a consistent state after running the spout test
// 		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
// 			if resp.Error != "" {
// 				return errors.New(resp.Error)
// 			}
// 			return nil
// 		}))
// 		require.NoError(t, c.DeleteAll())
// 	})
// 	t.Run("SpoutHammer", func(t *testing.T) {
// 		dataRepo := tu.UniqueString("TestSpoutHammer_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		// create a spout pipeline
// 		pipeline := tu.UniqueString("pipelinespoutbasic")
// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						"while [ : ]",
// 						"do",
// 						"echo \"\" | /pfs/out", // open and close pipe
// 						// no sleep so that it busy loops
// 						"date > date",
// 						"tar -cvf /pfs/out ./date*",
// 						"done"},
// 				},
// 				Spout: &pps.Spout{}, // this needs to be non-nil to make it a spout
// 			})
// 		require.NoError(t, err)

// 		// get 5 succesive commits, and ensure that the file size increases each time
// 		// since the spout should be appending to that file on each commit
// 		iter, err := c.SubscribeCommit(pipeline, "master", nil, "", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		var prevLength uint64
// 		for i := 0; i < 5; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(files))

// 			fileLength := files[0].SizeBytes
// 			if fileLength <= prevLength {
// 				t.Errorf("File length was expected to increase. Prev: %v, Cur: %v", prevLength, fileLength)
// 			}
// 			prevLength = fileLength
// 		}
// 		require.NoError(t, c.DeleteAll())
// 	})
// 	t.Run("SpoutPython", func(t *testing.T) {
// 		dataRepo := tu.UniqueString("TestSpoutPython_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		// create a spout pipeline for python
// 		pipeline := tu.UniqueString("pipelinespoutpython")
// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Image: "python:latest",
// 					Cmd:   []string{"/usr/bin/python"},
// 					Stdin: []string{`
// import io
// import random
// import string
// import tarfile
// import time
// with open("/pfs/out", "wb") as f:
//     for i in range(5):
//         with tarfile.open(fileobj=f, mode="w|", encoding="utf-8") as tar:
//             for j in range(2):
//                 content = ''.join(random.choice(string.ascii_lowercase) for _ in range(2048)).encode()
//                 tar_info = tarfile.TarInfo(str(0))
//                 tar_info.size = len(content)
//                 tar_info.mode = 0o600
//                 tar.addfile(tarinfo=tar_info, fileobj=io.BytesIO(content))
// time.Sleep(5)
// `},
// 				},
// 				Spout: &pps.Spout{},
// 			})
// 		require.NoError(t, err)

// 		// get 5 succesive commits, and ensure that the file size increases each time
// 		// since the spout should be appending to that file on each commit
// 		iter, err := c.SubscribeCommit(pipeline, "master", nil, "", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		var prevLength uint64
// 		for i := 0; i < 10; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(files))
// 			fileLength := files[0].SizeBytes
// 			if fileLength <= prevLength {
// 				t.Errorf("File length was expected to increase. Prev: %v, Cur: %v", prevLength, fileLength)
// 			}
// 			prevLength = fileLength
// 		}
// 		// make sure we can delete commits
// 		err = c.SquashCommit(pipeline, "master")
// 		require.NoError(t, err)

// 		downstreamPipeline := tu.UniqueString("pipelinespoutdownstream")
// 		require.NoError(t, c.CreatePipeline(
// 			downstreamPipeline,
// 			"",
// 			[]string{"/bin/bash"},
// 			[]string{"cp " + fmt.Sprintf("/pfs/%s/*", pipeline) + " /pfs/out/"},
// 			nil,
// 			client.NewPFSInput(pipeline, "/*"),
// 			"",
// 			false,
// 		))

// 		// we should have one job between pipeline and downstreamPipeline
// 		jobInfos, err := c.FlushJobAll([]*pfs.Commit{client.NewCommit(pipeline, "master")}, []string{downstreamPipeline})
// 		require.NoError(t, err)
// 		require.Equal(t, 1, len(jobInfos))

// 		// check that the spec commit for the pipeline has the correct subvenance -
// 		// there should be one entry for the output commit in the spout pipeline,
// 		// and one for the propagated commit in the downstream pipeline
// 		commitInfo, err := c.InspectCommit("__spec__", pipeline)
// 		require.NoError(t, err)
// 		require.Equal(t, 2, len(commitInfo.Subvenance))

// 		// finally, let's make sure that the provenance is in a consistent state after running the spout test
// 		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
// 			if resp.Error != "" {
// 				return errors.New(resp.Error)
// 			}
// 			return nil
// 		}))
// 		require.NoError(t, c.DeleteAll())
// 	})
// }

// func TestSpoutPachctl(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}

// 	// helper functions for SpoutPachctl
// 	putFileCommand := func(branch, flags, file string) string {
// 		return fmt.Sprintf("pachctl put file $PPS_PIPELINE_NAME@%s %s -f %s", branch, flags, file)
// 	}
// 	basicPutFile := func(file string) string {
// 		return putFileCommand("master", "", file)
// 	}

// 	t.Run("SpoutAuth", func(t *testing.T) {
// 		tu.DeleteAll(t)
// 		defer tu.DeleteAll(t)
// 		c := tu.GetAuthenticatedPachClient(t, tu.AdminUser)

// 		dataRepo := tu.UniqueString("TestSpoutAuth_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		// create a spout pipeline
// 		pipeline := tu.UniqueString("pipelinespoutauth")
// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						"while [ : ]",
// 						"do",
// 						"sleep 2",
// 						"date > date",
// 						basicPutFile("./date*"),
// 						"done"},
// 				},
// 				Spout: &pps.Spout{}, // this needs to be non-nil to make it a spout
// 			})
// 		require.NoError(t, err)

// 		// get 5 succesive commits, and ensure that the file size increases each time
// 		// since the spout should be appending to that file on each commit
// 		iter, err := c.SubscribeCommit(pipeline, "master", nil, "", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		var prevLength uint64
// 		for i := 0; i < 5; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(files))

// 			fileLength := files[0].SizeBytes
// 			if fileLength <= prevLength {
// 				t.Errorf("File length was expected to increase. Prev: %v, Cur: %v", prevLength, fileLength)
// 			}
// 			prevLength = fileLength
// 		}

// 		// make sure we can delete commits
// 		err = c.SquashCommit(pipeline, "master")
// 		require.NoError(t, err)

// 		// finally, let's make sure that the provenance is in a consistent state after running the spout test
// 		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
// 			if resp.Error != "" {
// 				return errors.New(resp.Error)
// 			}
// 			return nil
// 		}))
// 	})
// 	t.Run("SpoutAuthEnabledAfter", func(t *testing.T) {
// 		tu.DeleteAll(t)
// 		c := tu.GetPachClient(t)

// 		dataRepo := tu.UniqueString("TestSpoutAuthEnabledAfter_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		// create a spout pipeline
// 		pipeline := tu.UniqueString("pipelinespoutauthenabledafter")
// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						"while [ : ]",
// 						"do",
// 						"sleep 0.1",
// 						"pachctl auth whoami &> whoami",
// 						basicPutFile("./whoami*"),
// 						"done"},
// 				},
// 				Spout: &pps.Spout{}, // this needs to be non-nil to make it a spout
// 			})
// 		require.NoError(t, err)

// 		// get 5 succesive commits
// 		iter, err := c.SubscribeCommit(pipeline, "master", nil, "", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		for i := 0; i < 5; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(files))
// 		}

// 		// now let's authenticate, and make sure the spout fails due to a lack of authorization
// 		tu.ClearPachClientState(t)
// 		c = tu.GetAuthenticatedPachClient(t, tu.AdminUser)
// 		defer tu.DeleteAll(t)

// 		// make sure we can delete commits
// 		err = c.SquashCommit(pipeline, "master")
// 		require.NoError(t, err)

// 		// now let's update the pipeline and make sure it works again
// 		_, err = c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						"while [ : ]",
// 						"do",
// 						"sleep 2",
// 						"date > date",
// 						basicPutFile("./date*"),
// 						"done"},
// 				},
// 				Update:    true,
// 				Reprocess: true,         // to ensure subscribe commit will only read commits since the update
// 				Spout:     &pps.Spout{}, // this needs to be non-nil to make it a spout
// 			})
// 		require.NoError(t, err)

// 		// get 5 succesive commits
// 		iter, err = c.SubscribeCommit(pipeline, "master", nil, "", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		for i := 0; i < 5; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(files))
// 		}

// 		// finally, let's make sure that the provenance is in a consistent state after running the spout test
// 		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
// 			if resp.Error != "" {
// 				return errors.New(resp.Error)
// 			}
// 			return nil
// 		}))
// 	})

// 	testSpout(t, true)
// }

// func testSpout(t *testing.T, usePachctl bool) {
// 	tu.DeleteAll(t)
// 	defer tu.DeleteAll(t)
// 	c := tu.GetPachClient(t)

// 	putFileCommand := func(branch, flags, file string) string {
// 		if usePachctl {
// 			return fmt.Sprintf("pachctl put file $PPS_PIPELINE_NAME@%s %s -f %s", branch, flags, file)
// 		}
// 		return fmt.Sprintf("tar -cvf /pfs/out %s", file)
// 	}

// 	basicPutFile := func(file string) string {
// 		return putFileCommand("master", "", file)
// 	}

// 	t.Run("SpoutBasic", func(t *testing.T) {
// 		dataRepo := tu.UniqueString("TestSpoutBasic_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		// create a spout pipeline
// 		pipeline := tu.UniqueString("pipelinespoutbasic")
// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						"while [ : ]",
// 						"do",
// 						"sleep 2",
// 						"date > date",
// 						basicPutFile("./date*"),
// 						"done"},
// 				},
// 				Spout: &pps.Spout{}, // this needs to be non-nil to make it a spout
// 			})
// 		require.NoError(t, err)
// 		// get 5 succesive commits, and ensure that the file size increases each time
// 		// since the spout should be appending to that file on each commit
// 		iter, err := c.SubscribeCommit(pipeline, "master", nil, "", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		var prevLength uint64
// 		for i := 0; i < 5; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(files))

// 			fileLength := files[0].SizeBytes
// 			if fileLength <= prevLength {
// 				t.Errorf("File length was expected to increase. Prev: %v, Cur: %v", prevLength, fileLength)
// 			}
// 			prevLength = fileLength
// 		}
// 		// make sure we can delete commits
// 		err = c.SquashCommit(pipeline, "master")
// 		require.NoError(t, err)

// 		// and make sure we can attach a downstream pipeline
// 		downstreamPipeline := tu.UniqueString("pipelinespoutdownstream")
// 		require.NoError(t, c.CreatePipeline(
// 			downstreamPipeline,
// 			"",
// 			[]string{"/bin/bash"},
// 			[]string{"cp " + fmt.Sprintf("/pfs/%s/*", pipeline) + " /pfs/out/"},
// 			nil,
// 			client.NewPFSInput(pipeline, "/*"),
// 			"",
// 			false,
// 		))

// 		// we should have one job between pipeline and downstreamPipeline
// 		jobInfos, err := c.FlushJobAll([]*pfs.Commit{client.NewCommit(pipeline, "master")}, []string{downstreamPipeline})
// 		require.NoError(t, err)
// 		require.Equal(t, 1, len(jobInfos))

// 		// check that the spec commit for the pipeline has the correct subvenance -
// 		// there should be one entry for the output commit in the spout pipeline,
// 		// and one for the propagated commit in the downstream pipeline
// 		commitInfo, err := c.InspectCommit("__spec__", pipeline)
// 		require.NoError(t, err)
// 		require.Equal(t, 2, len(commitInfo.Subvenance))

// 		// finally, let's make sure that the provenance is in a consistent state after running the spout test
// 		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
// 			if resp.Error != "" {
// 				return errors.New(resp.Error)
// 			}
// 			return nil
// 		}))
// 		require.NoError(t, c.DeleteAll())
// 	})

// 	t.Run("SpoutOverwrite", func(t *testing.T) {
// 		dataRepo := tu.UniqueString("TestSpoutOverwrite_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		pipeline := tu.UniqueString("pipelinespoutoverwrite")
// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						// add extra command to get around issues with put file -o on a new repo
// 						"date > date",
// 						basicPutFile("./date*"),
// 						"while [ : ]",
// 						"do",
// 						"sleep 2",
// 						"date > date",
// 						putFileCommand("master", "-o", "./date*"),
// 						"done"},
// 				},
// 				Spout: &pps.Spout{
// 					Overwrite: true,
// 				},
// 			})
// 		require.NoError(t, err)

// 		// if the overwrite flag is enabled, then the spout will overwrite the file on each commit
// 		// so the commits should have files that stay the same size
// 		iter, err := c.SubscribeCommit(pipeline, "master", nil, "", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		var prevLength uint64
// 		for i := 0; i < 5; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(files))

// 			fileLength := files[0].SizeBytes
// 			if i > 0 && fileLength != prevLength {
// 				t.Errorf("File length was expected to stay the same. Prev: %v, Cur: %v", prevLength, fileLength)
// 			}
// 			prevLength = fileLength
// 		}
// 		// finally, let's make sure that the provenance is in a consistent state after running the spout test
// 		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
// 			if resp.Error != "" {
// 				return errors.New(resp.Error)
// 			}
// 			return nil
// 		}))
// 		require.NoError(t, c.DeleteAll())
// 	})

// 	t.Run("SpoutProvenance", func(t *testing.T) {
// 		dataRepo := tu.UniqueString("TestSpoutProvenance_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		// create a pipeline
// 		pipeline := tu.UniqueString("pipelinespoutprovenance")
// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						"while [ : ]",
// 						"do",
// 						"sleep 2",
// 						"date > date",
// 						basicPutFile("./date*"),
// 						"done"},
// 				},
// 				Spout: &pps.Spout{
// 					Overwrite: true,
// 				},
// 			})
// 		require.NoError(t, err)

// 		// get some commits
// 		pipelineInfo, err := c.InspectPipeline(pipeline)
// 		require.NoError(t, err)
// 		iter, err := c.SubscribeCommit(pipeline, "",
// 			client.NewCommitProvenance(ppsconsts.SpecRepo, pipeline, pipelineInfo.SpecCommit.ID),
// 			"", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)
// 		// and we want to make sure that these commits all have the same provenance
// 		provenanceID := ""
// 		for i := 0; i < 3; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(commitInfo.Provenance))
// 			provenance := commitInfo.Provenance[0].Commit
// 			if i == 0 {
// 				// set first one
// 				provenanceID = provenance.ID
// 			} else {
// 				require.Equal(t, provenanceID, provenance.ID)
// 			}
// 		}

// 		// now we'll update the pipeline
// 		_, err = c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						"while [ : ]",
// 						"do",
// 						"sleep 2",
// 						"date > date",
// 						basicPutFile("./date*"),
// 						"done"},
// 				},
// 				Spout:     &pps.Spout{},
// 				Update:    true,
// 				Reprocess: true,
// 			})
// 		require.NoError(t, err)

// 		pipelineInfo, err = c.InspectPipeline(pipeline)
// 		require.NoError(t, err)
// 		iter, err = c.SubscribeCommit(pipeline, "",
// 			client.NewCommitProvenance(ppsconsts.SpecRepo, pipeline, pipelineInfo.SpecCommit.ID),
// 			"", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		for i := 0; i < 3; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(commitInfo.Provenance))
// 			provenance := commitInfo.Provenance[0].Commit
// 			if i == 0 {
// 				// this time, we expect our commits to have different provenance from the commits earlier
// 				require.NotEqual(t, provenanceID, provenance.ID)
// 				provenanceID = provenance.ID
// 			} else {
// 				// but they should still have the same provenance as each other
// 				require.Equal(t, provenanceID, provenance.ID)
// 			}
// 		}
// 		// finally, let's make sure that the provenance is in a consistent state after running the spout test
// 		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
// 			if resp.Error != "" {
// 				return errors.New(resp.Error)
// 			}
// 			return nil
// 		}))
// 		require.NoError(t, c.DeleteAll())
// 	})
// 	t.Run("ServiceSpout", func(t *testing.T) {
// 		dataRepo := tu.UniqueString("TestServiceSpout_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		annotations := map[string]string{"foo": "bar"}

// 		// Create a pipeline that listens for tcp connections
// 		// on internal port 8000 and dumps whatever it receives
// 		// (should be in the form of a tar stream) to /pfs/out.

// 		var netcatCommand string

// 		pipeline := tu.UniqueString("pipelineservicespout")
// 		if usePachctl {
// 			netcatCommand = fmt.Sprintf("netcat -l -s 0.0.0.0 -p 8000  | tar -x --to-command 'pachctl put file %s@master:$TAR_FILENAME'", pipeline)
// 		} else {
// 			netcatCommand = "netcat -l -s 0.0.0.0 -p 8000 >/pfs/out"
// 		}

// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Metadata: &pps.Metadata{
// 					Annotations: annotations,
// 				},
// 				Transform: &pps.Transform{
// 					Image: "pachyderm/ubuntuplusnetcat:latest",
// 					Cmd:   []string{"sh"},
// 					Stdin: []string{netcatCommand},
// 				},
// 				ParallelismSpec: &pps.ParallelismSpec{
// 					Constant: 1,
// 				},
// 				Input:  client.NewPFSInput(dataRepo, "/"),
// 				Update: false,
// 				Spout: &pps.Spout{
// 					Service: &pps.Service{
// 						InternalPort: 8000,
// 						ExternalPort: 31800,
// 					},
// 				},
// 			})
// 		require.NoError(t, err)
// 		time.Sleep(20 * time.Second)

// 		host, _, err := net.SplitHostPort(c.GetAddress())
// 		require.NoError(t, err)
// 		serviceAddr := net.JoinHostPort(host, "31800")

// 		// Write a tar stream with a single file to
// 		// the tcp connection of the pipeline service's
// 		// external port.
// 		backoff.Retry(func() error {
// 			raddr, err := net.ResolveTCPAddr("tcp", serviceAddr)
// 			if err != nil {
// 				return err
// 			}

// 			conn, err := net.DialTCP("tcp", nil, raddr)
// 			if err != nil {
// 				return err
// 			}
// 			tarwriter := tar.NewWriter(conn)
// 			defer tarwriter.Close()
// 			headerinfo := &tar.Header{
// 				Name: "file1",
// 				Size: int64(len("foo")),
// 			}

// 			err = tarwriter.WriteHeader(headerinfo)
// 			if err != nil {
// 				return err
// 			}

// 			_, err = tarwriter.Write([]byte("foo"))
// 			if err != nil {
// 				return err
// 			}
// 			return nil
// 		}, backoff.NewTestingBackOff())
// 		iter, err := c.SubscribeCommit(pipeline, "master", nil, "", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		commitInfo, err := iter.Next()
// 		require.NoError(t, err)
// 		files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 		require.NoError(t, err)
// 		require.Equal(t, 1, len(files))

// 		// Confirm that a commit is made with the file
// 		// written to the external port of the pipeline's service.
// 		var buf bytes.Buffer
// 		err = c.GetFile(pipeline, commitInfo.Commit.ID, files[0].File.Path, 0, 0, &buf)
// 		require.NoError(t, err)
// 		require.Equal(t, buf.String(), "foo")

// 		// finally, let's make sure that the provenance is in a consistent state after running the spout test
// 		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
// 			if resp.Error != "" {
// 				return errors.New(resp.Error)
// 			}
// 			return nil
// 		}))
// 		require.NoError(t, c.DeleteAll())
// 	})

// 	t.Run("SpoutMarker", func(t *testing.T) {
// 		dataRepo := tu.UniqueString("TestSpoutMarker_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		// create a spout pipeline
// 		pipeline := tu.UniqueString("pipelinespoutmarker")

// 		// make sure it fails for an invalid filename
// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 				},
// 				Spout: &pps.Spout{
// 					Marker: "$$$*",
// 				},
// 			})
// 		require.YesError(t, err)

// 		var setupCommand string
// 		getMarkerCommand := "cp /pfs/mymark/test ./test"
// 		if usePachctl {
// 			setupCommand = "MARKER_HEAD=$(pachctl start commit $PPS_PIPELINE_NAME@marker)"
// 			getMarkerCommand = "pachctl get file $PPS_PIPELINE_NAME@marker:/mymark/test >test"
// 		}

// 		_, err = c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						setupCommand,
// 						getMarkerCommand,
// 						"mkdir mymark",
// 						"while [ : ]",
// 						"do",
// 						"sleep 1",
// 						"echo $(tail -1 test)x >> test",
// 						"cp test mymark/test",
// 						basicPutFile("test"),
// 						putFileCommand("$MARKER_HEAD", "", "./mymark/test*"),
// 						"done"},
// 				},
// 				Spout: &pps.Spout{
// 					Marker: "mymark",
// 				},
// 			})
// 		require.NoError(t, err)

// 		// get 5 succesive commits, and ensure that the file size increases each time
// 		// since the spout should be appending to that file on each commit
// 		followBranch := "marker"
// 		if usePachctl {
// 			// the spout never actually finishes its commit on marker, so follow master instead
// 			followBranch = "master"
// 		}
// 		iter, err := c.SubscribeCommit(pipeline, followBranch, nil, "", pfs.CommitState_FINISHED)
// 		require.NoError(t, err)

// 		var prevLength uint64
// 		for i := 0; i < 5; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(files))

// 			fileLength := files[0].SizeBytes
// 			if fileLength <= prevLength {
// 				t.Errorf("File length was expected to increase. Prev: %v, Cur: %v", prevLength, fileLength)
// 			}
// 			prevLength = fileLength
// 		}

// 		if usePachctl {
// 			require.NoError(t, c.FinishCommit(pipeline, "marker"))
// 		}
// 		_, err = c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						setupCommand,
// 						getMarkerCommand,
// 						"mkdir mymark",
// 						"while [ : ]",
// 						"do",
// 						"sleep 1",
// 						"echo $(tail -1 test). >> test",
// 						"cp test mymark/test",
// 						basicPutFile("test"),
// 						putFileCommand("$MARKER_HEAD", "", "./mymark/test*"),
// 						"done"},
// 				},
// 				Spout: &pps.Spout{
// 					Marker: "mymark",
// 				},
// 				Update: true,
// 			})
// 		require.NoError(t, err)

// 		for i := 0; i < 5; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(files))
// 		}

// 		// we want to check that the marker/test file looks like this:
// 		// x
// 		// xx
// 		// xxx
// 		// xxxx
// 		// xxxxx
// 		// xxxxx.
// 		// xxxxx..
// 		// xxxxx...
// 		// xxxxx....
// 		// xxxxx.....
// 		var buf bytes.Buffer
// 		err = c.GetFile(pipeline, "marker", "mymark/test", 0, 0, &buf)
// 		if err != nil {
// 			t.Errorf("Could not get file %v", err)
// 		}
// 		xs := 0
// 		for !errors.Is(err, io.EOF) {
// 			line := ""
// 			line, err = buf.ReadString('\n')

// 			if len(line) > 1 && line[len(line)-2:] == "x\n" {
// 				xs = len(line) - 1
// 			}
// 			if len(line) > 1 && line != strings.Repeat("x", xs)+strings.Repeat(".", len(line)-xs-1)+"\n" {
// 				t.Errorf("line did not have the expected form")
// 			}
// 		}
// 		if xs == 0 {
// 			t.Errorf("file has the wrong form, marker was likely overwritten")
// 		}
// 		// now let's reprocess the spout
// 		if usePachctl {
// 			require.NoError(t, c.FinishCommit(pipeline, "marker"))
// 		}
// 		_, err = c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						setupCommand,
// 						getMarkerCommand,
// 						"mkdir mymark",
// 						"while [ : ]",
// 						"do",
// 						"sleep 1",
// 						"echo $(tail -1 test). >> test",
// 						"cp test mymark/test",
// 						basicPutFile("test"),
// 						putFileCommand("$MARKER_HEAD", "", "./mymark/test*"),
// 						"done"},
// 				},
// 				Spout: &pps.Spout{
// 					Marker: "mymark",
// 				},
// 				Update:    true,
// 				Reprocess: true,
// 			})
// 		require.NoError(t, err)

// 		for i := 0; i < 5; i++ {
// 			commitInfo, err := iter.Next()
// 			require.NoError(t, err)
// 			files, err := c.ListFile(pipeline, commitInfo.Commit.ID, "")
// 			require.NoError(t, err)
// 			require.Equal(t, 1, len(files))
// 		}

// 		// we should get a single file with a pyramid of '.'s
// 		err = c.GetFile(pipeline, "marker", "mymark/test", 0, 0, &buf)
// 		if err != nil {
// 			t.Errorf("Could not get file %v", err)
// 		}
// 		for !errors.Is(err, io.EOF) {
// 			line := ""
// 			line, err = buf.ReadString('\n')

// 			if len(line) > 1 && line != strings.Repeat(".", len(line)-1)+"\n" {
// 				t.Errorf("line did not have the expected form %v, '%v'", len(line), line)
// 			}
// 		}
// 		// finally, let's make sure that the provenance is in a consistent state after running the spout test
// 		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
// 			if resp.Error != "" {
// 				return errors.New(resp.Error)
// 			}
// 			return nil
// 		}))
// 		require.NoError(t, c.DeleteAll())
// 	})

// 	t.Run("SpoutInputValidation", func(t *testing.T) {
// 		dataRepo := tu.UniqueString("TestSpoutInputValidation_data")
// 		require.NoError(t, c.CreateRepo(dataRepo))

// 		pipeline := tu.UniqueString("pipelinespoutinputvalidation")
// 		_, err := c.PpsAPIClient.CreatePipeline(
// 			c.Ctx(),
// 			&pps.CreatePipelineRequest{
// 				Pipeline: client.NewPipeline(pipeline),
// 				Transform: &pps.Transform{
// 					Cmd: []string{"/bin/sh"},
// 					Stdin: []string{
// 						"while [ : ]",
// 						"do",
// 						"sleep 2",
// 						"date > date",
// 						basicPutFile("./date*"),
// 						"done"},
// 				},
// 				Input: client.NewPFSInput(dataRepo, "/*"),
// 				Spout: &pps.Spout{
// 					Overwrite: true,
// 				},
// 			})
// 		require.YesError(t, err)
// 		// finally, let's make sure that the provenance is in a consistent state after running the spout test
// 		require.NoError(t, c.Fsck(false, func(resp *pfs.FsckResponse) error {
// 			if resp.Error != "" {
// 				return errors.New(resp.Error)
// 			}
// 			return nil
// 		}))
// 		require.NoError(t, c.DeleteAll())
// 	})
// }
