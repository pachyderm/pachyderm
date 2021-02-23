package track

import (
	"context"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
)

var _ Tracker = &postgresTracker{}

type postgresTracker struct {
	db *sqlx.DB
}

// NewPostgresTracker returns a
func NewPostgresTracker(db *sqlx.DB) Tracker {
	return &postgresTracker{db: db}
}

func (t *postgresTracker) DB() *sqlx.DB {
	return t.db
}

func (t *postgresTracker) CreateTx(tx *sqlx.Tx, id string, pointsTo []string, ttl time.Duration) error {
	// TODO: contraints on ttl? ttl = 0 has to be interpretted as no ttl, but all other values will make it through
	for _, dwn := range pointsTo {
		if dwn == id {
			return ErrSelfReference
		}
	}
	pointsTo = removeDuplicates(pointsTo)
	// if the object exists and has different downstream then error, otherwise return nil.
	exists, err := t.exists(tx, id)
	if err != nil {
		return err
	}
	if exists {
		dwn, err := t.getDownstream(tx, id)
		if err != nil {
			return err
		}
		if stringsMatch(pointsTo, dwn) {
			_, err := t.putObject(tx, id, ttl)
			return err
		}
		return ErrDifferentObjectExists
	}

	// create the object
	oid, err := t.putObject(tx, id, ttl)
	if err != nil {
		return err
	}
	var pointsToInts []int
	if err := tx.Select(&pointsToInts,
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
}

// putObject creates or updates the object at id, to have the max of the current and new ttl.
// If ttl == NoTTL, then the ttl is removed.
func (t *postgresTracker) putObject(tx *sqlx.Tx, id string, ttl time.Duration) (int, error) {
	var oid int
	if ttl != NoTTL {
		if err := tx.Get(&oid,
			`INSERT INTO storage.tracker_objects (str_id, expires_at)
			VALUES ($1, CURRENT_TIMESTAMP + $2 * interval '1 microsecond')
			ON CONFLICT (str_id) DO
				UPDATE SET expires_at = greatest(
					storage.tracker_objects.expires_at,
					(CURRENT_TIMESTAMP + $2 * interval '1 microsecond')
				)
				WHERE storage.tracker_objects.str_id = $1
			RETURNING int_id
		`, id, ttl.Microseconds()); err != nil {
			return 0, err
		}
	} else {
		if err := tx.Get(&oid,
			`INSERT INTO storage.tracker_objects (str_id)
			VALUES ($1)
			ON CONFLICT (str_id) DO
				UPDATE SET expires_at = NULL
				WHERE storage.tracker_objects.str_id = $1
			RETURNING int_id
		`, id); err != nil {
			return 0, err
		}
	}
	return oid, nil
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
	var dwn []string
	if err := dbutil.WithTx(ctx, t.db, func(tx *sqlx.Tx) error {
		var err error
		dwn, err = t.getDownstream(tx, id)
		return err
	}); err != nil {
		return nil, err
	}
	return dwn, nil
}

func (t *postgresTracker) getDownstream(tx *sqlx.Tx, id string) ([]string, error) {
	dwn := []string{}
	if err := tx.Select(&dwn,
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

func (t *postgresTracker) DeleteTx(tx *sqlx.Tx, id string) error {
	var count int
	if err := t.db.Get(&count, `
		WITH target AS (
			SELECT int_id FROM storage.tracker_objects WHERE str_id = $1
		)
		SELECT count(distinct from_id) FROM storage.tracker_refs WHERE to_id IN (SELECT int_id FROM TARGET)
	`, id); err != nil {
		return err
	}
	if count > 0 {
		return ErrDanglingRef
	}
	_, err := t.db.Exec(`
		WITH target AS (
			SELECT int_id FROM storage.tracker_objects WHERE str_id = $1
		)
		DELETE FROM storage.tracker_refs WHERE from_id IN (SELECT int_id FROM TARGET)
	`, id)
	if err != nil {
		return err
	}
	_, err = t.db.Exec(`DELETE FROM storage.tracker_objects WHERE str_id = $1`, id)
	return err
}

func (t *postgresTracker) IterateDeletable(ctx context.Context, cb func(id string) error) (retErr error) {
	rows, err := t.db.QueryxContext(ctx,
		`SELECT str_id FROM storage.tracker_objects
		WHERE int_id NOT IN (SELECT to_id FROM storage.tracker_refs)
		AND expires_at <= CURRENT_TIMESTAMP`)
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

func (t *postgresTracker) exists(tx *sqlx.Tx, id string) (bool, error) {
	var count int
	if err := tx.Get(&count, `SELECT count(*) FROM storage.tracker_objects WHERE str_id = $1`, id); err != nil {
		return false, err
	}
	return count > 0, nil
}

func stringsMatch(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func removeDuplicates(xs []string) []string {
	sort.Strings(xs)
	var countDeleted int
	for i := range xs {
		if i > 0 && xs[i] == xs[i-1] {
			countDeleted++
		} else {
			xs[i-countDeleted] = xs[i]
		}
	}
	return xs[:len(xs)-countDeleted]
}

// SetupPostgresTrackerV0 sets up the table for the postgres tracker
func SetupPostgresTrackerV0(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, schema)
	return err
}

var schema = `
	CREATE TABLE storage.tracker_objects (
		int_id BIGSERIAL PRIMARY KEY,
		str_id VARCHAR(4096) UNIQUE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP
	);

	CREATE TABLE storage.tracker_refs (
		from_id INT8 NOT NULL,
		to_id INT8 NOT NULL,
		PRIMARY KEY (from_id, to_id)
	);

	CREATE INDEX ON storage.tracker_refs (
		to_id,
		from_id
	);
`
