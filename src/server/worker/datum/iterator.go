package datum

import (
	"archive/tar"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"path"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	glob "github.com/pachyderm/ohmyglob"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
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

type pfsIterator struct {
	pachClient *client.APIClient
	input      *pps.PFSInput
}

func newPFSIterator(pachClient *client.APIClient, input *pps.PFSInput) Iterator {
	if input.Commit == "" {
		// this can happen if a pipeline with multiple inputs has been triggered
		// before all commits have inputs
		return &pfsIterator{}
	}
	return &pfsIterator{
		pachClient: pachClient,
		input:      input,
	}
}

func (pi *pfsIterator) Iterate(cb func(*Meta) error) error {
	if pi.input == nil {
		return nil
	}
	repo := pi.input.Repo
	branch := pi.input.Branch
	commit := pi.input.Commit
	pattern := pi.input.Glob
	return pi.pachClient.GlobFile(client.NewCommit(repo, branch, commit), pattern, func(fi *pfs.FileInfo) error {
		g := glob.MustCompile(pi.input.Glob, '/')
		// Remove the trailing slash to support glob replace on directory paths.
		p := strings.TrimRight(fi.File.Path, "/")
		joinOn := g.Replace(p, pi.input.JoinOn)
		groupBy := g.Replace(p, pi.input.GroupBy)
		return cb(&Meta{
			Inputs: []*common.Input{
				&common.Input{
					FileInfo:   fi,
					JoinOn:     joinOn,
					OuterJoin:  pi.input.OuterJoin,
					GroupBy:    groupBy,
					Name:       pi.input.Name,
					Lazy:       pi.input.Lazy,
					Branch:     pi.input.Branch,
					EmptyFiles: pi.input.EmptyFiles,
					S3:         pi.input.S3,
				},
			},
		})
	})
}

type unionIterator struct {
	iterators []Iterator
}

func newUnionIterator(pachClient *client.APIClient, inputs []*pps.Input) (Iterator, error) {
	ui := &unionIterator{}
	for _, input := range inputs {
		di, err := NewIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		ui.iterators = append(ui.iterators, di)
	}
	return ui, nil
}

func (ui *unionIterator) Iterate(cb func(*Meta) error) error {
	for _, iterator := range ui.iterators {
		if err := iterator.Iterate(cb); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

type crossIterator struct {
	iterators []Iterator
}

func newCrossIterator(pachClient *client.APIClient, inputs []*pps.Input) (Iterator, error) {
	ci := &crossIterator{}
	for _, input := range inputs {
		di, err := NewIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		ci.iterators = append(ci.iterators, di)
	}
	return ci, nil
}

func (ci *crossIterator) Iterate(cb func(*Meta) error) error {
	if len(ci.iterators) == 0 {
		return nil
	}
	return iterate(nil, ci.iterators, cb)
}

func iterate(crossInputs []*common.Input, iterators []Iterator, cb func(*Meta) error) error {
	if len(iterators) == 0 {
		return cb(&Meta{Inputs: crossInputs})
	}
	// TODO: Might want to exit fast for the zero datums case.
	err := iterators[0].Iterate(func(meta *Meta) error {
		return iterate(append(crossInputs, meta.Inputs...), iterators[1:], cb)
	})
	return errors.EnsureStack(err)
}

func newCronIterator(pachClient *client.APIClient, input *pps.CronInput) Iterator {
	return newPFSIterator(pachClient, &pps.PFSInput{
		Name:   input.Name,
		Repo:   input.Repo,
		Branch: "master",
		Commit: input.Commit,
		Glob:   "/*",
	})
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

type fileSetIterator struct {
	pachClient *client.APIClient
	commit     *pfs.Commit
}

// NewCommitIterator creates an iterator for the specified commit and repo.
func NewCommitIterator(pachClient *client.APIClient, commit *pfs.Commit) Iterator {
	return &fileSetIterator{
		pachClient: pachClient,
		commit:     commit,
	}
}

// NewFileSetIterator creates a new fileset iterator.
func NewFileSetIterator(pachClient *client.APIClient, fsID string) Iterator {
	return &fileSetIterator{
		pachClient: pachClient,
		commit:     client.NewRepo(client.FileSetsRepoName).NewCommit("", fsID),
	}
}

func newFileSetMultiIterator(pachClient *client.APIClient, fsID string) Iterator {
	return &fileSetMultiIterator{
		pachClient: pachClient,
		commit:     client.NewRepo(client.FileSetsRepoName).NewCommit("", fsID),
	}
}

func (fsi *fileSetIterator) Iterate(cb func(*Meta) error) error {
	r, err := fsi.pachClient.GetFileTAR(fsi.commit, path.Join("/", MetaPrefix, "*", MetaFileName))
	if err != nil {
		return err
	}
	tr := tar.NewReader(r)
	for {
		_, err := tr.Next()
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
		if err := cb(meta); err != nil {
			return err
		}
	}
}

type fileSetMultiIterator struct {
	pachClient *client.APIClient
	commit     *pfs.Commit
}

func (mi *fileSetMultiIterator) Iterate(cb func(*Meta) error) error {
	r, err := mi.pachClient.GetFileTAR(mi.commit, "/*")
	if err != nil {
		return err
	}
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

func existingMetaHash(meta *Meta) string {
	return meta.Hash
}

type joinIterator struct {
	pachClient *client.APIClient
	iterators  []Iterator
}

func newJoinIterator(pachClient *client.APIClient, inputs []*pps.Input) (Iterator, error) {
	ji := &joinIterator{
		pachClient: pachClient,
	}
	for _, input := range inputs {
		di, err := NewIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		ji.iterators = append(ji.iterators, di)
	}
	return ji, nil
}

func (ji *joinIterator) Iterate(cb func(*Meta) error) error {
	return ji.pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		filesets, err := computeDatumKeyFilesets(ji.pachClient, renewer, ji.iterators, true)
		if err != nil {
			return err
		}
		var dits []Iterator
		for _, fs := range filesets {
			dits = append(dits, newFileSetMultiIterator(ji.pachClient, fs))
		}
		return mergeByKey(dits, existingMetaHash, func(metas []*Meta) error {
			var crossInputs [][]*common.Input
			for _, m := range metas {
				crossInputs = append(crossInputs, m.Inputs)
			}
			err := newCrossListIterator(crossInputs).Iterate(func(meta *Meta) error {
				if len(meta.Inputs) == len(ji.iterators) {
					// all inputs represented, include all inputs
					return cb(meta)
				}
				var filtered []*common.Input
				for _, in := range meta.Inputs {
					if in.OuterJoin {
						filtered = append(filtered, in)
					}
				}
				if len(filtered) > 0 {
					return cb(&Meta{Inputs: filtered})
				}
				return nil
			})
			return errors.EnsureStack(err)
		})
	})
}

func computeDatumKeyFilesets(pachClient *client.APIClient, renewer *renew.StringSet, iterators []Iterator, isJoin bool) ([]string, error) {
	eg, ctx := errgroup.WithContext(pachClient.Ctx())
	pachClient = pachClient.WithCtx(ctx)
	filesets := make([]string, len(iterators))
	for i, di := range iterators {
		i := i
		di := di
		keyHasher := pfs.NewHash()
		marshaller := new(jsonpb.Marshaler)
		eg.Go(func() error {
			resp, err := pachClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
				err := di.Iterate(func(meta *Meta) error {
					for _, input := range meta.Inputs {
						// hash input keys to ensure consistently-shaped filepaths
						keyHasher.Reset()
						rawKey := input.GroupBy
						if isJoin {
							rawKey = input.JoinOn
						}
						keyHasher.Write([]byte(rawKey))
						key := hex.EncodeToString(keyHasher.Sum(nil))
						out, err := marshaller.MarshalToString(input)
						if err != nil {
							return errors.Wrap(err, "marshalling input for key aggregation")
						}
						if err := mf.PutFile(key, strings.NewReader(out), client.WithAppendPutFile()); err != nil {
							return errors.EnsureStack(err)
						}
					}
					return nil
				})
				return errors.EnsureStack(err)
			})
			if err != nil {
				return err
			}
			renewer.Add(ctx, resp.FileSetId)
			filesets[i] = resp.FileSetId
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return filesets, nil
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

type groupIterator struct {
	pachClient *client.APIClient
	iterators  []Iterator
}

func newGroupIterator(pachClient *client.APIClient, inputs []*pps.Input) (Iterator, error) {
	gi := &groupIterator{
		pachClient: pachClient,
	}
	for _, input := range inputs {
		di, err := NewIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		gi.iterators = append(gi.iterators, di)
	}
	return gi, nil
}

func (gi *groupIterator) Iterate(cb func(*Meta) error) error {
	return gi.pachClient.WithRenewer(func(ctx context.Context, renewer *renew.StringSet) error {
		filesets, err := computeDatumKeyFilesets(gi.pachClient, renewer, gi.iterators, false)
		if err != nil {
			return err
		}
		var dits []Iterator
		for _, fs := range filesets {
			dits = append(dits, newFileSetMultiIterator(gi.pachClient, fs))
		}
		return mergeByKey(dits, existingMetaHash, func(metas []*Meta) error {
			var allInputs []*common.Input
			for _, m := range metas {
				allInputs = append(allInputs, m.Inputs...)
			}
			return cb(&Meta{Inputs: allInputs})
		})
	})
}

type idGenerator = func(*Meta) string

func metaInputID(meta *Meta) string {
	return common.DatumID(meta.Inputs)
}

// Merge merges multiple datum iterators (key is datum ID).
func Merge(dits []Iterator, cb func([]*Meta) error) error {
	return mergeByKey(dits, metaInputID, cb)
}

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

// NewIterator creates a new datum iterator.
func NewIterator(pachClient *client.APIClient, input *pps.Input) (Iterator, error) {
	var iterator Iterator
	var err error
	switch {
	case input.Pfs != nil:
		iterator = newPFSIterator(pachClient, input.Pfs)
	case input.Union != nil:
		iterator, err = newUnionIterator(pachClient, input.Union)
		if err != nil {
			return nil, err
		}
	case input.Cross != nil:
		iterator, err = newCrossIterator(pachClient, input.Cross)
		if err != nil {
			return nil, err
		}
	case input.Join != nil:
		iterator, err = newJoinIterator(pachClient, input.Join)
		if err != nil {
			return nil, err
		}
	case input.Group != nil:
		iterator, err = newGroupIterator(pachClient, input.Group)
		if err != nil {
			return nil, err
		}
	case input.Cron != nil:
		iterator = newCronIterator(pachClient, input.Cron)
	default:
		return nil, errors.Errorf("unrecognized input type: %v", input)
	}
	return newIndexIterator(iterator), nil
}

type indexIterator struct {
	iterator Iterator
}

func newIndexIterator(iterator Iterator) Iterator {
	return &indexIterator{
		iterator: iterator,
	}
}

func (ii *indexIterator) Iterate(cb func(*Meta) error) error {
	index := int64(0)
	err := ii.iterator.Iterate(func(meta *Meta) error {
		meta.Index = index
		index++
		return cb(meta)
	})
	return errors.EnsureStack(err)
}
