package tracker

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

var _ Tracker = &PGTracker{}

type PGTracker struct {
	db *sqlx.DB
}

func NewPGTracker(db *sqlx.DB) *PGTracker {
	return &PGTracker{db: db}
}

func (t *PGTracker) CreateObject(ctx context.Context, id string, pointsTo []string, ttl time.Duration) error {
	panic("not implemented")
}

func (t *PGTracker) SetTTLPrefix(ctx context.Context, prefix string, ttl time.Duration) (time.Time, error) {
	panic("not implemented")
}

func (t *PGTracker) GetDownstream(ctx context.Context, id string) ([]string, error) {
	panic("not implemented")
}

func (t *PGTracker) GetUpstream(ctx context.Context, id string) ([]string, error) {
	panic("not implemented")
}

func (t *PGTracker) MarkTombstone(ctx context.Context, id string) error {
	panic("not implemented")
}

func (t *PGTracker) DeleteObject(ctx context.Context, id string) error {
	panic("not implemented")
}

func (t *PGTracker) IterateExpired(ctx context.Context, cb func(id string) error) error {
	panic("not implemented")
}

var schema = `
	CREATE TABLE storage.tracker_objects (
		int_id BIGSERIAL PRIMARY KEY,
		str_id VARCHAR(250) PRIMARY KEY,
		tombstone BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP
	);

	CREATE TABLE storage.tracker_refs (
		from INT8 NOT NULL,
		to INT8 NOT NULL,
		PRIMARY KEY (from, to)
	);	
`
