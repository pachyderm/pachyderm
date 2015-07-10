package pfs

type localAPIClient struct {
	apiServer ApiServer
}

func newLocalAPIClient(apiServer ApiServer) *localAPIClient {
	return &localAPIClient{
		apiServer,
	}
}
