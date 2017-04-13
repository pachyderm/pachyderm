package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/worker"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps/server"
	"google.golang.org/grpc"
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

	// Either pipeline name or job name must be set
	PPSPipelineName string `env:"PPS_PIPELINE_NAME"`
	PPSJobID        string `env:"PPS_JOB_ID"`
	PodName         string `env:"PPS_POD_NAME,required"`
}

func main() {
	cmdutil.Main(do, &appEnv{})
}

func validateEnv(appEnv *appEnv) error {
	if appEnv.PPSPipelineName == "" && appEnv.PPSJobID == "" {
		return fmt.Errorf("worker must recieve either pipeline name or job ID, but got neither")
	} else if appEnv.PPSPipelineName != "" && appEnv.PPSJobID != "" {
		return fmt.Errorf("worker must recieve either pipeline name or job ID, but got both")
	}
	return nil
}

// getPipelineInfo gets the PipelineInfo proto describing the pipeline that this
// worker is part of
func getPipelineInfo(etcdClient *etcd.Client, appEnv *appEnv) (*pps.PipelineInfo, error) {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	resp, err := etcdClient.Get(ctx, path.Join(appEnv.PPSPrefix, "pipelines", appEnv.PPSPipelineName))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("expected to find 1 pipeline, got %d: %v", len(resp.Kvs), resp)
	}
	pipelineInfo := new(pps.PipelineInfo)
	if err := proto.UnmarshalText(string(resp.Kvs[0].Value), pipelineInfo); err != nil {
		return nil, err
	}
	return pipelineInfo, nil
}

func getJobInfo(etcdClient *etcd.Client, appEnv *appEnv) (*pps.JobInfo, error) {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	resp, err := etcdClient.Get(ctx, path.Join(appEnv.PPSPrefix, "jobs", appEnv.PPSJobID))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("expected to find 1 job, got %d: %v", len(resp.Kvs), resp)
	}
	jobInfo := new(pps.JobInfo)
	if err := proto.UnmarshalText(string(resp.Kvs[0].Value), jobInfo); err != nil {
		return nil, err
	}
	return jobInfo, nil
}

func do(appEnvObj interface{}) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	appEnv := appEnvObj.(*appEnv)
	if err := validateEnv(appEnv); err != nil {
		return err
	}

	// get pachd client, so we can upload output data from the user binary
	pachClient, err := client.NewFromAddress(fmt.Sprintf("%v:650", appEnv.PachdAddress))
	if err != nil {
		return err
	}
	go pachClient.KeepConnected(make(chan bool)) // we never cancel the connection

	// Get etcd client, so we can register our IP (so pachd can discover us)
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", appEnv.EtcdAddress)},
		DialTimeout: 15 * time.Second,
	})
	if err != nil {
		return err
	}

	// Construct worker API server. Get relevant pipeline or job info, and then
	// use that to create a worker.APIServer.
	var workerRcName string
	var apiServer *worker.APIServer
	if appEnv.PPSPipelineName != "" {
		pipelineInfo, err := getPipelineInfo(etcdClient, appEnv)
		if err != nil {
			return err
		}
		workerRcName = ppsserver.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
		apiServer = worker.NewPipelineAPIServer(pachClient, pipelineInfo, appEnv.PodName)
	} else if appEnv.PPSJobID != "" {
		jobInfo, err := getJobInfo(etcdClient, appEnv)
		if err != nil {
			return err
		}
		workerRcName = ppsserver.JobRcName(jobInfo.Job.ID)
		apiServer = worker.NewJobAPIServer(pachClient, jobInfo, appEnv.PodName)
	}

	// Setup the hostPath mount to use a unique directory for this worker
	workerDir := filepath.Join(client.PPSHostPath, appEnv.PodName)
	if err := os.MkdirAll(workerDir, 0777); err != nil {
		return err
	}
	if err := os.Symlink(workerDir, client.PPSInputPrefix); err != nil {
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
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	resp, err := etcdClient.Grant(ctx, 10 /* seconds */)
	if err != nil {
		return err
	}
	etcdClient.KeepAlive(context.Background(), resp.ID) // keepalive forever

	// Actually write "key" into etcd
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second) // new ctx
	if _, err := etcdClient.Put(ctx, key, "", etcd.WithLease(resp.ID)); err != nil {
		return err
	}

	// If server ever exits, return error
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}
