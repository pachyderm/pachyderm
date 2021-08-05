package transactionenv

import (
	"bytes"
	"context"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
)

const defaultTTLSeconds = 60 * 10

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

func (env *TransactionEnv) NewRefresher(ctx context.Context, info *transaction.TransactionInfo) *refresher {
	out := &refresher{
		env:           env.serviceEnv,
		filesets:      map[fileset.ID]struct{}{},
		pipelineCache: map[string]*pps.PipelineInfo{},
	}
	for _, req := range info.Requests {
		if req.CreatePipeline != nil {
			out.pipelineCache[req.CreatePipeline.Pipeline.Name] = nil
		}
	}

	for _, resp := range info.Responses {
		if resp.CreatePipelineResponse != nil {
			id := resp.CreatePipelineResponse.FileSetId
			parsed, err := fileset.ParseID(id)
			if err == nil {
				_, err = env.serviceEnv.PfsServer().RenewFileSet(ctx, &pfs.RenewFileSetRequest{FileSetId: id, TtlSeconds: defaultTTLSeconds})
			}
			if err != nil {
				// don't error, just assume fileset is gone and start from scratch
				resp.CreatePipelineResponse.FileSetId = ""
				resp.CreatePipelineResponse.PrevPipelineVersion = 0
			} else {
				// add this to the set to be renewed
				out.filesets[*parsed] = struct{}{}
			}
		}
	}

	return out
}

func (env *TransactionEnv) withRefreshLoop(ctx context.Context, r *refresher, cb func(txncontext.FilesetManager) error) error {
	if r == nil {
		r = &refresher{
			env:           env.serviceEnv,
			filesets:      map[fileset.ID]struct{}{},
			pipelineCache: map[string]*pps.PipelineInfo{},
		}
	}
	for {
		r.refresh(ctx)
		if err := cb(r); err == nil || !isErrTransactionConflict(err) {
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
	if err != nil && errutil.IsNotFoundError(err) {
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
		if _, err := r.env.PfsServer().RenewFileSet(ctx, &pfs.RenewFileSetRequest{FileSetId: id.HexString(), TtlSeconds: defaultTTLSeconds}); err != nil {
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
