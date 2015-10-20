package server

import (
	"fmt"

	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/protolog"
)

type pipelineController struct {
	pfsAPIClient     pfs.APIClient
	persistAPIClient persist.APIClient

	pipeline        *persist.Pipeline
	cancelC         chan bool
	finishedCancelC chan bool
}

func newPipelineController(
	pfsAPIClient pfs.APIClient,
	persistAPIClient persist.APIClient,
	pipeline *persist.Pipeline,
) *pipelineController {
	return &pipelineController{
		pfsAPIClient,
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
	// TODO(pedge): use InitialCommitID
	lastCommitID := "scratch"
	if len(jobs) > 0 {
		lastJob := jobs[0]
		if len(lastJob.JobInput) == 0 {
			return fmt.Errorf("pachyderm.pps.watch.server: had job with no JobInput, this is not currently allowed, %v", lastJob)
		}
		if len(lastJob.JobInput) > 0 {
			return fmt.Errorf("pachyderm.pps.watch.server: had job with more than one JobInput, this is not currently allowed, %v", lastJob)
		}
		jobInput := lastJob.JobInput[0]
		if jobInput.GetHostDir() != "" {
			return fmt.Errorf("pachyderm.pps.watch.server: had job with host dir set, this is not allowed, %v", lastJob)
		}
		if jobInput.GetCommit() == nil {
			return fmt.Errorf("pachyderm.pps.watch.server: had job without commit set, this is not allowed, %v", lastJob)
		}
		lastCommitID = jobInput.GetCommit().Id
	}
	go func() {
		if err := p.run(lastCommitID); err != nil {
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

func (p *pipelineController) run(lastCommitID string) error {
	for {
		select {
		case <-p.cancelC:
			p.finishedCancelC <- true
			close(p.finishedCancelC)
			return nil
		default:
		}
	}
}
