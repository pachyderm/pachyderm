package watch

import etcd "github.com/coreos/etcd/clientv3"

// OpOption is a simple typedef for etcd.OpOption.
type OpOption etcd.OpOption

// WithFilterPut discards PUT events from the watcher.
func WithFilterPut() OpOption {
	return OpOption(etcd.WithFilterPut())
}
