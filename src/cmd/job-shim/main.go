package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/spf13/cobra"
	"go.pedge.io/env"
	"go.pedge.io/lion"
	"go.pedge.io/pkg/exec"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type appEnv struct {
	PachydermAddress string `env:"PACHD_PORT_650_TCP_ADDR,required"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	lion.SetLevel(lion.LevelDebug)
	appEnv := appEnvObj.(*appEnv)
	rootCmd := &cobra.Command{
		Use:   os.Args[0] + " job-id",
		Short: `Pachyderm job-shim, coordinates with ppsd to create an output commit and run user work.`,
		Long:  `Pachyderm job-shim, coordinates with ppsd to create an output commit and run user work.`,
		Run: func(cmd *cobra.Command, args []string) {
			client, err := getPachClient(appEnv)
			if err != nil {
				errorAndExit(err.Error())
			}
			response, err := client.StartJob(
				context.Background(),
				&pps.StartJobRequest{
					Job: &pps.Job{
						Id: args[0],
					}})
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(0)
			}

			mounter := fuse.NewMounter(appEnv.PachydermAddress, client)
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
			io := pkgexec.IO{
				Stdin:  strings.NewReader(response.Transform.Stdin),
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}
			success := true
			if err := pkgexec.RunIO(io, response.Transform.Cmd...); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				success = false
			}
			if _, err := client.FinishJob(
				context.Background(),
				&pps.FinishJobRequest{
					Job: &pps.Job{
						Id: args[0],
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

func getPachClient(appEnv *appEnv) (*pachyderm.APIClient, error) {
	clientConn, err := grpc.Dial(fmt.Sprintf("%s:650", appEnv.PachydermAddress), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pachyderm.NewAPIClient(clientConn), nil
}

func errorAndExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}
