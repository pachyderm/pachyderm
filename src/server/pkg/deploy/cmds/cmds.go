package cmds

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/pkg/exec"
)

// DeployCmd returns a cobra command for deploying a pachyderm cluster.
func DeployCmd() *cobra.Command {
	var shards int
	var hostPath string
	var dev bool
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "deploy [amazon bucket id secret token region volume-name volume-size-in-GB | google bucket volume-name volume-size-in-GB]",
		Short: "Print a kubernetes manifest for a Pachyderm cluster.",
		Long:  "Print a kubernetes manifest for a Pachyderm cluster.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 0, Max: 8}, func(args []string) error {
			version := version.PrettyPrintVersion(version.Version)
			if dev {
				version = "local"
			}
			var out io.Writer
			var manifest bytes.Buffer
			out = &manifest
			if dryRun {
				out = os.Stdout
			}
			if len(args) == 0 {
				assets.WriteLocalAssets(out, uint64(shards), hostPath, version)
			} else {
				switch args[0] {
				case "amazon":
					if len(args) != 8 {
						return fmt.Errorf("expected 8 args, got %d", len(args))
					}
					volumeName := args[6]
					volumeSize, err := strconv.Atoi(args[7])
					if err != nil {
						return fmt.Errorf("volume size needs to be an integer; instead got %v", args[7])
					}
					assets.WriteAmazonAssets(out, uint64(shards), args[1], args[2], args[3], args[4],
						args[5], volumeName, volumeSize, version)
				case "google":
					if len(args) != 4 {
						return fmt.Errorf("expected 4 args, got %d", len(args))
					}
					volumeName := args[2]
					volumeSize, err := strconv.Atoi(args[3])
					if err != nil {
						return fmt.Errorf("volume size needs to be an integer; instead got %v", args[3])
					}
					assets.WriteGoogleAssets(out, uint64(shards), args[1], volumeName, volumeSize, version)
				}
			}
			if !dryRun {
				return pkgexec.RunIO(
					pkgexec.IO{
						Stdin:  &manifest,
						Stdout: os.Stdout,
						Stderr: os.Stderr,
					}, "kubectl", "create", "-f", "-")
			}
			return nil
		}),
	}
	cmd.Flags().IntVarP(&shards, "shards", "s", 32, "The static number of shards for pfs.")
	cmd.Flags().StringVarP(&hostPath, "host-path", "p", "/tmp/pach", "the path on the host machine where data will be stored; this is only relevant if you are running pachyderm locally.")
	cmd.Flags().BoolVarP(&dev, "dev", "d", false, "Don't use a specific version of pachyderm/pachd.")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "", false, "Don't actually deploy pachyderm to Kubernetes, instead just print the manifest.")
	return cmd
}
