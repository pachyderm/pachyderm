package datum

import (
	"archive/tar"
	"context"
	"encoding/json"
	"io"
	"path"

	"github.com/gogo/protobuf/jsonpb"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
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
func NewIterator(pachClient *client.APIClient, taskDoer task.Doer, input *pps.Input) (Iterator, error) {
	fileSetID, err := Create(pachClient, taskDoer, input)
	if err != nil {
		return nil, err
	}
	return NewFileSetIterator(pachClient, fileSetID, nil), nil
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
	pachClient *client.APIClient
	commit     *pfs.Commit
	pathRange  *pfs.PathRange
}

// NewCommitIterator creates an iterator for the specified commit and repo.
func NewCommitIterator(pachClient *client.APIClient, commit *pfs.Commit, pathRange *pfs.PathRange) Iterator {
	return &commitIterator{
		pachClient: pachClient,
		commit:     commit,
		pathRange:  pathRange,
	}
}

func (ci *commitIterator) Iterate(cb func(*Meta) error) error {
	return iterateMeta(ci.pachClient, ci.commit, ci.pathRange, func(_ string, meta *Meta) error {
		return cb(meta)
	})
}

func iterateMeta(pachClient *client.APIClient, commit *pfs.Commit, pathRange *pfs.PathRange, cb func(string, *Meta) error) error {
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
	ctx, cancel := context.WithCancel(pachClient.Ctx())
	defer cancel()
	client, err := pachClient.PfsAPIClient.GetFileTAR(ctx, req)
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
		meta := &Meta{}
		if err := jsonpb.Unmarshal(tr, meta); err != nil {
			return errors.EnsureStack(err)
		}
		if err := cb(hdr.Name, meta); err != nil {
			return err
		}
	}
}

// NewFileSetIterator creates a new fileset iterator.
func NewFileSetIterator(pachClient *client.APIClient, fsID string, pathRange *pfs.PathRange) Iterator {
	return NewCommitIterator(pachClient, client.NewProjectRepo(pfs.DefaultProjectName, client.FileSetsRepoName).NewCommit("", fsID), pathRange)
}

type fileSetMultiIterator struct {
	pachClient *client.APIClient
	commit     *pfs.Commit
	pathRange  *pfs.PathRange
}

func newFileSetMultiIterator(pachClient *client.APIClient, fsID string, pathRange *pfs.PathRange) Iterator {
	return &fileSetMultiIterator{
		pachClient: pachClient,
		commit:     client.NewProjectRepo(pfs.DefaultProjectName, client.FileSetsRepoName).NewCommit("", fsID),
		pathRange:  pathRange,
	}
}

func (mi *fileSetMultiIterator) Iterate(cb func(*Meta) error) error {
	req := &pfs.GetFileRequest{
		File:      mi.commit.NewFile("/*"),
		PathRange: mi.pathRange,
	}
	ctx, cancel := context.WithCancel(mi.pachClient.Ctx())
	defer cancel()
	client, err := mi.pachClient.PfsAPIClient.GetFileTAR(ctx, req)
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
		decoder := json.NewDecoder(tr)
		for {
			input := new(common.Input)
			if err := jsonpb.UnmarshalNext(decoder, input); err != nil {
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
	ctx := context.TODO()
	type metaEntry struct {
		Meta *Meta
		ID   string
	}
	cpMetaEntry := func(dst, src *metaEntry) { *dst = *src }
	var ss []stream.Peekable[metaEntry]
	for _, dit := range dits {
		it := stream.NewFromForEach(ctx, cpMetaEntry, func(fn func(metaEntry) error) error {
			return dit.Iterate(func(x *Meta) error {
				return fn(metaEntry{
					Meta: x,
					ID:   idFunc(x),
				})
			})
		})
		pk := stream.NewPeekable(it, cpMetaEntry)
		ss = append(ss, pk)
	}
	m := stream.NewMerger(ss, func(a, b metaEntry) bool {
		return a.ID < b.ID
	})
	return stream.ForEach[stream.Merged[metaEntry]](ctx, m, func(x stream.Merged[metaEntry]) error {
		var metas []*Meta
		for _, me := range x.Values {
			metas = append(metas, me.Meta)
		}
		return cb(metas)
	})
}
