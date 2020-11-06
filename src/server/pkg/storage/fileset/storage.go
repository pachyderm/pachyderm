package fileset

import (
	"context"
	"fmt"
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
	// ErrNoFileSetFound is returned by the methods on Storage when a fileset does not exist
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

// NewUnorderedWriter creates a new unordered file set writer.
func (s *Storage) NewUnorderedWriter(ctx context.Context, fileSet, defaultTag string, opts ...UnorderedWriterOption) (*UnorderedWriter, error) {
	fileSet = applyPrefix(fileSet)
	return newUnorderedWriter(ctx, s, fileSet, s.memThreshold, defaultTag, opts...)
}

// NewWriter creates a new file set writer.
func (s *Storage) NewWriter(ctx context.Context, fileSet string, opts ...WriterOption) *Writer {
	fileSet = applyPrefix(fileSet)
	return s.newWriter(ctx, fileSet, opts...)
}

func (s *Storage) newWriter(ctx context.Context, fileSet string, opts ...WriterOption) *Writer {
	return newWriter(ctx, s.objC, s.chunks, fileSet, opts...)
}

// TODO: Expose some notion of read ahead (read a certain number of chunks in parallel).
// this will be necessary to speed up reading large files.
func (s *Storage) newReader(fileSet string, opts ...index.Option) *Reader {
	return newReader(s.objC, s.chunks, fileSet, opts...)
}

// Open opens a file set for reading.
// TODO: It might make sense to have some of the file set transforms as functional options here.
func (s *Storage) Open(ctx context.Context, fileSets []string, opts ...index.Option) (FileSet, error) {
	fileSets = applyPrefixes(fileSets)
	fs, err := s.open(ctx, fileSets, opts...)
	if err != nil {
		return nil, err
	}
	return NewDeleteFilter(fs), nil
}

func (s *Storage) open(ctx context.Context, fileSets []string, opts ...index.Option) (FileSet, error) {
	var fss []FileSet
	for _, fileSet := range fileSets {
		if err := s.objC.Walk(ctx, fileSet, func(name string) error {
			fss = append(fss, s.newReader(name, opts...))
			return nil
		}); err != nil {
			return nil, err
		}
	}
	// TODO: Error if no filesets found?
	if len(fss) == 1 {
		return fss[0], nil
	}
	return newMergeReader(s.chunks, fss), nil
}

// OpenWithDeletes opens a file set for reading and does not filter out deleted files.
// TODO: This should be a functional option on New.
func (s *Storage) OpenWithDeletes(ctx context.Context, fileSets []string, opts ...index.Option) (FileSet, error) {
	fileSets = applyPrefixes(fileSets)
	fs, err := s.open(ctx, fileSets, opts...)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

// Shard shards the file set into path ranges.
// TODO This should be extended to be more configurable (different criteria
// for creating shards).
func (s *Storage) Shard(ctx context.Context, fs FileSet, cb ShardCallback) error {
	return shard(ctx, fs, s.shardThreshold, cb)
}

// ShardCallback is a callback that returns a path range for each shard.
type ShardCallback func(*index.PathRange) error

// shard creates shards (path ranges) from the file set streams being merged.
// A shard is created when the size of the content for a path range is greater than
// the passed in shard threshold.
// For each shard, the callback is called with the path range for the shard.
func shard(ctx context.Context, fs FileSet, shardThreshold int64, cb ShardCallback) error {
	var size int64
	pathRange := &index.PathRange{}
	if err := fs.Iterate(ctx, func(f File) error {
		// A shard is created when we have encountered more than shardThreshold content bytes.
		if size >= shardThreshold {
			pathRange.Upper = f.Index().Path
			if err := cb(pathRange); err != nil {
				return err
			}
			size = 0
			pathRange = &index.PathRange{
				Lower: f.Index().Path,
			}
		}
		size += f.Index().SizeBytes
		return nil
	}); err != nil {
		return err
	}
	return cb(pathRange)
}

// Copy copies the fileset at srcPrefix to dstPrefix. It does *not* perform compaction
// ttl sets the time to live on the keys under dstPrefix if ttl == 0, it is ignored
func (s *Storage) Copy(ctx context.Context, srcPrefix, dstPrefix string, ttl time.Duration) error {
	srcPrefix = applyPrefix(srcPrefix)
	dstPrefix = applyPrefix(dstPrefix)
	// TODO: perform this atomically with postgres
	return s.objC.Walk(ctx, srcPrefix, func(srcPath string) error {
		dstPath := dstPrefix + srcPath[len(srcPrefix):]
		if err := obj.Copy(ctx, s.objC, s.objC, srcPath, dstPath); err != nil {
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
	fs, err := s.open(ctx, inputFileSets)
	if err != nil {
		return nil, err
	}
	if err := CopyFiles(ctx, w, fs); err != nil {
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
			if err := obj.Copy(ctx, s.objC, s.objC, src, dst); err != nil {
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

// WithRenewer calls cb with a Renewer, and a context which will be canceled if the renewer is unable to renew a path.
func (s *Storage) WithRenewer(ctx context.Context, ttl time.Duration, cb func(context.Context, *Renewer) error) error {
	renew := func(ctx context.Context, p string, ttl time.Duration) error {
		_, err := s.SetTTL(ctx, p, ttl)
		return err
	}
	return WithRenewer(ctx, ttl, renew, cb)
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
