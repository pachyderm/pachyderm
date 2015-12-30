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
	}
	if _, err := a.persistAPIClient.CreatePipelineInfo(ctx, persistPipelineInfo); err != nil {
		return nil, err
	}
	repo := pps.PipelineRepo(request.Pipeline)
	if _, err := a.pfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: repo}); err != nil {
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
		Transform: persistPipelineInfo.Transform,
		Shards:    persistPipelineInfo.Shards,
		InputRepo: persistPipelineInfo.InputRepo,
	}
}

func (a *apiServer) runPipeline(pipelineInfo *pps.PipelineInfo) error {
	ctx, cancel := context.WithCancel(context.Background())
	a.lock.Lock()
	a.cancelFuncs[*pipelineInfo.Pipeline] = cancel
	a.lock.Unlock()
	repoToBranches := make(map[string]map[string]bool)
	for _, inputRepo := range pipelineInfo.InputRepo {
		repoToBranches[inputRepo.Name] = make(map[string]bool)
	}
	for {
		var fromCommits []*pfs.Commit
		for repo, branches := range repoToBranches {
			for branch := range branches {
				fromCommits = append(
					fromCommits,
					&pfs.Commit{
						Repo: &pfs.Repo{Name: repo},
						Id:   branch,
					})
			}
		}
		listCommitRequest := &pfs.ListCommitRequest{
			Repo:       pipelineInfo.InputRepo,
			CommitType: pfs.CommitType_COMMIT_TYPE_READ,
			FromCommit: fromCommits,
			Block:      true,
		}
		commitInfos, err := a.pfsAPIClient.ListCommit(ctx, listCommitRequest)
		if err != nil {
			return err
		}
		for _, commitInfo := range commitInfos.CommitInfo {
			repoToBranches[commitInfo.Commit.Repo.Name][commitInfo.Commit.Id] = true
			if commitInfo.ParentCommit != nil {
				delete(repoToBranches[commitInfo.ParentCommit.Repo.Name], commitInfo.ParentCommit.Id)
			}
			// generate all the pemrutations of branches we could use this commit with
			commitSets := [][]*pfs.Commit{[]*pfs.Commit{}}
			for repoName, branches := range repoToBranches {
				if repoName == commitInfo.Commit.Repo.Name {
					continue
				}
				var newCommitSets [][]*pfs.Commit
				for _, commitSet := range commitSets {
					for branch := range branches {
						newCommitSet := make([]*pfs.Commit, len(commitSet)+1)
						copy(newCommitSet, commitSet)
						newCommitSet[len(commitSet)] = &pfs.Commit{
							Repo: &pfs.Repo{Name: repoName},
							Id:   branch,
						}
						newCommitSets = append(newCommitSets, newCommitSet)
					}
				}
				commitSets = newCommitSets
			}
			for _, commitSet := range commitSets {
				// + 1 as the commitSet doesn't contain the commit we just got
				if len(commitSet)+1 < len(pipelineInfo.InputRepo) {
					continue
				}
				var parentJob *pps.Job
				if commitInfo.ParentCommit != nil {
					parentJob, err = a.parentJob(ctx, pipelineInfo, append(commitSet, commitInfo.ParentCommit), commitInfo)
					if err != nil {
						return err
					}
				}
				if _, err = a.jobAPIClient.CreateJob(
					ctx,
					&pps.CreateJobRequest{
						Transform:   pipelineInfo.Transform,
						Pipeline:    pipelineInfo.Pipeline,
						Shards:      pipelineInfo.Shards,
						InputCommit: append(commitSet, commitInfo.Commit),
						ParentJob:   parentJob,
					},
				); err != nil {
					return err
				}
			}
		}
	}
}

func (a *apiServer) parentJob(
	ctx context.Context,
	pipelineInfo *pps.PipelineInfo,
	commitSet []*pfs.Commit,
	newCommit *pfs.CommitInfo,
) (*pps.Job, error) {
	jobInfo, err := a.jobAPIClient.ListJob(
		ctx,
		&pps.ListJobRequest{
			Pipeline:    pipelineInfo.Pipeline,
			InputCommit: append(commitSet, newCommit.ParentCommit),
		})
	if err != nil {
		return nil, err
	}
	if len(jobInfo.JobInfo) == 0 {
		return nil, nil
	}
	return jobInfo.JobInfo[0].Job, nil
}
