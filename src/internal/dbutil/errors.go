package dbutil

import (
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

// IsUniqueViolation returns true if the error is a UniqueContraintViolation
func IsUniqueViolation(err error) bool {
	pgErr := &pgconn.PgError{}
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.UniqueViolation
	}
	return false
}
