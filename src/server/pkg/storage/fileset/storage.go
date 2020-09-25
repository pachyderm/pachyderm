package fileset

import (
	"context"
	"fmt"
	"io"
	"math"
	"path"
	"strings"
	"time"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	semanticPrefix = "root"
	// TODO Not sure if these are the tags we should use, but the header and padding tag should show up before and after respectively in the
	// lexicographical ordering of file content tags.
	// headerTag is the tag used for the tar header bytes.
	headerTag = ""
	// paddingTag is the tag used for the padding bytes at the end of a tar entry.
	paddingTag = "~"
	// DefaultMemoryThreshold is the default for the memory threshold that must
	// be met before a file set part is serialized (excluding close).
	DefaultMemoryThreshold = 1024 * units.MB
	// DefaultShardThreshold is the default for the size threshold that must
	// be met before a shard is created by the shard function.
	DefaultShardThreshold = 1024 * units.MB
	// DefaultLevelZeroSize is the default size for level zero in the compacted
	// representation of a file set.
	DefaultLevelZeroSize = 1 * units.MB
	// DefaultLevelSizeBase is the default base of the exponential growth function
	// for level sizes in the compacted representation of a file set.
	DefaultLevelSizeBase = 10
	// Diff is the suffix of a path that points to the diff of the prefix.
	Diff = "diff"
	// Compacted is the suffix of a path that points to the compaction of the prefix.
	Compacted = "compacted"
)

var (
	ErrNoTTLSet       = errors.Errorf("no ttl set for fileset or any subfilesets")
	ErrNoFileSetFound = errors.Errorf("no fileset found")
)

// Storage is the abstraction that manages fileset storage.
type Storage struct {
	objC                         obj.Client
	chunks                       *chunk.Storage
	memThreshold, shardThreshold int64
	levelZeroSize                int64
	levelSizeBase                int
	filesetSem                   *semaphore.Weighted
}

// NewStorage creates a new Storage.
func NewStorage(objC obj.Client, chunks *chunk.Storage, opts ...StorageOption) *Storage {
	s := &Storage{
		objC:           objC,
		chunks:         chunks,
		memThreshold:   DefaultMemoryThreshold,
		shardThreshold: DefaultShardThreshold,
		levelZeroSize:  DefaultLevelZeroSize,
		levelSizeBase:  DefaultLevelSizeBase,
		filesetSem:     semaphore.NewWeighted(math.MaxInt64),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ChunkStorage returns the underlying chunk storage instance for this storage instance.
func (s *Storage) ChunkStorage() *chunk.Storage {
	return s.chunks
}

// New creates a new in-memory fileset.
func (s *Storage) New(ctx context.Context, fileSet, defaultTag string, opts ...UnorderedWriterOption) (*UnorderedWriter, error) {
	fileSet = applyPrefix(fileSet)
	return newUnorderedWriter(ctx, s, fileSet, s.memThreshold, defaultTag, opts...)
}

// NewWriter makes a Writer backed by the path `fileSet` in object storage.
func (s *Storage) NewWriter(ctx context.Context, fileSet string, opts ...WriterOption) *Writer {
	fileSet = applyPrefix(fileSet)
	return s.newWriter(ctx, fileSet, opts...)
}

func (s *Storage) newWriter(ctx context.Context, fileSet string, opts ...WriterOption) *Writer {
	return newWriter(ctx, s.objC, s.chunks, fileSet, opts...)
}

// NewReader makes a Reader backed by the path `fileSet` in object storage.
func (s *Storage) NewReader(ctx context.Context, fileSet string, opts ...index.Option) *Reader {
	fileSet = applyPrefix(fileSet)
	return s.newReader(ctx, fileSet, opts...)
}

// TODO Expose some notion of read ahead (read a certain number of chunks in parallel).
// this will be necessary to speed up reading large files.
func (s *Storage) newReader(ctx context.Context, fileSet string, opts ...index.Option) *Reader {
	return newReader(ctx, s.objC, s.chunks, fileSet, opts...)
}

// NewMergeReader returns a merge reader for a set for filesets.
func (s *Storage) NewMergeReader(ctx context.Context, fileSets []string, opts ...index.Option) (*MergeReader, error) {
	fileSets = applyPrefixes(fileSets)
	return s.newMergeReader(ctx, fileSets, opts...)
}

func (s *Storage) newMergeReader(ctx context.Context, fileSets []string, opts ...index.Option) (*MergeReader, error) {
	var rs []*Reader
	for _, fileSet := range fileSets {
		if err := s.objC.Walk(ctx, fileSet, func(name string) error {
			rs = append(rs, s.newReader(ctx, name, opts...))
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return newMergeReader(rs), nil
}

// OpenFileSet makes a source which will iterate over the prefix fileSet
func (s *Storage) OpenFileSet(ctx context.Context, fileSet string, opts ...index.Option) FileSet {
	return &mergeSource{
		s: s,
		getReader: func() (*MergeReader, error) {
			return s.NewMergeReader(ctx, []string{fileSet}, opts...)
		},
	}
}

// ResolveIndexes resolves index entries that are spread across multiple filesets.
// DEPRECATED: Use NewIndexResolver
func (s *Storage) ResolveIndexes(ctx context.Context, fileSets []string, cb func(*index.Index) error, opts ...index.Option) error {
	mr, err := s.NewMergeReader(ctx, fileSets, opts...)
	if err != nil {
		return err
	}
	w := s.newWriter(ctx, "", WithNoUpload(), WithIndexCallback(cb))
	if err := mr.WriteTo(w); err != nil {
		return err
	}
	return w.Close()
}

// Shard shards the merge of the file sets with the passed in prefix into file ranges.
// TODO This should be extended to be more configurable (different criteria
// for creating shards).
func (s *Storage) Shard(ctx context.Context, fileSets []string, shardFunc ShardFunc) error {
	mr, err := s.NewMergeReader(ctx, fileSets)
	if err != nil {
		return err
	}
	return shard(mr, s.shardThreshold, shardFunc)
}

// Copy copies the fileset at srcPrefix to dstPrefix. It does *not* perform compaction
// ttl sets the time to live on the keys under dstPrefix if ttl == 0, it is ignored
func (s *Storage) Copy(ctx context.Context, srcPrefix, dstPrefix string, ttl time.Duration) error {
	srcPrefix = applyPrefix(srcPrefix)
	dstPrefix = applyPrefix(dstPrefix)
	// TODO: perform this atomically with postgres
	return s.objC.Walk(ctx, srcPrefix, func(srcPath string) error {
		dstPath := dstPrefix + srcPath[len(srcPrefix):]
		if err := copyObject(ctx, s.objC, srcPath, dstPath); err != nil {
			return err
		}
		if ttl > 0 {
			_, err := s.SetTTL(ctx, removePrefix(dstPath), ttl)
			return err
		}
		return nil
	})
}

// CompactStats contains information about what was compacted.
type CompactStats struct {
	OutputSize int64
}

// Compact compacts a set of filesets into an output fileset.
func (s *Storage) Compact(ctx context.Context, outputFileSet string, inputFileSets []string, ttl time.Duration, opts ...index.Option) (*CompactStats, error) {
	outputFileSet = applyPrefix(outputFileSet)
	inputFileSets = applyPrefixes(inputFileSets)
	var size int64
	w := s.newWriter(ctx, outputFileSet, WithIndexCallback(func(idx *index.Index) error {
		size += idx.SizeBytes
		return nil
	}))
	mr, err := s.newMergeReader(ctx, inputFileSets, opts...)
	if err != nil {
		return nil, err
	}
	if err := mr.WriteTo(w); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	if ttl > 0 {
		if _, err := s.SetTTL(ctx, removePrefix(outputFileSet), ttl); err != nil {
			return nil, err
		}
	}
	return &CompactStats{OutputSize: size}, nil
}

// CompactSpec specifies the input and output for a compaction operation.
type CompactSpec struct {
	Output string
	Input  []string
}

// CompactSpec returns a compaction specification that determines the input filesets (the diff file set and potentially
// compacted filesets) and output fileset.
func (s *Storage) CompactSpec(ctx context.Context, fileSet string, compactedFileSet ...string) (*CompactSpec, error) {
	if len(compactedFileSet) > 1 {
		return nil, errors.Errorf("multiple compacted FileSets")
	}
	// Internal vs external path transforms
	fileSet = applyPrefix(fileSet)
	compactedFileSet = applyPrefixes(compactedFileSet)
	spec, err := s.compactSpec(ctx, fileSet, compactedFileSet...)
	if err != nil {
		return nil, err
	}
	spec.Input = removePrefixes(spec.Input)
	spec.Output = removePrefix(spec.Output)
	return spec, nil
}

func (s *Storage) compactSpec(ctx context.Context, fileSet string, compactedFileSet ...string) (ret *CompactSpec, retErr error) {
	idx, err := index.GetTopLevelIndex(ctx, s.objC, path.Join(fileSet, Diff))
	if err != nil {
		return nil, err
	}
	size := idx.SizeBytes
	spec := &CompactSpec{
		Input: []string{path.Join(fileSet, Diff)},
	}
	var level int
	// Handle first commit being compacted.
	if len(compactedFileSet) == 0 {
		for size > s.levelSize(level) {
			level++
		}
		spec.Output = path.Join(fileSet, Compacted, levelName(level))
		return spec, nil
	}
	// While we can't fit it all in the current level
	for {
		levelPath := path.Join(compactedFileSet[0], Compacted, levelName(level))
		idx, err := index.GetTopLevelIndex(ctx, s.objC, levelPath)
		if err != nil {
			if !s.objC.IsNotExist(err) {
				return nil, err
			}
		} else {
			spec.Input = append(spec.Input, levelPath)
			size += idx.SizeBytes
		}
		if size <= s.levelSize(level) {
			break
		}
		level++
	}
	// Now we know the output level
	spec.Output = path.Join(fileSet, Compacted, levelName(level))
	// Copy the other levels that may exist
	if err := s.objC.Walk(ctx, path.Join(compactedFileSet[0], Compacted), func(src string) error {
		lName := path.Base(src)
		l, err := parseLevel(lName)
		if err != nil {
			return err
		}
		if l > level {
			dst := path.Join(fileSet, Compacted, levelName(l))
			if err := copyObject(ctx, s.objC, src, dst); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	// Inputs should be ordered with priority from least to greatest.
	for i := 0; i < len(spec.Input)/2; i++ {
		spec.Input[i], spec.Input[len(spec.Input)-1-i] = spec.Input[len(spec.Input)-1-i], spec.Input[i]
	}
	return spec, nil
}

// Delete deletes a fileset.
func (s *Storage) Delete(ctx context.Context, fileSet string) error {
	fileSet = applyPrefix(fileSet)
	return s.objC.Walk(ctx, fileSet, func(name string) error {
		if err := s.objC.Delete(ctx, name); err != nil {
			return err
		}
		return s.chunks.DeleteSemanticReference(ctx, name)
	})
}

// WalkFileSet calls f with the path of every primitive fileSet under prefix.
func (s *Storage) WalkFileSet(ctx context.Context, prefix string, f func(string) error) error {
	return s.objC.Walk(ctx, applyPrefix(prefix), func(p string) error {
		return f(removePrefix(p))
	})
}

// SetTTL sets the time-to-live for the path p.
// if no fileset is found SetTTL returns ErrNoFileSetFound
func (s *Storage) SetTTL(ctx context.Context, p string, ttl time.Duration) (time.Time, error) {
	p = applyPrefix(p)
	if !s.objC.Exists(ctx, p) {
		return time.Time{}, ErrNoFileSetFound
	}
	return s.chunks.RenewReference(ctx, p, ttl)
}

// WithRenewal ensures that none of the filesets under prefix expire by repeatedly calling SetTTL in a separate goroutine
// until cb returns.
// RenewDuring attempts to delete the contents of prefix after cb has returned
func (s *Storage) WithRenewal(ctx context.Context, ttl time.Duration, ps []string, cb func(context.Context) error) error {
	// First we renew all of them, to fail without racing cb
	for _, p := range ps {
		_, err := s.SetTTL(ctx, p, ttl)
		if err != nil {
			return err
		}
	}
	// spawn the SetTTL loops
	eg, ctx2 := errgroup.WithContext(ctx)
	ctx2, cancel := context.WithCancel(ctx2)
	for _, p := range ps {
		p := p
		eg.Go(func() error {
			ticker := time.NewTicker(ttl / 2)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					_, err := s.SetTTL(ctx, p, ttl)
					if err != nil {
						return err
					}
				case <-ctx2.Done():
					return nil
				}
			}
		})
	}
	eg.Go(func() error {
		defer cancel()
		return cb(ctx2)
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	// Ensure that all the keys exist after cb has closed. If they exist now they must have existed for the duration of cb.
	for _, p := range ps {
		if !s.Exists(ctx, p) {
			return ErrNoFileSetFound
		}
	}
	// We delete to ensure no one tries to race for this data after the function ends.
	for _, p := range ps {
		if err := s.Delete(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) Exists(ctx context.Context, p string) (exists bool) {
	if err := s.WalkFileSet(ctx, p, func(p string) error {
		exists = true
		return nil
	}); err != nil {
		return false
	}
	return exists
}

func (s *Storage) GetExpiresAt(ctx context.Context, prefix string) (time.Time, error) {
	var expiresAt *time.Time
	if err := s.WalkFileSet(ctx, prefix, func(p string) error {
		t, err := s.getExpiresAt(ctx, p)
		if err != nil {
			if err != ErrNoTTLSet {
				return nil
			}
			return err
		}
		if expiresAt != nil && t.Before(*expiresAt) {
			expiresAt = &t
		}
		return nil
	}); err != nil {
		return time.Time{}, err
	}
	if expiresAt == nil {
		return time.Time{}, ErrNoTTLSet
	}
	return *expiresAt, nil
}

func (s *Storage) getExpiresAt(ctx context.Context, p string) (time.Time, error) {
	panic("not implemented")
}

func (s *Storage) levelSize(i int) int64 {
	return s.levelZeroSize * int64(math.Pow(float64(s.levelSizeBase), float64(i)))
}

func applyPrefix(fileSet string) string {
	fileSet = strings.TrimLeft(fileSet, "/")
	if strings.HasPrefix(fileSet, semanticPrefix) {
		log.Warn("may be double applying prefix in storage layer", fileSet)
	}
	return path.Join(semanticPrefix, fileSet)
}

func applyPrefixes(fileSets []string) []string {
	var prefixedFileSets []string
	for _, fileSet := range fileSets {
		prefixedFileSets = append(prefixedFileSets, applyPrefix(fileSet))
	}
	return prefixedFileSets
}

func removePrefix(fileSet string) string {
	if !strings.HasPrefix(fileSet, semanticPrefix) {
		panic(fileSet + " does not have prefix " + semanticPrefix)
	}
	return fileSet[len(semanticPrefix):]
}

func removePrefixes(xs []string) []string {
	ys := make([]string, len(xs))
	for i := range xs {
		ys[i] = removePrefix(xs[i])
	}
	return ys
}

const subFileSetFmt = "%020d"
const levelFmt = "level_" + subFileSetFmt

// SubFileSetStr returns the string representation of a subfileset.
func SubFileSetStr(subFileSet int64) string {
	return fmt.Sprintf(subFileSetFmt, subFileSet)
}

func levelName(i int) string {
	return fmt.Sprintf(levelFmt, i)
}

func parseLevel(x string) (int, error) {
	var y int
	_, err := fmt.Sscanf(x, levelFmt, &y)
	return y, err
}

func copyObject(ctx context.Context, objC obj.Client, src, dst string) error {
	w, err := objC.Writer(ctx, dst)
	if err != nil {
		return err
	}
	defer w.Close()
	r, err := objC.Reader(ctx, src, 0, 0)
	if err != nil {
		return err
	}
	defer r.Close()
	if _, err := io.Copy(w, r); err != nil {
		return err
	}
	return w.Close()
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
