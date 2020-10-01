package cmds

import (
	"bytes"
	"encoding/base64"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	_metrics "github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/spf13/cobra"
)

func CreateDeployMicrosoftCmd(dArgs *DeployCmdArgs) *cobra.Command {
	deployMicrosoft := &cobra.Command{
		Use:   "{{alias}} <container> <account-name> <account-key> <disk-size>",
		Short: "Deploy a Pachyderm cluster running on Microsoft Azure.",
		Long: `Deploy a Pachyderm cluster running on Microsoft Azure.
  <container>: An Azure container where Pachyderm will store PFS data.
  <disk-size>: Size of persistent volumes, in GB (assumed to all be the same).`,
		PreRun: dArgs.preRun,
		Run: cmdutil.RunFixedArgs(4, func(args []string) (retErr error) {
			start := time.Now()
			startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
			defer startMetricsWait()
			defer func() {
				finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
				finishMetricsWait()
			}()
			if _, err := base64.StdEncoding.DecodeString(args[2]); err != nil {
				return errors.Errorf("storage-account-key needs to be base64 encoded; instead got '%v'", args[2])
			}
			if dArgs.opts.EtcdVolume != "" {
				tempURI, err := url.ParseRequestURI(dArgs.opts.EtcdVolume)
				if err != nil {
					return errors.Errorf("volume URI needs to be a well-formed URI; instead got '%v'", dArgs.opts.EtcdVolume)
				}
				dArgs.opts.EtcdVolume = tempURI.String()
			}
			volumeSize, err := strconv.Atoi(args[3])
			if err != nil {
				return errors.Errorf("volume size needs to be an integer; instead got %v", args[3])
			}
			var buf bytes.Buffer
			container := strings.TrimPrefix(args[0], "wasb://")
			accountName, accountKey := args[1], args[2]
			if err = assets.WriteMicrosoftAssets(
				encoder(dArgs.globalFlags.OutputFormat, &buf), dArgs.opts, container, accountName, accountKey, volumeSize,
			); err != nil {
				return err
			}
			if err := kubectlCreate(dArgs.globalFlags.DryRun, buf.Bytes(), dArgs.opts); err != nil {
				return err
			}
			if !dArgs.globalFlags.DryRun || dArgs.contextFlags.CreateContext {
				if dArgs.contextFlags.ContextName == "" {
					dArgs.contextFlags.ContextName = "azure"
				}
				if err := contextCreate(dArgs.contextFlags.ContextName, dArgs.globalFlags.Namespace, dArgs.globalFlags.ServerCert); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	appendGlobalFlags(deployMicrosoft, dArgs.globalFlags)
	appendContextFlags(deployMicrosoft, dArgs.contextFlags)
	return deployMicrosoft
}
