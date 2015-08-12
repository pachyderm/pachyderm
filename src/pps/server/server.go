package server

import (
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

func NewAPIServer(storeClient store.Client, timer timing.Timer) pps.ApiServer {
	return newAPIServer(storeClient, timer)
}
