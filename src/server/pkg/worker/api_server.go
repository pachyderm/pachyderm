package worker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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
	for i, datum := range data {
		input := a.pipelineInfo.Inputs[i]
		if err := filesync.Pull(ctx, a.pachClient, filepath.Join(client.PPSInputPrefix, input.Name), datum, input.Lazy); err != nil {
			return err
		}
	}
	return nil
}

func (a *APIServer) runUserCode(ctx context.Context) error {
	if err := os.MkdirAll(client.PPSOutputPath, 0666); err != nil {
		return err
	}
	transform := a.pipelineInfo.Transform
	cmd := exec.Command(transform.Cmd[0], transform.Cmd[1:]...)
	cmd.Stdin = strings.NewReader(strings.Join(transform.Stdin, "\n") + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	success := true
	if err := cmd.Run(); err != nil {
		success = false
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				for _, returnCode := range transform.AcceptReturnCode {
					if int(returnCode) == status.ExitStatus() {
						success = true
					}
				}
			}
		}
		if !success {
			fmt.Fprintf(os.Stderr, "Error from exec: %s\n", err.Error())
		}
	}
	return nil
}

func (a *APIServer) uploadOutput(ctx context.Context) error {
	return nil
}

func (a *APIServer) Process(ctx context.Context, req *ProcessRequest) (resp *ProcessResponse, retErr error) {
	defer func(start time.Time) { a.Log(req, resp, retErr, time.Since(start)) }(time.Now())
	a.Lock()
	defer a.Unlock()
	if err := a.downloadData(ctx, req.Data); err != nil {
		return nil, err
	}
	if err := a.runUserCode(ctx); err != nil {
		return nil, err
	}
	if err := a.uploadOutput(ctx); err != nil {
		return nil, err
	}
	return &ProcessResponse{}, nil
}
