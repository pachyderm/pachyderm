package worker

import (
	"path"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/sync"
)

type APIServer struct {
	pachClient   *client.APIClient
	etcdClient   *etcd.Client
	pipelineInfo *pps.PipelineInfo
}

func NewAPIServer(pachClient *client.APIClient, etcdClient *etcd.Client, pipelineInfo *pps.PipelineInfo) *APIServer {
	return &APIServer{
		pachClient:   pachClient,
		etcdClient:   etcdClient,
		pipelineInfo: pipelineInfo,
	}
}

func (a *APIServer) downloadData(ctx context.Context, data []*pfs.FileInfo) error {
	for i, datum := range data {
		input := a.pipelineInfo.Inputs[i]
		if err := sync.Pull(ctx, a.pachClient, path.Join(client.PPSInputPrefix, input.Name), datum, input.Lazy); err != nil {
			return err
		}
	}
	return nil
}

func (a *APIServer) Process(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error) {
	if err := a.downloadData(ctx, req.Data); err != nil {
		return nil, err
	}
	return nil, nil
}
