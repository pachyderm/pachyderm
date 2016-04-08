package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
)

func DeployCmd() *cobra.Command {
	var shards int
	cmd := &cobra.Command{
		Use:   os.Args[0] + " [amazon bucket id secret token region | google bucket]",
		Short: "Print a kubernetes manifest for a Pachyderm cluster.",
		Long:  "Print a kubernetes manifest for a Pachyderm cluster.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 0, Max: 6}, func(args []string) error {
			if len(args) == 0 {
				assets.WriteLocalAssets(os.Stdout, uint64(shards))
			} else {
				switch args[0] {
				case "amazon":
					if len(args) != 6 {
						return fmt.Errorf("Expected 6 args, got %d", len(args))
					}
					assets.WriteAmazonAssets(os.Stdout, uint64(shards), args[1], args[2], args[3], args[4], args[5])
				case "google":
					if len(args) != 2 {
						return fmt.Errorf("Expected 2 args, got %d", len(args))
					}
					assets.WriteGoogleAssets(os.Stdout, uint64(shards), args[1])
				}
			}
			return nil
		}),
	}
	cmd.Flags().IntVarP(&shards, "shards", "s", 32, "The static number of shards for pfs.")
	return cmd
}
