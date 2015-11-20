package jobserver

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

type CombinedJobAPIServer interface {
	pps.JobAPIServer
	pps.InternalJobAPIServer
}

func NewAPIServer(
	pfsAPIClient pfs.APIClient,
	persistAPIClient persist.APIClient,
	client *kube.Client,
) CombinedJobAPIServer {
	return newAPIServer(
		pfsAPIClient,
		persistAPIClient,
		client,
	)
}
