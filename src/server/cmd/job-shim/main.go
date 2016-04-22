package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"github.com/spf13/cobra"
	"go.pedge.io/env"
	"go.pedge.io/pkg/exec"
	"golang.org/x/net/context"
)

type appEnv struct {
	PachydermAddress string `env:"PACHD_PORT_650_TCP_ADDR,required"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	rootCmd := &cobra.Command{
		Use:   os.Args[0] + " job-id",
		Short: `Pachyderm job-shim, coordinates with ppsd to create an output commit and run user work.`,
		Long:  `Pachyderm job-shim, coordinates with ppsd to create an output commit and run user work.`,
		Run: func(cmd *cobra.Command, args []string) {
			ppsClient, err := ppsserver.NewInternalJobAPIClientFromAddress(fmt.Sprintf("%v:650", appEnv.PachydermAddress))
			if err != nil {
				errorAndExit(err.Error())
			}
			response, err := ppsClient.StartJob(
				context.Background(),
				&ppsserver.StartJobRequest{
					Job: &ppsclient.Job{
						ID: args[0],
					}})
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(0)
			}

			pfsClient, err := client.NewFromAddress(fmt.Sprintf("%v:650", appEnv.PachydermAddress))
			if err != nil {
				errorAndExit(err.Error())
			}

			// We want to make sure that we actually have some input data to
			// operate on.
			var hasData bool
			for _, commitMount := range response.CommitMounts {
				fileInfos, err := pfsClient.ListFile(context.Background(), &pfsclient.ListFileRequest{
					File: &pfsclient.File{
						Commit: commitMount.Commit,
						Path:   "", // the root directory
					},
					Shard:      commitMount.Shard,
					FromCommit: commitMount.FromCommit,
					Recurse:    true,
				})
				if err != nil {
					errorAndExit(err.Error())
				}

				for _, fileInfo := range fileInfos.FileInfo {
					if fileInfo.SizeBytes > 0 {
						hasData = true
						break
					}
				}
			}

			// If this pod is not assigned any input data, we simply report success
			if !hasData {
				if _, err := ppsClient.FinishJob(
					context.Background(),
					&ppsserver.FinishJobRequest{
						Job: &ppsclient.Job{
							ID: args[0],
						},
						Index:   response.Index,
						Success: true,
					},
				); err != nil {
					errorAndExit(err.Error())
				}
				return
			}

			mounter := fuse.NewMounter(appEnv.PachydermAddress, pfsClient)
			ready := make(chan bool)
			go func() {
				if err := mounter.Mount(
					"/pfs",
					nil,
					response.CommitMounts,
					ready,
				); err != nil {
					errorAndExit(err.Error())
				}
			}()
			<-ready
			defer func() {
				if err := mounter.Unmount("/pfs"); err != nil {
					errorAndExit(err.Error())
				}
			}()
			var readers []io.Reader
			for _, line := range response.Transform.Stdin {
				readers = append(readers, strings.NewReader(line+"\n"))
			}
			io := pkgexec.IO{
				Stdin:  io.MultiReader(readers...),
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}
			success := true
			if err := pkgexec.RunIO(io, response.Transform.Cmd...); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				success = false
			}
			if _, err := ppsClient.FinishJob(
				context.Background(),
				&ppsserver.FinishJobRequest{
					Job: &ppsclient.Job{
						ID: args[0],
					},
					Index:   response.Index,
					Success: success,
				},
			); err != nil {
				errorAndExit(err.Error())
			}
		},
	}

	return rootCmd.Execute()
}

func errorAndExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}
