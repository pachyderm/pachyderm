package cmds

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/spf13/cobra"
)

func CreateDeployStorageAzureCmd(dArgs *DeployCmdArgs) *cobra.Command {
	deployStorageAzure := &cobra.Command{
		Use:    "{{alias}} <account-name> <account-key>",
		Short:  "Deploy credentials for the Azure storage provider.",
		Long:   "Deploy credentials for the Azure storage provider, so that Pachyderm can ingress data from and egress data to it.",
		PreRun: dArgs.preRun,
		Run: cmdutil.RunFixedArgs(2, func(args []string) error {
			return DeployStorageSecrets(assets.MicrosoftSecret("", args[0], args[1]), dArgs.globalFlags.DryRun, dArgs.globalFlags.OutputFormat, dArgs.opts)
		}),
	}
	appendGlobalFlags(deployStorageAzure, dArgs.globalFlags)
	return deployStorageAzure
}
