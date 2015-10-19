package watch // import "go.pachyderm.com/pachyderm/src/pps/watch"

func NewLocalAPIClient(apiServer APIServer) APIClient {
	return newLocalAPIClient(apiServer)
}
