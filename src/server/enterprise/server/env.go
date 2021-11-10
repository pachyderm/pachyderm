package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/server/auth"
	logrus "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Env struct {
	EtcdClient *clientv3.Client
	EtcdPrefix string

	AuthServer    auth.APIServer
	GetPachClient func(context.Context) *client.APIClient

	Logger            *logrus.Logger
	BackgroundContext context.Context
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv, etcdPrefix string) Env {
	return Env{
		EtcdClient: senv.GetEtcdClient(),
		EtcdPrefix: etcdPrefix,

		AuthServer:    senv.AuthServer(),
		GetPachClient: senv.GetPachClient,

		Logger:            senv.Logger(),
		BackgroundContext: senv.Context(),
	}
}
