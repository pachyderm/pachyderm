package migrations

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Env contains all the objects that can be manipulated during a migration.
// The Tx field will be overwritten with the transaction that the migration should be performed in.
type Env struct {
	// TODO: etcd
	ObjectClient obj.Client
	Tx           *pachsql.Tx
	EtcdClient   *clientv3.Client
}

// MakeEnv returns a new Env
// The only advantage to this contructor is you can be sure all the fields are set.
// You can also create an Env using a struct literal.
func MakeEnv(objC obj.Client, etcdC *clientv3.Client) Env {
	return Env{
		ObjectClient: objC,
		EtcdClient:   etcdC,
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
}

// Apply applies a Func to the state and returns a new state.
func (s State) Apply(name string, fn Func) State {
	return State{
		prev:   &s,
		change: fn,
		n:      s.n + 1,
		name:   strings.ToLower(name),
	}
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
			_, err := env.Tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS migrations (
				id BIGINT PRIMARY KEY,
				NAME VARCHAR(250) NOT NULL,
				start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				end_time TIMESTAMP
			);
			INSERT INTO migrations (id, name, start_time, end_time) VALUES (0, 'init', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING;
			`)
			return errors.EnsureStack(err)
		},
	}
}

// ApplyMigrations does the necessary work to actualize state.
// It will manipulate the objects available in baseEnv, and use the migrations table in db.
func ApplyMigrations(ctx context.Context, db *pachsql.DB, baseEnv Env, state State) error {
	for _, state := range collectStates(make([]State, 0, state.n+1), state) {
		if err := applyMigration(ctx, db, baseEnv, state); err != nil {
			return err
		}
	}
	return nil
}

// collectStates does a reverse order traversal of a linked list and adds each item to a slice
func collectStates(slice []State, s State) []State {
	if s.prev != nil {
		slice = collectStates(slice, *s.prev)
	}
	return append(slice, s)
}

func applyMigration(ctx context.Context, db *pachsql.DB, baseEnv Env, state State) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.EnsureStack(err)
	}
	env := baseEnv
	env.Tx = tx
	if err := func() error {
		if state.n == 0 {
			if err := state.change(ctx, env); err != nil {
				panic(err)
			}
		}
		_, err := tx.ExecContext(ctx, `LOCK TABLE migrations IN EXCLUSIVE MODE`)
		if err != nil {
			return errors.EnsureStack(err)
		}
		if finished, err := isFinished(ctx, tx, state); err != nil {
			return err
		} else if finished {
			// skip migration
			logrus.Infof("migration %d already applied", state.n)
			return nil
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO migrations (id, name, start_time) VALUES ($1, $2, CURRENT_TIMESTAMP)`, state.n, state.name); err != nil {
			return errors.EnsureStack(err)
		}
		logrus.Infof("applying migration %d %s", state.n, state.name)
		if err := state.change(ctx, env); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE migrations SET end_time = CURRENT_TIMESTAMP WHERE id = $1`, state.n); err != nil {
			return errors.EnsureStack(err)
		}
		logrus.Infof("successfully applied migration %d", state.n)
		return nil
	}(); err != nil {
		if err := tx.Rollback(); err != nil {
			logrus.Error(err)
		}
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(tx.Commit())
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
			if err := db.GetContext(ctx, &latest, `SELECT COALESCE(MAX(id), 0) FROM migrations`); err != nil && !errors.Is(err, sql.ErrNoRows) {
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
			return errors.EnsureStack(ctx.Err())
		case <-ticker.C:
		}
	}
}

func isFinished(ctx context.Context, tx *pachsql.Tx, state State) (bool, error) {
	var name string
	if err := tx.GetContext(ctx, &name, `
	SELECT name
	FROM migrations
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
