package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	debugclient "github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	logutil "github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/profileutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	debugserver "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	"github.com/pachyderm/pachyderm/v2/src/server/worker"
	workerserver "github.com/pachyderm/pachyderm/v2/src/server/worker/server"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"

	etcd "github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(logutil.FormatterFunc(logutil.Pretty))

	// append pachyderm bins to path to allow use of pachctl
	os.Setenv("PATH", os.Getenv("PATH")+":/pach-bin")

	cmdutil.Main(do, &serviceenv.WorkerFullConfiguration{})
}

// getPipelineInfo gets the PipelineInfo proto describing the pipeline that this
// worker is part of.
// getPipelineInfo has the side effect of adding auth to the passed pachClient
// which is necessary to get the PipelineInfo from pfs.
func getPipelineInfo(pachClient *client.APIClient, env serviceenv.ServiceEnv) (*pps.PipelineInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pipelines := ppsdb.Pipelines(env.GetDBClient(), env.GetPostgresListener())
	pipelineInfo := &pps.PipelineInfo{}
	// Notice we use the SpecCommitID from our env, not from postgres. This is
	// because the value in postgres might get updated while the worker pod is
	// being created and we don't want to run the transform of one version of
	// the pipeline in the image of a different verison.
	fmt.Sprintf("%s@%s", env.Config().PPSPipelineName, env.Config().PPSSpecCommitID)
	if err := pipelines.ReadOnly(ctx).Get(env.Config().PPSPipelineName, pipelineInfo); err != nil {
		return nil, err
	}
	pachClient.SetAuthToken(pipelineInfo.AuthToken)

	return pipelineInfo, nil
}

func do(config interface{}) error {
	// must run InstallJaegerTracer before InitWithKube/pach client initialization
	tracing.InstallJaegerTracerFromEnv()
	env := serviceenv.InitServiceEnv(serviceenv.NewConfiguration(config))

	// Enable cloud profilers if the configuration allows.
	profileutil.StartCloudProfiler("pachyderm-worker", env.Config())

	// Construct a client that connects to the sidecar.
	pachClient := env.GetPachClient(context.Background())
	pipelineInfo, err := getPipelineInfo(pachClient, env) // get pipeline creds for pachClient
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
