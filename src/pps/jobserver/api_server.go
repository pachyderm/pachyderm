package jobserver

import (
	"fmt"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/time"
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
}

func newJobState() *jobState {
	return &jobState{
		start:        0,
		finish:       0,
		outputCommit: nil,
		commitReady:  make(chan bool),
		finished:     make(chan bool),
	}
}

type apiServer struct {
	protorpclog.Logger
	pfsAPIClient     pfs.APIClient
	persistAPIClient persist.APIClient
	kubeClient       *kube.Client
	jobStates        map[string]*jobState
	lock             sync.Mutex
}

func newAPIServer(
	pfsAPIClient pfs.APIClient,
	persistAPIClient persist.APIClient,
	kubeClient *kube.Client,
) *apiServer {
	return &apiServer{
		protorpclog.NewLogger("pachyderm.pps.JobAPI"),
		pfsAPIClient,
		persistAPIClient,
		kubeClient,
		make(map[string]*jobState),
		sync.Mutex{},
	}
}

func (a *apiServer) CreateJob(ctx context.Context, request *pps.CreateJobRequest) (response *pps.Job, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if request.ParentJob == nil {
		return nil, fmt.Errorf("pachyderm.pps.jobserver: request.OutputParent cannot be nil")
	}
	if request.Shards == 0 {
		return nil, fmt.Errorf("pachyderm.pps.jobserver: request.Shards cannot be 0")
	}
	// TODO validate job to make sure input commits and output repo exist
	persistJobInfo := &persist.JobInfo{
		Shards:      request.Shards,
		Transform:   request.Transform,
		InputCommit: request.InputCommit,
		ParentJob:   request.ParentJob,
	}
	if request.Pipeline != nil {
		persistJobInfo.PipelineName = request.Pipeline.Name
	}
	persistJobInfo.JobId = uuid.NewWithoutDashes()
	persistJobInfo.CreatedAt = prototime.TimeToTimestamp(time.Now())
	if a.kubeClient == nil {
		return nil, fmt.Errorf("pachyderm.pps.jobserver: no job backend")
	}
	_, err := a.persistAPIClient.CreateJobInfo(ctx, persistJobInfo)
	if err != nil {
		return nil, err
	}
	if _, err := a.kubeClient.Jobs(api.NamespaceDefault).Create(job(persistJobInfo)); err != nil {
		return nil, err
	}
	return &pps.Job{
		Id: persistJobInfo.JobId,
	}, nil
}

func (a *apiServer) InspectJob(ctx context.Context, request *pps.InspectJobRequest) (response *pps.JobInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistJobInfo, err := a.persistAPIClient.GetJobInfo(ctx, request.Job)
	if err != nil {
		return nil, err
	}
	return a.persistJobInfoToJobInfo(ctx, persistJobInfo)
}

func (a *apiServer) ListJob(ctx context.Context, request *pps.ListJobRequest) (response *pps.JobInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	persistJobInfos, err := a.persistAPIClient.ListJobInfos(ctx, request)
	if err != nil {
		return nil, err
	}
	jobInfos := make([]*pps.JobInfo, len(persistJobInfos.JobInfo))
	for i, persistJobInfo := range persistJobInfos.JobInfo {
		jobInfo, err := a.persistJobInfoToJobInfo(ctx, persistJobInfo)
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
	jobInfo, err := a.persistAPIClient.GetJobInfo(ctx, request.Job)
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
		parentJobOutput, err := a.persistAPIClient.GetJobOutput(ctx, jobInfo.ParentJob)
		if err != nil {
			return nil, err
		}
		commit, err := a.pfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
			Parent: parentJobOutput.OutputCommit,
		})
		if err != nil {
			return nil, err
		}
		if _, err := a.persistAPIClient.CreateJobOutput(
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
	return &pps.StartJobResponse{
		Transform:    jobInfo.Transform,
		InputCommit:  jobInfo.InputCommit,
		OutputCommit: jobState.outputCommit,
		Shard: &pfs.Shard{
			Number:  shard,
			Modulus: jobInfo.Shards,
		},
	}, nil
}

func (a *apiServer) FinishJob(ctx context.Context, request *pps.FinishJobRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	jobInfo, err := a.persistAPIClient.GetJobInfo(ctx, request.Job)
	if err != nil {
		return nil, err
	}
	var jobState *jobState
	if err := func() error {
		a.lock.Lock()
		defer a.lock.Unlock()
		jobState = a.jobStates[request.Job.Id]
		if jobState == nil {
			return fmt.Errorf("job %s was never started", request.Job.Id)
		}
		jobState.finish++
		return nil
	}(); err != nil {
		return nil, err
	}
	if jobState.finish == jobInfo.Shards {
		// all of the shards have finished so we finish the commit
		jobOutput, err := a.persistAPIClient.GetJobOutput(ctx, request.Job)
		if err != nil {
			return nil, err
		}
		if jobOutput.OutputCommit == nil {
			return nil, fmt.Errorf("jobOutput.OutputCommit should not be nil (this is likely a bug)")
		}
		if _, err := a.pfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{
			Commit: jobOutput.OutputCommit,
		}); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) persistJobInfoToJobInfo(ctx context.Context, persistJobInfo *persist.JobInfo) (*pps.JobInfo, error) {
	job := &pps.Job{Id: persistJobInfo.JobId}
	jobInfo := &pps.JobInfo{
		Job:         job,
		Transform:   persistJobInfo.Transform,
		Pipeline:    &pps.Pipeline{Name: persistJobInfo.PipelineName},
		Shards:      persistJobInfo.Shards,
		InputCommit: persistJobInfo.InputCommit,
	}
	persistJobOutput, err := a.persistAPIClient.GetJobOutput(ctx, job)
	if err == nil && persistJobOutput != nil {
		jobInfo.OutputCommit = persistJobOutput.OutputCommit
	}
	return jobInfo, nil
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
