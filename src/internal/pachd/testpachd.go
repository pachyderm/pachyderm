// Package pachd needs to be documented.
//
// TODO: document
package pachd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-units"
	etcdcli "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	etcdwal "go.etcd.io/etcd/server/v3/wal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/v2/src/internal/cleanup"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	lokiclient "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

// TestPachdOption is used to customize a testpachd.
type TestPachdOption struct {
	noLogToFile      bool // A flag to turn off the LogToFileOption when it's the default.
	MutateContext    func(ctx context.Context) context.Context
	MutateEnv        func(env *Env)
	MutateConfig     func(config *pachconfig.PachdFullConfiguration)
	MutatePachd      func(full *Full)
	MutateFullOption func(fullOption *FullOption)
	OnReady          func(ctx context.Context, full *Full) error
	Cleanup          func(ctx context.Context) error
	CopyDBConfig     *dockertestenv.DBConfig // will copy the database configuration
}

// NoLogToFileOption is an option that disables NewTestPachd's default behavior of logging pachd
// logs to a file.
func NoLogToFileOption() TestPachdOption {
	return TestPachdOption{noLogToFile: true}
}

// ActivateAuthOption is an option that activates auth inside the created pachd.  Outside of tests,
// you must manually call pachd.AwaitAuth(ctx).
func ActivateAuthOption(rootToken string) TestPachdOption {
	if rootToken == "" {
		rootToken = "iamroot"
	}
	return TestPachdOption{
		MutateConfig: func(config *pachconfig.PachdFullConfiguration) {
			config.ActivateAuth = true
			config.AuthRootToken = rootToken
			config.LicenseKey = os.Getenv("ENT_ACT_CODE")
			config.EnterpriseSecret = "enterprisey"
		},
	}
}

func PJSWorkerAuthOption(pjsWorkerAuthToken string) TestPachdOption {
	if pjsWorkerAuthToken == "" {
		pjsWorkerAuthToken = "iamroot"
	}
	return TestPachdOption{
		MutateConfig: func(config *pachconfig.PachdFullConfiguration) {
			config.PJSWorkerAuthToken = pjsWorkerAuthToken
		},
	}
}

// GetK8sClient is an option that sets k to the internal fake k8s client.  Mutations made to this
// client will be visible to Pachyderm.
func GetK8sClient(k *kubernetes.Interface) TestPachdOption {
	return TestPachdOption{
		MutatePachd: func(full *Full) {
			*k = full.kubeClient
		},
	}
}

// WithS3Server runs the S3 gateway and sets the provided string pointer to its address.
func WithS3Server(t *testing.T, addr *string) TestPachdOption {
	l := testutil.Listen(t)
	var server *http.Server
	return TestPachdOption{
		OnReady: func(ctx context.Context, full *Full) error {
			full.s3Server.Addr = l.Addr().String()
			*addr = l.Addr().String()
			server = full.s3Server.Server
			go full.s3Server.Serve(l) //nolint:errcheck
			log.Info(ctx, "waiting for s3 server to be ready", zap.String("addr", *addr))
			for i := 0; i < 10; i++ {
				req, err := http.NewRequestWithContext(ctx, "get", fmt.Sprintf("http://%v/", *addr), nil)
				if err != nil {
					return errors.Wrapf(err, "wait for s3: attempt %d", i)
				}
				if _, err := http.DefaultClient.Do(req); err == nil {
					break
				}
				log.Debug(ctx, "s3 server not ready yet", zap.Error(err))
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		},
		Cleanup: func(ctx context.Context) error {
			if server != nil {
				if err := server.Close(); err != nil {
					return errors.Wrap(err, "close s3 server")
				}
			}
			return nil
		},
	}
}

// NewTestPachd creates an environment suitable for non-k8s tests
// and then calls pachd.NewFull with that environment.
func NewTestPachd(t testing.TB, opts ...TestPachdOption) *client.APIClient {
	ctx := pctx.TestContext(t)
	logToFile := true
	for _, o := range opts {
		if o.noLogToFile {
			logToFile = false
			break
		}
	}
	if logToFile {
		cfg := zap.NewProductionConfig()
		cfg.Sampling = nil
		cfg.OutputPaths = []string{filepath.Join(os.TempDir(), fmt.Sprintf("pachyderm-real-env-%s.log", url.PathEscape(t.Name())))}
		cfg.Level.SetLevel(zapcore.DebugLevel)
		logger, err := cfg.Build()
		require.NoError(t, err, "should be able to build a logger")
		ctx = pctx.Child(ctx, "", pctx.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return logger.Core()
		})))
	}
	ctx = mutateContext(ctx, opts...)

	dbcfg := dockertestenv.NewTestDBConfig(t)
	db := testutil.OpenDB(t, dbcfg.PGBouncer.DBOptions()...)
	directDB := testutil.OpenDB(t, dbcfg.Direct.DBOptions()...)
	dbListenerConfig := dbutil.GetDSN(
		ctx,
		dbutil.WithHostPort(dbcfg.Direct.Host, int(dbcfg.Direct.Port)),
		dbutil.WithDBName(dbcfg.Direct.DBName),
		dbutil.WithUserPassword(dbcfg.Direct.User, dbcfg.Direct.Password),
		dbutil.WithSSLMode("disable"))

	lis := testutil.Listen(t)
	bucket, _ := dockertestenv.NewTestBucket(ctx, t)
	etcd := testetcd.NewEnv(ctx, t).EtcdClient

	env := Env{
		DB:               db,
		DirectDB:         directDB,
		DBListenerConfig: dbListenerConfig,
		Bucket:           bucket,
		EtcdClient:       etcd,
		Listener:         lis,
	}
	pd := newTestPachd(env, opts)
	go func() {
		if err := pd.Run(ctx); err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				t.Log(err)
			}
		}
	}()
	pd.AwaitAuth(ctx)
	for _, opt := range opts {
		if opt.CopyDBConfig != nil {
			*opt.CopyDBConfig = dbcfg
		}
		if opt.OnReady != nil {
			if err := opt.OnReady(ctx, pd); err != nil {
				t.Fatalf("OnReady callback for option: %v", err)
			}
		}
		if opt.Cleanup != nil {
			t.Cleanup(func() {
				if err := opt.Cleanup(ctx); err != nil {
					t.Errorf("cleanup: %v", err)
				}
			})
		}
	}

	// client setup
	pachClient, err := pd.PachClient(ctx)
	require.NoError(t, err)
	require.NoErrorWithinTRetry(t, 5*time.Second, func() error { return pachClient.Health() })
	return pachClient
}

func mutateContext(ctx context.Context, opts ...TestPachdOption) context.Context {
	for _, o := range opts {
		if o.MutateContext != nil {
			ctx = o.MutateContext(ctx)
		}
	}
	return ctx
}

func newTestPachd(env Env, opts []TestPachdOption) *Full {
	config := pachconfig.PachdFullConfiguration{
		GlobalConfiguration: pachconfig.GlobalConfiguration{
			PeerPort:  netip.MustParseAddrPort(env.Listener.Addr().String()).Port(),
			Namespace: "default",
		},
		PachdSpecificConfiguration: pachconfig.PachdSpecificConfiguration{
			StorageConfiguration: pachconfig.StorageConfiguration{
				StorageMemoryThreshold:    units.GB,
				StorageLevelFactor:        10,
				StorageCompactionMaxFanIn: 10,
				StorageMemoryCacheSize:    20,
			},
		},
	}
	env.GetLokiClient = func() (*lokiclient.Client, error) {
		return nil, errors.New("no loki")
	}
	fullOption := &FullOption{}
	for _, opt := range opts {
		if opt.MutateEnv != nil {
			opt.MutateEnv(&env)
		}
		if opt.MutateConfig != nil {
			opt.MutateConfig(&config)
		}
		if opt.MutateFullOption != nil {
			opt.MutateFullOption(fullOption)
		}
	}
	pd := NewFull(env, config, fullOption)
	for _, opt := range opts {
		if opt.MutatePachd != nil {
			opt.MutatePachd(pd)
		}
	}
	return pd
}

// BuildTestPachd returns a test pachd that can be run outside of tests.  The returned cleanup
// handler frees all ephemeral resources associated with the instance.
func BuildTestPachd(ctx context.Context, opts ...TestPachdOption) (*Full, *cleanup.Cleaner, error) {
	cleaner := new(cleanup.Cleaner)

	// setup context
	ctx = mutateContext(ctx, opts...)

	// tmpdir
	tmpdir, err := os.MkdirTemp("", "testpachd-")
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "new tmpdir")
	}
	cleaner.AddCleanup("tmpdir", func() error {
		return errors.Wrapf(os.RemoveAll(tmpdir), "RemoveAll(%v)", tmpdir)
	})

	// database
	dbcfg, closeDB, err := dockertestenv.NewTestDBConfigCtx(ctx)
	cleaner.Subsume(closeDB)
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "test db config")
	}
	db, err := dbutil.NewDB(ctx, dbcfg.PGBouncer.DBOptions()...)
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "open pgbouncer connection")
	}
	cleaner.AddCleanup("pgbouncer connection", db.Close)
	directDB, err := dbutil.NewDB(ctx, dbcfg.Direct.DBOptions()...)
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "open direct db connection")
	}
	cleaner.AddCleanup("direct db connection", directDB.Close)
	dbListenerConfig := dbutil.GetDSN(
		ctx,
		dbutil.WithHostPort(dbcfg.Direct.Host, int(dbcfg.Direct.Port)),
		dbutil.WithDBName(dbcfg.Direct.DBName),
		dbutil.WithUserPassword(dbcfg.Direct.User, dbcfg.Direct.Password),
		dbutil.WithSSLMode("disable"),
	)

	// minio
	bucket, _, cleanupMinio, err := dockertestenv.NewTestBucketCtx(ctx)
	cleaner.AddCleanupCtx("minio", cleanupMinio)
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "create test bucket")
	}

	// etcd
	etcdwal.SegmentSizeBytes = 1 * 1000 * 1000 // 1 MB
	etcdConfig := embed.NewConfig()
	etcdConfig.Dir = path.Join(tmpdir, "etcd_data")
	etcdConfig.WalDir = path.Join(tmpdir, "etcd_wal")
	etcdConfig.UnsafeNoFsync = true
	etcdConfig.MaxTxnOps = 10000
	etcdConfig.InitialElectionTickAdvance = false
	etcdConfig.TickMs = 10
	etcdConfig.ElectionMs = 50
	level := log.AddLoggerToEtcdServer(ctx, etcdConfig)
	etcdLis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "create etcd listener")
	}
	if err := etcdLis.Close(); err != nil {
		return nil, cleaner, errors.Wrap(err, "close etcd listener")
	}
	clientURL, err := url.Parse(fmt.Sprintf("http://%s", etcdLis.Addr().String()))
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "parse etcd client URL")
	}
	etcdConfig.ListenPeerUrls = []url.URL{}
	etcdConfig.ListenClientUrls = []url.URL{*clientURL}
	level.SetLevel(zapcore.ErrorLevel)
	etcd, err := embed.StartEtcd(etcdConfig)
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "start etcd")
	}
	cleaner.AddCleanup("etcd", func() error { etcd.Close(); return nil })
	select {
	case <-etcd.Server.ReadyNotify():
		level.SetLevel(zapcore.InfoLevel)
	case <-ctx.Done():
		return nil, cleaner, errors.Wrap(context.Cause(ctx), "wait for etcd startup")
	}
	cleaner.AddCleanup("shut up etcd", func() error {
		level.SetLevel(zapcore.ErrorLevel)
		return nil
	})
	cfg := log.GetEtcdClientConfig(ctx)
	cfg.Endpoints = []string{clientURL.String()}
	cfg.DialOptions = client.DefaultDialOptions()
	etcdClient, err := etcdcli.New(cfg)
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "get etcd client")
	}

	// listener
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "listen on 127.0.0.1:0")
	}
	cleaner.AddCleanup("pachd listener", func() error {
		err := lis.Close()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				return errors.Wrap(err, "close listener")
			}
		}
		return nil
	})

	// build pachd
	env := Env{
		DB:               db,
		DirectDB:         directDB,
		DBListenerConfig: dbListenerConfig,
		Bucket:           bucket,
		EtcdClient:       etcdClient,
		Listener:         lis,
	}
	pd := newTestPachd(env, opts)
	return pd, cleaner, nil
}
