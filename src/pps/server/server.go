package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

func NewAPIServer(pfsAPIClient pfs.ApiClient, containerClient container.Client, storeClient store.Client, timer timing.Timer) pps.ApiServer {
	return newAPIServer(pfsAPIClient, containerClient, storeClient, timer)
}
