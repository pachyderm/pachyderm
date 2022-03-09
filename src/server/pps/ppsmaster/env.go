package ppsmaster

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
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
	Logger            *logrus.Logger
	Config            serviceenv.Configuration
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
		BackgroundContext: context.Background(),
		Logger:            senv.Logger(),
		Config:            *senv.Config(),
	}
}
