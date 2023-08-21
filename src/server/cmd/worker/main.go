package main

import (
	"context"
	"os"
	"path"
	"time"

	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	debugclient "github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/proc"
	"github.com/pachyderm/pachyderm/v2/src/internal/profileutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	debugserver "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	"github.com/pachyderm/pachyderm/v2/src/server/worker"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	workerapi "github.com/pachyderm/pachyderm/v2/src/worker"
)

func main() {
	log.InitWorkerLogger()
	ctx := pctx.Child(pctx.Background(""), "", pctx.WithFields(pps.WorkerIDField(os.Getenv(client.PPSPodNameEnv))))
	go log.WatchDroppedLogs(ctx, time.Minute)
	go proc.MonitorSelf(ctx)
	log.Debug(ctx, "version info", log.Proto("versionInfo", version.Version))

	// append pachyderm bins to path to allow use of pachctl
	os.Setenv("PATH", os.Getenv("PATH")+":/pach-bin")
	cmdutil.Main(ctx, do, &pachconfig.WorkerFullConfiguration{})
}

func do(ctx context.Context, config *pachconfig.WorkerFullConfiguration) error {
	// must run InstallJaegerTracer before InitWithKube/pach client initialization
	tracing.InstallJaegerTracerFromEnv()
	env := serviceenv.InitWithKube(ctx, pachconfig.NewConfiguration(config))

	// Enable cloud profilers if the configuration allows.
	profileutil.StartCloudProfiler(ctx, "pachyderm-worker", env.Config())

	// Construct a client that connects to the sidecar.
	pachClient := env.GetPachClient(ctx)
	p := &pps.Pipeline{
		Project: &pfs.Project{Name: env.Config().PPSProjectName},
		Name:    env.Config().PPSPipelineName,
	}
	pipelineInfo, err := ppsutil.GetWorkerPipelineInfo(
		pachClient,
		env.GetDBClient(),
		env.GetPostgresListener(),
		p,
		env.Config().PPSSpecCommitID,
	) // get pipeline creds for pachClient
	if err != nil {
		return errors.Wrapf(err, "worker: get pipelineInfo for %q", p)
	}
	ctx = pachClient.AddMetadata(ctx)

	// Construct worker API server.
	workerInstance, err := worker.NewWorker(pctx.Child(ctx, ""), env, pachClient, pipelineInfo, "/")
	if err != nil {
		return err
	}

	// grpc logger
	interceptor := logging.NewLoggingInterceptor(ctx)
	interceptor.Level = log.DebugLevel

	// Start worker api server
	server, err := grpcutil.NewServer(ctx, false,
		grpc.ChainUnaryInterceptor(interceptor.UnaryServerInterceptor),
		grpc.ChainStreamInterceptor(interceptor.StreamServerInterceptor),
	)
	if err != nil {
		return err
	}

	workerapi.RegisterWorkerServer(server.Server, workerInstance.APIServer)
	versionpb.RegisterAPIServer(server.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
	debugclient.RegisterDebugServer(server.Server, debugserver.NewDebugServer(env, env.Config().PodName, pachClient, env.GetDBClient()))

	// Put our IP address into etcd, so pachd can discover us
	workerRcName := ppsutil.PipelineRcName(pipelineInfo)
	key := path.Join(env.Config().PPSEtcdPrefix, workerserver.WorkerEtcdPrefix, workerRcName, env.Config().PPSWorkerIP)

	// Prepare to write "key" into etcd by creating lease -- if worker dies, our
	// IP will be removed from etcd
	leaseID, err := getETCDLease(ctx, env.GetEtcdClient(), 10*time.Second)
	if err != nil {
		return errors.Wrapf(err, "worker: get etcd lease")
	}

	// keepalive forever
	keepAliveChan, err := env.GetEtcdClient().KeepAlive(ctx, leaseID)
	if err != nil {
		return errors.Wrapf(err, "worker: etcd KeepAlive")
	}
	go func() {
		for {
			_, more := <-keepAliveChan
			if !more {
				log.Error(ctx, "failed to renew worker IP address etcd lease")
				return
			}
		}
	}()

	if err := writeKey(ctx, env.GetEtcdClient(), key, leaseID, 10*time.Second); err != nil {
		return errors.Wrapf(err, "worker: etcd key %s", key)
	}

	// If server ever exits, return error
	if _, err := server.ListenTCP("", env.Config().PPSWorkerPort); err != nil {
		return err
	}
	return server.Wait()
}

func getETCDLease(ctx context.Context, client *etcd.Client, duration time.Duration) (etcd.LeaseID, error) {
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	sec := int64(duration / time.Second)
	if sec == 0 { // do not aallow durations < 1 second to round down to 0 seconds
		sec = 1
	}
	resp, err := client.Grant(ctx, sec)
	if err != nil {
		return 0, errors.Wrapf(err, "getETCDLease: etcd grant")
	}
	return resp.ID, nil
}

func writeKey(ctx context.Context, client *etcd.Client, key string, id etcd.LeaseID, duration time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := client.Put(ctx, key, "", etcd.WithLease(id)); err != nil {
		return errors.Wrapf(err, "writeKey: etcd put")
	}
	return nil
}
