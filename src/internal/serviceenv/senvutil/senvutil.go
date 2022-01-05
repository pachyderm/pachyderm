// package senvutil provides helpers for converting from serviceenv types
// to component specific types.
// This package and serviceenv should eventually be removed.
package senvutil

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
)

func PPSConfig(config *serviceenv.Configuration) pps_server.Config {
	return pps_server.Config{
		Namespace:             config.Namespace,
		WorkerImage:           config.WorkerImage,
		WorkerSidecarImage:    config.WorkerSidecarImage,
		WorkerImagePullPolicy: config.WorkerImagePullPolicy,
		StorageRoot:           config.StorageRoot,
		StorageBackend:        config.StorageBackend,
		StorageHostPath:       config.StorageHostPath,
		ImagePullSecrets:      config.ImagePullSecrets,
		WorkerUsesRoot:        config.WorkerUsesRoot,
		PPSWorkerPort:         config.PPSWorkerPort,
		Port:                  config.Port,
		PeerPort:              config.PeerPort,
		GCPercent:             config.GCPercent,
		LokiLogging:           config.LokiLogging,

		PostgresPort:     uint16(config.PostgresPort),
		PostgresHost:     config.PostgresHost,
		PostgresUser:     config.PostgresUser,
		PostgresPassword: config.PostgresPassword,
		PostgresDBName:   config.PostgresDBName,

		PGBouncerHost: config.PGBouncerHost,
		PGBouncerPort: uint16(config.PGBouncerPort),

		DisableCommitProgressCounter:  config.DisableCommitProgressCounter,
		GoogleCloudProfilerProject:    config.GoogleCloudProfilerProject,
		PPSSpecCommitID:               config.PPSSpecCommitID,
		StorageUploadConcurrencyLimit: config.StorageUploadConcurrencyLimit,
		S3GatewayPort:                 config.S3GatewayPort,
		PPSPipelineName:               config.PPSPipelineName,
		PPSMaxConcurrentK8sRequests:   config.PPSMaxConcurrentK8sRequests,
	}
}

func PPSEnv(senv serviceenv.ServiceEnv, txnEnv *transactionenv.TransactionEnv, reporter *metrics.Reporter) pps_server.Env {
	etcdPrefix := path.Join(senv.Config().EtcdPrefix, senv.Config().PPSEtcdPrefix)
	return pps_server.Env{
		DB:            senv.GetDBClient(),
		TxnEnv:        txnEnv,
		Listener:      senv.GetPostgresListener(),
		KubeClient:    senv.GetKubeClient(),
		EtcdClient:    senv.GetEtcdClient(),
		GetLokiClient: senv.GetLokiClient,

		PFSServer:     senv.PfsServer(),
		AuthServer:    senv.AuthServer(),
		GetPachClient: senv.GetPachClient,
		TaskService:   senv.GetTaskService(etcdPrefix),

		Reporter:          reporter,
		BackgroundContext: context.Background(),
		Logger:            senv.Logger(),
	}
}

func PFSConfig(sc *serviceenv.Configuration) pfs_server.Config {
	return pfs_server.Config{
		StorageConfiguration: sc.StorageConfiguration,
	}
}

func PFSEnv(env serviceenv.ServiceEnv, txnEnv *transactionenv.TransactionEnv) (*pfs_server.Env, error) {
	// Setup etcd, object storage, and database clients.
	objClient, err := obj.NewClient(env.Config().StorageBackend, env.Config().StorageRoot)
	if err != nil {
		return nil, err
	}
	etcdPrefix := path.Join(env.Config().EtcdPrefix, env.Config().PFSEtcdPrefix)
	if env.AuthServer() == nil {
		panic("auth server cannot be nil")
	}
	return &pfs_server.Env{
		ObjectClient: objClient,
		DB:           env.GetDBClient(),
		TxnEnv:       txnEnv,
		Listener:     env.GetPostgresListener(),
		EtcdPrefix:   etcdPrefix,
		EtcdClient:   env.GetEtcdClient(),
		TaskService:  env.GetTaskService(etcdPrefix),

		AuthServer:    env.AuthServer(),
		GetPPSServer:  env.PpsServer,
		GetPachClient: env.GetPachClient,

		BackgroundContext: env.Context(),
		Logger:            env.Logger(),
	}, nil
}
