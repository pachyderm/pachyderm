package server

import (
	"bytes"
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

type FileAccessor struct {
	ctx       context.Context
	apiServer pfs.APIServer
	d         *driver
}

func (a *apiServer) NewFileAccessor(ctx context.Context) txncontext.FileAccessor {
	return &FileAccessor{
		ctx:       ctx,
		apiServer: a,
		d:         a.driver,
	}
}

// CreateFileset creates a fileset with a single file
// TODO(peter): make this immediately exit the transaction and lead to extra transaction-wide
// fileset handling logic, including renewal
func (a *FileAccessor) CreateFileset(path string, data []byte, append bool) (string, error) {
	id, err := a.d.createFileSet(a.ctx, func(uw *fileset.UnorderedWriter) error {
		return uw.Put(path, "", append, bytes.NewReader(data))
	})
	if err != nil {
		return "", nil
	}
	return id.HexString(), nil
}

func (a *FileAccessor) GetPipelineDetails(info *pps.PipelineInfo) error {
	return grpcutil.GetPipelineDetails(a.ctx, a.apiServer, info)
}
