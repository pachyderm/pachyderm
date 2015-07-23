package run

import (
	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/graph"
	"github.com/pachyderm/pachyderm/src/pps/source"
)

type runner struct {
	sourcer         source.Sourcer
	grapher         graph.Grapher
	containerClient container.Client
}

func newRunner(
	sourcer source.Sourcer,
	grapher graph.Grapher,
	containerClient container.Client,
) *runner {
	return &runner{
		sourcer,
		grapher,
		containerClient,
	}
}

func (r *runner) Run(pipelineSource *pps.PipelineSource) error {
	dirPath, pipeline, err := r.sourcer.GetDirPathAndPipeline(pipelineSource)
	if err != nil {
		return err
	}
	pipelineInfo, err := r.grapher.GetPipelineInfo(pipeline)
	if err != nil {
		return err
	}
	log.Printf("%v %v %v\n", dirPath, pipeline, pipelineInfo)
	return nil
}
