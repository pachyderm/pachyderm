package serviceenv

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"

	etcd "github.com/coreos/etcd/clientv3"
	loki "github.com/grafana/loki/pkg/logcli/client"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
	kube "k8s.io/client-go/kubernetes"
)

// TestServiceEnv is a simple implementation of ServiceEnv that can be constructed with
// existing clients.
type TestServiceEnv struct {
	Configuration    *Configuration
	PachClient       *client.APIClient
	EtcdClient       *etcd.Client
	KubeClient       *kube.Clientset
	LokiClient       *loki.Client
	DBClient         *sqlx.DB
	PostgresListener *col.PostgresListener
	Ctx              context.Context
}

func (s *TestServiceEnv) Config() *Configuration {
	return s.Configuration
}

func (s *TestServiceEnv) GetPachClient(ctx context.Context) *client.APIClient {
	return s.PachClient
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
func (s *TestServiceEnv) GetPostgresListener() *col.PostgresListener {
	return s.PostgresListener
}

func (s *TestServiceEnv) Context() context.Context {
	return s.Ctx
}

func (s *TestServiceEnv) Close() error {
	eg := &errgroup.Group{}
	eg.Go(s.GetPachClient(context.Background()).Close)
	eg.Go(s.GetEtcdClient().Close)
	eg.Go(s.GetDBClient().Close)
	eg.Go(s.GetPostgresListener().Close)
	return eg.Wait()
}
