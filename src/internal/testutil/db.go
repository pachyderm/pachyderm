package testutil

import (
	"crypto/rand"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/deploy/assets"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// set this to false if you want to keep the database around
var cleanup = true

const postgresMaxConnections = 100

// we want to divide the total number of connections we can have up among the
// concurrently running tests
var maxOpenConnsPerPool = postgresMaxConnections / runtime.GOMAXPROCS(0)

// TestDatabaseDeployment represents a deployment of postgres, and databases may
// be created for individual tests.
type TestDatabaseDeployment interface {
	NewDatabase(t testing.TB) (*sqlx.DB, *col.PostgresListener)
}

type postgresDeployment struct {
	db      *sqlx.DB
	connect func(testing.TB, string) (*sqlx.DB, *col.PostgresListener)
}

// NewPostgresDeployment creates a kubernetes namespaces containing a
// postgres instance, lazily generated up to the specified pool size. The pool
// will be destroyed at the end of the test or test suite that called this.
func NewPostgresDeployment(t testing.TB) TestDatabaseDeployment {
	manifest := &strings.Builder{}
	encoder, err := serde.GetEncoder("json", manifest, serde.WithIndent(2), serde.WithOrigName(true))
	require.NoError(t, err)

	namespaceName := UniqueString("postgres-test-")
	assetOpts := &assets.AssetOpts{
		PostgresOpts: assets.PostgresOpts{
			Nodes:      1,
			MemRequest: "500M",
			CPURequest: "0.1",
		},
		Namespace: namespaceName,
	}

	configMap := assets.PostgresInitConfigMap(assetOpts)
	require.NoError(t, encoder.Encode(configMap))

	storageClass, err := assets.PostgresStorageClass(assetOpts, assets.LocalBackend)
	require.NoError(t, err)
	require.NoError(t, encoder.Encode(storageClass))

	headlessService := assets.PostgresHeadlessService(assetOpts)
	require.NoError(t, encoder.Encode(headlessService))

	statefulSet := assets.PostgresStatefulSet(assetOpts, assets.LocalBackend, 1)
	require.NoError(t, encoder.Encode(statefulSet))

	service := assets.PostgresService(assetOpts)
	service.Spec.Ports[0].NodePort = 0
	require.NoError(t, encoder.Encode(service))

	// Connect to kubernetes and create the postgres items in their namespace
	kubeClient := GetKubeClient(t)
	namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}
	_, err = kubeClient.CoreV1().Namespaces().Create(namespace)
	require.NoError(t, err)

	if cleanup {
		t.Cleanup(func() {
			err := kubeClient.CoreV1().Namespaces().Delete(namespaceName, nil)
			require.NoError(t, err)
		})
	}

	cmd := Cmd("kubectl", "apply", "--namespace", namespaceName, "-f", "-")
	cmd.Stdin = strings.NewReader(manifest.String())
	err = cmd.Run()
	require.NoError(t, err)

	// Since persistent volumes are not tied to namespaces, use a label to delete
	// the volumes we created on shutdown.
	if cleanup {
		t.Cleanup(func() {
			selector := fmt.Sprintf("namespace=%s", namespaceName)
			err := kubeClient.CoreV1().PersistentVolumes().DeleteCollection(nil, metav1.ListOptions{LabelSelector: selector})
			require.NoError(t, err)
		})
	}

	// Read out the service details from kubernetes
	postgres, err := kubeClient.CoreV1().Services(namespaceName).Get("postgres", metav1.GetOptions{})
	require.NoError(t, err)

	var port int
	for _, servicePort := range postgres.Spec.Ports {
		if servicePort.Name == "client-port" {
			port = int(servicePort.NodePort)
		}
	}
	require.NotEqual(t, 0, port)

	// Get the IP address of the nodes (any _should_ work for the service port)
	nodes, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	require.NoError(t, err)

	// Minikube 'Hostname' address type didn't work when testing, use InternalIP
	var address string
	for _, addr := range nodes.Items[0].Status.Addresses {
		if addr.Type == "InternalIP" {
			address = addr.Address
		}
	}
	require.NotEqual(t, "", address)

	connect := func(t testing.TB, databaseName string) (*sqlx.DB, *col.PostgresListener) {
		options := []dbutil.Option{
			dbutil.WithHostPort(address, port),
			dbutil.WithDBName(databaseName),
		}

		db, err := dbutil.NewDB(options...)
		require.NoError(t, err)

		// Check the connection
		_, err = db.Exec("SELECT 1")
		require.NoError(t, err)
		db.SetMaxOpenConns(maxOpenConnsPerPool)

		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})

		listener := col.NewPostgresListener(dbutil.GetDSN(options...))
		t.Cleanup(func() {
			require.NoError(t, listener.Close())
		})

		return db, listener
	}

	// We don't actually need the listener at this level, just close it
	// immediately to preserve resources (it should be safe to close multiple
	// times).
	db, listener := connect(t, "")
	require.NoError(t, listener.Close())

	return &postgresDeployment{connect: connect, db: db}
}

func (pd *postgresDeployment) NewDatabase(t testing.TB) (*sqlx.DB, *col.PostgresListener) {
	dbName := ephemeralDBName(t)
	_, err := pd.db.Exec("CREATE DATABASE " + dbName)
	require.NoError(t, err)

	if cleanup {
		t.Cleanup(func() {
			_, err := pd.db.Exec("DROP DATABASE " + dbName)
			require.NoError(t, err)
		})
	}

	return pd.connect(t, dbName)
}

// NewTestDB connects to postgres using the default settings, creates a database with a unique name
// then calls cb with a sqlx.DB configured to use the newly created database.
// After cb returns the database is dropped.
func NewTestDB(t testing.TB) *sqlx.DB {
	dbName := ephemeralDBName(t)
	require.NoError(t, withDB(func(db *sqlx.DB) error {
		db.MustExec("CREATE DATABASE " + dbName)
		t.Log("database", dbName, "successfully created")
		return nil
	}))
	if cleanup {
		t.Cleanup(func() {
			require.NoError(t, withDB(func(db *sqlx.DB) error {
				db.MustExec("DROP DATABASE " + dbName)
				t.Log("database", dbName, "successfully deleted")
				return nil
			}))
		})
	}
	db2, err := dbutil.NewDB(dbutil.WithDBName(dbName))
	require.NoError(t, err)
	db2.SetMaxOpenConns(maxOpenConnsPerPool)
	t.Cleanup(func() {
		require.NoError(t, db2.Close())
	})
	return db2
}

// withDB creates a database connection that is scoped to the passed in callback.
func withDB(cb func(*sqlx.DB) error, opts ...dbutil.Option) (retErr error) {
	db, err := dbutil.NewDB(opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(db)
}

func ephemeralDBName(t testing.TB) string {
	buf := [8]byte{}
	n, err := rand.Reader.Read(buf[:])
	require.NoError(t, err)
	require.Equal(t, n, 8)

	// TODO: it looks like postgres is truncating identifiers to 32 bytes,
	// it should be 64 but we might be passing the name as non-ascii, i'm not really sure.
	// for now just use a random int, but it would be nice to go back to names with a timestamp.
	return fmt.Sprintf("test_%08x", buf)
	//now := time.Now()
	// test_<date>T<time>_<random int>
	// return fmt.Sprintf("test_%04d%02d%02dT%02d%02d%02d_%04x",
	// 	now.Year(), now.Month(), now.Day(),
	// 	now.Hour(), now.Minute(), now.Second(),
	// 	rand.Uint32())
}
