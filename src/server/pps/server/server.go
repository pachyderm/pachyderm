package server

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	ppsiface "github.com/pachyderm/pachyderm/v2/src/server/pps"
	etcd "go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes"
)

// Env contains the dependencies needed to create an API Server
type Env struct {
	DB          *pachsql.DB
	TxnEnv      *txnenv.TransactionEnv
	Listener    collection.PostgresListener
	KubeClient  kubernetes.Interface
	EtcdClient  *etcd.Client
	EtcdPrefix  string
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
	Config            pachconfig.Configuration
	PachwInSidecar    bool
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, reporter *metrics.Reporter) Env {
	etcdPrefix := path.Join(senv.Config().EtcdPrefix, senv.Config().PPSEtcdPrefix)
	return Env{
		DB:            senv.GetDBClient(),
		TxnEnv:        txnEnv,
		Listener:      senv.GetPostgresListener(),
		KubeClient:    senv.GetKubeClient(),
		EtcdClient:    senv.GetEtcdClient(),
		EtcdPrefix:    etcdPrefix,
		TaskService:   senv.GetTaskService(etcdPrefix),
		GetLokiClient: senv.GetLokiClient,

		PFSServer:     senv.PfsServer(),
		AuthServer:    senv.AuthServer(),
		GetPachClient: senv.GetPachClient,

		Reporter:          reporter,
		BackgroundContext: pctx.Child(senv.Context(), "PPS"),
		Config:            *senv.Config(),
		PachwInSidecar:    senv.Config().PachwInSidecars,
	}
}

// NewAPIServer creates an APIServer and runs the master loop in the background
func NewAPIServer(env Env) (ppsiface.APIServer, error) {
	srv, err := NewAPIServerNoMaster(env)
	if err != nil {
		return nil, err
	}
	apiServer := (srv).(*apiServer)
	if env.Config.EnablePreflightChecks {
		apiServer.validateKube(env.BackgroundContext)
	} else {
		log.Error(env.BackgroundContext, "Preflight checks are disabled. This is not recommended.")
	}
	go apiServer.master(env.BackgroundContext)
	go apiServer.worker(env.BackgroundContext)
	return apiServer, nil
}

// NewAPIServerNoMaster creates an APIServer without running the master
// loop in the background.
func NewAPIServerNoMaster(env Env) (ppsiface.APIServer, error) {
	config := env.Config
	apiServer := &apiServer{
		env:                   env,
		txnEnv:                env.TxnEnv,
		etcdPrefix:            env.EtcdPrefix,
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
		clusterDefaults:       ppsdb.ClusterDefaults(env.DB, env.Listener),
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
		env:             env,
		txnEnv:          env.TxnEnv,
		etcdPrefix:      env.EtcdPrefix,
		reporter:        env.Reporter,
		namespace:       namespace,
		workerUsesRoot:  true,
		pipelines:       ppsdb.Pipelines(env.DB, env.Listener),
		jobs:            ppsdb.Jobs(env.DB, env.Listener),
		clusterDefaults: ppsdb.ClusterDefaults(env.DB, env.Listener),
		workerGrpcPort:  workerGrpcPort,
		peerPort:        peerPort,
	}
	go apiServer.ServeSidecarS3G(pctx.Child(env.BackgroundContext, "s3gateway", pctx.WithServerID()))
	return apiServer, nil
}
