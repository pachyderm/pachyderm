package pjsdb

import (
	"context"
	"database/sql"
	"github.com/google/uuid"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

var (
	//HashFn the hash function used by the pjsdb library for hashing job context tokens.
	HashFn = hashFn
)

type JobContextToken []byte

type JobContext struct {
	Token JobContextToken
	Hash  []byte `db:"context_hash"`
}

// CreateJobContext generates a JobContext token and stores the hash in postgres. The token is returned to the caller.
// This token will be generated at-most-once per job. It is used to model a capabilities system.
func CreateJobContext(ctx context.Context, tx *pachsql.Tx, id JobID) (JobContext, error) {
	token := generateToken()
	jobCtx := JobContext{
		Token: token,
		Hash:  HashFn(token),
	}
	res, err := tx.ExecContext(ctx, `UPDATE pjs.jobs SET context_hash = $1 WHERE id = $2 AND context_hash IS NULL`,
		jobCtx.Hash, id)
	if err != nil {
		return JobContext{}, errors.Wrap(err, "create job context")
	}
	numRows, err := res.RowsAffected()
	if err != nil {
		return JobContext{}, errors.Wrap(err, "create job context")
	}
	if numRows == 0 {
		return JobContext{}, ErrJobContextAlreadyExists
	}
	return jobCtx, nil
}

// ResolveJobContext takes a JobContext and returns the associated job. The PJS gRPC API requests contain JobContext
// tokens rather than JobID fields, which necessitates this resolution.
func ResolveJobContext(ctx context.Context, tx *pachsql.Tx, token JobContextToken) (JobID, error) {
	var id JobID
	hash := HashFn(token)
	err := tx.GetContext(ctx, &id, `SELECT id FROM pjs.jobs WHERE context_hash = $1`, hash)
	if err != nil {
		if errors.As(err, sql.ErrNoRows) {
			return 0, sql.ErrNoRows
		}
		return 0, errors.Wrap(err, "resolve job context")
	}
	return id, nil
}

// RevokeJobContext revokes a JobContext token by deleting its hash from postgres.
// It returns an error if there were no tokens to revoke.
func RevokeJobContext(ctx context.Context, tx *pachsql.Tx, id JobID) error {
	res, err := tx.ExecContext(ctx, `UPDATE pjs.jobs SET context_hash = NULL WHERE id = $1 AND context_hash IS NOT NULL`, id)
	if err != nil {
		return errors.Wrap(err, "revoke job context")
	}
	numRows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "revoke job context")
	}
	if numRows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func generateToken() []byte {
	uuidv7, err := uuid.NewV7()
	if err != nil {
		panic(err) // the uuid generation can error if the random generator's mutex fails to unlock. Seems unlikely.
	}
	return []byte(uuidv7.String())
}

func hashFn(b []byte) []byte {
	return []byte(pachhash.EncodeHash(b))
}
