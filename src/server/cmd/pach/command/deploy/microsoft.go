package deploy

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"

	"github.com/urfave/cli"
)

func newMicrosoftCommand() cli.Command {
	return cli.Command{
		Name:        "microsoft",
		Aliases:     []string{"m"},
		Usage:       "Deploy a Pachyderm cluster running on Microsoft Azure.",
		ArgsUsage:   "<container> <storage account name> <storage account key> <volume URIs> <size of volumes (in GB)>",
		Description: descMicrosoft,
		Action:      actMicrosoft,
	}
}

var descMicrosoft = `Deploy a Pachyderm cluster running on Microsoft Azure. Arguments are:

   <container>: An Azure container where Pachyderm will store PFS data.
   <volume URIs>: A comma-separated list of persistent volumes, one per rethink shard (see --rethink-shards).
   <size of volumes>: Size of persistent volumes, in GB (assumed to all be the same).`

func actMicrosoft(c *cli.Context) (retErr error) {
	if c.GlobalBool("metrics") && !c.Bool("dev") {
		metricsFn := metrics.ReportAndFlushUserAction("Deploy")
		defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())
	}

	args := c.Args()
	if _, err := base64.StdEncoding.DecodeString(args[2]); err != nil {
		return fmt.Errorf("storage-account-key needs to be base64 encoded; instead got '%v'", args[2])
	}
	volumeURIs := strings.Split(args[3], ",")
	for i, uri := range volumeURIs {
		tempURI, err := url.ParseRequestURI(uri)
		if err != nil {
			return fmt.Errorf("All volume-uris needs to be a well-formed URI; instead got '%v'", uri)
		}
		volumeURIs[i] = tempURI.String()
	}
	volumeSize, err := strconv.Atoi(args[4])
	if err != nil {
		return fmt.Errorf("volume size needs to be an integer; instead got %v", args[4])
	}
	manifest := &bytes.Buffer{}
	if err = assets.WriteMicrosoftAssets(manifest, opts, args[0], args[1], args[2], volumeURIs, volumeSize); err != nil {
		return err
	}
	return maybeKcCreate(c, manifest)
}
