// Package transactiondb contains the database schema that Pachyderm
// transactions use.
package transactiondb

import (
	"path"

	etcd "github.com/coreos/etcd/clientv3"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
)

const (
	transactionsPrefix = "/transactions"
)

// Transactions returns a collection of open transactions
func Transactions(etcdClient *etcd.Client, etcdPrefix string) col.EtcdCollection {
	return col.NewEtcdCollection(
		etcdClient,
		path.Join(etcdPrefix, transactionsPrefix),
		nil,
		&transaction.TransactionInfo{},
		nil,
		nil,
	)
}
