package server //import "go.pachyderm.com/pachyderm/src/pps/server"

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pkg/timing"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/container"
	"go.pachyderm.com/pachyderm/src/pps/store"
)

func NewAPIServer(pfsAPIClient pfs.ApiClient, containerClient container.Client, storeClient store.Client, timer timing.Timer) pps.ApiServer {
	return newAPIServer(pfsAPIClient, containerClient, storeClient, timer)
}
