package dbutil

import (
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// IsUniqueViolation returns true if the error is a UniqueContraintViolation
func IsUniqueViolation(err error) bool {
	pgErr := &pgconn.PgError{}
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.UniqueViolation
	}
	return false
}
