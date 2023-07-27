package serviceenv

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"time"

	dex_storage "github.com/dexidp/dex/storage"
	"github.com/dlmiddlecote/sqlstats"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	mlc "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	auth_server "github.com/pachyderm/pachyderm/v2/src/server/auth"
	enterprise_server "github.com/pachyderm/pachyderm/v2/src/server/enterprise"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps"
)

const clusterIDKey = "cluster-id"

// ServiceEnv contains connections to other services in the
// cluster. In pachd, there is only one instance of this struct, but tests may
// create more, if they want to create multiple pachyderm "clusters" served in
// separate goroutines.
type ServiceEnv interface {
	AuthServer() auth_server.APIServer
	IdentityServer() identity.APIServer
	PfsServer() pfs_server.APIServer
	PpsServer() pps_server.APIServer
	EnterpriseServer() enterprise_server.APIServer
	SetAuthServer(auth_server.APIServer)
	SetIdentityServer(identity.APIServer)
	SetPfsServer(pfs_server.APIServer)
	SetPpsServer(pps_server.APIServer)
	SetEnterpriseServer(enterprise_server.APIServer)
	SetKubeClient(kube.Interface)

	Config() *pachconfig.Configuration
	GetPachClient(ctx context.Context) *client.APIClient
	GetEtcdClient() *etcd.Client
	GetTaskService(string) task.Service
	GetKubeClient() kube.Interface
	GetLokiClient() (*loki.Client, error)
	GetDBClient() *pachsql.DB
	GetDirectDBClient() *pachsql.DB
	GetPostgresListener() col.PostgresListener
	InitDexDB()
	GetDexDB() dex_storage.Storage
	ClusterID() string
	Context() context.Context
	Close() error
}

// NonblockingServiceEnv is an implementation of ServiceEnv that initializes
// clients in the background, and blocks in the getters until they're ready.
type NonblockingServiceEnv struct {
	config *pachconfig.Configuration

	// pachAddress is the domain name or hostport where pachd can be reached
	pachAddress string
	// pachClient is the "template" client other clients returned by this library
	// are based on. It contains the original GRPC client connection and has no
	// ctx and therefore no auth credentials or cancellation
	pachClient *client.APIClient
	// pachEg coordinates the initialization of pachClient.  Note that NonblockingServiceEnv
	// uses a separate error group for each client, rather than one for all
	// three clients, so that pachd can initialize a NonblockingServiceEnv inside of its own
	// initialization (if GetEtcdClient() blocked on intialization of 'pachClient'
	// and pachd/main.go couldn't start the pachd server until GetEtcdClient() had
	// returned, then pachd would be unable to start)
	pachEg errgroup.Group

	// etcdAddress is the domain name or hostport where etcd can be reached
	etcdAddress string
	// etcdClient is an etcd client that's shared by all users of this environment
	etcdClient *etcd.Client
	// etcdEg coordinates the initialization of etcdClient (see pachdEg)
	etcdEg errgroup.Group

	// kubeClient is a kubernetes client that, if initialized, is shared by all
	// users of this environment
	kubeClient kube.Interface
	// kubeEg coordinates the initialization of kubeClient (see pachdEg)
	kubeEg errgroup.Group

	// lokiClient is a loki (log aggregator) client that is shared by all users
	// of this environment, it doesn't require an initialization funcion, so
	// there's no errgroup associated with it.
	lokiClient *loki.Client

	// clusterId is the unique ID for this pach cluster
	clusterId   string
	clusterIdEg errgroup.Group

	// dexDB is a dex_storage connected to postgres
	dexDB   dex_storage.Storage
	dexDBEg errgroup.Group

	// dbClient and directDBClient are database clients.
	dbClient, directDBClient *pachsql.DB
	// dbEg coordinates the initialization of dbClient (see pachdEg)
	dbEg errgroup.Group

	// listener is a special database client for listening for changes
	listener col.PostgresListener

	authServer       auth_server.APIServer
	identityServer   identity.APIServer
	ppsServer        pps_server.APIServer
	pfsServer        pfs_server.APIServer
	enterpriseServer enterprise_server.APIServer

	// ctx is the background context for the environment that will be canceled
	// when the ServiceEnv is closed - this typically only happens for orderly
	// shutdown in tests
	ctx    context.Context
	cancel func()
}

func goCtx(eg *errgroup.Group, ctx context.Context, f func(context.Context) error) {
	eg.Go(func() error {
		return f(ctx)
	})
}

// InitPachOnlyEnv initializes this service environment. This dials a GRPC
// connection to pachd only (in a background goroutine), and creates the
// template pachClient used by future calls to GetPachClient.
//
// This call returns immediately, but GetPachClient will block
// until the client is ready.
func InitPachOnlyEnv(rctx context.Context, config *pachconfig.Configuration) *NonblockingServiceEnv {
	sctx, end := log.SpanContext(rctx, "serviceenv")
	ctx, cancel := pctx.WithCancel(sctx)
	env := &NonblockingServiceEnv{config: config, ctx: ctx, cancel: func() { cancel(); end() }}
	env.pachAddress = net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", env.config.PeerPort))
	goCtx(&env.pachEg, ctx, env.initPachClient)
	return env // env is not ready yet
}

// InitServiceEnv initializes this service environment. This dials a GRPC
// connection to pachd and etcd (in a background goroutine), and creates the
// template pachClient used by future calls to GetPachClient.
//
// This call returns immediately, but GetPachClient and GetEtcdClient block
// until their respective clients are ready.
func InitServiceEnv(ctx context.Context, config *pachconfig.Configuration) *NonblockingServiceEnv {
	env := InitPachOnlyEnv(ctx, config)
	env.etcdAddress = fmt.Sprintf("http://%s", net.JoinHostPort(env.config.EtcdHost, env.config.EtcdPort))
	goCtx(&env.etcdEg, env.ctx, env.initEtcdClient)
	goCtx(&env.clusterIdEg, env.ctx, env.initClusterID)
	goCtx(&env.dbEg, env.ctx, env.initDBClient)
	if !env.isWorker() {
		goCtx(&env.dbEg, env.ctx, env.initDirectDBClient)
	}
	env.listener = env.newListener()
	if lokiHost, lokiPort := env.config.LokiHost, env.config.LokiPort; lokiHost != "" && lokiPort != "" {
		env.lokiClient = &loki.Client{
			Address: fmt.Sprintf("http://%s", net.JoinHostPort(lokiHost, lokiPort)),
		}
	}
	return env // env is not ready yet
}

// InitWithKube is like InitNonblockingServiceEnv, but also assumes that it's run inside
// a kubernetes cluster and tries to connect to the kubernetes API server.
func InitWithKube(ctx context.Context, config *pachconfig.Configuration) *NonblockingServiceEnv {
	env := InitServiceEnv(ctx, config)
	goCtx(&env.kubeEg, env.ctx, env.initKubeClient)
	return env // env is not ready yet
}

// InitDBOnlyEnv is like InitServiceEnv, but only connects to the database.
func InitDBOnlyEnv(rctx context.Context, config *pachconfig.Configuration) *NonblockingServiceEnv {
	sctx, end := log.SpanContext(rctx, "serviceenv")
	ctx, cancel := pctx.WithCancel(sctx)
	env := &NonblockingServiceEnv{config: config, ctx: ctx, cancel: func() { cancel(); end() }}
	goCtx(&env.dbEg, env.ctx, env.initDBClient)
	return env
}

func (env *NonblockingServiceEnv) Config() *pachconfig.Configuration {
	return env.config
}

func (env *NonblockingServiceEnv) isWorker() bool {
	return env.config.PPSPipelineName != "" || env.config.IsPachw
}

func (env *NonblockingServiceEnv) initClusterID(ctx context.Context) (retErr error) {
	ctx, end := log.SpanContext(ctx, "initClusterID")
	defer end(log.Errorp(&retErr))
	client := env.GetEtcdClient()
	for {
		resp, err := client.Get(ctx, clusterIDKey)
		switch {
		case err != nil:
			log.Debug(ctx, "potential race setting cluster id", zap.Error(err))
			return errors.EnsureStack(err)
		case resp.Count == 0:
			// This might error if it races with another pachd trying to set the
			// cluster id so we ignore the error.
			id := uuid.NewWithoutDashes()
			done := log.Span(ctx, "putClusterID", zap.String("clusterID", id))
			_, err := client.Put(ctx, clusterIDKey, id)
			done(zap.Error(err))
			// Now iterate again and make sure we read it back.
		default:
			log.Debug(ctx, "cluster id already set", zap.ByteString("clusterID", resp.Kvs[0].Value))
			// We expect there to only be one value for this key
			env.clusterId = string(resp.Kvs[0].Value)
			return nil
		}
	}
}

func (env *NonblockingServiceEnv) initPachClient(ctx context.Context) error {
	// validate argument
	if env.pachAddress == "" {
		return errors.New("cannot initialize pach client with empty pach address")
	}
	// Initialize pach client
	return backoff.Retry(func() (retErr error) {
		defer log.Span(ctx, "initPachClient")(log.Errorp(&retErr))
		pachClient, err := client.NewFromURI(
			ctx,
			env.pachAddress,
			client.WithAdditionalUnaryClientInterceptors(grpc_prometheus.UnaryClientInterceptor, mlc.LogUnary),
			client.WithAdditionalStreamClientInterceptors(grpc_prometheus.StreamClientInterceptor, mlc.LogStream),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to initialize pach client")
		}
		env.pachClient = pachClient
		return nil
	}, backoff.RetryEvery(time.Second).For(5*time.Minute))
}

func (env *NonblockingServiceEnv) initEtcdClient(ctx context.Context) error {
	// validate argument
	if env.etcdAddress == "" {
		return errors.New("cannot initialize etcd client with empty etcd address")
	}
	// Initialize etcd
	opts := client.DefaultDialOptions() // SA1019 can't call grpc.Dial directly
	opts = append(opts,
		grpc.WithChainUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithChainStreamInterceptor(grpc_prometheus.StreamClientInterceptor))

	return backoff.Retry(func() (retErr error) {
		defer log.Span(ctx, "initEtcdClient")(log.Errorp(&retErr))
		cfg := log.GetEtcdClientConfig(env.Context())
		cfg.Endpoints = []string{env.etcdAddress}
		// Use a long timeout with Etcd so that Pachyderm doesn't crash loop
		// while waiting for etcd to come up (makes startup net faster)
		cfg.DialTimeout = 3 * time.Minute
		cfg.DialOptions = opts
		cfg.MaxCallSendMsgSize = math.MaxInt32
		cfg.MaxCallRecvMsgSize = math.MaxInt32

		var err error
		env.etcdClient, err = etcd.New(cfg)
		if err != nil {
			return errors.Wrapf(err, "failed to initialize etcd client")
		}
		return nil
	}, backoff.RetryEvery(time.Second).For(5*time.Minute))
}

func (env *NonblockingServiceEnv) initKubeClient(ctx context.Context) error {
	return backoff.Retry(func() (retErr error) {
		ctx, end := log.SpanContext(ctx, "initKubeClient")
		defer end(log.Errorp(&retErr))

		// Get secure in-cluster config
		var kubeAddr string
		var ok bool
		cfg, err := rest.InClusterConfig()
		if err != nil {
			// InClusterConfig failed, fall back to insecure config
			log.Info(ctx, "falling back to insecure kube client due to error from NewInCluster", zap.Error(err))
			kubeAddr, ok = os.LookupEnv("KUBERNETES_PORT_443_TCP_ADDR")
			if !ok {
				return errors.Wrapf(err, "can't fall back to insecure kube client due to missing env var (failed to retrieve in-cluster config")
			}
			kubePort, ok := os.LookupEnv("KUBERNETES_PORT")
			if !ok {
				kubePort = ":443"
			}
			bearerToken, _ := os.LookupEnv("KUBERNETES_BEARER_TOKEN_FILE")
			cfg = &rest.Config{
				Host:            fmt.Sprintf("%s:%s", kubeAddr, kubePort),
				BearerTokenFile: bearerToken,
				TLSClientConfig: rest.TLSClientConfig{
					Insecure: true,
				},
			}
		}
		cfg.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
			return promutil.InstrumentRoundTripper("kubernetes", rt)
		}
		env.kubeClient, err = kube.NewForConfig(cfg)
		if err != nil {
			return errors.Wrapf(err, "could not initialize kube client")
		}
		return nil
	}, backoff.RetryEvery(time.Second).For(5*time.Minute))
}

func (env *NonblockingServiceEnv) initDirectDBClient(ctx context.Context) error {
	ctx, end := log.SpanContext(ctx, "initDirectDBClient")
	defer end()

	db, err := dbutil.NewDB(
		dbutil.WithHostPort(env.config.PostgresHost, env.config.PostgresPort),
		dbutil.WithDBName(env.config.PostgresDBName),
		dbutil.WithUserPassword(env.config.PostgresUser, env.config.PostgresPassword),
		dbutil.WithMaxOpenConns(env.config.PostgresMaxOpenConns),
		dbutil.WithMaxIdleConns(env.config.PostgresMaxIdleConns),
		dbutil.WithConnMaxLifetime(time.Duration(env.config.PostgresConnMaxLifetimeSeconds)*time.Second),
		dbutil.WithConnMaxIdleTime(time.Duration(env.config.PostgresConnMaxIdleSeconds)*time.Second),
		dbutil.WithSSLMode(env.config.PostgresSSL),
		dbutil.WithQueryLog(env.config.PostgresQueryLogging, "pgx.direct"),
	)
	if err != nil {
		return err
	}
	env.directDBClient = db
	if err != nil {
		return err
	}
	if err := prometheus.Register(sqlstats.NewStatsCollector("direct", env.directDBClient.DB)); err != nil {
		log.Info(ctx, "problem registering stats collector for direct db client", zap.Error(err))
	}
	return nil
}

func (env *NonblockingServiceEnv) initDBClient(ctx context.Context) error {
	ctx, end := log.SpanContext(ctx, "initDBClient")
	defer end()
	db, err := dbutil.NewDB(
		dbutil.WithHostPort(env.config.PGBouncerHost, env.config.PGBouncerPort),
		dbutil.WithDBName(env.config.PostgresDBName),
		dbutil.WithUserPassword(env.config.PostgresUser, env.config.PostgresPassword),
		dbutil.WithMaxOpenConns(env.config.PGBouncerMaxOpenConns),
		dbutil.WithMaxIdleConns(env.config.PGBouncerMaxIdleConns),
		dbutil.WithConnMaxLifetime(time.Duration(env.config.PostgresConnMaxLifetimeSeconds)*time.Second),
		dbutil.WithConnMaxIdleTime(time.Duration(env.config.PostgresConnMaxIdleSeconds)*time.Second),
		dbutil.WithSSLMode(dbutil.SSLModeDisable),
		dbutil.WithQueryLog(env.config.PostgresQueryLogging, "pgx.bouncer"),
	)
	if err != nil {
		return err
	}
	env.dbClient = db
	if err := prometheus.Register(sqlstats.NewStatsCollector("pg_bouncer", env.dbClient.DB)); err != nil {
		log.Info(ctx, "problem registering stats collector for pg_bouncer db client", zap.Error(err))
	}
	return nil
}

func (env *NonblockingServiceEnv) newListener() col.PostgresListener {
	// TODO: Change this to be based on whether a direct connection to postgres is available.
	// A direct connection will not be available in the workers when the PG bouncer changes are in.
	if env.isWorker() {
		return env.newProxyListener()
	}
	return env.newDirectListener()
}

func (env *NonblockingServiceEnv) newProxyListener() col.PostgresListener {
	// The proxy postgres listener is lazily initialized to avoid consuming too many
	// gRPC resources by having idle client connections, so construction
	// can't fail.
	return client.NewProxyPostgresListener(env.newProxyClient)
}

func (env *NonblockingServiceEnv) newProxyClient() (proxy.APIClient, error) {
	var servicePachClient *client.APIClient
	if err := backoff.Retry(func() error {
		var err error
		servicePachClient, err = client.NewInCluster(pctx.TODO())
		return err
	}, backoff.RetryEvery(time.Second).For(5*time.Minute)); err != nil {
		return nil, errors.Wrapf(err, "failed to initialize service pach client")
	}
	return servicePachClient.ProxyClient, nil
}

func (env *NonblockingServiceEnv) newDirectListener() col.PostgresListener {
	dsn := dbutil.GetDSN(
		dbutil.WithHostPort(env.config.PostgresHost, env.config.PostgresPort),
		dbutil.WithDBName(env.config.PostgresDBName),
		dbutil.WithUserPassword(env.config.PostgresUser, env.config.PostgresPassword),
		dbutil.WithSSLMode(env.config.PostgresSSL),
		// Note: enabling query logs with WithQueryLog silentely causes Pachyderm to fail,
		// so don't do it.  nabling logs generates a synthetic DSN to pass to sql.DB (this
		// is how PGX maps its own non-string config to sql.DB connection strings), but it
		// only works for PGX connections, not PQ connections.  For the listener, we use the
		// PQ library instead of PGX, so the synthetic DSNs are not interchangeable.
		// Avoiding WithQueryLogs suppresses generating a synthetic DSN.
	)
	// The postgres listener is lazily initialized to avoid consuming too many
	// postgres resources by having idle client connections, so construction
	// can't fail.
	return col.NewPostgresListener(dsn)
}

// GetPachClient returns a pachd client with the same authentication
// credentials and cancellation as 'ctx' (ensuring that auth credentials are
// propagated through downstream RPCs).
//
// Functions that receive RPCs should call this to convert their RPC context to
// a Pachyderm client, and internal Pachyderm calls should accept clients
// returned by this call.
//
// (Warning) Do not call this function during server setup unless it is in a goroutine.
// A Pachyderm client is not available until the server has been setup.
func (env *NonblockingServiceEnv) GetPachClient(ctx context.Context) *client.APIClient {
	if err := env.pachEg.Wait(); err != nil {
		panic(err) // If env can't connect, there's no sensible way to recover
	}
	return env.pachClient.WithCtx(ctx)
}

// GetEtcdClient returns the already connected etcd client without modification.
func (env *NonblockingServiceEnv) GetEtcdClient() *etcd.Client {
	if err := env.etcdEg.Wait(); err != nil {
		panic(err) // If env can't connect, there's no sensible way to recover
	}
	if env.etcdClient == nil {
		panic("service env never connected to etcd")
	}
	return env.etcdClient
}

func (env *NonblockingServiceEnv) GetTaskService(prefix string) task.Service {
	return task.NewEtcdService(env.GetEtcdClient(), prefix)
}

// GetKubeClient returns the already connected Kubernetes API client without
// modification.
func (env *NonblockingServiceEnv) GetKubeClient() kube.Interface {
	if err := env.kubeEg.Wait(); err != nil {
		panic(err) // If env can't connect, there's no sensible way to recover
	}
	if env.kubeClient == nil {
		panic("service env never connected to kubernetes")
	}
	return env.kubeClient
}

// GetLokiClient returns the loki client, it doesn't require blocking on a
// connection because the client is just a dumb struct with no init function.
func (env *NonblockingServiceEnv) GetLokiClient() (*loki.Client, error) {
	if env.lokiClient == nil {
		return nil, errors.Errorf("loki not configured, is it running in the same namespace as pachd?")
	}
	return env.lokiClient, nil
}

// GetDBClient returns the already connected database client without modification.
func (env *NonblockingServiceEnv) GetDBClient() *pachsql.DB {
	if err := env.dbEg.Wait(); err != nil {
		panic(err) // If env can't connect, there's no sensible way to recover
	}
	if env.dbClient == nil {
		panic("service env never connected to the database")
	}
	return env.dbClient
}

func (env *NonblockingServiceEnv) GetDirectDBClient() *pachsql.DB {
	if env.isWorker() {
		panic("worker cannot get direct db client")
	}
	if err := env.dbEg.Wait(); err != nil {
		panic(err)
	}
	if env.directDBClient == nil {
		panic("service env never connected to the database")
	}
	return env.directDBClient
}

// GetPostgresListener returns the already constructed database client dedicated
// for listen operations without modification. Note that this listener lazily
// connects to the database on the first listen operation.
func (env *NonblockingServiceEnv) GetPostgresListener() col.PostgresListener {
	if env.listener == nil {
		panic("service env never created the listener")
	}
	return env.listener
}

func (env *NonblockingServiceEnv) GetDexDB() dex_storage.Storage {
	if err := env.dexDBEg.Wait(); err != nil {
		panic(err) // If env can't connect, there's no sensible way to recover
	}
	if env.dexDB == nil {
		panic("service env never connected to the Dex database")
	}
	return env.dexDB
}

func (env *NonblockingServiceEnv) ClusterID() string {
	if err := env.clusterIdEg.Wait(); err != nil {
		panic(err)
	}

	return env.clusterId
}

func (env *NonblockingServiceEnv) Context() context.Context {
	return env.ctx
}

func (env *NonblockingServiceEnv) Close() error {
	// Cancel anything using the ServiceEnv's context
	env.cancel()

	// Close all of the clients and return the first error.
	// Loki client and kube client do not have a Close method.
	eg := &errgroup.Group{}
	eg.Go(env.GetPachClient(context.Background()).Close)
	eg.Go(env.GetEtcdClient().Close)
	eg.Go(env.GetDBClient().Close)
	eg.Go(env.GetPostgresListener().Close)
	return errors.EnsureStack(eg.Wait())
}

// AuthServer returns the registered Auth APIServer
func (env *NonblockingServiceEnv) AuthServer() auth_server.APIServer {
	return env.authServer
}

// SetAuthServer registers an Auth APIServer with this service env
func (env *NonblockingServiceEnv) SetAuthServer(s auth_server.APIServer) {
	env.authServer = s
}

// IdentityServer returns the registered Identity APIServer
func (env *NonblockingServiceEnv) IdentityServer() identity.APIServer {
	return env.identityServer
}

// SetIdentityServer registers an Identity APIServer with this service env
func (env *NonblockingServiceEnv) SetIdentityServer(s identity.APIServer) {
	env.identityServer = s
}

// PpsServer returns the registered PPS APIServer
func (env *NonblockingServiceEnv) PpsServer() pps_server.APIServer {
	return env.ppsServer
}

// SetPpsServer registers a Pps APIServer with this service env
func (env *NonblockingServiceEnv) SetPpsServer(s pps_server.APIServer) {
	env.ppsServer = s
}

// PfsServer returns the registered PFS APIServer
func (env *NonblockingServiceEnv) PfsServer() pfs_server.APIServer {
	return env.pfsServer
}

// SetPfsServer registers a Pfs APIServer with this service env
func (env *NonblockingServiceEnv) SetPfsServer(s pfs_server.APIServer) {
	env.pfsServer = s
}

// EnterpriseServer returns the registered PFS APIServer
func (env *NonblockingServiceEnv) EnterpriseServer() enterprise_server.APIServer {
	return env.enterpriseServer
}

// SetEnterpriseServer registers a Enterprise APIServer with this service env
func (env *NonblockingServiceEnv) SetEnterpriseServer(s enterprise_server.APIServer) {
	env.enterpriseServer = s
}

// SetKubeClient can be used to override the kubeclient in testing.
func (env *NonblockingServiceEnv) SetKubeClient(s kube.Interface) {
	env.kubeClient = s
}
