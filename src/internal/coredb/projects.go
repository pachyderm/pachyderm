package coredb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// ErrProjectNotFound is returned by GetProject() when a project is not found in postgres.
type ErrProjectNotFound struct {
	Name string
	ID   pachsql.ID
}

// Error satisfies the error interface.
func (err ErrProjectNotFound) Error() string {
	if n := err.Name; n != "" {
		return fmt.Sprintf("project %q not found", n)
	}
	if id := err.ID; id != 0 {
		return fmt.Sprintf("project id=%d not found", id)
	}
	return "project not found"
}

func (err ErrProjectNotFound) Is(other error) bool {
	_, ok := other.(ErrProjectNotFound)
	return ok
}

func (err ErrProjectNotFound) GRPCStatus() *status.Status {
	return status.New(codes.NotFound, err.Error())
}

// ErrProjectAlreadyExists is returned by CreateProject() when a project with the same name already exists in postgres.
type ErrProjectAlreadyExists struct {
	Name string
}

// Error satisfies the error interface.
func (err ErrProjectAlreadyExists) Error() string {
	if n := err.Name; n != "" {
		return fmt.Sprintf("project %q already exists", n)
	}

	return "project already exists"
}

func (err ErrProjectAlreadyExists) Is(other error) bool {
	_, ok := other.(ErrProjectAlreadyExists)
	return ok
}

func (err ErrProjectAlreadyExists) GRPCStatus() *status.Status {
	return status.New(codes.AlreadyExists, err.Error())
}

func IsErrProjectAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

// ProjectIterator batches a page of projectRow entries. Entries can be retrieved using iter.Next().
type ProjectIterator struct {
	limit    int
	offset   int
	projects []projectRow
	index    int
	tx       *pachsql.Tx
}

type projectRow struct {
	ID          pachsql.ID `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// Next advances the iterator by one row. It returns a stream.EOS when there are no more entries.
func (iter *ProjectIterator) Next(ctx context.Context, dst **pfs.ProjectInfo) error {
	if dst == nil {
		return errors.Wrap(fmt.Errorf("project is nil"), "get next project")
	}
	var err error
	if iter.index >= len(iter.projects) {
		iter.index = 0
		iter.offset += iter.limit
		iter.projects, err = listProject(ctx, iter.tx, iter.limit, iter.offset)
		if err != nil {
			return errors.Wrap(err, "list project page")
		}
		if len(iter.projects) == 0 {
			return stream.EOS()
		}
	}
	row := iter.projects[iter.index]
	*dst = &pfs.ProjectInfo{
		Project:     &pfs.Project{Name: row.Name},
		Description: row.Description,
		CreatedAt:   timestamppb.New(row.CreatedAt),
	}
	iter.index++
	return nil
}

// ListProject returns a ProjectIterator that exposes a Next() function for retrieving *pfs.ProjectInfo references.
func ListProject(ctx context.Context, tx *pachsql.Tx) (*ProjectIterator, error) {
	limit := 100
	page, err := listProject(ctx, tx, limit, 0)
	if err != nil {
		return nil, errors.Wrap(err, "list projects")
	}
	iter := &ProjectIterator{
		projects: page,
		limit:    limit,
		tx:       tx,
	}
	return iter, nil
}

func listProject(ctx context.Context, tx *pachsql.Tx, limit, offset int) ([]projectRow, error) {
	var page []projectRow
	if err := tx.SelectContext(ctx, &page, "SELECT name,description,created_at FROM core.projects ORDER BY id ASC LIMIT $1 OFFSET $2", limit, offset); err != nil {
		return nil, errors.Wrap(err, "could not get project page")

	}
	return page, nil
}

// CreateProject creates an entry in the core.projects table.
func CreateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO core.projects (name, description) VALUES ($1, $2);", project.Project.Name, project.Description)
	//todo: insert project.authInfo into auth table.
	if err != nil && IsErrProjectAlreadyExists(err) {
		return ErrProjectAlreadyExists{Name: project.Project.Name}
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
		return ErrProjectNotFound{
			Name: projectName,
		}
	}
	return nil
}

// GetProject is like GetProjectByName, but retrieves an entry using the row id.
func GetProject(ctx context.Context, tx *pachsql.Tx, id pachsql.ID) (*pfs.ProjectInfo, error) {
	return getProject(ctx, tx, "id", id)
}

// GetProjectByName retrieves an entry from the core.projects table by project name.
func GetProjectByName(ctx context.Context, tx *pachsql.Tx, projectName string) (*pfs.ProjectInfo, error) {
	return getProject(ctx, tx, "name", projectName)
}

func getProject(ctx context.Context, tx *pachsql.Tx, where string, whereVal interface{}) (*pfs.ProjectInfo, error) {
	row := tx.QueryRowxContext(ctx, fmt.Sprintf("SELECT name, description, created_at FROM core.projects WHERE %s = $1", where), whereVal)
	project := &pfs.ProjectInfo{Project: &pfs.Project{}}
	var createdAt time.Time
	err := row.Scan(&project.Project.Name, &project.Description, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			if name, ok := whereVal.(string); ok {
				return nil, ErrProjectNotFound{Name: name}
			}
			return nil, ErrProjectNotFound{ID: whereVal.(pachsql.ID)}
		}
		return nil, errors.Wrap(err, "scanning project row")
	}
	project.CreatedAt = timestamppb.New(createdAt)
	return project, nil
}

// UpsertProject updates all fields of an existing project entry in the core.projects table by name. If 'upsert' is set to true, UpsertProject()
// will attempt to call CreateProject() if the entry does not exist.
func UpsertProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo) error {
	return updateProject(ctx, tx, project, "name", project.Project.Name, true)
}

// UpdateProject overwrites an existing project entry by ID.
func UpdateProject(ctx context.Context, tx *pachsql.Tx, id pachsql.ID, project *pfs.ProjectInfo) error {
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
