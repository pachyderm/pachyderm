package serviceenv

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"

	etcd "github.com/coreos/etcd/clientv3"
	loki "github.com/grafana/loki/pkg/logcli/client"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	kube "k8s.io/client-go/kubernetes"
)

// TestServiceEnv is a simple implementation of ServiceEnv that can be constructed with
// existing clients.
type TestServiceEnv struct {
	Configuration *Configuration
	PachClient    *client.APIClient
	EtcdClient    *etcd.Client
	KubeClient    *kube.Clientset
	LokiClient    *loki.Client
	DBClient      *sqlx.DB
	Ctx           context.Context
	Logger        *logrus.Logger
	Context       context.Context
}

func (s *TestServiceEnv) Config() *Configuration {
	return s.Configuration
}

func (s *TestServiceEnv) GetPachClient(ctx context.Context) *client.APIClient {
	return s.PachClient.WithCtx(ctx)
}
func (s *TestServiceEnv) GetEtcdClient() *etcd.Client {
	return s.EtcdClient
}
func (s *TestServiceEnv) GetKubeClient() *kube.Clientset {
	return s.KubeClient
}
func (s *TestServiceEnv) GetLokiClient() (*loki.Client, error) {
	return s.LokiClient, nil
}
func (s *TestServiceEnv) GetDBClient() *sqlx.DB {
	return s.DBClient
}

func (s *TestServiceEnv) Context() context.Context {
	return s.Ctx
}

func (s *TestServiceEnv) ClusterID() string {
	return "testing"
}

func (s *TestServiceEnv) Close() error {
	eg := &errgroup.Group{}
	eg.Go(s.GetPachClient(context.Background()).Close)
	eg.Go(s.GetEtcdClient().Close)
	eg.Go(s.GetDBClient().Close)
	return eg.Wait()
}

func (s *TestServiceEnv) GetLogger(service string) log.Logger {
	return log.NewWithLogger(service, s.Logger)
}
func (s *TestServiceEnv) GetContext() context.Context {
	return s.Context
}
