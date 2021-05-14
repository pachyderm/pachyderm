package serviceenv

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	auth_server "github.com/pachyderm/pachyderm/v2/src/server/auth"

	etcd "github.com/coreos/etcd/clientv3"
	dex_storage "github.com/dexidp/dex/storage"
	loki "github.com/grafana/loki/pkg/logcli/client"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	kube "k8s.io/client-go/kubernetes"
)

// TestServiceEnv is a simple implementation of ServiceEnv that can be constructed with
// existing clients.
type TestServiceEnv struct {
	Auth auth_server.APIServer

	Configuration    *Configuration
	PachClient       *client.APIClient
	EtcdClient       *etcd.Client
	KubeClient       *kube.Clientset
	LokiClient       *loki.Client
	DBClient         *sqlx.DB
	PostgresListener *col.PostgresListener
	DexDB            dex_storage.Storage
	Log              *log.Logger
	Ctx              context.Context

	// Ready is a channel that blocks `GetPachClient` until it's closed.
	// This avoids a race when we need to instantiate the server before
	// getting a client pointing at the same server.
	Ready chan interface{}
}

func (s *TestServiceEnv) Config() *Configuration {
	return s.Configuration
}

func (s *TestServiceEnv) GetPachClient(ctx context.Context) *client.APIClient {
	<-s.Ready
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
func (s *TestServiceEnv) GetPostgresListener() *col.PostgresListener {
	return s.PostgresListener
}

func (s *TestServiceEnv) Context() context.Context {
	return s.Ctx
}

func (s *TestServiceEnv) ClusterID() string {
	return "testing"
}

func (s *TestServiceEnv) Logger() *log.Logger {
	return s.Log
}

func (s *TestServiceEnv) GetDexDB() dex_storage.Storage {
	return s.DexDB
}

func (s *TestServiceEnv) Close() error {
	eg := &errgroup.Group{}
	if client := s.GetPachClient(context.Background()); client != nil {
		eg.Go(client.Close)
	}
	if client := s.GetEtcdClient(); client != nil {
		eg.Go(client.Close)
	}
	if client := s.GetDBClient(); client != nil {
		eg.Go(client.Close)
	}
	if client := s.GetDexDB(); client != nil {
		eg.Go(client.Close)
	}
	if listener := s.GetPostgresListener(); listener != nil {
		eg.Go(listener.Close)
	}
	return eg.Wait()
}

func (s *TestServiceEnv) AuthServer() auth_server.APIServer {
	return s.Auth
}
