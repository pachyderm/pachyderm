package v2_10_0

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/protobuf/proto"
)

var createUniqueIndex = "CREATE UNIQUE INDEX pip_version_idx ON collections.pipelines(idx_name,idx_version)"

// collect all pipeline versions for pipelines with duplicate versions
var duplicatePipelinesQuery = `
SELECT key, idx_version, proto
  FROM collections.pipelines
  WHERE idx_name
    IN (SELECT a.idx_name
        FROM collections.pipelines a
          INNER JOIN collections.pipelines b
          ON a.idx_name = b.idx_name
           AND a.idx_version = b.idx_version
           AND a.key != b.key
       )
  ORDER BY idx_name, createdat;
`

var listJobInfos = `
  SELECT key, proto
  FROM collections.jobs
`

type pipRow struct {
	Key        string `db:"key"`
	IdxVersion uint64 `db:"idx_version"`
	Proto      []byte `db:"proto"`
}

type jobRow struct {
	Key   string `db:"key"`
	Proto []byte `db:"proto"`
}

func deduplicatePipelineVersions(ctx context.Context, env migrations.Env) error {
	pipUpdates, pipVersionChanges, err := collectPipelineUpdates(ctx, env.Tx)
	if err != nil {
		return err
	}
	if len(pipUpdates) != 0 {
		var pipValues string
		for _, u := range pipUpdates {
			pipValues += fmt.Sprintf(" ('%s', %v, decode('%v', 'hex')),", u.Key, u.IdxVersion, hex.EncodeToString(u.Proto))
		}
		pipValues = pipValues[:len(pipValues)-1]
		stmt := fmt.Sprintf(`
                 UPDATE collections.pipelines AS p SET
                   p.idx_version = v.idx_version,
                   p.proto = v.proto
                 FROM (VALUES%s) AS v(key, idx_version, proto)
                 WHERE p.key = v.key;`, pipValues)
		if _, err := env.Tx.ExecContext(ctx, stmt); err != nil {
			return errors.Wrapf(err, "update pipeline rows statement: %v", stmt)
		}
	}
	jobUpdates, err := collectJobUpdates(ctx, env.Tx, pipVersionChanges)
	if err != nil {
		return err
	}
	if len(jobUpdates) != 0 {
		var jobValues string
		for _, u := range jobUpdates {
			jobValues += fmt.Sprintf(" ('%s', decode('%v', 'hex')),", u.Key, hex.EncodeToString(u.Proto))
		}
		jobValues = jobValues[:len(jobValues)-1]
		stmt := fmt.Sprintf(`
                 UPDATE collections.jobs AS j SET
                   j.proto = v.proto
                 FROM (VALUES%s) AS v(key, proto)
                 WHERE j.key = v.key;`, jobValues)
		if _, err := env.Tx.ExecContext(ctx, stmt); err != nil {
			return errors.Wrapf(err, "update job rows statement: %v", stmt)
		}
	}
	if _, err := env.Tx.ExecContext(ctx, createUniqueIndex); err != nil {
		return errors.Wrap(err, "create unique index pip_version_idx")
	}
	return nil
}

func collectPipelineUpdates(ctx context.Context, tx *pachsql.Tx) (rowUpdates []*pipRow,
	pipelineVersionChanges map[string]map[uint64]uint64,
	retErr error) {
	rr, err := tx.QueryxContext(ctx, duplicatePipelinesQuery)
	if err != nil {
		return nil, nil, errors.Wrap(err, "query duplicates in collections.pipelines")
	}
	defer rr.Close()
	pi := &pps.PipelineInfo{}
	pipLatestVersion := make(map[string]uint64)
	pipVersionChanges := make(map[string]map[uint64]uint64)
	var updates []*pipRow
	for rr.Next() {
		var row pipRow
		if err := rr.Err(); err != nil {
			return nil, nil, errors.Wrap(err, "row error")
		}
		if err := rr.StructScan(&row); err != nil {
			return nil, nil, errors.Wrap(err, "scan pipeline row")
		}
		if err := proto.Unmarshal(row.Proto, pi); err != nil {
			return nil, nil, errors.Wrapf(err, "unmarshal proto")
		}
		lastChange, ok := pipLatestVersion[pi.Pipeline.Name]
		if !ok {
			lastChange = 0
		}
		currVersion := pi.Version
		correctVersion := lastChange + 1
		if pi.Version != correctVersion {
			pi.Version = correctVersion
			data, err := proto.Marshal(pi)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "marshal pipeline info %v", pi)
			}
			updates = append(updates, &pipRow{Key: row.Key, IdxVersion: correctVersion, Proto: data})
		}
		pipLatestVersion[pi.Pipeline.Name] = correctVersion
		changes, ok := pipVersionChanges[pi.Pipeline.Name]
		if !ok {
			changes = make(map[uint64]uint64)
			pipVersionChanges[pi.Pipeline.Name] = changes
		}
		changes[currVersion] = correctVersion
	}
	return updates, pipVersionChanges, nil
}

func collectJobUpdates(ctx context.Context, tx *pachsql.Tx, pipVersionChanges map[string]map[uint64]uint64) ([]*jobRow, error) {
	rr, err := tx.QueryxContext(ctx, listJobInfos)
	if err != nil {
		return nil, errors.Wrap(err, "list jobs")
	}
	defer rr.Close()
	ji := &pps.JobInfo{}
	var updates []*jobRow
	for rr.Next() {
		var row jobRow
		if err := rr.Err(); err != nil {
			return nil, errors.Wrap(err, "row error")
		}
		if err := rr.StructScan(&row); err != nil {
			return nil, errors.Wrap(err, "scan job row")
		}
		if err := proto.Unmarshal(row.Proto, ji); err != nil {
			return nil, errors.Wrapf(err, "unmarshal proto")
		}
		if changes, ok := pipVersionChanges[ji.Job.Pipeline.String()]; ok {
			if new, ok := changes[ji.PipelineVersion]; ok {
				ji.PipelineVersion = new
				data, err := proto.Marshal(ji)
				if err != nil {
					return nil, errors.Wrapf(err, "marshal job info %v", ji)
				}
				updates = append(updates, &jobRow{Key: row.Key, Proto: data})
			}
		}
	}
	return updates, nil
}
