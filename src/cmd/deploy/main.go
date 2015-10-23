package main

import (
	"strconv"

	"github.com/spf13/cobra"
	"go.pachyderm.com/pachyderm/src/pkg/deploy"
	"go.pachyderm.com/pachyderm/src/pkg/deploy/server"
	"go.pedge.io/env"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/protolog/logrus"
	"golang.org/x/net/context"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

var (
	defaultEnv = map[string]string{
		"KUBERNETES_ADDRESS": "https://localhost:8080",
	}
)

type appEnv struct {
	KubernetesAddress string `env:"KUBERNETES_ADDRESS"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	logrus.Register()
	config := &client.Config{
		Host: appEnv.KubernetesAddress,
	}
	client, err := client.New(config)
	if err != nil {
		return err
	}
	apiServer := server.NewAPIServer(client)

	createCluster := &cobra.Command{
		Use:   "create-cluster cluster-name nodes shards replicas",
		Short: "Create a new pachyderm cluster.",
		Long:  "Create a new pachyderm cluster.",
		Run: pkgcobra.RunFixedArgs(4, func(args []string) error {
			nodes, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			shards, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}
			replicas, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}
			_, err = apiServer.CreateCluster(
				context.Background(),
				&deploy.CreateClusterRequest{
					Cluster: &deploy.Cluster{
						Name: args[0],
					},
					Nodes:    nodes,
					Shards:   shards,
					Replicas: replicas,
				})
			return err
		}),
	}

	rootCmd := &cobra.Command{
		Use: "deploy",
		Long: `Deploy Pachyderm clusters.

The environment variable KUBERNETES_ADDRESS controls the Kubernetes endpoint the CLI connects to, the default is https://localhost:8080.`,
	}
	rootCmd.AddCommand(createCluster)
	return rootCmd.Execute()
}
