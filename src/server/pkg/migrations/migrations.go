package migrations

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/sirupsen/logrus"
)

type Env struct {
	// TODO: etcd
	ObjectClient obj.Client
	Tx           *sqlx.Tx
}

func MakeEnv(objC obj.Client) Env {
	return Env{
		ObjectClient: objC,
	}
}

type Func func(ctx context.Context, env Env) error

type State struct {
	n      int
	prev   *State
	change Func
}

func (s State) Apply(fn Func) State {
	return State{
		prev:   &s,
		change: fn,
		n:      s.n + 1,
	}
}

func (s State) Number() int {
	return s.n
}

func InitialState() State {
	return State{
		change: func(ctx context.Context, env Env) error {
			_, err := env.Tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS migrations (
				id BIGINT PRIMARY KEY,
				start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				end_time TIMESTAMP
			);
			INSERT INTO migrations (id, start_time, end_time) VALUES (0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING;
			`)
			return err
		},
	}
}

func ApplyMigrations(ctx context.Context, db *sqlx.DB, baseEnv Env, state State) error {
	if state.prev != nil {
		if err := ApplyMigrations(ctx, db, baseEnv, *state.prev); err != nil {
			return err
		}
	}
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	env := baseEnv
	env.Tx = tx
	if err := func() error {
		if state.n == 0 {
			if err := state.change(ctx, env); err != nil {
				panic(err)
			}
		}
		_, err := tx.ExecContext(ctx, `LOCK TABLE migrations IN EXCLUSIVE MODE NOWAIT`)
		if err != nil {
			return err
		}
		if finished, err := isFinished(ctx, tx, state.n); err != nil {
			return err
		} else if finished {
			// skip migration
			logrus.Infof("migration %d already applied", state.n)
			return nil
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO migrations (id, start_time) VALUES ($1, CURRENT_TIMESTAMP)`, state.n); err != nil {
			return err
		}
		logrus.Info("applying migration", state.n)
		if err := state.change(ctx, env); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE migrations SET end_time = CURRENT_TIMESTAMP WHERE id = $1`, state.n); err != nil {
			return err
		}
		logrus.Info("successfully applied migration", state.n)
		return nil
	}(); err != nil {
		if err := tx.Rollback(); err != nil {
			logrus.Error(err)
		}
		return err
	}
	return tx.Commit()
}

func BlockUntil(ctx context.Context, db *sqlx.DB, state State) error {
	const (
		schemaName = "public"
		tableName  = "migrations"
	)
	// poll database until this state is registered
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		var tableExists bool
		if err := db.GetContext(ctx, &tableExists, `SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = $1
			AND table_name = $2
		)`, schemaName, tableName); err != nil {
			return err
		}
		if tableExists {
			var latest int
			if err := db.GetContext(ctx, &latest, `SELECT COALESCE(MAX(id), 0) FROM migrations`); err != nil && err != sql.ErrNoRows {
				return err
			}
			if latest == state.n {
				return nil
			} else if latest > state.n {
				return errors.Errorf("database state is newer than application is expecting")
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func isFinished(ctx context.Context, tx *sqlx.Tx, n int) (bool, error) {
	var count int
	if err := tx.GetContext(ctx, &count, `
	SELECT count(id)
	FROM migrations
	WHERE id = $1 AND end_time IS NOT NULL
	`, n); err != nil {
		return false, err
	}
	return count > 0, nil
}
