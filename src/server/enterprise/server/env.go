package server

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	kube "k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/v2/src/client"
	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/server/auth"
	logrus "github.com/sirupsen/logrus"
)

type Env struct {
	env          serviceenv.ServiceEnv
	txnEnv       *txnenv.TransactionEnv
	etcdPrefix   string
	mode         PauseMode
	unpausedMode string
}

// PauseMode represents whether a server is unpaused, paused, a sidecar or an enterprise server.
type PauseMode uint8

const (
	UnpausableMode PauseMode = iota
	FullMode
	PausedMode
)

type Option func(Env) Env

func WithUnpausedMode(mode string) Option {
	return func(e Env) Env {
		e.unpausedMode = mode
		return e
	}
}

func WithMode(mode PauseMode) Option {
	return func(e Env) Env {
		e.mode = mode
		return e
	}
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv, etcdPrefix string, txEnv *txnenv.TransactionEnv, options ...Option) *Env {
	e := Env{
		env:        senv,
		txnEnv:     txEnv,
		etcdPrefix: etcdPrefix,
	}
	for _, o := range options {
		e = o(e)
	}
	return &e
}

// Delegations to the service environment.  These are explicit in order to make
// it clear which parts of the service environment are relied upon.
func (e Env) DB() *pachsql.DB                                     { return e.env.GetDBClient() }
func (e Env) Listener() col.PostgresListener                      { return e.env.GetPostgresListener() }
func (e Env) EtcdClient() *clientv3.Client                        { return e.env.GetEtcdClient() }
func (e Env) AuthServer() auth.APIServer                          { return e.env.AuthServer() }
func (e Env) GetPachClient(ctx context.Context) *client.APIClient { return e.env.GetPachClient(ctx) }
func (e Env) getKubeClient() *kube.Clientset                      { return e.env.GetKubeClient() }
func (e Env) BackgroundContext() context.Context                  { return e.env.Context() }
func (e Env) namespace() string                                   { return e.env.Config().Namespace }
func (e Env) Logger() *logrus.Logger                              { return e.env.Logger() }
func (e Env) Config() serviceenv.Configuration                    { return *e.env.Config() }

func EnterpriseConfigCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		"enterpriseConfig",
		db,
		listener,
		&ec.EnterpriseConfig{},
		nil,
	)
}

// The enterpriseConfig collection stores the information necessary for the enterprise-service to
// heartbeat to the license service for ongoing license validity checks. For clusters with enterprise,
// if this information were lost, the cluster would eventually become locked out. We migrate
// This data is migrated to postgres so that the data stored in etcd can truly be considered ephemeral.
func EnterpriseConfigPostgresMigration(ctx context.Context, tx *pachsql.Tx, etcd *clientv3.Client) error {
	if err := col.SetupPostgresCollections(ctx, tx, EnterpriseConfigCollection(nil, nil)); err != nil {
		return err
	}
	config, err := checkForEtcdRecord(ctx, etcd)
	if err != nil {
		return err
	}
	if config != nil {
		return errors.EnsureStack(EnterpriseConfigCollection(nil, nil).ReadWrite(tx).Put(configKey, config))
	}
	return nil
}

func checkForEtcdRecord(ctx context.Context, etcd *clientv3.Client) (*ec.EnterpriseConfig, error) {
	etcdConfigCol := col.NewEtcdCollection(etcd, "", nil, &ec.EnterpriseConfig{}, nil, nil)
	var config ec.EnterpriseConfig
	if err := etcdConfigCol.ReadOnly(ctx).Get(configKey, &config); err != nil {
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
		return errors.EnsureStack(etcdConfigCol.ReadWrite(stm).Delete(configKey))
	}); err != nil {
		if !col.IsErrNotFound(err) {
			return err
		}
	}
	return nil
}
