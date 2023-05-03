package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

func schema(ctx context.Context, conn *pgx.Conn) {
	schema := `
	CREATE TABLE IF NOT EXISTS tracker_objects (
		int_id BIGSERIAL PRIMARY KEY,
		str_id VARCHAR(4096) UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tracker_refs (
		from_id INT8 NOT NULL,
		to_id INT8 NOT NULL,
		PRIMARY KEY (from_id, to_id)
	);
`
	if _, err := conn.Exec(ctx, schema); err != nil {
		log.Exit(ctx, "exec schema", zap.Error(err))
	}
}

func addSomeObjects(ctx context.Context, conn *pgx.Conn, n int) {
	for i := 0; i < n; i++ {
		if i%1000 == 0 {
			log.Info(ctx, "inserting...", zap.Int("i", i))
		}
		tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
		if err != nil {
			log.Exit(ctx, "begin", zap.Error(err))
		}

		name := fmt.Sprintf("%v/%v", rand.Int63(), rand.Int63())
		var expires sql.NullTime
		if rand.Intn(10) != 0 {
			expires.Valid = true
			expires.Time = time.Now().Add(time.Duration(i) * 20 * time.Millisecond)
		}
		row := tx.QueryRow(ctx, "insert into tracker_objects (str_id, expires_at) values ($1, $2) returning int_id", name, expires)
		var x int
		if err := row.Scan(&x); err != nil {
			log.Exit(ctx, "exec insert into tracker_objects; scan result", zap.Error(err))
		}
		referenced := map[int]struct{}{}
		for j := 0; j < rand.Intn(10); j++ {
			k := rand.Intn(x)
			if _, ok := referenced[k]; ok {
				continue
			}
			referenced[k] = struct{}{}
			if _, err := tx.Exec(ctx, "insert into tracker_refs (from_id, to_id) VALUES ($1, $2)", x, k); err != nil {
				log.Exit(ctx, "exec insert into tracker_refs", zap.Error(err))
			}
		}

		if err := tx.Commit(ctx); err != nil {
			log.Exit(ctx, "commit", zap.Error(err))
		}
	}

}

func gc(ctx context.Context, conn *pgx.Conn) {
	ctx, done := log.SpanContextL(ctx, "gc", log.InfoLevel)
	var n int
	defer done(zap.Intp("n", &n))
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		log.Exit(ctx, "begin", zap.Error(err))
	}
	rows, err := tx.Query(ctx, `SELECT int_id FROM tracker_objects as objs WHERE NOT EXISTS (SELECT 1 FROM tracker_refs as refs where objs.int_id = refs.to_id) AND expires_at <= CURRENT_TIMESTAMP;`)
	if err != nil {
		log.Exit(ctx, "gc query", zap.Error(err))
	}
	defer rows.Close()
	var getRidOf []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Exit(ctx, "scan", zap.Error(err))
		}
		getRidOf = append(getRidOf, id)
	}
	rows.Close()
	n += len(getRidOf)
	for _, id := range getRidOf {
		if _, err := tx.Exec(ctx, `DELETE FROM tracker_objects WHERE int_id=$1`, id); err != nil {
			log.Exit(ctx, "delete from tracker_objects", zap.Int("str_id", id), zap.Error(err))
		}
		if _, err := tx.Exec(ctx, `DELETE FROM tracker_refs WHERE from_id=$1`, id); err != nil {
			log.Exit(ctx, "delete from tracker_refs", zap.Int("str_id", id), zap.Error(err))
		}
	}
	if err := tx.Commit(ctx); err != nil {
		log.Exit(ctx, "commit", zap.Error(err))
	}
}

func main() {
	log.InitPachctlLogger()
	log.SetLevel(log.InfoLevel)
	ctx, c := signal.NotifyContext(pctx.Background(""), os.Interrupt)
	defer c()
	cfg, err := pgx.ParseConfig("user=postgres host=localhost port=5432 database=foo")
	if err != nil {
		log.Exit(ctx, "parse config", zap.Error(err))
	}
	cfg.Logger = log.NewPGX("pgx")

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		log.Exit(ctx, "connect", zap.Error(err))
	}

	schema(ctx, conn)
	for {
		addSomeObjects(ctx, conn, 20000)
		gc(ctx, conn)
	}
}
