package run

import (
	"go.pachyderm.com/pachyderm/src/pkg/graph"
	"go.pachyderm.com/pachyderm/src/pkg/timing"
	"go.pachyderm.com/pachyderm/src/pps/container"
	"go.pachyderm.com/pachyderm/src/pps/store"
)

type Runner interface {
	Start(pipelineRunID string) error
}

func NewRunner(
	grapher graph.Grapher,
	containerClient container.Client,
	storeClient store.Client,
	timer timing.Timer,
) Runner {
	return newRunner(
		grapher,
		containerClient,
		storeClient,
		timer,
	)
}
