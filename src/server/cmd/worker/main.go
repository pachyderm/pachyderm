package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	debugclient "github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/client/version/versionpb"
	debugserver "github.com/pachyderm/pachyderm/src/server/debug/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/worker"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

// appEnv stores the environment variables that this worker needs
type appEnv struct {
	// Address of etcd, so that worker can write its own IP there for discoverh
	EtcdAddress string `env:"ETCD_PORT_2379_TCP_ADDR,required"`

	// Prefix in etcd for all pachd-related records
	PPSPrefix string `env:"PPS_ETCD_PREFIX,required"`

	// worker gets its own IP here, via the k8s downward API. It then writes that
	// IP back to etcd so that pachd can discover it
	PPSWorkerIP string `env:"PPS_WORKER_IP,required"`

	// The name of the pipeline that this worker belongs to
	PPSPipelineName string `env:"PPS_PIPELINE_NAME,required"`

	// The ID of the commit that contains the pipeline spec.
	PPSSpecCommitID string `env:"PPS_SPEC_COMMIT,required"`

	// The name of this pod
	PodName string `env:"PPS_POD_NAME,required"`

	// The namespace in which Pachyderm is deployed
	Namespace string `env:"PPS_NAMESPACE,required"`

	// StorageRoot is where we store hashtrees
	StorageRoot string `env:"PACH_ROOT,default=/pach"`
}

func main() {
	// Copy the contents of /pach-bin/certs into /etc/ssl/certs. Don't return an
	// error (which would cause 'Walk()' to exit early) but do record if any certs
	// are known to be missing so we can inform the user
	copyErr := false
	if err := filepath.Walk("/pach-bin/certs", func(inPath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Warnf("skipping \"%s\", could not stat path: %v", inPath, err)
			copyErr = true
			return nil // Don't try and fix any errors encountered by Walk() itself
		}
		if info.IsDir() {
			return nil // We'll just copy the children of any directories when we traverse them
		}

		// Open input file (src)
		in, err := os.OpenFile(inPath, os.O_RDONLY, 0)
		if err != nil {
			log.Warnf("could not read \"%s\": %v", inPath, err)
			copyErr = true
			return nil
		}
		defer in.Close()

		// Create output file (dest) and open for writing
		outRelPath, err := filepath.Rel("/pach-bin/certs", inPath)
		if err != nil {
			log.Warnf("skipping \"%s\", could not extract relative path: %v", inPath, err)
			copyErr = true
			return nil
		}
		outPath := filepath.Join("/etc/ssl/certs", outRelPath)
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			log.Warnf("skipping \"%s\", could not create directory \"%s\": %v", inPath, outDir, err)
			copyErr = true
			return nil
		}
		out, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			log.Warnf("skipping \"%s\", could not create output file \"%s\": %v", inPath, outPath, err)
			copyErr = true
			return nil
		}
		defer out.Close()

		// Copy src -> dest
		if _, err := io.Copy(out, in); err != nil {
			log.Warnf("could not copy \"%s\" to \"%s\": %v", inPath, outPath, err)
			copyErr = true
			return nil
		}
		return nil
	}); err != nil {
		// Should never happen
		log.Warnf("could not copy /pach-bin/certs to /etc/ssl/certs: %v", err)
	}
	if copyErr {
		log.Warnf("Errors were encountered while copying /pach-bin/certs to /etc/ssl/certs (see above--might result in subsequent SSL/TLS errors)")
	}
	cmdutil.Main(do, &appEnv{})
}

// getPipelineInfo gets the PipelineInfo proto describing the pipeline that this
// worker is part of.
// getPipelineInfo has the side effect of adding auth to the passed pachClient
// which is necessary to get the PipelineInfo from pfs.
func getPipelineInfo(etcdClient *etcd.Client, pachClient *client.APIClient, appEnv *appEnv) (*pps.PipelineInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := etcdClient.Get(ctx, path.Join(appEnv.PPSPrefix, "pipelines", appEnv.PPSPipelineName))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("expected to find 1 pipeline (%s), got %d: %v", appEnv.PPSPipelineName, len(resp.Kvs), resp)
	}
	var pipelinePtr pps.EtcdPipelineInfo
	if err := pipelinePtr.Unmarshal(resp.Kvs[0].Value); err != nil {
		return nil, err
	}
	pachClient.SetAuthToken(pipelinePtr.AuthToken)
	// Notice we use the SpecCommitID from our env, not from etcd. This is
	// because the value in etcd might get updated while the worker pod is
	// being created and we don't want to run the transform of one version of
	// the pipeline in the image of a different verison.
	pipelinePtr.SpecCommit.ID = appEnv.PPSSpecCommitID
	return ppsutil.GetPipelineInfo(pachClient, &pipelinePtr)
}

func do(appEnvObj interface{}) error {
	go func() {
		log.Println(http.ListenAndServe(":651", nil))
	}()

	appEnv := appEnvObj.(*appEnv)

	// Construct a client that connects to the sidecar.
	pachClient, err := client.NewFromAddress("localhost:653")
	if err != nil {
		return fmt.Errorf("error constructing pachClient: %v", err)
	}

	// Get etcd client, so we can register our IP (so pachd can discover us)
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", appEnv.EtcdAddress)},
		DialOptions: client.DefaultDialOptions(),
	})
	if err != nil {
		return fmt.Errorf("error constructing etcdClient: %v", err)
	}

	pipelineInfo, err := getPipelineInfo(etcdClient, pachClient, appEnv)
	if err != nil {
		return fmt.Errorf("error getting pipelineInfo: %v", err)
	}

	// Construct worker API server.
	workerRcName := ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	apiServer, err := worker.NewAPIServer(pachClient, etcdClient, appEnv.PPSPrefix, pipelineInfo, appEnv.PodName, appEnv.Namespace, appEnv.StorageRoot)
	if err != nil {
		return err
	}

	// Start worker api server
	eg := errgroup.Group{}
	ready := make(chan error)
	eg.Go(func() error {
		return grpcutil.Serve(
			grpcutil.ServerOptions{
				MaxMsgSize: grpcutil.MaxMsgSize,
				Port:       client.PPSWorkerPort,
				RegisterFunc: func(s *grpc.Server) error {
					defer close(ready)
					worker.RegisterWorkerServer(s, apiServer)
					versionpb.RegisterAPIServer(s, version.NewAPIServer(version.Version, version.APIServerOptions{}))
					debugclient.RegisterDebugServer(s, debugserver.NewDebugServer(appEnv.PodName, etcdClient, appEnv.PPSPrefix))
					return nil
				},
			},
		)
	})

	// Wait until server is ready, then put our IP address into etcd, so pachd can
	// discover us
	<-ready
	key := path.Join(appEnv.PPSPrefix, worker.WorkerEtcdPrefix, workerRcName, appEnv.PPSWorkerIP)

	// Prepare to write "key" into etcd by creating lease -- if worker dies, our
	// IP will be removed from etcd
	ctx, cancel := context.WithTimeout(pachClient.Ctx(), 10*time.Second)
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
	return eg.Wait()
}
