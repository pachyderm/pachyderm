package cmds

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type PreRun func(cmd *cobra.Command, args []string)

type DeployCmdArgs struct {
	preRun       PreRun
	opts         *assets.AssetOpts
	globalFlags  *GlobalFlags
	contextFlags *ContextFlags
}

type GlobalFlags struct {
	DryRun                     bool
	OutputFormat               string
	Namespace                  string
	ServerCert                 string
	BlockCacheSize             string
	DashImage                  string
	DashOnly                   bool
	EtcdCPURequest             string
	EtcdMemRequest             string
	EtcdNodes                  int
	EtcdStorageClassName       string
	EtcdVolume                 string
	ExposeObjectAPI            bool
	ImagePullSecret            string
	LocalRoles                 bool
	LogLevel                   string
	StorageV2                  bool
	NoDash                     bool
	NoExposeDockerSocket       bool
	NoGuaranteed               bool
	NoRBAC                     bool
	PachdCPURequest            string
	PachdNonCacheMemRequest    string
	PachdShards                int
	Registry                   string
	TLSCertKey                 string
	UploadConcurrencyLimit     int
	PutFileConcurrencyLimit    int
	ClusterDeploymentID        string
	RequireCriticalServersOnly bool
	WorkerServiceAccountName   string
}

type S3Flags struct {
	Retries        int
	Timeout        string
	UploadACL      string
	Reverse        bool
	PartSize       int64
	MaxUploadParts int
	DisableSSL     bool
	NoVerifySSL    bool
}

type ContextFlags struct {
	ContextName   string
	CreateContext bool
}

func appendGlobalFlags(cmd *cobra.Command, flags *GlobalFlags) {
	cmd.Flags().IntVar(&flags.PachdShards, "shards", 16, "(rarely set) The maximum number of pachd nodes allowed in the cluster; increasing this number blindly can result in degraded performance.")
	cmd.Flags().IntVar(&flags.EtcdNodes, "dynamic-etcd-nodes", 0, "Deploy etcd as a StatefulSet with the given number of pods.  The persistent volumes used by these pods are provisioned dynamically.  Note that StatefulSet is currently a beta kubernetes feature, which might be unavailable in older versions of kubernetes.")
	cmd.Flags().StringVar(&flags.EtcdVolume, "static-etcd-volume", "", "Deploy etcd as a ReplicationController with one pod.  The pod uses the given persistent volume.")
	cmd.Flags().StringVar(&flags.EtcdStorageClassName, "etcd-storage-class", "", "If set, the name of an existing StorageClass to use for etcd storage. Ignored if --static-etcd-volume is set.")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Don't actually deploy pachyderm to Kubernetes, instead just print the manifest. Note that a pachyderm context will not be created, unless you also use `--create-context`.")
	cmd.Flags().StringVarP(&flags.OutputFormat, "output", "o", "json", "Output format. One of: json|yaml")
	cmd.Flags().StringVar(&flags.LogLevel, "log-level", "info", "The level of log messages to print options are, from least to most verbose: \"error\", \"info\", \"debug\".")
	cmd.Flags().BoolVar(&flags.DashOnly, "dashboard-only", false, "Only deploy the Pachyderm UI (experimental), without the rest of pachyderm. This is for launching the UI adjacent to an existing Pachyderm cluster. After deployment, run \"pachctl port-forward\" to connect")
	cmd.Flags().BoolVar(&flags.NoDash, "no-dashboard", false, "Don't deploy the Pachyderm UI alongside Pachyderm (experimental).")
	cmd.Flags().StringVar(&flags.Registry, "registry", "", "The registry to pull images from.")
	cmd.Flags().StringVar(&flags.ImagePullSecret, "image-pull-secret", "", "A secret in Kubernetes that's needed to pull from your private registry.")
	cmd.Flags().StringVar(&flags.DashImage, "dash-image", "", "Image URL for pachyderm dashboard")
	cmd.Flags().BoolVar(&flags.NoGuaranteed, "no-guaranteed", false, "Don't use guaranteed QoS for etcd and pachd deployments. Turning this on (turning guaranteed QoS off) can lead to more stable local clusters (such as on Minikube), it should normally be used for production clusters.")
	cmd.Flags().BoolVar(&flags.NoRBAC, "no-rbac", false, "Don't deploy RBAC roles for Pachyderm. (for k8s versions prior to 1.8)")
	cmd.Flags().BoolVar(&flags.LocalRoles, "local-roles", false, "Use namespace-local roles instead of cluster roles. Ignored if --no-rbac is set.")
	cmd.Flags().StringVar(&flags.Namespace, "namespace", "", "Kubernetes namespace to deploy Pachyderm to.")
	cmd.Flags().BoolVar(&flags.NoExposeDockerSocket, "no-expose-docker-socket", false, "Don't expose the Docker socket to worker containers. This limits the privileges of workers which prevents them from automatically setting the container's working dir and user.")
	cmd.Flags().BoolVar(&flags.ExposeObjectAPI, "expose-object-api", false, "If set, instruct pachd to serve its object/block API on its public port (not safe with auth enabled, do not set in production).")
	cmd.Flags().StringVar(&flags.TLSCertKey, "tls", "", "string of the form \"<cert path>,<key path>\" of the signed TLS certificate and private key that Pachd should use for TLS authentication (enables TLS-encrypted communication with Pachd)")
	cmd.Flags().BoolVar(&flags.StorageV2, "storage-v2", false, "Deploy Pachyderm using V2 storage (alpha)")
	cmd.Flags().IntVar(&flags.UploadConcurrencyLimit, "upload-concurrency-limit", assets.DefaultUploadConcurrencyLimit, "The maximum number of concurrent object storage uploads per Pachd instance.")
	cmd.Flags().IntVar(&flags.PutFileConcurrencyLimit, "put-file-concurrency-limit", assets.DefaultPutFileConcurrencyLimit, "The maximum number of files to upload or fetch from remote sources (HTTP, blob storage) using PutFile concurrently.")
	cmd.Flags().StringVar(&flags.ClusterDeploymentID, "cluster-deployment-id", "", "Set an ID for the cluster deployment. Defaults to a random value.")
	cmd.Flags().BoolVar(&flags.RequireCriticalServersOnly, "require-critical-servers-only", assets.DefaultRequireCriticalServersOnly, "Only require the critical Pachd servers to startup and run without errors.")
	cmd.Flags().StringVar(&flags.WorkerServiceAccountName, "worker-service-account", assets.DefaultWorkerServiceAccountName, "The Kubernetes service account for workers to use when creating S3 gateways.")

	// Flags for setting pachd resource requests. These should rarely be set --
	// only if we get the defaults wrong, or users have an unusual access pattern
	//
	// All of these are empty by default, because the actual default values depend
	// on the backend to which we're. The defaults are set in
	// s/s/pkg/deploy/assets/assets.go
	cmd.Flags().StringVar(&flags.PachdCPURequest,
		"pachd-cpu-request", "", "(rarely set) The size of Pachd's CPU "+
			"request, which we give to Kubernetes. Size is in cores (with partial "+
			"cores allowed and encouraged).")
	cmd.Flags().StringVar(&flags.BlockCacheSize, "block-cache-size", "",
		"Size of pachd's in-memory cache for PFS files. Size is specified in "+
			"bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc).")
	cmd.Flags().StringVar(&flags.PachdNonCacheMemRequest,
		"pachd-memory-request", "", "(rarely set) The size of PachD's memory "+
			"request in addition to its block cache (set via --block-cache-size). "+
			"Size is in bytes, with SI suffixes (M, K, G, Mi, Ki, Gi, etc).")
	cmd.Flags().StringVar(&flags.EtcdCPURequest,
		"etcd-cpu-request", "", "(rarely set) The size of etcd's CPU request, "+
			"which we give to Kubernetes. Size is in cores (with partial cores "+
			"allowed and encouraged).")
	cmd.Flags().StringVar(&flags.EtcdMemRequest,
		"etcd-memory-request", "", "(rarely set) The size of etcd's memory "+
			"request. Size is in bytes, with SI suffixes (M, K, G, Mi, Ki, Gi, "+
			"etc).")
}

func AppendS3Flags(cmd *cobra.Command, flags *S3Flags) {
	cmd.Flags().IntVar(&flags.Retries, "retries", obj.DefaultRetries, "(rarely set) Set a custom number of retries for object storage requests.")
	cmd.Flags().StringVar(&flags.Timeout, "timeout", obj.DefaultTimeout, "(rarely set) Set a custom timeout for object storage requests.")
	cmd.Flags().StringVar(&flags.UploadACL, "upload-acl", obj.DefaultUploadACL, "(rarely set) Set a custom upload ACL for object storage uploads.")
	cmd.Flags().BoolVar(&flags.Reverse, "reverse", obj.DefaultReverse, "(rarely set) Reverse object storage paths.")
	cmd.Flags().Int64Var(&flags.PartSize, "part-size", obj.DefaultPartSize, "(rarely set) Set a custom part size for object storage uploads.")
	cmd.Flags().IntVar(&flags.MaxUploadParts, "max-upload-parts", obj.DefaultMaxUploadParts, "(rarely set) Set a custom maximum number of upload parts.")
	cmd.Flags().BoolVar(&flags.DisableSSL, "disable-ssl", obj.DefaultDisableSSL, "(rarely set) Disable SSL.")
	cmd.Flags().BoolVar(&flags.NoVerifySSL, "no-verify-ssl", obj.DefaultNoVerifySSL, "(rarely set) Skip SSL certificate verification (typically used for enabling self-signed certificates).")
}

func appendContextFlags(cmd *cobra.Command, flags *ContextFlags) {
	cmd.Flags().StringVarP(&flags.ContextName, "context", "c", "", "Name of the context to add to the pachyderm config. If unspecified, a context name will automatically be derived.")
	cmd.Flags().BoolVar(&flags.CreateContext, "create-context", false, "Create a context, even with `--dry-run`.")
}

func preRunInternal(args []string, globalFlags *GlobalFlags, opts *assets.AssetOpts) error {
	cfg, err := config.Read(false)
	if err != nil {
		log.Warningf("could not read config to check whether cluster metrics "+
			"will be enabled: %v.\n", err)
	}

	if globalFlags.Namespace == "" {
		kubeConfig := config.KubeConfig(nil)
		var err error
		globalFlags.Namespace, _, err = kubeConfig.Namespace()
		if err != nil {
			log.Warningf("using namespace \"default\" (couldn't load namespace "+
				"from kubernetes config: %v)\n", err)
			globalFlags.Namespace = "default"
		}
	}

	if globalFlags.DashImage == "" {
		globalFlags.DashImage = fmt.Sprintf("%s:%s", defaultDashImage, getCompatibleVersion("dash", "", defaultDashVersion))
	}

	opts = &assets.AssetOpts{
		FeatureFlags: assets.FeatureFlags{
			StorageV2: globalFlags.StorageV2,
		},
		StorageOpts: assets.StorageOpts{
			UploadConcurrencyLimit:  globalFlags.UploadConcurrencyLimit,
			PutFileConcurrencyLimit: globalFlags.PutFileConcurrencyLimit,
		},
		PachdShards:                uint64(globalFlags.PachdShards),
		Version:                    version.PrettyPrintVersion(version.Version),
		LogLevel:                   globalFlags.LogLevel,
		Metrics:                    cfg == nil || cfg.V2.Metrics,
		PachdCPURequest:            globalFlags.PachdCPURequest,
		PachdNonCacheMemRequest:    globalFlags.PachdNonCacheMemRequest,
		BlockCacheSize:             globalFlags.BlockCacheSize,
		EtcdCPURequest:             globalFlags.EtcdCPURequest,
		EtcdMemRequest:             globalFlags.EtcdMemRequest,
		EtcdNodes:                  globalFlags.EtcdNodes,
		EtcdVolume:                 globalFlags.EtcdVolume,
		EtcdStorageClassName:       globalFlags.EtcdStorageClassName,
		DashOnly:                   globalFlags.DashOnly,
		NoDash:                     globalFlags.NoDash,
		DashImage:                  globalFlags.DashImage,
		Registry:                   globalFlags.Registry,
		ImagePullSecret:            globalFlags.ImagePullSecret,
		NoGuaranteed:               globalFlags.NoGuaranteed,
		NoRBAC:                     globalFlags.NoRBAC,
		LocalRoles:                 globalFlags.LocalRoles,
		Namespace:                  globalFlags.Namespace,
		NoExposeDockerSocket:       globalFlags.NoExposeDockerSocket,
		ExposeObjectAPI:            globalFlags.ExposeObjectAPI,
		ClusterDeploymentID:        globalFlags.ClusterDeploymentID,
		RequireCriticalServersOnly: globalFlags.RequireCriticalServersOnly,
		WorkerServiceAccountName:   globalFlags.WorkerServiceAccountName,
	}
	if globalFlags.TLSCertKey != "" {
		// TODO(msteffen): If either the cert path or the key path contains a
		// comma, this doesn't work
		certKey := strings.Split(globalFlags.TLSCertKey, ",")
		if len(certKey) != 2 {
			return fmt.Errorf("could not split TLS certificate and key correctly; must have two parts but got: %#v", certKey)
		}
		opts.TLS = &assets.TLSOpts{
			ServerCert: certKey[0],
			ServerKey:  certKey[1],
		}

		serverCertBytes, err := ioutil.ReadFile(certKey[0])
		if err != nil {
			return errors.Wrapf(err, "could not read server cert at %q", certKey[0])
		}
		globalFlags.ServerCert = base64.StdEncoding.EncodeToString([]byte(serverCertBytes))
	}
	return nil
}

func standardDeployCmds() []*cobra.Command {
	var commands []*cobra.Command
	var opts *assets.AssetOpts

	globalFlags := new(GlobalFlags)

	s3Flags := new(S3Flags)

	contextFlags := new(ContextFlags)

	preRun := cmdutil.Run(func(args []string) error {
		return preRunInternal(args, globalFlags, opts)
	})

	deployPreRun := cmdutil.Run(func(args []string) error {
		if version.IsUnstable() {
			fmt.Printf("WARNING: The version of Pachyderm you are deploying (%s) is an unstable pre-release build and may not support data migration.\n\n", version.PrettyVersion())

			if ok, err := cmdutil.InteractiveConfirm(); err != nil {
				return err
			} else if !ok {
				return errors.New("deploy aborted")
			}
		}
		return preRunInternal(args, globalFlags, opts)
	})

	deployCmdArgs := DeployCmdArgs{
		preRun:       deployPreRun,
		opts:         opts,
		globalFlags:  globalFlags,
		contextFlags: contextFlags,
	}

	commands = append(commands, cmdutil.CreateAlias(CreateDeployLocalCmd(deployCmdArgs), "deploy local"))
	commands = append(commands, cmdutil.CreateAlias(CreateDeployGoogleCmd(deployCmdArgs), "deploy google"))
	commands = append(commands, cmdutil.CreateAlias(CreateDeployCustomCmd(deployCmdArgs, s3Flags), "deploy custom"))
	commands = append(commands, cmdutil.CreateAlias(CreateDeployAmazonCmd(deployCmdArgs, s3Flags), "deploy amazon"))
	commands = append(commands, cmdutil.CreateAlias(CreateDeployMicrosoftCmd(deployCmdArgs), "deploy microsoft"))

	storageCmdArgs := DeployCmdArgs{
		preRun:       preRun,
		opts:         opts,
		globalFlags:  globalFlags,
		contextFlags: contextFlags,
	}

	commands = append(commands, cmdutil.CreateAlias(CreateDeployStorageAmazonCmd(storageCmdArgs, s3Flags), "deploy storage amazon"))
	commands = append(commands, cmdutil.CreateAlias(CreateDeployStorageGoogleCmd(storageCmdArgs), "deploy storage google"))
	commands = append(commands, cmdutil.CreateAlias(CreateDeployStorageAzureCmd(storageCmdArgs), "deploy storage microsoft"))

	deployStorage := &cobra.Command{
		Short: "Deploy credentials for a particular storage provider.",
		Long:  "Deploy credentials for a particular storage provider, so that Pachyderm can ingress data from and egress data to it.",
	}
	commands = append(commands, cmdutil.CreateAlias(deployStorage, "deploy storage"))

	listImages := CreateListImagesCmd(preRun, opts)
	appendGlobalFlags(listImages, globalFlags)
	commands = append(commands, cmdutil.CreateAlias(listImages, "deploy list-images"))

	exportImages := CreateExportImagesCmd(preRun, opts)
	appendGlobalFlags(exportImages, globalFlags)
	commands = append(commands, cmdutil.CreateAlias(exportImages, "deploy export-images"))

	importImages := CreateImportImagesCmd(preRun, opts)
	appendGlobalFlags(importImages, globalFlags)
	commands = append(commands, cmdutil.CreateAlias(importImages, "deploy import-images"))

	return commands
}

// Cmds returns a list of cobra commands for deploying Pachyderm clusters.
func Cmds() []*cobra.Command {
	commands := standardDeployCmds()

	deployIDE := CreateDeployIDECommand()
	commands = append(commands, cmdutil.CreateAlias(deployIDE, "deploy ide"))

	deploy := &cobra.Command{
		Short: "Deploy a Pachyderm cluster.",
		Long:  "Deploy a Pachyderm cluster.",
	}
	commands = append(commands, cmdutil.CreateAlias(deploy, "deploy"))

	undeploy := CreateUndeployCmd()
	commands = append(commands, cmdutil.CreateAlias(undeploy, "undeploy"))

	updateDash := CreateUpdateDashCmd()
	commands = append(commands, cmdutil.CreateAlias(updateDash, "update-dash"))

	return commands
}
