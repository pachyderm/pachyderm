package serviceenv

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	auth_server "github.com/pachyderm/pachyderm/v2/src/server/auth"
	enterprise_server "github.com/pachyderm/pachyderm/v2/src/server/enterprise"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps"

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
	Configuration    *Configuration
	PachClient       *client.APIClient
	EtcdClient       *etcd.Client
	KubeClient       *kube.Clientset
	LokiClient       *loki.Client
	DBClient         *sqlx.DB
	PostgresListener col.PostgresListener
	DexDB            dex_storage.Storage
	Log              *log.Logger
	Ctx              context.Context

	// Auth is the registered auth APIServer
	Auth auth_server.APIServer

	// Pps is the registered pps APIServer
	Pps pps_server.APIServer

	// Pfs is the registered pfs APIServer
	Pfs pfs_server.APIServer

	// Enterprise is the registered pfs APIServer
	Enterprise enterprise_server.APIServer

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
func (s *TestServiceEnv) GetPostgresListener() col.PostgresListener {
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

// AuthServer returns the registered PFS APIServer
func (env *TestServiceEnv) AuthServer() auth_server.APIServer {
	return env.Auth
}

// PpsServer returns the registered PPS APIServer
func (env *TestServiceEnv) PpsServer() pps_server.APIServer {
	return env.Pps
}

// PfsServer returns the registered PFS APIServer
func (env *TestServiceEnv) PfsServer() pfs_server.APIServer {
	return env.Pfs
}

// SetAuthServer returns the registered PFS APIServer
func (env *TestServiceEnv) SetAuthServer(s auth_server.APIServer) {
	env.Auth = s
}

// SetPpsServer returns the registered PPS APIServer
func (env *TestServiceEnv) SetPpsServer(s pps_server.APIServer) {
	env.Pps = s
}

// SetPfsServer returns the registered PFS APIServer
func (env *TestServiceEnv) SetPfsServer(s pfs_server.APIServer) {
	env.Pfs = s
}

// EnterpriseServer returns the registered Enterprise APIServer
func (env *TestServiceEnv) EnterpriseServer() enterprise_server.APIServer {
	return env.Enterprise
}

// SetEnterpriseServer returns the registered Enterprise APIServer
func (env *TestServiceEnv) SetEnterpriseServer(s enterprise_server.APIServer) {
	env.Enterprise = s
}
