package pipelineserver

import (
	"fmt"

	"golang.org/x/net/context"

	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/protolog"
)

type pipelineController struct {
	pfsAPIClient     pfs.APIClient
	jobAPIClient     pps.JobAPIClient
	persistAPIClient persist.APIClient

	pipeline        *persist.Pipeline
	cancelC         chan bool
	finishedCancelC chan bool
}

func newPipelineController(
	pfsAPIClient pfs.APIClient,
	jobAPIClient pps.JobAPIClient,
	persistAPIClient persist.APIClient,
	pipeline *persist.Pipeline,
) *pipelineController {
	return &pipelineController{
		pfsAPIClient,
		jobAPIClient,
		persistAPIClient,
		pipeline,
		make(chan bool),
		make(chan bool),
	}
}

func (p *pipelineController) Start() error {
	// TODO(pedge): do not get all jobs each time, need a limit call on persist, more
	// generally, need all persist calls to have a limit
	jobs, err := getJobsByPipelineName(p.persistAPIClient, p.pipeline.Name)
	if err != nil {
		return err
	}
	repo, err := getRepoForPipeline(p.pipeline)
	if err != nil {
		return err
	}
	lastCommit := &pfs.Commit{
		Repo: repo,
		// TODO(pedge): use initial commit id when moved to pfs package
		Id: "scratch",
	}
	if len(jobs) > 0 {
		lastCommit, err = getCommitForJob(jobs[0])
		if err != nil {
			return err
		}
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
					Repo:  lastCommit.Repo,
					From:  lastCommit,
					Block: true,
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
				if err := p.createAndStartJobForCommitInfo(commitInfo); err != nil {
					return err
				}
			}
		}
	}
}

// TODO(pedge): implement
func (p *pipelineController) createAndStartJobForCommitInfo(commitInfo *pfs.CommitInfo) error {
	return nil
}

func getRepoForPipeline(pipeline *persist.Pipeline) (*pfs.Repo, error) {
	if len(pipeline.PipelineInput) == 0 {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: had pipeline with no PipelineInput, this is not currently allowed, %v", pipeline)
	}
	if len(pipeline.PipelineInput) > 0 {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: had pipeline with more than one PipelineInput, this is not currently allowed, %v", pipeline)
	}
	pipelineInput := pipeline.PipelineInput[0]
	if pipelineInput.GetHostDir() != "" {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: had pipeline with host dir set, this is not allowed, %v", pipeline)
	}
	if pipelineInput.GetRepo() == nil {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: had pipeline without repo set, this is not allowed, %v", pipeline)
	}
	return pipelineInput.GetRepo(), nil
}

func getCommitForJob(job *persist.Job) (*pfs.Commit, error) {
	if len(job.JobInput) == 0 {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: had job with no JobInput, this is not currently allowed, %v", job)
	}
	if len(job.JobInput) > 0 {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: had job with more than one JobInput, this is not currently allowed, %v", job)
	}
	jobInput := job.JobInput[0]
	if jobInput.GetHostDir() != "" {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: had job with host dir set, this is not allowed, %v", job)
	}
	if jobInput.GetCommit() == nil {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: had job without commit set, this is not allowed, %v", job)
	}
	return jobInput.GetCommit(), nil
}
