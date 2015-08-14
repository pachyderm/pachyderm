package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

func NewAPIServer(pfsAPIClient pfs.ApiClient, storeClient store.Client, timer timing.Timer) pps.ApiServer {
	return newAPIServer(pfsAPIClient, storeClient, timer)
}
