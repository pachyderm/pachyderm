package cmds

import (
	"strconv"

	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
	"golang.org/x/net/context"
	kube "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/pachyderm/pachyderm/src/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/pkg/deploy/server"
	provider "github.com/pachyderm/pachyderm/src/pkg/provider"
)

func Cmds(
	kubernetesAddress string,
	kubernetesUsername string,
	kubernetesPassword string,
	providerName string,
	gceProject string,
	gceZone string,
) ([]*cobra.Command, error) {
	createCluster := &cobra.Command{
		Use:   "create-cluster replicas shards ",
		Short: "Create a new pachyderm cluster.",
		Long:  "Create a new pachyderm cluster.",
		Run: pkgcobra.RunFixedArgs(2, func(args []string) error {
			apiServer, err := getAPIServer(kubernetesAddress, kubernetesUsername, kubernetesPassword, providerName, gceProject, gceZone)
			if err != nil {
				return err
			}
			nodes, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			shards, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			_, err = apiServer.CreateCluster(
				context.Background(),
				&deploy.CreateClusterRequest{
					Cluster: &deploy.Cluster{
						Name: args[0],
					},
					Nodes:  nodes,
					Shards: shards,
				})
			return err
		}),
	}

	var result []*cobra.Command
	result = append(result, createCluster)
	return result, nil
}

func getAPIServer(
	kubernetesAddress string,
	kubernetesUsername string,
	kubernetesPassword string,
	providerName string,
	gceProject string,
	gceZone string,
) (server.APIServer, error) {
	config := &kube.Config{
		Host:     kubernetesAddress,
		Insecure: true,
		Username: kubernetesUsername,
		Password: kubernetesPassword,
	}
	kubeClient, err := kube.New(config)
	if err != nil {
		return nil, err
	}
	provider, err := getProvider(providerName, gceProject, gceZone)
	if err != nil {
		return nil, err
	}
	return server.NewAPIServer(kubeClient, provider), nil
}

func getProvider(providerName string, gceProject string, gceZone string) (provider.Provider, error) {
	if providerName == "gce" {
		return provider.NewGoogleProvider(context.TODO(), gceProject, gceZone)
	}
	return nil, nil
}
