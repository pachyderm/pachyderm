package dbutil

import (
	"strings"

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

// IsErrDatabaseConnection returns true if the error occurs during database connection flakiness
func IsDatabaseDisconnect(err error) bool {
	return strings.Contains(err.Error(), "broken pipe") || strings.Contains(err.Error(), "unexpected EOF")
}
