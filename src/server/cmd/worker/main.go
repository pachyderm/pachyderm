package main

import (
	"context"
	"os"
	"path"
	"time"

	debugclient "github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	logutil "github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/profileutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	debugserver "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	"github.com/pachyderm/pachyderm/v2/src/server/worker"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"

	log "github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
)

func main() {
	log.SetFormatter(logutil.FormatterFunc(logutil.Pretty))

	// append pachyderm bins to path to allow use of pachctl
	os.Setenv("PATH", os.Getenv("PATH")+":/pach-bin")

	cmdutil.Main(do, &serviceenv.WorkerFullConfiguration{})
}

func do(config interface{}) error {
	// must run InstallJaegerTracer before InitWithKube/pach client initialization
	tracing.InstallJaegerTracerFromEnv()
	env := serviceenv.InitServiceEnv(serviceenv.NewConfiguration(config))

	// Enable cloud profilers if the configuration allows.
	profileutil.StartCloudProfiler("pachyderm-worker", env.Config())

	// Construct a client that connects to the sidecar.
	pachClient := env.GetPachClient(context.Background())
	pipelineInfo, err := ppsutil.GetWorkerPipelineInfo(
		pachClient,
		env.GetDBClient(),
		env.GetPostgresListener(),
		env.Config().PPSPipelineName,
		env.Config().PPSSpecCommitID,
	) // get pipeline creds for pachClient
	if err != nil {
		return errors.Wrapf(err, "error getting pipelineInfo")
	}

	// Construct worker API server.
	workerRcName := ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	workerInstance, err := worker.NewWorker(env, pachClient, pipelineInfo, "/")
	if err != nil {
		return err
	}

	// Start worker api server
	server, err := grpcutil.NewServer(context.Background(), false)
	if err != nil {
		return err
	}

	workerserver.RegisterWorkerServer(server.Server, workerInstance.APIServer)
	versionpb.RegisterAPIServer(server.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
	debugclient.RegisterDebugServer(server.Server, debugserver.NewDebugServer(env, env.Config().PodName, pachClient))

	// Put our IP address into etcd, so pachd can discover us
	key := path.Join(env.Config().PPSEtcdPrefix, workerserver.WorkerEtcdPrefix, workerRcName, env.Config().PPSWorkerIP)

	// Prepare to write "key" into etcd by creating lease -- if worker dies, our
	// IP will be removed from etcd
	ctx, cancel := context.WithTimeout(pachClient.Ctx(), 10*time.Second)
	defer cancel()

	resp, err := env.GetEtcdClient().Grant(ctx, 10 /* seconds */)
	if err != nil {
		return errors.Wrapf(err, "error granting lease")
	}

	// keepalive forever
	if _, err := env.GetEtcdClient().KeepAlive(context.Background(), resp.ID); err != nil {
		return errors.Wrapf(err, "error with KeepAlive")
	}

	// Actually write "key" into etcd
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second) // new ctx
	defer cancel()
	if _, err := env.GetEtcdClient().Put(ctx, key, "", etcd.WithLease(resp.ID)); err != nil {
		return errors.Wrapf(err, "error putting IP address")
	}

	// If server ever exits, return error
	if _, err := server.ListenTCP("", env.Config().PPSWorkerPort); err != nil {
		return err
	}
	return server.Wait()
}
