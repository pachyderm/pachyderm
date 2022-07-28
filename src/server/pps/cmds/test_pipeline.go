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
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tabwriter"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
)

const (
	name   = "pfs"
	branch = "dev"
)

func testPipeline(client *pachdclient.APIClient, pipelinePath string, datumLimit int, failedJob string) (retErr error) {
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
		var lastSuccessJi *pps.JobInfo
		if failedJob != "" {
			client.ListJobF(pipeline, nil, -1, false, func(ji *pps.JobInfo) error {
				if ji.State == pps.JobState_JOB_SUCCESS {
					lastSuccessJi = ji
					return errutil.ErrBreak
				}
				return nil
			})
		}
		if err := txnC.UpdateRepo(pipeline); err != nil {
			return err
		}
		var commit *pfs.Commit
		if lastSuccessJi != nil {
			commit, err = txnC.StartCommitParent(pipeline, branch, lastSuccessJi.OutputCommit.Branch.Name, lastSuccessJi.Job.ID)
		} else {
			commit, err = txnC.StartCommit(pipeline, branch)
		}
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
		// Rewrite the input to reference development branches.
		if err := pps.VisitInput(request.Input, func(input *pps.Input) error {
			if input.Pfs != nil && outCommits[input.Pfs.Repo] != nil {
				input.Pfs.Branch = branch
			}
			return nil
		}); err != nil {
			return err
		}
		fmt.Printf("Running pipeline: %s\n", request.Pipeline.Name)
		datums := 0
		runDatum := func(di *ppsclient.DatumInfo) error {
			if di.State == pps.DatumState_SUCCESS || di.State == pps.DatumState_SKIPPED {
				return nil
			}
			datums++
			if datumLimit != 0 && datums > datumLimit {
				return errutil.ErrBreak
			}
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
		}
		var ji *pps.JobInfo
		if failedJob != "" {
			ji, err = client.InspectJob(request.Pipeline.Name, failedJob, false)
			if err != nil {
				return err
			}
		}
		if ji != nil && ji.State == pps.JobState_JOB_FAILURE {
			if err := client.ListDatum(request.Pipeline.Name, failedJob, runDatum); err != nil {
				return err
			}
		} else {
			if err := client.ListDatumInput(request.Input, runDatum); err != nil {
				return err
			}
		}
		if err := client.FinishCommit(request.Pipeline.Name, branch, commitID); err != nil {
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
