package pachd

import (
	"testing"

	"github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewTestPachd creates an environment suitable for non-k8s tests
// and then calls pachd.NewFull with that environment.
func NewTestPachd(t testing.TB) *client.APIClient {
	ctx := pctx.TestContext(t)
	cfg := zap.NewProductionConfig()
	cfg.Sampling = nil
	// cfg.OutputPaths = []string{filepath.Join(os.TempDir(), fmt.Sprintf("pachyderm-real-env-%s.log", url.PathEscape(t.Name())))}
	cfg.Level.SetLevel(zapcore.DebugLevel)
	logger, err := cfg.Build()
	require.NoError(t, err, "should be able to make a realenv logger")
	ctx = pctx.Child(ctx, "", pctx.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return logger.Core()
	})))

	dbcfg := dockertestenv.NewTestDBConfig(t)
	objC := dockertestenv.NewTestObjClient(ctx, t)
	lis := testutil.Listen(t)
	env := Env{
		DB:         testutil.OpenDB(t, dbcfg.PGBouncer.DBOptions()...),
		DirectDB:   testutil.OpenDB(t, dbcfg.Direct.DBOptions()...), // TODO
		ObjClient:  objC,
		EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient,
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

	// create pachd and run it
	pd := NewFull(env, config)
	go func() {
		if err := pd.Run(ctx); err != nil {
			t.Log(err)
		}
	}()

	// client setup
	addr, err := grpcutil.ParsePachdAddress("http://" + lis.Addr().String())
	require.NoError(t, err)
	pachClient, err := client.NewFromPachdAddress(ctx, addr)
	require.NoError(t, err)

	return pachClient
}
