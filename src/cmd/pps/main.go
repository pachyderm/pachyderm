package main

import (
	"strings"

	"go.pedge.io/env"

	"github.com/pachyderm/pachyderm/src/cmd/pps/cmds"
	"github.com/spf13/cobra"
)

type appEnv struct {
	PachydermPpsd1Port string `env:"PACHYDERM_PPSD_1_PORT"`
	Address            string `env:"PPS_ADDRESS,default=0.0.0.0:651"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	address := appEnv.PachydermPpsd1Port
	if address == "" {
		address = appEnv.Address
	} else {
		address = strings.Replace(address, "tcp://", "", -1)
	}
	rootCmd := &cobra.Command{
		Use: "pps",
		Long: `Access the PPS API.

Envronment variables:
  PPS_ADDRESS=0.0.0.0:651, the PPS server to connect to.`,
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
