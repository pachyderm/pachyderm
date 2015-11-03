package main

import (
	"fmt"
	"os"
	"strings"

	"go.pedge.io/env"

	"github.com/pachyderm/pachyderm/src/cmd/pfs/cmds"
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
	cmds, err := cmds.Cmds(address)
	if err != nil {
		return err
	}
	rootCmd := &cobra.Command{
		Use: "pfs",
		Long: `Access the PFS API.

Note that this CLI is experimental and does not even check for common errors.
The environment variable PFS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:650.`,
	}
	for _, cmd := range cmds {
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
