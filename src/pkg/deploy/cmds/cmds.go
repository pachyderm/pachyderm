package cmds

import (
	"os"

	"github.com/pachyderm/pachyderm/src/pkg/deploy/assets"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
)

func Cmds(
	kubernetesAddress string,
	kubernetesUsername string,
	kubernetesPassword string,
	providerName string,
	gceProject string,
	gceZone string,
) ([]*cobra.Command, error) {
	var shards int
	createCluster := &cobra.Command{
		Use:   "create-cluster",
		Short: "Create a new pachyderm cluster.",
		Long:  "Create a new pachyderm cluster.",
		Run: pkgcobra.RunFixedArgs(0, func([]string) error {
			assets.PrintAssets(os.Stdout, uint64(shards))
			return nil
		}),
	}
	createCluster.Flags().IntVarP(&shards, "shards", "s", 1, "The static number of shards for pfs.")

	var result []*cobra.Command
	result = append(result, createCluster)
	return result, nil
}
