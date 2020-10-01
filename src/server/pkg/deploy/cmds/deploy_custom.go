package cmds

import (
	"bytes"
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	_metrics "github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/spf13/cobra"
)

func CreateDeployCustomCmd(dArgs DeployCmdArgs, s3Flags *S3Flags) *cobra.Command {
	var objectStoreBackend string
	var persistentDiskBackend string
	var secure bool
	var isS3V2 bool
	deployCustom := &cobra.Command{
		Use:   "{{alias}} --persistent-disk <persistent disk backend> --object-store <object store backend> <persistent disk args> <object store args>",
		Short: "Deploy a custom Pachyderm cluster configuration",
		Long: `Deploy a custom Pachyderm cluster configuration.
If <object store backend> is \"s3\", then the arguments are:
    <volumes> <size of volumes (in GB)> <bucket> <id> <secret> <endpoint>`,
		PreRun: dArgs.preRun,
		Run: cmdutil.RunBoundedArgs(4, 7, func(args []string) (retErr error) {
			start := time.Now()
			startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
			defer startMetricsWait()
			defer func() {
				finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
				finishMetricsWait()
			}()
			// Setup advanced configuration.
			advancedConfig := &obj.AmazonAdvancedConfiguration{
				Retries:        s3Flags.Retries,
				Timeout:        s3Flags.Timeout,
				UploadACL:      s3Flags.UploadACL,
				Reverse:        s3Flags.Reverse,
				PartSize:       s3Flags.PartSize,
				MaxUploadParts: s3Flags.MaxUploadParts,
				DisableSSL:     s3Flags.DisableSSL,
				NoVerifySSL:    s3Flags.NoVerifySSL,
			}
			if isS3V2 {
				fmt.Printf("DEPRECATED: Support for the S3V2 option is being deprecated. It will be removed in a future version\n\n")
			}
			// Generate manifest and write assets.
			var buf bytes.Buffer
			if err := assets.WriteCustomAssets(
				encoder(dArgs.globalFlags.OutputFormat, &buf), dArgs.opts, args, objectStoreBackend,
				persistentDiskBackend, secure, isS3V2, advancedConfig,
			); err != nil {
				return err
			}
			if err := kubectlCreate(dArgs.globalFlags.DryRun, buf.Bytes(), dArgs.opts); err != nil {
				return err
			}
			if !dArgs.globalFlags.DryRun || dArgs.contextFlags.CreateContext {
				if dArgs.contextFlags.ContextName == "" {
					dArgs.contextFlags.ContextName = "custom"
				}
				if err := contextCreate(dArgs.contextFlags.ContextName, dArgs.globalFlags.Namespace, dArgs.globalFlags.ServerCert); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	appendGlobalFlags(deployCustom, dArgs.globalFlags)
	appendS3Flags(deployCustom, s3Flags)
	appendContextFlags(deployCustom, dArgs.contextFlags)
	// (bryce) secure should be merged with disableSSL, but it would be a breaking change.
	deployCustom.Flags().BoolVarP(&secure, "secure", "s", false, "Enable secure access to a Minio server.")
	deployCustom.Flags().StringVar(&persistentDiskBackend, "persistent-disk", "aws",
		"(required) Backend providing persistent local volumes to stateful pods. "+
			"One of: aws, google, or azure.")
	deployCustom.Flags().StringVar(&objectStoreBackend, "object-store", "s3",
		"(required) Backend providing an object-storage API to pachyderm. One of: "+
			"s3, gcs, or azure-blob.")
	deployCustom.Flags().BoolVar(&isS3V2, "isS3V2", false, "Enable S3V2 client (DEPRECATED)")
	return deployCustom
}
