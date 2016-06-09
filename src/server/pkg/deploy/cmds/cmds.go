package cmds

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
)

func DeployCmd() *cobra.Command {
	var shards int
	var hostPath string
	cmd := &cobra.Command{
		Use:   os.Args[0] + " [amazon bucket id secret token region [volume-name volume-size-in-GB] | google bucket [volume-name volume-size-in-GB]]",
		Short: "Print a kubernetes manifest for a Pachyderm cluster.",
		Long:  "Print a kubernetes manifest for a Pachyderm cluster.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 0, Max: 8}, func(args []string) error {
			if len(args) == 0 {
				assets.WriteLocalAssets(os.Stdout, uint64(shards), hostPath)
			} else {
				var volumeName string
				var volumeSize int
				var err error
				// If we are given volume name and size...
				if len(args) == 8 || len(args) == 4 {
					volumeName = args[len(args)-2]
					volumeSize, err = strconv.Atoi(args[len(args)-1])
					if err != nil {
						return fmt.Errorf("volume size needs to be an integer; instead got %v", args[len(args)-1])
					}
				}
				switch args[0] {
				case "amazon":
					if len(args) != 6 && len(args) != 8 {
						return fmt.Errorf("Expected 6 or 8 args, got %d", len(args))
					}
					assets.WriteAmazonAssets(os.Stdout, uint64(shards), args[1], args[2], args[3], args[4], args[5], volumeName, volumeSize)
				case "google":
					if len(args) != 2 && len(args) != 4 {
						return fmt.Errorf("Expected 2 or 4 args, got %d", len(args))
					}
					assets.WriteGoogleAssets(os.Stdout, uint64(shards), args[1], volumeName, volumeSize)
				}
			}
			return nil
		}),
	}
	cmd.Flags().IntVarP(&shards, "shards", "s", 32, "The static number of shards for pfs.")
	cmd.Flags().StringVarP(&hostPath, "host-path", "p", "", "the path on the host machine where data will be stored; this is only relevant if you are running pachyderm locally; by default, a directory under /tmp is used")
	return cmd
}
