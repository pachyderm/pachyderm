package transactionenv

import (
	"bytes"
	"context"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

type errTransactionConflict struct{}

func (err errTransactionConflict) Is(other error) bool {
	_, ok := other.(errTransactionConflict)
	return ok
}

func (err errTransactionConflict) Error() string {
	return "transaction conflict, will be reattempted"
}

func isErrTransactionConflict(err error) bool {
	return errors.Is(err, errTransactionConflict{})
}

type basicPutFile struct {
	dest *string
	path string
	data []byte
}

type refresher struct {
	env           serviceenv.ServiceEnv
	filesetQueue  []basicPutFile
	filesets      map[fileset.ID]struct{}
	pipelineCache map[string]*pps.PipelineInfo
}

func (env *TransactionEnv) withRefreshLoop(ctx context.Context, cb func(txncontext.FilesetManager) error) error {
	r := refresher{
		env:           env.serviceEnv,
		filesets:      map[fileset.ID]struct{}{},
		pipelineCache: map[string]*pps.PipelineInfo{},
	}
	for {
		r.refresh(ctx)
		if err := cb(&r); err == nil || !isErrTransactionConflict(err) {
			return err
		}
	}
}

// CreateFileset creates a fileset with a single file
func (r *refresher) CreateFileset(idDest *string, path string, data []byte) error {
	if *idDest != "" {
		// try to clean up
		if parsed, err := fileset.ParseID(*idDest); err != nil {
			delete(r.filesets, *parsed)
		}
	}
	r.filesetQueue = append(r.filesetQueue, basicPutFile{
		dest: idDest,
		path: path,
		data: data,
	})
	// return error to force transaction restart
	return errTransactionConflict{}
}

func (r *refresher) LatestPipelineInfo(txnCtx *txncontext.TransactionContext, pipeline *pps.Pipeline) (*pps.PipelineInfo, error) {
	newInfo, err := r.env.PpsServer().InspectPipelineInTransaction(txnCtx, pipeline.Name)
	if err != nil && col.IsErrNotFound(err) {
		// pipeline doesn't exist, so we can return nil info, meaning it's up to date in this transaction
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if cached, ok := r.pipelineCache[pipeline.Name]; ok && cached.Version == newInfo.Version {
		// up to date
		return r.pipelineCache[pipeline.Name], nil
	}

	// clear space for updated info and fail the transaction
	r.pipelineCache[pipeline.Name] = nil
	return nil, errTransactionConflict{}
}

func (r *refresher) refresh(ctx context.Context) error {
	// renew existing filesets
	for id := range r.filesets {
		if _, err := r.env.PfsServer().RenewFileSet(ctx, &pfs.RenewFileSetRequest{FileSetId: id.HexString(), TtlSeconds: 60 * 10}); err != nil {
			return err
		}
	}

	// create new filesets
	for _, req := range r.filesetQueue {
		id, err := r.env.PfsServer().CreateFileSetCallback(ctx, func(uw *fileset.UnorderedWriter) error {
			return uw.Put(req.path, "", false, bytes.NewReader(req.data))
		})
		if err != nil {
			return nil
		}
		// add to set of filesets and store in destination
		r.filesets[*id] = struct{}{}
		*req.dest = id.HexString()
	}
	r.filesetQueue = r.filesetQueue[:0]

	// refetch pipeline infos
	for pipeline := range r.pipelineCache {
		info, err := r.env.PpsServer().InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: client.NewPipeline(pipeline), Details: true})
		if errutil.IsNotFoundError(err) {
			delete(r.pipelineCache, pipeline)
		} else if err != nil {
			return err
		} else {
			r.pipelineCache[pipeline] = info
		}
	}
	return nil
}
