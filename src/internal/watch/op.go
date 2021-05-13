package watch

import etcd "github.com/coreos/etcd/clientv3"

// WatchOptions is a set of options that can be used when watching
type WatchOptions struct {
	SortTarget    etcd.SortTarget
	SortOrder     etcd.SortOrder
	IncludePut    bool
	IncludeDelete bool
}

func DefaultWatchOptions() WatchOptions {
	// Sort by mod revision--how the items would have been returned if we watched
	// them from the beginning.
	return WatchOptions{
		SortTarget:    etcd.SortByModRevision,
		SortOrder:     etcd.SortAscend,
		IncludePut:    true,
		IncludeDelete: true,
	}
}

func SumOptions(opts ...Option) WatchOptions {
	result := DefaultWatchOptions()
	for _, opt := range opts {
		result = opt(result)
	}
	return result
}

// Option is a function that modifies the set of options to use when watching
type Option func(WatchOptions) WatchOptions

// IgnorePut discards PUT events from the watcher
func IgnorePut(opt WatchOptions) WatchOptions {
	opt.IncludePut = false
	return opt
}

// IgnoreDelete discards DELETE events from the watcher
func IgnoreDelete(opt WatchOptions) WatchOptions {
	opt.IncludeDelete = false
	return opt
}

// WithSort specifies the sort to use for the watcher
func WithSort(sortTarget etcd.SortTarget, sortOrder etcd.SortOrder) Option {
	return func(opt WatchOptions) WatchOptions {
		opt.SortTarget = sortTarget
		opt.SortOrder = sortOrder
		return opt
	}
}
