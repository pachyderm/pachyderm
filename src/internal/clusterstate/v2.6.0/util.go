package v2_6_0

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

func repoKey(repo *pfs.Repo) string {
	return repo.Project.Name + "/" + repo.Name + "." + repo.Type
}

func oldCommitKey(commit *pfs.Commit) string {
	return branchKey(commit.Branch) + "=" + commit.Id
}

func commitBranchlessKey(commit *pfs.Commit) string {
	return repoKey(commit.Branch.Repo) + "@" + commit.Id
}

func branchKey(branch *pfs.Branch) string {
	return repoKey(branch.Repo) + "@" + branch.Name
}

// jobKey is a string representation of a Job suitable for use as an indexing
// key.  It will include the project if the project name is not the empty
// string.
func jobKey(j *pps.Job) string {
	return fmt.Sprintf("%s@%s", j.Pipeline, j.Id)
}
func pipelineCommitKey(commit *pfs.Commit) string {
	return fmt.Sprintf("%s/%s@%s", commit.Repo.Project.Name, commit.Repo.Name, commit.Id)
}

func forEachCollectionProtos[T proto.Message](ctx context.Context, tx *pachsql.Tx, table string, val T, f func(T)) error {
	rr, err := tx.QueryContext(ctx, fmt.Sprintf("SELECT proto FROM collections.%s;", table))
	if err != nil {
		return errors.Wrap(err, "could not read table")
	}
	defer rr.Close()
	for rr.Next() {
		var pb []byte
		if err := rr.Err(); err != nil {
			return errors.Wrap(err, "row error")
		}
		if err := rr.Scan(&pb); err != nil {
			return errors.Wrap(err, "could not scan row")
		}
		if err := proto.Unmarshal(pb, val); err != nil {
			return errors.Wrapf(err, "could not unmarshal proto")
		}
		f(val)
	}
	return nil
}

func listCollectionProtos[T proto.Message](ctx context.Context, tx *pachsql.Tx, table string, val T) ([]T, error) {
	log.Info(ctx, "listing collection protos", zap.String("table", table))
	protos := make([]T, 0)
	if err := forEachCollectionProtos(ctx, tx, table, val, func(T) {
		protos = append(protos, proto.Clone(val).(T))
		proto.Reset(val)
	}); err != nil {
		return nil, err
	}
	return protos, nil
}

func getCollectionProto[T proto.Message](ctx context.Context, tx *pachsql.Tx, table string, key string, v T) error {
	var pb []byte
	stmt := fmt.Sprintf("SELECT proto FROM collections.%s WHERE key=$1", table)
	if err := tx.GetContext(ctx, &pb, stmt, key); err != nil {
		return errors.Wrapf(err, "get collections.%s with key %q", table, key)
	}
	if err := proto.Unmarshal(pb, v); err != nil {
		return errors.Wrapf(err, "could not unmarshal proto")
	}
	return nil
}

func updateCollectionProto[T proto.Message](ctx context.Context, tx *pachsql.Tx, table string, oldKey string, newKey string, pb T) error {
	data, err := proto.Marshal(pb)
	if err != nil {
		return errors.Wrapf(err, "unmarshal info proto from %q table for key %q", table, oldKey)
	}
	proto.Reset(pb)
	if err := proto.Unmarshal(data, pb); err != nil {
		return errors.Wrap(err, "FAILED REMARSHALL")
	}
	stmt := fmt.Sprintf("UPDATE collections.%s SET key=$1, proto=$2 WHERE key=$3;", table)
	_, err = tx.ExecContext(ctx, stmt, newKey, data, oldKey)
	return errors.Wrapf(err, "update collections.%s with key %q", table, oldKey)
}
