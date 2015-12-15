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
	kubernetesManifest := &cobra.Command{
		Use:   "manifest",
		Short: "Print a kubernetes manifest for a Pachyderm cluster.",
		Long:  "Print a kubernetes manifest for a Pachyderm cluster.",
		Run: pkgcobra.RunFixedArgs(0, func([]string) error {
			assets.WriteAssets(os.Stdout, uint64(shards))
			return nil
		}),
	}
	kubernetesManifest.Flags().IntVarP(&shards, "shards", "s", 1, "The static number of shards for pfs.")

	var result []*cobra.Command
	result = append(result, kubernetesManifest)
	return result, nil
}
