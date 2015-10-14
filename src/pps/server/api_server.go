package server

import (
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
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
	return nil, nil
}

func (a *apiServer) StartJob(ctx context.Context, request *pps.StartJobRequest) (response *google_protobuf.Empty, err error) {
	return nil, nil
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
