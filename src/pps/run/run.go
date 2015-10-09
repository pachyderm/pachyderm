package run //import "go.pachyderm.com/pachyderm/src/pps/run"

import (
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps/store"
	"go.pedge.io/pkg/graph"
	"go.pedge.io/pkg/time"
)

type Runner interface {
	Start(pipelineRunID string) error
}

func NewRunner(
	grapher pkggraph.Grapher,
	containerClient container.Client,
	storeClient store.Client,
	timer pkgtime.Timer,
) Runner {
	return newRunner(
		grapher,
		containerClient,
		storeClient,
		timer,
	)
}
