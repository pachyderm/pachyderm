package cmds

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fuse"
	pachdclient "github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tabwriter"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
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
	var requests []*ppsclient.CreatePipelineRequest
	for {
		request, err := pipelineReader.NextCreatePipelineRequest()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		requests = append(requests, request)
	}
	txn, err := client.StartTransaction()
	if err != nil {
		return err
	}
	txnC := client.WithTransaction(txn)
	outCommits := make(map[string]*pfs.Commit)
	var commitID string
	for _, request := range requests {
		pipeline := request.Pipeline.Name
		if err := txnC.UpdateRepo(pipeline); err != nil {
			return err
		}
		commit, err := txnC.StartCommit(pipeline, "master")
		if err != nil {
			return err
		}
		commitID = commit.ID
		outCommits[pipeline] = commit
	}
	if _, err := client.FinishTransaction(txn); err != nil {
		return err
	}
	finishedCommits := make(map[string]bool)
	defer func() {
		for pipeline, outC := range outCommits {
			if !finishedCommits[pipeline] {
				if err := client.FinishCommit(outC.Branch.Repo.Name, outC.Branch.Name, outC.ID); err != nil && retErr == nil {
					retErr = err
				}
			}
		}
	}()
	for _, request := range requests {
		fmt.Printf("Running pipeline: %s\n", request.Pipeline.Name)
		if err := client.ListDatumInput(request.Input, func(di *ppsclient.DatumInfo) error {
			fmt.Printf("Running datum: %s with files:\n", di.Datum.ID)
			writer := tabwriter.NewWriter(os.Stdout, pretty.FileHeaderWithRepoAndCommit)
			for _, fi := range di.Data {
				pretty.PrintFileInfo(writer, fi, false, true, true)
			}
			if err := writer.Flush(); err != nil {
				return err
			}
			mountOpts := make(map[string]*fuse.RepoOptions)
			for _, fi := range di.Data {
				mountOpts[fi.File.Commit.Branch.Repo.Name] = &fuse.RepoOptions{
					Name: fi.File.Commit.Branch.Repo.Name,
					File: fi.File,
				}
			}
			mountOpts["out"] = &fuse.RepoOptions{
				Name:  "out",
				File:  &pfs.File{Commit: outCommits[request.Pipeline.Name]},
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
		}); err != nil {
			return err
		}
		if err := client.FinishCommit(request.Pipeline.Name, "master", commitID); err != nil {
			return err
		}
		finishedCommits[request.Pipeline.Name] = true
	}
	mountOpts := make(map[string]*fuse.RepoOptions)
	for _, outC := range outCommits {
		mountOpts[outC.Branch.Repo.Name] = &fuse.RepoOptions{
			Name: outC.Branch.Repo.Name,
			File: &pfs.File{Commit: outC},
		}
	}
	fmt.Println("Mounting output at /pfs, CTRL-c to unmount.")
	return fuse.Mount(client, "/pfs", &fuse.Options{
		RepoOptions: mountOpts,
		Fuse: &fs.Options{
			MountOptions: gofuse.MountOptions{
				FsName: name,
				Name:   name,
			}},
	})
}
