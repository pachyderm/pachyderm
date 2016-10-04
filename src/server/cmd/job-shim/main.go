package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

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
}

func main() {
	env.Main(do, &appEnv{})
}

func downloadInput(c *client.APIClient, commitMounts []*fuse.CommitMount) error {
	for _, commitMount := range commitMounts {
		repo := commitMount.Commit.Repo.Name
		commitID := commitMount.Commit.ID
		fromCommitID := commitMount.DiffMethod.FromCommit.ID
		fullFile := commitMount.DiffMethod.FullFile
		shard := commitMount.Shard
		if commitMount.Alias == "prev" || commitMount.Alias == "out" {
			continue
		}
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
			if err := c.GetFile(repo, commitID, path, 0, 0, fromCommitID, fullFile, shard, w); err != nil {
				return err
			}
			if err := w.Flush(); w != nil {
				return err
			}
		}
	}
	return os.MkdirAll(PFSOutputPrefix, 0777)
}

func uploadOutput(c *client.APIClient, out *fuse.CommitMount) error {
	repo := out.Commit.Repo.Name
	commit := out.Commit.ID
	return filepath.Walk(PFSOutputPrefix, func(path string, info os.FileInfo, err error) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(PFSOutputPrefix, path)
		if err != nil {
			return err
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
			ppsClient, err := ppsserver.NewInternalJobAPIClientFromAddress(fmt.Sprintf("%v:650", appEnv.PachydermAddress))
			if err != nil {
				return err
			}
			response, err := ppsClient.StartJob(
				context.Background(),
				&ppsserver.StartJobRequest{
					Job: &ppsclient.Job{
						ID: args[0],
					}})
			if err != nil {
				lion.Errorf("error from StartJob: %s", err.Error())
				return err
			}

			if response.Transform.Debug {
				lion.SetLevel(lion.LevelDebug)
			}
			// We want to make sure that we only send FinishJob once.
			// The most bulletproof way would be to check that on server side,
			// but this is easier.
			var finished bool
			// Make sure that we call FinishJob even if something caused a panic
			defer func() {
				if r := recover(); r != nil && !finished {
					fmt.Println("job shim crashed; this is like a bug in pachyderm")
					if _, err := ppsClient.FinishJob(
						context.Background(),
						&ppsserver.FinishJobRequest{
							Job: &ppsclient.Job{
								ID: args[0],
							},
							Success:  false,
							PodIndex: response.PodIndex,
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
				if _, err := ppsClient.FinishJob(
					context.Background(),
					&ppsserver.FinishJobRequest{
						Job:      client.NewJob(args[0]),
						Success:  false,
						PodIndex: response.PodIndex,
					},
				); err != nil {
					return err
				}
				finished = true
				return
			}
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
			if _, err := ppsClient.FinishJob(
				context.Background(),
				&ppsserver.FinishJobRequest{
					Job:      client.NewJob(args[0]),
					Success:  success,
					PodIndex: response.PodIndex,
				},
			); err != nil {
				return err
			}
			finished = true

			var outputMount *fuse.CommitMount
			for _, c := range response.CommitMounts {
				if c.Alias == "out" {
					outputMount = c
					break
				}
			}
			return uploadOutput(c, outputMount)
		}),
	}

	return rootCmd.Execute()
}
