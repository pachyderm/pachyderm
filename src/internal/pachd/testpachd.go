package pachd

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/cleanup"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	clientv3 "go.etcd.io/etcd/client/v3"
	etcdcli "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	etcdwal "go.etcd.io/etcd/server/v3/wal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gocloud.dev/blob"
)

// NewTestPachd creates an environment suitable for non-k8s tests
// and then calls pachd.NewFull with that environment.
func NewTestPachd(t testing.TB) *client.APIClient {
	ctx := pctx.TestContext(t)
	cfg := zap.NewProductionConfig()
	cfg.Sampling = nil
	cfg.OutputPaths = []string{filepath.Join(os.TempDir(), fmt.Sprintf("pachyderm-real-env-%s.log", url.PathEscape(t.Name())))}
	cfg.Level.SetLevel(zapcore.DebugLevel)
	logger, err := cfg.Build()
	require.NoError(t, err, "should be able to make a realenv logger")
	ctx = pctx.Child(ctx, "", pctx.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return logger.Core()
	})))

	dbcfg := dockertestenv.NewTestDBConfig(t)
	db := testutil.OpenDB(t, dbcfg.PGBouncer.DBOptions()...)
	directDB := testutil.OpenDB(t, dbcfg.Direct.DBOptions()...)

	lis := testutil.Listen(t)
	bucket, _ := dockertestenv.NewTestBucket(ctx, t)
	etcd := testetcd.NewEnv(ctx, t).EtcdClient

	pd := newTestPachd(db, directDB, bucket, etcd, lis)
	go func() {
		if err := pd.Run(ctx); err != nil {
			t.Log(err)
		}
	}()

	// client setup
	pachClient, err := pd.PachClient(ctx)
	require.NoError(t, err)
	require.NoErrorWithinTRetry(t, 5*time.Second, func() error { return pachClient.Health() })
	return pachClient
}

func newTestPachd(db, directDB *pachsql.DB, bucket *blob.Bucket, etcd *clientv3.Client, lis net.Listener) *Full {
	env := Env{
		DB:         db,
		DirectDB:   directDB,
		Bucket:     bucket,
		EtcdClient: etcd,
		Listener:   lis,
	}
	config := pachconfig.PachdFullConfiguration{
		PachdSpecificConfiguration: pachconfig.PachdSpecificConfiguration{
			StorageConfiguration: pachconfig.StorageConfiguration{
				StorageMemoryThreshold:    units.GB,
				StorageLevelFactor:        10,
				StorageCompactionMaxFanIn: 10,
				StorageMemoryCacheSize:    20,
			},
		},
	}
	pd := NewFull(env, config)
	return pd
}

func BuildTestPachd(ctx context.Context) (*Full, *cleanup.Cleaner, error) {
	cleaner := new(cleanup.Cleaner)

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
	db, err := dbutil.NewDB(dbcfg.PGBouncer.DBOptions()...)
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "open pgbouncer connection")
	}
	cleaner.AddCleanup("pgbouncer connection", db.Close)
	directDB, err := dbutil.NewDB(dbcfg.Direct.DBOptions()...)
	if err != nil {
		return nil, cleaner, errors.Wrap(err, "open direct db connection")
	}
	cleaner.AddCleanup("direct db connection", directDB.Close)

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
		return nil, cleaner, errors.Wrap(err, "wait for etcd startup")
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
	pd := newTestPachd(db, directDB, bucket, etcdClient, lis)
	return pd, cleaner, nil
}
