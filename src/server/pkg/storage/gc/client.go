package gc

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/prometheus/client_golang/prometheus"
)

type Reference struct {
	sourcetype string
	source     string
	chunk      chunk.Chunk
}

type Client interface {
	ReserveChunks(context.Context, string, []chunk.Chunk) error
	UpdateReferences(context.Context, []Reference, []Reference, string) error
}

type ClientImpl struct {
	server Server
	db     *gorm.DB
}

func MakeClient(ctx context.Context, server Server, host string, port uint16, registry prometheus.Registerer) (Client, error) {
	// Opening a connection is done lazily, initialization will connect
	db, err := openDatabase(host, port)
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

	return &ClientImpl{
		db:     db,
		server: server,
	}, nil
}

func (gcc *ClientImpl) reserveChunksInDatabase(ctx context.Context, job string, chunks []chunk.Chunk) ([]chunk.Chunk, error) {
	chunkIds := []string{}
	for _, chunk := range chunks {
		chunkIds = append(chunkIds, fmt.Sprintf("('%s')", chunk.Hash))
	}
	sort.Strings(chunkIds)

	var chunksToFlush []chunkModel
	statements := []stmtCallback{
		func(txn *gorm.DB) *gorm.DB {
			return txn.Exec("set local synchronous_commit = off;")
		},
		func(txn *gorm.DB) *gorm.DB {
			return txn.Raw(`
with
added_chunks as (
 insert into chunks (chunk) values `+strings.Join(chunkIds, ",")+`
 on conflict (chunk) do update set chunk = excluded.chunk
 returning chunk, deleting
),
added_refs as (
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

	if err := runTransaction(gcc.db, ctx, statements, reserveChunksStats); err != nil {
		return nil, err
	}

	return convertChunks(chunksToFlush), nil
}

func (gcc *ClientImpl) ReserveChunks(ctx context.Context, job string, chunks []chunk.Chunk) (retErr error) {
	defer func(startTime time.Time) { applyRequestStats("ReserveChunks", retErr, startTime) }(time.Now())
	if len(chunks) == 0 {
		return nil
	}

	var err error
	for len(chunks) > 0 {
		chunks, err = gcc.reserveChunksInDatabase(ctx, job, chunks)
		if err != nil {
			return err
		}

		if len(chunks) > 0 {
			if err := gcc.server.FlushDeletes(ctx, chunks); err != nil {
				return err
			}
		}
	}
	return nil
}

func (gcc *ClientImpl) UpdateReferences(ctx context.Context, add []Reference, remove []Reference, releaseJob string) (retErr error) {
	defer func(startTime time.Time) { applyRequestStats("UpdateReferences", retErr, startTime) }(time.Now())

	removeValues := make([][]interface{}, 0, len(remove))
	for _, ref := range remove {
		removeValues = append(removeValues, []interface{}{ref.sourcetype, ref.source, ref.chunk.Hash})
	}

	addValues := make([][]interface{}, len(add))
	for _, ref := range add {
		addValues = append(addValues, []interface{}{ref.sourcetype, ref.source, ref.chunk.Hash})
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
),
counts as (
 select chunk, count(*) - 1 as count from refs join del_refs using (chunk) group by 1 order by 1
)

select chunk from counts where count = 0
      `, refQuery).Scan(&chunksToDelete)
		},
	}

	if err := runTransaction(gcc.db, ctx, statements, updateReferencesStats); err != nil {
		return err
	}

	if len(chunksToDelete) > 0 {
		if err := gcc.server.DeleteChunks(ctx, convertChunks(chunksToDelete)); err != nil {
			return err
		}
	}
	return nil
}
