package server

import (
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

func NewAPIServer(storeClient store.Client) pps.ApiServer {
	return newAPIServer(storeClient)
}
