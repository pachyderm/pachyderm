package pfssync

// TODO(2.0 optional): Rewrite these tests to work with the new sync package in V2.
//suite.Run("SyncPullPush", func(t *testing.T) {
//	t.Parallel()
//  env := testpachd.NewRealEnv(t, tu.NewTestDBConfig(t))
//
//	repo1 := "repo1"
//	require.NoError(t, env.PachClient.CreateRepo(repo1))
//
//	commit1, err := env.PachClient.StartCommit(repo1, "master")
//	require.NoError(t, err)
//	_, err = env.PachClient.PutFile(repo1, commit1.ID, "foo", strings.NewReader("foo\n"))
//	require.NoError(t, err)
//	_, err = env.PachClient.PutFile(repo1, commit1.ID, "dir/bar", strings.NewReader("bar\n"))
//	require.NoError(t, err)
//	require.NoError(t, env.PachClient.FinishCommit(repo1, commit1.ID))
//
//  tmpDir := t.TempDir()
//	puller := pfssync.NewPuller()
//	require.NoError(t, puller.Pull(env.PachClient, tmpDir, repo1, commit1.ID, "/", false, false, 2, nil, ""))
//	_, err = puller.CleanUp()
//	require.NoError(t, err)
//
//	repo2 := "repo2"
//	require.NoError(t, env.PachClient.CreateRepo(repo2))
//
//	commit2, err := env.PachClient.StartCommit(repo2, "master")
//	require.NoError(t, err)
//
//	require.NoError(t, pfssync.Push(env.PachClient, tmpDir, commit2, false))
//	require.NoError(t, env.PachClient.FinishCommit(repo2, commit2.ID))
//
//	var buffer bytes.Buffer
//	require.NoError(t, env.PachClient.GetFile(repo2, commit2.ID, "foo", &buffer))
//	require.Equal(t, "foo\n", buffer.String())
//	buffer.Reset()
//	require.NoError(t, env.PachClient.GetFile(repo2, commit2.ID, "dir/bar", &buffer))
//	require.Equal(t, "bar\n", buffer.String())
//
//	fileInfos, err := env.PachClient.ListFile(repo2, commit2.ID, "")
//	require.NoError(t, err)
//	require.Equal(t, 2, len(fileInfos))
//
//	commit3, err := env.PachClient.StartCommit(repo2, "master")
//	require.NoError(t, err)
//
//	// Test the overwrite flag.
//	// After this Push operation, all files should still look the same, since
//	// the old files were overwritten.
//	require.NoError(t, pfssync.Push(env.PachClient, tmpDir, commit3, true))
//	require.NoError(t, env.PachClient.FinishCommit(repo2, commit3.ID))
//
//	buffer.Reset()
//	require.NoError(t, env.PachClient.GetFile(repo2, commit3.ID, "foo", &buffer))
//	require.Equal(t, "foo\n", buffer.String())
//	buffer.Reset()
//	require.NoError(t, env.PachClient.GetFile(repo2, commit3.ID, "dir/bar", &buffer))
//	require.Equal(t, "bar\n", buffer.String())
//
//	fileInfos, err = env.PachClient.ListFile(repo2, commit3.ID, "")
//	require.NoError(t, err)
//	require.Equal(t, 2, len(fileInfos))
//
//	// Test Lazy files
//  tmpDir2 := t.TempDir()
//	puller = pfssync.NewPuller()
//	require.NoError(t, puller.Pull(env.PachClient, tmpDir2, repo1, "master", "/", true, false, 2, nil, ""))
//
//	data, err := ioutil.ReadFile(path.Join(tmpDir2, "dir/bar"))
//	require.NoError(t, err)
//	require.Equal(t, "bar\n", string(data))
//
//	_, err = puller.CleanUp()
//	require.NoError(t, err)
//})
//
//suite.Run("SyncFile", func(t *testing.T) {
//	t.Parallel()
//  env := testpachd.NewRealEnv(t, tu.NewTestDBConfig(t))
//
//	repo := "repo"
//	require.NoError(t, env.PachClient.CreateRepo(repo))
//
//	content1 := random.String(int(pfs.ChunkSize))
//
//	commit1, err := env.PachClient.StartCommit(repo, "master")
//	require.NoError(t, err)
//	require.NoError(t, pfssync.PushFile(env.PachClient, env.PachClient, &pfs.File{
//		Commit: commit1,
//		Path:   "file",
//	}, strings.NewReader(content1)))
//	require.NoError(t, env.PachClient.FinishCommit(repo, commit1.ID))
//
//	var buffer bytes.Buffer
//	require.NoError(t, env.PachClient.GetFile(repo, commit1.ID, "file", &buffer))
//	require.Equal(t, content1, buffer.String())
//
//	content2 := random.String(int(pfs.ChunkSize * 2))
//
//	commit2, err := env.PachClient.StartCommit(repo, "master")
//	require.NoError(t, err)
//	require.NoError(t, pfssync.PushFile(env.PachClient, env.PachClient, &pfs.File{
//		Commit: commit2,
//		Path:   "file",
//	}, strings.NewReader(content2)))
//	require.NoError(t, env.PachClient.FinishCommit(repo, commit2.ID))
//
//	buffer.Reset()
//	require.NoError(t, env.PachClient.GetFile(repo, commit2.ID, "file", &buffer))
//	require.Equal(t, content2, buffer.String())
//
//	content3 := content2 + random.String(int(pfs.ChunkSize))
//
//	commit3, err := env.PachClient.StartCommit(repo, "master")
//	require.NoError(t, err)
//	require.NoError(t, pfssync.PushFile(env.PachClient, env.PachClient, &pfs.File{
//		Commit: commit3,
//		Path:   "file",
//	}, strings.NewReader(content3)))
//	require.NoError(t, env.PachClient.FinishCommit(repo, commit3.ID))
//
//	buffer.Reset()
//	require.NoError(t, env.PachClient.GetFile(repo, commit3.ID, "file", &buffer))
//	require.Equal(t, content3, buffer.String())
//})
//
//suite.Run("SyncEmptyDir", func(t *testing.T) {
//	t.Parallel()
//  env := testpachd.NewRealEnv(t, tu.NewTestDBConfig(t))
//
//	repo := "repo"
//	require.NoError(t, env.PachClient.CreateRepo(repo))
//
//	commit, err := env.PachClient.StartCommit(repo, "master")
//	require.NoError(t, err)
//	require.NoError(t, env.PachClient.FinishCommit(repo, commit.ID))
//
//  tmpDir := t.TempDir()
//
//	// We want to make sure that Pull creates an empty directory
//	// when the path that we are cloning is empty.
//	dir := filepath.Join(tmpDir, "tmp")
//
//	puller := pfssync.NewPuller()
//	require.NoError(t, puller.Pull(env.PachClient, dir, repo, commit.ID, "/", false, false, 0, nil, ""))
//	_, err = os.Stat(dir)
//	require.NoError(t, err)
//	_, err = puller.CleanUp()
//	require.NoError(t, err)
//})
