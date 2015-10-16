package watch // import "go.pachyderm.com/pachyderm/src/pps/watch"

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
)

func NewAPIServer(
	ppsAPIClient pps.APIClient,
	pfsAPIClient pfs.ApiClient,
	persistAPIClient persist.APIClient,
) APIServer {
	return newAPIServer(
		ppsAPIClient,
		pfsAPIClient,
		persistAPIClient,
	)
}

func NewLocalAPIClient(apiServer APIServer) APIClient {
	return newLocalAPIClient(apiServer)
}
