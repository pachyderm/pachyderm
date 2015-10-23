package pipelineserver

import (
	"fmt"

	"golang.org/x/net/context"

	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pedge.io/protolog"
)

type pipelineController struct {
	pfsAPIClient      pfs.APIClient
	jobAPIClient      pps.JobAPIClient
	pipelineAPIClient pps.PipelineAPIClient

	pipelineInfo    *pps.PipelineInfo
	cancelC         chan bool
	finishedCancelC chan bool
}

func newPipelineController(
	pfsAPIClient pfs.APIClient,
	jobAPIClient pps.JobAPIClient,
	pipelineAPIClient pps.PipelineAPIClient,
	pipelineInfo *pps.PipelineInfo,
) *pipelineController {
	return &pipelineController{
		pfsAPIClient,
		jobAPIClient,
		pipelineAPIClient,
		pipelineInfo,
		make(chan bool),
		make(chan bool),
	}
}

func (p *pipelineController) Start() error {
	// TODO(pedge): do not get all jobs each time, need a limit call on persist, more
	// generally, need all persist calls to have a limit
	jobInfos, err := p.jobAPIClient.ListJob(context.Background(), &pps.ListRequest{Pipeline: p.pipelineInfo.Pipeline})
	if err != nil {
		return err
	}
	lastCommit := &pfs.Commit{
		Repo: a.PipelineInfo.Input,
		// TODO(pedge): use initial commit id when moved to pfs package
		Id: "scratch",
	}
	if len(jobs.Job) > 0 {
		lastCommit = jobs.Job[0].Input
	}
	go func() {
		if err := p.run(lastCommit); err != nil {
			// TODO(pedge): what to do with error?
			protolog.Errorln(err.Error())
		}
	}()
	return nil
}

func (p *pipelineController) Cancel() {
	p.cancelC <- true
	close(p.cancelC)
	<-p.finishedCancelC
}

func (p *pipelineController) run(lastCommit *pfs.Commit) error {
	for {
		// TODO(pedge): this is not what we want, the ListCommit
		// context should take the cancel chan and the pfs api server implementation
		// should handle it, for now we do this, but this also means the ListCommit
		// call will not be cancelled and the goroutine will continue to run
		// just overall not good, we need a discussion about handling cancel and
		// look into if gRPC does this automatically
		var commitInfos *pfs.CommitInfos
		var err error
		done := make(chan bool)
		go func() {
			commitInfos, err = p.pfsAPIClient.ListCommit(
				context.Background(),
				&pfs.ListCommitRequest{
					Repo:       lastCommit.Repo,
					CommitType: pfs.CommitType_COMMIT_TYPE_READ,
					From:       lastCommit,
					Block:      true,
				},
			)
			done <- true
			close(done)
		}()
		select {
		case <-p.cancelC:
			p.finishedCancelC <- true
			close(p.finishedCancelC)
			return nil
		case <-done:
			if err != nil {
				return err
			}
			if len(commitInfos.CommitInfo) == 0 {
				return fmt.Errorf("pachyderm.pps.pipelineserver: we expected at least one *pfs.CommitInfo returned from blocking call, but no *pfs.CommitInfo structs were returned for %v", lastCommit)
			}
			// going in reverse order, oldest to newest
			for _, commitInfo := range commitInfos.CommitInfo {
				if err := p.createJobForCommitInfo(commitInfo); err != nil {
					return err
				}
			}
		}
	}
}

// TODO(pedge): implement
func (p *pipelineController) createJobForCommitInfo(commitInfo *pfs.CommitInfo) error {
	return nil
}
