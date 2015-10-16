package server // import "go.pachyderm.com/pachyderm/src/pps/watch/server"

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pachyderm.com/pachyderm/src/pps/watch"
)

func NewAPIServer(
	pfsAPIClient pfs.ApiClient,
	persistAPIClient persist.APIClient,
) watch.APIServer {
	return newAPIServer(
		pfsAPIClient,
		persistAPIClient,
	)
}
