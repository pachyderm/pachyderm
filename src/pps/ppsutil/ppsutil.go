package ppsutil

import (
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/peter-edge/go-google-protobuf"
	"golang.org/x/net/context"
)

func GetVersion(apiClient pps.ApiClient) (*pps.GetVersionResponse, error) {
	return apiClient.GetVersion(
		context.Background(),
		&google_protobuf.Empty{},
	)
}

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
) error {
	_, err := apiClient.StartPipelineRun(
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
	return err
}
