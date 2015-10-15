package server // import "go.pachyderm.com/pachyderm/src/pps/server"

import (
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
)

func NewAPIServer(persistAPIClient persist.APIClient, containerClient container.Client) pps.APIServer {
	return newLogAPIServer(newAPIServer(persistAPIClient, containerClient))
}

func NewLocalAPIClient(apiServer pps.APIServer) pps.APIClient {
	return newLocalAPIClient(apiServer)
}
