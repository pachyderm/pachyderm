package gc

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/prometheus/client_golang/prometheus"
)

// Reference describes a reference to a chunk in chunk storage.  If a chunk has
// no references, it will be deleted.
//  * Sourcetype - the type of reference, one of:
//   * 'temporary' - a temporary reference to a chunk
//   * 'chunk' - a cross-chunk reference, from one chunk to another
//   * 'semantic' - a reference to a chunk by some semantic name
//  * Source - the source of the reference, this may be a temporary id, chunk id, or a
//    semantic name
//  * Chunk - the target chunk being referenced
type Reference struct {
	Sourcetype string
	Source     string
	Chunk      string
}

// Client is the interface provided by the garbage collector client, for use on
// worker nodes.  It will directly perform reference-counting operations on the
// cluster's postgres database, but synchronize chunk deletions with the
// garbage collector service running in pachd.
type Client interface {
	// ReserveChunk ensures that a chunk is not deleted during the course of a
	// temporary reference.  It will add a temporary reference to the given chunk, even
	// if it doesn't exist yet.  If the specified chunk is currently
	// being deleted, this call will block while it flushes the delete through
	// the garbage collector service.
	ReserveChunk(context.Context, string, string) error

	AddReference(context.Context, *Reference) error
	RemoveReference(context.Context, *Reference) error
}

type client struct {
	server Server
	db     *gorm.DB
}

// NewClient constructs a garbage collector client:
//  * server - an object implementing the Server interface that will be
//      called for forwarding deletion candidates and flushing deletes.
//  * postgresHost, postgresPort - the host and port of the postgres instance
//      which is used for coordinating garbage collection reference counts with
//      the garbage collector clients
//  * registry (optional) - a Prometheus stats registry for tracking usage and
//    performance
func NewClient(server Server, postgresHost string, postgresPort uint16, registry prometheus.Registerer) (Client, error) {
	// Opening a connection is done lazily, initialization will connect
	db, err := openDatabase(postgresHost, postgresPort)
	if err != nil {
		return nil, err
	}

	db.LogMode(false)

	err = initializeDb(db)
	if err != nil {
		return nil, err
	}

	if registry != nil {
		initPrometheus(registry)
	}

	return &client{
		server: server,
		db:     db,
	}, nil
}

func (c *client) ReserveChunk(ctx context.Context, chunk, tmpID string) (retErr error) {
	defer func(startTime time.Time) { applyRequestStats("ReserveChunk", retErr, startTime) }(time.Now())
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
					INSERT INTO refs (sourcetype, source, chunk, created)
					SELECT 'temporary'::reftype, ?, chunk, NOW()
					FROM added_chunk
					WHERE deleting IS NULL
				)
				SELECT chunk
				FROM added_chunk
				WHERE deleting IS NOT NULL
			`, chunk, tmpID).Scan(&flushChunk)
		},
	}
	for {
		if err := runTransaction(ctx, c.db, stmtFuncs, reserveChunkStats); err != nil {
			return err
		}
		if len(flushChunk) == 0 {
			return nil
		}
		if err := c.server.FlushDelete(ctx, chunk); err != nil {
			return err
		}
	}
}

func (c *client) AddReference(ctx context.Context, ref *Reference) (retErr error) {
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			// Insert the reference.
			// Does nothing if the reference already exists.
			// (bryce) should we consider the possibility of a very slow client
			// not adding a reference before the temporary reference times out?
			// We could do some reachability validation, but we would probably want
			// to do this when the semantic reference is setup to minimize the cost on
			// the common path.
			return txn.Exec(`
				INSERT INTO refs (sourcetype, source, chunk, created)
				VALUES (?, ?, ?, NOW())
				ON CONFLICT DO NOTHING
			`, ref.Sourcetype, ref.Source, ref.Chunk)
		},
	}
	return runTransaction(ctx, c.db, stmtFuncs, nil)
}

func (c *client) RemoveReference(ctx context.Context, ref *Reference) (retErr error) {
	var chunksDeleted []chunkModel
	stmtFuncs := []statementFunc{
		func(txn *gorm.DB) *gorm.DB {
			// Delete the references with the same sourcetype and source (chunk is ignored).
			// Return the chunks that should be deleted (the chunks will have zero references
			// after the references are removed).
			return txn.Raw(`
				WITH deleted_refs AS (
					DELETE FROM refs
					WHERE sourcetype = ?
					AND source = ?
					RETURNING chunk
				)
				SELECT deleted_refs.chunk
				FROM deleted_refs
				JOIN refs
				ON deleted_refs.chunk = refs.chunk
				GROUP BY 1
				HAVING COUNT(*) = 1
		      `, ref.Sourcetype, ref.Source).Scan(&chunksDeleted)
		},
	}
	if err := runTransaction(ctx, c.db, stmtFuncs, nil); err != nil {
		return err
	}
	for _, chunk := range chunksDeleted {
		c.server.DeleteChunk(ctx, chunk.Chunk)
	}
	return nil
}
