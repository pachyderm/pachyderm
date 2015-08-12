package run

import (
	"github.com/pachyderm/pachyderm/src/pkg/graph"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/source"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

type Runner interface {
	Start(pipelineSource *pps.PipelineSource) (string, error)
}

func NewRunner(
	sourcer source.Sourcer,
	grapher graph.Grapher,
	containerClient container.Client,
	storeClient store.Client,
	timer timing.Timer,
) Runner {
	return newRunner(
		sourcer,
		grapher,
		containerClient,
		storeClient,
		timer,
	)
}
