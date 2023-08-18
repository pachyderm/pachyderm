package pfsdb

import (
	"context"
	"database/sql"
	"fmt"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// CommitsChannelName is used to watch events for the commits table.
	CommitsChannelName = "pfs_commits"
)

// CommitID is the row id for a commit entry in postgres.
// A separate type is defined for safety so row ids must be explicitly cast for use in another table.
type CommitID uint64

// ErrCommitNotFound is returned by GetCommit() when a commit is not found in postgres.
type ErrCommitNotFound struct {
	Repo        string
	ID          CommitID
	CommitSetID string
}

// Error satisfies the error interface.
func (err ErrCommitNotFound) Error() string {
	if id := err.ID; id != 0 {
		return fmt.Sprintf("commit id=%d not found", id)
	}
	return fmt.Sprintf("commit %s/%s not found", err.Repo, err.CommitSetID)
}

func (err ErrCommitNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

func IsErrCommitNotFound(err error) bool {
	return errors.As(err, &ErrCommitNotFound{})
}

// ErrCommitAlreadyExists is returned by CreateCommit() when a commit with the same name already exists in postgres.
type ErrCommitAlreadyExists struct {
	Repo        string
	CommitSetID string
}

// Error satisfies the error interface.
func (err ErrCommitAlreadyExists) Error() string {
	if n, t := err.CommitSetID, err.Repo; n != "" && t != "" {
		return fmt.Sprintf("commit %s.%s already exists", n, t)
	}
	return "commit already exists"
}

func (err ErrCommitAlreadyExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, err.Error())
}

// CommitPair is an (id, commitInfo) tuple returned by the commit iterator.
type CommitPair struct {
	ID         CommitID
	CommitInfo *pfs.CommitInfo
}

// this dropped global variable instantiation forces the compiler to check whether CommitIterator implements stream.Iterator.
var _ stream.Iterator[CommitPair] = &CommitIterator{}

// CommitIterator batches a page of commitRow entries. Entries can be retrieved using iter.Next().
type CommitIterator struct {
	limit   int
	offset  int
	commits []commitRow
	index   int
	tx      *pachsql.Tx
	filter  CommitListFilter
}

type commitRow struct {
	ID             CommitID  `db:"int_id"`
	CommitSetID    string    `db:"commit_set_id"`
	RepoID         RepoID    `db:"repo_id"`
	Origin         string    `db:"origin"`
	Description    string    `db:"description"`
	StartTime      time.Time `db:"start_time"`
	FinishingTime  time.Time `db:"finishing_time"`
	FinishedTime   time.Time `db:"finished_time"`
	CompactingTime int64     `db:"compacting_time"`
	ValidatingTime int64     `db:"validating_time"`
	Error          string    `db:"error"`
	Size           uint64    `db:"size"`
	CreatedAt      time.Time `db:"created_at"`
	RepoName       string    `db:"repo_name"`
	RepoType       string    `db:"repo_type"`
	ProjectName    string    `db:"proj_name"`
	//UpdatedAt      time.Time `db:"updated_at"`
}

// Next advances the iterator by one row. It returns a stream.EOS when there are no more entries.
func (iter *CommitIterator) Next(ctx context.Context, dst *CommitPair) error {
	if dst == nil {
		return errors.Wrap(fmt.Errorf("commit is nil"), "get next commit")
	}
	var err error
	if iter.index >= len(iter.commits) {
		iter.index = 0
		iter.offset += iter.limit
		iter.commits, err = listCommitPage(ctx, iter.tx, iter.limit, iter.offset, iter.filter)
		if err != nil {
			return errors.Wrap(err, "list commit page")
		}
		if len(iter.commits) == 0 {
			return stream.EOS()
		}
	}
	row := iter.commits[iter.index]
	commit, err := getCommitFromCommitRow(&row)
	if err != nil {
		return errors.Wrap(err, "getting commitInfo from commit row")
	}
	*dst = CommitPair{
		CommitInfo: commit,
		ID:         row.ID,
	}
	iter.index++
	return nil
}

// CommitFields is used in the ListCommitFilter and defines specific field names for type safety.
// This should hopefully prevent a library user from misconfiguring the filter.
type CommitFields string

var (
	CommitSetIDs  = CommitFields("commit_set_id")
	CommitOrigins = CommitFields("origin")
	CommitRepos   = CommitFields("repo_id")
)

// CommitListFilter is a filter for listing commits. It ANDs together separate keys, but ORs together the key values:
// where commit.<key_1> IN (<key_1:value_1>, <key_2:value_2>, ...) AND commit.<key_2> IN (<key_2:value_1>,<key_2:value_2>,...)
type CommitListFilter map[CommitFields][]string

// ListCommit returns a CommitIterator that exposes a Next() function for retrieving *pfs.CommitInfo references.
func ListCommit(ctx context.Context, tx *pachsql.Tx, filter CommitListFilter) (*CommitIterator, error) {
	limit := 100
	page, err := listCommitPage(ctx, tx, limit, 0, filter)
	if err != nil {
		return nil, errors.Wrap(err, "list commits")
	}
	iter := &CommitIterator{
		commits: page,
		limit:   limit,
		tx:      tx,
		filter:  filter,
	}
	return iter, nil
}

func listCommitPage(ctx context.Context, tx *pachsql.Tx, limit, offset int, filter CommitListFilter) ([]commitRow, error) {
	return nil, nil
}

func sanitizeTimestamppb(timestamp *timestamppb.Timestamp) time.Time {
	if timestamp != nil {
		return timestamp.AsTime()
	}
	return time.Time{}
}

func durationpbToBigInt(duration *durationpb.Duration) int64 {
	if duration != nil {
		return duration.Seconds
	}
	return 0
}

// CreateCommit creates an entry in the pfs.commits table.
func CreateCommit(ctx context.Context, tx *pachsql.Tx, commit *pfs.CommitInfo) error {
	if commit.Commit.Repo == nil {
		return errors.New(fmt.Sprintf("commit.Repo is nil: %+v", commit.Commit))
	}
	if commit.Details == nil { // stub in an empty details struct to avoid panics.
		commit.Details = &pfs.CommitInfo_Details{}
	}
	query := `
		INSERT INTO pfs.commits 
    	(commit_set_id, repo_id, origin, start_time, description, finishing_time, finished_time, compacting_time, 
    	 validating_time, error, size) 
		VALUES ($1, (SELECT id from pfs.repos WHERE name=$2 AND type=$3 AND project_id=(SELECT id from core.projects WHERE name=$4)), 
		        $5::pfs.commit_origin, $6, $7, $8, $9, $10, $11, $12, $13);`
	_, err := tx.ExecContext(ctx, query, commit.Commit.Id, commit.Commit.Repo.Name, commit.Commit.Repo.Type,
		commit.Commit.Repo.Project.Name, strings.ToLower(commit.Origin.Kind.String()), sanitizeTimestamppb(commit.Started),
		commit.Description, sanitizeTimestamppb(commit.Finishing), sanitizeTimestamppb(commit.Finished),
		durationpbToBigInt(commit.Details.CompactingTime), durationpbToBigInt(commit.Details.ValidatingTime),
		commit.Error, commit.Details.SizeBytes)
	if err != nil && IsDuplicateKeyErr(err) { // a duplicate key implies that an entry for the repo already exists.
		return ErrCommitAlreadyExists{CommitSetID: commit.Commit.Id, Repo: RepoKey(commit.Commit.Repo)}
	}
	return errors.Wrap(err, "create commit")
}

// DeleteCommit deletes an entry in the pfs.commits table.
func DeleteCommit(ctx context.Context, tx *pachsql.Tx, commitSetID string) error {
	return nil
}

func GetCommitID(ctx context.Context, tx *pachsql.Tx, commitSetID string) (CommitID, error) {
	row, err := getCommitRowByName(ctx, tx, commitSetID)
	if err != nil {
		return CommitID(0), err
	}
	return row.ID, nil
}

func getCommitRowByName(ctx context.Context, tx *pachsql.Tx, commitSetID string) (*commitRow, error) {
	row := &commitRow{}
	return row, nil
}

func GetCommit(ctx context.Context, tx *pachsql.Tx, id CommitID) (*pfs.CommitInfo, error) {
	if id == 0 {
		return nil, errors.New("invalid id: 0")
	}
	row := &commitRow{}
	getCommit := `
	SELECT 
    	commit.int_id, commit.commit_set_id, commit.repo_id, commit.origin, commit.description, commit.start_time, 
    	commit.finishing_time, commit.finished_time, commit.compacting_time, commit.validating_time, commit.created_at, 
    	commit.error, commit.size, repo.name AS repo_name, repo.type AS repo_type, project.name AS proj_name
		FROM pfs.commits commit
		JOIN pfs.repos repo ON commit.repo_id = repo.id
		JOIN core.projects project ON repo.project_id = project.id`
	err := tx.QueryRowxContext(ctx, fmt.Sprintf("%s WHERE int_id=$1", getCommit), id).StructScan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCommitNotFound{ID: id}
		}
		return nil, errors.Wrap(err, "scanning commit row")
	}
	fmt.Printf("%+v", row)
	return getCommitFromCommitRow(row)
}

func getCommitFromCommitRow(row *commitRow) (*pfs.CommitInfo, error) {
	repo := &pfs.Repo{
		Name: row.RepoName,
		Type: row.RepoType,
		Project: &pfs.Project{
			Name: row.ProjectName,
		},
	}
	commitInfo := &pfs.CommitInfo{
		Commit: &pfs.Commit{
			Repo: repo,
			Id:   row.CommitSetID,
			Branch: &pfs.Branch{
				Repo: repo,
				Name: "", //todo(fahad): figure out how to get this.
			},
		},
	}
	return commitInfo, nil
}

// UpsertCommit will attempt to insert a commit. If the commit already exists, it will update its description.
func UpsertCommit(ctx context.Context, tx *pachsql.Tx, commit *pfs.CommitInfo) error {
	return nil
}

// UpdateCommit overwrites an existing commit entry by CommitID.
func UpdateCommit(ctx context.Context, tx *pachsql.Tx, id CommitID, commit *pfs.CommitInfo) error {
	return nil
}
