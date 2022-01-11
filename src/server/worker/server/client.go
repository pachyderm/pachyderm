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

	log "github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
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
	var result []*pps.WorkerStatus
	if err := ForEachWorker(ctx, pipelineRcName, etcdClient, etcdPrefix, workerGrpcPort, func(c Client) error {
		ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
		status, err := c.Status(ctx, &types.Empty{})
		if err != nil {
			log.Warnf("error getting worker status: %v", err)
			return nil
		}
		result = append(result, status)
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// Cancel cancels a set of datums running on workers.
// pipelineRcName is the name of the pipeline's RC and can be gotten with
// ppsutil.PipelineRcName.
func Cancel(ctx context.Context, pipelineRcName string, etcdClient *etcd.Client,
	etcdPrefix string, workerGrpcPort uint16, jobID string, dataFilter []string) (retErr error) {
	success := false
	if err := ForEachWorker(ctx, pipelineRcName, etcdClient, etcdPrefix, workerGrpcPort, func(c Client) error {
		resp, err := c.Cancel(ctx, &CancelRequest{
			JobID:       jobID,
			DataFilters: dataFilter,
		})
		if err != nil {
			return err
		}
		if resp.Success {
			success = true
		}
		return nil
	}); err != nil {
		return err
	}
	if !success {
		return errors.Errorf("datum matching filter %+v could not be found for job ID %s", dataFilter, jobID)
	}
	return nil
}

// Client combines the WorkerAPI and the DebugAPI into a single client.
type Client struct {
	WorkerClient
	debug.DebugClient
	clientConn *grpc.ClientConn
}

func newClient(conn *grpc.ClientConn) Client {
	return Client{
		NewWorkerClient(conn),
		debug.NewDebugClient(conn),
		conn,
	}
}

func (c *Client) Close() error {
	return errors.EnsureStack(c.clientConn.Close())
}

// NewClient returns a worker client for the worker at the IP address passed in.
func NewClient(address string) (Client, error) {
	port, err := strconv.Atoi(os.Getenv(client.PPSWorkerPortEnv))
	if err != nil {
		return Client{}, errors.EnsureStack(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", address, port),
		append(client.DefaultDialOptions(), grpc.WithInsecure())...)
	if err != nil {
		return Client{}, errors.EnsureStack(err)
	}
	return newClient(conn), nil
}

// TODO: It may make sense to parallelize this, but we have to consider that this operation can have multiple instances running in parallel.
func ForEachWorker(ctx context.Context, pipelineRcName string, etcdClient *etcd.Client, etcdPrefix string, workerGrpcPort uint16, cb func(Client) error) error {
	resp, err := etcdClient.Get(ctx, path.Join(etcdPrefix, WorkerEtcdPrefix, pipelineRcName), etcd.WithPrefix())
	if err != nil {
		return errors.EnsureStack(err)
	}
	for _, kv := range resp.Kvs {
		workerIP := path.Base(string(kv.Key))
		if err := WithClient(ctx, workerIP, workerGrpcPort, cb); err != nil {
			return err
		}
	}
	return nil
}

func WithClient(ctx context.Context, address string, port uint16, cb func(Client) error) (retErr error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", address, port),
		append(client.DefaultDialOptions(), grpc.WithInsecure())...)
	if err != nil {
		return errors.EnsureStack(err)
	}
	c := newClient(conn)
	defer func() {
		if err := c.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(c)
}
