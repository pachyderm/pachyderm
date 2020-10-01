package cmds

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/spf13/cobra"
)

func CreateDeployStorageAmazonCmd(dArgs *DeployCmdArgs, s3Flags *S3Flags) *cobra.Command {
	deployStorageAmazon := &cobra.Command{
		Use:    "{{alias}} <region> <access-key-id> <secret-access-key> [<session-token>]",
		Short:  "Deploy credentials for the Amazon S3 storage provider.",
		Long:   "Deploy credentials for the Amazon S3 storage provider, so that Pachyderm can ingress data from and egress data to it.",
		PreRun: dArgs.preRun,
		Run: cmdutil.RunBoundedArgs(3, 4, func(args []string) error {
			var token string
			if len(args) == 4 {
				token = args[3]
			}
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
			return DeployStorageSecrets(assets.AmazonSecret(args[0], "", args[1], args[2], token, "", "", advancedConfig), dArgs.globalFlags.DryRun, dArgs.globalFlags.OutputFormat, dArgs.opts)
		}),
	}
	appendGlobalFlags(deployStorageAmazon, dArgs.globalFlags)
	appendS3Flags(deployStorageAmazon, s3Flags)
	return deployStorageAmazon
}
