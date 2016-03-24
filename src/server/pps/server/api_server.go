package server

import (
	"fmt"
	"sync"
	"time"

	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/shard"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"
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
	outputCommit *pfsclient.Commit // the output commit
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
	hasher           *ppsserver.Hasher
	router           shard.Router
	pfsAddress       string
	pfsAPIClient     pfsclient.APIClient
	pfsClientOnce    sync.Once
	persistAPIServer persist.APIServer
	kubeClient       *kube.Client
	jobStates        map[string]*jobState
	jobStatesLock    sync.Mutex
	cancelFuncs      map[ppsclient.Pipeline]func()
	cancelFuncsLock  sync.Mutex
	version          int64
	// versionLock protects the version field.
	// versionLock must be held BEFORE reading from version and UNTIL all
	// requests using version have returned
	versionLock sync.RWMutex
}

func (a *apiServer) CreateJob(ctx context.Context, request *ppsclient.CreateJobRequest) (response *ppsclient.Job, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	defer func() {
		if retErr == nil {
			metrics.AddJobs(1)
		}
	}()
	if request.Shards == 0 {
		return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: request.Shards cannot be 0")
	}
	repoSet := make(map[string]bool)
	for _, input := range request.Inputs {
		repoSet[input.Commit.Repo.Name] = true
	}
	if len(repoSet) < len(request.Inputs) {
		return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: duplicate repo in job")
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
		return nil, fmt.Errorf("pachyderm.ppsclient.jobserver: no job backend")
	}
	_, err := a.persistAPIServer.CreateJobInfo(ctx, persistJobInfo)
	if err != nil {
		return nil, err
	}
	defer func() {
		if retErr != nil {
			if _, err := a.persistAPIServer.CreateJobState(ctx, &persist.JobState{
				JobID: persistJobInfo.JobID,
				State: ppsclient.JobState_JOB_STATE_FAILURE,
			}); err != nil {
				protolion.Errorf("error from CreateJobState %s", err.Error())
			}
		}
	}()
	if _, err := a.kubeClient.Jobs(api.NamespaceDefault).Create(job(persistJobInfo)); err != nil {
		return nil, err
	}
	return &ppsclient.Job{
		ID: persistJobInfo.JobID,
	}, nil
}

func (a *apiServer) InspectJob(ctx context.Context, request *ppsclient.InspectJobRequest) (response *ppsclient.JobInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistJobInfo, err := a.persistAPIServer.InspectJob(ctx, request)
	if err != nil {
		return nil, err
	}
	return newJobInfo(persistJobInfo)
}

func (a *apiServer) ListJob(ctx context.Context, request *ppsclient.ListJobRequest) (response *ppsclient.JobInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistJobInfos, err := a.persistAPIServer.ListJobInfos(ctx, request)
	if err != nil {
		return nil, err
	}
	jobInfos := make([]*ppsclient.JobInfo, len(persistJobInfos.JobInfo))
	for i, persistJobInfo := range persistJobInfos.JobInfo {
		jobInfo, err := newJobInfo(persistJobInfo)
		if err != nil {
			return nil, err
		}
		jobInfos[i] = jobInfo
	}
	return &ppsclient.JobInfos{
		JobInfo: jobInfos,
	}, nil
}

func (a *apiServer) StartJob(ctx context.Context, request *ppsserver.StartJobRequest) (response *ppsserver.StartJobResponse, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	inspectJobRequest := &ppsclient.InspectJobRequest{Job: request.Job}
	jobInfo, err := a.persistAPIServer.InspectJob(ctx, inspectJobRequest)
	if err != nil {
		return nil, err
	}
	if jobInfo.Transform == nil {
		return nil, fmt.Errorf("jobInfo.Transform should not be nil (this is likely a bug)")
	}
	a.jobStatesLock.Lock()
	jobState, ok := a.jobStates[request.Job.ID]
	if !ok {
		jobState = newJobState()
		a.jobStates[request.Job.ID] = jobState
	}
	shard := jobState.start
	if jobState.start < jobInfo.Shards {
		jobState.start++
	}
	a.jobStatesLock.Unlock()
	if shard == jobInfo.Shards {
		return nil, fmt.Errorf("job %s already has %d shards", request.Job.ID, jobInfo.Shards)
	}
	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return nil, err
	}
	var parentJobInfo *persist.JobInfo
	if jobInfo.ParentJob != nil {
		inspectJobRequest := &ppsclient.InspectJobRequest{Job: jobInfo.ParentJob}
		parentJobInfo, err = a.persistAPIServer.InspectJob(ctx, inspectJobRequest)
		if err != nil {
			return nil, err
		}
	}
	repoToFromCommit := make(map[string]*pfsclient.Commit)
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
		startCommitRequest := &pfsclient.StartCommitRequest{}
		if parentJobInfo == nil || reduce {
			if jobInfo.PipelineName == "" {
				startCommitRequest.Repo = ppsserver.JobRepo(request.Job)
				if _, err := pfsAPIClient.CreateRepo(ctx, &pfsclient.CreateRepoRequest{Repo: startCommitRequest.Repo}); err != nil {
					return nil, err
				}
			} else {
				startCommitRequest.Repo = ppsserver.PipelineRepo(&ppsclient.Pipeline{Name: jobInfo.PipelineName})
			}
		} else {
			startCommitRequest.Repo = parentJobInfo.OutputCommit.Repo
			startCommitRequest.ParentID = parentJobInfo.OutputCommit.ID
		}
		commit, err := pfsAPIClient.StartCommit(ctx, startCommitRequest)
		if err != nil {
			return nil, err
		}
		if _, err := a.persistAPIServer.CreateJobOutput(
			ctx,
			&persist.JobOutput{
				JobID:        request.Job.ID,
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
			Shard: &pfsclient.Shard{
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
	return &ppsserver.StartJobResponse{
		Transform:    jobInfo.Transform,
		CommitMounts: commitMounts,
		OutputCommit: jobState.outputCommit,
		Index:        shard,
	}, nil
}

func (a *apiServer) FinishJob(ctx context.Context, request *ppsserver.FinishJobRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	inspectJobRequest := &ppsclient.InspectJobRequest{Job: request.Job}
	jobInfo, err := a.persistAPIServer.InspectJob(ctx, inspectJobRequest)
	if err != nil {
		return nil, err
	}
	var finished bool
	persistJobState := ppsclient.JobState_JOB_STATE_FAILURE
	if err := func() error {
		a.jobStatesLock.Lock()
		defer a.jobStatesLock.Unlock()
		jobState, ok := a.jobStates[request.Job.ID]
		if !ok {
			return fmt.Errorf("job %s was never started", request.Job.ID)
		}
		jobState.success = jobState.success && request.Success
		if jobState.success {
			persistJobState = ppsclient.JobState_JOB_STATE_SUCCESS
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
		if _, err := pfsAPIClient.FinishCommit(ctx, &pfsclient.FinishCommitRequest{
			Commit: jobInfo.OutputCommit,
		}); err != nil {
			return nil, err
		}
		if _, err := a.persistAPIServer.CreateJobState(ctx, &persist.JobState{
			JobID: request.Job.ID,
			State: persistJobState,
		}); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *ppsclient.CreatePipelineRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	defer func() {
		if retErr == nil {
			metrics.AddPipelines(1)
		}
	}()
	pfsAPIClient, err := a.getPfsClient()
	if request.Pipeline == nil {
		return nil, fmt.Errorf("pachyderm.ppsclient.pipelineserver: request.Pipeline cannot be nil")
	}
	repoSet := make(map[string]bool)
	for _, input := range request.Inputs {
		if _, err := pfsAPIClient.InspectRepo(ctx, &pfsclient.InspectRepoRequest{Repo: input.Repo}); err != nil {
			return nil, err
		}
		repoSet[input.Repo.Name] = true
	}
	if len(repoSet) < len(request.Inputs) {
		return nil, fmt.Errorf("pachyderm.ppsclient.pipelineserver: duplicate input repos")
	}
	repo := ppsserver.PipelineRepo(request.Pipeline)
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
	if _, err := pfsAPIClient.CreateRepo(ctx, &pfsclient.CreateRepoRequest{Repo: repo}); err != nil {
		return nil, err
	}
	go func() {
		if err := a.runPipeline(newPipelineInfo(persistPipelineInfo)); err != nil {
			protolion.Printf("pipeline errored: %s", err.Error())
		}
	}()
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectPipeline(ctx context.Context, request *ppsclient.InspectPipelineRequest) (response *ppsclient.PipelineInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelineInfo, err := a.persistAPIServer.GetPipelineInfo(ctx, request.Pipeline)
	if err != nil {
		return nil, err
	}
	return newPipelineInfo(persistPipelineInfo), nil
}

func (a *apiServer) ListPipeline(ctx context.Context, request *ppsclient.ListPipelineRequest) (response *ppsclient.PipelineInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelineInfos, err := a.persistAPIServer.ListPipelineInfos(ctx, google_protobuf.EmptyInstance)
	if err != nil {
		return nil, err
	}
	pipelineInfos := make([]*ppsclient.PipelineInfo, len(persistPipelineInfos.PipelineInfo))
	for i, persistPipelineInfo := range persistPipelineInfos.PipelineInfo {
		pipelineInfos[i] = newPipelineInfo(persistPipelineInfo)
	}
	return &ppsclient.PipelineInfos{
		PipelineInfo: pipelineInfos,
	}, nil
}

func (a *apiServer) DeletePipeline(ctx context.Context, request *ppsclient.DeletePipelineRequest) (response *google_protobuf.Empty, err error) {
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
	pipelineInfos, err := a.ListPipeline(context.Background(), &ppsclient.ListPipelineRequest{})
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

func newPipelineInfo(persistPipelineInfo *persist.PipelineInfo) *ppsclient.PipelineInfo {
	return &ppsclient.PipelineInfo{
		Pipeline: &ppsclient.Pipeline{
			Name: persistPipelineInfo.PipelineName,
		},
		Transform:  persistPipelineInfo.Transform,
		Shards:     persistPipelineInfo.Shards,
		Inputs:     persistPipelineInfo.Inputs,
		OutputRepo: persistPipelineInfo.OutputRepo,
	}
}

func (a *apiServer) runPipeline(pipelineInfo *ppsclient.PipelineInfo) error {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFuncsLock.Lock()
	a.cancelFuncs[*pipelineInfo.Pipeline] = cancel
	a.cancelFuncsLock.Unlock()
	repoToLeaves := make(map[string]map[string]bool)
	repoToInput := make(map[string]*ppsclient.PipelineInput)
	var inputRepos []*pfsclient.Repo
	for _, input := range pipelineInfo.Inputs {
		repoToLeaves[input.Repo.Name] = make(map[string]bool)
		repoToInput[input.Repo.Name] = input
		inputRepos = append(inputRepos, &pfsclient.Repo{Name: input.Repo.Name})
	}
	pfsAPIClient, err := a.getPfsClient()
	if err != nil {
		return err
	}
	for {
		var fromCommits []*pfsclient.Commit
		for repo, leaves := range repoToLeaves {
			for leaf := range leaves {
				fromCommits = append(
					fromCommits,
					&pfsclient.Commit{
						Repo: &pfsclient.Repo{Name: repo},
						ID:   leaf,
					})
			}
		}
		listCommitRequest := &pfsclient.ListCommitRequest{
			Repo:       inputRepos,
			CommitType: pfsclient.CommitType_COMMIT_TYPE_READ,
			FromCommit: fromCommits,
			Block:      true,
		}
		commitInfos, err := pfsAPIClient.ListCommit(ctx, listCommitRequest)
		if err != nil {
			return err
		}
		for _, commitInfo := range commitInfos.CommitInfo {
			repoToLeaves[commitInfo.Commit.Repo.Name][commitInfo.Commit.ID] = true
			if commitInfo.ParentCommit != nil {
				delete(repoToLeaves[commitInfo.ParentCommit.Repo.Name], commitInfo.ParentCommit.ID)
			}
			// generate all the permutations of leaves we could use this commit with
			commitSets := [][]*pfsclient.Commit{[]*pfsclient.Commit{}}
			for repoName, leaves := range repoToLeaves {
				if repoName == commitInfo.Commit.Repo.Name {
					continue
				}
				var newCommitSets [][]*pfsclient.Commit
				for _, commitSet := range commitSets {
					for leaf := range leaves {
						newCommitSet := make([]*pfsclient.Commit, len(commitSet)+1)
						copy(newCommitSet, commitSet)
						newCommitSet[len(commitSet)] = &pfsclient.Commit{
							Repo: &pfsclient.Repo{Name: repoName},
							ID:   leaf,
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
				var parentJob *ppsclient.Job
				if commitInfo.ParentCommit != nil {
					parentJob, err = a.parentJob(ctx, pipelineInfo, commitSet, commitInfo)
					if err != nil {
						return err
					}
				}
				var inputs []*ppsclient.JobInput
				for _, commit := range append(commitSet, commitInfo.Commit) {
					inputs = append(inputs, &ppsclient.JobInput{
						Commit: commit,
						Reduce: repoToInput[commit.Repo.Name].Reduce,
					})
				}
				if _, err = a.CreateJob(
					ctx,
					&ppsclient.CreateJobRequest{
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
	pipelineInfo *ppsclient.PipelineInfo,
	commitSet []*pfsclient.Commit,
	newCommit *pfsclient.CommitInfo,
) (*ppsclient.Job, error) {
	jobInfo, err := a.ListJob(
		ctx,
		&ppsclient.ListJobRequest{
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

func (a *apiServer) getPfsClient() (pfsclient.APIClient, error) {
	if a.pfsAPIClient == nil {
		var onceErr error
		a.pfsClientOnce.Do(func() {
			clientConn, err := grpc.Dial(a.pfsAddress, grpc.WithInsecure())
			if err != nil {
				onceErr = err
			}
			a.pfsAPIClient = pfsclient.NewAPIClient(clientConn)
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return a.pfsAPIClient, nil
}

func newJobInfo(persistJobInfo *persist.JobInfo) (*ppsclient.JobInfo, error) {
	job := &ppsclient.Job{ID: persistJobInfo.JobID}
	return &ppsclient.JobInfo{
		Job:          job,
		Transform:    persistJobInfo.Transform,
		Pipeline:     &ppsclient.Pipeline{Name: persistJobInfo.PipelineName},
		Shards:       persistJobInfo.Shards,
		Inputs:       persistJobInfo.Inputs,
		ParentJob:    persistJobInfo.ParentJob,
		CreatedAt:    persistJobInfo.CreatedAt,
		OutputCommit: persistJobInfo.OutputCommit,
		State:        persistJobInfo.State,
	}, nil
}

func job(jobInfo *persist.JobInfo) *extensions.Job {
	app := jobInfo.JobID
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
			Name:   jobInfo.JobID,
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
					Name:   jobInfo.JobID,
					Labels: labels(app),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:    "user",
							Image:   image,
							Command: []string{"/job-shim", jobInfo.JobID},
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
