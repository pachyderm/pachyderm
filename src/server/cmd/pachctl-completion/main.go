package main

import (
	"os"

	"github.com/pachyderm/pachyderm/src/server/cmd/pachctl/cmd"
	"go.pedge.io/env"
)

type appEnv struct{}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	// passing an empty address but that's fine because we're not going to
	// execute the command but print docs with it
	rootCmd, err := cmd.PachctlCmd("")
	if err != nil {
		return err
	}
	return rootCmd.GenBashCompletion(os.Stdout)
}
