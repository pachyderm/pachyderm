package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/net/context"

	"github.com/fsouza/go-dockerclient"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/spf13/cobra"
)

const badJSON1 = `
{
"356weryt

}
`

const badJSON2 = `
{
    "a": 1,
    "b": [23,4,4,64,56,36,7456,7],
    "c": {a,f,g,h,j,j},
    "d": 3452.36456,
}
`

func rootCmd() *cobra.Command {
	rootCmd := &cobra.Command{}
	AddCmds("0.0.0.0:30650", rootCmd)
	return rootCmd
}

func testJSONSyntaxErrorsReported(inputFile string, inputFileValue string, inputCommand []string) {
	ioutil.WriteFile(inputFile, []byte(inputFileValue), 0644)
	os.Args = inputCommand
	rootCmd().Execute()
}

func testBadJSON(t *testing.T, testName string, inputFile string, inputFileValue string, inputCommand []string, expectedOutput string) {

	// Setup user config prior to test
	// So that stdout only contains JSON no warnings
	_, err := config.Read()
	require.NoError(t, err)

	if os.Getenv("BE_CRASHER") == "1" {
		testJSONSyntaxErrorsReported(inputFile, inputFileValue, inputCommand)
		return
	}

	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%v", testName))
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	out, err := cmd.CombinedOutput()

	// Do our cleanup here, since we have an exit 1 in the actual run:
	wd, _ := os.Getwd()
	fileName := filepath.Join(wd, inputFile)
	os.Remove(fileName)

	require.YesError(t, err)
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		require.Equal(t, expectedOutput, string(out))
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)

}

func TestJSONSyntaxErrorsReportedCreateJob(t *testing.T) {
	descriptiveOutput := `Syntax Error on line 3:

"356weryt

         ^
invalid character '\n' in string literal
`
	cmd := []string{"pachctl", "create-job", "-f", "bad1.json"}
	testBadJSON(t, "TestJSONSyntaxErrorsReportedCreateJob", "bad1.json", badJSON1, cmd, descriptiveOutput)
}

func TestJSONSyntaxErrorsReportedCreateJob2(t *testing.T) {
	descriptiveOutput := `Syntax Error on line 5:

    "c": {a
          ^
invalid character 'a' looking for beginning of object key string
`
	cmd := []string{"pachctl", "create-job", "-f", "bad2.json"}
	testBadJSON(t, "TestJSONSyntaxErrorsReportedCreateJob2", "bad2.json", badJSON2, cmd, descriptiveOutput)
}

func TestJSONSyntaxErrorsReportedCreatePipeline(t *testing.T) {
	descriptiveOutput := `Syntax Error on line 5:

    "c": {a
          ^
invalid character 'a' looking for beginning of object key string
`
	cmd := []string{"pachctl", "create-pipeline", "-f", "bad2.json"}
	testBadJSON(t, "TestJSONSyntaxErrorsReportedCreatePipeline", "bad2.json", badJSON2, cmd, descriptiveOutput)
}

func TestJSONSyntaxErrorsReportedRunPipeline(t *testing.T) {
	descriptiveOutput := `Syntax Error on line 5:

    "c": {a
          ^
invalid character 'a' looking for beginning of object key string
`
	cmd := []string{"pachctl", "run-pipeline", "-f", "bad2.json", "somePipelineName"}
	testBadJSON(t, "TestJSONSyntaxErrorsReportedRunPipeline", "bad2.json", badJSON2, cmd, descriptiveOutput)
}

func TestJSONSyntaxErrorsReportedCreatePipelineFromStdin(t *testing.T) {
	descriptiveOutput := `Reading from stdin.
Syntax Error on line 5:

    "c": {a
          ^
invalid character 'a' looking for beginning of object key string
`
	rawCmd := []string{"pachctl", "create-pipeline"}
	testName := "TestJSONSyntaxErrorsReportedCreatePipelineFromStdin"

	if os.Getenv("BE_CRASHER") == "1" {
		os.Args = rawCmd
		ioutil.WriteFile("bad2.json", []byte(badJSON2), 0644)
		os.Stdin, _ = os.Open("bad2.json")
		rootCmd().Execute()
		return
	}

	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%v", testName))
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	out, err := cmd.CombinedOutput()

	// Put our cleanup here, since we have an exit 1 in the actual run:

	wd, _ := os.Getwd()
	os.Remove(filepath.Join(wd, "bad2.json"))

	require.YesError(t, err)
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		require.Equal(t, descriptiveOutput, string(out))
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)

}

func TestJSONSyntaxErrorsReportedUpdatePipeline(t *testing.T) {
	descriptiveOutput := `Syntax Error on line 5:

    "c": {a
          ^
invalid character 'a' looking for beginning of object key string
`
	cmd := []string{"pachctl", "update-pipeline", "-f", "bad2.json"}
	testBadJSON(t, "TestJSONSyntaxErrorsReportedCreatePipeline", "bad2.json", badJSON2, cmd, descriptiveOutput)
}

func TestPushImages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	client, err := docker.NewClientFromEnv()
	require.NoError(t, err)
	require.NoError(t, client.TagImage("pachyderm/job-shim", docker.TagImageOptions{
		Repo:    "test-job-shim",
		Context: context.Background(),
	}))
	ioutil.WriteFile("test-push-images.json", []byte(`{
  "pipeline": {
    "name": "test_push_images"
  },
  "transform": {
    "cmd": [ "true" ],
	"image": "test-job-shim"
  }
}`), 0644)
	os.Args = []string{"pachctl", "create-pipeline", "--push-images", "-f", "test-push-images.json"}
	require.NoError(t, rootCmd().Execute())
}
