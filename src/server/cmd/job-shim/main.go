package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

const (
	PFSInputPrefix  = "/pfs"
	PFSOutputPrefix = "/pfs/out"
)

type appEnv struct {
	PachydermAddress string `env:"PACHD_PORT_650_TCP_ADDR,required"`
	PodName          string `env:"PPS_POD_NAME,required"`
}

func main() {
	env.Main(do, &appEnv{})
}

func downloadInput(c *client.APIClient, commitMounts []*fuse.CommitMount) error {
	for _, commitMount := range commitMounts {
		if commitMount.Alias == "prev" || commitMount.Alias == "out" {
			continue
		}
		repo := commitMount.Commit.Repo.Name
		commitID := commitMount.Commit.ID
		var fromCommitID string
		if commitMount.DiffMethod != nil && commitMount.DiffMethod.FromCommit != nil {
			fromCommitID = commitMount.DiffMethod.FromCommit.ID
		}
		var fullFile bool
		if commitMount.DiffMethod != nil {
			fullFile = commitMount.DiffMethod.FullFile
		}
		shard := commitMount.Shard
		fileInfos, err := c.ListFile(repo, commitID, "/", fromCommitID, fullFile, shard, false)
		if err != nil {
			return err
		}
		for _, fileInfo := range fileInfos {
			if err := os.MkdirAll(filepath.Join(PFSInputPrefix, repo), 0777); err != nil {
				return err
			}
			path := filepath.Join(PFSInputPrefix, repo, fileInfo.File.Path)
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			defer f.Close()
			w := bufio.NewWriter(f)
			if err := c.GetFile(repo, commitID, fileInfo.File.Path, 0, 0, fromCommitID, fullFile, shard, w); err != nil {
				return err
			}
			if err := w.Flush(); err != nil {
				return err
			}
		}
	}
	return os.MkdirAll(PFSOutputPrefix, 0777)
}

func uploadOutput(c *client.APIClient, out *fuse.CommitMount, overwrite bool) error {
	repo := out.Commit.Repo.Name
	commit := out.Commit.ID
	return filepath.Walk(PFSOutputPrefix, func(path string, info os.FileInfo, err error) error {
		if path == PFSOutputPrefix {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		relPath, err := filepath.Rel(PFSOutputPrefix, path)
		if err != nil {
			return err
		}
		if overwrite {
			if err := c.DeleteFile(repo, commit, relPath); err != nil {
				return err
			}
		}
		_, err = c.PutFile(repo, commit, relPath, f)
		return err
	})
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

			if err := downloadInput(c, response.CommitMounts); err != nil {
				return err
			}

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

			tick := time.Tick(10 * time.Second)
			for {
				select {
				case success := <-cmdCh:
					var outputMount *fuse.CommitMount
					for _, c := range response.CommitMounts {
						if c.Alias == "out" {
							outputMount = c
							break
						}
					}
					if err := uploadOutput(c, outputMount, response.Transform.Overwrite); err != nil {
						fmt.Printf("err from uploading output: %s\n", err)
						success = false
					}

					res, err := ppsClient.FinishPod(
						context.Background(),
						&ppsserver.FinishPodRequest{
							ChunkID: response.ChunkID,
							PodName: appEnv.PodName,
							Success: success,
						},
					)
					if err != nil {
						return err
					}
					finished = true
					if res.Fail {
						return errors.New("restarting")
					}
					return nil
				case <-tick:
					res, err := ppsClient.ContinuePod(
						context.Background(),
						&ppsserver.ContinuePodRequest{
							ChunkID: response.ChunkID,
							PodName: appEnv.PodName,
						},
					)
					if err != nil {
						return err
					}
					if res.Exit {
						return nil
					}
				}
			}
			return nil
		}),
	}

	return rootCmd.Execute()
}
