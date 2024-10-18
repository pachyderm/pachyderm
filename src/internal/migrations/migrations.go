// Package migrations implements database migrations.
package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// Env contains all the objects that can be manipulated during a migration.
// The Tx field will be overwritten with the transaction that the migration should be performed in.
type Env struct {
	// TODO: etcd
	Tx             *pachsql.Tx
	EtcdClient     *clientv3.Client
	WithTableLocks bool
}

func (env Env) LockTables(ctx context.Context, tables ...string) error {
	if !env.WithTableLocks {
		return nil
	}
	for _, table := range tables {
		if _, err := env.Tx.ExecContext(ctx, fmt.Sprintf("LOCK TABLE %s IN EXCLUSIVE MODE", table)); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

// MakeEnv returns a new Env
// The only advantage to this contructor is you can be sure all the fields are set.
// You can also create an Env using a struct literal.
func MakeEnv(etcdC *clientv3.Client) Env {
	return Env{
		EtcdClient:     etcdC,
		WithTableLocks: true,
	}
}

// Func is the type of functions that perform migrations.
type Func func(ctx context.Context, env Env) error

// State represents a state of the cluster, including all the steps taken to get there.
type State struct {
	n      int
	prev   *State
	change Func
	name   string
	squash bool
}

type ApplyOpt func(*State)

func Squash(s *State) {
	s.squash = true
}

// Apply applies a Func to the state and returns a new state.
func (s State) Apply(name string, fn Func, opts ...ApplyOpt) State {
	s2 := State{
		prev:   &s,
		change: fn,
		n:      s.n + 1,
		name:   strings.ToLower(name),
	}
	for _, opt := range opts {
		opt(&s2)
	}
	return s2
}

// Number returns the number of changes to be applied before the state can be actualized.
// The number of the initial state is 0
// State number n requires n changes from the initial state.
func (s State) Number() int {
	return s.n
}

// InitialState returns a cluster that has no migrations.
// The initial state contains a change which is just to create the migration table.
func InitialState() State {
	return State{
		name: "init",
		change: func(ctx context.Context, env Env) error {
			_, err := env.Tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS public.migrations (
				id BIGINT PRIMARY KEY,
				NAME VARCHAR(250) NOT NULL,
				start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				end_time TIMESTAMP
			);
			INSERT INTO public.migrations (id, name, start_time, end_time) VALUES (0, 'init', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING;
			`)
			return errors.EnsureStack(err)
		},
	}
}

// ApplyMigrations does the necessary work to actualize state.
// It will manipulate the objects available in baseEnv, and use the migrations table in db.
func ApplyMigrations(ctx context.Context, db *pachsql.DB, baseEnv Env, state State) (retErr error) {
	ctx, end := log.SpanContextL(ctx, "ApplyMigrations", log.InfoLevel)
	defer end(log.Errorp(&retErr))

	env := baseEnv
	states := CollectStates(state)
	for i, state := range states {
		// if this state is not squashed into the previous, open a new transaction.
		// otherwise do nothing, and reuse the previous transaction.
		if !state.squash {
			tx, err := db.BeginTxx(ctx, &sql.TxOptions{
				Isolation: sql.LevelSnapshot,
			})
			if err != nil {
				return errors.EnsureStack(err)
			}
			env.Tx = tx
		}
		// Apply the migration
		if err := ApplyMigrationTx(ctx, env, state); err != nil {
			if err := env.Tx.Rollback(); err != nil {
				log.Error(ctx, "problem rolling back migrations", zap.Error(err))
			}
			return errors.EnsureStack(err)
		}
		// If this is the last migation, or the next migration is not squashed, commit.
		if i == len(states)-1 || !states[i+1].squash {
			if err := env.Tx.Commit(); err != nil {
				log.Error(ctx, "failed to commit migration", zap.Error(err))
				return errors.EnsureStack(err)
			}
			env.Tx = nil
		}
	}
	return nil
}

// CollectStates does a reverse order traversal of a linked list and adds each item to a slice
func CollectStates(s State) []State {
	out := make([]State, 0, s.n+1)
	return collectStates(out, s)
}

func collectStates(slice []State, s State) []State {
	if s.prev != nil {
		slice = collectStates(slice, *s.prev)
	}
	return append(slice, s)
}

func ApplyMigrationTx(ctx context.Context, env Env, state State) error {
	tx := env.Tx
	if state.n == 0 {
		if err := state.change(ctx, env); err != nil {
			panic(err)
		}
	}
	rows, err := tx.QueryxContext(ctx, `SELECT CONCAT(table_schema, '.', table_name) AS schema_table_name 
		FROM information_schema.tables WHERE table_schema NOT IN ('pg_catalog', 'information_schema');`)
	if err != nil {
		return errors.Wrap(err, "getting tables to lock")
	}
	tables := make([]string, 0)
	for rows.Next() {
		table := ""
		if err := rows.Scan(&table); err != nil {
			return errors.Wrap(err, "scanning row into response")
		}
		tables = append(tables, table)
	}
	if err := env.LockTables(ctx, tables...); err != nil {
		return errors.Wrap(err, "locking all tables before migration")
	}
	if finished, err := isFinished(ctx, tx, state); err != nil {
		return err
	} else if finished {
		// skip migration
		msg := fmt.Sprintf("migration %d already applied", state.n)
		log.Info(ctx, msg) // avoid log rate limit
		return nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO public.migrations (id, name, start_time) VALUES ($1, $2, CURRENT_TIMESTAMP)`, state.n, state.name); err != nil {
		return errors.EnsureStack(err)
	}
	msg := fmt.Sprintf("applying migration %d: %s", state.n, state.name)
	log.Info(ctx, msg) // avoid log rate limit
	if err := state.change(ctx, env); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE public.migrations SET end_time = CURRENT_TIMESTAMP WHERE id = $1`, state.n); err != nil {
		return errors.EnsureStack(err)
	}
	msg = fmt.Sprintf("successfully applied migration %d", state.n)
	log.Info(ctx, msg) // avoid log rate limit
	return nil
}

// BlockUntil blocks until state is actualized.
// It makes no attempt to perform migrations, hopefully another process is working on that
// by calling ApplyMigrations.
// If the cluster ever enters a state newer than the state passed to BlockUntil, it errors.
func BlockUntil(ctx context.Context, db *pachsql.DB, state State) error {
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
			return errors.EnsureStack(err)
		}
		if tableExists {
			var latest int
			if err := db.GetContext(ctx, &latest, `SELECT COALESCE(MAX(id), 0) FROM public.migrations`); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return errors.EnsureStack(err)
			}
			if latest == state.n {
				return nil
			} else if latest > state.n {
				return errors.Errorf("database state is newer than application is expecting")
			}
		}
		select {
		case <-ctx.Done():
			return errors.EnsureStack(context.Cause(ctx))
		case <-ticker.C:
		}
	}
}

func isFinished(ctx context.Context, tx *pachsql.Tx, state State) (bool, error) {
	var name string
	if err := tx.GetContext(ctx, &name, `
	SELECT name
	FROM public.migrations
	WHERE id = $1
	`, state.n); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, errors.EnsureStack(err)
	}
	if name != state.name {
		return false, errors.Errorf("migration mismatch %d HAVE: %s WANT: %s", state.n, name, state.name)
	}
	return true, nil
}
