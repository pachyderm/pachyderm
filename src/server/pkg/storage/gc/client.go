package gc

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

// TODO: remove debugging timing fn
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}

type Client struct {
	pachClient *client.APIClient
	db         *sql.DB
}

type Reference struct {
	sourcetype string
	source     string
	chunk      chunk.Chunk
}

func initializeDb(db *sql.DB, ctx context.Context) error {
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

	_, err = db.ExecContext(ctx, `create index on refs (chunk)`)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `create index on refs (sourcetype, source)`)
	if err != nil {
		return err
	}

	return nil
}

func NewClient(pachClient *client.APIClient, host string, port int16) (*Client, error) {
	connStr := fmt.Sprintf("host=%s port=%d dbname=pgc user=pachyderm password=elephantastic sslmode=disable", host, port)
	connector, err := pq.NewConnector(connStr)
	if err != nil {
		return nil, err
	}

	// Opening a connection is done lazily, statement preparation will connect
	db := sql.OpenDB(connector)
	ctx := pachClient.Ctx()

	err = initializeDb(db, ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		db:         db,
		pachClient: pachClient,
	}, nil
}

func (gcc *Client) ReserveChunks(job string, chunks []chunk.Chunk) error {
	defer timeTrack(time.Now(), "ReserveChunks")
	if len(chunks) == 0 {
		return nil
	}

	ctx := gcc.pachClient.Ctx()

	// TODO: check for conflict errors and retry in a loop
	txn, err := gcc.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	questions := []string{}
	for i := 0; i < len(chunks); i++ {
		questions = append(questions, fmt.Sprintf("($%d)", 2+i))
	}

	query := `
with
added_chunks as (
 insert into chunks (chunk)
  values ` + strings.Join(questions, ",") + `
 on conflict (chunk) do update set chunk = excluded.chunk
 returning chunk, deleting
),
added_refs as (
 insert into refs (chunk, source, sourcetype)
  select
   chunk,
   ($1) as source,
   'job'::reftype as sourcetype
  from added_chunks where
   deleting is null
)

select chunk from added_chunks where deleting is not null;
	`

	parameters := []interface{}{job}
	for _, chunk := range chunks {
		parameters = append(parameters, chunk.Hash)
	}

	rows, err := txn.QueryContext(ctx, query, parameters...)
	if err != nil {
		return err
	}

	// The cursor must be closed before committing
	rows.Close()

	if err := txn.Commit(); err != nil {
		return err
	}

	// If any returned rows have `removing` or `deleting` non-nil, flush those
	// rows through the `FlushDeletes` RPC.

	return nil
}

func (gcc *Client) UpdateReferences(add []Reference, remove []Reference, releaseJobs []string) error {
	defer timeTrack(time.Now(), "UpdateReferences")
	ctx := gcc.pachClient.Ctx()

	// TODO: check for conflict errors and retry in a loop
	txn, err := gcc.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	insertStmt, err := txn.Prepare(pq.CopyIn("refs", "sourcetype", "source", "chunk"))
	if err != nil {
		return err
	}
	defer insertStmt.Close()
	for _, ref := range add {
		_, err := insertStmt.Exec(ref.sourcetype, ref.source, ref.chunk.Hash)
		if err != nil {
			return err
		}
	}
	_, err = insertStmt.Exec()
	if err != nil {
		return err
	}

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

	var jobStr string
	if len(releaseJobs) == 0 {
		jobStr = "null"
	} else {
		jobs := []string{}
		for _, job := range releaseJobs {
			jobs = append(jobs, fmt.Sprintf("('job', '%s')", job))
		}
		jobStr = strings.Join(jobs, ",")
	}

	query := `
with
del_refs as (
 delete from refs where
  (sourcetype, source, chunk) in (` + removeStr + `) or 
	(sourcetype, source) in (` + jobStr + `)
 returning chunk
),
counts as (
 select chunk, count(*) - 1 as count from refs join del_refs using (chunk) group by 1
)

update chunks set
 deleting = now()
from counts where
 counts.chunk = chunks.chunk and
 count = 0
returning chunks.chunk;
	`

	rows, err := txn.QueryContext(ctx, query)
	if err != nil {
		return err
	}

	rows.Close()

	if err := txn.Commit(); err != nil {
		return err
	}

	// Verify that the rows are not removing, deleting, or missing - if so, return
	// an error - the user will need to reserve the chunks then try again

	return nil
}
