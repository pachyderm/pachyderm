package pachd

import (
	"path"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	admin_server "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
	auth_server "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	debug_server "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	enterprise_server "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	license_server "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	pachw_server "github.com/pachyderm/pachyderm/v2/src/server/pachw/server"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
)

func AdminEnv(senv serviceenv.ServiceEnv, paused bool) admin_server.Env {
	return admin_server.Env{
		ClusterID: senv.ClusterID(),
		Config:    senv.Config(),
		PFSServer: senv.PfsServer(),
		Paused:    paused,
		DB:        senv.GetDBClient(),
	}
}

func AuthEnv(senv serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv) auth_server.Env {
	return auth_server.Env{
		DB:         senv.GetDBClient(),
		EtcdClient: senv.GetEtcdClient(),
		Listener:   senv.GetPostgresListener(),
		TxnEnv:     txnEnv,

		GetEnterpriseServer: senv.EnterpriseServer,
		GetIdentityServer:   senv.IdentityServer,
		GetPfsServer:        senv.PfsServer,
		GetPpsServer:        senv.PpsServer,

		BackgroundContext: pctx.Child(senv.Context(), "auth"),
		Config:            *senv.Config(),
	}
}

func EnterpriseEnv(senv serviceenv.ServiceEnv, etcdPrefix string, txEnv *txnenv.TransactionEnv) *enterprise_server.Env {
	e := &enterprise_server.Env{
		DB:       senv.GetDBClient(),
		Listener: senv.GetPostgresListener(),
		TxnEnv:   txEnv,

		EtcdClient: senv.GetEtcdClient(),
		EtcdPrefix: etcdPrefix,

		AuthServer:    senv.AuthServer(),
		GetPachClient: senv.GetPachClient,
		GetKubeClient: senv.GetKubeClient,

		BackgroundContext: senv.Context(),
		Namespace:         senv.Config().Namespace,
		Config:            *senv.Config(),
	}
	return e
}

func LicenseEnv(senv serviceenv.ServiceEnv) *license_server.Env {
	return &license_server.Env{
		DB:               senv.GetDBClient(),
		Listener:         senv.GetPostgresListener(),
		Config:           senv.Config(),
		EnterpriseServer: senv.EnterpriseServer(),
	}
}

func PachwEnv(env serviceenv.ServiceEnv) (*pachw_server.Env, error) {
	PFSEtcdPrefix := path.Join(env.Config().EtcdPrefix, env.Config().PFSEtcdPrefix)
	PPSEtcdPrefix := path.Join(env.Config().EtcdPrefix, env.Config().PPSEtcdPrefix)
	if env.AuthServer() == nil {
		panic("auth server cannot be nil")
	}
	return &pachw_server.Env{
		EtcdPrefix:        env.Config().EtcdPrefix,
		EtcdClient:        env.GetEtcdClient(),
		PFSTaskService:    env.GetTaskService(PFSEtcdPrefix),
		PPSTaskService:    env.GetTaskService(PPSEtcdPrefix),
		KubeClient:        env.GetKubeClient(),
		Namespace:         env.Config().Namespace,
		MinReplicas:       env.Config().PachwMinReplicas,
		MaxReplicas:       env.Config().PachwMaxReplicas,
		BackgroundContext: env.Context(),
	}, nil
}

func PFSEnv(env serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv) (*pfs_server.Env, error) {
	if env.AuthServer() == nil {
		panic("auth server cannot be nil")
	}
	cfg := env.Config()
	bucket, err := obj.NewBucket(env.Context(), cfg.StorageBackend, cfg.StorageRoot, cfg.StorageURL)
	if err != nil {
		return nil, errors.Wrap(err, "pfs env")
	}
	etcdPrefix := path.Join(cfg.EtcdPrefix, cfg.PFSEtcdPrefix)
	return &pfs_server.Env{
		Bucket:      bucket,
		DB:          env.GetDBClient(),
		TxnEnv:      txnEnv,
		Listener:    env.GetPostgresListener(),
		EtcdPrefix:  etcdPrefix,
		EtcdClient:  env.GetEtcdClient(),
		TaskService: env.GetTaskService(etcdPrefix),

		Auth:                 env.AuthServer(),
		GetPipelineInspector: func() pfs_server.PipelineInspector { return env.PpsServer() },

		StorageConfig: cfg.StorageConfiguration,
		GetPPSServer:  env.PpsServer,
	}, nil
}

func StorageEnv(env serviceenv.ServiceEnv) (*storage.Env, error) {
	cfg := env.Config()
	bucket, err := obj.NewBucket(env.Context(), cfg.StorageBackend, cfg.StorageRoot, cfg.StorageURL)
	if err != nil {
		return nil, errors.Wrap(err, "storage env")
	}
	return &storage.Env{
		Bucket: bucket,
		DB:     env.GetDBClient(),
		Config: cfg.StorageConfiguration,
	}, nil

}

func PFSWorkerEnv(env serviceenv.ServiceEnv) (*pfs_server.WorkerEnv, error) {
	cfg := env.Config()
	bucket, err := obj.NewBucket(env.Context(), cfg.StorageBackend, cfg.StorageRoot, cfg.StorageURL)
	if err != nil {
		return nil, errors.Wrap(err, "pfs worker")
	}
	etcdPrefix := path.Join(cfg.EtcdPrefix, cfg.PFSEtcdPrefix)
	return &pfs_server.WorkerEnv{
		Bucket:      bucket,
		DB:          env.GetDBClient(),
		TaskService: env.GetTaskService(etcdPrefix),
	}, nil
}

func PPSEnv(senv serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, reporter *metrics.Reporter) pps_server.Env {
	etcdPrefix := path.Join(senv.Config().EtcdPrefix, senv.Config().PPSEtcdPrefix)
	return pps_server.Env{
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
	}
}

func DebugEnv(env serviceenv.ServiceEnv) debug_server.Env {
	return debug_server.Env{
		Config:               *env.Config(),
		Name:                 env.Config().PachdPodName,
		DB:                   env.GetDBClient(),
		SidecarClient:        nil,
		GetKubeClient:        env.GetKubeClient,
		GetDynamicKubeClient: env.GetDynamicKubeClient,
		GetLokiClient:        env.GetLokiClient,
		GetPachClient:        env.GetPachClient,
		TaskService:          env.GetTaskService(env.Config().EtcdPrefix),
	}
}
