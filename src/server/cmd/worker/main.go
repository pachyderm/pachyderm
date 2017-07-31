package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"path"
	"time"

	"golang.org/x/sync/errgroup"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"github.com/pachyderm/pachyderm/src/server/worker"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

// appEnv stores the environment variables that this worker needs
type appEnv struct {
	// Address of etcd, so that worker can write its own IP there for discoverh
	EtcdAddress string `env:"ETCD_PORT_2379_TCP_ADDR,required"`

	// Address for connecting to pachd (so this can download input data)
	PachdAddress string `env:"PACHD_PORT_650_TCP_ADDR,required"`

	// Prefix in etcd for all pachd-related records
	PPSPrefix string `env:"PPS_ETCD_PREFIX,required"`

	// worker gets its own IP here, via the k8s downward API. It then writes that
	// IP back to etcd so that pachd can discover it
	PPSWorkerIP string `env:"PPS_WORKER_IP,required"`

	// The name of the pipeline that this worker belongs to
	PPSPipelineName string `env:"PPS_PIPELINE_NAME,required"`

	// The name of this pod
	PodName string `env:"PPS_POD_NAME,required"`

	// The namespace in which Pachyderm is deployed
	Namespace string `env:"PPS_NAMESPACE,required"`
}

func main() {
	cmdutil.Main(do, &appEnv{})
}

// getPipelineInfo gets the PipelineInfo proto describing the pipeline that this
// worker is part of
func getPipelineInfo(etcdClient *etcd.Client, appEnv *appEnv) (*pps.PipelineInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := etcdClient.Get(ctx, path.Join(appEnv.PPSPrefix, "pipelines", appEnv.PPSPipelineName))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("expected to find 1 pipeline, got %d: %v", len(resp.Kvs), resp)
	}
	pipelineInfo := new(pps.PipelineInfo)

	if err := pipelineInfo.Unmarshal(resp.Kvs[0].Value); err != nil {
		return nil, err
	}
	return pipelineInfo, nil
}

func do(appEnvObj interface{}) error {
	go func() {
		log.Println(http.ListenAndServe(":652", nil))
	}()

	appEnv := appEnvObj.(*appEnv)

	// Construct a client that connects to the sidecar.
	pachClient, err := client.NewFromAddress("localhost:650")
	if err != nil {
		return fmt.Errorf("error constructing pachClient: %v", err)
	}

	// Get etcd client, so we can register our IP (so pachd can discover us)
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", appEnv.EtcdAddress)},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return fmt.Errorf("error constructing etcdClient: %v", err)
	}

	pipelineInfo, err := getPipelineInfo(etcdClient, appEnv)
	if err != nil {
		return fmt.Errorf("error getting pipelineInfo: %v", err)
	}

	// Set the auth token that will be used for this client
	pachClient.SetAuthToken(pipelineInfo.Capability)

	// Construct worker API server.
	workerRcName := ppsserver.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	apiServer, err := worker.NewAPIServer(pachClient, etcdClient, appEnv.PPSPrefix, pipelineInfo, appEnv.PodName, appEnv.Namespace)
	if err != nil {
		return err
	}

	// Start worker api server
	eg := errgroup.Group{}
	ready := make(chan error)
	eg.Go(func() error {
		return grpcutil.Serve(
			func(s *grpc.Server) {
				worker.RegisterWorkerServer(s, apiServer)
				close(ready)
			},
			grpcutil.ServeOptions{
				Version:    version.Version,
				MaxMsgSize: grpcutil.MaxMsgSize,
			},
			grpcutil.ServeEnv{
				GRPCPort: client.PPSWorkerPort,
			},
		)
	})

	// Wait until server is ready, then put our IP address into etcd, so pachd can
	// discover us
	<-ready
	key := path.Join(appEnv.PPSPrefix, "workers", workerRcName, appEnv.PPSWorkerIP)

	// Prepare to write "key" into etcd by creating lease -- if worker dies, our
	// IP will be removed from etcd
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := etcdClient.Grant(ctx, 10 /* seconds */)
	if err != nil {
		return fmt.Errorf("error granting lease: %v", err)
	}

	// keepalive forever
	if _, err := etcdClient.KeepAlive(context.Background(), resp.ID); err != nil {
		return fmt.Errorf("error with KeepAlive: %v", err)
	}

	// Actually write "key" into etcd
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second) // new ctx
	defer cancel()
	if _, err := etcdClient.Put(ctx, key, "", etcd.WithLease(resp.ID)); err != nil {
		return fmt.Errorf("error putting IP address: %v", err)
	}

	// If server ever exits, return error
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}
