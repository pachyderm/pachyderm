package worker

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	filesync "github.com/pachyderm/pachyderm/src/server/pkg/sync"
)

type APIServer struct {
	sync.Mutex
	protorpclog.Logger
	pachClient   *client.APIClient
	etcdClient   *etcd.Client
	pipelineInfo *pps.PipelineInfo
}

func NewAPIServer(pachClient *client.APIClient, etcdClient *etcd.Client, pipelineInfo *pps.PipelineInfo) *APIServer {
	return &APIServer{
		Mutex:        sync.Mutex{},
		Logger:       protorpclog.NewLogger(""),
		pachClient:   pachClient,
		etcdClient:   etcdClient,
		pipelineInfo: pipelineInfo,
	}
}

func (a *APIServer) downloadData(ctx context.Context, data []*pfs.FileInfo) error {
	fmt.Println("BP0")
	for i, datum := range data {
		input := a.pipelineInfo.Inputs[i]
		fmt.Println("BP1")
		if err := filesync.Pull(ctx, a.pachClient, filepath.Join(client.PPSInputPrefix, input.Name), datum, input.Lazy); err != nil {
			return err
		}
		fmt.Println("BP2")
	}
	return nil
}

func (a *APIServer) Process(ctx context.Context, req *ProcessRequest) (resp *ProcessResponse, retErr error) {
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	a.Lock()
	defer a.Unlock()
	if err := a.downloadData(ctx, req.Data); err != nil {
		return nil, err
	}
	return &ProcessResponse{}, nil
}
