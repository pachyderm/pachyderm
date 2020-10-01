package cmds

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	_metrics "github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/spf13/cobra"
)

func CreateDeployGoogleCmd(dArgs DeployCmdArgs) *cobra.Command {
	deployGoogle := &cobra.Command{
		Use:   "{{alias}} <bucket-name> <disk-size> [<credentials-file>]",
		Short: "Deploy a Pachyderm cluster running on Google Cloud Platform.",
		Long: `Deploy a Pachyderm cluster running on Google Cloud Platform.
  <bucket-name>: A Google Cloud Storage bucket where Pachyderm will store PFS data.
  <disk-size>: Size of Google Compute Engine persistent disks in GB (assumed to all be the same).
  <credentials-file>: A file containing the private key for the account (downloaded from Google Compute Engine).`,
		PreRun: dArgs.preRun,
		Run: cmdutil.RunBoundedArgs(2, 3, func(args []string) (retErr error) {
			start := time.Now()
			startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
			defer startMetricsWait()
			defer func() {
				finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
				finishMetricsWait()
			}()
			volumeSize, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Errorf("volume size needs to be an integer; instead got %v", args[1])
			}
			var buf bytes.Buffer
			dArgs.opts.BlockCacheSize = "0G" // GCS is fast so we want to disable the block cache. See issue #1650
			var cred string
			if len(args) == 3 {
				credBytes, err := ioutil.ReadFile(args[2])
				if err != nil {
					return errors.Wrapf(err, "error reading creds file %s", args[2])
				}
				cred = string(credBytes)
			}
			bucket := strings.TrimPrefix(args[0], "gs://")
			if err = assets.WriteGoogleAssets(
				encoder(dArgs.globalFlags.OutputFormat, &buf), dArgs.opts, bucket, cred, volumeSize,
			); err != nil {
				return err
			}
			if err := kubectlCreate(dArgs.globalFlags.DryRun, buf.Bytes(), dArgs.opts); err != nil {
				return err
			}
			if !dArgs.globalFlags.DryRun || dArgs.contextFlags.CreateContext {
				if dArgs.contextFlags.ContextName == "" {
					dArgs.contextFlags.ContextName = "gcs"
				}
				if err := contextCreate(dArgs.contextFlags.ContextName, dArgs.globalFlags.Namespace, dArgs.globalFlags.ServerCert); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	appendGlobalFlags(deployGoogle, dArgs.globalFlags)
	appendContextFlags(deployGoogle, dArgs.contextFlags)
	return deployGoogle
}
