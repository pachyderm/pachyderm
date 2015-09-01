package ppsutil

import (
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pps"
	"golang.org/x/net/context"
)

const (
	registryDirectory = "grpcutil/registry/pps"
)

func NewPpsRegistry(discoveryClient discovery.Client) grpcutil.Registry {
	return grpcutil.NewRegistry(
		discovery.NewRegistry(
			discoveryClient,
			registryDirectory,
		),
	)
}

func NewPpsProvider(discoveryClient discovery.Client, dialer grpcutil.Dialer) grpcutil.Provider {
	return grpcutil.NewProvider(
		discovery.NewRegistry(
			discoveryClient,
			registryDirectory,
		),
		dialer,
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
