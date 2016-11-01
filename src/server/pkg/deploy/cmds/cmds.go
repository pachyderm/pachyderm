package cmds

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/pkg/exec"
)

func maybeKcCreate(dryRun bool, manifest *bytes.Buffer) error {
	if dryRun {
		_, err := os.Stdout.Write(manifest.Bytes())
		return err
	}
	return pkgexec.RunIO(
		pkgexec.IO{
			Stdin:  manifest,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}, "kubectl", "create", "-f", "-")
}

// DeployCmd returns a cobra command for deploying a pachyderm cluster.
func DeployCmd() *cobra.Command {
	var shards int
	var hostPath string
	var dryRun bool
	var registry bool
	var rethinkdbCacheSize string
	var curVersion = version.PrettyPrintVersion(version.Version)

	deployLocal := &cobra.Command{
		Use:   "local",
		Short: "Deploy a local, single-node Pachyderm cluster.",
		Long:  "Deploy a local, single-node Pachyderm cluster.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 0, Max: 0}, func(args []string) error {
			manifest := &bytes.Buffer{}
			assets.WriteLocalAssets(manifest, uint64(shards), hostPath, registry, rethinkdbCacheSize, deploy.DevVersionTag)
			return maybeKcCreate(dryRun, manifest)
		}),
	}
	deployLocal.Flags().StringVar(&hostPath, "host-path", "/tmp/pach", "Location on the host machine where PFS metadata will be stored.")

	deployGoogle := &cobra.Command{
		Use:   "google <GCS bucket> <GCE persistent disk> <Disk size (in GB)>",
		Short: "Deploy a Pachyderm cluster running on GCP.",
		Long:  "Deploy a Pachyderm cluster running on GCP.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 3, Max: 3}, func(args []string) error {
			volumeName := args[1]
			volumeSize, err := strconv.Atoi(args[2])
			if err != nil {
				return fmt.Errorf("volume size needs to be an integer; instead got %v", args[2])
			}
			manifest := &bytes.Buffer{}
			assets.WriteGoogleAssets(manifest, uint64(shards), args[0],
				volumeName, volumeSize, registry, rethinkdbCacheSize, curVersion)
			return maybeKcCreate(dryRun, manifest)
		}),
	}

	deployAmazon := &cobra.Command{
		Use:   "amazon <S3 bucket> <id> <secret> <token> <region> <EBS volume name> <volume size (in GB)>",
		Short: "Deploy a Pachyderm cluster running on AWS.",
		Long:  "Deploy a Pachyderm cluster running on AWS.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 7, Max: 7}, func(args []string) error {
			volumeName := args[5]
			volumeSize, err := strconv.Atoi(args[6])
			if err != nil {
				return fmt.Errorf("volume size needs to be an integer; instead got %v", args[6])
			}
			manifest := &bytes.Buffer{}
			assets.WriteAmazonAssets(manifest, uint64(shards), args[0], args[1], args[2], args[3],
				args[4], volumeName, volumeSize, registry, rethinkdbCacheSize, curVersion)
			return maybeKcCreate(dryRun, manifest)
		}),
	}

	deployMicrosoft := &cobra.Command{
		Use:   "microsoft <container> <storage account name> <storage account key> <volume uri> <volume size in GB>",
		Short: "Deploy a Pachyderm cluster running on Microsoft Azure.",
		Long:  "Deploy a Pachyderm cluster running on Microsoft Azure.",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 5, Max: 5}, func(args []string) error {
			_, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("storage-account-key needs to be base64 encoded; instead got '%v'", args[2])
			}
			volumeURI, err := url.ParseRequestURI(args[3])
			if err != nil {
				return fmt.Errorf("volume-uri needs to be a well-formed URI; instead got '%v'", args[3])
			}
			volumeSize, err := strconv.Atoi(args[4])
			if err != nil {
				return fmt.Errorf("volume size needs to be an integer; instead got %v", args[4])
			}
			manifest := &bytes.Buffer{}
			assets.WriteMicrosoftAssets(manifest, uint64(shards), args[0], args[1], args[2],
				volumeURI.String(), volumeSize, registry, rethinkdbCacheSize, curVersion)
			return maybeKcCreate(dryRun, manifest)
		}),
	}

	cmd := &cobra.Command{
		Use:   "deploy amazon|google|microsoft|local",
		Short: "Deploy a Pachyderm cluster.",
		Long:  "Deploy a Pachyderm cluster.",
	}
	cmd.PersistentFlags().IntVarP(&shards, "shards", "s", 32, "The static number of shards for pfs.")
	cmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "", false, "Don't actually deploy pachyderm to Kubernetes, instead just print the manifest.")
	cmd.PersistentFlags().BoolVarP(&registry, "registry", "r", true, "Deploy a docker registry along side pachyderm.")
	cmd.PersistentFlags().StringVar(&rethinkdbCacheSize, "rethinkdb-cache-size", "768M", "Size of in-memory cache to use for Pachyderm's RethinkDB instance, "+
		"e.g. \"2G\". Default is \"768M\". Size is specified in bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc)")
	cmd.AddCommand(deployLocal)
	cmd.AddCommand(deployAmazon)
	cmd.AddCommand(deployGoogle)
	cmd.AddCommand(deployMicrosoft)
	return cmd
}
