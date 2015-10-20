package server

import (
	"fmt"

	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

func getPipeline(persistAPIClient persist.APIClient, name string) (*persist.Pipeline, error) {
	pipelines, err := persistAPIClient.GetPipelinesByName(context.Background(), &google_protobuf.StringValue{Value: name})
	if err != nil {
		return nil, err
	}
	if len(pipelines.Pipeline) == 0 {
		return nil, fmt.Errorf("pachyderm.pps.watch.server: no piplines for name %s", name)
	}
	return pipelines.Pipeline[0], nil
}

func getAllPipelines(persistAPIClient persist.APIClient) ([]*persist.Pipeline, error) {
	protoPipelines, err := persistAPIClient.GetAllPipelines(context.Background(), emptyInstance)
	if err != nil {
		return nil, err
	}
	pipelineMap := make(map[string]*persist.Pipeline)
	for _, pipeline := range protoPipelines.Pipeline {
		// pipelines are ordered newest to oldest, so if we have already
		// seen a pipeline with the same name, it is newer
		if _, ok := pipelineMap[pipeline.Name]; !ok {
			pipelineMap[pipeline.Name] = pipeline
		}
	}
	pipelines := make([]*persist.Pipeline, len(pipelineMap))
	i := 0
	for _, pipeline := range pipelineMap {
		pipelines[i] = pipeline
		i++
	}
	return pipelines, nil
}

func getJobsByPipelineName(persistAPIClient persist.APIClient, name string) ([]*persist.Job, error) {
	pipelines, err := persistAPIClient.GetPipelinesByName(context.Background(), &google_protobuf.StringValue{Value: name})
	if err != nil {
		return nil, err
	}
	if len(pipelines.Pipeline) == 0 {
		return nil, fmt.Errorf("pachyderm.pps.watch.server: no piplines for name %s", name)
	}
	var jobs []*persist.Job
	for _, pipeline := range pipelines.Pipeline {
		protoJobs, err := persistAPIClient.GetJobsByPipelineID(context.Background(), &google_protobuf.StringValue{Value: pipeline.Id})
		if err != nil {
			return nil, err
		}
		if len(protoJobs.Job) > 0 {
			jobs = append(jobs, protoJobs.Job...)
		}
	}
	// TODO(pedge): sort by timestamp
	return jobs, nil
}
