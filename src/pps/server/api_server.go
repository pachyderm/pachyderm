package server

import (
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

var (
	emptyInstance = &google_protobuf.Empty{}
)

type apiServer struct {
	persistAPIClient persist.APIClient
}

func newAPIServer(persistAPIClient persist.APIClient) *apiServer {
	return &apiServer{persistAPIClient}
}

func (a *apiServer) CreateJob(ctx context.Context, request *pps.CreateJobRequest) (response *pps.Job, err error) {
	persistJob, err := a.persistAPIClient.CreateJob(ctx, jobToPersist(request.Job))
	if err != nil {
		return nil, err
	}
	return persistToJob(persistJob), nil
}

func (a *apiServer) GetJob(ctx context.Context, request *pps.GetJobRequest) (response *pps.Job, err error) {
	persistJob, err := a.persistAPIClient.GetJobByID(ctx, &google_protobuf.StringValue{Value: request.JobId})
	if err != nil {
		return nil, err
	}
	return persistToJob(persistJob), nil
}

func (a *apiServer) GetJobsByPipelineName(ctx context.Context, request *pps.GetJobsByPipelineNameRequest) (response *pps.Jobs, err error) {
	persistPipelines, err := a.persistAPIClient.GetPipelinesByName(ctx, &google_protobuf.StringValue{Value: request.PipelineName})
	if err != nil {
		return nil, err
	}
	var jobs []*pps.Job
	for _, persistPipeline := range persistPipelines.Pipeline {
		persistJobs, err := a.persistAPIClient.GetJobsByPipelineID(ctx, &google_protobuf.StringValue{Value: persistPipeline.Id})
		if err != nil {
			return nil, err
		}
		iJobs := persistToJobs(persistJobs)
		jobs = append(jobs, iJobs.Job...)
	}
	return &pps.Jobs{
		Job: jobs,
	}, nil
}

func (a *apiServer) StartJob(ctx context.Context, request *pps.StartJobRequest) (response *google_protobuf.Empty, err error) {
	persistJob, err := a.persistAPIClient.GetJobByID(ctx, &google_protobuf.StringValue{Value: request.JobId})
	if err != nil {
		return nil, err
	}
	if err := a.startPersistJob(persistJob); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *apiServer) GetJobStatus(ctx context.Context, request *pps.GetJobStatusRequest) (response *pps.JobStatus, err error) {
	return nil, nil
}

func (a *apiServer) GetJobLogs(request *pps.GetJobLogsRequest, responseServer pps.API_GetJobLogsServer) (err error) {
	return nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *pps.Pipeline, err error) {
	return nil, nil
}

func (a *apiServer) GetPipeline(ctx context.Context, request *pps.GetPipelineRequest) (response *pps.Pipeline, err error) {
	return nil, nil
}

func (a *apiServer) GetAllPipelines(ctx context.Context, request *google_protobuf.Empty) (response *pps.Pipelines, err error) {
	return nil, nil
}

func (a *apiServer) startPersistJob(persistJob *persist.Job) error {
	return nil
}

func jobToPersist(job *pps.Job) *persist.Job {
	persistJob := &persist.Job{
		Id:        job.Id,
		JobInput:  job.JobInput,
		JobOutput: job.JobOutput,
	}
	if job.GetTransform() != nil {
		persistJob.Spec = &persist.Job_Transform{
			Transform: job.GetTransform(),
		}
	} else if job.GetPipelineId() != "" {
		persistJob.Spec = &persist.Job_PipelineId{
			PipelineId: job.GetPipelineId(),
		}
	}
	return persistJob
}

func persistToJob(persistJob *persist.Job) *pps.Job {
	job := &pps.Job{
		Id:        persistJob.Id,
		JobInput:  persistJob.JobInput,
		JobOutput: persistJob.JobOutput,
	}
	if persistJob.GetTransform() != nil {
		job.Spec = &pps.Job_Transform{
			Transform: persistJob.GetTransform(),
		}
	} else if persistJob.GetPipelineId() != "" {
		job.Spec = &pps.Job_PipelineId{
			PipelineId: persistJob.GetPipelineId(),
		}
	}
	return job
}

func jobsToPersist(jobs *pps.Jobs) *persist.Jobs {
	persistJobs := make([]*persist.Job, len(jobs.Job))
	for i, job := range jobs.Job {
		persistJobs[i] = jobToPersist(job)
	}
	return &persist.Jobs{
		Job: persistJobs,
	}
}

func persistToJobs(persistJobs *persist.Jobs) *pps.Jobs {
	jobs := make([]*pps.Job, len(persistJobs.Job))
	for i, persistJob := range persistJobs.Job {
		jobs[i] = persistToJob(persistJob)
	}
	return &pps.Jobs{
		Job: jobs,
	}
}
