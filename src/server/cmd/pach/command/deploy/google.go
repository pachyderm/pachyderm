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

func newGoogleCommand() cli.Command {
	return cli.Command{
		Name:        "google",
		Aliases:     []string{"g"},
		Usage:       "Deploy a Pachyderm cluster running on GCP.",
		ArgsUsage:   "<GCS bucket> <GCE persistent disks> <size of disks (in GB)>",
		Description: descGoogle,
		Action:      actGoogle,
	}

}

var descGoogle = `Deploy a Pachyderm cluster running on GCP. Arguments are:

     <GCS bucket>: A GCS bucket where Pachyderm will store PFS data.
     <GCE persistent disks>: A comma-separated list of GCE persistent disks, one per rethink shard (see --rethink-shards).
     <size of disks>: Size of GCE persistent disks in GB (assumed to all be the same).`

func actGoogle(c *cli.Context) (retErr error) {
	if c.GlobalBool("metrics") && !c.Bool("dev") {
		metricsFn := metrics.ReportAndFlushUserAction("Deploy")
		defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())
	}
	args := c.Args()
	volumeNames := strings.Split(args[1], ",")
	volumeSize, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("volume size needs to be an integer; instead got %v", args[2])
	}
	manifest := &bytes.Buffer{}
	if err = assets.WriteGoogleAssets(manifest, opts, args[0], volumeNames, volumeSize); err != nil {
		return err
	}
	return maybeKcCreate(c, manifest)
}
