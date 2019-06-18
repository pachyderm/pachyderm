package gc

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/lib/pq"
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
	server                    Server
	db                        *sql.DB
	reserveChunksCounter      *prometheus.CounterVec
	reserveChunksHistogram    prometheus.Histogram
	updateReferencesCounter   *prometheus.CounterVec
	updateReferencesHistogram prometheus.Histogram
}

func initializeDb(ctx context.Context, db *sql.DB) error {
	// TODO: move initialization somewhere more consistent
	_, err := db.ExecContext(ctx, `
do $$ begin
 create type reftype as enum ('chunk', 'job', 'semantic');
exception
 when duplicate_object then null;
end $$
  `)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
create table if not exists refs (
 sourcetype reftype not null,
 source text not null,
 chunk text not null,
 primary key(sourcetype, source, chunk)
)`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
create table if not exists chunks (
 chunk text primary key,
 deleting timestamp
)`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
create index if not exists idx_chunk on refs (chunk)
`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
create index if not exists idx_sourcetype_source on refs (sourcetype, source)
`)
	if err != nil {
		return err
	}

	return nil
}

func MakeClient(ctx context.Context, server Server, host string, port int16, metrics prometheus.Registerer) (Client, error) {
	connStr := fmt.Sprintf("host=%s port=%d dbname=pgc user=pachyderm password=elephantastic sslmode=disable", host, port)
	connector, err := pq.NewConnector(connStr)
	if err != nil {
		return nil, err
	}

	// Opening a connection is done lazily, initialization will connect
	db := sql.OpenDB(connector)

	err = initializeDb(ctx, db)
	if err != nil {
		return nil, err
	}

	reserveChunksCounter, reserveChunksHistogram := makeCollectors("reserve_chunks", "")
	updateReferencesCounter, updateReferencesHistogram := makeCollectors("update_references", "")

	if metrics != nil {
		collectors := []prometheus.Collector{
			reserveChunksCounter,
			reserveChunksHistogram,
			updateReferencesCounter,
			updateReferencesHistogram,
		}

		for _, c := range collectors {
			if err := metrics.Register(c); err != nil {
				// metrics may be redundantly registered; ignore these errors
				if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
					fmt.Errorf("error registering prometheus metric: %v", err)
				}
			}
		}
	}

	return &ClientImpl{
		db:                        db,
		server:                    server,
		reserveChunksCounter:      reserveChunksCounter,
		reserveChunksHistogram:    reserveChunksHistogram,
		updateReferencesCounter:   updateReferencesCounter,
		updateReferencesHistogram: updateReferencesHistogram,
	}, nil
}

func readChunksFromCursor(cursor *sql.Rows) []chunk.Chunk {
	chunks := []chunk.Chunk{}
	for cursor.Next() {
		var hash string
		if err := cursor.Scan(&hash); err != nil {
			return nil
		}
		chunks = append(chunks, chunk.Chunk{Hash: hash})
	}
	return chunks
}

func isRetriableError(err error, counter *prometheus.CounterVec) bool {
	if err, ok := err.(*pq.Error); ok {
		fmt.Printf("error: %s\n", err.Code.Name())
		counter.With(prometheus.Labels{"result": err.Code.Name()}).Add(1)
		name := err.Code.Class().Name()
		return name == "transaction_rollback"
	}
	return false
}

func (gcc *ClientImpl) runReserveSql(ctx context.Context, query string) ([]chunk.Chunk, error) {
	txn, err := gcc.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}

	var chunksToFlush []chunk.Chunk
	err = func() (retErr error) {
		defer func() {
			if retErr != nil {
				txn.Rollback()
			}
		}()

		_, err = txn.ExecContext(ctx, "set local synchronous_commit = off;")
		if err != nil {
			return err
		}

		cursor, err := txn.QueryContext(ctx, query)
		if err != nil {
			return err
		}

		// Flush returned chunks through the server
		chunksToFlush = readChunksFromCursor(cursor)
		cursor.Close()
		if err := cursor.Err(); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return nil, err
	}

	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return chunksToFlush, nil
}

func (gcc *ClientImpl) reserveChunksInDatabase(ctx context.Context, job string, chunks []chunk.Chunk) ([]chunk.Chunk, error) {
	chunkIds := []string{}
	for _, chunk := range chunks {
		chunkIds = append(chunkIds, fmt.Sprintf("('%s')", chunk.Hash))
	}
	sort.Strings(chunkIds)

	query := `
with
added_chunks as (
 insert into chunks (chunk)
  values ` + strings.Join(chunkIds, ",") + `
 on conflict (chunk) do update set chunk = excluded.chunk
 returning chunk, deleting
),
added_refs as (
 insert into refs (chunk, source, sourcetype)
  select
   chunk, '` + job + `', 'job'::reftype
  from added_chunks where
   deleting is null
	order by 1
)

select chunk from added_chunks where deleting is not null;
	`

	parameters := []interface{}{job}
	for _, chunk := range chunks {
		parameters = append(parameters, chunk.Hash)
	}

	var chunksToFlush []chunk.Chunk
	for {
		var err error
		chunksToFlush, err = gcc.runReserveSql(ctx, query)
		if err == nil {
			break
		}
		if !isRetriableError(err, gcc.reserveChunksCounter) {
			return nil, err
		}
	}

	gcc.reserveChunksCounter.With(prometheus.Labels{"result": "success"}).Add(1)
	return chunksToFlush, nil
}

func (gcc *ClientImpl) ReserveChunks(ctx context.Context, job string, chunks []chunk.Chunk) error {
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

func (gcc *ClientImpl) UpdateReferences(ctx context.Context, add []Reference, remove []Reference, releaseJob string) error {
	var removeStr string
	if len(remove) == 0 {
		removeStr = "null"
	} else {
		removes := []string{}
		for _, ref := range remove {
			removes = append(removes, fmt.Sprintf("('%s', '%s', '%s')", ref.sourcetype, ref.source, ref.chunk.Hash))
		}
		removeStr = strings.Join(removes, ",")
	}

	var addStr string
	if len(add) == 0 {
		addStr = ""
	} else {
		adds := []string{}
		for _, ref := range add {
			adds = append(adds, fmt.Sprintf("('%s', '%s', '%s')", ref.sourcetype, ref.source, ref.chunk.Hash))
		}
		addStr = `
added_refs as (
 insert into refs (sourcetype, source, chunk) values ` + strings.Join(adds, ",") + `
 on conflict do nothing
),
		`
	}

	var jobStr string
	if releaseJob == "" {
		jobStr = "null"
	} else {
		jobStr = fmt.Sprintf("('job', '%s')", releaseJob)
	}

	query := `
with
` + addStr + `
del_refs as (
 delete from refs using (
	 select sourcetype, source, chunk from refs
	 where
		(sourcetype, source, chunk) in (` + removeStr + `) or
		(sourcetype, source) in (` + jobStr + `)
	 order by 1, 2, 3
 ) del
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
	`

	var chunksToDelete []chunk.Chunk
	for {
		txn, err := gcc.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return err
		}

		cursor, err := txn.QueryContext(ctx, query)
		if err != nil {
			if err := txn.Rollback(); err != nil {
				return err
			}
			if isRetriableError(err, gcc.updateReferencesCounter) {
				continue
			}
			return err
		}

		chunksToDelete = readChunksFromCursor(cursor)
		cursor.Close()
		if err := cursor.Err(); err != nil {
			if err := txn.Rollback(); err != nil {
				return err
			}
			if isRetriableError(err, gcc.updateReferencesCounter) {
				continue
			}
			return err
		}

		if err := txn.Commit(); err != nil {
			if isRetriableError(err, gcc.updateReferencesCounter) {
				continue
			}
			return err
		}
		break
	}

	if len(chunksToDelete) > 0 {
		if err := gcc.server.DeleteChunks(ctx, chunksToDelete); err != nil {
			return err
		}
	}
	gcc.reserveChunksCounter.With(prometheus.Labels{"result": "success"}).Add(1)
	return nil
}
