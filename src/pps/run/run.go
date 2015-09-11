package run

import (
	"github.com/pachyderm/pachyderm/src/pkg/graph"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/store"
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
