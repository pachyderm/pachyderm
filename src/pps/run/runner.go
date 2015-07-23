package run

import (
	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/graph"
	"github.com/pachyderm/pachyderm/src/pps/source"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

type runner struct {
	sourcer         source.Sourcer
	grapher         graph.Grapher
	containerClient container.Client
	storeClient     store.Client
}

func newRunner(
	sourcer source.Sourcer,
	grapher graph.Grapher,
	containerClient container.Client,
	storeClient store.Client,
) *runner {
	return &runner{
		sourcer,
		grapher,
		containerClient,
		storeClient,
	}
}

func (r *runner) Run(pipelineSource *pps.PipelineSource) (string, error) {
	dirPath, pipeline, err := r.sourcer.GetDirPathAndPipeline(pipelineSource)
	if err != nil {
		return "", err
	}
	pipelineInfo, err := r.grapher.GetPipelineInfo(pipeline)
	if err != nil {
		return "", err
	}
	log.Printf("%v %v %v\n", dirPath, pipeline, pipelineInfo)
	return "1234", nil
}
