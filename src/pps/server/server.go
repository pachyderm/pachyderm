package server // import "go.pachyderm.com/pachyderm/src/pps/server"

import (
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
)

func NewAPIServer(persistAPIClient persist.APIClient) pps.APIServer {
	return newLogAPIServer(newAPIServer(persistAPIClient))
}

func NewLocalAPIClient(apiServer pps.APIServer) pps.APIClient {
	return newLocalAPIClient(apiServer)
}
