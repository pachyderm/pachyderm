package cmds

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	_metrics "github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/pkg/exec"
)

// DeployCmd returns a cobra command for deploying a pachyderm cluster.
func DeployCmd(metrics bool) *cobra.Command {
	var shards int
	var hostPath string
	var dev bool
	var dryRun bool
	var registry bool
	var rethinkdbCacheSize string
	var logLevel string
	cmd := &cobra.Command{
		Use:   "deploy [amazon bucket id secret token region volume-name volume-size-in-GB | google bucket volume-name volume-size-in-GB | microsoft container storage-account-name storage-account-key volume-uri volume-size-in-GB]",
		Short: "Print a kubernetes manifest for a Pachyderm cluster.",
		Long:  "Print a kubernetes manifest for a Pachyderm cluster.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 0, Max: 8}, func(args []string) (retErr error) {
			if metrics && !dev {
				finalMetrics := _metrics.ReportAndFlushUserAction("Deploy")
				defer func(start time.Time) { finalMetrics(start, retErr) }(time.Now())
			}
			version := version.PrettyPrintVersion(version.Version)
			if dev {
				version = deploy.DevVersionTag
			}
			var out io.Writer
			var manifest bytes.Buffer
			out = &manifest
			if dryRun {
				out = os.Stdout
			}
			opts := &assets.AssetOpts{
				Shards:             uint64(shards),
				Registry:           registry,
				RethinkdbCacheSize: rethinkdbCacheSize,
				Version:            version,
				LogLevel:           logLevel,
				Metrics:            metrics,
			}
			if len(args) == 0 {
				assets.WriteLocalAssets(out, opts, hostPath)
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
					assets.WriteAmazonAssets(out, opts, args[1], args[2], args[3], args[4], args[5], volumeName, volumeSize)
				case "google":
					if len(args) != 4 {
						return fmt.Errorf("expected 4 args, got %d", len(args))
					}
					volumeName := args[2]
					volumeSize, err := strconv.Atoi(args[3])
					if err != nil {
						return fmt.Errorf("volume size needs to be an integer; instead got %v", args[3])
					}
					assets.WriteGoogleAssets(out, opts, args[1], volumeName, volumeSize)
				case "microsoft":
					if len(args) != 6 {
						return fmt.Errorf("expected 6 args, got %d", len(args))
					}
					_, err := base64.StdEncoding.DecodeString(args[3])
					if err != nil {
						return fmt.Errorf("storage-account-key needs to be base64 encoded; instead got '%v'", args[3])
					}
					volumeURI, err := url.ParseRequestURI(args[4])
					if err != nil {
						return fmt.Errorf("volume-uri needs to be a well-formed URI; instead got '%v'", args[4])
					}
					volumeSize, err := strconv.Atoi(args[5])
					if err != nil {
						return fmt.Errorf("volume size needs to be an integer; instead got %v", args[5])
					}
					assets.WriteMicrosoftAssets(out, opts, args[1], args[2], args[3], volumeURI.String(), volumeSize)
				default:
					return fmt.Errorf("expected one of google, amazon, or microsoft; instead got '%v'", args[0])
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
	cmd.Flags().BoolVarP(&registry, "registry", "r", true, "Deploy a docker registry along side pachyderm.")
	cmd.Flags().StringVar(&rethinkdbCacheSize, "rethinkdb-cache-size", "768M", "Size of in-memory cache to use for Pachyderm's RethinkDB instance, "+
		"e.g. \"2G\". Default is \"768M\". Size is specified in bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc)")
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "The level of log messages to print options are, from least to most verbose: \"error\", \"info\", \"debug\".")
	return cmd
}
