package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"go.pedge.io/lion/proto"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

var (
	trueVal = true
	suite   = "pachyderm"
)

type jobState struct {
	start        uint64      // the number of shards started
	finish       uint64      // the number of shards finished
	outputCommit *pfs.Commit // the output commit
	commitReady  chan bool   // closed when outCommit has been started (and is non nil)
	finished     chan bool   // closed when the job has been finished, the jobState will be deleted afterward
	success      bool
}

func newJobState() *jobState {
	return &jobState{
		start:        0,
		finish:       0,
		outputCommit: nil,
		commitReady:  make(chan bool),
		finished:     make(chan bool),
		success:      true,
	}
}

type apiServer struct {
	protorpclog.Logger
	hasher           *pps.Hasher
	router           shard.Router
	pfsAddress       string
	pfsAPIClient     pfs.APIClient
	pfsClientOnce    sync.Once
	persistAPIServer persist.APIServer
	kubeClient       *kube.Client
	jobStates        map[string]*jobState
	jobStatesLock    sync.Mutex
	cancelFuncs      map[pps.Pipeline]func()
	cancelFuncsLock  sync.Mutex
	version          int64
	// versionLock protects the version field.
	// versionLock must be held BEFORE reading from version and UNTIL all
	// requests using version have returned
	versionLock sync.RWMutex
}

func (a *apiServer) CreateJob(ctx context.Context, request *pps.CreateJobRequest) (response *pps.Job, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if request.Shards == 0 {
		return nil, fmt.Errorf("pachyderm.pps.jobserver: request.Shards cannot be 0")
	}
	repoSet := make(map[string]bool)
	for _, input := range request.Inputs {
		repoSet[input.Commit.Repo.Name] = true
	}
	if len(repoSet) < len(request.Inputs) {
		return nil, fmt.Errorf("pachyderm.pps.jobserver: duplicate repo in job")
	}
	// TODO validate job to make sure input commits and output repo exist
	persistJobInfo := &persist.JobInfo{
		Shards:    request.Shards,
		Transform: request.Transform,
		Inputs:    request.Inputs,
		ParentJob: request.ParentJob,
	}
	if request.Pipeline != nil {
		persistJobInfo.PipelineName = request.Pipeline.Name
	}
	if a.kubeClient == nil {
		return nil, fmt.Errorf("pachyderm.pps.jobserver: no job backend")
	}
	_, err := a.persistAPIServer.CreateJobInfo(ctx, persistJobInfo)
	if err != nil {
		return nil, err
	}
	defer func() {
		if retErr != nil {
			if _, err := a.persistAPIServer.CreateJobState(ctx, &persist.JobState{
				JobId: persistJobInfo.JobId,
				State: pps.JobState_JOB_STATE_FAILURE,
			}); err != nil {
				protolion.Errorf("error from CreateJobState %s", err.Error())
			}
		}
	}()
	if _, err := a.kubeClient.Jobs(api.NamespaceDefault).Create(job(persistJobInfo)); err != nil {
		return nil, err
	}
	return &pps.Job{
		Id: persistJobInfo.JobId,
	}, nil
}

func (a *apiServer) InspectJob(ctx context.Context, request *pps.InspectJobRequest) (response *pps.JobInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistJobInfo, err := a.persistAPIServer.InspectJob(ctx, request)
	if err != nil {
		return nil, err
	}
	return newJobInfo(persistJobInfo)
}

func (a *apiServer) ListJob(ctx context.Context, request *pps.ListJobRequest) (response *pps.JobInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistJobInfos, err := a.persistAPIServer.ListJobInfos(ctx, request)
	if err != nil {
		return nil, err
	}
	jobInfos := make([]*pps.JobInfo, len(persistJobInfos.JobInfo))
	for i, persistJobInfo := range persistJobInfos.JobInfo {
		jobInfo, err := newJobInfo(persistJobInfo)
		if err != nil {
			return nil, err
		}
		jobInfos[i] = jobInfo
	}
	return &pps.JobInfos{
		JobInfo: jobInfos,
	}, nil
}

func (a *apiServer) StartJob(ctx context.Context, request *pps.StartJobRequest) (response *pps.StartJobResponse, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	inspectJobRequest := &pps.InspectJobRequest{Job: request.Job}
	jobInfo, err := a.persistAPIServer.InspectJob(ctx, inspectJobRequest)
	if err != nil {
		return nil, err
	}
	if jobInfo.Transform == nil {
		return nil, fmt.Errorf("jobInfo.Transform should not be nil (this is likely a bug)")
	}
	a.jobStatesLock.Lock()
	jobState, ok := a.jobStates[request.Job.Id]
	if !ok {
		jobState = newJobState()
		a.jobStates[request.Job.Id] = jobState
	}
	shard := jobState.start
	if jobState.start < jobInfo.Shards {
		jobState.start++
	}
	a.jobStatesLock.Unlock()
	if shard == jobInfo.Shards {
		return nil, fmt.Errorf("job %s already has %d shards", request.Job.Id, jobInfo.Shards)
	}
	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}
	var parentJobInfo *persist.JobInfo
	if jobInfo.ParentJob != nil {
		inspectJobRequest := &pps.InspectJobRequest{Job: jobInfo.ParentJob}
		parentJobInfo, err = a.persistAPIServer.InspectJob(ctx, inspectJobRequest)
		if err != nil {
			return nil, err
		}
	}
	repoToFromCommit := make(map[string]*pfs.Commit)
	reduce := false
	if parentJobInfo != nil {
		for _, jobInput := range parentJobInfo.Inputs {
			if jobInput.Reduce {
				reduce = true
			} else {
				// input isn't being reduced, do it incrementally
				repoToFromCommit[jobInput.Commit.Repo.Name] = jobInput.Commit
			}
		}
	}
	if shard == 0 {
		startCommitRequest := &pfs.StartCommitRequest{}
		if parentJobInfo == nil || reduce {
			if jobInfo.PipelineName == "" {
				startCommitRequest.Repo = pps.JobRepo(request.Job)
				if _, err := pfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: startCommitRequest.Repo}); err != nil {
					return nil, err
				}
			} else {
				startCommitRequest.Repo = pps.PipelineRepo(&pps.Pipeline{Name: jobInfo.PipelineName})
			}
		} else {
			startCommitRequest.Repo = parentJobInfo.OutputCommit.Repo
			startCommitRequest.ParentId = parentJobInfo.OutputCommit.Id
		}
		commit, err := pfsAPIClient.StartCommit(ctx, startCommitRequest)
		if err != nil {
			return nil, err
		}
		if _, err := a.persistAPIServer.CreateJobOutput(
			ctx,
			&persist.JobOutput{
				JobId:        request.Job.Id,
				OutputCommit: commit,
			}); err != nil {
			return nil, err
		}
		jobState.outputCommit = commit
		close(jobState.commitReady)
	}
	<-jobState.commitReady
	if jobState.outputCommit == nil {
		return nil, fmt.Errorf("jobState.outputCommit should not be nil (this is likely a bug)")
	}
	var commitMounts []*fuse.CommitMount
	for _, jobInput := range jobInfo.Inputs {
		commitMount := &fuse.CommitMount{
			Commit:     jobInput.Commit,
			FromCommit: repoToFromCommit[jobInput.Commit.Repo.Name],
			Shard: &pfs.Shard{
				FileModulus:  1,
				BlockModulus: 1,
			},
		}
		if jobInput.Reduce {
			commitMount.Shard.FileNumber = shard
			commitMount.Shard.FileModulus = jobInfo.Shards
		} else {
			commitMount.Shard.BlockNumber = shard
			commitMount.Shard.BlockModulus = jobInfo.Shards
		}
		commitMounts = append(commitMounts, commitMount)
	}
	outputCommitMount := &fuse.CommitMount{
		Commit: jobState.outputCommit,
		Alias:  "out",
	}
	commitMounts = append(commitMounts, outputCommitMount)
	return &pps.StartJobResponse{
		Transform:    jobInfo.Transform,
		CommitMounts: commitMounts,
		OutputCommit: jobState.outputCommit,
		Index:        shard,
	}, nil
}

func (a *apiServer) FinishJob(ctx context.Context, request *pps.FinishJobRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	inspectJobRequest := &pps.InspectJobRequest{Job: request.Job}
	jobInfo, err := a.persistAPIServer.InspectJob(ctx, inspectJobRequest)
	if err != nil {
		return nil, err
	}
	var finished bool
	persistJobState := pps.JobState_JOB_STATE_FAILURE
	if err := func() error {
		a.jobStatesLock.Lock()
		defer a.jobStatesLock.Unlock()
		jobState, ok := a.jobStates[request.Job.Id]
		if !ok {
			return fmt.Errorf("job %s was never started", request.Job.Id)
		}
		jobState.success = jobState.success && request.Success
		if jobState.success {
			persistJobState = pps.JobState_JOB_STATE_SUCCESS
		}
		jobState.finish++
		finished = (jobState.finish == jobInfo.Shards)
		return nil
	}(); err != nil {
		return nil, err
	}
	if finished {
		if jobInfo.OutputCommit == nil {
			return nil, fmt.Errorf("jobInfo.OutputCommit should not be nil (this is likely a bug)")
		}
		pfsAPIClient, err := a.getPfsClient()
		if err != nil {
			return nil, err
		}
		if _, err := pfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
			Commit: jobInfo.OutputCommit,
		}); err != nil {
			return nil, err
		}
		if _, err := a.persistAPIServer.CreateJobState(ctx, &persist.JobState{
			JobId: request.Job.Id,
			State: persistJobState,
		}); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	pfsAPIClient, err := a.getPfsClient()
	if request.Pipeline == nil {
		return nil, fmt.Errorf("pachyderm.pps.pipelineserver: request.Pipeline cannot be nil")
	}
	repoSet := make(map[string]bool)
	for _, input := range request.Inputs {
		if _, err := pfsAPIClient.InspectRepo(ctx, &pfs.InspectRepoRequest{Repo: input.Repo}); err != nil {
			return nil, err
		}
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
	a.cancelFuncsLock.Lock()
	defer a.cancelFuncsLock.Unlock()
	a.cancelFuncs[*request.Pipeline]()
	delete(a.cancelFuncs, *request.Pipeline)
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) Version(version int64) error {
	a.versionLock.Lock()
	defer a.versionLock.Unlock()
	a.version = version
	return nil
}

func (a *apiServer) AddShard(shard uint64, version int64) error {
	pipelineInfos, err := a.ListPipeline(context.Background(), &pps.ListPipelineRequest{})
	if err != nil {
		return err
	}
	for _, pipelineInfo := range pipelineInfos.PipelineInfo {
		if a.hasher.HashPipeline(pipelineInfo.Pipeline) == shard {
			pipelineInfo := pipelineInfo
			go func() {
				if err := a.runPipeline(pipelineInfo); err != nil {
					protolion.Printf("pipeline errored: %s", err.Error())
				}
			}()
		}
	}
	return nil
}

func (a *apiServer) RemoveShard(shard uint64, version int64) error {
	a.cancelFuncsLock.Lock()
	defer a.cancelFuncsLock.Unlock()
	for pipeline, cancelFunc := range a.cancelFuncs {
		if a.hasher.HashPipeline(&pipeline) == shard {
			cancelFunc()
		}
	}
	return nil
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
	a.cancelFuncsLock.Lock()
	a.cancelFuncs[*pipelineInfo.Pipeline] = cancel
	a.cancelFuncsLock.Unlock()
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
			// generate all the permutations of leaves we could use this commit with
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
					parentJob, err = a.parentJob(ctx, pipelineInfo, commitSet, commitInfo)
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
				if _, err = a.CreateJob(
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
	jobInfo, err := a.ListJob(
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

func newJobInfo(persistJobInfo *persist.JobInfo) (*pps.JobInfo, error) {
	job := &pps.Job{Id: persistJobInfo.JobId}
	return &pps.JobInfo{
		Job:          job,
		Transform:    persistJobInfo.Transform,
		Pipeline:     &pps.Pipeline{Name: persistJobInfo.PipelineName},
		Shards:       persistJobInfo.Shards,
		Inputs:       persistJobInfo.Inputs,
		ParentJob:    persistJobInfo.ParentJob,
		CreatedAt:    persistJobInfo.CreatedAt,
		OutputCommit: persistJobInfo.OutputCommit,
		State:        persistJobInfo.State,
	}, nil
}

func job(jobInfo *persist.JobInfo) *extensions.Job {
	app := jobInfo.JobId
	shards := int(jobInfo.Shards)
	image := "pachyderm/job-shim"
	if jobInfo.Transform.Image != "" {
		image = jobInfo.Transform.Image
	}
	return &extensions.Job{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   jobInfo.JobId,
			Labels: labels(app),
		},
		Spec: extensions.JobSpec{
			Selector: &unversioned.LabelSelector{
				MatchLabels: labels(app),
			},
			Parallelism: &shards,
			Completions: &shards,
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Name:   jobInfo.JobId,
					Labels: labels(app),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:    "user",
							Image:   image,
							Command: []string{"/job-shim", jobInfo.JobId},
							SecurityContext: &api.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
						},
					},
					RestartPolicy: "OnFailure",
				},
			},
		},
	}
}

func labels(app string) map[string]string {
	return map[string]string{
		"app":   app,
		"suite": suite,
	}
}
