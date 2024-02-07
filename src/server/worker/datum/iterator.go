package datum

import (
	"archive/tar"
	"context"
	"io"
	"path"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
)

// Iterator is the standard interface for a datum iterator.
type Iterator interface {
	// Iterate iterates over a set of datums.
	Iterate(func(*Meta) error) error
}

// NewIterator creates a new datum iterator.
// TODO: Maybe add a renewer parameter to keep file set alive?
func NewIterator(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, input *pps.Input) (Iterator, error) {
	fileSetID, err := Create(ctx, c, taskDoer, input)
	if err != nil {
		return nil, err
	}
	return NewFileSetIterator(ctx, c, fileSetID, nil), nil
}

func NewCreateDatumStreamIterator(ctx context.Context, c pfs.APIClient, taskDoer task.Doer, input *pps.Input) Iterator {
	ctx = pctx.Child(ctx, "CreateDatumStream")
	cds := &createDatumStream{
		ctx:               ctx,
		c:                 c,
		taskDoer:          taskDoer,
		input:             input,
		requestDatumsChan: make(chan bool),
		fsidChan:          make(chan string),
		errChan:           make(chan error, 1),
		doneChan:          make(chan bool),
	}
	go cds.create(cds.input, cds.fsidChan)
	return &createDatumStreamIterator{
		createDatumStream: cds,
		metaBuffer:        make([]*Meta, 0),
	}
}

type createDatumStreamIterator struct {
	*createDatumStream
	metaBuffer []*Meta
}

// Return value of nil means all datums have been returned.
// Return value of errutil.ErrBreak means that iteration is paused and
// will be resumed when client asks for more.
// Return value of anything else means an error.
func (it *createDatumStreamIterator) Iterate(cb func(*Meta) error) error {
	for {
		// Consume leftover datums from the buffer first before asking
		// to create more.
		if len(it.metaBuffer) == 0 {
			select {
			case <-it.doneChan:
				return nil
			case it.requestDatumsChan <- true:
			}
			select {
			case fsid := <-it.fsidChan:
				fsidIt := NewFileSetIterator(it.ctx, it.c, fsid, nil)
				if err := fsidIt.Iterate(func(meta *Meta) error {
					err := cb(meta)
					// If callback has consumed all the datums it needs,
					// add the leftover datums to the buffer.
					if errors.Is(err, errutil.ErrBreak) {
						it.metaBuffer = append(it.metaBuffer, meta)
						return nil
					}
					return err
				}); err != nil {
					return errors.Wrap(err, "single file set iteration")
				}
				if len(it.metaBuffer) > 0 {
					return errutil.ErrBreak
				}
			case err := <-it.errChan:
				return errors.Wrap(err, "create()")
			}
		} else {
			meta := it.metaBuffer[0]
			if err := cb(meta); err != nil {
				return errors.Wrap(err, "callback error")
			}
			it.metaBuffer = it.metaBuffer[1:]
		}
	}
}


// Hasher is the standard interface for a datum hasher.
type Hasher interface {
	// Hash computes the datum hash based on the inputs.
	Hash([]*common.Input) string
}

type jobIterator struct {
	iterator Iterator
	job      *pps.Job
	hasher   Hasher
}

// NewJobIterator creates a new job iterator.
func NewJobIterator(iterator Iterator, job *pps.Job, hasher Hasher) Iterator {
	return &jobIterator{
		iterator: iterator,
		job:      job,
		hasher:   hasher,
	}
}

func (ji *jobIterator) Iterate(cb func(*Meta) error) error {
	err := ji.iterator.Iterate(func(meta *Meta) error {
		meta.Job = ji.job
		meta.Hash = ji.hasher.Hash(meta.Inputs)
		return cb(meta)
	})
	return errors.EnsureStack(err)
}

type commitIterator struct {
	ctx       context.Context
	c         pfs.APIClient
	commit    *pfs.Commit
	pathRange *pfs.PathRange
}

// NewCommitIterator creates an iterator for the specified commit and repo.
func NewCommitIterator(ctx context.Context, c pfs.APIClient, commit *pfs.Commit, pathRange *pfs.PathRange) Iterator {
	ctx = pctx.Child(ctx, "commitIterator")
	return &commitIterator{
		ctx:       ctx,
		c:         c,
		commit:    commit,
		pathRange: pathRange,
	}
}

func (ci *commitIterator) Iterate(cb func(*Meta) error) error {
	return iterateMeta(ci.ctx, ci.c, ci.commit, ci.pathRange, func(_ string, meta *Meta) error {
		return cb(meta)
	})
}

type fileTarGetter interface {
	GetFileTAR(ctx context.Context, in *pfs.GetFileRequest, opts ...grpc.CallOption) (pfs.API_GetFileTARClient, error)
}

func iterateMeta(ctx context.Context, tarGetter fileTarGetter, commit *pfs.Commit, pathRange *pfs.PathRange, cb func(string, *Meta) error) error {
	// TODO: This code ensures that we only read metadata for the meta files.
	// There may be a better way to do this.
	if pathRange == nil {
		pathRange = &pfs.PathRange{}
	}
	if pathRange.Lower == "" {
		pathRange.Lower = path.Join("/", common.MetaPrefix)
	}
	if pathRange.Upper == "" {
		pathRange.Upper = path.Join("/", common.MetaPrefix+"_")
	}
	req := &pfs.GetFileRequest{
		File:      commit.NewFile(path.Join("/", common.MetaFilePath("*"))),
		PathRange: pathRange,
	}
	ctx, cancel := pctx.WithCancel(ctx)
	defer cancel()
	client, err := tarGetter.GetFileTAR(ctx, req)
	if err != nil {
		return errors.EnsureStack(err)
	}
	r := grpcutil.NewStreamingBytesReader(client, nil)
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if pfsserver.IsFileNotFoundErr(err) || errors.Is(err, io.EOF) {
				return nil
			}
			return errors.Wrap(err, "next")
		}
		meta := &Meta{}
		// This intentionally reads only the first Meta message stored in the meta file.
		//
		// TODO: should this return the first, or the last?  No-one on
		// the team is currently certain.
		decoder := protoutil.NewProtoJSONDecoder(tr, protojson.UnmarshalOptions{})
		if err := decoder.UnmarshalNext(meta); err != nil {
			return errors.Wrapf(err, "could not unmarshal protojson meta in %s %v", req.File, req.PathRange)
		}
		migrateMetaInputsV2_6_0(meta)
		if err := cb(hdr.Name, meta); err != nil {
			return err
		}
	}
}

// TODO(provenance): call getter internally, and leave note for future
func migrateMetaInputsV2_6_0(meta *Meta) {
	for _, i := range meta.Inputs {
		i.FileInfo.File.Commit.Repo = i.FileInfo.File.Commit.AccessRepo()
	}
}

// NewFileSetIterator creates a new fileset iterator.
func NewFileSetIterator(ctx context.Context, c pfs.APIClient, fsID string, pathRange *pfs.PathRange) Iterator {
	return NewCommitIterator(ctx, c, client.NewRepo(pfs.DefaultProjectName, client.FileSetsRepoName).NewCommit("", fsID), pathRange)
}

type fileSetMultiIterator struct {
	ctx       context.Context
	pfs       pfs.APIClient
	commit    *pfs.Commit
	pathRange *pfs.PathRange
}

func newFileSetMultiIterator(ctx context.Context, c pfs.APIClient, fsID string, pathRange *pfs.PathRange) Iterator {
	return &fileSetMultiIterator{
		ctx:       ctx,
		pfs:       c,
		commit:    client.NewRepo(pfs.DefaultProjectName, client.FileSetsRepoName).NewCommit("", fsID),
		pathRange: pathRange,
	}
}

func (mi *fileSetMultiIterator) Iterate(cb func(*Meta) error) error {
	req := &pfs.GetFileRequest{
		File:      mi.commit.NewFile("/*"),
		PathRange: mi.pathRange,
	}
	ctx, cancel := pctx.WithCancel(mi.ctx)
	defer cancel()
	client, err := mi.pfs.GetFileTAR(ctx, req)
	if err != nil {
		return errors.EnsureStack(err)
	}
	r := grpcutil.NewStreamingBytesReader(client, nil)
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if pfsserver.IsFileNotFoundErr(err) || errors.Is(err, io.EOF) {
				return nil
			}
			return errors.EnsureStack(err)
		}
		var meta Meta
		// kind of an abuse of the field, just stick this to key off of
		meta.Hash = hdr.Name
		decoder := protoutil.NewProtoJSONDecoder(tr, protojson.UnmarshalOptions{})
		for {
			input := new(common.Input)
			if err := decoder.UnmarshalNext(input); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return errors.Wrap(err, "error unmarshalling input")
			}
			meta.Inputs = append(meta.Inputs, input)
		}
		if err := cb(&meta); err != nil {
			return err
		}
	}
}

type crossIterator struct {
	iterators []Iterator
}

func (ci *crossIterator) Iterate(cb func(*Meta) error) error {
	if len(ci.iterators) == 0 {
		return nil
	}
	return iterate(nil, ci.iterators, cb)
}

func newCrossListIterator(crossInputs [][]*common.Input) Iterator {
	ci := &crossIterator{}
	for _, inputs := range crossInputs {
		ci.iterators = append(ci.iterators, newListIterator(inputs))
	}
	return ci
}

type listIterator struct {
	inputs []*common.Input
}

func newListIterator(inputs []*common.Input) Iterator {
	return &listIterator{
		inputs: inputs,
	}
}

func (li *listIterator) Iterate(cb func(*Meta) error) error {
	for _, input := range li.inputs {
		if err := cb(&Meta{
			Inputs: []*common.Input{input},
		}); err != nil {
			return err
		}
	}
	return nil
}

// Merge merges multiple datum iterators (key is datum ID).
func Merge(dits []Iterator, cb func([]*Meta) error) error {
	return mergeByKey(dits, metaInputID, cb)
}

func metaInputID(meta *Meta) string {
	return common.DatumID(meta.Inputs)
}

type idGenerator = func(*Meta) string

func mergeByKey(dits []Iterator, idFunc idGenerator, cb func([]*Meta) error) error {
	var ss []stream.Stream
	for _, dit := range dits {
		ss = append(ss, newDatumStream(dit, len(ss), idFunc))
	}
	pq := stream.NewPriorityQueue(ss, compare)
	return pq.Iterate(func(ss []stream.Stream) error {
		var metas []*Meta
		for _, s := range ss {
			metas = append(metas, s.(*datumStream).meta)
		}
		return cb(metas)
	})
}

type datumStream struct {
	meta     *Meta
	id       string
	metaChan chan *Meta
	errChan  chan error
	idFunc   idGenerator
}

func newDatumStream(dit Iterator, priority int, idFunc idGenerator) *datumStream {
	metaChan := make(chan *Meta)
	errChan := make(chan error, 1)
	go func() {
		if err := dit.Iterate(func(meta *Meta) error {
			metaChan <- meta
			return nil
		}); err != nil {
			errChan <- err
			return
		}
		close(metaChan)
	}()
	return &datumStream{
		metaChan: metaChan,
		errChan:  errChan,
		idFunc:   idFunc,
	}
}

func (ds *datumStream) Next() error {
	select {
	case meta, more := <-ds.metaChan:
		if !more {
			return io.EOF
		}
		ds.meta = meta
		ds.id = ds.idFunc(meta)
		return nil
	case err := <-ds.errChan:
		return err
	}
}

func compare(s1, s2 stream.Stream) int {
	ds1 := s1.(*datumStream)
	ds2 := s2.(*datumStream)
	return strings.Compare(ds1.id, ds2.id)
}
