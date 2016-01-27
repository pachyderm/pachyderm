package jobserver

import (
	"fmt"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/fuse"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"go.pedge.io/lion/proto"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
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
	pfsAPIClient     pfs.APIClient
	persistAPIServer persist.APIServer
	kubeClient       *kube.Client
	jobStates        map[string]*jobState
	lock             sync.Mutex
}

func newAPIServer(
	pfsAPIClient pfs.APIClient,
	persistAPIServer persist.APIServer,
	kubeClient *kube.Client,
) *apiServer {
	return &apiServer{
		protorpclog.NewLogger("pachyderm.pps.JobAPI"),
		pfsAPIClient,
		persistAPIServer,
		kubeClient,
		make(map[string]*jobState),
		sync.Mutex{},
	}
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
	a.lock.Lock()
	jobState, ok := a.jobStates[request.Job.Id]
	if !ok {
		jobState = newJobState()
		a.jobStates[request.Job.Id] = jobState
	}
	shard := jobState.start
	if jobState.start < jobInfo.Shards {
		jobState.start++
	}
	a.lock.Unlock()
	if shard == jobInfo.Shards {
		return nil, fmt.Errorf("job %s already has %d shards", request.Job.Id, jobInfo.Shards)
	}
	if shard == 0 {
		var parentCommit *pfs.Commit
		if jobInfo.ParentJob == nil {
			var repo *pfs.Repo
			if jobInfo.PipelineName == "" {
				repo = pps.JobRepo(request.Job)
				if _, err := a.pfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: repo}); err != nil {
					return nil, err
				}
			} else {
				repo = pps.PipelineRepo(&pps.Pipeline{Name: jobInfo.PipelineName})
			}
			parentCommit = &pfs.Commit{Repo: repo}
		} else {
			inspectJobRequest := &pps.InspectJobRequest{Job: jobInfo.ParentJob}
			parentJobInfo, err := a.persistAPIServer.InspectJob(ctx, inspectJobRequest)
			if err != nil {
				return nil, err
			}
			parentCommit = parentJobInfo.OutputCommit
		}
		commit, err := a.pfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
			Parent: parentCommit,
		})
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
			Commit: jobInput.Commit,
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
		a.lock.Lock()
		defer a.lock.Unlock()
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
		if _, err := a.pfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
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
			Selector: &extensions.PodSelector{
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
