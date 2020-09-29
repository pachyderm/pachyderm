package cmds

import (
	"bytes"
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/spf13/cobra"
)

const (
	defaultDashImage   = "pachyderm/dash"
	defaultDashVersion = "0.5.48"
)

func CreateUpdateDashCmd() *cobra.Command {
	var updateDashDryRun bool
	var updateDashOutputFormat string
	updateDash := &cobra.Command{
		Short: "Update and redeploy the Pachyderm Dashboard at the latest compatible version.",
		Long:  "Update and redeploy the Pachyderm Dashboard at the latest compatible version.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			cfg, err := config.Read(false)
			if err != nil {
				return err
			}
			_, activeContext, err := cfg.ActiveContext(false)
			if err != nil {
				return err
			}

			// Undeploy the dash
			if !updateDashDryRun {
				if err := kubectl(nil, activeContext, "delete", "deploy", "-l", "suite=pachyderm,app=dash"); err != nil {
					return err
				}
				if err := kubectl(nil, activeContext, "delete", "svc", "-l", "suite=pachyderm,app=dash"); err != nil {
					return err
				}
			}

			// Redeploy the dash
			var buf bytes.Buffer
			opts := &assets.AssetOpts{
				DashOnly:  true,
				DashImage: fmt.Sprintf("%s:%s", defaultDashImage, getCompatibleVersion("dash", "", defaultDashVersion)),
			}
			if err := assets.WriteDashboardAssets(
				encoder(updateDashOutputFormat, &buf), opts,
			); err != nil {
				return err
			}
			return kubectlCreate(updateDashDryRun, buf.Bytes(), opts)
		}),
	}
	updateDash.Flags().BoolVar(&updateDashDryRun, "dry-run", false, "Don't actually deploy Pachyderm Dash to Kubernetes, instead just print the manifest.")
	updateDash.Flags().StringVarP(&updateDashOutputFormat, "output", "o", "json", "Output format. One of: json|yaml")
	return updateDash
}
