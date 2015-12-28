package pipelineserver

import (
	"fmt"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/protolog"
	"golang.org/x/net/context"
)

type apiServer struct {
	protorpclog.Logger
	pfsAPIClient     pfs.APIClient
	jobAPIClient     pps.JobAPIClient
	persistAPIClient persist.APIClient

	cancelFuncs map[pps.Pipeline]func()
	lock        sync.Mutex
}

func newAPIServer(
	pfsAPIClient pfs.APIClient,
	jobAPIClient pps.JobAPIClient,
	persistAPIClient persist.APIClient,
) *apiServer {
	return &apiServer{
		protorpclog.NewLogger("pachyderm.pps.PipelineAPI"),
		pfsAPIClient,
		jobAPIClient,
		persistAPIClient,
		make(map[pps.Pipeline]func()),
		sync.Mutex{},
	}
}

func (a *apiServer) Start() error {
	pipelineInfos, err := a.ListPipeline(context.Background(), &pps.ListPipelineRequest{})
	if err != nil {
		return err
	}
	for _, pipelineInfo := range pipelineInfos.PipelineInfo {
		pipelineInfo := pipelineInfo
		go func() {
			if err := a.runPipeline(pipelineInfo); err != nil {
				protolog.Printf("pipeline errored: %s", err.Error())
			}
		}()
	}
	return nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	if request.Pipeline == nil {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: request.Pipeline cannot be nil")
	}
	persistPipelineInfo := &persist.PipelineInfo{
		PipelineName: request.Pipeline.Name,
		Transform:    request.Transform,
		Shards:       request.Shards,
		InputRepo:    request.InputRepo,
		OutputRepo:   request.OutputRepo,
	}
	if _, err := a.persistAPIClient.CreatePipelineInfo(ctx, persistPipelineInfo); err != nil {
		return nil, err
	}
	go func() {
		if err := a.runPipeline(persistPipelineInfoToPipelineInfo(persistPipelineInfo)); err != nil {
			protolog.Printf("pipeline errored: %s", err.Error())
		}
	}()
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectPipeline(ctx context.Context, request *pps.InspectPipelineRequest) (response *pps.PipelineInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelineInfo, err := a.persistAPIClient.GetPipelineInfo(ctx, request.Pipeline)
	if err != nil {
		return nil, err
	}
	return persistPipelineInfoToPipelineInfo(persistPipelineInfo), nil
}

func (a *apiServer) ListPipeline(ctx context.Context, request *pps.ListPipelineRequest) (response *pps.PipelineInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelineInfos, err := a.persistAPIClient.ListPipelineInfos(ctx, google_protobuf.EmptyInstance)
	if err != nil {
		return nil, err
	}
	pipelineInfos := make([]*pps.PipelineInfo, len(persistPipelineInfos.PipelineInfo))
	for i, persistPipelineInfo := range persistPipelineInfos.PipelineInfo {
		pipelineInfos[i] = persistPipelineInfoToPipelineInfo(persistPipelineInfo)
	}
	return &pps.PipelineInfos{
		PipelineInfo: pipelineInfos,
	}, nil
}

func (a *apiServer) DeletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) (response *google_protobuf.Empty, err error) {
	if _, err := a.persistAPIClient.DeletePipelineInfo(ctx, request.Pipeline); err != nil {
		return nil, err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	a.cancelFuncs[*request.Pipeline]()
	delete(a.cancelFuncs, *request.Pipeline)
	return google_protobuf.EmptyInstance, nil
}

func persistPipelineInfoToPipelineInfo(persistPipelineInfo *persist.PipelineInfo) *pps.PipelineInfo {
	return &pps.PipelineInfo{
		Pipeline: &pps.Pipeline{
			Name: persistPipelineInfo.PipelineName,
		},
		Transform:  persistPipelineInfo.Transform,
		Shards:     persistPipelineInfo.Shards,
		InputRepo:  persistPipelineInfo.InputRepo,
		OutputRepo: persistPipelineInfo.OutputRepo,
	}
}

func (a *apiServer) runPipeline(pipelineInfo *pps.PipelineInfo) error {
	ctx, cancel := context.WithCancel(context.Background())
	a.lock.Lock()
	a.cancelFuncs[*pipelineInfo.Pipeline] = cancel
	a.lock.Unlock()
	var loopErr error
	//TODO this gets really weird with branching... we need to figure out what that looks like.
	repoToLastCommit := make(map[pfs.Repo]*pfs.Commit)
	var lock sync.Mutex
	var wg sync.WaitGroup
	for _, inputRepo := range pipelineInfo.InputRepo {
		inputRepo := inputRepo
		wg.Add(1)
		go func() {
			defer wg.Done()
			var lastCommit *pfs.Commit
			for {
				listCommitRequest := &pfs.ListCommitRequest{
					Repo:       inputRepo,
					CommitType: pfs.CommitType_COMMIT_TYPE_READ,
					FromCommit: []*pfs.Commit{lastCommit},
					Block:      true,
				}
				commitInfos, err := a.pfsAPIClient.ListCommit(ctx, listCommitRequest)
				if err != nil && loopErr == nil {
					loopErr = err
					return
				}
				for _, commitInfo := range commitInfos.CommitInfo {
					lock.Lock()
					repoToLastCommit[*inputRepo] = commitInfo.Commit
					lastCommit = commitInfo.Commit
					var commits []*pfs.Commit
					for _, commit := range repoToLastCommit {
						commits = append(commits, commit)
					}
					lock.Unlock()
					if len(commits) < len(pipelineInfo.InputRepo) {
						// we don't yet have a commit for every input repo so there's no way to run the job
						// TODO is this actually the right policy? maybe we should run with empty commits
						continue
					}
					outParentCommit, err := a.bestParent(pipelineInfo, commitInfo)
					if err != nil && loopErr == nil {
						loopErr = err
						return
					}
					protolog.Printf("running job: repoToLastCommit: %+v, outParentCommit: %+v", repoToLastCommit, outParentCommit)
					_, err = a.jobAPIClient.CreateJob(
						ctx,
						&pps.CreateJobRequest{
							Transform:    pipelineInfo.Transform,
							Pipeline:     pipelineInfo.Pipeline,
							Shards:       pipelineInfo.Shards,
							InputCommit:  []*pfs.Commit{commitInfo.Commit},
							OutputParent: outParentCommit,
						},
					)
				}
			}
		}()
	}
	wg.Wait()
	return loopErr
}

func (a *apiServer) bestParent(pipelineInfo *pps.PipelineInfo, inputCommitInfo *pfs.CommitInfo) (*pfs.Commit, error) {
	for {
		jobInfos, err := a.jobAPIClient.ListJob(
			context.Background(),
			&pps.ListJobRequest{
				Pipeline: pipelineInfo.Pipeline,
				Input:    inputCommitInfo.Commit,
			},
		)
		if err != nil {
			return nil, err
		}
		// newest to oldest assumed
		for _, jobInfo := range jobInfos.JobInfo {
			if jobInfo.OutputCommit != nil {
				outputCommitInfo, err := a.pfsAPIClient.InspectCommit(context.TODO(), &pfs.InspectCommitRequest{Commit: jobInfo.OutputCommit})
				if err != nil {
					return nil, err
				}
				if outputCommitInfo.CommitType == pfs.CommitType_COMMIT_TYPE_READ {
					return outputCommitInfo.Commit, nil
				}
			}
		}
		if inputCommitInfo.ParentCommit == nil {
			return &pfs.Commit{Repo: pipelineInfo.OutputRepo}, nil
		}
		inputCommitInfo, err = a.pfsAPIClient.InspectCommit(context.TODO(), &pfs.InspectCommitRequest{Commit: inputCommitInfo.ParentCommit})
		if err != nil {
			return nil, err
		}
	}
}
