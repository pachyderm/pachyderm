package main

import (
	"strconv"

	"github.com/pachyderm/pachyderm/src/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/pkg/deploy/server"
	"github.com/pachyderm/pachyderm/src/pkg/provider"
	"github.com/spf13/cobra"
	"go.pedge.io/env"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/protolog/logrus"
	"golang.org/x/net/context"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

var (
	defaultEnv = map[string]string{
		"KUBERNETES_ADDRESS":  "http://localhost:8080",
		"KUBERNETES_USERNAME": "admin",
	}
)

type appEnv struct {
	KubernetesAddress  string `env:"KUBERNETES_ADDRESS"`
	KubernetesUsername string `env:"KUBERNETES_USERNAME"`
	KubernetesPassword string `env:"KUBERNETES_PASSWORD"`
	GCEProject         string `env:"GCE_PROJECT"`
	GCEZone            string `env:"GCE_ZONE"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	logrus.Register()
	config := &client.Config{
		Host:     appEnv.KubernetesAddress,
		Insecure: true,
		Username: appEnv.KubernetesUsername,
		Password: appEnv.KubernetesPassword,
	}
	client, err := client.New(config)
	if err != nil {
		return err
	}
	provider, err := provider.NewGoogleProvider(context.TODO(), appEnv.GCEProject, appEnv.GCEZone)
	if err != nil {
		return err
	}
	apiServer := server.NewAPIServer(client, provider)

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

	deleteCluster := &cobra.Command{
		Use:   "delete-cluster cluster-name",
		Short: "Delete a cluster.",
		Long:  "Delete a cluster.",
		Run: pkgcobra.RunFixedArgs(1, func(args []string) error {
			_, err = apiServer.DeleteCluster(
				context.Background(),
				&deploy.DeleteClusterRequest{
					Cluster: &deploy.Cluster{
						Name: args[0],
					},
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
	rootCmd.AddCommand(deleteCluster)
	return rootCmd.Execute()
}
