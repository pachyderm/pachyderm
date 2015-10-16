package pps // import "go.pachyderm.com/pachyderm/src/pps"

func NewLocalAPIClient(apiServer APIServer) APIClient {
	return newLocalAPIClient(apiServer)
}
