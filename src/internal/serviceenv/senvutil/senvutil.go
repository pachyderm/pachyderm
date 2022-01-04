package senvutil

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
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

		Reporter:          reporter,
		BackgroundContext: context.Background(),
		Logger:            senv.Logger(),
	}
}
