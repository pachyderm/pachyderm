package run

import (
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/graph"
	"github.com/pachyderm/pachyderm/src/pps/source"
)

type Runner interface {
	Run(pipelineSource *pps.PipelineSource) error
}

func NewRunner(
	sourcer source.Sourcer,
	grapher graph.Grapher,
	containerClient container.Client,
) Runner {
	return newRunner(
		sourcer,
		grapher,
		containerClient,
	)
}
