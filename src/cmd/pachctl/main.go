package main

import (
	"fmt"
	"os"
	"strings"

	"go.pedge.io/env"

	pfscmds "github.com/pachyderm/pachyderm/src/pfs/cmds"
	deploycmds "github.com/pachyderm/pachyderm/src/pkg/deploy/cmds"
	ppscmds "github.com/pachyderm/pachyderm/src/pps/cmds"
	"github.com/spf13/cobra"
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	PfsAddress         string `env:"PFS_ADDRESS,default=0.0.0.0:650"`
	PachydermPpsd1Port string `env:"PACHYDERM_PPSD_1_PORT"`
	PpsAddress         string `env:"PPS_ADDRESS,default=0.0.0.0:651"`
	KubernetesAddress  string `env:"KUBERNETES_ADDRESS,default=http://localhost:8080"`
	KubernetesUsername string `env:"KUBERNETES_USERNAME,default=admin"`
	KubernetesPassword string `env:"KUBERNETES_PASSWORD"`
	Provider           string `env:"PROVIDER"`
	GCEProject         string `env:"GCE_PROJECT"`
	GCEZone            string `env:"GCE_ZONE"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Long: `Access the Pachyderm API.

Envronment variables:
  PFS_ADDRESS=0.0.0.0:650, the PFS server to connect to.
  PPS_ADDRESS=0.0.0.0:651, the PPS server to connect to.
  KUBERNETES_ADDRESS=http://localhost:8080, the Kubernetes endpoint to connect to.
  KUBERNETES_USERNAME=admin
  KUBERNETES_PASSWORD
  PROVIDER, which provider to use for cluster creation (currently only supports GCE).
  GCE_PROJECT
  GCE_ZONE`,
	}
	pfsCmds, err := pfscmds.Cmds(getPfsdAddress(appEnv))
	if err != nil {
		return err
	}
	for _, cmd := range pfsCmds {
		rootCmd.AddCommand(cmd)
	}
	ppsCmds, err := ppscmds.Cmds(getPpsdAddress(appEnv))
	if err != nil {
		return err
	}
	for _, cmd := range ppsCmds {
		rootCmd.AddCommand(cmd)
	}

	deployCmds, err := deploycmds.Cmds(
		appEnv.KubernetesAddress,
		appEnv.KubernetesUsername,
		appEnv.KubernetesAddress,
		appEnv.Provider,
		appEnv.GCEProject,
		appEnv.GCEZone,
	)
	if err != nil {
		return err
	}
	for _, cmd := range deployCmds {
		rootCmd.AddCommand(cmd)
	}

	return rootCmd.Execute()
}

func getPfsdAddress(appEnv *appEnv) string {
	if pfsdAddr := os.Getenv("PFSD_PORT_650_TCP_ADDR"); pfsdAddr != "" {
		return fmt.Sprintf("%s:650", pfsdAddr)
	}
	if appEnv.PachydermPfsd1Port != "" {
		return strings.Replace(appEnv.PachydermPfsd1Port, "tcp://", "", -1)
	}
	return appEnv.PfsAddress
}

func getPpsdAddress(appEnv *appEnv) string {
	if ppsdAddr := os.Getenv("PPSD_PORT_651_TCP_ADDR"); ppsdAddr != "" {
		return fmt.Sprintf("%s:651", ppsdAddr)
	}
	if appEnv.PachydermPpsd1Port != "" {
		return strings.Replace(appEnv.PachydermPpsd1Port, "tcp://", "", -1)
	}
	return appEnv.PpsAddress
}
