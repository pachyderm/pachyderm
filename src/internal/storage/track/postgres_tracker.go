package track

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

var _ Tracker = &postgresTracker{}

type postgresTracker struct {
	db *pachsql.DB
}

// NewPostgresTracker returns a
func NewPostgresTracker(db *pachsql.DB) Tracker {
	return &postgresTracker{db: db}
}

func (t *postgresTracker) DB() *pachsql.DB {
	return t.db
}

func (t *postgresTracker) CreateTx(tx *pachsql.Tx, id string, pointsTo []string, ttl time.Duration) error {
	for _, dwn := range pointsTo {
		if dwn == id {
			return ErrSelfReference
		}
	}
	pointsTo = dedupedStrings(pointsTo)
	// create an object or update the ttl of an existing one
	intID, created, err := t.putObject(tx, id, ttl)
	if err != nil {
		return err
	}
	if !created {
		dwn, err := t.getDownstream(tx, intID)
		if err != nil {
			return err
		}
		if !stringsMatch(pointsTo, dwn) {
			return ErrDifferentObjectExists
		}
		return nil
	}
	return t.addReferences(tx, intID, pointsTo)
}

// putObject creates or updates the object at id, to have the max of the current and new ttl.
// If ttl == NoTTL, then the ttl is removed.
func (t *postgresTracker) putObject(tx *pachsql.Tx, id string, ttl time.Duration) (int, bool, error) {
	// About xmax https://stackoverflow.com/a/39204667
	res := struct {
		IntID int `db:"int_id"`
		XMax  int `db:"xmax"`
	}{}
	if ttl != NoTTL {
		if err := tx.Get(&res,
			`INSERT INTO storage.tracker_objects (str_id, expires_at)
			VALUES ($1, CURRENT_TIMESTAMP + $2 * interval '1 microsecond')
			ON CONFLICT (str_id) DO
				UPDATE SET expires_at = greatest(
					storage.tracker_objects.expires_at,
					(CURRENT_TIMESTAMP + $2 * interval '1 microsecond')
				)
				WHERE storage.tracker_objects.str_id = $1
			RETURNING int_id, xmax
		`, id, ttl.Microseconds()); err != nil {
			return 0, false, errors.EnsureStack(err)
		}
	} else {
		if err := tx.Get(&res,
			`INSERT INTO storage.tracker_objects (str_id)
			VALUES ($1)
			ON CONFLICT (str_id) DO
				UPDATE SET expires_at = NULL
				WHERE storage.tracker_objects.str_id = $1
			RETURNING int_id, xmax
		`, id); err != nil {
			return 0, false, errors.EnsureStack(err)
		}
	}
	inserted := res.XMax == 0
	return res.IntID, inserted, nil
}

func (t *postgresTracker) addReferences(tx *pachsql.Tx, intID int, pointsTo []string) error {
	if len(pointsTo) == 0 {
		return nil
	}
	var pointsToInts []int
	if err := tx.Select(&pointsToInts,
		`INSERT INTO storage.tracker_refs (from_id, to_id)
			SELECT $1, int_id FROM storage.tracker_objects WHERE str_id = ANY($2)
		RETURNING to_id`,
		intID, pointsTo); err != nil {
		return errors.EnsureStack(err)
	}
	if len(pointsToInts) != len(pointsTo) {
		return ErrDanglingRef
	}
	return nil
}

func (t *postgresTracker) SetTTL(ctx context.Context, id string, ttl time.Duration) (time.Time, error) {
	var expiresAt time.Time
	err := t.db.GetContext(ctx, &expiresAt,
		`UPDATE storage.tracker_objects
		SET expires_at = CURRENT_TIMESTAMP + $2 * interval '1 microsecond'
		WHERE str_id = $1
		RETURNING expires_at
	`, id, ttl.Microseconds())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = pacherr.NewNotExist("tracker", id)
		}
		return time.Time{}, err
	}
	return expiresAt, nil
}

func (t *postgresTracker) SetTTLPrefix(ctx context.Context, prefix string, ttl time.Duration) (time.Time, int, error) {
	var x struct {
		Count     int       `db:"count"`
		ExpiresAt time.Time `db:"expires_at"`
	}
	err := t.db.GetContext(ctx, &x,
		`WITH rows AS (
			UPDATE storage.tracker_objects
			SET expires_at = CURRENT_TIMESTAMP + $2 * interval '1 microsecond'
			WHERE str_id LIKE $1 || '%'
			RETURNING expires_at
		)
		SELECT COUNT(*) as count, COALESCE(MIN(expires_at), CURRENT_TIMESTAMP + $2 * interval '1 microsecond') as expires_at FROM rows
		`, prefix, ttl.Microseconds())
	if err != nil {
		return time.Time{}, 0, errors.EnsureStack(err)
	}
	return x.ExpiresAt, x.Count, nil
}

func (t *postgresTracker) GetDownstream(ctx context.Context, id string) ([]string, error) {
	var dwn []string
	if err := t.db.SelectContext(ctx, &dwn, `
		WITH target AS (
			SELECT int_id FROM storage.tracker_objects WHERE str_id = $1
		)
		SELECT str_id
		FROM storage.tracker_objects
		WHERE int_id IN (
			SELECT to_id FROM storage.tracker_refs WHERE from_id IN (SELECT int_id FROM target)
		)
	`, id); err != nil {
		return nil, errors.EnsureStack(err)
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
		return nil, errors.EnsureStack(err)
	}
	return ups, nil
}

func (t *postgresTracker) GetExpiresAt(ctx context.Context, id string) (time.Time, error) {
	var expiresAt time.Time
	if err := t.db.GetContext(ctx, &expiresAt,
		`SELECT expires_at FROM storage.tracker_objects WHERE str_id = $1
	`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, pacherr.NewNotExist("tracker", id)
		}
		return time.Time{}, errors.EnsureStack(err)
	}
	return expiresAt, nil
}

func (t *postgresTracker) DeleteTx(tx *pachsql.Tx, id string) error {
	var count int
	if err := tx.Get(&count, `
		WITH target AS (
			SELECT int_id FROM storage.tracker_objects WHERE str_id = $1
		)
		SELECT count(distinct from_id) FROM storage.tracker_refs WHERE to_id IN (SELECT int_id FROM target)
	`, id); err != nil {
		return errors.EnsureStack(err)
	}
	if count > 0 {
		return ErrDanglingRef
	}
	_, err := tx.Exec(`
		WITH target AS (
			SELECT int_id FROM storage.tracker_objects WHERE str_id = $1
		)
		DELETE FROM storage.tracker_refs WHERE from_id IN (SELECT int_id FROM target)
	`, id)
	if err != nil {
		return errors.EnsureStack(err)
	}
	_, err = tx.Exec(`DELETE FROM storage.tracker_objects WHERE str_id = $1`, id)
	return errors.EnsureStack(err)
}

func (t *postgresTracker) IterateDeletable(ctx context.Context, cb func(id string) error) (retErr error) {
	var toDelete []string
	// select 1 in inner query as we don't actually care about the results, just existence
	// set arbitrary limit to guarantee we can iterate, doesn't matter for GC as we run this repeatedly
	err := t.db.SelectContext(ctx, &toDelete,
		`SELECT str_id FROM storage.tracker_objects as objs
		WHERE NOT EXISTS (SELECT 1 FROM storage.tracker_refs as refs where objs.int_id = refs.to_id)
		AND expires_at <= CURRENT_TIMESTAMP LIMIT 10000`)
	if err != nil {
		return errors.EnsureStack(err)
	}

	for _, id := range toDelete {
		if err := cb(id); err != nil {
			return err
		}
	}
	return nil
}

func (t *postgresTracker) getDownstream(tx *pachsql.Tx, intID int) ([]string, error) {
	dwn := []string{}
	if err := tx.Select(&dwn, `
		SELECT str_id FROM storage.tracker_objects
		JOIN storage.tracker_refs ON int_id = to_id
		WHERE from_id = $1
	`, intID); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return dwn, nil
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

func dedupedStrings(xs []string) []string {
	ys := append([]string{}, xs...)
	return removeDuplicates(ys)
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
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func SetupPostgresTrackerV0(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, schema)
	return errors.EnsureStack(err)
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
