package migrations

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/sirupsen/logrus"
)

type Env struct {
	// TODO: etcd
	ObjectClient obj.Client
	Tx           *sqlx.Tx
}

type MigrationFunc func(ctx context.Context, env Env) error

type State struct {
	n      int
	prev   *State
	change MigrationFunc
}

func (s State) Apply(fn MigrationFunc) State {
	return State{
		prev:   &s,
		change: fn,
		n:      s.n + 1,
	}
}

func InitialState() State {
	return State{
		change: func(ctx context.Context, env Env) error {
			_, err := env.Tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS migrations (
				id BIGSERIAL PRIMARY KEY,
				start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			);`)
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
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelLinearizable,
	})
	if err != nil {
		return err
	}
	env := baseEnv
	env.Tx = tx
	if err := func() error {
		if state.n > 0 {
			_, err := tx.ExecContext(ctx, `LOCK TABLE migrations IN EXCLUSIVE MODE NO WAIT`)
			if err != nil {
				return err
			}
			var x int
			if err := tx.GetContext(ctx, &x, `SELECT count(id) FROM migrations WHERE id = $1`, state.n); err != nil {
				return err
			}
			if x > 0 {
				// skip migration
				return nil
			}
		}
		return state.change(ctx, env)
	}(); err != nil {
		if err := tx.Rollback(); err != nil {
			logrus.Error(err)
		}
		return err
	}
	return tx.Commit()
}

func BlockUntil(ctx context.Context, db *sqlx.DB, state *State) error {
	// poll database until this state is registered
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		var latest int
		if err := db.GetContext(ctx, &latest, `SELECT MAX(id) FROM migrations`); err != nil && err != sql.ErrNoRows {
			return err
		}
		if latest == state.n {
			return nil
		} else if latest > state.n {

		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
