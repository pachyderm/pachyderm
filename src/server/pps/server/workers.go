package server

import (
	"context"
	"fmt"
	"path"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	workerpkg "github.com/pachyderm/pachyderm/src/server/worker"

	etcd "github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
)

const (
	workerEtcdPrefix = "workers"
)

func status(ctx context.Context, id string, etcdClient *etcd.Client, etcdPrefix string) ([]*pps.WorkerStatus, error) {
	workerClients, err := workerClients(ctx, id, etcdClient, etcdPrefix)
	if err != nil {
		return nil, err
	}
	var result []*pps.WorkerStatus
	for _, workerClient := range workerClients {
		status, err := workerClient.Status(ctx, &types.Empty{})
		if err != nil {
			return nil, err
		}
		result = append(result, status)
	}
	return result, nil
}

func cancel(ctx context.Context, id string, etcdClient *etcd.Client,
	etcdPrefix string, jobID string, dataFilter []string) error {
	workerClients, err := workerClients(ctx, id, etcdClient, etcdPrefix)
	if err != nil {
		return err
	}
	success := false
	for _, workerClient := range workerClients {
		resp, err := workerClient.Cancel(ctx, &workerpkg.CancelRequest{
			JobID:       jobID,
			DataFilters: dataFilter,
		})
		if err != nil {
			return err
		}
		if resp.Success {
			success = true
		}
	}
	if !success {
		return fmt.Errorf("datum matching filter %+v could not be found for jobID %s", dataFilter, jobID)
	}
	return nil
}

func workerClients(ctx context.Context, id string, etcdClient *etcd.Client, etcdPrefix string) ([]workerpkg.WorkerClient, error) {
	resp, err := etcdClient.Get(ctx, path.Join(etcdPrefix, workerEtcdPrefix, id), etcd.WithPrefix())
	if err != nil {
		return nil, err
	}

	var result []workerpkg.WorkerClient
	for _, kv := range resp.Kvs {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", path.Base(string(kv.Key)), client.PPSWorkerPort),
			client.PachDialOptions()...)
		if err != nil {
			return nil, err
		}
		result = append(result, workerpkg.NewWorkerClient(conn))
	}
	return result, nil
}
