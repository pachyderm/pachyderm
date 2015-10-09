package server //import "go.pachyderm.com/pachyderm/src/pps/server"

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/store"
	"go.pedge.io/pkg/time"
)

func NewAPIServer(pfsAPIClient pfs.ApiClient, containerClient container.Client, storeClient store.Client, timer pkgtime.Timer) pps.ApiServer {
	return newAPIServer(pfsAPIClient, containerClient, storeClient, timer)
}
