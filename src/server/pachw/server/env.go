package server

import (
	"context"

	etcd "go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/v2/src/internal/task"
)

// Env is the dependencies needed to run the pachW Controller
type Env struct {
	EtcdPrefix        string
	EtcdClient        *etcd.Client
	TaskService       task.Service
	KubeClient        kubernetes.Interface
	Namespace         string
	MaxReplicas       int
	MinReplicas       int
	BackgroundContext context.Context
}
