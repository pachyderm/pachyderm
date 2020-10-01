package cmds

import (
	"io/ioutil"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/spf13/cobra"
)

func CreateDeployStorageGoogleCmd(dArgs *DeployCmdArgs) *cobra.Command {
	deployStorageGoogle := &cobra.Command{
		Use:    "{{alias}} <credentials-file>",
		Short:  "Deploy credentials for the Google Cloud storage provider.",
		Long:   "Deploy credentials for the Google Cloud storage provider, so that Pachyderm can ingress data from and egress data to it.",
		PreRun: dArgs.preRun,
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			credBytes, err := ioutil.ReadFile(args[0])
			if err != nil {
				return errors.Wrapf(err, "error reading credentials file %s", args[0])
			}
			return DeployStorageSecrets(assets.GoogleSecret("", string(credBytes)), dArgs.globalFlags.DryRun, dArgs.globalFlags.OutputFormat, dArgs.opts)
		}),
	}
	appendGlobalFlags(deployStorageGoogle, dArgs.globalFlags)
	return deployStorageGoogle
}
