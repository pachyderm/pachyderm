package cmds

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	_metrics "github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	awsAccessKeyIDRE = regexp.MustCompile("^[A-Z0-9]{20}$")
	awsSecretRE      = regexp.MustCompile("^[A-Za-z0-9/+=]{40}$")
	awsRegionRE      = regexp.MustCompile("^[a-z]{2}(?:-gov)?-[a-z]+-[0-9]$")
)

func CreateDeployAmazonCmd(dArgs DeployCmdArgs, s3Flags *S3Flags) *cobra.Command {
	var cloudfrontDistribution string
	var creds string
	var iamRole string
	var vault string
	deployAmazon := &cobra.Command{
		Use:   "{{alias}} <bucket-name> <region> <disk-size>",
		Short: "Deploy a Pachyderm cluster running on AWS.",
		Long: `Deploy a Pachyderm cluster running on AWS.
  <bucket-name>: An S3 bucket where Pachyderm will store PFS data.
  <region>: The AWS region where Pachyderm is being deployed (e.g. us-west-1)
  <disk-size>: Size of EBS volumes, in GB (assumed to all be the same).`,
		PreRun: dArgs.preRun,
		Run: cmdutil.RunFixedArgs(3, func(args []string) (retErr error) {
			start := time.Now()
			startMetricsWait := _metrics.StartReportAndFlushUserAction("Deploy", start)
			defer startMetricsWait()
			defer func() {
				finishMetricsWait := _metrics.FinishReportAndFlushUserAction("Deploy", retErr, start)
				finishMetricsWait()
			}()
			if creds == "" && vault == "" && iamRole == "" {
				return errors.Errorf("one of --credentials, --vault, or --iam-role needs to be provided")
			}

			// populate 'amazonCreds' & validate
			var amazonCreds *assets.AmazonCreds
			s := bufio.NewScanner(os.Stdin)
			if creds != "" {
				parts := strings.Split(creds, ",")
				if len(parts) < 2 || len(parts) > 3 || containsEmpty(parts[:2]) {
					return errors.Errorf("incorrect format of --credentials")
				}
				amazonCreds = &assets.AmazonCreds{ID: parts[0], Secret: parts[1]}
				if len(parts) > 2 {
					amazonCreds.Token = parts[2]
				}

				if !awsAccessKeyIDRE.MatchString(amazonCreds.ID) {
					fmt.Fprintf(os.Stderr, "The AWS Access Key seems invalid (does not "+
						"match %q). Do you want to continue deploying? [yN]\n",
						awsAccessKeyIDRE)
					if s.Scan(); s.Text()[0] != 'y' && s.Text()[0] != 'Y' {
						os.Exit(1)
					}
				}

				if !awsSecretRE.MatchString(amazonCreds.Secret) {
					fmt.Fprintf(os.Stderr, "The AWS Secret seems invalid (does not "+
						"match %q). Do you want to continue deploying? [yN]\n", awsSecretRE)
					if s.Scan(); s.Text()[0] != 'y' && s.Text()[0] != 'Y' {
						os.Exit(1)
					}
				}
			}
			if vault != "" {
				if amazonCreds != nil {
					return errors.Errorf("only one of --credentials, --vault, or --iam-role needs to be provided")
				}
				parts := strings.Split(vault, ",")
				if len(parts) != 3 || containsEmpty(parts) {
					return errors.Errorf("incorrect format of --vault")
				}
				amazonCreds = &assets.AmazonCreds{VaultAddress: parts[0], VaultRole: parts[1], VaultToken: parts[2]}
			}
			if iamRole != "" {
				if amazonCreds != nil {
					return errors.Errorf("only one of --credentials, --vault, or --iam-role needs to be provided")
				}
				dArgs.opts.IAMRole = iamRole
			}
			volumeSize, err := strconv.Atoi(args[2])
			if err != nil {
				return errors.Errorf("volume size needs to be an integer; instead got %v", args[2])
			}
			if strings.TrimSpace(cloudfrontDistribution) != "" {
				log.Warningf("you specified a cloudfront distribution; deploying on " +
					"AWS with cloudfront is currently an alpha feature. No security " +
					"restrictions have been applied to cloudfront, making all data " +
					"public (obscured but not secured)\n")
			}
			bucket, region := strings.TrimPrefix(args[0], "s3://"), args[1]
			if !awsRegionRE.MatchString(region) {
				fmt.Fprintf(os.Stderr, "The AWS region seems invalid (does not match "+
					"%q). Do you want to continue deploying? [yN]\n", awsRegionRE)
				if s.Scan(); s.Text()[0] != 'y' && s.Text()[0] != 'Y' {
					os.Exit(1)
				}
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
			// Generate manifest and write assets.
			var buf bytes.Buffer
			if err = assets.WriteAmazonAssets(
				encoder(dArgs.globalFlags.OutputFormat, &buf), dArgs.opts, region, bucket, volumeSize,
				amazonCreds, cloudfrontDistribution, advancedConfig,
			); err != nil {
				return err
			}
			if err := kubectlCreate(dArgs.globalFlags.DryRun, buf.Bytes(), dArgs.opts); err != nil {
				return err
			}
			if !dArgs.globalFlags.DryRun || dArgs.contextFlags.CreateContext {
				if dArgs.contextFlags.ContextName == "" {
					dArgs.contextFlags.ContextName = "aws"
				}
				if err := contextCreate(dArgs.contextFlags.ContextName, dArgs.globalFlags.Namespace, dArgs.globalFlags.ServerCert); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	appendGlobalFlags(deployAmazon, dArgs.globalFlags)
	AppendS3Flags(deployAmazon, s3Flags)
	appendContextFlags(deployAmazon, dArgs.contextFlags)
	deployAmazon.Flags().StringVar(&cloudfrontDistribution, "cloudfront-distribution", "",
		"Deploying on AWS with cloudfront is currently "+
			"an alpha feature. No security restrictions have been"+
			"applied to cloudfront, making all data public (obscured but not secured)")
	deployAmazon.Flags().StringVar(&creds, "credentials", "", "Use the format \"<id>,<secret>[,<token>]\". You can get a token by running \"aws sts get-session-token\".")
	deployAmazon.Flags().StringVar(&vault, "vault", "", "Use the format \"<address/hostport>,<role>,<token>\".")
	deployAmazon.Flags().StringVar(&iamRole, "iam-role", "", fmt.Sprintf("Use the given IAM role for authorization, as opposed to using static credentials. The given role will be applied as the annotation %s, this used with a Kubernetes IAM role management system such as kube2iam allows you to give pachd credentials in a more secure way.", assets.IAMAnnotation))
	return deployAmazon
}
