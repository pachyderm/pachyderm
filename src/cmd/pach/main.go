package main

import (
	"strings"

	"go.pedge.io/env"

	deploycmds "github.com/pachyderm/pachyderm/src/cmd/deploy/cmds"
	pfscmds "github.com/pachyderm/pachyderm/src/cmd/pfs/cmds"
	ppscmds "github.com/pachyderm/pachyderm/src/cmd/pps/cmds"
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
		Use: "pach",
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
	pfsAddress := appEnv.PachydermPfsd1Port
	if pfsAddress == "" {
		pfsAddress = appEnv.PfsAddress
	} else {
		pfsAddress = strings.Replace(pfsAddress, "tcp://", "", -1)
	}
	pfsCmds, err := pfscmds.Cmds(pfsAddress)
	if err != nil {
		return err
	}
	for _, cmd := range pfsCmds {
		rootCmd.AddCommand(cmd)
	}
	ppsAddress := appEnv.PachydermPpsd1Port
	if ppsAddress == "" {
		ppsAddress = appEnv.PpsAddress
	} else {
		ppsAddress = strings.Replace(ppsAddress, "tcp://", "", -1)
	}
	ppsCmds, err := ppscmds.Cmds(ppsAddress)
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
