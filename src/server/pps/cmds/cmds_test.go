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

func testJSONSyntaxErrorsReported() {
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

	ioutil.WriteFile("bad1.json", []byte(badJSON1), 0644)
	defer func() {
		os.Remove("bad1.json")
	}()
	os.Args = []string{"pachctl", "create-job", "-f", "bad1.json"}

	rootCmd.Execute()
}

func TestJSONSyntaxErrorsReported(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		testJSONSyntaxErrorsReported()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestJSONSyntaxErrorsReported")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	out, err := cmd.CombinedOutput()

	require.YesError(t, err)
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {

		descriptiveOutput := `Error parsing job spec: Syntax Error on line 3:

"356weryt

         ^
invalid character '\n' in string literal
`

		require.Equal(t, descriptiveOutput, string(out))
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)

	fmt.Printf("err? %v\n", err)
	fmt.Printf("asdf")

}
