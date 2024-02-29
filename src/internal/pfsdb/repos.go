package pfsdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch/postgres"
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
			repo.project_id AS "project.id",
			project.name AS "project.name",
			array_agg(branch.name) AS "branches",
			repo.created_at,
			repo.updated_at
		FROM pfs.repos repo 
			JOIN core.projects project ON repo.project_id = project.id
			LEFT JOIN pfs.branches branch ON branch.repo_id = repo.id
	`
	noBranches    = "{NULL}"
	reposPageSize = 100
)

// RepoNotFoundError is returned by GetRepo() when a repo is not found in postgres.
type RepoNotFoundError struct {
	Project string
	Name    string
	Type    string
	ID      RepoID
}

// Error satisfies the error interface.
func (err *RepoNotFoundError) Error() string {
	return fmt.Sprintf("repo (id=%d, project=%s, name=%s, type=%s) not found", err.ID, err.Project, err.Name, err.Type)
}

func (err *RepoNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

func IsErrRepoNotFound(err error) bool {
	return errors.As(err, &RepoNotFoundError{})
}

func IsDuplicateKeyErr(err error) bool {
	targetErr := &pgconn.PgError{}
	ok := errors.As(err, targetErr)
	if !ok {
		return false
	}
	return targetErr.Code == "23505" // duplicate key SQLSTATE
}

// RepoInfoWithID is an (id, repoInfo) tuple returned by the repo iterator.
type RepoInfoWithID struct {
	ID RepoID
	*pfs.RepoInfo
	Revision int64
}

// this dropped global variable instantiation forces the compiler to check whether RepoIterator implements stream.Iterator.
var _ stream.Iterator[RepoInfoWithID] = &RepoIterator{}

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
		if _, err := GetProjectByName(ctx, tx, repoProject); err != nil {
			if errors.As(err, new(*ProjectNotFoundError)) {
				return errors.Join(err, &RepoNotFoundError{Project: repoProject, Name: repoName, Type: repoType})
			}
			return errors.Wrapf(err, "could not get project %v for delete repo", repoProject)
		}
		return &RepoNotFoundError{Project: repoProject, Name: repoName, Type: repoType}
	}
	return nil
}

func GetRepoID(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) (RepoID, error) {
	row, err := getRepoByName(ctx, tx, repoProject, repoName, repoType)
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

func GetRepoInfoWithID(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) (*RepoInfoWithID, error) {
	repo, err := getRepoByName(ctx, tx, repoProject, repoName, repoType)
	if err != nil {
		return nil, errors.Wrap(err, "get repo info with id")
	}
	repoInfo, err := repo.PbInfo()
	if err != nil {
		return nil, errors.Wrap(err, "get repo info with id")
	}
	return &RepoInfoWithID{
		ID:       repo.ID,
		RepoInfo: repoInfo,
	}, nil
}

// GetRepo retrieves an entry from the pfs.repos table by using the row id.
func GetRepo(ctx context.Context, tx *pachsql.Tx, id RepoID) (*pfs.RepoInfo, error) {
	if id == 0 {
		return nil, errors.New("invalid id: 0")
	}
	repo := &Repo{}
	err := tx.GetContext(ctx, repo, fmt.Sprintf("%s WHERE repo.id=$1 GROUP BY repo.id, project.name, project.id;", getRepoAndBranches), id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &RepoNotFoundError{ID: id}
		}
		return nil, errors.Wrap(err, "scanning repo row")
	}
	return repo.PbInfo()
}

// GetRepoByName retrieves an entry from the pfs.repos table by project, repo name, and type.
func GetRepoByName(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) (*pfs.RepoInfo, error) {
	repo, err := getRepoByName(ctx, tx, repoProject, repoName, repoType)
	if err != nil {
		return nil, err
	}
	return repo.PbInfo()
}

func getRepoByName(ctx context.Context, tx *pachsql.Tx, repoProject, repoName, repoType string) (*Repo, error) {
	if repoProject == "" {
		repoProject = pfs.DefaultProjectName
	}
	repo := &Repo{}
	if err := tx.GetContext(ctx, repo,
		fmt.Sprintf("%s WHERE repo.project_id=(SELECT id from core.projects where name=$1) "+
			"AND repo.name=$2 AND repo.type=$3 GROUP BY repo.id, project.name, project.id;", getRepoAndBranches),
		repoProject, repoName, repoType,
	); err != nil {
		if err == sql.ErrNoRows {
			if _, err := GetProjectByName(ctx, tx, repoProject); err != nil {
				if errors.As(err, new(*ProjectNotFoundError)) {
					return nil, errors.Join(err, &RepoNotFoundError{Project: repoProject, Name: repoName, Type: repoType})
				}
				return nil, errors.Wrapf(err, "could not get project for get repo", repoProject)
			}
			return nil, &RepoNotFoundError{Project: repoProject, Name: repoName, Type: repoType}
		}
		return nil, errors.Wrap(err, "scanning repo row")
	}
	return repo, nil
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

type repoColumn string

const (
	RepoColumnID          = repoColumn("repo.id")
	RepoColumnCreatedAt   = repoColumn("repo.created_at")
	RepoColumnUpdatedAt   = repoColumn("repo.updated_at")
	RepoColumnName        = repoColumn("repo.name")
	RepoColumnProjectName = repoColumn("project.name")
)

type OrderByRepoColumn OrderByColumn[repoColumn]

func NewRepoIterator(ctx context.Context, ext sqlx.ExtContext, startPage, pageSize, maxPages uint64, filter *pfs.Repo, orderBys ...OrderByRepoColumn) (*RepoIterator, error) {
	var conditions []string
	var values []any
	if filter != nil {
		if filter.Project != nil && filter.Project.Name != "" {
			conditions = append(conditions, "project.name = ?")
			values = append(values, filter.Project.Name)
		}
		if filter.Name != "" {
			conditions = append(conditions, "repo.name = ?")
			values = append(values, filter.Name)
		}
		if filter.Type != "" {
			conditions = append(conditions, "repo.type = ?")
			values = append(values, filter.Type)
		}
	}
	query := getRepoAndBranches
	if len(conditions) > 0 {
		query += fmt.Sprintf("\nWHERE %s", strings.Join(conditions, " AND "))
	}
	query += "\nGROUP BY repo.id, project.name, project.id"
	var orderByGeneric []OrderByColumn[repoColumn]
	if len(orderBys) == 0 {
		orderByGeneric = []OrderByColumn[repoColumn]{{Column: RepoColumnID, Order: SortOrderAsc}}
	} else {
		for _, orderBy := range orderBys {
			orderByGeneric = append(orderByGeneric, OrderByColumn[repoColumn](orderBy))
		}
	}
	query += "\n" + OrderByQuery[repoColumn](orderByGeneric...)
	query = ext.Rebind(query)
	return &RepoIterator{
		paginator: newPageIterator[Repo](ctx, query, values, startPage, pageSize, maxPages),
		ext:       ext,
	}, nil
}

func ForEachRepo(ctx context.Context, tx *pachsql.Tx, filter *pfs.Repo, page *pfs.RepoPage, cb func(repoWithID RepoInfoWithID) error, orderBys ...OrderByRepoColumn) error {
	var maxPages uint64
	if page == nil {
		page = &pfs.RepoPage{
			PageIndex: 0,
			PageSize:  100,
		}
	} else {
		maxPages = 1
	}
	if len(orderBys) == 0 {
		var err error
		orderBys, err = makePageOrderBys(page.Order)
		if err != nil {
			return err
		}
	}
	iter, err := NewRepoIterator(ctx, tx, uint64(page.PageIndex), uint64(page.PageSize), maxPages, filter, orderBys...)
	if err != nil {
		return errors.Wrap(err, "for each repo")
	}
	if err := stream.ForEach[RepoInfoWithID](ctx, iter, cb); err != nil {
		return errors.Wrap(err, "for each repo")
	}
	return nil
}

func makePageOrderBys(ordering pfs.RepoPage_Ordering) ([]OrderByRepoColumn, error) {
	if ordering == pfs.RepoPage_PROJECT_REPO {
		return []OrderByRepoColumn{
			{Column: RepoColumnProjectName, Order: SortOrderAsc},
			{Column: RepoColumnName, Order: SortOrderAsc},
		}, nil
	}
	return nil, errors.Errorf("cannot make page order bys for ordering %v", ordering)
}

func ListRepo(ctx context.Context, tx *pachsql.Tx, filter *pfs.Repo, page *pfs.RepoPage, orderBys ...OrderByRepoColumn) ([]RepoInfoWithID, error) {
	var repos []RepoInfoWithID
	if err := ForEachRepo(ctx, tx, filter, page, func(repoWithID RepoInfoWithID) error {
		repos = append(repos, repoWithID)
		return nil
	}, orderBys...); err != nil {
		return nil, errors.Wrap(err, "list repo")
	}
	return repos, nil
}

type RepoIterator struct {
	paginator pageIterator[Repo]
	ext       sqlx.ExtContext
}

func (i *RepoIterator) Next(ctx context.Context, dst *RepoInfoWithID) error {
	if dst == nil {
		return errors.Errorf("dst RepoInfo cannot be nil")
	}
	repo, rev, err := i.paginator.next(ctx, i.ext)
	if err != nil {
		return err
	}
	repoInfo, err := repo.PbInfo()
	if err != nil {
		return err
	}
	dst.ID = repo.ID
	dst.RepoInfo = repoInfo
	dst.Revision = rev
	return nil
}

// Helper functions for watching repos.
type repoUpsertHandler func(id RepoID, repoInfo *pfs.RepoInfo) error
type repoDeleteHandler func(id RepoID) error

func WatchRepos(ctx context.Context, db *pachsql.DB, listener collection.PostgresListener, onUpsert repoUpsertHandler, onDelete repoDeleteHandler) error {
	watcher, err := postgres.NewWatcher(db, listener, randutil.UniqueString("watch-repos-"), ReposChannelName)
	if err != nil {
		return err
	}
	defer watcher.Close()
	snapshot, err := NewRepoIterator(ctx, db, 0, reposPageSize, 0, nil, OrderByRepoColumn{Column: RepoColumnID, Order: SortOrderAsc})
	if err != nil {
		return err
	}
	return watchRepos(ctx, db, snapshot, watcher.Watch(), onUpsert, onDelete)
}

func watchRepos(ctx context.Context, db *pachsql.DB, snapshot stream.Iterator[RepoInfoWithID], events <-chan *postgres.Event, onUpsert repoUpsertHandler, onDelete repoDeleteHandler) error {
	// Handle snapshot.
	if err := stream.ForEach[RepoInfoWithID](ctx, snapshot, func(r RepoInfoWithID) error {
		return onUpsert(r.ID, r.RepoInfo)
	}); err != nil {
		return err
	}
	// Handle events.
	for {
		select {
		case event, ok := <-events:
			if !ok {
				return errors.New("watch repos: events channel closed")
			}
			if event.Err != nil {
				return event.Err
			}
			id := RepoID(event.Id)
			switch event.Type {
			case postgres.EventDelete:
				if err := onDelete(id); err != nil {
					return err
				}
			case postgres.EventInsert, postgres.EventUpdate:
				var repoInfo *pfs.RepoInfo
				if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
					var err error
					repoInfo, err = GetRepo(ctx, tx, id)
					if err != nil {
						return err
					}
					return nil
				}); err != nil {
					return err
				}
				if err := onUpsert(id, repoInfo); err != nil {
					return err
				}
			default:
				return errors.Errorf("unknown event type: %v", event.Type)
			}
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "watch repos")
		}
	}
}
