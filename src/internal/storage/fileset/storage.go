package fileset

import (
	"context"
	"fmt"
	"math"
	"path"
	"strings"
	"time"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/internal/errors"
	"github.com/pachyderm/pachyderm/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/src/internal/storage/track"
	"golang.org/x/sync/semaphore"
)

const (
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
	// TrackerPrefix is used for creating tracker objects for filesets
	TrackerPrefix = "fileset/"
)

var (
	// ErrNoFileSetFound is returned by the methods on Storage when a fileset does not exist
	ErrNoFileSetFound = errors.Errorf("no fileset found")
)

// Storage is the abstraction that manages fileset storage.
type Storage struct {
	tracker                      track.Tracker
	store                        Store
	chunks                       *chunk.Storage
	memThreshold, shardThreshold int64
	levelZeroSize                int64
	levelSizeBase                int
	filesetSem                   *semaphore.Weighted
}

// NewStorage creates a new Storage.
func NewStorage(store Store, tr track.Tracker, chunks *chunk.Storage, opts ...StorageOption) *Storage {
	s := &Storage{
		store:          store,
		tracker:        tr,
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

// Store returns the underlying store.
// TODO Store is just used to poke through the information about file set sizes.
// I think there might be a cleaner way to handle this through the file set interface, and changing
// the metadata we expose for a file set as a set of metadata entries.
func (s *Storage) Store() Store {
	return s.store
}

// ChunkStorage returns the underlying chunk storage instance for this storage instance.
func (s *Storage) ChunkStorage() *chunk.Storage {
	return s.chunks
}

// NewUnorderedWriter creates a new unordered file set writer.
func (s *Storage) NewUnorderedWriter(ctx context.Context, fileSet, defaultTag string, opts ...UnorderedWriterOption) (*UnorderedWriter, error) {
	return newUnorderedWriter(ctx, s, fileSet, s.memThreshold, defaultTag, opts...)
}

// NewWriter creates a new file set writer.
func (s *Storage) NewWriter(ctx context.Context, fileSet string, opts ...WriterOption) *Writer {
	return s.newWriter(ctx, fileSet, opts...)
}

func (s *Storage) newWriter(ctx context.Context, fileSet string, opts ...WriterOption) *Writer {
	return newWriter(ctx, s.store, s.tracker, s.chunks, fileSet, opts...)
}

// TODO: Expose some notion of read ahead (read a certain number of chunks in parallel).
// this will be necessary to speed up reading large files.
func (s *Storage) newReader(fileSet string, opts ...index.Option) *Reader {
	return newReader(s.store, s.chunks, fileSet, opts...)
}

// Open opens a file set for reading.
// TODO: It might make sense to have some of the file set transforms as functional options here.
func (s *Storage) Open(ctx context.Context, fileSets []string, opts ...index.Option) (FileSet, error) {
	var fss []FileSet
	for _, fileSet := range fileSets {
		if err := s.store.Walk(ctx, fileSet, func(p string) error {
			fss = append(fss, s.newReader(p, opts...))
			return nil
		}); err != nil {
			return nil, err
		}
	}
	if len(fss) == 0 {
		return nil, errors.Errorf("error opening fileset: non-existent fileset: %v", fileSets)
	}
	if len(fss) == 1 {
		return fss[0], nil
	}
	return newMergeReader(s.chunks, fss), nil
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
		size += index.SizeBytes(f.Index())
		return nil
	}); err != nil {
		return err
	}
	return cb(pathRange)
}

// Copy copies the fileset at srcPrefix to dstPrefix. It does *not* perform compaction
// ttl sets the time to live on the keys under dstPrefix if ttl == 0, it is ignored
func (s *Storage) Copy(ctx context.Context, srcPrefix, dstPrefix string, ttl time.Duration) error {
	// TODO: perform this atomically with postgres
	return s.store.Walk(ctx, srcPrefix, func(srcPath string) error {
		dstPath := dstPrefix + srcPath[len(srcPrefix):]
		return copyPath(ctx, s.store, s.store, srcPath, dstPath, s.tracker, ttl)
	})
}

// CompactStats contains information about what was compacted.
type CompactStats struct {
	OutputSize int64
}

// Compact compacts a set of filesets into an output fileset.
func (s *Storage) Compact(ctx context.Context, outputFileSet string, inputFileSets []string, ttl time.Duration, opts ...index.Option) (*CompactStats, error) {
	var size int64
	w := s.newWriter(ctx, outputFileSet, WithTTL(ttl), WithIndexCallback(func(idx *index.Index) error {
		size += index.SizeBytes(idx)
		return nil
	}))
	fs, err := s.Open(ctx, inputFileSets)
	if err != nil {
		return nil, err
	}
	if err := CopyFiles(ctx, w, fs, true); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
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
	spec, err := s.compactSpec(ctx, fileSet, compactedFileSet...)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func (s *Storage) compactSpec(ctx context.Context, fileSet string, compactedFileSet ...string) (ret *CompactSpec, retErr error) {
	md, err := s.store.Get(ctx, path.Join(fileSet, Diff))
	if err != nil {
		return nil, err
	}
	size := md.SizeBytes
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
		md, err := s.store.Get(ctx, levelPath)
		if err != nil {
			if err != ErrPathNotExists {
				return nil, err
			}
		} else {
			spec.Input = append(spec.Input, levelPath)
			size += md.SizeBytes
		}
		if size <= s.levelSize(level) {
			break
		}
		level++
	}
	// Now we know the output level
	spec.Output = path.Join(fileSet, Compacted, levelName(level))
	// Copy the other levels that may exist
	if err := s.store.Walk(ctx, path.Join(compactedFileSet[0], Compacted), func(src string) error {
		lName := path.Base(src)
		l, err := parseLevel(lName)
		if err != nil {
			return err
		}
		if l > level {
			dst := path.Join(fileSet, Compacted, levelName(l))
			if err := copyPath(ctx, s.store, s.store, src, dst, s.tracker, 0); err != nil {
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
	return s.store.Walk(ctx, fileSet, func(name string) error {
		oid := filesetObjectID(name)
		if err := s.store.Delete(ctx, name); err != nil {
			return err
		}
		return s.tracker.MarkTombstone(ctx, oid)
	})
}

// SetTTL sets the time-to-live for the prefix p.
func (s *Storage) SetTTL(ctx context.Context, p string, ttl time.Duration) (time.Time, error) {
	oid := filesetObjectID(p)
	return s.tracker.SetTTLPrefix(ctx, oid, ttl)
}

// WithRenewer calls cb with a Renewer, and a context which will be canceled if the renewer is unable to renew a path.
func (s *Storage) WithRenewer(ctx context.Context, ttl time.Duration, cb func(context.Context, *renew.StringSet) error) error {
	rf := func(ctx context.Context, p string, ttl time.Duration) error {
		_, err := s.SetTTL(ctx, p, ttl)
		return err
	}
	return renew.WithStringSet(ctx, ttl, rf, cb)
}

// GC creates a track.GarbageCollector with a Deleter that can handle deleting filesets and chunks
func (s *Storage) GC(ctx context.Context) error {
	const period = 10 * time.Second
	tmpDeleter := track.NewTmpDeleter()
	chunkDeleter := s.chunks.NewDeleter()
	filesetDeleter := &deleter{
		store: s.store,
	}
	mux := track.DeleterMux(func(id string) track.Deleter {
		switch {
		case strings.HasPrefix(id, track.TmpTrackerPrefix):
			return tmpDeleter
		case strings.HasPrefix(id, chunk.TrackerPrefix):
			return chunkDeleter
		case strings.HasPrefix(id, TrackerPrefix):
			return filesetDeleter
		default:
			return nil
		}
	})
	gc := track.NewGarbageCollector(s.tracker, period, mux)
	return gc.Run(ctx)
}

func (s *Storage) levelSize(i int) int64 {
	return s.levelZeroSize * int64(math.Pow(float64(s.levelSizeBase), float64(i)))
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

func filesetObjectID(p string) string {
	return "fileset/" + p
}

var _ track.Deleter = &deleter{}

type deleter struct {
	store Store
}

// TODO: This needs to be implemented, temporary filesets are still in Postgres.
func (d *deleter) Delete(ctx context.Context, id string) error {
	return nil
}
