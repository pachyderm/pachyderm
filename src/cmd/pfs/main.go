package main

import (
	"strings"

	"go.pedge.io/env"

	"github.com/pachyderm/pachyderm/src/cmd/pfs/cmds"
	"github.com/spf13/cobra"
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	Address            string `env:"PFS_ADDRESS,default=0.0.0.0:650"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	address := appEnv.PachydermPfsd1Port
	if address == "" {
		address = appEnv.Address
	} else {
		address = strings.Replace(address, "tcp://", "", -1)
	}
	rootCmd := &cobra.Command{
		Use: "pfs",
		Long: `Access the PFS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PFS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:650.`,
	}
	cmds, err := cmds.Cmds(address)
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}

	return rootCmd.Execute()
}
