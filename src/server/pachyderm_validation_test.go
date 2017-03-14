package server

// "fmt"

// "strings"

// func TestInvalidSimpleService(t *testing.T) {
// 	t.Parallel()
// 	if testing.Short() {
// 		t.Skip("Skipping integration tests in short mode")
// 	}
// 	c := getPachClient(t)
// 	dataRepo := uniqueString("TestService_data")
// 	require.NoError(t, c.CreateRepo(dataRepo))
// 	commit, err := c.StartCommit(dataRepo, "master")
// 	require.NoError(t, err)
// 	fileContent := "hai\n"
// 	_, err = c.PutFile(dataRepo, commit.ID, "file", strings.NewReader(fileContent))
// 	require.NoError(t, err)
// 	require.NoError(t, c.FinishCommit(dataRepo, commit.ID))
// 	// Only specifying an external port should error
// 	_, err = c.CreateJob(
// 		"pachyderm_netcat",
// 		[]string{"sh"},
// 		[]string{
// 			fmt.Sprintf("ls /pfs/%v/file", dataRepo),
// 			fmt.Sprintf("ls /pfs/%v/file > /pfs/out/filelist", dataRepo),
// 			fmt.Sprintf("while true; do nc -l 30003 < /pfs/%v/file; done", dataRepo),
// 		},
// 		&ppsclient.ParallelismSpec{
// 			Strategy: ppsclient.ParallelismSpec_CONSTANT,
// 			Constant: uint64(1),
// 		},
// 		[]*ppsclient.JobInput{{
// 			Commit: commit,
// 		}},
// 		"",
// 		0,
// 		30004,
// 	)
// 	require.YesError(t, err)
//
// 	// Using anything but parallelism 1 should error
// 	_, err = c.CreateJob(
// 		"pachyderm_netcat",
// 		[]string{"sh"},
// 		[]string{
// 			fmt.Sprintf("ls /pfs/%v/file", dataRepo),
// 			fmt.Sprintf("ls /pfs/%v/file > /pfs/out/filelist", dataRepo),
// 			fmt.Sprintf("while true; do nc -l 30003 < /pfs/%v/file; done", dataRepo),
// 		},
// 		&ppsclient.ParallelismSpec{
// 			Strategy: ppsclient.ParallelismSpec_CONSTANT,
// 			Constant: uint64(2),
// 		},
// 		[]*ppsclient.JobInput{{
// 			Commit: commit,
// 		}},
// 		"",
// 		0,
// 		30004,
// 	)
// 	require.YesError(t, err)
// }

// Make sure that pipeline validation requires:
// - No dash in pipeline name
// - Input must have branch and glob
//func TestInvalidCreatePipeline(t *testing.T) {
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}
//t.Parallel()
//c := getPachClient(t)

//// Set up repo
//dataRepo := uniqueString("TestDuplicatedJob_data")
//require.NoError(t, c.CreateRepo(dataRepo))

//// Create pipeline with no name
//pipelineName := uniqueString("pipeline")
//cmd := []string{"cp", path.Join("/pfs", dataRepo, "file"), "/pfs/out/file"}
//err := c.CreatePipeline(
//pipelineName,
//"",
//cmd,
//nil,
//&ppsclient.ParallelismSpec{
//Strategy: ppsclient.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*ppsclient.PipelineInput{{
//Repo:   &pfsclient.Repo{Name: dataRepo},
//Branch: "master",
//Glob:   "/*",
//}},
//false,
//)
//require.YesError(t, err)
//require.Matches(t, "name", err.Error())

//// Create pipeline named "out"
//err = c.CreatePipeline(
//pipelineName,
//"",
//cmd,
//nil,
//&ppsclient.ParallelismSpec{
//Strategy: ppsclient.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*ppsclient.PipelineInput{{
//Name:   "out",
//Repo:   &pfsclient.Repo{Name: dataRepo},
//Branch: "master",
//Glob:   "/*",
//}},
//false,
//)
//require.YesError(t, err)
//require.Matches(t, "out", err.Error())

//// Create pipeline with no branch
//err = c.CreatePipeline(
//pipelineName,
//"",
//cmd,
//nil,
//&ppsclient.ParallelismSpec{
//Strategy: ppsclient.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*ppsclient.PipelineInput{{
//Name: "input",
//Repo: &pfsclient.Repo{Name: dataRepo},
//Glob: "/*",
//}},
//false,
//)
//require.YesError(t, err)
//require.Matches(t, "branch", err.Error())

//// Create pipeline with no glob
//err = c.CreatePipeline(
//pipelineName,
//"",
//cmd,
//nil,
//&ppsclient.ParallelismSpec{
//Strategy: ppsclient.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*ppsclient.PipelineInput{{
//Name:   "input",
//Repo:   &pfsclient.Repo{Name: dataRepo},
//Branch: "master",
//}},
//false,
//)
//require.YesError(t, err)
//require.Matches(t, "glob", err.Error())
//}

//// Make sure that pipeline validation checks that all inputs exist
//func TestPipelineThatUseNonexistentInputs(t *testing.T) {
//if testing.Short() {
//t.Skip("Skipping integration tests in short mode")
//}
//t.Parallel()
//c := getPachClient(t)

//pipelineName := uniqueString("pipeline")
//require.YesError(t, c.CreatePipeline(
//pipelineName,
//"",
//[]string{"bash"},
//[]string{""},
//&ppsclient.ParallelismSpec{
//Strategy: ppsclient.ParallelismSpec_CONSTANT,
//Constant: 1,
//},
//[]*ppsclient.PipelineInput{
//{
//Name:   "whatever",
//Branch: "master",
//Glob:   "/*",
//Repo:   &pfsclient.Repo{Name: "nonexistent"},
//},
//},
//false,
//))
//}
