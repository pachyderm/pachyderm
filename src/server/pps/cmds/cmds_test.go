package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

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

func testJSONSyntaxErrorsReported(inputFile string, inputFileValue string, inputCommand []string) {
	address := "0.0.0.0:30650"

	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Long: `Access the Pachyderm API.

Envronment variables:
  ADDRESS=0.0.0.0:30650, the server to connect to.
`,
	}

	cmds, _ := Cmds(address)
	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}

	// Call create job

	ioutil.WriteFile(inputFile, []byte(inputFileValue), 0644)
	defer func() {
		os.Remove(inputFile)
	}()
	os.Args = inputCommand
	rootCmd.Execute()
}

func testBadJSON(t *testing.T, testName string, inputFile string, inputFileValue string, inputCommand []string, expectedOutput string) {

	if os.Getenv("BE_CRASHER") == "1" {
		testJSONSyntaxErrorsReported(inputFile, inputFileValue, inputCommand)
		return
	}
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%v", testName))
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	out, err := cmd.CombinedOutput()

	require.YesError(t, err)
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		require.Equal(t, expectedOutput, string(out))
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)

}

func TestJSONSyntaxErrorsReportedCreateJob(t *testing.T) {
	descriptiveOutput := `Error parsing job spec: Syntax Error on line 3:

"356weryt

         ^
invalid character '\n' in string literal
`
	cmd := []string{"pachctl", "create-job", "-f", "bad1.json"}
	testBadJSON(t, "TestJSONSyntaxErrorsReportedCreateJob", "bad1.json", badJSON1, cmd, descriptiveOutput)
}

func TestJSONSyntaxErrorsReportedCreateJob2(t *testing.T) {
	descriptiveOutput := `Error parsing job spec: Syntax Error on line 5:

    "c": {a
          ^
invalid character 'a' looking for beginning of object key string
`
	cmd := []string{"pachctl", "create-job", "-f", "bad2.json"}
	testBadJSON(t, "TestJSONSyntaxErrorsReportedCreateJob2", "bad2.json", badJSON2, cmd, descriptiveOutput)
}

func TestJSONSyntaxErrorsReportedCreatePipeline(t *testing.T) {
	descriptiveOutput := `Error parsing pipeline spec: Syntax Error on line 5:

    "c": {a
          ^
invalid character 'a' looking for beginning of object key string
`
	cmd := []string{"pachctl", "create-pipeline", "-f", "bad2.json"}
	testBadJSON(t, "TestJSONSyntaxErrorsReportedCreatePipeline", "bad2.json", badJSON2, cmd, descriptiveOutput)
}

func TestJSONSyntaxErrorsReportedRunPipeline(t *testing.T) {
	descriptiveOutput := `Error parsing pipeline spec: Syntax Error on line 5:

    "c": {a
          ^
invalid character 'a' looking for beginning of object key string
`
	cmd := []string{"pachctl", "run-pipeline", "-f", "bad2.json", "somePipelineName"}
	testBadJSON(t, "TestJSONSyntaxErrorsReportedRunPipeline", "bad2.json", badJSON2, cmd, descriptiveOutput)
}
