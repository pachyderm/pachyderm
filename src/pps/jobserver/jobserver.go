package jobserver

import (
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

type CombinedJobAPIServer interface {
	pps.JobAPIServer
	pps.InternalJobAPIServer
}

func NewAPIServer(
	pfsAddress string,
	persistAPIServer persist.APIServer,
	client *kube.Client,
) CombinedJobAPIServer {
	return newAPIServer(
		pfsAddress,
		persistAPIServer,
		client,
	)
}
