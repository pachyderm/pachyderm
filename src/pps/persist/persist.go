package persist // import "go.pachyderm.com/pachyderm/src/pps/persist"

func NewLocalAPIClient(apiServer APIServer) APIClient {
	return newLocalAPIClient(apiServer)
}
