package migrations

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
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
	return State{}
}

func ApplyMigrations(ctx context.Context, db *sqlx.DB, state *State) error {
	if state.prev == nil {
		// check migrations table, skip if applied

		return nil
	} else {
		ApplyMigrations(ctx, db, state.prev)
	}
}

func BlockUntil(ctx context.Context, db *sqlx.DB, state *State) error {
	// poll database until this state is registered
	return nil
}

var DesiredState 