package pfsdb

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// ErrRepoNotFound is returned by GetRepo() when a repo is not found in postgres.
type ErrRepoNotFound struct {
	Project string
	Name    string
	ID      pachsql.ID
}

// Error satisfies the error interface.
func (err ErrRepoNotFound) Error() string {
	if n := err.Name; n != "" {
		return fmt.Sprintf("repo %q not found", n)
	}
	if id := err.ID; id != 0 {
		return fmt.Sprintf("repo id=%d not found", id)
	}
	return "repo not found"
}

func (err ErrRepoNotFound) Is(other error) bool {
	_, ok := other.(ErrRepoNotFound)
	return ok
}

func (err ErrRepoNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

// ErrRepoAlreadyExists is returned by CreateRepo() when a repo with the same name already exists in postgres.
type ErrRepoAlreadyExists struct {
	Project string
	Name    string
}

// Error satisfies the error interface.
func (err ErrRepoAlreadyExists) Error() string {
	if n := err.Name; n != "" {
		return fmt.Sprintf("repo %q already exists", n)
	}

	return "repo already exists"
}

func (err ErrRepoAlreadyExists) Is(other error) bool {
	_, ok := other.(ErrRepoAlreadyExists)
	return ok
}

func (err ErrRepoAlreadyExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, err.Error())
}

func IsErrRepoAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

// RepoIterator batches a page of repoRow entries. Entries can be retrieved using iter.Next().
type RepoIterator struct {
	limit  int
	offset int
	repos  []repoRow
	index  int
	tx     *pachsql.Tx
}

type repoRow struct {
	ID          uint64    `db:"id"`
	Name        string    `db:"name"`
	ProjectID   string    `db:"project_id"`
	Description string    `db:"description"`
	RepoType    string    `db:"type"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Next advances the iterator by one row. It returns a stream.EOS when there are no more entries.
func (iter *RepoIterator) Next(ctx context.Context, dst **pfs.RepoInfo) error {
	if dst == nil {
		return errors.Wrap(fmt.Errorf("repo is nil"), "get next repo")
	}
	var err error
	if iter.index >= len(iter.repos) {
		iter.index = 0
		iter.offset += iter.limit
		iter.repos, err = listRepo(ctx, iter.tx, iter.limit, iter.offset)
		if err != nil {
			return errors.Wrap(err, "list repo page")
		}
		if len(iter.repos) == 0 {
			return stream.EOS()
		}
	}
	row := iter.repos[iter.index]
	proj, err := getProjectFromRepoRow(ctx, iter.tx, &row)
	if err != nil {
		return errors.Wrap(err, "list repo")
	}
	*dst = &pfs.RepoInfo{
		Repo: &pfs.Repo{
			Name:    row.Name,
			Type:    row.RepoType,
			Project: proj.Project,
		},
		Description: row.Description,
	}
	iter.index++
	return nil
}

// ListRepo returns a RepoIterator that exposes a Next() function for retrieving *pfs.RepoInfo references.
func ListRepo(ctx context.Context, tx *pachsql.Tx) (*RepoIterator, error) {
	limit := 100
	page, err := listRepo(ctx, tx, limit, 0)
	if err != nil {
		return nil, errors.Wrap(err, "list repos")
	}
	iter := &RepoIterator{
		repos: page,
		limit: limit,
		tx:    tx,
	}
	return iter, nil
}

func listRepo(ctx context.Context, tx *pachsql.Tx, limit, offset int) ([]repoRow, error) {
	var page []repoRow
	if err := tx.SelectContext(ctx, &page, "SELECT name, type, project_id, description FROM pfs.repos ORDER BY id ASC LIMIT $1 OFFSET $2", limit, offset); err != nil {
		return nil, errors.Wrap(err, "could not get repo page")

	}
	return page, nil
}

// CreateRepo creates an entry in the pfs.repos table.
func CreateRepo(ctx context.Context, tx *pachsql.Tx, repo *pfs.RepoInfo) error {
	if repo.Repo.Type == "" {
		repo.Repo.Type = "unknown"
	}
	if repo.Repo.Project == nil {
		repo.Repo.Project = &pfs.Project{Name: "default"}
	}
	_, err := tx.ExecContext(ctx, "INSERT INTO pfs.repos (name, type, project_id, description) VALUES ($1, $2::pfs.repo_type, (SELECT id from core.projects where name=$3), $4);",
		repo.Repo.Name, repo.Repo.Type, repo.Repo.Project.Name, repo.Description)
	//todo: figure out how to handle branches?
	if err != nil && IsErrRepoAlreadyExists(err) {
		return ErrRepoAlreadyExists{Project: repo.Repo.Project.Name, Name: repo.Repo.Name}
	}
	return errors.Wrap(err, "create repo")
}

// DeleteRepo deletes an entry in the pfs.repos table.
func DeleteRepo(ctx context.Context, tx *pachsql.Tx, repoName string) error {
	result, err := tx.ExecContext(ctx, "DELETE FROM pfs.repos WHERE name = $1;", repoName)
	if err != nil {
		return errors.Wrap(err, "delete repo")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "could not get affected rows")
	}
	if rowsAffected == 0 {
		return ErrRepoNotFound{
			Name: repoName,
		}
	}
	return nil
}

func DeleteAllRepos(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, "TRUNCATE pfs.repos;")
	return errors.Wrap(err, "could not delete all repo rows")
}

// todo(fahad): GetByIndex() Repo Type

// GetRepo is like GetRepoByName, but retrieves an entry using the row id.
func GetRepo(ctx context.Context, tx *pachsql.Tx, id pachsql.ID) (*pfs.RepoInfo, error) {
	return getRepo(ctx, tx, "id", id)
}

// GetRepoByName retrieves an entry from the pfs.repos table by repo name.
func GetRepoByName(ctx context.Context, tx *pachsql.Tx, repoName string) (*pfs.RepoInfo, error) {
	return getRepo(ctx, tx, "name", repoName)
}

func getRepo(ctx context.Context, tx *pachsql.Tx, where string, whereVal interface{}) (*pfs.RepoInfo, error) {
	row := &repoRow{}
	err := tx.QueryRowxContext(ctx, fmt.Sprintf("SELECT name, type, project_id, description FROM pfs.repos WHERE %s = $1",
		where), whereVal).StructScan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			if name, ok := whereVal.(string); ok {
				return nil, ErrRepoNotFound{Name: name}
			}
			return nil, ErrRepoNotFound{ID: whereVal.(pachsql.ID)}
		}
		return nil, errors.Wrap(err, "get repo: scanning repo row")
	}
	proj, err := getProjectFromRepoRow(ctx, tx, row)
	repoInfo := &pfs.RepoInfo{
		Repo: &pfs.Repo{
			Name:    row.Name,
			Type:    row.RepoType,
			Project: proj.Project,
		},
		Description: row.Description,
	}
	return repoInfo, nil
}

func getProjectFromRepoRow(ctx context.Context, tx *pachsql.Tx, row *repoRow) (*pfs.ProjectInfo, error) {
	id, err := strconv.ParseUint(row.ProjectID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "get project from repo row: parsing project ID")
	}
	projInfo, err := coredb.GetProject(ctx, tx, pachsql.ID(id))
	if err != nil {
		return nil, errors.Wrap(err, "get project from repo row: get project")
	}
	return projInfo, nil
}

// UpsertRepo updates all fields of an existing repo entry in the pfs.repos table by name. If 'upsert' is set to true, UpsertRepo()
// will attempt to call CreateRepo() if the entry does not exist.
func UpsertRepo(ctx context.Context, tx *pachsql.Tx, repo *pfs.RepoInfo) error {
	return updateRepo(ctx, tx, repo, "name", repo.Repo.Name, true)
}

// UpdateRepo overwrites an existing repo entry by pachsql.ID.
func UpdateRepo(ctx context.Context, tx *pachsql.Tx, id pachsql.ID, repo *pfs.RepoInfo) error {
	return updateRepo(ctx, tx, repo, "id", id, false)
}

func updateRepo(ctx context.Context, tx *pachsql.Tx, repo *pfs.RepoInfo, where string, whereVal interface{}, upsert bool) error {
	if repo.Repo.Type == "" {
		repo.Repo.Type = "unknown"
	}
	if repo.Repo.Project == nil {
		repo.Repo.Project = &pfs.Project{Name: "default"}
	}
	res, err := tx.ExecContext(ctx,
		fmt.Sprintf("UPDATE pfs.repos SET name = $1, type = $2::pfs.repo_type, project_id = (SELECT id FROM core.projects WHERE name = $3), description = $4 WHERE %s = $5;", where),
		repo.Repo.Name, repo.Repo.Type, repo.Repo.Project.Name, repo.Description, whereVal)
	if err != nil {
		return errors.Wrap(err, "update repo")
	}
	numRows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get affected rows")
	}
	if numRows == 0 {
		if upsert {
			return CreateRepo(ctx, tx, repo)
		}
		return errors.New(fmt.Sprintf("%s not found in pfs.repos", repo.Repo.Name))
	}
	return nil
}
