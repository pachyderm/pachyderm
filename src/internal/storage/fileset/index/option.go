package index

// Option configures an index reader.
type Option func(r *Reader)

// PathRange is a range of paths.
// The range is inclusive, exclusive: [Lower, Upper).
type PathRange struct {
	Lower, Upper string
}

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

func (r *PathRange) atStart(path string) bool {
	if r.Lower == "" {
		return true
	}
	return path >= r.Lower
}

func (r *PathRange) atEnd(path string) bool {
	if r.Upper == "" {
		return false
	}
	return path >= r.Upper
}
