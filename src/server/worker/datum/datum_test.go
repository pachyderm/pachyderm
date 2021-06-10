package datum

// TODO: This test needs to be reworked.
//func TestSet(t *testing.T) {
//	t.Parallel()
//  env := testpachd.NewRealEnv(t, testutil.NewTestDBConfig(t))
//
//	c := env.PachClient
//	inputRepo := tu.UniqueString(t.Name() + "_input")
//	require.NoError(t, c.CreateRepo(inputRepo))
//	outputRepo := tu.UniqueString(t.Name() + "_output")
//	require.NoError(t, c.CreateRepo(outputRepo))
//	inputCommit, err := c.StartCommit(inputRepo, "master")
//	require.NoError(t, err)
//	for i := 0; i < 50; i++ {
//		require.NoError(t, c.PutFile(inputRepo, inputCommit.ID, fmt.Sprintf("/foo%v", i), strings.NewReader("input")))
//	}
//	require.NoError(t, c.FinishCommit(inputRepo, inputCommit.ID))
//	inputName := "test"
//	in := client.NewPFSInput(inputRepo, "/foo*")
//	in.Pfs.Name = inputName
//	in.Pfs.Commit = inputCommit.ID
//	outputCommit, err := c.StartCommit(outputRepo, "master")
//	require.NoError(t, err)
//	var allInputs [][]*common.Input
//	// Create datum fileset.
//	require.NoError(t, c.WithModifyFileClient(outputRepo, outputCommit.ID, func(mfc *client.ModifyFileClient) error {
//    storageRoot := t.TempDir()
//		require.NoError(t, WithSet(c, storageRoot, func(s *Set) error {
//			di, err := NewIterator(c, in)
//			if err != nil {
//				return err
//			}
//			return di.Iterate(func(meta *Meta) error {
//				allInputs = append(allInputs, meta.Inputs)
//				return s.WithDatum(meta, func(d *Datum) error {
//					return processFiles(path.Join(d.PFSStorageRoot(), OutputPrefix), path.Join(d.PFSStorageRoot(), inputName), func(_ []byte) []byte {
//						return []byte("output")
//					})
//				})
//			})
//		}, WithMetaOutput(newDatumClient(mfc))))
//		return nil
//	}))
//	require.NoError(t, c.FinishCommit(outputRepo, outputCommit.ID))
//	// Check output.
//  storageRoot := t.TempDir()
//	require.NoError(t, WithSet(c, storageRoot, func(s *Set) error {
//		fsi := NewFileSetIterator(c, outputRepo, outputCommit.ID)
//		return fsi.Iterate(func(meta *Meta) error {
//			require.Equal(t, allInputs[0], meta.Inputs)
//			allInputs = allInputs[1:]
//			return nil
//		})
//	}))
//}
//
//func processFiles(outputDir, inputDir string, cb func([]byte) []byte) error {
//	return filepath.Walk(inputDir, func(file string, fi os.FileInfo, err error) (retErr error) {
//		if err != nil {
//			return err
//		}
//		if file == inputDir {
//			return nil
//		}
//		buf := &bytes.Buffer{}
//		inputF, err := os.Open(file)
//		if err != nil {
//			return err
//		}
//		defer func() {
//			if err := inputF.Close(); retErr == nil {
//				retErr = err
//			}
//		}()
//		if _, err := io.Copy(buf, inputF); err != nil {
//			return err
//		}
//		data := cb(buf.Bytes())
//		outputName := path.Base(file)
//		outputPath := path.Join(outputDir, outputName)
//		outputF, err := os.Create(outputPath)
//		if err != nil {
//			return err
//		}
//		defer func() {
//			if err := outputF.Close(); retErr == nil {
//				retErr = err
//			}
//		}()
//		_, err = outputF.Write(data)
//		return err
//	})
//}
