package main

import (
	"fmt"
	"os"
	"strings"

	"go.pedge.io/env"

	pfscmds "github.com/pachyderm/pachyderm/src/cmd/pfs/cmds"
	ppscmds "github.com/pachyderm/pachyderm/src/cmd/pps/cmds"
	"github.com/spf13/cobra"
)

var (
	defaultEnv = map[string]string{
		"PFS_ADDRESS": "0.0.0.0:650",
	}
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	Address            string `env:"PFS_ADDRESS"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
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
		Use: "pach",
		Long: `Access the Pachyderm API.

The environment variable PFS_ADDRESS controls which PFS server the CLI connects to, the default is 0.0.0.0:650.`,
	}
	pfscmds, err := pfscmds.Cmds(address)
	if err != nil {
		return err
	}
	for _, cmd := range pfscmds {
		rootCmd.AddCommand(cmd)
	}
	ppscmds, err := ppscmds.Cmds(address)
	if err != nil {
		return err
	}
	for _, cmd := range ppscmds {
		rootCmd.AddCommand(cmd)
	}

	return rootCmd.Execute()
}

func getPfsdAddress() string {
	pfsdAddr := os.Getenv("PFSD_PORT_650_TCP_ADDR")
	if pfsdAddr == "" {
		return "0.0.0.0:650"
	}
	return fmt.Sprintf("%s:650", pfsdAddr)
}
