package transactionenv

import (
	"bytes"
	"context"
	"crypto/md5"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
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
	path string
	data []byte
}

func (b basicPutFile) Hash() [16]byte {
	hasher := md5.New()
	hasher.Write([]byte(b.path))
	hasher.Write([]byte{0})
	hasher.Write(b.data)
	return md5.Sum(nil)
}

type refresher struct {
	env           serviceenv.ServiceEnv
	filesetQueue  []basicPutFile
	filesets      map[[16]byte]string
	pipelineCache map[string]pipelineResult
}

func (env *TransactionEnv) NewRefresher(info *transaction.TransactionInfo) *refresher {
	out := &refresher{
		env:           env.serviceEnv,
		filesets:      map[[16]byte]string{},
		pipelineCache: map[string]pipelineResult{},
	}
	for _, req := range info.Requests {
		if req.CreatePipeline != nil {
			out.pipelineCache[req.CreatePipeline.Pipeline.Name] = pipelineResult{}
		}
	}

	for _, resp := range info.Responses {
		if resp.CreatePipelineResponse != nil {
			// start from scratch
			// TODO(peter): store fileset hashes or read the data so we don't have to do this
			resp.CreatePipelineResponse.FileSetId = ""
			resp.CreatePipelineResponse.PrevPipelineVersion = 0
		}
	}

	return out
}

func (env *TransactionEnv) withRefreshLoop(ctx context.Context, r *refresher, cb func(txncontext.FilesetManager) error) error {
	if r == nil {
		r = &refresher{
			env:           env.serviceEnv,
			filesets:      map[[16]byte]string{},
			pipelineCache: map[string]pipelineResult{},
		}
	}
	for {
		r.refresh(ctx)
		if err := cb(r); err == nil || !isErrTransactionConflict(err) {
			return err
		}
	}
}

// CreateFileset checks if a fileset with the given data has already been created
func (r *refresher) CreateFileset(path string, data []byte) (string, error) {
	req := basicPutFile{
		path: path,
		data: data,
	}
	if id, ok := r.filesets[req.Hash()]; ok {
		return id, nil
	}
	r.filesetQueue = append(r.filesetQueue, req)
	// return error to force transaction restart
	return "", errTransactionConflict{}
}

type pipelineResult struct {
	info *pps.PipelineInfo
	err  error
}

func (r *refresher) LatestPipelineInfo(txnCtx *txncontext.TransactionContext, pipeline *pps.Pipeline) (*pps.PipelineInfo, error) {
	newInfo, err := r.env.PpsServer().InspectPipelineInTransaction(txnCtx, pipeline.Name)
	if err != nil {
		if errutil.IsNotFoundError(err) {
			// pipeline doesn't exist, drop it from the refresh cache
			delete(r.pipelineCache, pipeline.Name)
		}
		return nil, err
	}

	if cached, ok := r.pipelineCache[pipeline.Name]; ok && (cached.err != nil || cached.info.GetVersion() == newInfo.Version) {
		// up to date
		return cached.info, cached.err
	}

	// clear space for updated info and fail the transaction
	r.pipelineCache[pipeline.Name] = pipelineResult{}
	return nil, errTransactionConflict{}
}

func (r *refresher) refresh(ctx context.Context) error {
	// renew existing filesets
	for _, id := range r.filesets {
		if _, err := r.env.PfsServer().RenewFileSet(ctx, &pfs.RenewFileSetRequest{FileSetId: id, TtlSeconds: defaultTTLSeconds}); err != nil {
			return err
		}
	}

	// create new filesets
	for _, req := range r.filesetQueue {
		resp, err := r.env.GetPachClient(ctx).WithCreateFileSetClient(func(m client.ModifyFile) error {
			return m.PutFile(req.path, bytes.NewReader(req.data))
		})
		if err != nil {
			return nil
		}
		// add to set of filesets and store in destination
		r.filesets[req.Hash()] = resp.FileSetId
	}
	r.filesetQueue = r.filesetQueue[:0]

	// refetch pipeline infos, but save error for caller rather than exiting
	for pipeline := range r.pipelineCache {
		info, err := r.env.PpsServer().InspectPipeline(ctx, &pps.InspectPipelineRequest{Pipeline: client.NewPipeline(pipeline), Details: true})
		r.pipelineCache[pipeline] = pipelineResult{info: info, err: err}
	}
	return nil
}
