package main

import (
	"github.com/pachyderm/pachyderm/src/cmd/deploy/cmds"
	"github.com/spf13/cobra"
	"go.pedge.io/env"
	"go.pedge.io/protolog/logrus"
)

type appEnv struct {
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
	logrus.Register()
	rootCmd := &cobra.Command{
		Use: "deploy",
		Long: `Deploy Pachyderm clusters.

Envronment variables:
  KUBERNETES_ADDRESS=http://localhost:8080
  KUBERNETES_USERNAME=admin
  KUBERNETES_PASSWORD
  PROVIDER, which provider to use for cluster creation (currently only supports GCE).
  GCE_PROJECT
  GCE_ZONE`,
	}
	cmds, err := cmds.Cmds(
		appEnv.KubernetesAddress,
		appEnv.KubernetesUsername,
		appEnv.KubernetesPassword,
		appEnv.Provider,
		appEnv.GCEProject,
		appEnv.GCEZone,
	)
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}
	return rootCmd.Execute()
}
