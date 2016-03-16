package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
)

func Cmds() []*cobra.Command {
	var shards int
	kubernetesManifest := &cobra.Command{
		Use:   "manifest [amazon-secret bucket id secret token region]",
		Short: "Print a kubernetes manifest for a Pachyderm cluster.",
		Long:  "Print a kubernetes manifest for a Pachyderm cluster.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 0, Max: 5}, func(args []string) error {
			switch len(args) {
			case 0:
				assets.WriteAssets(os.Stdout, uint64(shards), false)
			case 5:
				assets.WriteAssets(os.Stdout, uint64(shards), true)
				assets.WriteAmazonSecret(os.Stdout, args[0], args[1], args[2], args[3], args[4])
			default:
				return fmt.Errorf("Expected 0 or 5 args, got %d", len(args))
			}
			return nil
		}),
	}
	kubernetesManifest.Flags().IntVarP(&shards, "shards", "s", 1, "The static number of shards for pfs.")

	var result []*cobra.Command
	result = append(result, kubernetesManifest)
	return result
}
