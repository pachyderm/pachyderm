package pjsdb

import (
	"context"
	"crypto/rand"
	"database/sql"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

var (
	//HashFn the hash function used by the pjsdb library for hashing job context tokens.
	// Its exposed for ease-of-access. This also allows it to be overridden in testing.
	HashFn = hashFn
)

// JobContextToken is a bearer token used throughout the PJS API to model a capabilities system that enables a PJS worker
// to act on behalf of a job and scope down its view of the job system to that job's view.
type JobContextToken []byte

// String prints out a token's hash.
func (j JobContextToken) String() string {
	return string(HashFn(j))
}

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

// JobContextTokenFromHex decodes the tokens that come through the API into JobContextToken objects.
func JobContextTokenFromHex(hex string) (JobContextToken, error) {
	token, err := pachhash.ParseHex([]byte(hex))
	if err != nil {
		return nil, errors.Wrap(err, "job context token from hash")
	}
	return token[:], nil
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

func generateToken() JobContextToken {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		panic(err) // the random generator can error if a mutex fails to unlock. Seems unlikely.
	}
	return token
}

func hashFn(b []byte) []byte {
	bHash := pachhash.Sum(b)
	return bHash[:]
}
