package server

import (
	"context"
	"path"

	"github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
)

// Env is the dependencies needed to run the pachW Master
type Env struct {
	EtcdPrefix        string
	EtcdClient        *etcd.Client
	TaskService       task.Service
	KubeClient        kubernetes.Interface
	Namespace         string
	Enabled           bool
	MaxReplicas       int
	MinReplicas       int
	BackgroundContext context.Context
	Logger            *logrus.Logger
}

func EnvFromServiceEnv(env serviceenv.ServiceEnv) (*Env, error) {
	etcdPrefix := path.Join(env.Config().EtcdPrefix, env.Config().PFSEtcdPrefix)
	if env.AuthServer() == nil {
		panic("auth server cannot be nil")
	}
	return &Env{
		EtcdPrefix:        etcdPrefix,
		EtcdClient:        env.GetEtcdClient(),
		TaskService:       env.GetTaskService(etcdPrefix),
		KubeClient:        env.GetKubeClient(),
		Namespace:         env.Config().Namespace,
		Enabled:           env.Config().PachwEnabled,
		MinReplicas:       env.Config().PachwMinReplicas,
		MaxReplicas:       env.Config().PachwMaxReplicas,
		BackgroundContext: env.Context(),
		Logger:            env.Logger(),
	}, nil
}
