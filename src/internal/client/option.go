package client

import "github.com/pachyderm/pachyderm/v2/src/pfs"

type putFileConfig struct {
	datum             string
	append            bool
	importConcurrency uint32
}

// PutFileOption configures a PutFile call.
type PutFileOption func(*putFileConfig)

// WithAppendPutFile configures the PutFile call to append to existing files.
func WithAppendPutFile() PutFileOption {
	return func(pf *putFileConfig) {
		pf.append = true
	}
}

// WithDatumPutFile configures the PutFile call to apply to a particular datum.
func WithDatumPutFile(datum string) PutFileOption {
	return func(pf *putFileConfig) {
		pf.datum = datum
	}
}

// WithImportConcurrency configures the maximum number of tasks in flight created by PutFileURL.
func WithImportConcurrency(importConcurrency uint32) PutFileOption {
	return func(pf *putFileConfig) {
		pf.importConcurrency = importConcurrency
	}
}

type deleteFileConfig struct {
	datum     string
	recursive bool
}

// DeleteFileOption configures a DeleteFile call.
type DeleteFileOption func(*deleteFileConfig)

// WithDatumDeleteFile configures the DeleteFile call to apply to a particular datum.
func WithDatumDeleteFile(datum string) DeleteFileOption {
	return func(dfc *deleteFileConfig) {
		dfc.datum = datum
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

// WithDatumCopyFile configures the CopyFile call to apply to a particular datum.
func WithDatumCopyFile(datum string) CopyFileOption {
	return func(cf *pfs.CopyFile) {
		cf.Datum = datum
	}
}

// GetFileOption configures a GetFile call
type GetFileOption func(*pfs.GetFileRequest)

// WithDatumGetFile sets the datum for the get file request
func WithDatumGetFile(datum string) GetFileOption {
	return func(gf *pfs.GetFileRequest) {
		gf.File.Datum = datum
	}
}

func WithOffset(offset int64) GetFileOption {
	return func(gf *pfs.GetFileRequest) {
		gf.Offset = offset
	}
}
