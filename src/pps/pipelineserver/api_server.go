package pipelineserver

import (
	"fmt"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"go.pedge.io/lion/proto"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type apiServer struct {
	protorpclog.Logger
	pfsAddress       string
	pfsAPIClient     pfs.APIClient
	pfsClientOnce    sync.Once
	jobAPIClient     pps.JobAPIClient
	persistAPIServer persist.APIServer
	cancelFuncs      map[pps.Pipeline]func()
	lock             sync.Mutex
}

func newAPIServer(
	pfsAddress string,
	jobAPIClient pps.JobAPIClient,
	persistAPIServer persist.APIServer,
) *apiServer {
	return &apiServer{
		protorpclog.NewLogger("pachyderm.pps.PipelineAPI"),
		pfsAddress,
		nil,
		sync.Once{},
		jobAPIClient,
		persistAPIServer,
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
				protolion.Printf("pipeline errored: %s", err.Error())
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
	repoSet := make(map[string]bool)
	for _, input := range request.Inputs {
		repoSet[input.Repo.Name] = true
	}
	if len(repoSet) < len(request.Inputs) {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: duplicate input repos")
	}
	repo := pps.PipelineRepo(request.Pipeline)
	persistPipelineInfo := &persist.PipelineInfo{
		PipelineName: request.Pipeline.Name,
		Transform:    request.Transform,
		Shards:       request.Shards,
		Inputs:       request.Inputs,
		OutputRepo:   repo,
	}
	if _, err := a.persistAPIServer.CreatePipelineInfo(ctx, persistPipelineInfo); err != nil {
		return nil, err
	}
	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}
	if _, err := pfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: repo}); err != nil {
		return nil, err
	}
	go func() {
		if err := a.runPipeline(newPipelineInfo(persistPipelineInfo)); err != nil {
			protolion.Printf("pipeline errored: %s", err.Error())
		}
	}()
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectPipeline(ctx context.Context, request *pps.InspectPipelineRequest) (response *pps.PipelineInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelineInfo, err := a.persistAPIServer.GetPipelineInfo(ctx, request.Pipeline)
	if err != nil {
		return nil, err
	}
	return newPipelineInfo(persistPipelineInfo), nil
}

func (a *apiServer) ListPipeline(ctx context.Context, request *pps.ListPipelineRequest) (response *pps.PipelineInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelineInfos, err := a.persistAPIServer.ListPipelineInfos(ctx, google_protobuf.EmptyInstance)
	if err != nil {
		return nil, err
	}
	pipelineInfos := make([]*pps.PipelineInfo, len(persistPipelineInfos.PipelineInfo))
	for i, persistPipelineInfo := range persistPipelineInfos.PipelineInfo {
		pipelineInfos[i] = newPipelineInfo(persistPipelineInfo)
	}
	return &pps.PipelineInfos{
		PipelineInfo: pipelineInfos,
	}, nil
}

func (a *apiServer) DeletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) (response *google_protobuf.Empty, err error) {
	if _, err := a.persistAPIServer.DeletePipelineInfo(ctx, request.Pipeline); err != nil {
		return nil, err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	a.cancelFuncs[*request.Pipeline]()
	delete(a.cancelFuncs, *request.Pipeline)
	return google_protobuf.EmptyInstance, nil
}

func newPipelineInfo(persistPipelineInfo *persist.PipelineInfo) *pps.PipelineInfo {
	return &pps.PipelineInfo{
		Pipeline: &pps.Pipeline{
			Name: persistPipelineInfo.PipelineName,
		},
		Transform:  persistPipelineInfo.Transform,
		Shards:     persistPipelineInfo.Shards,
		Inputs:     persistPipelineInfo.Inputs,
		OutputRepo: persistPipelineInfo.OutputRepo,
	}
}

func (a *apiServer) runPipeline(pipelineInfo *pps.PipelineInfo) error {
	ctx, cancel := context.WithCancel(context.Background())
	a.lock.Lock()
	a.cancelFuncs[*pipelineInfo.Pipeline] = cancel
	a.lock.Unlock()
	repoToLeaves := make(map[string]map[string]bool)
	repoToInput := make(map[string]*pps.PipelineInput)
	var inputRepos []*pfs.Repo
	for _, input := range pipelineInfo.Inputs {
		repoToLeaves[input.Repo.Name] = make(map[string]bool)
		repoToInput[input.Repo.Name] = input
		inputRepos = append(inputRepos, &pfs.Repo{Name: input.Repo.Name})
	}
	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return err
	}
	for {
		var fromCommits []*pfs.Commit
		for repo, leaves := range repoToLeaves {
			for leaf := range leaves {
				fromCommits = append(
					fromCommits,
					&pfs.Commit{
						Repo: &pfs.Repo{Name: repo},
						Id:   leaf,
					})
			}
		}
		listCommitRequest := &pfs.ListCommitRequest{
			Repo:       inputRepos,
			CommitType: pfs.CommitType_COMMIT_TYPE_READ,
			FromCommit: fromCommits,
			Block:      true,
		}
		commitInfos, err := pfsAPIClient.ListCommit(ctx, listCommitRequest)
		if err != nil {
			return err
		}
		for _, commitInfo := range commitInfos.CommitInfo {
			repoToLeaves[commitInfo.Commit.Repo.Name][commitInfo.Commit.Id] = true
			if commitInfo.ParentCommit != nil {
				delete(repoToLeaves[commitInfo.ParentCommit.Repo.Name], commitInfo.ParentCommit.Id)
			}
			// generate all the pemrutations of leaves we could use this commit with
			commitSets := [][]*pfs.Commit{[]*pfs.Commit{}}
			for repoName, leaves := range repoToLeaves {
				if repoName == commitInfo.Commit.Repo.Name {
					continue
				}
				var newCommitSets [][]*pfs.Commit
				for _, commitSet := range commitSets {
					for leaf := range leaves {
						newCommitSet := make([]*pfs.Commit, len(commitSet)+1)
						copy(newCommitSet, commitSet)
						newCommitSet[len(commitSet)] = &pfs.Commit{
							Repo: &pfs.Repo{Name: repoName},
							Id:   leaf,
						}
						newCommitSets = append(newCommitSets, newCommitSet)
					}
				}
				commitSets = newCommitSets
			}
			for _, commitSet := range commitSets {
				// + 1 as the commitSet doesn't contain the commit we just got
				if len(commitSet)+1 < len(pipelineInfo.Inputs) {
					continue
				}
				var parentJob *pps.Job
				if commitInfo.ParentCommit != nil {
					parentJob, err = a.parentJob(ctx, pipelineInfo, append(commitSet, commitInfo.ParentCommit), commitInfo)
					if err != nil {
						return err
					}
				}
				var inputs []*pps.JobInput
				for _, commit := range append(commitSet, commitInfo.Commit) {
					inputs = append(inputs, &pps.JobInput{
						Commit: commit,
						Reduce: repoToInput[commit.Repo.Name].Reduce,
					})
				}
				if _, err = a.jobAPIClient.CreateJob(
					ctx,
					&pps.CreateJobRequest{
						Transform: pipelineInfo.Transform,
						Pipeline:  pipelineInfo.Pipeline,
						Shards:    pipelineInfo.Shards,
						Inputs:    inputs,
						ParentJob: parentJob,
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

func (a *apiServer) getPfsClient() (pfs.APIClient, error) {
	if a.pfsAPIClient == nil {
		var onceErr error
		a.pfsClientOnce.Do(func() {
			clientConn, err := grpc.Dial(a.pfsAddress, grpc.WithInsecure())
			if err != nil {
				onceErr = err
			}
			a.pfsAPIClient = pfs.NewAPIClient(clientConn)
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return a.pfsAPIClient, nil
}
