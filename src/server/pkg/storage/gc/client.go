package gc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

type Client struct {
	pachClient           *client.APIClient
	db                   *sql.DB
	reserveChunksStmt    *sql.Stmt
	updateReferencesStmt *sql.Stmt
}

type Reference struct {
	chunk      chunk.Chunk
	sourcetype string
	source     string
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

	fmt.Printf("making reserve chunks statement\n")
	reserveChunksStmt, err := db.PrepareContext(ctx, `
with
added_chunks as (
 insert into chunks (chunk)
  values ($2)
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
	`)
	if err != nil {
		return nil, err
	}

	fmt.Printf("making update references statement\n")

	updateReferencesStmt, err := db.PrepareContext(ctx, `
with
added_refs as (
 insert into refs (sourcetype, source, chunk) values $1
),
del_refs as (
 delete from refs where (sourcetype, source, chunk) in ($2) or (sourcetype, source) in ($3)
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
	`)
	if err != nil {
		return nil, err
	}

	return &Client{
		db:                   db,
		pachClient:           pachClient,
		reserveChunksStmt:    reserveChunksStmt,
		updateReferencesStmt: updateReferencesStmt,
	}, nil
}

func (gcc *Client) ReserveChunks(job string, chunks []chunk.Chunk) error {
	ctx := gcc.pachClient.Ctx()

	// TODO: check for conflict errors and retry in a loop
	txn, err := gcc.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	strChunks := []string{}
	for _, chunk := range chunks {
		strChunks = append(strChunks, chunk.Hash)
	}

	txnStmt := txn.StmtContext(ctx, gcc.reserveChunksStmt)
	rows, err := txnStmt.QueryContext(ctx, job, pq.Array(strChunks)) // TODO: probably wrong way to insert chunks
	if err != nil {
		return err
	}
	defer rows.Close()

	if err := txn.Commit(); err != nil {
		return err
	}

	// If any returned rows have `removing` or `deleting` non-nil, flush those
	// rows through the `FlushDeletes` RPC.

	return nil
}

func (gcc *Client) UpdateReferences(add []Reference, remove []Reference, releaseJobs []string) error {
	ctx := gcc.pachClient.Ctx()

	// TODO: check for conflict errors and retry in a loop
	txn, err := gcc.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	adds := []string{}
	removes := []string{}
	jobs := []string{}

	txnStmt := txn.StmtContext(ctx, gcc.updateReferencesStmt)
	rows, err := txnStmt.QueryContext(ctx, pq.Array(adds), pq.Array(removes), pq.Array(jobs)) // TODO: probably wrong way to insert things
	if err != nil {
		return err
	}
	defer rows.Close()

	if err := txn.Commit(); err != nil {
		return err
	}

	// Verify that the rows are not removing, deleting, or missing - if so, return
	// an error - the user will need to reserve the chunks then try again

	return nil
}
