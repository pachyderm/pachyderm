package fileset

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	units "github.com/docker/go-units"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"golang.org/x/sync/semaphore"
)

const (
	// DefaultMemoryThreshold is the default for the memory threshold that must
	// be met before a file set part is serialized (excluding close).
	DefaultMemoryThreshold = units.GB
	// DefaultCompactionLevelFactor is the default factor that level sizes increase by in a compacted fileset.
	DefaultCompactionLevelFactor = 10
	DefaultPrefetchLimit         = 10
	DefaultBatchThreshold        = units.MB
	// DefaultIndexCacheSize is the default size of the index cache.
	DefaultIndexCacheSize = 100

	// TrackerPrefix is used for creating tracker objects for filesets
	TrackerPrefix = "fileset/"

	// DefaultFileDatum is the default file datum.
	DefaultFileDatum = "default"
)

var (
	// ErrNoFileSetFound is returned by the methods on Storage when a fileset does not exist
	ErrNoFileSetFound = errors.Errorf("no fileset found")
)

// Storage is an abstraction for interfacing with file sets.
// A storage instance:
// - Provides methods for writing file sets, opening file sets for reading, and managing file sets.
// - Manages tracker state to keep internal file sets alive while writing file sets.
// - Manages an internal index cache that supports logarithmic lookup in the multilevel indexes.
// - Provides methods for processing file set compaction tasks.
// - Provides a method for creating a garbage collector.
type Storage struct {
	tracker          track.Tracker
	store            MetadataStore
	chunks           *chunk.Storage
	idxCache         *index.Cache
	memThreshold     int64
	shardConfig      *index.ShardConfig
	compactionConfig *CompactionConfig
	filesetSem       *semaphore.Weighted
	prefetchLimit    int
}

type CompactionConfig struct {
	LevelFactor int64
}

// NewStorage creates a new Storage.
func NewStorage(mds MetadataStore, tr track.Tracker, chunks *chunk.Storage, opts ...StorageOption) *Storage {
	s := &Storage{
		store:        mds,
		tracker:      tr,
		chunks:       chunks,
		idxCache:     index.NewCache(chunks, DefaultIndexCacheSize),
		memThreshold: DefaultMemoryThreshold,
		shardConfig: &index.ShardConfig{
			NumFiles:  index.DefaultShardNumThreshold,
			SizeBytes: index.DefaultShardSizeThreshold,
		},
		compactionConfig: &CompactionConfig{
			LevelFactor: DefaultCompactionLevelFactor,
		},
		filesetSem:    semaphore.NewWeighted(math.MaxInt64),
		prefetchLimit: DefaultPrefetchLimit,
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.compactionConfig.LevelFactor < 1 {
		panic("level factor cannot be < 1")
	}
	return s
}

func (s *Storage) ShardConfig() *index.ShardConfig {
	return s.shardConfig
}

// NewUnorderedWriter creates a new unordered file set writer.
func (s *Storage) NewUnorderedWriter(ctx context.Context, opts ...UnorderedWriterOption) (*UnorderedWriter, error) {
	return newUnorderedWriter(ctx, s, s.memThreshold, s.shardConfig.NumFiles/2, opts...)
}

// NewWriter creates a new file set writer.
func (s *Storage) NewWriter(ctx context.Context, opts ...WriterOption) *Writer {
	return s.newWriter(ctx, opts...)
}

func (s *Storage) newWriter(ctx context.Context, opts ...WriterOption) *Writer {
	ctx = pctx.Child(ctx, "fileSetWriter")
	return newWriter(ctx, s, opts...)
}

func (s *Storage) newReader(handle *Handle) *Reader {
	return newReader(s.store, s.chunks, s.idxCache, handle)
}

// Open opens a file set for reading.
func (s *Storage) Open(ctx context.Context, handles []*Handle) (FileSet, error) {
	var err error
	handles, err = s.FlattenAll(ctx, handles)
	if err != nil {
		return nil, err
	}
	var fss []FileSet
	for _, handle := range handles {
		fss = append(fss, s.newReader(handle))
	}
	if len(fss) == 0 {
		return emptyFileSet{}, nil
	}
	if len(fss) == 1 {
		return fss[0], nil
	}
	return newMergeReader(s.chunks, fss), nil
}

// Compose produces a composite fileset from the filesets under ids.
// It does not perform a merge or check that the filesets at ids in any way
// other than ensuring that they exist.
func (s *Storage) Compose(ctx context.Context, handles []*Handle, ttl time.Duration) (*Handle, error) {
	var result *Handle
	if err := dbutil.WithTx(ctx, s.store.DB(), func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		result, err = s.ComposeTx(tx, handles, ttl)
		return err
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// ComposeTx produces a composite fileset from the filesets under ids.
// It does not perform a merge or check that the filesets at ids in any way
// other than ensuring that they exist.
func (s *Storage) ComposeTx(tx *pachsql.Tx, handles []*Handle, ttl time.Duration) (*Handle, error) {
	c := &Composite{
		Layers: handlesToTokenHexStrings(handles),
	}
	return s.newCompositeTx(tx, c, ttl)
}

// Clone creates a new fileset, identical to the fileset at id, but with the specified ttl.
// The ttl can be ignored by using track.NoTTL
func (s *Storage) Clone(ctx context.Context, handle *Handle, ttl time.Duration) (*Handle, error) {
	var result *Handle
	if err := dbutil.WithTx(ctx, s.store.DB(), func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		result, err = s.CloneTx(tx, handle, ttl)
		return err
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// CloneTx creates a new fileset, identical to the fileset at id, but with the specified ttl.
// The ttl can be ignored by using track.NoTTL
func (s *Storage) CloneTx(tx *pachsql.Tx, handle *Handle, ttl time.Duration) (*Handle, error) {
	md, err := s.store.GetTx(tx, handle.token)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	switch x := md.Value.(type) {
	case *Metadata_Primitive:
		return s.newPrimitiveTx(tx, x.Primitive, ttl)
	case *Metadata_Composite:
		return s.newCompositeTx(tx, x.Composite, ttl)
	default:
		return nil, errors.Errorf("cannot clone type %T", md.Value)
	}
}

// Flatten iterates through IDs and replaces references to composite file sets
// with all their layers in place and executes the user provided callback
// against each primitive file set.
func (s *Storage) Flatten(ctx context.Context, handles []*Handle, cb func(handle *Handle) error) error {
	for _, handle := range handles {
		md, err := s.store.Get(ctx, handle.token)
		if err != nil {
			return err
		}
		switch x := md.Value.(type) {
		case *Metadata_Primitive:
			if err := cb(handle); err != nil {
				if errors.Is(err, errutil.ErrBreak) {
					return nil
				}
				return err
			}
		case *Metadata_Composite:
			handles, err := x.Composite.PointsTo()
			if err != nil {
				return err
			}
			if err := s.Flatten(ctx, handles, cb); err != nil {
				if errors.Is(err, errutil.ErrBreak) {
					return nil
				}
				return err
			}
		default:
			// TODO: should it be?
			return errors.Errorf("Flatten is not defined for empty filesets")
		}
	}
	return nil
}

// FlattenAll is like Flatten, but collects the primitives to return to the user.
func (s *Storage) FlattenAll(ctx context.Context, handles []*Handle) ([]*Handle, error) {
	flattened := make([]*Handle, 0, len(handles))
	if err := s.Flatten(ctx, handles, func(handle *Handle) error {
		flattened = append(flattened, handle)
		return nil
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return flattened, nil
}

func (s *Storage) flattenPrimitives(ctx context.Context, handles []*Handle) ([]*Primitive, error) {
	handles, err := s.FlattenAll(ctx, handles)
	if err != nil {
		return nil, err
	}
	return s.getPrimitives(ctx, handles)
}

func (s *Storage) getPrimitives(ctx context.Context, handles []*Handle) ([]*Primitive, error) {
	var prims []*Primitive
	for _, handle := range handles {
		prim, err := s.getPrimitive(ctx, handle)
		if err != nil {
			return nil, err
		}
		prims = append(prims, prim)
	}
	return prims, nil
}

// Concat is a special case of Merge, where the filesets each contain paths for distinct ranges.
// The path ranges must be non-overlapping and the ranges must be lexigraphically sorted.
// Concat always returns the ID of a primitive fileset.
func (s *Storage) Concat(ctx context.Context, handles []*Handle, ttl time.Duration) (*Handle, error) {
	var size int64
	additive := index.NewWriter(ctx, s.chunks, "additive-index-writer")
	deletive := index.NewWriter(ctx, s.chunks, "deletive-index-writer")
	for _, handle := range handles {
		md, err := s.store.Get(ctx, handle.token)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		prim := md.GetPrimitive()
		if prim == nil {
			return nil, errors.Errorf("file set %v is not primitive", handle.token)
		}
		if prim.Additive != nil {
			if err := additive.WriteIndex(prim.Additive); err != nil {
				return nil, err
			}
		}
		if prim.Deletive != nil {
			if err := deletive.WriteIndex(prim.Deletive); err != nil {
				return nil, err
			}
		}
		size += prim.SizeBytes
	}
	additiveIdx, err := additive.Close()
	if err != nil {
		return nil, err
	}
	deletiveIdx, err := deletive.Close()
	if err != nil {
		return nil, err
	}
	return s.newPrimitive(ctx, &Primitive{
		Additive:  additiveIdx,
		Deletive:  deletiveIdx,
		SizeBytes: size,
	}, ttl)
}

// Drop allows a fileset to be deleted if it is not otherwise referenced.
func (s *Storage) Drop(ctx context.Context, handle *Handle) error {
	_, err := s.SetTTL(ctx, handle, track.ExpireNow)
	return err
}

// SetTTL sets the time-to-live for the fileset at id
func (s *Storage) SetTTL(ctx context.Context, handle *Handle, ttl time.Duration) (time.Time, error) {
	oid := handle.token.TrackerID()
	res, err := s.tracker.SetTTL(ctx, oid, ttl)
	return res, err
}

// SizeUpperBound returns an upper bound for the size of the data in the file set in bytes.
// The upper bound is cheaper to compute than the actual size.
func (s *Storage) SizeUpperBound(ctx context.Context, handle *Handle) (int64, error) {
	prims, err := s.flattenPrimitives(ctx, []*Handle{handle})
	if err != nil {
		return 0, err
	}
	var total int64
	for _, prim := range prims {
		total += prim.SizeBytes
	}
	return total, nil
}

// Size returns the size of the data in the file set in bytes.
func (s *Storage) Size(ctx context.Context, handle *Handle) (int64, error) {
	fs, err := s.Open(ctx, []*Handle{handle})
	if err != nil {
		return 0, err
	}
	var total int64
	if err := fs.Iterate(ctx, func(f File) error {
		total += index.SizeBytes(f.Index())
		return nil
	}); err != nil {
		return 0, err
	}
	return total, nil
}

// WithRenewer calls cb with a Renewer, and a context which will be canceled if the renewer is unable to renew a path.
func (s *Storage) WithRenewer(ctx context.Context, ttl time.Duration, cb func(context.Context, *Renewer) error) (retErr error) {
	r := newRenewer(ctx, s, ttl)
	defer func() {
		if err := r.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(r.Context(), r)
}

func (s *Storage) NewGC(d time.Duration) *track.GarbageCollector {
	tmpDeleter := renew.NewTmpDeleter()
	chunkDeleter := s.chunks.NewDeleter()
	filesetDeleter := &deleter{
		store: s.store,
	}
	mux := track.DeleterMux(func(id string) track.Deleter {
		switch {
		case strings.HasPrefix(id, renew.TmpTrackerPrefix):
			return tmpDeleter
		case strings.HasPrefix(id, chunk.TrackerPrefix):
			return chunkDeleter
		case strings.HasPrefix(id, TrackerPrefix):
			return filesetDeleter
		default:
			return nil
		}
	})
	return track.NewGarbageCollector(s.tracker, d, mux)
}

func (s *Storage) exists(ctx context.Context, handle *Handle) (bool, error) {
	exists, err := s.store.Exists(ctx, handle.token)
	return exists, err
}

func (s *Storage) newPrimitive(ctx context.Context, prim *Primitive, ttl time.Duration) (*Handle, error) {
	var result *Handle
	if err := dbutil.WithTx(ctx, s.store.DB(), func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		result, err = s.newPrimitiveTx(tx, prim, ttl)
		return err
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Storage) newPrimitiveTx(tx *pachsql.Tx, prim *Primitive, ttl time.Duration) (*Handle, error) {
	token := newToken()
	md := &Metadata{
		Value: &Metadata_Primitive{
			Primitive: prim,
		},
	}
	var pointsTo []string
	for _, chunkID := range prim.PointsTo() {
		pointsTo = append(pointsTo, chunkID.TrackerID())
	}
	if err := s.store.SetTx(tx, token, md); err != nil {
		return nil, err
	}
	if err := s.tracker.CreateTx(tx, token.TrackerID(), pointsTo, ttl); err != nil {
		return nil, err
	}
	id, err := computeId(tx, s.store, md)
	if err != nil {
		return nil, err
	}
	return &Handle{
		token: token,
		id:    id,
	}, nil
}

func (s *Storage) newComposite(ctx context.Context, comp *Composite, ttl time.Duration) (*Handle, error) {
	var result *Handle
	if err := dbutil.WithTx(ctx, s.store.DB(), func(ctx context.Context, tx *pachsql.Tx) error {
		var err error
		result, err = s.newCompositeTx(tx, comp, ttl)
		return err
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Storage) newCompositeTx(tx *pachsql.Tx, comp *Composite, ttl time.Duration) (*Handle, error) {
	token := newToken()
	md := &Metadata{
		Value: &Metadata_Composite{
			Composite: comp,
		},
	}
	handles, err := comp.PointsTo()
	if err != nil {
		return nil, err
	}
	var pointsTo []string
	for _, handle := range handles {
		pointsTo = append(pointsTo, handle.token.TrackerID())
	}
	if err := s.store.SetTx(tx, token, md); err != nil {
		return nil, err
	}
	if err := s.tracker.CreateTx(tx, token.TrackerID(), pointsTo, ttl); err != nil {
		return nil, err
	}
	id, err := computeId(tx, s.store, md)
	if err != nil {
		return nil, err
	}
	return &Handle{
		token: token,
		id:    id,
	}, nil
}

func (s *Storage) getPrimitive(ctx context.Context, handle *Handle) (*Primitive, error) {
	md, err := s.store.Get(ctx, handle.token)
	if err != nil {
		return nil, err
	}
	prim := md.GetPrimitive()
	if prim == nil {
		return nil, errors.Errorf("fileset %v is not primitive", handle.token)
	}
	return prim, nil
}

var _ track.Deleter = &deleter{}

type deleter struct {
	store MetadataStore
}

func (d *deleter) DeleteTx(tx *pachsql.Tx, oid string) error {
	if !strings.HasPrefix(oid, TrackerPrefix) {
		return errors.Errorf("don't know how to delete %v", oid)
	}
	token, err := parseToken([]byte(oid[len(TrackerPrefix):]))
	if err != nil {
		return err
	}
	return d.store.DeleteTx(tx, token)
}

type PinnedFileset Token

// Pin clones a fileset, keeping it alive forever.
/* 	TODO(Fahad): Replace cloning with a Pin that is a big int.
	A pin will point to a fileset ID, where the ID is a stable hash of the root index.
   	Fileset trees must be convergent in order to achieve this. */
func (s *Storage) Pin(tx *pachsql.Tx, handle *Handle) (PinnedFileset, error) {
	handle, err := s.CloneTx(tx, handle, track.NoTTL)
	if err != nil {
		return PinnedFileset{}, errors.Wrap(err, "pin")
	}
	return PinnedFileset(handle.token), nil
}

type ChunkSetID uint64

func (s *Storage) CreateChunkSet(ctx context.Context, tx *sqlx.Tx) (ChunkSetID, error) {
	ctx = pctx.Child(ctx, "createChunkset")
	// Insert ChunkSet into ChunkSet table.
	var chunksetID ChunkSetID
	if err := tx.GetContext(ctx, &chunksetID, `INSERT INTO storage.chunksets DEFAULT VALUES RETURNING id`); err != nil {
		return 0, errors.Wrapf(err, "get chunk set id")
	}

	// encode chunkset into string for tracker
	chunksetStrID := chunksetStringID(chunksetID)

	// List all of the filesets.
	var pointsTo []string
	if err := tx.SelectContext(ctx, &pointsTo, `SELECT str_id FROM storage.tracker_objects WHERE str_id LIKE 'fileset/%'`); err != nil {
		return 0, errors.Wrap(err, "get filesets from db")
	}
	if err := s.tracker.CreateTx(tx, chunksetStrID, pointsTo, track.NoTTL); err != nil {
		return 0, errors.Wrap(err, "create tracker object and references")
	}

	return chunksetID, nil
}

func (s *Storage) DropChunkSet(ctx context.Context, tx *sqlx.Tx, id ChunkSetID) error {
	result, err := tx.Exec("DELETE FROM storage.chunksets WHERE id = $1", id)
	if err != nil {
		return errors.Wrap(err, "delete chunkset in db")
	}
	// Check the number of affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "rows effected")
	}
	if rowsAffected == 0 {
		return errors.Errorf("no chunkset found with the given id: %d", id)
	}
	// encode chunkset into string for tracker
	strID := chunksetStringID(id)
	return errors.Wrap(s.tracker.DeleteTx(tx, strID), "delete tracker object and references")
}

func chunksetStringID(id ChunkSetID) string {
	return "chunkset/" + strconv.FormatUint(uint64(id), 10)
}
