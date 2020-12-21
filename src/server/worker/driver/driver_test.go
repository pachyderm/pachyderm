package driver

// TODO: Make work with V2?
//var inputRepo = "inputRepo"
//var inputGitRepo = "https://github.com/pachyderm/test-artifacts.git"
//var inputGitRepoFake = "https://github.com/pachyderm/test-artifacts-fake.git"
//
//func testPipelineInfo() *pps.PipelineInfo {
//	return &pps.PipelineInfo{
//		Pipeline: client.NewPipeline("testPipeline"),
//		Transform: &pps.Transform{
//			Cmd: []string{"cp", filepath.Join("pfs", inputRepo, "file"), "pfs/out/file"},
//		},
//		ParallelismSpec: &pps.ParallelismSpec{
//			Constant: 1,
//		},
//		ResourceRequests: &pps.ResourceSpec{
//			Memory: "100M",
//			Cpu:    0.5,
//		},
//		Input: client.NewPFSInput(inputRepo, "/*"),
//	}
//}
//
//type testEnv struct {
//	testpachd.MockEnv
//	driver *driver
//}
//
//func withTestEnv(cb func(*testEnv)) error {
//	return testpachd.WithMockEnv(func(mockEnv *testpachd.MockEnv) (err error) {
//		env := &testEnv{MockEnv: *mockEnv}
//
//		var d Driver
//		d, err = NewDriver(
//			testPipelineInfo(),
//			env.PachClient,
//			env.EtcdClient,
//			tu.UniqueString("driverTest"),
//			filepath.Clean(filepath.Join(env.Directory, "pfs")),
//			"namespace",
//		)
//		if err != nil {
//			return err
//		}
//		d = d.WithContext(env.Context)
//		env.driver = d.(*driver)
//		env.driver.pipelineInfo.Transform.WorkingDir = env.Directory
//
//		cb(env)
//
//		return nil
//	})
//}
//
//// collectLogs provides the given callback with a mock TaggedLogger object which
//// will be used to collect all the logs and return them. This is pretty naive
//// and just splits log statements based on newlines because when running user
//// code, it is just used as an io.Writer and doesn't know when one message ends
//// and the next begins.
//func collectLogs(cb func(logs.TaggedLogger)) []string {
//	logger := logs.NewMockLogger()
//	buffer := &bytes.Buffer{}
//	logger.Writer = buffer
//	logger.Job = "job-id"
//
//	cb(logger)
//
//	logStmts := strings.Split(buffer.String(), "\n")
//	if len(logStmts) > 0 && logStmts[len(logStmts)-1] == "" {
//		return logStmts[0 : len(logStmts)-1]
//	}
//	return logStmts
//}
//
//// requireLogs wraps collectLogs and ensures that certain log statements were
//// made. These are specified as regular expressions in the patterns parameter,
//// and each pattern must match at least one log line. The patterns are run
//// separately against each log line, not against the entire output. If the
//// patterns parameter is nil, we require that there are no log statements.
//func requireLogs(t *testing.T, patterns []string, cb func(logs.TaggedLogger)) {
//	logStmts := collectLogs(cb)
//
//	if patterns == nil {
//		require.Equal(t, 0, len(logStmts), "callback should not have logged anything")
//	} else {
//		for _, pattern := range patterns {
//			require.OneOfMatches(t, pattern, logStmts, "callback did not log the expected message")
//		}
//	}
//}
//
//// Test that user code will successfully run and the output will be forwarded to logs
//func TestRunUserCode(t *testing.T) {
//	t.Parallel()
//	logMessage := "this is a user code log message"
//	err := withTestEnv(func(env *testEnv) {
//		env.driver.pipelineInfo.Transform.Cmd = []string{"echo", logMessage}
//		requireLogs(t, []string{logMessage}, func(logger logs.TaggedLogger) {
//			err := env.driver.RunUserCode(context.Background(), logger, []string{})
//			require.NoError(t, err)
//		})
//	})
//	require.NoError(t, err)
//}
//
//func TestRunUserCodeError(t *testing.T) {
//	t.Parallel()
//	err := withTestEnv(func(env *testEnv) {
//		env.driver.pipelineInfo.Transform.Cmd = []string{"false"}
//		requireLogs(t, []string{"exit status 1"}, func(logger logs.TaggedLogger) {
//			err := env.driver.RunUserCode(context.Background(), logger, []string{})
//			require.YesError(t, err)
//		})
//	})
//	require.NoError(t, err)
//}
//
//func TestRunUserCodeNoCommand(t *testing.T) {
//	t.Parallel()
//	err := withTestEnv(func(env *testEnv) {
//		env.driver.pipelineInfo.Transform.Cmd = []string{}
//		requireLogs(t, []string{"no command specified"}, func(logger logs.TaggedLogger) {
//			err := env.driver.RunUserCode(context.Background(), logger, []string{})
//			require.YesError(t, err)
//		})
//	})
//	require.NoError(t, err)
//}
//
//func TestRunUserCodeEnv(t *testing.T) {
//	t.Parallel()
//	err := withTestEnv(func(env *testEnv) {
//		env.driver.pipelineInfo.Transform.Cmd = []string{"env"}
//		requireLogs(t, []string{"FOO=password", "BAR=hunter2"}, func(logger logs.TaggedLogger) {
//			err := env.driver.RunUserCode(context.Background(), logger, []string{"FOO=password", "BAR=hunter2"})
//			require.NoError(t, err)
//		})
//	})
//	require.NoError(t, err)
//}
