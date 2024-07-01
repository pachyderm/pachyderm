package pjsdb

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"github.com/pkg/errors"
)

type JobID uint64

func CreateJob(ctx context.Context, tx *pachsql.Tx, job *pjs.JobInfo) (JobID, error) {
	var id JobID
	hasher := pachhash.New()
	hash := hasher.Sum(job.Spec.Value)
	row := tx.QueryRowxContext(ctx, `
		INSERT INTO pjs.jobs (input, spec, spec_hash, parent) VALUES 
		($1, $2, $3, $4) RETURNING id`, job.Input, job.Spec)
	if err := row.Scan(&id); err != nil {
		return 0, errors.Wrap(err, "scanning id from create commitInfo")
	}

}
