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
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
)

func schema(ctx context.Context, conn *pgxpool.Pool) {
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

func addSomeObjects(ctx context.Context, conn *pgxpool.Pool) {
	ch := make(chan int)
	do := func(i int) {
		tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
		if err != nil {
			log.Exit(ctx, "begin", zap.Error(err))
		}

		name := fmt.Sprintf("%v/%v", rand.Int63(), rand.Int63())
		var expires sql.NullTime
		if rand.Intn(10) != 0 {
			expires.Valid = true
			expires.Time = time.Now().Add(time.Minute + time.Duration(i)*20*time.Microsecond)
		}
		row := tx.QueryRow(ctx, "insert into tracker_objects (str_id, expires_at) values ($1, $2) returning int_id", name, expires)
		var x int
		if err := row.Scan(&x); err != nil {
			log.Error(ctx, "exec insert into tracker_objects; scan result", zap.Error(err))
			tx.Rollback(ctx)
			return
		}
		referenced := map[int]struct{}{}
		for j := 0; j < rand.Intn(50); j++ {
			k := rand.Intn(x)
			if _, ok := referenced[k]; ok {
				continue
			}
			referenced[k] = struct{}{}
			if _, err := tx.Exec(ctx, "insert into tracker_refs (from_id, to_id) VALUES ($1, $2)", k, x); err != nil {
				log.Error(ctx, "exec insert into tracker_refs", zap.Error(err))
				tx.Rollback(ctx)
				return
			}
		}

		if err := tx.Commit(ctx); err != nil {
			log.Error(ctx, "commit", zap.Error(err))
			return
		}
	}
	for i := 0; i < 64; i++ {
		go func() {
			for n := range ch {
				do(n)
			}
		}()
	}

	for i := 0; ; i++ {
		if i%10000 == 0 {
			tx, err := conn.Begin(ctx)
			if err != nil {
				log.Error(ctx, "begin size query", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}
			row := tx.QueryRow(ctx, `select count(1) from tracker_objects`)
			var n int
			if err := row.Scan(&n); err != nil {
				log.Error(ctx, "scan size", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}
			tx.Rollback(ctx)
			if n > 100e6 {
				log.Info(ctx, "done inserting for a while", zap.Int("n", n))
				close(ch)
				return
			}
			log.Info(ctx, "inserting...", zap.Int("i", i), zap.Int("n", n))
		}
		ch <- i
	}
}

func gc(ctx context.Context, conn *pgxpool.Pool) {
	ctx, done := log.SpanContextL(ctx, "gc", log.InfoLevel)
	defer done()
	for {
		tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
		if err != nil {
			log.Error(ctx, "begin", zap.Error(err))
			return
		}
		rows, err := tx.Query(ctx, `SELECT int_id FROM tracker_objects as objs WHERE NOT EXISTS (SELECT 1 FROM tracker_refs as refs where objs.int_id = refs.to_id) AND expires_at <= CURRENT_TIMESTAMP limit 50000;`)
		if err != nil {
			log.Error(ctx, "gc query", zap.Error(err))
			tx.Rollback(ctx)
			return
		}
		defer rows.Close()
		var getRidOf []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				log.Error(ctx, "scan", zap.Error(err))
				tx.Rollback(ctx)
				return
			}
			getRidOf = append(getRidOf, id)
		}
		log.Info(ctx, "collect", zap.Int("n", len(getRidOf)))
		rows.Close()
		if len(getRidOf) == 0 {
			return
		}
		var n int64
		for _, id := range getRidOf {
			if _, err := tx.Exec(ctx, `DELETE FROM tracker_objects WHERE int_id=$1`, id); err != nil {
				log.Error(ctx, "delete from tracker_objects", zap.Int("str_id", id), zap.Error(err))
				tx.Rollback(ctx)
				return
			}
			result, err := tx.Exec(ctx, `DELETE FROM tracker_refs WHERE from_id=$1`, id)
			if err != nil {
				log.Error(ctx, "delete from tracker_refs", zap.Int("str_id", id), zap.Error(err))
				tx.Rollback(ctx)
				return
			}
			n += result.RowsAffected()
		}
		log.Info(ctx, "delete refs", zap.Int64("n", n))
		if err := tx.Commit(ctx); err != nil {
			log.Error(ctx, "commit", zap.Error(err))
			tx.Rollback(ctx)
			return
		}
	}
}

func main() {
	log.InitPachctlLogger()
	log.SetLevel(log.InfoLevel)
	ctx, c := signal.NotifyContext(pctx.Background(""), os.Interrupt)
	defer c()

	cfg, err := pgxpool.ParseConfig("user=postgres host=localhost port=5432 database=foo")
	if err != nil {
		log.Exit(ctx, "parse config", zap.Error(err))
	}
	cfg.ConnConfig.Logger = log.NewPGX("pgx")
	cfg.MaxConns = 64
	conn, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		log.Exit(ctx, "connect", zap.Error(err))
	}

	schema(ctx, conn)
	go func() {
		for {
			addSomeObjects(ctx, conn)
			time.Sleep(30 * time.Second)
		}
	}()
	go func() {
		for {
			gc(ctx, conn)
			time.Sleep(10 * time.Second)
		}
	}()
	<-ctx.Done()
}
