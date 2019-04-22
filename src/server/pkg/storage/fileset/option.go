package fileset

type Option func(f *FileSet)

func WithRoot(root string) Option {
	return func(f *FileSet) {
		op.root = root
	}
}

func WithParent(parentName string) Option {
	return func(f *FileSet) {
		f.parentName = parentName
	}
}

func WithScratch(memThreshold int64, scratch Scratch) Option {
	return func(f *FileSet) {
		f.memThreshold = memThreshold
		f.scratch = scratch
	}
}
