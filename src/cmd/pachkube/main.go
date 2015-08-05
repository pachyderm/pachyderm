package main

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/pkg/cobramainutil"
	"github.com/pachyderm/pachyderm/src/pkg/mainutil"
	"github.com/spf13/cobra"
)

var (
	defaultEnv = map[string]string{}
)

type appEnv struct{}

func main() {
	mainutil.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	//appEnv := appEnvObj.(*appEnv)

	runCmd := cobramainutil.Command{
		Use:  "run",
		Long: "run",
		Run: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("here\n")
			return nil
		},
	}.ToCobraCommand()

	rootCmd := &cobra.Command{
		Use:  "pachkube",
		Long: `pachkube`,
	}

	rootCmd.AddCommand(runCmd)
	return rootCmd.Execute()
}
