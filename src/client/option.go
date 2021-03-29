package client

import "github.com/pachyderm/pachyderm/v2/src/pfs"

// PutFileOption configures a PutFile call.
type PutFileOption func(*pfs.PutFile)

// WithAppendPutFile configures the PutFile call to append to existing files.
func WithAppendPutFile() PutFileOption {
	return func(pf *pfs.PutFile) {
		pf.Append = true
	}
}

// WithTagPutFile configures the PutFile call to apply to a particular tag.
func WithTagPutFile(tag string) PutFileOption {
	return func(pf *pfs.PutFile) {
		pf.Tag = tag
	}
}

// DeleteFileOption configures a DeleteFile call.
type DeleteFileOption func(*pfs.DeleteFile)

// WithTagDeleteFile configures the DeleteFile call to apply to a particular tag.
func WithTagDeleteFile(tag string) DeleteFileOption {
	return func(df *pfs.DeleteFile) {
		df.Tag = tag
	}
}

// CopyFileOption configures a CopyFile call.
type CopyFileOption func(*pfs.CopyFileRequest)

// WithAppendCopyFile configures the CopyFile call to append to existing files.
func WithAppendCopyFile() CopyFileOption {
	return func(cfr *pfs.CopyFileRequest) {
		cfr.Append = true
	}
}

// WithTagCopyFile configures the CopyFile call to apply to a particular tag.
func WithTagCopyFile(tag string) CopyFileOption {
	return func(cfr *pfs.CopyFileRequest) {
		cfr.Tag = tag
	}
}
