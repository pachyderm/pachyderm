package main

import (
	"os"

	"github.com/pachyderm/pachyderm/src/server/cmd/pachctl/cmd"
	"github.com/spf13/pflag"
	"go.pedge.io/env"
)

type appEnv struct {
	Address string `env:"ADDRESS,default=0.0.0.0:30650"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	appEnv := appEnvObj.(*appEnv)
	rootCmd, err := cmd.PachctlCmd(appEnv.Address)
	if err != nil {
		return err
	}
	return rootCmd.Execute()
}
