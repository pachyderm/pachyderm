package gc

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/prometheus/client_golang/prometheus"
)

// Reference describes a reference to a chunk in chunk storage.  If a chunk has
// no references, it will be deleted.
//  * Sourcetype - the type of reference, one of:
//   * 'job' - a temporary reference to the chunk, tied to the lifetime of a job
//   * 'chunk' - a cross-chunk reference, from one chunk to another
//   * 'semantic' - a reference to a chunk by some semantic name
//  * Source - the source of the reference, this may be a job id, chunk id, or a
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
	// ReserveChunks ensures that chunks are not deleted during the course of a
	// job.  It will add a temporary reference to each of the given chunks, even
	// if they don't exist yet.  If one of the specified chunks is currently
	// being deleted, this call will block while it flushes the deletes through
	// the garbage collector service running in pachd.
	ReserveChunks(ctx context.Context, jobID string, chunks []string) error

	// UpdateReferences should be run _after_ ReserveChunks, once new chunks have
	// be written to object storage.  This will add and remove the specified
	// references, and remove all references held by the given job ID.  Any
	// chunks which are no longer referenced will be forwarded to the garbage
	// collector service in pachd for summary destruction.
	UpdateReferences(ctx context.Context, add []Reference, remove []Reference, releaseJobID string) error
}

type clientImpl struct {
	server Server
	db     *gorm.DB
}

// MakeClient constructs a garbage collector server:
//  * server - an object implementing the Server interface that will be
//      called for forwarding deletion candidates and flushing deletes.
//  * postgresHost, postgresPort - the host and port of the postgres instance
//      which is used for coordinating garbage collection reference counts with
//      the garbage collector clients
//  * registry (optional) - a Prometheus stats registry for tracking usage and
//    performance
func MakeClient(
	server Server,
	postgresHost string,
	postgresPort uint16,
	registry prometheus.Registerer,
) (Client, error) {
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

	return &clientImpl{
		db:     db,
		server: server,
	}, nil
}

func (ci *clientImpl) reserveChunksInDatabase(ctx context.Context, job string, chunks []string) ([]string, error) {
	chunkIDs := []string{}
	for _, chunk := range chunks {
		chunkIDs = append(chunkIDs, fmt.Sprintf("('%s')", chunk))
	}
	sort.Strings(chunkIDs)

	var chunksToFlush []chunkModel
	statements := []stmtCallback{
		func(txn *gorm.DB) *gorm.DB {
			return txn.Exec("set local synchronous_commit = off;")
		},
		func(txn *gorm.DB) *gorm.DB {
			return txn.Raw(`
with
added_chunks as (
 insert into chunks (chunk) values `+strings.Join(chunkIDs, ",")+`
 on conflict (chunk) do update set chunk = excluded.chunk
 returning chunk, deleting
), added_refs as (
 insert into refs (chunk, source, sourcetype)
  select
   chunk, ?, 'job'::reftype
  from added_chunks where
   deleting is null
	order by 1
)

select chunk from added_chunks where deleting is not null;
			`, job).Scan(&chunksToFlush)
		},
	}

	if err := runTransaction(ctx, ci.db, statements, reserveChunksStats); err != nil {
		return nil, err
	}

	return convertChunks(chunksToFlush), nil
}

func (ci *clientImpl) ReserveChunks(ctx context.Context, job string, chunks []string) (retErr error) {
	defer func(startTime time.Time) { applyRequestStats("ReserveChunks", retErr, startTime) }(time.Now())
	chunkIDs := []string{}
	for _, c := range chunks {
		chunkIDs = append(chunkIDs, c)
	}
	fmt.Printf("ReserveChunks (%s): %v\n", job, strings.Join(chunkIDs, ", "))
	if len(chunks) == 0 {
		return nil
	}

	var err error
	for len(chunks) > 0 {
		chunks, err = ci.reserveChunksInDatabase(ctx, job, chunks)
		if err != nil {
			return err
		}

		if len(chunks) > 0 {
			if err := ci.server.FlushDeletes(ctx, chunks); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ci *clientImpl) UpdateReferences(ctx context.Context, add []Reference, remove []Reference, releaseJob string) (retErr error) {
	defer func(startTime time.Time) { applyRequestStats("UpdateReferences", retErr, startTime) }(time.Now())
	addStr := []string{}
	removeStr := []string{}
	for _, ref := range add {
		addStr = append(addStr, fmt.Sprintf("%s/%s:%s", ref.Sourcetype, ref.Source, ref.Chunk))
	}
	for _, ref := range remove {
		removeStr = append(removeStr, fmt.Sprintf("%s/%s:%s", ref.Sourcetype, ref.Source, ref.Chunk))
	}
	fmt.Printf("UpdateReferences (%s):\n", releaseJob)
	fmt.Printf("  add:\n")
	for _, x := range addStr {
		fmt.Printf("    %v\n", x)
	}
	fmt.Printf("  remove:\n")
	for _, x := range removeStr {
		fmt.Printf("    %v\n", x)
	}

	removeValues := make([][]interface{}, 0, len(remove))
	for _, ref := range remove {
		removeValues = append(removeValues, []interface{}{ref.Sourcetype, ref.Source, ref.Chunk})
	}

	addValues := make([][]interface{}, len(add))
	for _, ref := range add {
		addValues = append(addValues, []interface{}{ref.Sourcetype, ref.Source, ref.Chunk})
	}

	var chunksToDelete []chunkModel
	statements := []stmtCallback{
		func(txn *gorm.DB) *gorm.DB {
			if len(addValues) > 0 {
				return txn.Exec(`
insert into refs (sourcetype, source, chunk) values ?
on conflict do nothing
				`, addValues)
			}
			return txn
		},
		func(txn *gorm.DB) *gorm.DB {
			builder := txn.Table(refTable)
			if len(removeValues) > 0 {
				builder = builder.Where("(sourcetype, source, chunk) in (?)", removeValues)
			} else {
				builder = builder.Where("(sourcetype, source, chunk) in (?)", nil)
			}
			if releaseJob != "" {
				builder = builder.Or("sourcetype = 'job' AND source = ?", releaseJob)
			}
			refQuery := builder.Order("sourcetype").Order("source").Order("chunk").QueryExpr()

			return txn.Raw(`
with del_refs as (
 delete from refs using (?) del
 where
  refs.sourcetype = del.sourcetype and
	refs.source = del.source and
	refs.chunk = del.chunk
 returning refs.chunk
), deleted as (
 select chunk, count(*) from del_refs group by 1 order by 1
), before as (
 select chunk, count(*) as count from refs join deleted using (chunk) group by 1 order by 1
)

select chunk from before join deleted using (chunk) where before.count - deleted.count = 0
      `, refQuery).Scan(&chunksToDelete)
		},
	}

	if err := runTransaction(ctx, ci.db, statements, updateReferencesStats); err != nil {
		return err
	}

	if len(chunksToDelete) > 0 {
		if err := ci.server.DeleteChunks(ctx, convertChunks(chunksToDelete)); err != nil {
			return err
		}
	}
	return nil
}
