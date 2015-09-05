package ppsutil

import (
	"github.com/pachyderm/pachyderm/src/pps"
	"golang.org/x/net/context"
)

func GetPipeline(
	apiClient pps.ApiClient,
	pipelineSourceID string,
) (*pps.GetPipelineResponse, error) {
	return apiClient.GetPipeline(
		context.Background(),
		&pps.GetPipelineRequest{
			PipelineSourceId: pipelineSourceID,
		},
	)
}

func StartPipelineRun(
	apiClient pps.ApiClient,
	pipelineRunID string,
) error {
	_, err := apiClient.StartPipelineRun(
		context.Background(),
		&pps.StartPipelineRunRequest{
			PipelineRunId: pipelineRunID,
		},
	)
	return err
}

func GetPipelineRunStatus(
	apiClient pps.ApiClient,
	pipelineRunID string,
	all bool,
) (*pps.GetPipelineRunStatusResponse, error) {
	return apiClient.GetPipelineRunStatus(
		context.Background(),
		&pps.GetPipelineRunStatusRequest{
			PipelineRunId: pipelineRunID,
			All:           all,
		},
	)
}

func GetPipelineRunLogs(
	apiClient pps.ApiClient,
	pipelineRunID string,
	node string,
) (*pps.GetPipelineRunLogsResponse, error) {
	return apiClient.GetPipelineRunLogs(
		context.Background(),
		&pps.GetPipelineRunLogsRequest{
			PipelineRunId: pipelineRunID,
			Node:          node,
		},
	)
}
