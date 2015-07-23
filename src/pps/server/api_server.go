package server

import (
	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/source"
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
	_, pipeline, err := source.NewSourcer().GetDirPathAndPipeline(getPipelineRequest.PipelineSource)
	if err != nil {
		return nil, err
	}
	return &pps.GetPipelineResponse{
		Pipeline: pipeline,
	}, nil
}
