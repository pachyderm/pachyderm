package pfsdb

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	getRepoAndBranches = "SELECT repo.id, repo.name, repo.type, repo.project_id, " +
		"repo.description, array_agg(branch.proto) AS branches, repo.created_at, repo.updated_at FROM pfs.repos repo " +
		"LEFT JOIN core.projects project ON repo.project_id = project.id " +
		"LEFT JOIN collections.branches branch ON project.name || '/' || repo.name || '.' || repo.type = branch.idx_repo "
	noBranches = "{NULL}"
)

// ErrRepoNotFound is returned by GetRepo() when a repo is not found in postgres.
type ErrRepoNotFound struct {
	Project string
	Name    string
	Type    string
	ID      pachsql.ID
}

// Error satisfies the error interface.
func (err ErrRepoNotFound) Error() string {
	if id := err.ID; id != 0 {
		return fmt.Sprintf("repo id=%d not found", id)
	}
	return fmt.Sprintf("repo %s/%s.%s not found", err.Project, err.Name, err.Type)
}

func (err ErrRepoNotFound) Is(other error) bool {
	_, ok := other.(ErrRepoNotFound)
	return ok
}

func (err ErrRepoNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

func IsErrRepoNotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}

// ErrRepoAlreadyExists is returned by CreateRepo() when a repo with the same name already exists in postgres.
type ErrRepoAlreadyExists struct {
	Project string
	Name    string
	Type    string
}

// Error satisfies the error interface.
func (err ErrRepoAlreadyExists) Error() string {
	if n, t := err.Name, err.Type; n != "" && t != "" {
		return fmt.Sprintf("repo %s.%s already exists", n, t)
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
	limit    int
	offset   int
	repos    []repoRow
	index    int
	tx       *pachsql.Tx
	where    string
	whereVal interface{}
}

type repoRow struct {
	ID                 uint64    `db:"id"`
	Name               string    `db:"name"`
	ProjectID          string    `db:"project_id"`
	ProjectName        string    `db:"proj_name"`
	ProjectDescription string    `db:"proj_desc"`
	Description        string    `db:"description"`
	RepoType           string    `db:"type"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
	// Branches is a string that contains an array of hex-encoded branchInfos. The array is enclosed with curly braces.
	// Each entry is prefixed with '//x' and entries are delimited by a ','
	Branches string `db:"branches"`
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
		iter.repos, err = listRepoPage(ctx, iter.tx, iter.limit, iter.offset, iter.where, iter.whereVal)
		if err != nil {
			return errors.Wrap(err, "list repo page")
		}
		if len(iter.repos) == 0 {
			return stream.EOS()
		}
	}
	row := iter.repos[iter.index]
	*dst, err = getRepoFromRepoRow(ctx, iter.tx, &row)
	if err != nil {
		return errors.Wrap(err, "getting repoInfo from repo row")
	}
	iter.index++
	return nil
}

// ListRepo returns a RepoIterator that exposes a Next() function for retrieving *pfs.RepoInfo references.
func ListRepo(ctx context.Context, tx *pachsql.Tx) (*RepoIterator, error) {
	return listRepo(ctx, tx, "", nil)
}

// ListRepoByIdxType is like ListRepo but only iterates over repo.type = repoType.
func ListRepoByIdxType(ctx context.Context, tx *pachsql.Tx, repoType string) (*RepoIterator, error) {
	return listRepo(ctx, tx, "type", repoType)
}

// ListRepoByIdxName is like ListRepo but only iterates over repo.name = name.
func ListRepoByIdxName(ctx context.Context, tx *pachsql.Tx, repoName string) (*RepoIterator, error) {
	return listRepo(ctx, tx, "name", repoName)
}

// ListRepo returns a RepoIterator that exposes a Next() function for retrieving *pfs.RepoInfo references.
func listRepo(ctx context.Context, tx *pachsql.Tx, where string, whereVal interface{}) (*RepoIterator, error) {
	limit := 100
	page, err := listRepoPage(ctx, tx, limit, 0, where, whereVal)
	if err != nil {
		return nil, errors.Wrap(err, "list repos")
	}
	iter := &RepoIterator{
		repos: page,
		limit: limit,
		tx:    tx,
	}
	if where != "" && whereVal != nil {
		iter.where = where
		iter.whereVal = whereVal
	}
	return iter, nil
}

func listRepoPage(ctx context.Context, tx *pachsql.Tx, limit, offset int, where string, whereVal interface{}) ([]repoRow, error) {
	var page []repoRow
	if where != "" && whereVal != nil {
		if err := tx.SelectContext(ctx, &page,
			fmt.Sprintf("%s WHERE repo.%s = $1 GROUP BY repo.id ORDER BY repo.id ASC LIMIT $2 OFFSET $3;", getRepoAndBranches, where),
			whereVal, limit, offset); err != nil {
			return nil, errors.Wrap(err, "could not get repo page")
		}
		return page, nil
	}
	err := tx.SelectContext(ctx, &page, getRepoAndBranches+"GROUP BY repo.id ORDER BY repo.id ASC LIMIT $1 OFFSET $2 ;", limit, offset)
	if err != nil {
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
	if err != nil && IsErrRepoAlreadyExists(err) {
		return ErrRepoAlreadyExists{Project: repo.Repo.Project.Name, Name: repo.Repo.Name}
	}
	return errors.Wrap(err, "create repo")
}

// DeleteRepo deletes an entry in the pfs.repos table.
func DeleteRepo(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) error {
	result, err := tx.ExecContext(ctx, "DELETE FROM pfs.repos WHERE project_id=(SELECT id FROM core.projects WHERE name=$1) AND name=$2 AND type=$3;", repoProject, repoName, repoType)
	if err != nil {
		return errors.Wrap(err, "delete repo")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "could not get affected rows")
	}
	if rowsAffected == 0 {
		return ErrRepoNotFound{Project: repoProject, Name: repoName, Type: repoType}
	}
	return nil
}

func GetRepoID(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) (pachsql.ID, error) {
	row, err := getRepoRowByName(ctx, tx, repoProject, repoName, repoType)
	if err != nil {
		return pachsql.ID(0), err
	}
	return pachsql.ID(row.ID), nil
}

func getRepoRowByName(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) (*repoRow, error) {
	row := &repoRow{}
	if repoProject == "" {
		repoProject = "default"
	}
	err := tx.QueryRowxContext(ctx, fmt.Sprintf("%s WHERE repo.project_id=(SELECT id from core.projects where name=$1) AND repo.name=$2 AND repo.type=$3 GROUP BY repo.id;", getRepoAndBranches), repoProject, repoName, repoType).StructScan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRepoNotFound{Project: repoProject, Name: repoName, Type: repoType}
		}
		return nil, errors.Wrap(err, "scanning repo row")
	}
	return row, nil
}

// todo(fahad): rewrite branch related code during the branches migration.
// todo(fahad): do we need to worry about a repo with more than the default postgres LIMIT number of branches?
// GetRepo retrieves an entry from the pfs.repos table by using the row id.
func GetRepo(ctx context.Context, tx *pachsql.Tx, id pachsql.ID) (*pfs.RepoInfo, error) {
	row := &repoRow{}
	err := tx.QueryRowxContext(ctx, fmt.Sprintf("%s WHERE repo.id=$1 GROUP BY repo.id;", getRepoAndBranches), id).StructScan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRepoNotFound{ID: id}
		}
		return nil, errors.Wrap(err, "scanning repo row")
	}
	return getRepoFromRepoRow(ctx, tx, row)
}

// GetRepoByName retrieves an entry from the pfs.repos table by project, repo name, and type.
func GetRepoByName(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) (*pfs.RepoInfo, error) {
	row, err := getRepoRowByName(ctx, tx, repoProject, repoName, repoType)
	if err != nil {
		return nil, err
	}
	return getRepoFromRepoRow(ctx, tx, row)
}

func getRepoFromRepoRow(ctx context.Context, tx *pachsql.Tx, row *repoRow) (*pfs.RepoInfo, error) {
	proj, err := getProjectFromRepoRow(ctx, tx, row)
	if err != nil {
		return nil, errors.Wrap(err, "getting project from repo row")
	}
	branches, err := getBranchesFromRepoRow(row)
	if err != nil {
		return nil, errors.Wrap(err, "getting branches from repo row")
	}
	repoInfo := &pfs.RepoInfo{
		Repo: &pfs.Repo{
			Name:    row.Name,
			Type:    row.RepoType,
			Project: proj.Project,
		},
		Description: row.Description,
		Branches:    branches,
		Created:     timestamppb.New(row.CreatedAt),
	}
	return repoInfo, nil
}

// todo(fahad): should this be a join too?
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

func getBranchesFromRepoRow(row *repoRow) ([]*pfs.Branch, error) {
	if row == nil {
		return nil, errors.Wrap(fmt.Errorf("repo row is nil"), "")
	}
	branches := make([]*pfs.Branch, 0)
	if row.Branches == noBranches {
		return branches, nil
	}
	// after aggregation, braces, quotes, and leading hex prefixes need to be removed from the encoded branch strings.
	for _, branchStr := range strings.Split(strings.Trim(row.Branches, "{}"), ",") {
		branchHex := strings.Trim(strings.Trim(branchStr, "\""), "\\x")
		decodedString, err := hex.DecodeString(branchHex)
		if err != nil {
			return nil, errors.Wrap(err, "branch not hex encoded")
		}
		branchInfo := &pfs.BranchInfo{}
		if err := proto.Unmarshal(decodedString, branchInfo); err != nil {
			return nil, errors.Wrap(err, "error unmarshalling")
		}
		branches = append(branches, branchInfo.Branch)
	}
	return branches, nil
}

// UpsertRepo updates all fields of an existing repo entry in the pfs.repos table by name. If 'upsert' is set to true, UpsertRepo()
// will attempt to call CreateRepo() if the entry does not exist.
func UpsertRepo(ctx context.Context, tx *pachsql.Tx, repo *pfs.RepoInfo) error {
	id, err := GetRepoID(ctx, tx, repo.Repo.Project.Name, repo.Repo.Name, repo.Repo.Type)
	if err != nil {
		if IsErrRepoNotFound(err) {
			return errors.Wrap(CreateRepo(ctx, tx, repo), "create repo if not exists")
		}
		return errors.Wrap(err, "get repo id")
	}
	return errors.Wrap(UpdateRepo(ctx, tx, id, repo), "update existing repo")
}

// UpdateRepo overwrites an existing repo entry by pachsql.ID.
func UpdateRepo(ctx context.Context, tx *pachsql.Tx, id pachsql.ID, repo *pfs.RepoInfo) error {
	if repo.Repo.Type == "" {
		repo.Repo.Type = "unknown"
	}
	if repo.Repo.Project == nil {
		repo.Repo.Project = &pfs.Project{Name: "default"}
	}
	res, err := tx.ExecContext(ctx, "UPDATE pfs.repos SET name=$1, type=$2::pfs.repo_type, project_id=(SELECT id FROM core.projects WHERE name=$3), description=$4 WHERE id=$5;",
		repo.Repo.Name, repo.Repo.Type, repo.Repo.Project.Name, repo.Description, id)
	if err != nil {
		return errors.Wrap(err, "update repo")
	}
	numRows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get affected rows")
	}
	if numRows == 0 {
		return ErrRepoNotFound{Project: repo.Repo.Project.Name, Name: repo.Repo.Name, Type: repo.Repo.Type}
	}
	return nil
}
