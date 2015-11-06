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
		Use:   "create-cluster cluster-name nodes shards replicas",
		Short: "Create a new pachyderm cluster.",
		Long:  "Create a new pachyderm cluster.",
		Run: pkgcobra.RunFixedArgs(4, func(args []string) error {
			apiServer, err := getAPIServer(kubernetesAddress, kubernetesUsername, kubernetesPassword, providerName, gceProject, gceZone)
			if err != nil {
				return err
			}
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
			apiServer, err := getAPIServer(kubernetesAddress, kubernetesUsername, kubernetesPassword, providerName, gceProject, gceZone)
			if err != nil {
				return err
			}
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

	var result []*cobra.Command
	result = append(result, createCluster)
	result = append(result, deleteCluster)
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
