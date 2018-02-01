package server

import (
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
)

type apiServer struct {
	log.Logger
	pachClient *client.APIClient
}

func (a *apiServer) Extract(request *admin.ExtractRequest, extractServer admin.API_ExtractServer) error {
	return nil
}

func (a *apiServer) Restore(restoreServer admin.API_RestoreServer) error {
	return nil
}
