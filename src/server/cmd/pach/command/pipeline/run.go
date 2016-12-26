package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/client"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	pkgcmd "github.com/pachyderm/pachyderm/src/server/pkg/cmd"
	"github.com/pachyderm/pachyderm/src/server/pps/example"

	"github.com/urfave/cli"
)

func newRunCommand() cli.Command {
	return cli.Command{
		Name:        "run",
		Aliases:     []string{"r"},
		Usage:       "Run a pipeline once.",
		ArgsUsage:   "pipeline-name [-f job.json]",
		Description: descRun(),
		Action:      actRun,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "file, f",
				Usage: "The file containing the run-pipeline spec, - reads from stdin.",
				Value: "-",
			},
		},
	}
}

func descRun() string {
	str, err := (&jsonpb.Marshaler{Indent: "  "}).MarshalToString(example.RunPipelineSpec)
	if err != nil {
		pkgcmd.ErrorAndExit("error from RunPipeline: %s", err.Error())
	}
	return fmt.Sprintf("Run a pipeline once, optionally overriding some pipeline options by providing a spec.  The spec looks like this:\n%s", str)
}

func actRun(c *cli.Context) (retErr error) {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	request := &ppsclient.CreateJobRequest{
		Pipeline: &ppsclient.Pipeline{
			Name: c.Args().First(),
		},
		Force: true,
	}

	var buf bytes.Buffer
	var specReader io.Reader
	specPath := c.String("file")
	if specPath == "-" {
		specReader = io.TeeReader(os.Stdin, &buf)
		fmt.Print("Reading from stdin.\n")
	} else if specPath != "" {
		specFile, err := os.Open(specPath)
		if err != nil {
			return err
		}

		defer func() {
			if err := specFile.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()

		specReader = io.TeeReader(specFile, &buf)
		decoder := json.NewDecoder(specReader)
		s, err := replaceMethodAliases(decoder)
		if err != nil {
			return describeSyntaxError(err, buf)
		}
		if err := jsonpb.UnmarshalString(s, request); err != nil {
			return err
		}
	}

	job, err := clnt.PpsAPIClient.CreateJob(context.Background(), request)
	if err != nil {
		return sanitizeErr(err)
	}
	fmt.Println(job.ID)
	return nil
}
