package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmd"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"github.com/spf13/cobra"
	"go.pedge.io/env"
	"go.pedge.io/lion"
	"golang.org/x/net/context"
)

type appEnv struct {
	PachydermAddress string `env:"PACHD_PORT_650_TCP_ADDR,required"`
	PodName          string `env:"PPS_POD_NAME,required"`
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
		Run: cmd.RunFixedArgs(1, func(args []string) (retErr error) {
			ppsClient, err := ppsserver.NewInternalPodAPIClientFromAddress(fmt.Sprintf("%v:650", appEnv.PachydermAddress))
			if err != nil {
				return err
			}
			response, err := ppsClient.StartPod(
				context.Background(),
				&ppsserver.StartPodRequest{
					Job: &ppsclient.Job{
						ID: args[0],
					},
					PodName: appEnv.PodName,
				})
			if err != nil {
				lion.Errorf("error from StartPod: %s", err.Error())
				return err
			}

			// Start sending ContinuePod to PPS to signal that we are alive
			exitCh := make(chan struct{})
			go func() {
				tick := time.Tick(10 * time.Second)
				for {
					<-tick
					res, err := ppsClient.ContinuePod(
						context.Background(),
						&ppsserver.ContinuePodRequest{
							ChunkID: response.ChunkID,
							PodName: appEnv.PodName,
						},
					)
					if err != nil {
						lion.Errorf("error from ContinuePod: %s", err.Error())
					}
					if res != nil && res.Exit {
						select {
						case exitCh <- struct{}{}:
							// If someone received this signal, then they are
							// responsible to exiting the program and release
							// all resources.
							return
						default:
							// Otherwise, we just terminate the program.
							lion.Errorf("chunk was revoked. restarting...")
							os.Exit(1)
						}
					}
				}
			}()

			if response.Transform.Debug {
				lion.SetLevel(lion.LevelDebug)
			}
			// We want to make sure that we only send FinishPod once.
			// The most bulletproof way would be to check that on server side,
			// but this is easier.
			var finished bool
			// Make sure that we call FinishPod even if something caused a panic
			defer func() {
				if r := recover(); r != nil && !finished {
					fmt.Println("job shim crashed; this is like a bug in pachyderm")
					if _, err := ppsClient.FinishPod(
						context.Background(),
						&ppsserver.FinishPodRequest{
							ChunkID: response.ChunkID,
							PodName: appEnv.PodName,
							Success: false,
						},
					); err != nil && retErr == nil {
						retErr = err
					}
				}
			}()

			c, err := client.NewFromAddress(fmt.Sprintf("%v:650", appEnv.PachydermAddress))
			if err != nil {
				return err
			}

			mounter := fuse.NewMounter(appEnv.PachydermAddress, c)
			ready := make(chan bool)
			errCh := make(chan error)
			go func() {
				if err := mounter.MountAndCreate(
					"/pfs",
					nil,
					response.CommitMounts,
					ready,
					response.Transform.Debug,
					false,
				); err != nil {
					errCh <- err
				}
			}()
			select {
			case <-ready:
			case err := <-errCh:
				return err
			}
			defer func() {
				if err := mounter.Unmount("/pfs"); err != nil && retErr == nil {
					retErr = err
				}
			}()
			var readers []io.Reader
			for _, line := range response.Transform.Stdin {
				readers = append(readers, strings.NewReader(line+"\n"))
			}
			if len(response.Transform.Cmd) == 0 {
				fmt.Println("unable to run; a cmd needs to be provided")
				if _, err := ppsClient.FinishPod(
					context.Background(),
					&ppsserver.FinishPodRequest{
						ChunkID: response.ChunkID,
						PodName: appEnv.PodName,
						Success: false,
					},
				); err != nil {
					return err
				}
				finished = true
				return
			}

			cmdCh := make(chan bool)
			go func() {
				cmd := exec.Command(response.Transform.Cmd[0], response.Transform.Cmd[1:]...)
				cmd.Stdin = io.MultiReader(readers...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				success := true
				if err := cmd.Run(); err != nil {
					success = false
					if exiterr, ok := err.(*exec.ExitError); ok {
						if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
							for _, returnCode := range response.Transform.AcceptReturnCode {
								if int(returnCode) == status.ExitStatus() {
									success = true
								}
							}
						}
					}
					if !success {
						fmt.Fprintf(os.Stderr, "Error from exec: %s\n", err.Error())
					}
				}
				cmdCh <- success
			}()

			var success bool
			select {
			case <-exitCh:
				return fmt.Errorf("chunk was revoked. restarting...")
			case success = <-cmdCh:
			}
			var outputMount *fuse.CommitMount
			for _, c := range response.CommitMounts {
				if c.Alias == "out" {
					outputMount = c
					break
				}
			}
			return nil
		}),
	}

	return rootCmd.Execute()
}
