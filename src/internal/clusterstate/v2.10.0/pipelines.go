package v2_10_0

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var createUniqueIndex = "CREATE UNIQUE INDEX pip_version_idx ON collections.pipelines(idx_name,idx_version)"

// find pipelines with duplicate versions
// each returned row has a pipeline with duplicate versions
// the newest createdat timestamp is returned
// all keys and protos are returned in arrays ordered by createdat
// so newest duplicate is last entry in array
// duplicated idx_version is returned whole and split into pipeline name and version number

// select count(1) count, max(createdat) max_createdat, array_agg(key order by createdat) keys, array_agg(proto order by createdat) protos, idx_version, split_part(idx_version, '@', 1) pipeline_path, cast(split_part(idx_version, '@', 2) as int8) version_number from collections.pipelines group by idx_name, idx_version having count(1) > 1;

var duplicatePipelinesQuery = `
select
  count(1) count,
  max(createdat) max_createdat,
  array_agg(key order by createdat) keys,
  array_agg(proto order by createdat) protos,
  idx_version,
  split_part(idx_version, '@', 1) pipeline_path,
  cast(split_part(idx_version, '@', 2) as int8) version_number
from collections.pipelines
group by idx_name, idx_version
having count(1) > 1
`

type pipDBRow struct {
	Count         int64     `db:"count"`
	MaxCreatedAt  time.Time `db:"max_createdat"`
	Keys          []string  `db:"keys"`
	Protos        [][]byte  `db:"protos"`
	IdxVersion    string    `db:"idx_version"`
	PipelinePath  string    `db:"pipeline_path"`
	VersionNumber int64     `db:"version_number"`
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

func DeduplicatePipelineVersions(ctx context.Context, env migrations.Env) error {
	pipUpdates, pipVersionChanges, err := collectPipelineUpdates(ctx, env.Tx)
	if err != nil {
		return err
	}
	log.Info(ctx, "detected updates for pipeline versions", zap.Int("updates_count", len(pipUpdates)))
	if err := UpdatePipelineRows(ctx, env.Tx, pipUpdates); err != nil {
		return err
	}
	jobUpdates, err := collectJobUpdates(ctx, env.Tx, pipVersionChanges)
	if err != nil {
		return err
	}
	if err := UpdateJobRows(ctx, env.Tx, jobUpdates); err != nil {
		return err
	}
	if _, err := env.Tx.ExecContext(ctx, createUniqueIndex); err != nil {
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
		if err := rr.Err(); err != nil {
			return nil, nil, errors.Wrap(err, "row error")
		}
		if err := rr.StructScan(&row); err != nil {
			return nil, nil, errors.Wrap(err, "scan pipeline row")
		}
		key := row.Keys[len(row.Keys)-1]
		rowProto := row.Protos[len(row.Protos)-1]
		if err := proto.Unmarshal(rowProto, pi); err != nil {
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
			project, pipeline, _, err := parsePipelineKey(key)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "parse key %q", key)
			}
			idxVersion := VersionKey(project, pipeline, correctVersion)
			updates = append(updates, &PipUpdateRow{Key: key, IdxVersion: idxVersion, Proto: data})
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

func UpdatePipelineRows(ctx context.Context, tx *pachsql.Tx, pipUpdates []*PipUpdateRow) error {
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
		log.Info(ctx, "deduplicate pipeline versions statement", zap.String("stmt", stmt))
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return errors.Wrapf(err, "update pipeline rows statement: %v", stmt)
		}
	}
	return nil
}

func UpdateJobRows(ctx context.Context, tx *pachsql.Tx, jobUpdates []*jobRow) error {
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
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return errors.Wrapf(err, "update job rows statement: %v", stmt)
		}
	}
	return nil
}
