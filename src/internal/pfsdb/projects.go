package pfsdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// ProjectNotFoundError is returned by GetProject() when a project is not found in postgres.
type ProjectNotFoundError struct {
	Name string
	ID   ProjectID
}

// Error satisfies the error interface.
func (err *ProjectNotFoundError) Error() string {
	if n := err.Name; n != "" {
		return fmt.Sprintf("project %q not found", n)
	}
	if id := err.ID; id != 0 {
		return fmt.Sprintf("project id=%d not found", id)
	}
	return "project not found"
}

func (err *ProjectNotFoundError) Is(other error) bool {
	return errors.As(other, &ProjectNotFoundError{})
}

func (err *ProjectNotFoundError) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

// ProjectAlreadyExistsError is returned by CreateProject() when a project with the same name already exists in postgres.
type ProjectAlreadyExistsError struct {
	Name string
}

// Error satisfies the error interface.
func (err *ProjectAlreadyExistsError) Error() string {
	if n := err.Name; n != "" {
		return fmt.Sprintf("project %q already exists", n)
	}

	return "project already exists"
}

func (err *ProjectAlreadyExistsError) Is(other error) bool {
	_, ok := other.(*ProjectAlreadyExistsError)
	return ok
}

func (err *ProjectAlreadyExistsError) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, err.Error())
}

func IsErrProjectAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

type projectColumn string

var (
	ProjectColumnID        = projectColumn("project.id")
	ProjectColumnCreatedAt = projectColumn("project.created_at")
	ProjectColumnUpdatedAt = projectColumn("project.updated_at")
)

type OrderByProjectColumn OrderByColumn[projectColumn]

type ProjectIterator struct {
	paginator pageIterator[Project]
	extCtx    sqlx.ExtContext
}

type ProjectWithID struct {
	*pfs.ProjectInfo
	ID       ProjectID
	Revision int64
}

func (i *ProjectIterator) Next(ctx context.Context, dst *ProjectWithID) error {
	if dst == nil {
		return errors.Errorf("dst ProjectInfo cannot be nil")
	}
	project, rev, err := i.paginator.next(ctx, i.extCtx)
	if err != nil {
		return err
	}
	projectInfo := project.PbInfo()
	dst.ID = project.ID
	dst.ProjectInfo = projectInfo
	dst.Revision = rev
	return nil
}

func NewProjectIterator(ctx context.Context, extCtx sqlx.ExtContext, startPage, pageSize uint64, filter *pfs.Project, orderBys ...OrderByProjectColumn) (*ProjectIterator, error) {
	var conditions []string
	var values []any
	if filter != nil {
		if filter.Name != "" {
			conditions = append(conditions, "project.name = ?")
			values = append(values, filter.Name)
		}
	}
	query := "SELECT id,name,description,created_at FROM core.projects project"
	if len(conditions) > 0 {
		query += "\n" + fmt.Sprintf("WHERE %s", strings.Join(conditions, " AND "))
	}
	// Compute ORDER BY
	var orderByGeneric []OrderByColumn[projectColumn]
	if len(orderBys) == 0 {
		orderByGeneric = []OrderByColumn[projectColumn]{{Column: ProjectColumnID, Order: SortOrderAsc}}
	} else {
		for _, orderBy := range orderBys {
			orderByGeneric = append(orderByGeneric, OrderByColumn[projectColumn](orderBy))
		}
	}
	query += "\n" + OrderByQuery[projectColumn](orderByGeneric...)
	query = extCtx.Rebind(query)
	return &ProjectIterator{
		paginator: newPageIterator[Project](ctx, query, values, startPage, pageSize, 0),
		extCtx:    extCtx,
	}, nil
}

func ForEachProject(ctx context.Context, tx *pachsql.Tx, cb func(projectWithID ProjectWithID) error) error {
	iter, err := NewProjectIterator(ctx, tx, 0, 100, nil)
	if err != nil {
		return errors.Wrap(err, "for each project")
	}
	if err = stream.ForEach[ProjectWithID](ctx, iter, cb); err != nil {
		return errors.Wrap(err, "for each project")
	}
	return nil
}

func ListProject(ctx context.Context, tx *pachsql.Tx) ([]ProjectWithID, error) {
	var projects []ProjectWithID
	if err := ForEachProject(ctx, tx, func(projectWithID ProjectWithID) error {
		projects = append(projects, projectWithID)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list project")
	}
	return projects, nil
}

// CreateProject creates an entry in the core.projects table.
func CreateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO core.projects (name, description) VALUES ($1, $2);", project.Project.Name, project.Description)
	//todo: insert project.authInfo into auth table.
	if err != nil && IsErrProjectAlreadyExists(err) {
		return &ProjectAlreadyExistsError{Name: project.Project.Name}
	}
	return errors.Wrap(err, "create project")
}

// DeleteProject deletes an entry in the core.projects table.
func DeleteProject(ctx context.Context, tx *pachsql.Tx, projectName string) error {
	result, err := tx.ExecContext(ctx, "DELETE FROM core.projects WHERE name = $1;", projectName)
	if err != nil {
		return errors.Wrap(err, "delete project")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "could not get affected rows")
	}
	if rowsAffected == 0 {
		return &ProjectNotFoundError{
			Name: projectName,
		}
	}
	return nil
}

// GetProject is like GetProjectByName, but retrieves an entry using the row id.
func GetProject(ctx context.Context, tx *pachsql.Tx, id ProjectID) (*pfs.ProjectInfo, error) {
	proj, err := getProject(ctx, tx, "id", id)
	if err != nil {
		return nil, errors.Wrap(err, "get project by name")
	}
	return proj.ProjectInfo, nil
}

// GetProjectByName retrieves an entry from the core.projects table by project name.
func GetProjectByName(ctx context.Context, tx *pachsql.Tx, projectName string) (*pfs.ProjectInfo, error) {
	proj, err := getProject(ctx, tx, "name", projectName)
	if err != nil {
		return nil, errors.Wrap(err, "get project by name")
	}
	return proj.ProjectInfo, nil
}

// GetProjectWithID is like GetProjectByName, but retrieves an entry along with its id.
func GetProjectWithID(ctx context.Context, tx *pachsql.Tx, projectName string) (*ProjectWithID, error) {
	return getProject(ctx, tx, "name", projectName)
}

func getProject(ctx context.Context, tx *pachsql.Tx, where string, whereVal interface{}) (*ProjectWithID, error) {
	row := tx.QueryRowxContext(ctx, fmt.Sprintf("SELECT name, description, created_at, id FROM core.projects WHERE %s = $1", where), whereVal)
	project := &pfs.ProjectInfo{Project: &pfs.Project{}}
	id := 0
	var createdAt time.Time
	err := row.Scan(&project.Project.Name, &project.Description, &createdAt, &id)
	if err != nil {
		if err == sql.ErrNoRows {
			if name, ok := whereVal.(string); ok {
				return nil, &ProjectNotFoundError{Name: name}
			}
			return nil, &ProjectNotFoundError{ID: whereVal.(ProjectID)}
		}
		return nil, errors.Wrap(err, "scanning project row")
	}
	project.CreatedAt = timestamppb.New(createdAt)
	return &ProjectWithID{
		ID:          ProjectID(id),
		ProjectInfo: project,
	}, nil
}

// UpsertProject updates all fields of an existing project entry in the core.projects table by name. If 'upsert' is set to true, UpsertProject()
// will attempt to call CreateProject() if the entry does not exist.
func UpsertProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo) error {
	return updateProject(ctx, tx, project, "name", project.Project.Name, true)
}

// UpdateProject overwrites an existing project entry by ID.
func UpdateProject(ctx context.Context, tx *pachsql.Tx, id ProjectID, project *pfs.ProjectInfo) error {
	return updateProject(ctx, tx, project, "id", id, false)
}

func updateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo, where string, whereVal interface{}, upsert bool) error {
	res, err := tx.ExecContext(ctx, fmt.Sprintf("UPDATE core.projects SET name = $1, description = $2 WHERE %s = $3;", where),
		project.Project.Name, project.Description, whereVal)
	if err != nil {
		return errors.Wrap(err, "update project")
	}
	numRows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "get affected rows")
	}
	if numRows == 0 {
		if upsert {
			return CreateProject(ctx, tx, project)
		}
		return errors.New(fmt.Sprintf("%s not found in core.projects", project.Project.Name))
	}
	return nil
}
