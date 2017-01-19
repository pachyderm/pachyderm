package deploy

import (
	"bytes"
	"time"

	"github.com/pachyderm/pachyderm/src/server/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"

	"github.com/urfave/cli"
)

func newLocalCommand() cli.Command {
	return cli.Command{
		Name:        "local",
		ArgsUsage:   "local",
		Usage:       "Deploy a single-node Pachyderm cluster with local metadata storage.",
		Description: "Deploy a single-node Pachyderm cluster with local metadata storage.",
		Aliases:     []string{"l"},
		Action:      actLocal,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "host-path, a",
				Usage: "Location on the host machine where PFS metadata will be stored.",
				Value: "/var/pachyderm",
			},
		},
	}
}

func actLocal(c *cli.Context) (retErr error) {
	if c.GlobalBool("metrics") && !c.Bool("dev") {
		metricsFn := metrics.ReportAndFlushUserAction("Deploy")
		defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())
	}
	manifest := &bytes.Buffer{}
	if c.Bool("dev") {
		opts.Version = deploy.DevVersionTag
	}
	if err := assets.WriteLocalAssets(manifest, opts, c.String("host-path")); err != nil {
		return err
	}
	return maybeKcCreate(c, manifest)
}
