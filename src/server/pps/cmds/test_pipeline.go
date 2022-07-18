package cmds

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fuse"
	pachdclient "github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/fuse"
)

const (
	name = "pfs"
)

func testPipeline(client *pachdclient.APIClient, pipelinePath string) (retErr error) {
	if pipelinePath == "" {
		pipelinePath = "-"
	}
	pipelineBytes, err := readPipelineBytes(pipelinePath)
	if err != nil {
		return err
	}
	pipelineReader, err := ppsutil.NewPipelineManifestReader(pipelineBytes)
	if err != nil {
		return err
	}
	request, err := pipelineReader.NextCreatePipelineRequest()
	if err != nil {
		return err
	}
	if err := client.UpdateRepo(request.Pipeline.Name); err != nil {
		return err
	}
	commit, err := client.StartCommit(request.Pipeline.Name, "master")
	if err != nil {
		return err
	}
	defer func() {
		if err := client.FinishCommit(request.Pipeline.Name, "master", commit.ID); err != nil {
			retErr = err
		}
	}()
	return client.ListDatumInput(request.Input, func(di *ppsclient.DatumInfo) error {
		mountOpts := make(map[string]*fuse.RepoOptions)
		for _, fi := range di.Data {
			mountOpts[fi.File.Commit.Branch.Repo.Name] = &fuse.RepoOptions{
				Name: fi.File.Commit.Branch.Repo.Name,
				File: fi.File,
			}
		}
		mountOpts["out"] = &fuse.RepoOptions{
			Name:  "out",
			File:  &pfs.File{Commit: commit},
			Write: true,
		}
		unmount := make(chan struct{})
		unmounted := make(chan struct{})
		var mountErr error
		go func() {
			mountErr = fuse.Mount(client, "/pfs", &fuse.Options{
				Write:       true,
				RepoOptions: mountOpts,
				Unmount:     unmount,
				Fuse: &fs.Options{
					MountOptions: gofuse.MountOptions{
						FsName: name,
						Name:   name,
					}},
			})
			close(unmounted)
		}()
		defer func() { <-unmounted }()
		defer func() { close(unmount) }()
		time.Sleep(time.Second)
		cmd := exec.Command(request.Transform.Cmd[0], request.Transform.Cmd[1:]...)
		if request.Transform.Stdin != nil {
			cmd.Stdin = strings.NewReader(strings.Join(request.Transform.Stdin, "\n") + "\n")
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
	return nil
}
