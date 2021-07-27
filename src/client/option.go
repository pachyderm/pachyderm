package client

import "github.com/pachyderm/pachyderm/v2/src/pfs"

type putFileConfig struct {
	tag    string
	append bool
}

// PutFileOption configures a PutFile call.
type PutFileOption func(*putFileConfig)

// WithAppendPutFile configures the PutFile call to append to existing files.
func WithAppendPutFile() PutFileOption {
	return func(pf *putFileConfig) {
		pf.append = true
	}
}

// WithTagPutFile configures the PutFile call to apply to a particular tag.
func WithTagPutFile(tag string) PutFileOption {
	return func(pf *putFileConfig) {
		pf.tag = tag
	}
}

type deleteFileConfig struct {
	tag       string
	recursive bool
}

// DeleteFileOption configures a DeleteFile call.
type DeleteFileOption func(*deleteFileConfig)

// WithTagDeleteFile configures the DeleteFile call to apply to a particular tag.
func WithTagDeleteFile(tag string) DeleteFileOption {
	return func(dfc *deleteFileConfig) {
		dfc.tag = tag
	}
}

// WithRecursiveDeleteFile configures the DeleteFile call to recursively delete the files in a directory.
func WithRecursiveDeleteFile() DeleteFileOption {
	return func(dfc *deleteFileConfig) {
		dfc.recursive = true
	}
}

// CopyFileOption configures a CopyFile call.
type CopyFileOption func(*pfs.CopyFile)

// WithAppendCopyFile configures the CopyFile call to append to existing files.
func WithAppendCopyFile() CopyFileOption {
	return func(cf *pfs.CopyFile) {
		cf.Append = true
	}
}

// WithTagCopyFile configures the CopyFile call to apply to a particular tag.
func WithTagCopyFile(tag string) CopyFileOption {
	return func(cf *pfs.CopyFile) {
		cf.Tag = tag
	}
}
