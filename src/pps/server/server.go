package server // import "go.pachyderm.com/pachyderm/src/pps/server"

import (
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pachyderm.com/pachyderm/src/pps/watch"
)

func NewAPIServer(
	persistAPIClient persist.APIClient,
	watchAPIClient watch.APIClient,
	containerClient container.Client,
) pps.APIServer {
	return newLogAPIServer(newAPIServer(persistAPIClient, watchAPIClient, containerClient))
}
