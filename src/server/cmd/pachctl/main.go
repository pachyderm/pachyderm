package main

import (
	"os"

	"github.com/pachyderm/pachyderm/src/server/cmd/pachctl/cmd"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/spf13/pflag"
)

type appEnv struct {
	Address string `env:"ADDRESS,default=0.0.0.0:30650"`
}

func main() {
	cmdutil.Main(do, &appEnv{})
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
