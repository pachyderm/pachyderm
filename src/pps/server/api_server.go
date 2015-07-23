package server

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pkg/clone"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/parse"
	"github.com/peter-edge/go-google-protobuf"
	"golang.org/x/net/context"
)

var (
	getVersionResponseInstance = &pps.GetVersionResponse{
		Version: &pps.Version{
			Major:      common.MajorVersion,
			Minor:      common.MinorVersion,
			Micro:      common.MicroVersion,
			Additional: common.AdditionalVersion,
		},
	}
)

type apiServer struct{}

func newAPIServer() *apiServer {
	return &apiServer{}
}

func (a *apiServer) GetVersion(ctx context.Context, empty *google_protobuf.Empty) (*pps.GetVersionResponse, error) {
	return getVersionResponseInstance, nil
}

func (a *apiServer) GetPipeline(ctx context.Context, getPipelineRequest *pps.GetPipelineRequest) (*pps.GetPipelineResponse, error) {
	var pipeline *pps.Pipeline
	if getPipelineRequest.PipelineSource.GithubPipelineSource != nil {
		dirPath, err := ioutil.TempDir("", "pachyderm")
		if err != nil {
			return nil, err
		}
		if err := clone.GithubClone(
			dirPath,
			getPipelineRequest.PipelineSource.GithubPipelineSource.User,
			getPipelineRequest.PipelineSource.GithubPipelineSource.Repository,
			getPipelineRequest.PipelineSource.GithubPipelineSource.Branch,
			"",
			getPipelineRequest.PipelineSource.GithubPipelineSource.AccessToken,
		); err != nil {
			return nil, err
		}
		pipeline, err = parse.NewParser().ParsePipeline(filepath.Clean(filepath.Join(dirPath, getPipelineRequest.PipelineSource.GithubPipelineSource.ContextDir)))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("must specify pipeline source")
	}
	return &pps.GetPipelineResponse{
		Pipeline: pipeline,
	}, nil
}
