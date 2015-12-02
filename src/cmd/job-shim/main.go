package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/spf13/cobra"
	"go.pedge.io/env"
	"go.pedge.io/pkg/exec"
	"google.golang.org/grpc"
)

type appEnv struct {
	PachydermPfsd1Port string `env:"PACHYDERM_PFSD_1_PORT"`
	PfsAddress         string `env:"PFS_ADDRESS,default=0.0.0.0:650"`
	PachydermPpsd1Port string `env:"PACHYDERM_PPSD_1_PORT"`
	PpsAddress         string `env:"PPS_ADDRESS,default=0.0.0.0:651"`
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
			pfsAPIClient, err := getPfsAPIClient(getPfsdAddress(appEnv))
			if err != nil {
				errorAndExit(err.Error())
			}

			ppsAPIClient, err := getPpsAPIClient(getPpsdAddress(appEnv))
			if err != nil {
				errorAndExit(err.Error())
			}

			response, err := ppsAPIClient.StartJob(
				context.Background(),
				&pps.StartJobRequest{
					Job: &pps.Job{
						Id: args[0],
					}})
			if err != nil {
				errorAndExit(err.Error())
			}

			mounter := fuse.NewMounter(getPfsdAddress(appEnv), pfsAPIClient)
			ready := make(chan bool)
			go func() {
				if err := mounter.Mount("/pfs", response.Shard, append(response.InputCommit, response.OutputCommit), ready); err != nil {
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
				Stdin:  os.Stdin,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}
			if err := pkgexec.RunIO(io, response.Transform.Cmd...); err != nil {
				errorAndExit(err.Error())
			}
			if _, err := ppsAPIClient.FinishJob(
				context.Background(),
				&pps.FinishJobRequest{
					Job: &pps.Job{
						Id: args[0],
					}},
			); err != nil {
				errorAndExit(err.Error())
			}
		},
	}

	return rootCmd.Execute()
}

func getPfsdAddress(appEnv *appEnv) string {
	if pfsdAddr := os.Getenv("PFSD_PORT_650_TCP_ADDR"); pfsdAddr != "" {
		return fmt.Sprintf("%s:650", pfsdAddr)
	}
	if appEnv.PachydermPfsd1Port != "" {
		return strings.Replace(appEnv.PachydermPfsd1Port, "tcp://", "", -1)
	}
	return appEnv.PfsAddress
}

func getPpsdAddress(appEnv *appEnv) string {
	if ppsdAddr := os.Getenv("PPSD_PORT_651_TCP_ADDR"); ppsdAddr != "" {
		return fmt.Sprintf("%s:651", ppsdAddr)
	}
	if appEnv.PachydermPpsd1Port != "" {
		return strings.Replace(appEnv.PachydermPpsd1Port, "tcp://", "", -1)
	}
	return appEnv.PpsAddress
}

func getPfsAPIClient(address string) (pfs.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pfs.NewAPIClient(clientConn), nil
}

func getPpsAPIClient(address string) (pps.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pps.NewAPIClient(clientConn), nil
}

func errorAndExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", fmt.Sprintf(format, args...))
	os.Exit(1)
}
