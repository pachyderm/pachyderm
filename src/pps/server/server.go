package server // import "go.pachyderm.com/pachyderm/src/pps/server"

import "go.pachyderm.com/pachyderm/src/pps"

func NewAPIServer() pps.APIServer {
	return newLogAPIServer(newAPIServer())
}

func NewLocalAPIClient(apiServer pps.APIServer) pps.APIClient {
	return newLocalAPIClient(apiServer)
}
