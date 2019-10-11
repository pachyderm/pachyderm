package fileset

// Option configures a file set.
type Option func(f *FileSet)

// WithRoot sets the root path of the file set.
func WithRoot(root string) Option {
	return func(f *FileSet) {
		f.root = root
	}
}

// WithMemThreshold sets the memory threshold of the file set.
func WithMemThreshold(threshold int64) Option {
	return func(f *FileSet) {
		f.memAvailable = threshold
		f.memThreshold = threshold
	}
}
