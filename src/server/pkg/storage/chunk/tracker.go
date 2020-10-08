package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	io "io"
	"path"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

type ChunkID []byte

func ChunkIDFromHex(h string) (ChunkID, error) {
	return hex.DecodeString(h)
}

func (id ChunkID) HexString() string {
	return hex.EncodeString(id)
}

// Tracker tracks chunk metadata and keeps track of chunk sets being uploaded or downloaded.
type Tracker interface {
	// ChunkSet methods. Keep track of active chunks

	NewChunkSet(ctx context.Context, chunkSet string, delete bool, ttl time.Duration) error
	// SetTTL sets the TTL on a chunkSet
	SetTTL(ctx context.Context, chunkSet string, ttl time.Duration) (expiresAt time.Time, err error)
	// LockChunk adds a chunk id to a chunk set.
	LockChunk(ctx context.Context, chunkSet string, chunkID ChunkID) error
	// DropChunkSet causes all chunks in a chunk set to be deleted. It returns the number of chunks dropped.
	DropChunkSet(ctx context.Context, chunkSet string) (int, error)

	// Chunk metadata

	// SetChunkInfo adds chunk metadata to the tracker
	SetChunkInfo(ctx context.Context, chunkID ChunkID, chunkInfo *ChunkInfo, pointsTo []ChunkID) error
	// DeleteChunkInfo removes chunk metadata from the tracker
	DeleteChunkInfo(ctx context.Context, chunkID ChunkID) error
}

func UploadChunk(ctx context.Context, tracker Tracker, objc obj.Client, cinfo *ChunkInfo, r io.Reader) error {
	if err := tracker.PostChunkInfo(ctx, chunkInfo); err != nil {
		return err
	}
	// Skip the upload if the chunk already exists.
	if w.objC.Exists(w.ctx, path) {
		return nil
	}
	p := path.Join(prefix, cinfo.Hash)
	objW, err := objC.Writer(w.ctx, path)
	if err != nil {
		return err
	}
	defer objW.Close()
	gzipW, err := gzip.NewWriterLevel(objW, gzip.BestSpeed)
	if err != nil {
		return err
	}
	defer gzipW.Close()
	// TODO Encryption?
	_, err = io.Copy(gzipW, bytes.NewReader(chunkBytes))
	return err
}

func GetChunk(ctx context.Context, _ Tracker, objc obj.Client, chunkID ChunkID, w io.Writer) (io.ReadCloser, error) {
	objR, err := dr.objC.Reader(dr.ctx, path.Join(prefix, string(chunkID)), 0, 0)
	if err != nil {
		return err
	}
	defer objR.Close()
	gzipR, err := gzip.NewReader(objR)
	if err != nil {
		return err
	}
	defer gzipR.Close()
	return io.Copy(r, gzipR)
}

func DeleteChunk(ctx context.Context, tracker Tracker, objc obj.Client, chunkID ChunkID) error {
	// TODO: fix the race where a chunk is uploaded after it is deleted,
	if err := tracker.DeleteChunkInfo(ctx, chunkID); err != nil {
		return err
	}

	return objc.Delete(ctx, chunkPath(chunkID))
}

func chunkPath(chunkID ChunkID) string {
	return path.Join(prefix, chunkID.HexString())
}

type PGTracker struct {
	db *sqlx.DB
}

func NewPGTracker(db *sqlx.DB) *PGTracker {
	return &PGTracker{db: db}
}

func (tr *PGTracker) NewChunkSet(ctx context.Context, chunkSet string, delete bool, ttl time.Duration) (time.Time, error) {
	panic("not implemented")
}

func (tr *PGTracker) SetTTL(ctx context.Context, chunkSet string, ttl time.Duration) (time.Time, error) {
	panic("not implemented")
}

const schema = `
	CREATE SCHEMA IF NOT EXISTS storage;

	CREATE TABLE storage.chunks (
		int_id BIGSERIAL PRIMARY KEY,
		hash_id BYTEA NOT NULL UNIQUE,
		size INT8 NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE storage.chunk_refs (
		from INT8 NOT NULL,
		to INT8 NOT NULL,
		PRIMARY KEY(from, to)
	);

	CREATE TABLE storage.chunk_locks (
		chunkset_id VARCHAR(64) NOT NULL PRIMARY KEY
		chunk_id INT8 NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL
	);
`
