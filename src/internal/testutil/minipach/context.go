package minipach

import (
	"fmt"
	"path"
	"runtime/debug"
	"testing"

	adminclient "github.com/pachyderm/pachyderm/v2/src/admin"
	authclient "github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	debugclient "github.com/pachyderm/pachyderm/v2/src/debug"
	eprsclient "github.com/pachyderm/pachyderm/v2/src/enterprise"
	healthclient "github.com/pachyderm/pachyderm/v2/src/health"
	identityclient "github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	licenseclient "github.com/pachyderm/pachyderm/v2/src/license"
	pfsclient "github.com/pachyderm/pachyderm/v2/src/pfs"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
	adminserver "github.com/pachyderm/pachyderm/v2/src/server/admin/server"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	debugserver "github.com/pachyderm/pachyderm/v2/src/server/debug/server"
	eprsserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	"github.com/pachyderm/pachyderm/v2/src/server/health"
	identity_server "github.com/pachyderm/pachyderm/v2/src/server/identity/server"
	licenseserver "github.com/pachyderm/pachyderm/v2/src/server/license/server"
	pfs_server "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	pps_server "github.com/pachyderm/pachyderm/v2/src/server/pps/server"
	txnserver "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
	transactionclient "github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/random"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
)

type TestContext struct {
	env *serviceenv.TestServiceEnv
}

func GetTestContext(t testing.TB) *TestContext {
	t.Helper()
	env := testetcd.NewEnv(t)
	testId := random.String(20)
	clientSocketPath := path.Join(env.Directory, "pachd_socket")
	fmt.Printf("Test context %s - %s\n", testId, env.Directory)

	fullConfig := &serviceenv.PachdFullConfiguration{}
	cmdutil.Populate(fullConfig)
	config := serviceenv.NewConfiguration(fullConfig)
	config.PostgresServiceSSL = "disable"
	config.StorageRoot = path.Join(env.Directory, "pach_root")
	config.CacheRoot = path.Join(env.Directory, "cache_root")

	db := dbutil.NewTestDB(t)
	require.NoError(t, migrations.ApplyMigrations(context.Background(), db, migrations.Env{}, clusterstate.DesiredClusterState))

	senv := &serviceenv.TestServiceEnv{
		Configuration: config,
		EtcdClient:    env.EtcdClient,
		DBClient:      db,
	}
	require.NoError(t, SetupServer(senv, clientSocketPath))

	var err error
	senv.PachClient, err = client.NewFromSocket("unix://" + clientSocketPath)
	require.NoError(t, err)
	return &TestContext{env: senv}
}

func (c *TestContext) GetAuthenticatedPachClient(tb testing.TB, subject string) *client.APIClient {
	tb.Helper()
	rootClient := c.GetUnauthenticatedPachClient(tb)
	ActivateAuth(tb, rootClient)
	rootClient.SetAuthToken(RootToken)
	if subject == authclient.RootUser {
		return rootClient
	}
	token, err := rootClient.GetRobotToken(rootClient.Ctx(), &authclient.GetRobotTokenRequest{Robot: subject})
	require.NoError(tb, err)
	client := c.GetUnauthenticatedPachClient(tb)
	client.SetAuthToken(token.Token)
	return client
}

// GetUnauthenticatedPachClient returns a copy of the testing pach client with no auth token
func (c *TestContext) GetUnauthenticatedPachClient(tb testing.TB) *client.APIClient {
	tb.Helper()
	return c.env.PachClient.WithCtx(context.Background())
}

func SetupServer(env serviceenv.ServiceEnv, socketPath string) error {
	debug.SetGCPercent(env.Config().GCPercent)

	kubeNamespace := ""
	var reporter *metrics.Reporter

	authInterceptor := auth.NewInterceptor(env)
	externalServer, err := grpcutil.NewServer(
		context.Background(),
		true,
		grpc.ChainUnaryInterceptor(
			tracing.UnaryServerInterceptor(),
			authInterceptor.InterceptUnary,
		),
		grpc.ChainStreamInterceptor(
			tracing.StreamServerInterceptor(),
			authInterceptor.InterceptStream,
		),
	)
	if err != nil {
		return err
	}

	identityStorageProvider, err := identity_server.NewStorageProvider(env)
	if err != nil {
		return err
	}

	if err := logGRPCServerSetup("External Pachd", func() error {
		txnEnv := &txnenv.TransactionEnv{}
		var pfsAPIServer pfs_server.APIServer
		if err := logGRPCServerSetup("PFS API", func() error {
			pfsAPIServer, err = pfs_server.NewAPIServer(env, txnEnv, path.Join(env.Config().EtcdPrefix, env.Config().PFSEtcdPrefix), env.GetDBClient())
			if err != nil {
				return err
			}
			pfsclient.RegisterAPIServer(externalServer.Server, pfsAPIServer)
			return nil
		}); err != nil {
			return err
		}
		var ppsAPIServer pps_server.APIServer
		if err := logGRPCServerSetup("PPS API", func() error {
			ppsAPIServer, err = pps_server.NewAPIServer(
				env,
				txnEnv,
				path.Join(env.Config().EtcdPrefix, env.Config().PPSEtcdPrefix),
				kubeNamespace,
				env.Config().WorkerImage,
				env.Config().WorkerSidecarImage,
				env.Config().WorkerImagePullPolicy,
				env.Config().StorageRoot,
				env.Config().StorageBackend,
				env.Config().StorageHostPath,
				env.Config().CacheRoot,
				env.Config().IAMRole,
				env.Config().ImagePullSecret,
				env.Config().NoExposeDockerSocket,
				reporter,
				env.Config().WorkerUsesRoot,
				env.Config().PPSWorkerPort,
				env.Config().Port,
				env.Config().HTTPPort,
				env.Config().PeerPort,
				env.Config().GCPercent,
			)
			if err != nil {
				return err
			}
			ppsclient.RegisterAPIServer(externalServer.Server, ppsAPIServer)
			return nil
		}); err != nil {
			return err
		}

		if err := logGRPCServerSetup("Identity API", func() error {
			idAPIServer := identity_server.NewIdentityServer(
				env,
				identityStorageProvider,
				false,
			)
			if err != nil {
				return err
			}
			identityclient.RegisterAPIServer(externalServer.Server, idAPIServer)
			return nil
		}); err != nil {
			return err
		}

		var authAPIServer authserver.APIServer
		if err := logGRPCServerSetup("Auth API", func() error {
			authAPIServer, err = authserver.NewAuthServer(
				env, txnEnv, path.Join(env.Config().EtcdPrefix, env.Config().AuthEtcdPrefix), false, false, true)
			if err != nil {
				return err
			}
			authclient.RegisterAPIServer(externalServer.Server, authAPIServer)
			return nil
		}); err != nil {
			return err
		}
		var transactionAPIServer txnserver.APIServer
		if err := logGRPCServerSetup("Transaction API", func() error {
			transactionAPIServer, err = txnserver.NewAPIServer(
				env,
				txnEnv,
				path.Join(env.Config().EtcdPrefix, env.Config().PFSEtcdPrefix),
			)
			if err != nil {
				return err
			}
			transactionclient.RegisterAPIServer(externalServer.Server, transactionAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Enterprise API", func() error {
			enterpriseAPIServer, err := eprsserver.NewEnterpriseServer(
				env, path.Join(env.Config().EtcdPrefix, env.Config().EnterpriseEtcdPrefix))
			if err != nil {
				return err
			}
			eprsclient.RegisterAPIServer(externalServer.Server, enterpriseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("License API", func() error {
			licenseAPIServer, err := licenseserver.New(
				env, path.Join(env.Config().EtcdPrefix, env.Config().EnterpriseEtcdPrefix))
			if err != nil {
				return err
			}
			licenseclient.RegisterAPIServer(externalServer.Server, licenseAPIServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Admin API", func() error {
			adminclient.RegisterAPIServer(externalServer.Server, adminserver.NewAPIServer(&adminclient.ClusterInfo{
				ID:           "",
				DeploymentID: env.Config().DeploymentID,
			}))
			return nil
		}); err != nil {
			return err
		}
		healthServer := health.NewHealthServer()
		if err := logGRPCServerSetup("Health", func() error {
			healthclient.RegisterHealthServer(externalServer.Server, healthServer)
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Version API", func() error {
			versionpb.RegisterAPIServer(externalServer.Server, version.NewAPIServer(version.Version, version.APIServerOptions{}))
			return nil
		}); err != nil {
			return err
		}
		if err := logGRPCServerSetup("Debug", func() error {
			debugclient.RegisterDebugServer(externalServer.Server, debugserver.NewDebugServer(
				env,
				env.Config().PachdPodName,
				nil,
			))
			return nil
		}); err != nil {
			return err
		}
		txnEnv.Initialize(env, transactionAPIServer, authAPIServer, pfsAPIServer, ppsAPIServer)
		if err := externalServer.ListenSocket(socketPath); err != nil {
			return err
		}
		healthServer.Ready()
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func logGRPCServerSetup(name string, f func() error) (retErr error) {
	log.Printf("started setting up %v GRPC Server", name)
	defer func() {
		if retErr != nil {
			retErr = errors.Wrapf(retErr, "error setting up %v GRPC Server", name)
		} else {
			log.Printf("finished setting up %v GRPC Server", name)
		}
	}()
	return f()
}
