package server

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy"
	"github.com/pachyderm/pachyderm/src/server/pkg/provider"

	k8s "k8s.io/kubernetes/pkg/client/unversioned"
)

type APIServer interface {
	deploy.APIServer
}

func NewAPIServer(client *k8s.Client, provider provider.Provider) APIServer {
	return newAPIServer(client, provider)
}
