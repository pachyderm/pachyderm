package cmds

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	_metrics "github.com/pachyderm/pachyderm/src/server/pkg/metrics"

	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
)

var defaultDashImage = "pachyderm/dash:0.4.0"

func maybeKcCreate(dryRun bool, manifest *bytes.Buffer, opts *assets.AssetOpts) error {
	if dryRun {
		_, err := os.Stdout.Write(manifest.Bytes())
		return err
	}
	ret := cmdutil.RunIO(
		cmdutil.IO{
			Stdin:  manifest,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}, "kubectl", "create", "-f", "-")
	if !dryRun {
		fmt.Println("\nPachyderm is launching. Check it's status with \"kubectl get all\"")
		if opts.DashOnly || opts.EnableDash {
			fmt.Println("Once launched, access the dashboard by running \"pachctl port-forward\"")
		}
		fmt.Println("")
	}
	return ret
}

// DeployCmd returns a cobra.Command to deploy pachyderm.
func DeployCmd(noMetrics *bool) *cobra.Command {
	metrics := !*noMetrics
	var pachdShards int
	var hostPath string
	var dev bool
	var dryRun bool
	var secure bool
	var etcdNodes int
	var etcdVolume string
	var pachdCPURequest string
	var pachdNonCacheMemRequest string
	var blockCacheSize string
	var etcdCPURequest string
	var etcdMemRequest string
	var logLevel string
	var persistentDiskBackend string
	var objectStoreBackend string
	var opts *assets.AssetOpts
	var enableDash bool
	var dashOnly bool
	var dashImage string

	deployLocal := &cobra.Command{
		Use:   "local",
		Short: "Deploy a single-node Pachyderm cluster with local metadata storage.",
		Long:  "Deploy a single-node Pachyderm cluster with local metadata storage.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			if metrics && !dev {
				start := time.Now()
				startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
				defer startMetricsWait()
				defer func() {
					finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
					finishMetricsWait()
				}()
			}
			manifest := &bytes.Buffer{}
			if dev {
				opts.Version = deploy.DevVersionTag
			}
			if err := assets.WriteLocalAssets(manifest, opts, hostPath); err != nil {
				return err
			}
			return maybeKcCreate(dryRun, manifest, opts)
		}),
	}
	deployLocal.Flags().StringVar(&hostPath, "host-path", "/var/pachyderm", "Location on the host machine where PFS metadata will be stored.")
	deployLocal.Flags().BoolVarP(&dev, "dev", "d", false, "Don't use a specific version of pachyderm/pachd.")

	deployGoogle := &cobra.Command{
		Use:   "google <GCS bucket> <size of disk(s) (in GB)>",
		Short: "Deploy a Pachyderm cluster running on GCP.",
		Long: "Deploy a Pachyderm cluster running on GCP.\n" +
			"Arguments are:\n" +
			"  <GCS bucket>: A GCS bucket where Pachyderm will store PFS data.\n" +
			"  <GCE persistent disks>: A comma-separated list of GCE persistent disks, one per etcd node (see --etcd-nodes).\n" +
			"  <size of disks>: Size of GCE persistent disks in GB (assumed to all be the same).\n",
		Run: cmdutil.RunFixedArgs(2, func(args []string) (retErr error) {
			if metrics && !dev {
				start := time.Now()
				startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
				defer startMetricsWait()
				defer func() {
					finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
					finishMetricsWait()
				}()
			}
			volumeSize, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("volume size needs to be an integer; instead got %v", args[1])
			}
			manifest := &bytes.Buffer{}
			opts.BlockCacheSize = "0G" // GCS is fast so we want to disable the block cache. See issue #1650
			if err = assets.WriteGoogleAssets(manifest, opts, args[0], volumeSize); err != nil {
				return err
			}
			return maybeKcCreate(dryRun, manifest, opts)
		}),
	}

	deployCustom := &cobra.Command{
		Use:   "custom --persistent-disk <persistent disk backend> --object-store <object store backend> <persistent disk args> <object store args>",
		Short: "(in progress) Deploy a custom Pachyderm cluster configuration",
		Long: "(in progress) Deploy a custom Pachyderm cluster configuration.\n" +
			"If <object store backend> is \"s3\", then the arguments are:\n" +
			"    <volumes> <size of volumes (in GB)> <bucket> <id> <secret> <endpoint>\n",
		Run: pkgcobra.RunBoundedArgs(pkgcobra.Bounds{Min: 4, Max: 7}, func(args []string) (retErr error) {
			if metrics && !dev {
				start := time.Now()
				startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
				defer startMetricsWait()
				defer func() {
					finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
					finishMetricsWait()
				}()
			}
			manifest := &bytes.Buffer{}
			err := assets.WriteCustomAssets(manifest, opts, args, objectStoreBackend, persistentDiskBackend, secure)
			if err != nil {
				return err
			}
			return maybeKcCreate(dryRun, manifest, opts)
		}),
	}
	deployCustom.Flags().BoolVarP(&secure, "secure", "s", false, "Enable secure access to a Minio server.")
	deployCustom.Flags().StringVar(&persistentDiskBackend, "persistent-disk", "aws",
		"(required) Backend providing persistent local volumes to stateful pods. "+
			"One of: aws, google, or azure.")
	deployCustom.Flags().StringVar(&objectStoreBackend, "object-store", "s3",
		"(required) Backend providing an object-storage API to pachyderm. One of: "+
			"s3, gcs, or azure-blob.")
	var cloudfrontDistribution string
	deployAmazon := &cobra.Command{
		Use:   "amazon <S3 bucket> <id> <secret> <token> <region> <size of volumes (in GB)>",
		Short: "Deploy a Pachyderm cluster running on AWS.",
		Long: "Deploy a Pachyderm cluster running on AWS. Arguments are:\n" +
			"  <S3 bucket>: An S3 bucket where Pachyderm will store PFS data.\n" +
			"  <id>, <secret>, <token>: Session token details, used for authorization. You can get these by running 'aws sts get-session-token'\n" +
			"  <region>: The aws region where pachyderm is being deployed (e.g. us-west-1)\n" +
			"  <size of volumes>: Size of EBS volumes, in GB (assumed to all be the same).\n",
		Run: cmdutil.RunFixedArgs(6, func(args []string) (retErr error) {
			if metrics && !dev {
				start := time.Now()
				startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
				defer startMetricsWait()
				defer func() {
					finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
					finishMetricsWait()
				}()
			}
			volumeSize, err := strconv.Atoi(args[5])
			if err != nil {
				return fmt.Errorf("volume size needs to be an integer; instead got %v", args[5])
			}
			if strings.TrimSpace(cloudfrontDistribution) != "" {
				fmt.Printf("WARNING: You specified a cloudfront distribution. Deploying on AWS with cloudfront is currently " +
					"an alpha feature. No security restrictions have been applied to cloudfront, making all data public (obscured but not secured)\n")
			}
			manifest := &bytes.Buffer{}
			if err = assets.WriteAmazonAssets(manifest, opts, args[0], args[1], args[2], args[3], args[4], volumeSize, cloudfrontDistribution); err != nil {
				return err
			}
			return maybeKcCreate(dryRun, manifest, opts)
		}),
	}
	deployAmazon.Flags().StringVar(&cloudfrontDistribution, "cloudfront-distribution", "",
		"Deploying on AWS with cloudfront is currently "+
			"an alpha feature. No security restrictions have been"+
			"applied to cloudfront, making all data public (obscured but not secured)")

	deployMicrosoft := &cobra.Command{
		Use:   "microsoft <container> <storage account name> <storage account key> <size of volumes (in GB)>",
		Short: "Deploy a Pachyderm cluster running on Microsoft Azure.",
		Long: "Deploy a Pachyderm cluster running on Microsoft Azure. Arguments are:\n" +
			"  <container>: An Azure container where Pachyderm will store PFS data.\n" +
			"  <size of volumes>: Size of persistent volumes, in GB (assumed to all be the same).\n",
		Run: cmdutil.RunFixedArgs(4, func(args []string) (retErr error) {
			if metrics && !dev {
				start := time.Now()
				startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
				defer startMetricsWait()
				defer func() {
					finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
					finishMetricsWait()
				}()
			}
			if _, err := base64.StdEncoding.DecodeString(args[2]); err != nil {
				return fmt.Errorf("storage-account-key needs to be base64 encoded; instead got '%v'", args[2])
			}
			if opts.EtcdVolume != "" {
				tempURI, err := url.ParseRequestURI(opts.EtcdVolume)
				if err != nil {
					return fmt.Errorf("Volume URI needs to be a well-formed URI; instead got '%v'", opts.EtcdVolume)
				}
				opts.EtcdVolume = tempURI.String()
			}
			volumeSize, err := strconv.Atoi(args[3])
			if err != nil {
				return fmt.Errorf("volume size needs to be an integer; instead got %v", args[3])
			}
			manifest := &bytes.Buffer{}
			if err = assets.WriteMicrosoftAssets(manifest, opts, args[0], args[1], args[2], volumeSize); err != nil {
				return err
			}
			return maybeKcCreate(dryRun, manifest, opts)
		}),
	}

	deploy := &cobra.Command{
		Use:   "deploy amazon|google|microsoft|local|custom",
		Short: "Deploy a Pachyderm cluster.",
		Long:  "Deploy a Pachyderm cluster.",
		PersistentPreRun: cmdutil.Run(func([]string) error {
			opts = &assets.AssetOpts{
				PachdShards:             uint64(pachdShards),
				Version:                 version.PrettyPrintVersion(version.Version),
				LogLevel:                logLevel,
				Metrics:                 metrics,
				PachdCPURequest:         pachdCPURequest,
				PachdNonCacheMemRequest: pachdNonCacheMemRequest,
				BlockCacheSize:          blockCacheSize,
				EtcdCPURequest:          etcdCPURequest,
				EtcdMemRequest:          etcdMemRequest,
				EtcdNodes:               etcdNodes,
				EtcdVolume:              etcdVolume,
				EnableDash:              enableDash,
				DashOnly:                dashOnly,
				DashImage:               dashImage,
			}
			return nil
		}),
	}
	deploy.PersistentFlags().IntVar(&pachdShards, "shards", 16, "(rarely set) The maximum number of pachd nodes allowed in the cluster; increasing this number blindly can result in degraded performance.")
	deploy.PersistentFlags().IntVar(&etcdNodes, "dynamic-etcd-nodes", 0, "Deploy etcd as a StatefulSet with the given number of pods.  The persistent volumes used by these pods are provisioned dynamically.  Note that StatefulSet is currently a beta kubernetes feature, which might be unavailable in older versions of kubernetes.")
	deploy.PersistentFlags().StringVar(&etcdVolume, "static-etcd-volume", "", "Deploy etcd as a ReplicationController with one pod.  The pod uses the given persistent volume.")
	deploy.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Don't actually deploy pachyderm to Kubernetes, instead just print the manifest.")
	deploy.PersistentFlags().StringVar(&logLevel, "log-level", "info", "The level of log messages to print options are, from least to most verbose: \"error\", \"info\", \"debug\".")
	deploy.PersistentFlags().BoolVar(&enableDash, "dashboard", false, "Deploy the Pachyderm UI along with Pachyderm (experimental). After deployment, run \"pachctl port-forward\" to connect")
	deploy.PersistentFlags().BoolVar(&dashOnly, "dashboard-only", false, "Only deploy the Pachyderm UI (experimental), without the rest of pachyderm. This is for launching the UI adjacent to an existing Pachyderm cluster. After deployment, run \"pachctl port-forward\" to connect")
	deploy.PersistentFlags().StringVar(&dashImage, "dash-image", defaultDashImage, "Image URL for pachyderm dashboard")
	deploy.AddCommand(deployLocal)
	deploy.AddCommand(deployAmazon)
	deploy.AddCommand(deployGoogle)
	deploy.AddCommand(deployMicrosoft)
	deploy.AddCommand(deployCustom)

	// Flags for setting pachd and rethink resource requests. These should rarely
	// be set -- only if we get the defaults wrong, or users have an unusual
	// access pattern.
	//
	// All of these are empty by default, because the actual default values depend
	// on the backend to which we're. The defaults are set in
	// s/s/pkg/deploy/assets/assets.go
	deploy.PersistentFlags().StringVar(&pachdCPURequest,
		"pachd-cpu-request", "", "(rarely set) The size of Pachd's CPU "+
			"request, which we give to Kubernetes. Size is in cores (with partial "+
			"cores allowed and encouraged).")
	deploy.PersistentFlags().StringVar(&blockCacheSize, "block-cache-size", "",
		"Size of pachd's in-memory cache for PFS files. Size is specified in "+
			"bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc).")
	deploy.PersistentFlags().StringVar(&pachdNonCacheMemRequest,
		"pachd-memory-request", "", "(rarely set) The size of PachD's memory "+
			"request in addition to its block cache (set via --block-cache-size). "+
			"Size is in bytes, with SI suffixes (M, K, G, Mi, Ki, Gi, etc).")
	deploy.PersistentFlags().StringVar(&etcdCPURequest,
		"etcd-cpu-request", "", "(rarely set) The size of etcd's CPU request, "+
			"which we give to Kubernetes. Size is in cores (with partial cores "+
			"allowed and encouraged).")
	deploy.PersistentFlags().StringVar(&etcdMemRequest,
		"etcd-memory-request", "", "(rarely set) The size of etcd's memory "+
			"request. Size is in bytes, with SI suffixes (M, K, G, Mi, Ki, Gi, "+
			"etc).")
	return deploy
}

// Cmds returns a list of cobra commands for deploying Pachyderm clusters.
func Cmds(noMetrics *bool) []*cobra.Command {
	deploy := DeployCmd(noMetrics)
	var all bool
	undeploy := &cobra.Command{
		Use:   "undeploy",
		Short: "Tear down a deployed Pachyderm cluster.",
		Long:  "Tear down a deployed Pachyderm cluster.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			if all {
				fmt.Printf(`
By using the --all flag, you are going to delete everything, including the
persistent volumes where metadata is stored.  If your persistent volumes
were dynamically provisioned (i.e. if you used the "--dynamic-etcd-nodes"
flag), the underlying volumes will be removed, making metadata such repos,
commits, pipelines, and jobs unrecoverable. If your persistent volume was
manually provisioned (i.e. if you used the "--static-etcd-volume" flag), the
underlying volume will not be removed.

Are you sure you want to proceed? yN
`)
				r := bufio.NewReader(os.Stdin)
				bytes, err := r.ReadBytes('\n')
				if err != nil {
					return err
				}
				if !(bytes[0] == 'y' || bytes[0] == 'Y') {
					return nil
				}
			}
			io := cmdutil.IO{
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}
			if err := cmdutil.RunIO(io, "kubectl", "delete", "job", "-l", "suite=pachyderm"); err != nil {
				return err
			}
			if err := cmdutil.RunIO(io, "kubectl", "delete", "all", "-l", "suite=pachyderm"); err != nil {
				return err
			}
			if err := cmdutil.RunIO(io, "kubectl", "delete", "sa", "-l", "suite=pachyderm"); err != nil {
				return err
			}
			if err := cmdutil.RunIO(io, "kubectl", "delete", "secret", "-l", "suite=pachyderm"); err != nil {
				return err
			}
			if all {
				if err := cmdutil.RunIO(io, "kubectl", "delete", "storageclass", "-l", "suite=pachyderm"); err != nil {
					return err
				}
				if err := cmdutil.RunIO(io, "kubectl", "delete", "pvc", "-l", "suite=pachyderm"); err != nil {
					return err
				}
				if err := cmdutil.RunIO(io, "kubectl", "delete", "pv", "-l", "suite=pachyderm"); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	undeploy.Flags().BoolVarP(&all, "all", "a", false, `
Delete everything, including the persistent volumes where metadata
is stored.  If your persistent volumes were dynamically provisioned (i.e. if
you used the "--dynamic-etcd-nodes" flag), the underlying volumes will be
removed, making metadata such repos, commits, pipelines, and jobs
unrecoverable. If your persistent volume was manually provisioned (i.e. if
you used the "--static-etcd-volume" flag), the underlying volume will not be
removed.`)
	return []*cobra.Command{deploy, undeploy}
}
