// Package transactiondb contains the database schema that Pachyderm
// transactions use.
package transactiondb

import (
	"path"

	etcd "go.etcd.io/etcd/clientv3"

	"github.com/pachyderm/pachyderm/src/client/transaction"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

const (
	transactionsPrefix = "/transactions"
)

// Transactions returns a collection of open transactions
func Transactions(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, transactionsPrefix),
		nil,
		&transaction.TransactionInfo{},
		nil,
		nil,
	)
}
