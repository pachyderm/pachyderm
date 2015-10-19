package server

import (
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
)

func jobToPersist(job *pps.Job) *persist.Job {
	persistJob := &persist.Job{
		Id:        job.Id,
		JobInput:  job.JobInput,
		JobOutput: job.JobOutput,
		// TODO(pedge): this relies on the current implementation where
		// the watch implementation directly calls persist api client,
		// if the watch implementation were to call the api client,
		// then this setting of the field would be incorrect
		CreatedFromWatch: false,
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

func persistToJobStatus(persistJobStatus *persist.JobStatus) *pps.JobStatus {
	return &pps.JobStatus{
		Type:      persistJobStatus.Type,
		Timestamp: persistJobStatus.Timestamp,
		Message:   persistJobStatus.Message,
	}
}

func pipelineToPersist(pipeline *pps.Pipeline) *persist.Pipeline {
	return &persist.Pipeline{
		Name:           pipeline.Name,
		Transform:      pipeline.Transform,
		PipelineInput:  pipeline.PipelineInput,
		PipelineOutput: pipeline.PipelineOutput,
	}
}

func persistToPipeline(persistPipeline *persist.Pipeline) *pps.Pipeline {
	return &pps.Pipeline{
		Name:           persistPipeline.Name,
		Transform:      persistPipeline.Transform,
		PipelineInput:  persistPipeline.PipelineInput,
		PipelineOutput: persistPipeline.PipelineOutput,
	}
}
