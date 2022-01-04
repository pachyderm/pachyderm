package server

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	ppsiface "github.com/pachyderm/pachyderm/v2/src/server/pps"
	logrus "github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes"
)

// Env contains the dependencies needed to create an API Server
type Env struct {
	DB          *pachsql.DB
	TxnEnv      *txnenv.TransactionEnv
	Listener    collection.PostgresListener
	KubeClient  *kubernetes.Clientset
	EtcdClient  *etcd.Client
	TaskService task.Service
	// TODO: make this just a *loki.Client
	// This is not a circular dependency
	GetLokiClient func() (*loki.Client, error)

	PFSServer  pfsserver.APIServer
	AuthServer authserver.APIServer
	// TODO: This should just be a pach client for the needed services.
	// serviceenv blocks until everything is done though, so we can't get it until after setup is done.
	GetPachClient func(context.Context) *client.APIClient

	Reporter          *metrics.Reporter
	BackgroundContext context.Context
	Logger            *logrus.Logger
}

// Config is configuration parameters for the API Server
// Config does not contain instantiated dependencies.
// If it is a user provided, primitive value (bool, string, int, etc.), then it belongs in here.
type Config struct {
	Namespace string

	WorkerImage           string
	WorkerSidecarImage    string
	WorkerImagePullPolicy string
	ImagePullSecrets      string
	WorkerUsesRoot        bool

	StorageRoot     string
	StorageBackend  string
	StorageHostPath string

	EtcdPrefix    string
	PPSEtcdPrefix string

	PPSWorkerPort uint16
	Port          uint16
	PeerPort      uint16
	GCPercent     int

	LokiLogging bool

	PostgresPort     uint16
	PostgresHost     string
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string

	PGBouncerHost string
	PGBouncerPort uint16

	DisableCommitProgressCounter  bool
	GoogleCloudProfilerProject    string
	PPSSpecCommitID               string
	StorageUploadConcurrencyLimit int
	S3GatewayPort                 uint16
	PPSPipelineName               string
	PPSMaxConcurrentK8sRequests   int
}

// NewAPIServer creates an APIServer.
func NewAPIServer(env Env, config Config) (ppsiface.APIServer, error) {
	srv, err := NewAPIServerNoMaster(env, config)
	if err != nil {
		return nil, err
	}
	apiServer := srv.(*apiServer)
	apiServer.validateKube(apiServer.env.BackgroundContext)
	go apiServer.master()
	return srv, nil
}

func NewAPIServerNoMaster(env Env, config Config) (ppsiface.APIServer, error) {
	etcdPrefix := path.Join(config.EtcdPrefix, config.PPSEtcdPrefix)
	apiServer := &apiServer{
		Logger:                log.NewLogger("pps.API", env.Logger),
		env:                   env,
		txnEnv:                env.TxnEnv,
		etcdPrefix:            etcdPrefix,
		namespace:             config.Namespace,
		workerImage:           config.WorkerImage,
		workerSidecarImage:    config.WorkerSidecarImage,
		workerImagePullPolicy: config.WorkerImagePullPolicy,
		storageRoot:           config.StorageRoot,
		storageBackend:        config.StorageBackend,
		storageHostPath:       config.StorageHostPath,
		imagePullSecrets:      config.ImagePullSecrets,
		reporter:              env.Reporter,
		workerUsesRoot:        config.WorkerUsesRoot,
		pipelines:             ppsdb.Pipelines(env.DB, env.Listener),
		jobs:                  ppsdb.Jobs(env.DB, env.Listener),
		workerGrpcPort:        config.PPSWorkerPort,
		port:                  config.Port,
		peerPort:              config.PeerPort,
		gcPercent:             config.GCPercent,
	}
	return apiServer, nil
}

// NewSidecarAPIServer creates an APIServer that has limited functionalities
// and is meant to be run as a worker sidecar.  It cannot, for instance,
// create pipelines.
func NewSidecarAPIServer(
	env Env,
	namespace string,
	workerGrpcPort uint16,
	peerPort uint16,
) (*apiServer, error) {
	apiServer := &apiServer{
		Logger:         log.NewLogger("pps.API", env.Logger),
		env:            env,
		txnEnv:         env.TxnEnv,
		reporter:       env.Reporter,
		namespace:      namespace,
		workerUsesRoot: true,
		pipelines:      ppsdb.Pipelines(env.DB, env.Listener),
		jobs:           ppsdb.Jobs(env.DB, env.Listener),
		workerGrpcPort: workerGrpcPort,
		peerPort:       peerPort,
	}
	go apiServer.ServeSidecarS3G()
	return apiServer, nil
}
