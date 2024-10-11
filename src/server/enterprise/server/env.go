package server

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	kube "k8s.io/client-go/kubernetes"

	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/server/auth"
)

type Env struct {
	DB       *pachsql.DB
	Listener col.PostgresListener
	TxnEnv   *txnenv.TransactionEnv

	EtcdClient *clientv3.Client
	EtcdPrefix string

	AuthServer    auth.APIServer
	GetPachClient func(context.Context) *client.APIClient
	GetKubeClient func() kube.Interface
	Namespace     string

	BackgroundContext context.Context
	Config            pachconfig.Configuration
}

// PauseMode represents whether a server is unpaused, paused, a sidecar or an enterprise server.
type PauseMode uint8

const (
	UnpausableMode PauseMode = iota
	FullMode
	PausedMode
)

type Config struct {
	Heartbeat    bool
	Mode         PauseMode
	UnpausedMode string
}

func EnterpriseConfigCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		"enterpriseConfig",
		db,
		listener,
		&ec.EnterpriseConfig{},
		nil,
	)
}

// EnterpriseConfigPostgresMigration is not properly documented.
//
// The enterpriseConfig collection stores the information necessary for the enterprise-service to
// heartbeat to the license service for ongoing license validity checks. For clusters with enterprise,
// if this information were lost, the cluster would eventually become locked out. We migrate
// This data is migrated to postgres so that the data stored in etcd can truly be considered ephemeral.
//
// TODO: document.
func EnterpriseConfigPostgresMigration(ctx context.Context, tx *pachsql.Tx, etcd *clientv3.Client) error {
	if err := col.SetupPostgresCollections(ctx, tx, EnterpriseConfigCollection(nil, nil)); err != nil {
		return err
	}
	config, err := checkForEtcdRecord(ctx, etcd)
	if err != nil {
		return err
	}
	if config != nil {
		return errors.EnsureStack(EnterpriseConfigCollection(nil, nil).ReadWrite(tx).Put(ctx, configKey, config))
	}
	return nil
}

func checkForEtcdRecord(ctx context.Context, etcd *clientv3.Client) (*ec.EnterpriseConfig, error) {
	etcdConfigCol := col.NewEtcdCollection(etcd, "", nil, &ec.EnterpriseConfig{}, nil, nil)
	var config ec.EnterpriseConfig
	if err := etcdConfigCol.ReadOnly().Get(ctx, configKey, &config); err != nil {
		if col.IsErrNotFound(err) {
			return nil, nil
		}
		return nil, errors.EnsureStack(err)
	}
	return &config, nil
}

func DeleteEnterpriseConfigFromEtcd(ctx context.Context, etcd *clientv3.Client) error {
	if _, err := col.NewSTM(ctx, etcd, func(stm col.STM) error {
		etcdConfigCol := col.NewEtcdCollection(etcd, "", nil, &ec.EnterpriseConfig{}, nil, nil)
		return errors.EnsureStack(etcdConfigCol.ReadWrite(stm).Delete(ctx, configKey))
	}); err != nil {
		if !col.IsErrNotFound(err) {
			return err
		}
	}
	return nil
}

// StopWorkers stops all workers
func StopWorkers(ctx context.Context, kc kube.Interface, namespace string) error {
	return scaleDownWorkers(ctx, kc, namespace)
}
