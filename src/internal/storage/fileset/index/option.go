package index

// Option configures an index reader.
type Option func(r *Reader)

// WithRange sets a range filter for the read.
func WithRange(pathRange *PathRange) Option {
	return func(r *Reader) {
		r.filter = &pathFilter{pathRange: pathRange}
	}
}

// WithPrefix sets a prefix filter for the read.
func WithPrefix(prefix string) Option {
	return func(r *Reader) {
		r.filter = &pathFilter{prefix: prefix}
	}
}

// WithDatum adds a datum filter that matches a single datum.
func WithDatum(datum string) Option {
	return func(r *Reader) {
		r.datum = datum
	}
}

// WithShardConfig sets the sharding configuration.
func WithShardConfig(config *ShardConfig) Option {
	return func(r *Reader) {
		r.shardConfig = config
	}
}
