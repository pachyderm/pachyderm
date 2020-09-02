package gc

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

const (
	flushingDeletion = "flushing deletion"
)

var (
	errFlush = errors.Errorf("flushing")
)

type SourceType string

const (
	STTemporary = SourceType("temporary")
	STChunk     = SourceType("chunk")
	STSemantic  = SourceType("semantic")
)

// Reference describes a reference to a chunk in object storage.  If a chunk has
// no references, it will be deleted.
//  * Sourcetype - the type of reference, one of:
//   * 'temporary' - a temporary reference to a chunk.
//   * 'chunk' - a cross-chunk reference, from one chunk to another.
//   * 'semantic' - a reference to a chunk by some semantic name.
//  * Source - the source of the reference, this may be a temporary id, chunk id, or a
//    semantic name.
//  * Chunk - the target chunk being referenced.
//  * ExpiresAt - the time after which a temporary chunk will be destroyed
type Reference struct {
	Sourcetype SourceType
	Source     string
	Chunk      string
	ExpiresAt  *time.Time
}

// Client is the interface provided by the garbage collector client, for use on
// worker nodes.  It will directly perform reference-counting operations on the
// cluster's Postgres database, and block on deleting chunks.
type Client interface {
	// ReserveChunk ensures that a chunk is not deleted by the garbage collector.
	// It will add a temporary reference to the given chunk, even
	// if it doesn't exist yet.  If the specified chunk is currently
	// being deleted, this call will block while it is being deleted.
	ReserveChunk(ctx context.Context, chunk string, tmpID string, expiresAt time.Time) error

	// CreateReference creates a reference.
	CreateReference(context.Context, *Reference) error
	// DeleteReference deletes a reference.
	DeleteReference(context.Context, *Reference) error
	// RenewReference will set expiresAt
	// The SourceType will be assumed to be 'temporary'
	RenewReference(ctx context.Context, src string, expiresAt time.Time) error
}

type client struct {
	db *gorm.DB
}

// NewClient creates a new client.
func NewClient(db *gorm.DB) (Client, error) {
	return &client{db: db}, nil
}

func (c *client) ReserveChunk(ctx context.Context, chunk, tmpID string, expiresAt time.Time) error {
	var flushChunk []chunkModel
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			return txn.Exec("SET LOCAL synchronous_commit = off;")
		},
		func(txn *gorm.DB) *gorm.DB {
			// Insert the chunk to reserve and add a temporary reference to it.
			// If the chunk already exists, then just add a temporary reference.
			// Return the chunk if it is being deleted (flush is necessary).
			return txn.Raw(`
				WITH added_chunk AS (
					INSERT INTO chunks (chunk)
					VALUES (?)
					ON CONFLICT (chunk) DO UPDATE SET chunk = EXCLUDED.chunk
					RETURNING chunk, deleting
				), added_ref AS (
					INSERT INTO refs (sourcetype, source, chunk, created, expires_at)
					SELECT 'temporary'::reftype, ?, chunk, NOW(), ?
					FROM added_chunk
					WHERE deleting IS NULL
				)
				SELECT chunk
				FROM added_chunk
				WHERE deleting IS NOT NULL
			`, chunk, tmpID, expiresAt).Scan(&flushChunk)
		},
	}
	return retry(ctx, flushingDeletion, func() error {
		if err := runTransaction(ctx, c.db, stmtFuncs); err != nil {
			return err
		}
		if len(flushChunk) > 0 {
			return errFlush
		}
		return nil
	})
}

func (c *client) CreateReference(ctx context.Context, ref *Reference) (retErr error) {
	if ref.Sourcetype == "temporary" && ref.ExpiresAt == nil {
		t := time.Now().Add(defaultTimeout)
		ref.ExpiresAt = &t
	}
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			// Insert the reference.
			// Does nothing if the reference already exists.
			// TODO We should consider the possibility of a very slow client
			// not adding a reference before the temporary reference times out.
			// We could do some reachability validation, but we would probably want
			// to do this when the semantic reference is setup to minimize the cost on
			// the common path.
			return txn.Exec(`
				INSERT INTO refs (sourcetype, source, chunk, created, expires_at)
				VALUES (?, ?, ?, NOW(), ?)
				ON CONFLICT DO NOTHING
			`, ref.Sourcetype, ref.Source, ref.Chunk, ref.ExpiresAt)
		},
	}
	return runTransaction(ctx, c.db, stmtFuncs)
}

func (c *client) RenewReference(ctx context.Context, src string, expiresAt time.Time) error {
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			return txn.Model(&refModel{}).
				Where(`sourcetype = ?`, STTemporary).
				Where(`source = ?`, src).
				Update("expires_at", expiresAt)
		},
	}
	return runTransaction(ctx, c.db, stmtFuncs)
}

func (c *client) DeleteReference(ctx context.Context, ref *Reference) (retErr error) {
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			// Delete the references with the same sourcetype and source (chunk is ignored).
			// Return the chunks that should be deleted (the chunks will have zero references
			// after the references are removed).
			return txn.Exec(`
				DELETE FROM refs
				WHERE sourcetype = ?
				AND source = ?
		      `, ref.Sourcetype, ref.Source)
		},
	}
	return runTransaction(ctx, c.db, stmtFuncs)
}

type mockClient struct{}

// NewMockClient creates a new mock client.
func NewMockClient() Client {
	return &mockClient{}
}

func (c *mockClient) ReserveChunk(ctx context.Context, chunk, tmp string, expiresAt time.Time) error {
	return nil
}

func (c *mockClient) CreateReference(ctx context.Context, ref *Reference) error {
	return nil
}

func (c *mockClient) DeleteReference(ctx context.Context, ref *Reference) error {
	return nil
}

func (c *mockClient) RenewReference(context.Context, string, time.Time) error {
	return nil
}
