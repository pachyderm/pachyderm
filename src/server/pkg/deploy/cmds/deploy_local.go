package cmds

import (
	"bytes"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	_metrics "github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/spf13/cobra"
)

func CreateDeployLocalCmd(dArgs DeployCmdArgs) *cobra.Command {
	var dev bool
	var hostPath string
	deployLocal := &cobra.Command{
		Short:  "Deploy a single-node Pachyderm cluster with local metadata storage.",
		Long:   "Deploy a single-node Pachyderm cluster with local metadata storage.",
		PreRun: dArgs.preRun,
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			if !dev {
				start := time.Now()
				startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
				defer startMetricsWait()
				defer func() {
					finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
					finishMetricsWait()
				}()
			}
			if dev {
				// Use dev build instead of release build
				dArgs.opts.Version = deploy.DevVersionTag

				// we turn metrics off if this is a dev cluster. The default
				// is set by deploy.PersistentPreRun, below.
				dArgs.opts.Metrics = false

				// Disable authentication, for tests
				dArgs.opts.DisableAuthentication = true

				// Serve the Pachyderm object/block API locally, as this is needed by
				// our tests (and authentication is disabled anyway)
				dArgs.opts.ExposeObjectAPI = true
			}
			var buf bytes.Buffer
			if err := assets.WriteLocalAssets(
				encoder(dArgs.globalFlags.OutputFormat, &buf), dArgs.opts, hostPath,
			); err != nil {
				return err
			}
			if err := kubectlCreate(dArgs.globalFlags.DryRun, buf.Bytes(), dArgs.opts); err != nil {
				return err
			}
			if !dArgs.globalFlags.DryRun || dArgs.contextFlags.CreateContext {
				if dArgs.contextFlags.ContextName == "" {
					dArgs.contextFlags.ContextName = "local"
				}
				if err := contextCreate(dArgs.contextFlags.ContextName, dArgs.globalFlags.Namespace, dArgs.globalFlags.ServerCert); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	AppendGlobalFlags(deployLocal, dArgs.globalFlags)
	appendContextFlags(deployLocal, dArgs.contextFlags)
	deployLocal.Flags().StringVar(&hostPath, "host-path", "/var/pachyderm", "Location on the host machine where PFS metadata will be stored.")
	deployLocal.Flags().BoolVarP(&dev, "dev", "d", false, "Deploy pachd with local version tags, disable metrics, expose Pachyderm's object/block API, and use an insecure authentication mechanism (do not set on any cluster with sensitive data)")
	return deployLocal
}
