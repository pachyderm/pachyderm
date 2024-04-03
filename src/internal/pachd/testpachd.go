package pachd

import (
	"context"
	stderrors "errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-units"
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
	etcdcli "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	etcdwal "go.etcd.io/etcd/server/v3/wal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

// TestPachdOptions allow a testpachd to be customized.
type TestPachdOption struct {
	noLogToFile  bool // A flag to turn off the LogToFileOption when it's the default.
	MutateEnv    func(env *Env)
	MutateConfig func(config *pachconfig.PachdFullConfiguration)
	MutatePachd  func(full *Full)
}

// NoLogToFileOption is an option that disable's NewTestPachd's default behavior of logging pachd
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

	dbcfg := dockertestenv.NewTestDBConfig(t)
	db := testutil.OpenDB(t, dbcfg.PGBouncer.DBOptions()...)
	directDB := testutil.OpenDB(t, dbcfg.Direct.DBOptions()...)
	dbListenerConfig := dbutil.GetDSN(
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

	// client setup
	pachClient, err := pd.PachClient(ctx)
	require.NoError(t, err)
	require.NoErrorWithinTRetry(t, 5*time.Second, func() error { return pachClient.Health() })
	return pachClient
}

func newTestPachd(env Env, opts []TestPachdOption) *Full {
	config := pachconfig.PachdFullConfiguration{
		GlobalConfiguration: pachconfig.GlobalConfiguration{
			PeerPort: netip.MustParseAddrPort(env.Listener.Addr().String()).Port(),
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
	for _, opt := range opts {
		if opt.MutateEnv != nil {
			opt.MutateEnv(&env)
		}
		if opt.MutateConfig != nil {
			opt.MutateConfig(&config)
		}
	}
	pd := NewFull(env, config)
	for _, opt := range opts {
		if opt.MutatePachd != nil {
			opt.MutatePachd(pd)
		}
	}
	return pd
}

// BuildTestPachd returns a test pachd that can be run outside of tests.  The returned cleanup
// handler frees all ephemeral resources associated with the instance.
func BuildAndRunTestPachd(ctx context.Context, eg *errgroup.Group, opts ...TestPachdOption) (*Full, error) {
	var pdCh = make(chan *Full)

	eg.Go(func() (errs error) {
		// tmpdir
		tmpdir, err := os.MkdirTemp("", "testpachd-")
		if err != nil {
			return errors.Wrap(err, "new tmpdir")
		}
		defer func() {
			errs = stderrors.Join(errs, errors.Wrapf(os.RemoveAll(tmpdir), "RemoveAll(%v)", tmpdir))
		}()

		// database
		dbcfg, err := dockertestenv.NewTestDBConfigCtx(ctx)
		defer func() {
			errs = stderrors.Join(errs, dbcfg.Cleanup(pctx.Background("cleanup DB config")))
		}()
		if err != nil {
			return errors.Wrap(err, "test db config")
		}
		db, err := dbutil.NewDB(dbcfg.PGBouncer.DBOptions()...)
		if err != nil {
			return errors.Wrap(err, "open pgbouncer connection")
		}
		defer func() {
			errs = stderrors.Join(errs, db.Close())
		}()
		directDB, err := dbutil.NewDB(dbcfg.Direct.DBOptions()...)
		if err != nil {
			return errors.Wrap(err, "open direct db connection")
		}
		defer func() {
			errs = stderrors.Join(errs, directDB.Close())
		}()
		dbListenerConfig := dbutil.GetDSN(
			dbutil.WithHostPort(dbcfg.Direct.Host, int(dbcfg.Direct.Port)),
			dbutil.WithDBName(dbcfg.Direct.DBName),
			dbutil.WithUserPassword(dbcfg.Direct.User, dbcfg.Direct.Password),
			dbutil.WithSSLMode("disable"),
		)

		// minio
		bucket, _, cleanupMinio, err := dockertestenv.NewTestBucketCtx(ctx)
		defer func() {
			errs = stderrors.Join(errs, cleanupMinio(pctx.Background("cleanup minio")))
		}()
		if err != nil {
			return errors.Wrap(err, "create test bucket")
		}

		// etcd
		etcdwal.SegmentSizeBytes = 1 * 1000 * 1000 // 1 MB
		etcdConfig := embed.NewConfig()
		etcdConfig.Dir = path.Join(tmpdir, "etcd_data")
		etcdConfig.WalDir = path.Join(tmpdir, "etcd_wal")
		etcdConfig.MaxTxnOps = 10000
		etcdConfig.InitialElectionTickAdvance = false
		etcdConfig.TickMs = 10
		etcdConfig.ElectionMs = 50
		level := log.AddLoggerToEtcdServer(ctx, etcdConfig)
		etcdLis, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			return errors.Wrap(err, "create etcd listener")
		}
		if err := etcdLis.Close(); err != nil {
			return errors.Wrap(err, "close etcd listener")
		}
		clientURL, err := url.Parse(fmt.Sprintf("http://%s", etcdLis.Addr().String()))
		if err != nil {
			return errors.Wrap(err, "parse etcd client URL")
		}
		etcdConfig.ListenPeerUrls = []url.URL{}
		etcdConfig.ListenClientUrls = []url.URL{*clientURL}
		level.SetLevel(zapcore.ErrorLevel)
		etcd, err := embed.StartEtcd(etcdConfig)
		if err != nil {
			return errors.Wrap(err, "start etcd")
		}
		defer etcd.Close()
		select {
		case <-etcd.Server.ReadyNotify():
			level.SetLevel(zapcore.InfoLevel)
		case <-ctx.Done():
			return errors.Wrap(err, "wait for etcd startup")
		}
		defer level.SetLevel(zapcore.ErrorLevel)
		cfg := log.GetEtcdClientConfig(ctx)
		cfg.Endpoints = []string{clientURL.String()}
		cfg.DialOptions = client.DefaultDialOptions()
		etcdClient, err := etcdcli.New(cfg)
		if err != nil {
			return errors.Wrap(err, "get etcd client")
		}

		// listener
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return errors.Wrap(err, "listen on 127.0.0.1:0")
		}
		defer func() {
			if err := lis.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
				errs = stderrors.Join(errs, errors.Wrap(err, "close listener"))
			}
		}()

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
		pdCh <- pd
		if err := pd.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return errors.Wrap(err, "pd.Run")
		}
		return nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err() //nolint:wrapcheck
	case pd := <-pdCh:
		return pd, nil
	}
}
