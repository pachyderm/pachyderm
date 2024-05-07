package v2_10_0

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var CreateUniqueIndex = `
  CREATE UNIQUE INDEX pip_version_idx ON collections.pipelines(idx_version);
`

// find all pipelines with duplicate versions, and return all of the pipeline versions for each of those pipelines
var duplicatePipelinesQuery = `
  SELECT key, proto
  FROM collections.pipelines
  WHERE idx_name
    IN (
      SELECT idx_name
      FROM collections.pipelines
      GROUP BY idx_name, idx_version
      HAVING COUNT(key) > 1
    )
  ORDER BY idx_name, createdat;
`

func DeduplicatePipelineVersions(ctx context.Context, env migrations.Env) error {
	pipUpdates, pipVersionChanges, err := collectPipelineUpdates(ctx, env.Tx)
	if err != nil {
		return err
	}
	log.Info(ctx, "detected updates for pipeline versions", zap.Int("updates_count", len(pipUpdates)))
	if err := UpdatePipelineRows(ctx, env.Tx, pipUpdates); err != nil {
		return err
	}
	if err := UpdateJobPipelineVersions(ctx, env.Tx, pipVersionChanges); err != nil {
		return err
	}
	if _, err := env.Tx.ExecContext(ctx, CreateUniqueIndex); err != nil {
		return errors.Wrap(err, "create unique index pip_version_idx")
	}
	return nil
}

func collectPipelineUpdates(ctx context.Context, tx *pachsql.Tx) (rowUpdates []*PipUpdateRow,
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
	var updates []*PipUpdateRow
	for rr.Next() {
		var row pipDBRow
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
			project, pipeline, _, err := parsePipelineKey(row.Key)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "parse key %q", row.Key)
			}
			idxVersion := VersionKey(project, pipeline, correctVersion)
			updates = append(updates, &PipUpdateRow{Key: row.Key, IdxVersion: idxVersion, Proto: data})
		}
		pipLatestVersion[pi.Pipeline.Name] = correctVersion
		changes, ok := pipVersionChanges[pi.Pipeline.Name]
		if !ok {
			changes = make(map[uint64]uint64)
			pipVersionChanges[pi.Pipeline.Name] = changes
		}
		changes[currVersion] = correctVersion
	}
	if err := rr.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "row error")
	}
	return updates, pipVersionChanges, nil
}

func UpdateJobPipelineVersions(ctx context.Context, tx *pachsql.Tx, pipVersionChanges map[string]map[uint64]uint64) error {
	jobUpdates, err := collectJobUpdates(ctx, tx, pipVersionChanges)
	if err != nil {
		return err
	}
	if err := updateJobRows(ctx, tx, jobUpdates); err != nil {
		return err
	}
	return nil
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
	if err := rr.Err(); err != nil {
		return nil, errors.Wrap(err, "row error")
	}
	return updates, nil
}

func UpdatePipelineRows(ctx context.Context, tx *pachsql.Tx, pipUpdates []*PipUpdateRow) error {
	if len(pipUpdates) == 0 {
		return nil
	}
	var pipValues string
	for _, u := range pipUpdates {
		pipValues += fmt.Sprintf(" ('%s', '%s', decode('%v', 'hex')),", u.Key, u.IdxVersion, hex.EncodeToString(u.Proto))
	}
	pipValues = pipValues[:len(pipValues)-1]
	stmt := fmt.Sprintf(`
                 UPDATE collections.pipelines AS p SET
                   idx_version = v.idx_version,
                   proto = v.proto
                 FROM (VALUES%s) AS v(key, idx_version, proto)
                 WHERE p.key = v.key;`, pipValues)
	log.Info(ctx, "deduplicate pipeline versions statement", zap.String("stmt", stmt))
	if _, err := tx.ExecContext(ctx, stmt); err != nil {
		return errors.Wrapf(err, "update pipeline rows statement: %v", stmt)
	}
	return nil
}

func updateJobRows(ctx context.Context, tx *pachsql.Tx, jobUpdates []*jobRow) error {
	if len(jobUpdates) == 0 {
		return nil
	}
	var jobValues string
	for _, u := range jobUpdates {
		jobValues += fmt.Sprintf(" ('%s', decode('%v', 'hex')),", u.Key, hex.EncodeToString(u.Proto))
	}
	jobValues = jobValues[:len(jobValues)-1]
	stmt := fmt.Sprintf(`
                 UPDATE collections.jobs AS j SET
                   proto = v.proto
                 FROM (VALUES%s) AS v(key, proto)
                 WHERE j.key = v.key;`, jobValues)
	if _, err := tx.ExecContext(ctx, stmt); err != nil {
		return errors.Wrapf(err, "update job rows statement: %v", stmt)
	}
	return nil
}

type pipDBRow struct {
	Key   string `db:"key"`
	Proto []byte `db:"proto"`
}

type PipUpdateRow struct {
	Key        string
	Proto      []byte
	IdxVersion string
}

var listJobInfos = `
  SELECT key, proto
  FROM collections.jobs
`

type jobRow struct {
	Key   string `db:"key"`
	Proto []byte `db:"proto"`
}

// COPIED from src/internal/ppsdb/ppsdb.go
func VersionKey(project, pipeline string, version uint64) string {
	// zero pad in case we want to sort
	return fmt.Sprintf("%s/%s@%08d", project, pipeline, version)
}

// COPIED from src/internal/ppsdb/ppsdb.go
//
// parsePipelineKey expects keys to either be of the form <pipeline>@<id> or
// <project>/<pipeline>@<id>.
func parsePipelineKey(key string) (projectName, pipelineName, id string, err error) {
	parts := strings.Split(key, "@")
	if len(parts) != 2 || !uuid.IsUUIDWithoutDashes(parts[1]) {
		return "", "", "", errors.Errorf("key %s is not of form [<project>/]<pipeline>@<id>", key)
	}
	id = parts[1]
	parts = strings.Split(parts[0], "/")
	if len(parts) == 0 {
		return "", "", "", errors.Errorf("key %s is not of form [<project>/]<pipeline>@<id>")
	}
	pipelineName = parts[len(parts)-1]
	if len(parts) == 1 {
		return
	}
	projectName = strings.Join(parts[0:len(parts)-1], "/")
	return
}
