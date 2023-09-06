package pfsdb

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/jackc/pgconn"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	// ReposChannelName is used to watch events for the repos table.
	ReposChannelName = "pfs_repos"

	getRepoAndBranches = `
	SELECT
		repo.id,
		repo.name,
		repo.type,
		repo.description,
		repo.project_id as "project.id",
		project.name AS "project.name",
		array_agg(branch.proto) AS branches,
		repo.created_at,
		repo.updated_at
	FROM
		pfs.repos repo
			JOIN core.projects project ON repo.project_id = project.id
			LEFT JOIN collections.branches branch ON project.name || '/' || repo.name || '.' || repo.type = branch.idx_repo
	`
	noBranches = "{NULL}"
)

// ErrRepoNotFound is returned by GetRepo() when a repo is not found in postgres.
type ErrRepoNotFound struct {
	Project string
	Name    string
	Type    string
	ID      RepoID
}

// Error satisfies the error interface.
func (err ErrRepoNotFound) Error() string {
	return fmt.Sprintf("repo (id=%d, project=%s, name=%s, type=%s) not found", err.ID, err.Project, err.Name, err.Type)
}

func (err ErrRepoNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

func IsErrRepoNotFound(err error) bool {
	return errors.As(err, &ErrRepoNotFound{})
}

func IsDuplicateKeyErr(err error) bool {
	targetErr := &pgconn.PgError{}
	ok := errors.As(err, targetErr)
	if !ok {
		return false
	}
	return targetErr.Code == "23505" // duplicate key SQLSTATE
}

// RepoPair is an (id, repoInfo) tuple returned by the repo iterator.
type RepoPair struct {
	ID       RepoID
	RepoInfo *pfs.RepoInfo
}

// this dropped global variable instantiation forces the compiler to check whether RepoIterator implements stream.Iterator.
var _ stream.Iterator[RepoPair] = &RepoIterator{}

// RepoIterator batches a page of repoRow entries. Entries can be retrieved using iter.Next().
type RepoIterator struct {
	limit  int
	offset int
	repos  []Repo
	index  int
	tx     *pachsql.Tx
	filter RepoListFilter
}

// Next advances the iterator by one row. It returns a stream.EOS when there are no more entries.
func (iter *RepoIterator) Next(ctx context.Context, dst *RepoPair) error {
	if dst == nil {
		return errors.New("repo is nil, get next repo")
	}
	var err error
	if iter.index >= len(iter.repos) {
		iter.index = 0
		iter.offset += iter.limit
		iter.repos, err = listRepoPage(ctx, iter.tx, iter.limit, iter.offset, iter.filter)
		if err != nil {
			return errors.Wrap(err, "list repo page")
		}
		if len(iter.repos) == 0 {
			return stream.EOS()
		}
	}
	row := iter.repos[iter.index]
	repo, err := getRepoFromRepoRow(&row)
	if err != nil {
		return errors.Wrap(err, "getting repoInfo from repo row")
	}
	*dst = RepoPair{
		RepoInfo: repo,
		ID:       row.ID,
	}
	iter.index++
	return nil
}

// RepoFields is used in the ListRepoFilter and defines specific field names for type safety.
// This should hopefully prevent a library user from misconfiguring the filter.
type RepoFields string

var (
	RepoTypes    = RepoFields("type")
	RepoProjects = RepoFields("project_id")
	RepoNames    = RepoFields("name")
)

// RepoListFilter is a filter for listing repos. It ANDs together separate keys, but ORs together the key values:
// where repo.<key_1> IN (<key_1:value_1>, <key_2:value_2>, ...) AND repo.<key_2> IN (<key_2:value_1>,<key_2:value_2>,...)
type RepoListFilter map[RepoFields][]string

// ListRepo returns a RepoIterator that exposes a Next() function for retrieving *pfs.RepoInfo references.
func ListRepo(ctx context.Context, tx *pachsql.Tx, filter RepoListFilter) (*RepoIterator, error) {
	limit := 100
	page, err := listRepoPage(ctx, tx, limit, 0, filter)
	if err != nil {
		return nil, errors.Wrap(err, "list repos")
	}
	iter := &RepoIterator{
		repos:  page,
		limit:  limit,
		tx:     tx,
		filter: filter,
	}
	return iter, nil
}

func listRepoPage(ctx context.Context, tx *pachsql.Tx, limit, offset int, filter RepoListFilter) ([]Repo, error) {
	var page []Repo
	where := ""
	conditions := make([]string, 0)
	for key, vals := range filter {
		if len(vals) == 0 {
			continue
		}
		quotedVals := make([]string, 0)
		for _, val := range vals {
			quotedVals = append(quotedVals, fmt.Sprintf("'%s'", val))
		}
		if key == RepoProjects {
			conditions = append(conditions, fmt.Sprintf("repo.%s IN (SELECT id FROM core.projects WHERE name IN (%s))", string(key), strings.Join(quotedVals, ",")))
		} else {
			conditions = append(conditions, fmt.Sprintf("repo.%s IN (%s)", string(key), strings.Join(quotedVals, ",")))
		}
	}
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}
	if err := tx.SelectContext(ctx, &page, fmt.Sprintf("%s %s GROUP BY repo.id, project.name ORDER BY repo.id ASC LIMIT $1 OFFSET $2;",
		getRepoAndBranches, where), limit, offset); err != nil {
		return nil, errors.Wrap(err, "could not get repo page")
	}
	return page, nil
}

// DeleteRepo deletes an entry in the pfs.repos table.
func DeleteRepo(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) error {
	result, err := tx.ExecContext(ctx, "DELETE FROM pfs.repos "+
		"WHERE project_id=(SELECT id FROM core.projects WHERE name=$1) AND name=$2 AND type=$3;", repoProject, repoName, repoType)
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

func GetRepoID(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) (RepoID, error) {
	row, err := getRepoRowByName(ctx, tx, repoProject, repoName, repoType)
	if err != nil {
		return RepoID(0), err
	}
	return row.ID, nil
}

func getRepoRowByName(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) (*Repo, error) {
	repo := &Repo{}
	if repoProject == "" {
		repoProject = pfs.DefaultProjectName
	}
	err := tx.QueryRowxContext(ctx, fmt.Sprintf("%s WHERE repo.project_id=(SELECT id from core.projects where name=$1) "+
		"AND repo.name=$2 AND repo.type=$3 GROUP BY repo.id, project.name;", getRepoAndBranches), repoProject, repoName, repoType).StructScan(repo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRepoNotFound{Project: repoProject, Name: repoName, Type: repoType}
		}
		return nil, errors.Wrap(err, "scanning repo row")
	}
	return repo, nil
}

// todo(fahad): rewrite branch related code during the branches migration.
// GetRepo retrieves an entry from the pfs.repos table by using the row id.
func GetRepo(ctx context.Context, tx *pachsql.Tx, id RepoID) (*pfs.RepoInfo, error) {
	if id == 0 {
		return nil, errors.New("invalid id: 0")
	}
	row := &Repo{}
	err := tx.QueryRowxContext(ctx, fmt.Sprintf("%s WHERE repo.id=$1 GROUP BY repo.id, project.name;", getRepoAndBranches), id).StructScan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRepoNotFound{ID: id}
		}
		return nil, errors.Wrap(err, "scanning repo row")
	}
	return getRepoFromRepoRow(row)
}

// GetRepoByName retrieves an entry from the pfs.repos table by project, repo name, and type.
func GetRepoByName(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) (*pfs.RepoInfo, error) {
	row, err := getRepoRowByName(ctx, tx, repoProject, repoName, repoType)
	if err != nil {
		return nil, err
	}
	return getRepoFromRepoRow(row)
}

func getRepoFromRepoRow(row *Repo) (*pfs.RepoInfo, error) {
	proj := &pfs.Project{
		Name: row.Project.Name,
	}
	branches, err := getBranchesFromRepoRow(row)
	if err != nil {
		return nil, errors.Wrap(err, "getting branches from repo row")
	}
	repoInfo := &pfs.RepoInfo{
		Repo: &pfs.Repo{
			Name:    row.Name,
			Type:    row.Type,
			Project: proj,
		},
		Description: row.Description,
		Branches:    branches,
		Created:     timestamppb.New(row.CreatedAt),
	}
	return repoInfo, nil
}

func getBranchesFromRepoRow(row *Repo) ([]*pfs.Branch, error) {
	if row == nil {
		return nil, errors.New("repo row is nil")
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

// UpsertRepo will attempt to insert a repo, and return its ID. If the repo already exists, it will update its description.
func UpsertRepo(ctx context.Context, tx *pachsql.Tx, repo *pfs.RepoInfo) (RepoID, error) {
	if repo.Repo.Name == "" {
		return 0, errors.Errorf("repo name is required: %+v", repo.Repo)
	}
	if repo.Repo.Type == "" {
		return 0, errors.Errorf("repo type is required: %+v", repo.Repo)
	}
	if repo.Repo.Project == nil {
		return 0, errors.Errorf("project is required: %+v", repo.Repo)
	}
	var repoID RepoID
	if err := tx.QueryRowContext(ctx,
		`
		INSERT INTO pfs.repos (name, type, project_id, description)
		VALUES ($1, $2, (SELECT id from core.projects where name=$3), $4)
		ON CONFLICT (name, type, project_id) DO UPDATE SET description= EXCLUDED.description
		RETURNING id
		`,
		repo.Repo.Name, repo.Repo.Type, repo.Repo.Project.Name, repo.Description,
	).Scan(&repoID); err != nil {
		return 0, errors.Wrap(err, "upsert repo")
	}
	return repoID, nil
}
