package server

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	etcd "github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	// WorkerEtcdPrefix is the prefix in etcd that we use to store worker information.
	WorkerEtcdPrefix = "workers"

	defaultTimeout = time.Second * 5
)

// Status returns the statuses of workers referenced by pipelineRcName.
// pipelineRcName is the name of the pipeline's RC and can be gotten with
// ppsutil.PipelineRcName. You can also pass "" for pipelineRcName to get all
// clients for all workers.
func Status(ctx context.Context, pipelineRcName string, etcdClient *etcd.Client, etcdPrefix string, workerGrpcPort uint16) ([]*pps.WorkerStatus, error) {
	workerClients, err := Clients(ctx, pipelineRcName, etcdClient, etcdPrefix, workerGrpcPort)
	if err != nil {
		return nil, err
	}
	var result []*pps.WorkerStatus
	for _, workerClient := range workerClients {
		ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
		status, err := workerClient.Status(ctx, &types.Empty{})
		if err != nil {
			log.Warnf("error getting worker status: %v", err)
			continue
		}
		result = append(result, status)
	}
	return result, nil
}

// Cancel cancels a set of datums running on workers.
// pipelineRcName is the name of the pipeline's RC and can be gotten with
// ppsutil.PipelineRcName.
func Cancel(ctx context.Context, pipelineRcName string, etcdClient *etcd.Client,
	etcdPrefix string, workerGrpcPort uint16, jobID string, dataFilter []string) error {
	workerClients, err := Clients(ctx, pipelineRcName, etcdClient, etcdPrefix, workerGrpcPort)
	if err != nil {
		return err
	}
	success := false
	for _, workerClient := range workerClients {
		resp, err := workerClient.Cancel(ctx, &CancelRequest{
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
		return errors.Errorf("datum matching filter %+v could not be found for job ID %s", dataFilter, jobID)
	}
	return nil
}

// Conns returns a slice of connections to worker servers.
// pipelineRcName is the name of the pipeline's RC and can be gotten with
// ppsutil.PipelineRcName. You can also pass "" for pipelineRcName to get all
// clients for all workers.
func Conns(ctx context.Context, pipelineRcName string, etcdClient *etcd.Client, etcdPrefix string, workerGrpcPort uint16, workerIP ...string) ([]*grpc.ClientConn, error) {
	resp, err := etcdClient.Get(ctx, path.Join(etcdPrefix, WorkerEtcdPrefix, pipelineRcName), etcd.WithPrefix())
	if err != nil {
		return nil, err
	}
	var result []*grpc.ClientConn
	for _, kv := range resp.Kvs {
		wIP := path.Base(string(kv.Key))
		if len(workerIP) > 0 && wIP != workerIP[0] {
			continue
		}
		ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
		conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", wIP, workerGrpcPort),
			append(client.DefaultDialOptions(), grpc.WithInsecure())...)
		if err != nil {
			return nil, err
		}
		result = append(result, conn)
	}
	return result, nil
}

// Client combines the WorkerAPI and the DebugAPI into a single client.
type Client struct {
	WorkerClient
	debug.DebugClient
}

func newClient(conn *grpc.ClientConn) Client {
	return Client{
		NewWorkerClient(conn),
		debug.NewDebugClient(conn),
	}
}

// Clients returns a slice of worker clients for a pipeline.
// pipelineRcName is the name of the pipeline's RC and can be gotten with
// ppsutil.PipelineRcName. You can also pass "" for pipelineRcName to get all
// clients for all workers.
func Clients(ctx context.Context, pipelineRcName string, etcdClient *etcd.Client, etcdPrefix string, workerGrpcPort uint16, workerIP ...string) ([]Client, error) {
	conns, err := Conns(ctx, pipelineRcName, etcdClient, etcdPrefix, workerGrpcPort, workerIP...)
	if err != nil {
		return nil, err
	}
	var result []Client
	for _, conn := range conns {
		result = append(result, newClient(conn))
	}
	return result, nil
}

// NewClient returns a worker client for the worker at the IP address passed in.
func NewClient(address string) (Client, error) {
	port, err := strconv.Atoi(os.Getenv(client.PPSWorkerPortEnv))
	if err != nil {
		return Client{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", address, port),
		append(client.DefaultDialOptions(), grpc.WithInsecure())...)
	if err != nil {
		return Client{}, err
	}
	return newClient(conn), nil
}
