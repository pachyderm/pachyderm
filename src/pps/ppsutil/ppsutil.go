package ppsutil

import (
	"github.com/pachyderm/pachyderm/src/pps"
	"golang.org/x/net/context"
)

func GetPipelineGithub(
	apiClient pps.ApiClient,
	contextDir string,
	user string,
	repository string,
	branch string,
	accessToken string,
) (*pps.GetPipelineResponse, error) {
	return apiClient.GetPipeline(
		context.Background(),
		&pps.GetPipelineRequest{
			PipelineSource: &pps.PipelineSource{
				GithubPipelineSource: &pps.GithubPipelineSource{
					ContextDir:  contextDir,
					User:        user,
					Repository:  repository,
					Branch:      branch,
					AccessToken: accessToken,
				},
			},
		},
	)
}

func StartPipelineRunGithub(
	apiClient pps.ApiClient,
	contextDir string,
	user string,
	repository string,
	branch string,
	accessToken string,
) (*pps.StartPipelineRunResponse, error) {
	return apiClient.StartPipelineRun(
		context.Background(),
		&pps.StartPipelineRunRequest{
			PipelineSource: &pps.PipelineSource{
				GithubPipelineSource: &pps.GithubPipelineSource{
					ContextDir:  contextDir,
					User:        user,
					Repository:  repository,
					Branch:      branch,
					AccessToken: accessToken,
				},
			},
		},
	)
}

func GetPipelineRunStatus(
	apiClient pps.ApiClient,
	pipelineRunID string,
) (*pps.GetPipelineRunStatusResponse, error) {
	return apiClient.GetPipelineRunStatus(
		context.Background(),
		&pps.GetPipelineRunStatusRequest{
			PipelineRunId: pipelineRunID,
		},
	)
}
