package cmds

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	deployclient "github.com/pachyderm/pachyderm/src/client/deploy"

	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/images"
	_metrics "github.com/pachyderm/pachyderm/src/server/pkg/metrics"

	"github.com/spf13/cobra"
	"go.pedge.io/pkg/cobra"
)

var defaultDashImage = "pachyderm/dash:0.5.10"

func maybeKcCreate(dryRun bool, manifest *bytes.Buffer, opts *assets.AssetOpts, metrics bool) error {
	if dryRun {
		_, err := os.Stdout.Write(manifest.Bytes())
		return err
	}
	io := cmdutil.IO{
		Stdin:  manifest,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	// we set --validate=false due to https://github.com/kubernetes/kubernetes/issues/53309
	if err := cmdutil.RunIO(io, "kubectl", "create", "-f", "-", "--validate=false"); err != nil {
		return err
	}
	if !dryRun {
		fmt.Println("\nPachyderm is launching. Check its status with \"kubectl get all\"")
		if opts.DashOnly || opts.EnableDash {
			fmt.Println("Once launched, access the dashboard by running \"pachctl port-forward\"")
		}
		fmt.Println("")
	}
	return nil
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
	var registry string
	var imagePullSecret string
	var noGuaranteed bool

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
				// Use dev build instead of release build
				opts.Version = deploy.DevVersionTag

				// we turn metrics off this is a dev cluster. The default is set by
				// deploy.PersistentPreRun, below.
				opts.Metrics = false

				// Disable authentication, for tests
				opts.DisableAuthentication = true
			}
			if err := assets.WriteLocalAssets(manifest, opts, hostPath); err != nil {
				return err
			}
			return maybeKcCreate(dryRun, manifest, opts, metrics)
		}),
	}
	deployLocal.Flags().StringVar(&hostPath, "host-path", "/var/pachyderm", "Location on the host machine where PFS metadata will be stored.")
	deployLocal.Flags().BoolVarP(&dev, "dev", "d", false, "Deploy pachd built locally, disable metrics, and use insecure authentication")

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
			return maybeKcCreate(dryRun, manifest, opts, metrics)
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
			return maybeKcCreate(dryRun, manifest, opts, metrics)
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
	var creds string
	var iamRole string
	deployAmazon := &cobra.Command{
		Use:   "amazon <S3 bucket> <region> <size of volumes (in GB)>",
		Short: "Deploy a Pachyderm cluster running on AWS.",
		Long: "Deploy a Pachyderm cluster running on AWS. Arguments are:\n" +
			"  <S3 bucket>: An S3 bucket where Pachyderm will store PFS data.\n" +
			"\n" +
			"  <region>: The aws region where pachyderm is being deployed (e.g. us-west-1)\n" +
			"  <size of volumes>: Size of EBS volumes, in GB (assumed to all be the same).\n",
		Run: cmdutil.RunFixedArgs(3, func(args []string) (retErr error) {
			if metrics && !dev {
				start := time.Now()
				startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
				defer startMetricsWait()
				defer func() {
					finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
					finishMetricsWait()
				}()
			}
			if creds == "" && iamRole == "" {
				return fmt.Errorf("Either the --credentials or the --iam-role flag needs to be provided")
			}
			var id, secret, token string
			if creds != "" {
				parts := strings.Split(creds, ",")
				if len(parts) != 3 {
					return fmt.Errorf("Incorrect format of --credentials")
				}
				id, secret, token = parts[0], parts[1], parts[2]
			}
			opts.IAMRole = iamRole
			volumeSize, err := strconv.Atoi(args[2])
			if err != nil {
				return fmt.Errorf("volume size needs to be an integer; instead got %v", args[2])
			}
			if strings.TrimSpace(cloudfrontDistribution) != "" {
				fmt.Printf("WARNING: You specified a cloudfront distribution. Deploying on AWS with cloudfront is currently " +
					"an alpha feature. No security restrictions have been applied to cloudfront, making all data public (obscured but not secured)\n")
			}
			manifest := &bytes.Buffer{}
			if err = assets.WriteAmazonAssets(manifest, opts, args[0], id, secret, token, args[1], volumeSize, cloudfrontDistribution); err != nil {
				return err
			}
			return maybeKcCreate(dryRun, manifest, opts, metrics)
		}),
	}
	deployAmazon.Flags().StringVar(&cloudfrontDistribution, "cloudfront-distribution", "",
		"Deploying on AWS with cloudfront is currently "+
			"an alpha feature. No security restrictions have been"+
			"applied to cloudfront, making all data public (obscured but not secured)")
	deployAmazon.Flags().StringVar(&creds, "credentials", "", "Use the format '--credentials=<id>,<secret>,<token>'\n<id>, <secret>, and <token> are session token details, used for authorization. You can get these by running 'aws sts get-session-token'")
	deployAmazon.Flags().StringVar(&iamRole, "iam-role", "", "Use the given IAM role for authorization, as opposed to using static credentials.  The nodes on which Pachyderm is deployed needs to have the given IAM role.")

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
			return maybeKcCreate(dryRun, manifest, opts, metrics)
		}),
	}

	deployStorage := &cobra.Command{
		Use:   "storage <backend> ...",
		Short: "Deploy credentials for a particular storage provider.",
		Long: `
Deploy credentials for a particular storage provider, so that Pachyderm can
ingress data from and egress data to it.  Currently three backends are
supported: aws, google, and azure.  To see the required arguments for a
particular backend, run "pachctl deploy storage <backend>"`,
		Run: cmdutil.RunBoundedArgs(1, 5, func(args []string) (retErr error) {
			var data map[string][]byte
			switch args[0] {
			case "aws":
				// Need at least 4 arguments: backend, bucket, id, secret
				if len(args) < 4 {
					return fmt.Errorf("Usage: pachctl deploy storage aws <region> <id> <secret> <token>\n\n<token> is optional")
				}
				var token string
				if len(args) == 5 {
					token = args[4]
				}
				data = assets.AmazonSecret("", "", args[2], args[3], token, args[1])
			case "google":
				return fmt.Errorf("deploying credentials for GCS storage is not currently supported")
			case "azure":
				// Need 3 arguments: backend, account name, account key
				if len(args) != 3 {
					return fmt.Errorf("Usage: pachctl deploy storage azure <account name> <account key>")
				}
				data = assets.MicrosoftSecret("", args[1], args[2])
			}

			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return fmt.Errorf("error constructing pachyderm client: %v", err)
			}

			_, err = c.DeployStorageSecret(context.Background(), &deployclient.DeployStorageSecretRequest{
				Secrets: data,
			})
			if err != nil {
				return fmt.Errorf("error deploying storage secret to pachd: %v", err)
			}
			return nil
		}),
	}

	listImages := &cobra.Command{
		Use:   "list-images",
		Short: "Output the list of images in a deployment.",
		Long:  "Output the list of images in a deployment.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			for _, image := range assets.Images(opts) {
				fmt.Println(image)
			}
			return nil
		}),
	}

	exportImages := &cobra.Command{
		Use:   "export-images output-file",
		Short: "Export a tarball (to stdout) containing all of the images in a deployment.",
		Long:  "Export a tarball (to stdout) containing all of the images in a deployment.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			file, err := os.Create(args[0])
			if err != nil {
				return err
			}
			defer func() {
				if err := file.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			return images.Export(opts, file)
		}),
	}

	importImages := &cobra.Command{
		Use:   "import-images input-file",
		Short: "Import a tarball (from stdin) containing all of the images in a deployment and push them to a private registry.",
		Long:  "Import a tarball (from stdin) containing all of the images in a deployment and push them to a private registry.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			file, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer func() {
				if err := file.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			return images.Import(opts, file)
		}),
	}

	deploy := &cobra.Command{
		Use:   "deploy amazon|google|microsoft|local|custom|storage",
		Short: "Deploy a Pachyderm cluster.",
		Long:  "Deploy a Pachyderm cluster.",
		PersistentPreRun: cmdutil.Run(func([]string) error {
			dashImage = getDefaultOrLatestDashImage(dashImage, dryRun)
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
				Registry:                registry,
				NoGuaranteed:            noGuaranteed,
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
	deploy.PersistentFlags().StringVar(&registry, "registry", "", "The registry to pull images from.")
	deploy.PersistentFlags().StringVar(&imagePullSecret, "image-pull-secret", "", "A secret in Kubernetes that's needed to pull from your private registry.")
	deploy.PersistentFlags().StringVar(&dashImage, "dash-image", "", "Image URL for pachyderm dashboard")
	deploy.PersistentFlags().BoolVar(&noGuaranteed, "no-guaranteed", false, "Don't use guaranteed QoS for etcd and pachd deployments. Turning this on (turning guaranteed QoS off) can lead to more stable local clusters (such as a on Minikube), it should normally be used for production clusters.")

	deploy.AddCommand(
		deployLocal,
		deployAmazon,
		deployGoogle,
		deployMicrosoft,
		deployCustom,
		deployStorage,
		listImages,
		exportImages,
		importImages,
	)

	// Flags for setting pachd resource requests. These should rarely be set --
	// only if we get the defaults wrong, or users have an unusual access pattern
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

	var updateDashDryRun bool
	updateDash := &cobra.Command{
		Use:   "update-dash",
		Short: "Update and redeploy the Pachyderm Dashboard at the latest compatible version.",
		Long:  "Update and redeploy the Pachyderm Dashboard at the latest compatible version.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			// Undeploy the dash
			if !updateDashDryRun {
				io := cmdutil.IO{
					Stdout: os.Stdout,
					Stderr: os.Stderr,
				}
				if err := cmdutil.RunIO(io, "kubectl", "delete", "deploy", "-l", "suite=pachyderm,app=dash"); err != nil {
					return err
				}
				if err := cmdutil.RunIO(io, "kubectl", "delete", "svc", "-l", "suite=pachyderm,app=dash"); err != nil {
					return err
				}
			}
			// Redeploy the dash
			manifest := &bytes.Buffer{}
			dashImage := getDefaultOrLatestDashImage("", updateDashDryRun)
			opts := &assets.AssetOpts{
				DashOnly:  true,
				DashImage: dashImage,
			}
			assets.WriteDashboardAssets(manifest, opts)
			return maybeKcCreate(updateDashDryRun, manifest, opts, false)
		}),
	}
	updateDash.Flags().BoolVar(&updateDashDryRun, "dry-run", false, "Don't actually deploy Pachyderm Dash to Kubernetes, instead just print the manifest.")

	return []*cobra.Command{deploy, undeploy, updateDash}
}

func getDefaultOrLatestDashImage(dashImage string, dryRun bool) string {
	var err error
	version := version.PrettyPrintVersion(version.Version)
	defer func() {
		if err != nil && !dryRun {
			fmt.Printf("Error retrieving latest dash image for pachctl %v: %v Falling back to dash image %v\n", version, err, defaultDashImage)
		}
	}()
	if dashImage != "" {
		// It has been supplied explicitly by version on the command line
		return dashImage
	}
	dashImage = defaultDashImage
	compatibleDashVersionsURL := fmt.Sprintf("https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/compatibility/%v", version)
	resp, err := http.Get(compatibleDashVersionsURL)
	if err != nil {
		return dashImage
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return dashImage
	}
	if resp.StatusCode != 200 {
		err = errors.New(string(body))
		return dashImage
	}
	allVersions := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(allVersions) < 1 {
		return dashImage
	}
	latestVersion := strings.TrimSpace(allVersions[len(allVersions)-1])

	return fmt.Sprintf("pachyderm/dash:%v", latestVersion)
}
