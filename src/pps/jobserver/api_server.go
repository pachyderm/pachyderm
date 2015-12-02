package jobserver

import (
	"fmt"
	"sync"
	"time"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	kube "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
)

var (
	trueVal = true
	suite   = "pachyderm"
)

type apiServer struct {
	protorpclog.Logger
	pfsAPIClient     pfs.APIClient
	persistAPIClient persist.APIClient
	kubeClient       *kube.Client
	startJobCounter  map[pps.Job]uint64
	startJobLock     sync.Mutex
	finishJobCounter map[pps.Job]uint64
	finishJobLock    sync.Mutex
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
		make(map[pps.Job]uint64),
		sync.Mutex{},
		make(map[pps.Job]uint64),
		sync.Mutex{},
	}
}

func (a *apiServer) CreateJob(ctx context.Context, request *pps.CreateJobRequest) (response *pps.Job, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if request.OutputParent == nil {
		return nil, fmt.Errorf("pachyderm.pps.jobserver: request.OutputParent cannot be nil")
	}
	persistJobInfo := &persist.JobInfo{
		Shards:       request.Shards,
		InputCommit:  request.InputCommit,
		OutputParent: request.OutputParent,
	}
	if request.GetTransform() != nil {
		persistJobInfo.Spec = &persist.JobInfo_Transform{
			Transform: request.GetTransform(),
		}
	} else if request.GetPipeline() != nil {
		persistJobInfo.Spec = &persist.JobInfo_PipelineName{
			PipelineName: request.GetPipeline().Name,
		}
	} else {
		return nil, fmt.Errorf("pachyderm.pps.jobserver: both transform and pipeline are not set on %v", request)
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
	var shard uint64
	if err := func() error {
		a.startJobLock.Lock()
		defer a.startJobLock.Unlock()
		shard = a.startJobCounter[*request.Job]
		a.startJobCounter[*request.Job] = shard + 1
		if shard == 0 {
			commit, err := a.pfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{
				Parent: jobInfo.OutputParent,
			})
			if err != nil {
				return err
			}
			if _, err := a.persistAPIClient.CreateJobOutput(
				ctx,
				&persist.JobOutput{
					JobId:        request.Job.Id,
					OutputCommit: commit,
				}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return nil, err
	}
	jobOutput, err := a.persistAPIClient.GetJobOutput(ctx, request.Job)
	if err != nil {
		return nil, err
	}
	if jobOutput.OutputCommit == nil {
		return nil, fmt.Errorf("jobOutput.OutputCommit should not be nil (this is likely a bug)")
	}
	return &pps.StartJobResponse{
		InputCommit:  jobInfo.InputCommit,
		OutputCommit: jobOutput.OutputCommit,
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
	var finished uint64
	func() {
		a.finishJobLock.Lock()
		defer a.finishJobLock.Unlock()
		a.finishJobCounter[*request.Job] = a.finishJobCounter[*request.Job] + 1
		finished = a.finishJobCounter[*request.Job]
	}()
	if finished == jobInfo.Shards {
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
		Shards:      persistJobInfo.Shards,
		InputCommit: persistJobInfo.InputCommit,
	}
	if persistJobInfo.GetTransform() != nil {
		jobInfo.Spec = &pps.JobInfo_Transform{
			Transform: persistJobInfo.GetTransform(),
		}
	}
	if persistJobInfo.GetPipelineName() != "" {
		jobInfo.Spec = &pps.JobInfo_Pipeline{
			Pipeline: &pps.Pipeline{
				Name: persistJobInfo.GetPipelineName(),
			},
		}
	}
	persistJobOutput, err := a.persistAPIClient.GetJobOutput(ctx, job)
	if err != nil {
		return nil, err
	}
	if persistJobOutput != nil {
		jobInfo.OutputCommit = persistJobOutput.OutputCommit
	}
	return jobInfo, nil
}

func job(jobInfo *persist.JobInfo) *extensions.Job {
	app := jobInfo.JobId
	shards := int(jobInfo.Shards)
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
							Image:   "pachyderm/job-shim",
							Command: append([]string{"/job-shim", jobInfo.JobId}, jobInfo.GetTransform().Cmd...),
							SecurityContext: &api.SecurityContext{
								Privileged: &trueVal, // god is this dumb
							},
						},
					},
					RestartPolicy: "Never",
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
