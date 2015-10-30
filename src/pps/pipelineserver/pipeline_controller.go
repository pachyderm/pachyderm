package pipelineserver

import (
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"go.pedge.io/protolog"
)

type pipelineController struct {
	pfsAPIClient      pfs.APIClient
	jobAPIClient      pps.JobAPIClient
	pipelineAPIClient pps.PipelineAPIClient
	pipelineInfo      *pps.PipelineInfo

	ctx       context.Context
	cancel    context.CancelFunc
	waitGroup *sync.WaitGroup
}

func newPipelineController(
	pfsAPIClient pfs.APIClient,
	jobAPIClient pps.JobAPIClient,
	pipelineAPIClient pps.PipelineAPIClient,
	pipelineInfo *pps.PipelineInfo,
) *pipelineController {
	ctx, cancel := context.WithCancel(context.Background())
	return &pipelineController{
		pfsAPIClient,
		jobAPIClient,
		pipelineAPIClient,
		pipelineInfo,
		ctx,
		cancel,
		&sync.WaitGroup{},
	}
}

func (p *pipelineController) Start() error {
	// TODO: do not get all jobs each time, need a limit call on persist, more
	// generally, need all persist calls to have a limit
	jobInfos, err := p.jobAPIClient.ListJob(context.Background(), &pps.ListJobRequest{Pipeline: p.pipelineInfo.Pipeline})
	if err != nil {
		return err
	}
	lastCommit := &pfs.Commit{
		Repo: p.pipelineInfo.Input,
		// TODO: use initial commit id when moved to pfs package
		Id: "scratch",
	}
	if len(jobInfos.JobInfo) > 0 {
		lastCommit = jobInfos.JobInfo[0].Input
	}
	p.waitGroup.Add(1)
	go func() {
		defer p.waitGroup.Done()
		if err := p.run(lastCommit); ignoreCanceledError(err) != nil {
			// TODO: what to do with error?
			protolog.Errorln(err.Error())
		}
	}()
	return nil
}

func (p *pipelineController) Cancel() error {
	p.cancel()
	// does not block until run is complete, but run will be in the process of cancelling
	<-p.ctx.Done()
	// wait until run completes
	p.waitGroup.Wait()
	return ignoreCanceledError(p.ctx.Err())
}

func (p *pipelineController) run(lastCommit *pfs.Commit) error {
	for {
		// http://blog.golang.org/context
		commitErrorPairC := make(chan commitErrorPair, 1)
		go func() { commitErrorPairC <- p.runInner(p.ctx, lastCommit) }()
		select {
		case <-p.ctx.Done():
			_ = <-commitErrorPairC
			return ignoreCanceledError(p.ctx.Err())
		case commitErrorPair := <-commitErrorPairC:
			if ignoreCanceledError(commitErrorPair.Err) != nil {
				return commitErrorPair.Err
			}
			lastCommit = commitErrorPair.Commit
		}
	}
}

type commitErrorPair struct {
	Commit *pfs.Commit
	Err    error
}

func (p *pipelineController) runInner(ctx context.Context, lastCommit *pfs.Commit) commitErrorPair {
	commitInfos, err := p.pfsAPIClient.ListCommit(
		ctx,
		&pfs.ListCommitRequest{
			Repo:       lastCommit.Repo,
			CommitType: pfs.CommitType_COMMIT_TYPE_READ,
			From:       lastCommit,
			Block:      true,
		},
	)
	if err != nil {
		return commitErrorPair{Err: err}
	}
	if len(commitInfos.CommitInfo) == 0 {
		return commitErrorPair{Err: fmt.Errorf("pachyderm.pps.pipelineserver: we expected at least one *pfs.CommitInfo returned from blocking call, but no *pfs.CommitInfo structs were returned for %v", lastCommit)}
	}
	// going in reverse order, oldest to newest
	for _, commitInfo := range commitInfos.CommitInfo {
		if err := p.createJobForCommitInfo(ctx, commitInfo); err != nil {
			return commitErrorPair{Err: err}
		}
	}
	return commitErrorPair{Commit: commitInfos.CommitInfo[len(commitInfos.CommitInfo)-1].Commit}
}

func (p *pipelineController) createJobForCommitInfo(ctx context.Context, commitInfo *pfs.CommitInfo) error {
	parentOutputCommit, err := p.getParentOutputCommit(ctx, commitInfo)
	if err != nil {
		return err
	}
	_, err = p.jobAPIClient.CreateJob(
		ctx,
		&pps.CreateJobRequest{
			Spec: &pps.CreateJobRequest_Pipeline{
				Pipeline: p.pipelineInfo.Pipeline,
			},
			Input:        commitInfo.Commit,
			OutputParent: parentOutputCommit,
		},
	)
	return err
}

func (p *pipelineController) getParentOutputCommit(ctx context.Context, commitInfo *pfs.CommitInfo) (*pfs.Commit, error) {
	for commitInfo.ParentCommit != nil && commitInfo.ParentCommit.Id != pfs.InitialCommitID {
		outputCommit, err := p.getOutputCommit(ctx, commitInfo.ParentCommit)
		if err != nil {
			return nil, err
		}
		if outputCommit != nil {
			return outputCommit, nil
		}
	}
	return &pfs.Commit{
		Repo: commitInfo.Commit.Repo,
		Id:   pfs.InitialCommitID,
	}, nil
}

func (p *pipelineController) getOutputCommit(ctx context.Context, inputCommit *pfs.Commit) (*pfs.Commit, error) {
	jobInfos, err := p.jobAPIClient.ListJob(
		ctx,
		&pps.ListJobRequest{
			Pipeline: p.pipelineInfo.Pipeline,
			Input:    inputCommit,
		},
	)
	if err != nil {
		return nil, err
	}
	// newest to oldest assumed
	for _, jobInfo := range jobInfos.JobInfo {
		if jobInfo.Output != nil && containsSuccessJobStatus(jobInfo.JobStatus) {
			return jobInfo.Output, nil
		}
	}
	return nil, nil
}

// TODO: not assuming that last status is success
func containsSuccessJobStatus(jobStatuses []*pps.JobStatus) bool {
	for _, jobStatus := range jobStatuses {
		if jobStatus.Type == pps.JobStatusType_JOB_STATUS_TYPE_SUCCESS {
			return true
		}
	}
	return false
}

func ignoreCanceledError(err error) error {
	if err != context.Canceled && grpc.Code(err) != codes.Canceled {
		return err
	}
	return nil
}
