package watch

import etcd "github.com/coreos/etcd/clientv3"

// OpOption is a simple typedef for etcd.OpOption.
type OpOption struct {
	Get   etcd.OpOption
	Watch etcd.OpOption
}

// WithFilterPut discards PUT events from the watcher.
func WithFilterPut() OpOption {
	return OpOption{Watch: etcd.WithFilterPut(), Get: nil}
}

// WithSort specifies the sort to use for the watcher
func WithSort(sortBy etcd.SortTarget, sortOrder etcd.SortOrder) OpOption {
	return OpOption{Get: etcd.WithSort(sortBy, sortOrder), Watch: nil}
}

// WithFilterDelete discards DELETE events from the watcher.
func WithFilterDelete() OpOption {
	return OpOption{Watch: etcd.WithFilterDelete()}
}
