package gc

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

func initializeDb(ctx context.Context, db *sql.DB) error {
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

func isRetriableError(err error) bool {
	if err, ok := err.(*pq.Error); ok {
		return err.Code.Class().Name() == "transaction_rollback"
	}
	return false
}

func applyRequestMetrics(request string, err error, start time.Time) {
	var result string
	switch err {
	//TODO: add more resolution here
	case nil:
		result = "success"
	default:
		result = "error"
	}
	requestResults.WithLabelValues(request, result).Inc()
	requestTime.WithLabelValues(request).Observe(float64(time.Since(start).Seconds()))
}

// TODO: include time spent in sql
func applySqlMetrics(operation string, err error, start time.Time) {
	var result string
	switch x := err.(type) {
	case nil:
		result = "success"
	case *pq.Error:
		result = x.Code.Name()
	default:
		result = "unknown"
	}
	sqlResults.WithLabelValues(operation, result).Inc()
	sqlTime.WithLabelValues(operation).Observe(float64(time.Since(start).Seconds()))
}

func applyDeleteMetrics(err error, start time.Time) {
	var result string
	switch x := err.(type) {
	case nil:
		result = "success"
	case *pq.Error:
		result = x.Code.Name()
	default:
		result = "unknown"
	}
	deleteResults.WithLabelValues(result).Inc()
	deleteTime.Observe(float64(time.Since(start).Seconds()))
}
