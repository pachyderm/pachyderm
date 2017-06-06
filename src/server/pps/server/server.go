package server

import (
	"path"
	"sync"

	"github.com/pachyderm/pachyderm/src/client"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"

	etcd "github.com/coreos/etcd/clientv3"
	"go.pedge.io/proto/rpclog"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	pipelinesPrefix = "/pipelines"
	jobsPrefix      = "/jobs"
)

var (
	// Index mapping pipeline to jobs started by the pipeline
	jobsPipelineIndex = col.Index{"Pipeline", false}

	// Index mapping job inputs (repos + pipeline version) to output commit. This
	// is how we know if we need to start a job
	jobsInputIndex = col.Index{"Input", false}

	// Index mapping 1.4.5 and earlier style job inputs (repos + pipeline
	// version) to output commit. This is how we know if we need to start a job
	// Needed for legacy compatibility.
	jobsInputsIndex = col.Index{"Inputs", false}

	// Index of pipelines and jobs that have been stopped (state is "success" or
	// "failure" for jobs, or "stopped" or "failure" for pipelines). See
	// (Job|Pipeline)StateToStopped in s/s/pps/server/api_server.go
	stoppedIndex = col.Index{"Stopped", false}
)

// NewAPIServer creates an APIServer.
func NewAPIServer(
	etcdAddress string,
	etcdPrefix string,
	hasher *ppsserver.Hasher,
	address string,
	kubeClient *kube.Client,
	namespace string,
	workerImage string,
	workerSidecarImage string,
	workerImagePullPolicy string,
	storageRoot string,
	storageBackend string,
	storageHostPath string,
	reporter *metrics.Reporter,
) (ppsclient.APIServer, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return nil, err
	}

	apiServer := &apiServer{
		Logger:                protorpclog.NewLogger("pps.API"),
		etcdPrefix:            etcdPrefix,
		hasher:                hasher,
		address:               address,
		etcdClient:            etcdClient,
		pachConnOnce:          sync.Once{},
		kubeClient:            kubeClient,
		namespace:             namespace,
		workerImage:           workerImage,
		workerSidecarImage:    workerSidecarImage,
		workerImagePullPolicy: workerImagePullPolicy,
		storageRoot:           storageRoot,
		storageBackend:        storageBackend,
		storageHostPath:       storageHostPath,
		reporter:              reporter,
		pipelines: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, pipelinesPrefix),
			[]col.Index{stoppedIndex},
			&ppsclient.PipelineInfo{},
		),
		jobs: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, jobsPrefix),
			[]col.Index{jobsPipelineIndex, stoppedIndex, jobsInputIndex},
			&ppsclient.JobInfo{},
		),
	}
	go apiServer.master()
	return apiServer, nil
}
