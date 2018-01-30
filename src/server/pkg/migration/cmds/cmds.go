package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/client"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pps"

	"github.com/spf13/cobra"
)

func Cmds(noMetrics *bool) ([]*cobra.Command, error) {
	metrics := !*noMetrics
	marshaller := &jsonpb.Marshaler{Indent: "  "}

	migrate1_6to1_7 := &cobra.Command{
		Use:   "migrate-1.6-to-1.7",
		Short: "Migrate a 1.6 cluster to 1.7.",
		Long:  "Migrate a 1.6 cluster to 1.7.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			client, err := client.NewOnUserMachine(metrics, "user")
			if err != nil {
				return err
			}
			fmt.Print(`Pachctl will perform the following operations to migrate:
1. Extract pipeline manifests from 1.6 cluster to ./pipelines.json.
2. Delete pipelines from 1.6 cluster.
3. Upgrade cluster to 1.7.
4. Recreate pipelines in 1.7 cluster.`)
			pipelineInfos, err := client.ListPipeline()
			if err != nil {
				return err
			}
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
			if _, err := client.PpsAPIClient.DeletePipeline(context.Background(), &ppsclient.DeletePipelineRequest{
				DeleteJobs: true,
				All:        true,
			}); err != nil {
				return err
			}
			return nil
		}),
	}
	return []*cobra.Command{migrate1_6to1_7}, nil
}
