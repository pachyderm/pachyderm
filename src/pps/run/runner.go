package run

import (
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
	return nil
}
