package server

import (
	"github.com/pachyderm/pachyderm/src/pkg/deploy"

	k8s "k8s.io/kubernetes/pkg/client/unversioned"
)

type APIServer interface {
	deploy.APIServer
}

func NewAPIServer(client *k8s.Client) APIServer {
	return newAPIServer(client)
}
