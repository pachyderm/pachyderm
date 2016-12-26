package deploy

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"

	"github.com/urfave/cli"
)

func newAmazonCommand() cli.Command {
	return cli.Command{
		Name:        "amazon",
		Aliases:     []string{"a"},
		Usage:       "Deploy a Pachyderm cluster running on AWS.",
		ArgsUsage:   "<S3 bucket> <id> <secret> <token> <region> <EBS volume names> <size of volumes (in GB)>",
		Description: descAmazon,
		Action:      actAmazon,
	}
}

var descAmazon = `Deploy a Pachyderm cluster running on AWS. Arguments are:

     <S3 bucket>: An S3 bucket where Pachyderm will store PFS data.
     <id>, <secret>, <token>: Session token details, used for authorization. You can get these by running 'aws sts get-session-token'
     <region>: The aws region where pachyderm is being deployed (e.g. us-west-1)
     <EBS volume names>: A comma-separated list of EBS volumes, one per rethink shard (see --rethink-shards).
     <size of volumes>: Size of EBS volumes, in GB (assumed to all be the same).`

func actAmazon(c *cli.Context) (retErr error) {
	if c.GlobalBool("metrics") && !c.Bool("dev") {
		metricsFn := metrics.ReportAndFlushUserAction("Deploy")
		defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())
	}

	args := c.Args()
	volumeNames := strings.Split(args[5], ",")
	volumeSize, err := strconv.Atoi(args[6])
	if err != nil {
		return fmt.Errorf("volume size needs to be an integer; instead got %v", args[6])
	}
	manifest := &bytes.Buffer{}
	if err = assets.WriteAmazonAssets(manifest, opts, args[0], args[1], args[2], args[3], args[4], volumeNames, volumeSize); err != nil {
		return err
	}
	return maybeKcCreate(c, manifest)
}
