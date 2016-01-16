package main

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/cmd/pachctl/cmd"
	"go.pedge.io/env"
)

type appEnv struct {
	PfsAddress string `env:"PFS_ADDRESS,default=0.0.0.0:650"`
	PpsAddress string `env:"PPS_ADDRESS,default=0.0.0.0:651"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	pfsdAddress := getPfsdAddress(appEnv)
	ppsdAddress := getPpsdAddress(appEnv)
	rootCmd, err := cmd.PachctlCmd(pfsdAddress, ppsdAddress)
	if err != nil {
		return err
	}
	return rootCmd.Execute()
}

func getPfsdAddress(appEnv *appEnv) string {
	if pfsdAddr := os.Getenv("PFSD_PORT_650_TCP_ADDR"); pfsdAddr != "" {
		return fmt.Sprintf("%s:650", pfsdAddr)
	}
	return appEnv.PfsAddress
}

func getPpsdAddress(appEnv *appEnv) string {
	if ppsdAddr := os.Getenv("PPSD_PORT_651_TCP_ADDR"); ppsdAddr != "" {
		return fmt.Sprintf("%s:651", ppsdAddr)
	}
	return appEnv.PpsAddress
}
