package watch

import etcd "github.com/coreos/etcd/clientv3"

// OpOption is a simple typedef for etcd.OpOption.
type OpOption etcd.OpOption

// WithPrevKV gets the previous key-value pair before the event happens. If the previous KV is already compacted,
// nothing will be returned.
func WithPrevKV() OpOption {
	return OpOption(etcd.WithPrevKV())
}

// WithFilterPut discards PUT events from the watcher.
func WithFilterPut() OpOption {
	return OpOption(etcd.WithFilterPut())
}
