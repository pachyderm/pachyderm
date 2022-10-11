package realenv

import (
	"net"
	"net/url"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"testing"

	units "github.com/docker/go-units"
	testclient "k8s.io/client-go/kubernetes/fake"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	authapi "github.com/pachyderm/pachyderm/v2/src/server/auth"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	"github.com/pachyderm/pachyderm/v2/src/server/enterprise"
	enterpriseserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	identityserver "github.com/pachyderm/pachyderm/v2/src/server/identity/server"
	licenseserver "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	pfsapi "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	ppsapi "github.com/pachyderm/pachyderm/v2/src/server/pps"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	proxyserver "github.com/pachyderm/pachyderm/v2/src/server/proxy/server"
	txnserver "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
	"github.com/pachyderm/pachyderm/v2/src/version"
	pb "github.com/pachyderm/pachyderm/v2/src/version/versionpb"
)

// RealEnv contains a setup for running end-to-end pachyderm tests locally.  It
// includes the base MockEnv struct as well as a real instance of the API server.
// These calls can still be mocked, but they default to calling into the real
// server endpoints.
type RealEnv struct {
	testpachd.MockEnv

	ServiceEnv               serviceenv.ServiceEnv
	AuthServer               authapi.APIServer
	IdentityServer           identity.APIServer
	EnterpriseServer         enterprise.APIServer
	LicenseServer            license.APIServer
	PPSServer                ppsapi.APIServer
	PFSServer                pfsapi.APIServer
	TransactionServer        txnserver.APIServer
	VersionServer            pb.APIServer
	ProxyServer              proxy.APIServer
	MockPPSTransactionServer *testpachd.MockPPSTransactionServer
}

// NewRealEnv constructs a MockEnv, then forwards all API calls to go to API
// server instances for supported operations. PPS uses a fake clientset which allows
// some PPS behavior to work.
func NewRealEnv(t testing.TB, customOpts ...serviceenv.ConfigOption) *RealEnv {
	return newRealEnv(t, false, testpachd.AuthMiddlewareInterceptor, customOpts...)
}

// NewRealEnvWithIdentity creates a new real Env and then activates an identity server with dex.
func NewRealEnvWithIdentity(t testing.TB, customOpts ...serviceenv.ConfigOption) *RealEnv {
	env := newRealEnv(t, false, testpachd.AuthMiddlewareInterceptor, customOpts...)
	env.ActivateIdentity(t)
	return env
}

// NewRealEnvWithPPSTransactionMock constructs a MockEnv, then forwards all API calls to go to API
// server instances for supported operations. A mock implementation of PPS Transactions are used.
func NewRealEnvWithPPSTransactionMock(t testing.TB, customOpts ...serviceenv.ConfigOption) *RealEnv {
	noInterceptor := func(mock *testpachd.MockPachd) grpcutil.Interceptor {
		return grpcutil.Interceptor{}
	}
	return newRealEnv(t, true, noInterceptor, customOpts...)
}

func (realEnv *RealEnv) ActivateIdentity(t testing.TB) {
	dir, err := os.Getwd()
	require.NoError(t, err)
	dexDir := strings.Split(dir, "src")[0] + "dex-assets"
	_, err = os.Stat(dexDir)
	require.NoError(t, err)
	realEnv.ServiceEnv.InitDexDB()
	identityEnv := identityserver.EnvFromServiceEnv(realEnv.ServiceEnv)
	identityAddr := realEnv.PachClient.GetAddress().Host + ":" + strconv.Itoa(int(realEnv.PachClient.GetAddress().Port+8))
	realEnv.IdentityServer = identityserver.NewIdentityServer(identityEnv, true, identityserver.UnitTestOption(identityAddr, dexDir))
	realEnv.ServiceEnv.SetIdentityServer(realEnv.IdentityServer)
	linkServers(&realEnv.MockPachd.Identity, realEnv.IdentityServer)
}

func newRealEnv(t testing.TB, mockPPSTransactionServer bool, interceptor testpachd.InterceptorOption, customOpts ...serviceenv.ConfigOption) *RealEnv {
	mockEnv := testpachd.NewMockEnv(t, interceptor)

	realEnv := &RealEnv{MockEnv: *mockEnv}
	etcdClientURL, err := url.Parse(realEnv.EtcdClient.Endpoints()[0])
	require.NoError(t, err)

	opts := []serviceenv.ConfigOption{
		func(config *serviceenv.Configuration) {
			require.NoError(t, cmdutil.PopulateDefaults(config))
			config.StorageBackend = obj.Local
			config.StorageRoot = path.Join(realEnv.Directory, "localStorage")
		},
		DefaultConfigOptions,
		serviceenv.WithEtcdHostPort(etcdClientURL.Hostname(), etcdClientURL.Port()),
		serviceenv.WithPachdPeerPort(uint16(realEnv.MockPachd.Addr.(*net.TCPAddr).Port)),
		serviceenv.WithOidcPort(uint16(realEnv.MockPachd.Addr.(*net.TCPAddr).Port + 7)),
	}
	opts = append(opts, customOpts...) // Overwrite with any custom options
	realEnv.ServiceEnv = serviceenv.InitServiceEnv(serviceenv.ConfigFromOptions(opts...))

	// Overwrite the mock pach client with the ServiceEnv's client so it gets closed earlier
	realEnv.PachClient = realEnv.ServiceEnv.GetPachClient(realEnv.ServiceEnv.Context())

	t.Cleanup(func() {
		// There is a race condition here, although not too serious because this only
		// happens in tests and the errors should not propagate back to the clients -
		// ideally we would close the client connection first and wait for the server
		// RPCs to end before closing the underlying service connections (like
		// postgres and etcd), so we don't get spurious errors. Instead, some RPCs may
		// fail because of losing the database connection.
		// TODO: It appears the postgres db.Close() may return errors due to
		// background goroutines using a closed TCP session because we don't do an
		// orderly shutdown, so we don't check the error here.
		realEnv.ServiceEnv.Close()
	})

	// database migrations
	err = migrations.ApplyMigrations(realEnv.ServiceEnv.Context(), realEnv.ServiceEnv.GetDBClient(), migrations.Env{EtcdClient: realEnv.EtcdClient}, clusterstate.DesiredClusterState)
	require.NoError(t, err)
	err = migrations.BlockUntil(realEnv.ServiceEnv.Context(), realEnv.ServiceEnv.GetDBClient(), clusterstate.DesiredClusterState)
	require.NoError(t, err)

	txnEnv := txnenv.New()
	// AUTH
	authEnv := authserver.EnvFromServiceEnv(realEnv.ServiceEnv, txnEnv)
	realEnv.AuthServer, err = authserver.NewAuthServer(authEnv, true, false, true)
	require.NoError(t, err)
	realEnv.ServiceEnv.SetAuthServer(realEnv.AuthServer)

	// ENTERPRISE
	entEnv := enterpriseserver.EnvFromServiceEnv(realEnv.ServiceEnv, path.Join("", "enterprise"), txnEnv)
	realEnv.EnterpriseServer, err = enterpriseserver.NewEnterpriseServer(entEnv, true)
	require.NoError(t, err)
	realEnv.ServiceEnv.SetEnterpriseServer(realEnv.EnterpriseServer)
	mockEnv.MockPachd.GetAuthServer = realEnv.ServiceEnv.AuthServer

	// LICENSE
	licenseEnv := licenseserver.EnvFromServiceEnv(realEnv.ServiceEnv)
	realEnv.LicenseServer, err = licenseserver.New(licenseEnv)
	require.NoError(t, err)

	// PFS
	pfsEnv, err := pfsserver.EnvFromServiceEnv(realEnv.ServiceEnv, txnEnv)
	require.NoError(t, err)
	pfsEnv.EtcdPrefix = ""
	realEnv.PFSServer, err = pfsserver.NewAPIServer(*pfsEnv)
	require.NoError(t, err)
	realEnv.ServiceEnv.SetPfsServer(realEnv.PFSServer)

	// TRANSACTION
	realEnv.TransactionServer, err = txnserver.NewAPIServer(realEnv.ServiceEnv, txnEnv)
	require.NoError(t, err)
	realEnv.ProxyServer = proxyserver.NewAPIServer(proxyserver.Env{Listener: realEnv.ServiceEnv.GetPostgresListener()})

	// VERSION
	realEnv.VersionServer = version.NewAPIServer(version.Version, version.APIServerOptions{})

	txnEnv.Initialize(realEnv.ServiceEnv, realEnv.TransactionServer)

	// PPS
	if mockPPSTransactionServer {
		realEnv.MockPPSTransactionServer = testpachd.NewMockPPSTransactionServer()
		realEnv.ServiceEnv.SetPpsServer(&realEnv.MockPPSTransactionServer.Api)
		realEnv.MockPPSTransactionServer.InspectPipelineInTransaction.
			Use(func(txnctx *txncontext.TransactionContext, pipeline *pps.Pipeline) (*pps.PipelineInfo, error) {
				return nil, col.ErrNotFound{
					Type: "pipelines",
					Key:  pipeline.String(),
				}
			})
	} else {
		reporter := metrics.NewReporter(realEnv.ServiceEnv)
		clientset := testclient.NewSimpleClientset()
		realEnv.ServiceEnv.SetKubeClient(clientset)
		ppsEnv := ppsserver.EnvFromServiceEnv(realEnv.ServiceEnv, txnEnv, reporter)
		realEnv.PPSServer, err = ppsserver.NewAPIServer(ppsEnv)
		realEnv.ServiceEnv.SetPpsServer(realEnv.PPSServer)
		require.NoError(t, err)
		linkServers(&realEnv.MockPachd.PPS, realEnv.PPSServer)
	}

	linkServers(&realEnv.MockPachd.PFS, realEnv.PFSServer)
	linkServers(&realEnv.MockPachd.Auth, realEnv.AuthServer)
	linkServers(&realEnv.MockPachd.Enterprise, realEnv.EnterpriseServer)
	linkServers(&realEnv.MockPachd.License, realEnv.LicenseServer)
	linkServers(&realEnv.MockPachd.Transaction, realEnv.TransactionServer)
	linkServers(&realEnv.MockPachd.Version, realEnv.VersionServer)
	linkServers(&realEnv.MockPachd.Proxy, realEnv.ProxyServer)

	return realEnv
}

// DefaultConfigOptions is a serviceenv config option with the defaults used for tests
func DefaultConfigOptions(config *serviceenv.Configuration) {
	config.StorageMemoryThreshold = units.GB
	config.StorageLevelFactor = 10
	config.StorageCompactionMaxFanIn = 10
	config.StorageMemoryCacheSize = 20
}

// linkServers can be used to default a mock server to make calls to a real api
// server. Due to some reflection shenanigans, mockServerPtr must explicitly be
// a pointer to the mock server instance.
func linkServers(mockServerPtr interface{}, realServer interface{}) {
	mockValue := reflect.ValueOf(mockServerPtr).Elem()
	realValue := reflect.ValueOf(realServer)
	mockType := mockValue.Type()
	for i := 0; i < mockType.NumField(); i++ {
		field := mockType.Field(i)
		if field.Name != "api" {
			mock := mockValue.FieldByName(field.Name)
			realMethod := realValue.MethodByName(field.Name)

			// We need a pointer to the mock field to call the right method
			mockPtr := reflect.New(reflect.PtrTo(mock.Type()))
			mockPtrValue := mockPtr.Elem()
			mockPtrValue.Set(mock.Addr())

			useFn := mockPtrValue.MethodByName("Use")
			useFn.Call([]reflect.Value{realMethod})
		}
	}
}
