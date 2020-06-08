package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	debugclient "github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/client/version/versionpb"
	debugserver "github.com/pachyderm/pachyderm/src/server/debug/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	logutil "github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/worker"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(logutil.FormatterFunc(logutil.Pretty))

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
		// Should never happen, but just log if it does
		copyErr = true
		log.Warnf("walk failed with: %v", err)
	}
	if copyErr {
		log.Warnf(
			"pachyderm's worker binary encountered errors while copying " +
				"/pach-bin/certs to /etc/ssl/certs (see above). This might cause the " +
				"worker binary to error while communicating with object storage for " +
				"egress pipelines or for merging pipeline outputs")
	}
	cmdutil.Main(do, &serviceenv.WorkerFullConfiguration{})
}

// getPipelineInfo gets the PipelineInfo proto describing the pipeline that this
// worker is part of.
// getPipelineInfo has the side effect of adding auth to the passed pachClient
// which is necessary to get the PipelineInfo from pfs.
func getPipelineInfo(pachClient *client.APIClient, env *serviceenv.ServiceEnv) (*pps.PipelineInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := env.GetEtcdClient().Get(ctx, path.Join(env.PPSEtcdPrefix, "pipelines", env.PPSPipelineName))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("expected to find 1 pipeline (%s), got %d: %v", env.PPSPipelineName, len(resp.Kvs), resp)
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
	pipelinePtr.SpecCommit.ID = env.PPSSpecCommitID
	return ppsutil.GetPipelineInfo(pachClient, &pipelinePtr)
}

func do(config interface{}) error {
	// must run InstallJaegerTracer before InitWithKube/pach client initialization
	tracing.InstallJaegerTracerFromEnv()
	env := serviceenv.InitServiceEnv(serviceenv.NewConfiguration(config))

	// Construct a client that connects to the sidecar.
	pachClient := env.GetPachClient(context.Background())
	pipelineInfo, err := getPipelineInfo(pachClient, env) // get pipeline creds for pachClient
	if err != nil {
		return fmt.Errorf("error getting pipelineInfo: %v", err)
	}

	// Construct worker API server.
	workerRcName := ppsutil.PipelineRcName(pipelineInfo.Pipeline.Name, pipelineInfo.Version)
	apiServer, err := worker.NewAPIServer(pachClient, env.GetEtcdClient(), env.PPSEtcdPrefix, pipelineInfo, env.PodName, env.Namespace, env.StorageRoot)
	if err != nil {
		return err
	}

	// Start worker api server
	server, err := grpcutil.NewServer(context.Background(), false)
	if err != nil {
		return err
	}

	worker.RegisterWorkerServer(server.Server, apiServer)
	versionpb.RegisterAPIServer(server.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
	debugclient.RegisterDebugServer(server.Server, debugserver.NewDebugServer(env.PodName, env.GetEtcdClient(), env.PPSEtcdPrefix, env.PPSWorkerPort, "", pachClient))

	// Put our IP address into etcd, so pachd can discover us
	key := path.Join(env.PPSEtcdPrefix, worker.WorkerEtcdPrefix, workerRcName, env.PPSWorkerIP)

	// Prepare to write "key" into etcd by creating lease -- if worker dies, our
	// IP will be removed from etcd
	ctx, cancel := context.WithTimeout(pachClient.Ctx(), 10*time.Second)
	defer cancel()
	resp, err := env.GetEtcdClient().Grant(ctx, 10 /* seconds */)
	if err != nil {
		return fmt.Errorf("error granting lease: %v", err)
	}

	// keepalive forever
	if _, err := env.GetEtcdClient().KeepAlive(context.Background(), resp.ID); err != nil {
		return fmt.Errorf("error with KeepAlive: %v", err)
	}

	// Actually write "key" into etcd
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second) // new ctx
	defer cancel()
	if _, err := env.GetEtcdClient().Put(ctx, key, "", etcd.WithLease(resp.ID)); err != nil {
		return fmt.Errorf("error putting IP address: %v", err)
	}

	// If server ever exits, return error
	if _, err := server.ListenTCP("", env.PPSWorkerPort); err != nil {
		return err
	}
	return server.Wait()
}
