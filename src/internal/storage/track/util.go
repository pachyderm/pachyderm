package track

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
)

// Create creates uses tracker to create the object id.
func Create(ctx context.Context, tr Tracker, id string, pointsTo []string, ttl time.Duration) error {
	return dbutil.WithTx(ctx, tr.DB(), func(tx *sqlx.Tx) error {
		return tr.CreateTx(tx, id, pointsTo, ttl)
	})
}

// Delete deletes id from the tracker
func Delete(ctx context.Context, tr Tracker, id string) error {
	return dbutil.WithTx(ctx, tr.DB(), func(tx *sqlx.Tx) error {
		return tr.DeleteTx(tx, id)
	})
}

// Drop sets the object at id to expire now
func Drop(ctx context.Context, tr Tracker, id string) error {
	_, err := tr.SetTTL(ctx, id, ExpireNow)
	return err
}

// DropPrefix sets all objects with prefix to expire now.
// It returns the number of objects affected or an error.
func DropPrefix(ctx context.Context, tr Tracker, prefix string) (int, error) {
	_, n, err := tr.SetTTLPrefix(ctx, prefix, ExpireNow)
	return n, err
}
