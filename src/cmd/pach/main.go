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
	GCEProject         string `env:"GCE_PROJECT,required"`
	GCEZone            string `env:"GCE_ZONE,required"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	rootCmd := &cobra.Command{
		Use: "pach",
		Long: `Access the Pachyderm API.

The environment variable PFS_ADDRESS controls which PFS server the CLI connects to, the default is 0.0.0.0:650.
The environment variable PPS_ADDRESS controls what server the CLI connects to, the default is 0.0.0.0:651.`,
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

	deployCmds, err := deploycmds.Cmds(appEnv.KubernetesAddress, appEnv.KubernetesUsername, appEnv.KubernetesAddress, appEnv.GCEProject, appEnv.GCEZone)
	if err != nil {
		return err
	}
	for _, cmd := range deployCmds {
		rootCmd.AddCommand(cmd)
	}

	return rootCmd.Execute()
}
