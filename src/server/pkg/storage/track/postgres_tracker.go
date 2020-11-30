package track

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var _ Tracker = &postgresTracker{}

type postgresTracker struct {
	db *sqlx.DB
}

// NewPostgresTracker returns a
func NewPostgresTracker(db *sqlx.DB) Tracker {
	return &postgresTracker{db: db}
}

func (t *postgresTracker) CreateObject(ctx context.Context, id string, pointsTo []string, ttl time.Duration) error {
	// TODO: contraints on ttl? ttl = 0 has to be interpretted as no ttl, but all other values will make it through
	for _, dwn := range pointsTo {
		if dwn == id {
			return ErrSelfReference
		}
	}
	return t.withTx(ctx, func(tx *sqlx.Tx) error {
		var oid int
		if err := func() error {
			if ttl > 0 {
				return tx.GetContext(ctx, &oid,
					`INSERT INTO storage.tracker_objects (str_id, expires_at)
				VALUES ($1, CURRENT_TIMESTAMP + $2 * interval '1 microsecond')
				ON CONFLICT (str_id) DO NOTHING
				RETURNING int_id
				`, id, ttl.Microseconds())
			}
			return tx.GetContext(ctx, &oid,
				`INSERT INTO storage.tracker_objects (str_id)
				VALUES ($1)
				ON CONFLICT (str_id) DO NOTHING
				RETURNING int_id
				`, id)
		}(); err != nil {
			if err == sql.ErrNoRows {
				err = ErrObjectExists
			}
			return err
		}
		var pointsToInts []int
		if err := tx.SelectContext(ctx, &pointsToInts,
			`INSERT INTO storage.tracker_refs (from_id, to_id)
			SELECT $1, int_id FROM storage.tracker_objects WHERE str_id = ANY($2)
			RETURNING to_id`,
			oid, pq.StringArray(pointsTo)); err != nil {
			return err
		}
		if len(pointsToInts) != len(pointsTo) {
			return ErrDanglingRef
		}
		return nil
	})
}

func (t *postgresTracker) SetTTLPrefix(ctx context.Context, prefix string, ttl time.Duration) (time.Time, error) {
	var expiresAt time.Time
	err := t.db.GetContext(ctx, &expiresAt,
		`UPDATE storage.tracker_objects
		SET expires_at = CURRENT_TIMESTAMP + $2 * interval '1 microsecond'
		WHERE str_id LIKE $1 || '%'
		RETURNING expires_at`, prefix, ttl.Microseconds())
	if err != nil {
		return time.Time{}, err
	}
	return expiresAt, nil
}

func (t *postgresTracker) GetDownstream(ctx context.Context, id string) ([]string, error) {
	dwn := []string{}
	if err := t.db.SelectContext(ctx, &dwn,
		`WITH target AS (
			SELECT int_id FROM storage.tracker_objects WHERE str_id = $1
		)
		SELECT str_id
		FROM storage.tracker_objects
		WHERE int_id IN (
			SELECT to_id FROM storage.tracker_refs WHERE from_id IN (SELECT int_id FROM target)
		)`, id); err != nil {
		return nil, err
	}
	return dwn, nil
}

func (t *postgresTracker) GetUpstream(ctx context.Context, id string) ([]string, error) {
	ups := []string{}
	if err := t.db.SelectContext(ctx, &ups,
		`WITH target AS (
			SELECT int_id FROM storage.tracker_objects WHERE str_id = $1
		)
		SELECT str_id
		FROM storage.tracker_objects
		WHERE int_id IN (
			SELECT from_id FROM storage.tracker_refs WHERE to_id IN (SELECT int_id FROM TARGET)
		)`, id); err != nil {
		return nil, err
	}
	return ups, nil
}

func (t *postgresTracker) MarkTombstone(ctx context.Context, id string) error {
	var tombstones []bool
	if err := t.db.SelectContext(ctx, &tombstones, `	
		UPDATE storage.tracker_objects
		SET tombstone = (	
			CASE
				WHEN NOT EXISTS (
					SELECT from_id FROM storage.tracker_refs
					WHERE to_id IN (
						SELECT int_id FROM storage.tracker_objects
						WHERE str_id = $1
					)
				) THEN TRUE
				ELSE FALSE
			END
		)
		WHERE str_id = $1
		RETURNING tombstone
	`, id); err != nil {
		return err
	}
	// if we get no results back, then it doesn't exist
	if len(tombstones) == 0 {
		return nil
	}
	// if the tombstone is not set, then it was because it would create a dangling ref
	if !tombstones[0] {
		return ErrDanglingRef
	}
	return nil
}

func (t *postgresTracker) FinishDelete(ctx context.Context, id string) error {
	err := t.withTx(ctx, func(tx *sqlx.Tx) error {
		var tombstone bool
		if err := tx.GetContext(ctx, &tombstone,
			`DELETE FROM storage.tracker_objects
			WHERE str_id = $1
			RETURNING tombstone
			`, id); err != nil {
			return err
		}
		if !tombstone {
			return ErrNotTombstone
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM storage.tracker_refs WHERE from_id IN (
			SELECT int_id FROM storage.tracker_objects WHERE str_id = $1
		)
		`, id); err != nil {
			return err
		}
		return nil
	})
	if err == sql.ErrNoRows {
		return nil
	}
	return nil
}

func (t *postgresTracker) IterateDeletable(ctx context.Context, cb func(id string) error) (retErr error) {
	rows, err := t.db.QueryxContext(ctx,
		`SELECT str_id FROM storage.tracker_objects
		WHERE int_id NOT IN (SELECT to_id FROM storage.tracker_refs)
		AND (expires_at <= CURRENT_TIMESTAMP OR tombstone)`)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); retErr == nil {
			retErr = err
		}
	}()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := cb(id); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (t *postgresTracker) withTx(ctx context.Context, cb func(tx *sqlx.Tx) error) error {
	tx, err := t.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	if err := cb(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// SetupPostgresTracker sets up the table for the postgres tracker
func SetupPostgresTracker(db *sqlx.DB) {
	db.MustExec(schema)
}

var schema = `
	CREATE TABLE storage.tracker_objects (
		int_id BIGSERIAL PRIMARY KEY,
		str_id VARCHAR(4096) UNIQUE,
		tombstone BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP
	);

	CREATE TABLE storage.tracker_refs (
		from_id INT8 NOT NULL,
		to_id INT8 NOT NULL,
		PRIMARY KEY (from_id, to_id)
	);	
`
