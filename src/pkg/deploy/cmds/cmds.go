package cmds

import (
	"os"

	"github.com/pachyderm/pachyderm/src/pkg/deploy/assets"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
)

func Cmds() []*cobra.Command {
	var shards int
	var secrets bool
	kubernetesManifest := &cobra.Command{
		Use:   "manifest",
		Short: "Print a kubernetes manifest for a Pachyderm cluster.",
		Long:  "Print a kubernetes manifest for a Pachyderm cluster.",
		Run: pkgcobra.RunFixedArgs(0, func([]string) error {
			assets.WriteAssets(os.Stdout, uint64(shards), secrets)
			return nil
		}),
	}
	kubernetesManifest.Flags().IntVarP(&shards, "shards", "s", 1, "The static number of shards for pfs.")
	kubernetesManifest.Flags().BoolVarP(&secrets, "secrets", "t", false, "Whether to print storage secrets.")

	amazonSecret := &cobra.Command{
		Use:   "amazon-secret bucket id secret token region",
		Short: "Print a kubernetes secret containing amazon credentials.",
		Long:  "Print a kubernetes secret containing amazon credentials.",
		Run: pkgcobra.RunFixedArgs(5, func(args []string) error {
			assets.WriteAmazonSecret(os.Stdout, args[0], args[1], args[2], args[3], args[4])
			return nil
		}),
	}

	var result []*cobra.Command
	result = append(result, kubernetesManifest)
	result = append(result, amazonSecret)
	return result
}
