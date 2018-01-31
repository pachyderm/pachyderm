package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/client"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pps"

	"github.com/spf13/cobra"
)

func Cmds(noMetrics *bool) []*cobra.Command {
	metrics := !*noMetrics
	marshaller := &jsonpb.Marshaler{Indent: "  "}

	migrate1_6to1_7 := &cobra.Command{
		Use:   "migrate-1.6-to-1.7",
		Short: "Migrate a 1.6 cluster to 1.7.",
		Long:  "Migrate a 1.6 cluster to 1.7.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			c, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			fmt.Print(`Pachctl will perform the following operations to migrate:
1. Extract pipeline manifests from 1.6 cluster to ./pipelines.json.
2. Delete pipelines from 1.6 cluster.
3. Upgrade cluster to 1.7.
4. Recreate pipelines in 1.7 cluster.
`)
			fmt.Println("Extracting to pipelines.json.")
			pipelineInfos, err := c.ListPipeline()
			if err != nil {
				return err
			}
			if err := func() (retErr error) {
				f, err := os.Create("pipelines.json")
				if err != nil {
					return err
				}
				defer func() {
					if err := f.Close(); err != nil && retErr == nil {
						retErr = err
					}
				}()
				for _, pipelineInfo := range pipelineInfos {
					if err := marshaller.Marshal(f, pps.PipelineReqFromInfo(pipelineInfo)); err != nil {
						return err
					}
				}
				return nil
			}(); err != nil {
				return err
			}
			fmt.Println("Deleting pipelines.")
			if _, err := c.PpsAPIClient.DeletePipeline(context.Background(), &ppsclient.DeletePipelineRequest{
				DeleteJobs: true,
				All:        true,
			}); err != nil {
				return err
			}
			fmt.Println("Deploying 1.7.0 image.")
			cmdIO := cmdutil.IO{
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}
			if err := cmdutil.RunIO(cmdIO, "kubectl", "set", "image", "deployment/pachd", "pachd=pachyderm/pachd:1.7.0"); err != nil {
				return err
			}
			fmt.Println("Redploying pipelines.")
			if err := backoff.RetryNotify(func() error {
				c, err = client.NewOnUserMachine(metrics, "user")
				if err != nil {
					return err
				}
				return nil
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				fmt.Printf("Error connecting to 1.7 cluster: %v, retry in: %v.\n", err, d)
				return nil
			}); err != nil {
				return err
			}

			cfgReader, err := pps.NewPipelineManifestReader("pipelines.json")
			if err != nil {
				return err
			}
			for {
				request, err := cfgReader.NextCreatePipelineRequest()
				if err == io.EOF {
					break
				} else if err != nil {
					return err
				}
				if _, err := c.PpsAPIClient.CreatePipeline(context.Background(), request); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	return []*cobra.Command{migrate1_6to1_7}
}
