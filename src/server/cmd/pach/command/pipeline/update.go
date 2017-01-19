package pipeline

import (
	"context"
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/client"

	"github.com/urfave/cli"
)

func newUpdateCommand() cli.Command {
	pipelineSpec := string(pachyderm.MustAsset("doc/deployment/pipeline_spec.md"))
	return cli.Command{
		Name:        "update",
		Aliases:     []string{"u"},
		Usage:       "Update an existing Pachyderm pipeline.",
		ArgsUsage:   "-f pipeline.json",
		Description: fmt.Sprintf("Update a Pachyderm pipeline with a new spec\n\n%s", pipelineSpec),
		Action:      actUpdate,
		Flags: []cli.Flag{
			cli.BoolTFlag{
				Name:  "archive, a",
				Usage: "Whether or not to archive existing commits in this pipeline's output repo.",
			},
			cli.StringFlag{
				Name:  "file, f",
				Usage: "The file containing the job, it can be a url or local file. - reads from stdin.",
				Value: "-",
			},
			cli.BoolFlag{
				Name:  "push-images, p",
				Usage: "If true, push local docker images into the cluster registry.",
			},
			cli.StringFlag{
				Name:  "registry, r",
				Usage: "The registry to push images to.",
				Value: "docker.io",
			},
			cli.StringFlag{
				Name:  "username, u",
				Usage: "The username to push images as, defaults to your OS username.",
			},
			cli.StringFlag{
				Name:  "password, pw",
				Usage: "Your password for the registry being pushed to.",
			},
		},
	}
}

func actUpdate(c *cli.Context) error {
	cfgReader, err := newPipelineManifestReader(c.String("file"))
	if err != nil {
		return err
	}
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	for {
		request, err := cfgReader.nextCreatePipelineRequest()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		request.Update = true
		request.NoArchive = !c.BoolT("archive")
		if c.Bool("push-images") {
			pushedImage, err := pushImage(c.String("registry"), c.String("username"), c.String("password"), request.Transform.Image)
			if err != nil {
				return err
			}
			request.Transform.Image = pushedImage
		}
		if _, err := clnt.PpsAPIClient.CreatePipeline(context.Background(), request); err != nil {
			return sanitizeErr(err)
		}
	}
	return nil
}
