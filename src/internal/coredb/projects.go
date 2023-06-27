package coredb

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// ProjectIterator batches a page of projectRow entries. Entries can be retrieved using iter.Next().
type ProjectIterator struct {
	limit    int
	offset   int
	projects []projectRow
	index    int
	tx       *pachsql.Tx
}

// ID is the auto-incrementing primary key used for entries in postgres tables.
type ID uint64

type projectRow struct {
	ID          ID        `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Next advances the iterator by one row. It returns a stream.EOS when there are no more entries.
func (iter *ProjectIterator) Next(ctx context.Context, dst **pfs.ProjectInfo) error {
	if dst == nil {
		return errors.Wrap(fmt.Errorf("project is nil"), "failed to get next project")
	}
	var err error
	if iter.index >= len(iter.projects) {
		iter.index = 0
		iter.offset += iter.limit
		iter.projects, err = listProject(ctx, iter.tx, iter.limit, iter.offset)
		if err != nil {
			return errors.Wrap(err, "failed to list project page")
		}
		if len(iter.projects) == 0 {
			return stream.EOS()
		}
	}
	row := iter.projects[iter.index]
	projectTimestamp, err := types.TimestampProto(row.CreatedAt)
	if err != nil {
		return errors.Wrap(err, "failed to convert time.Time to proto timestamp")
	}
	*dst = &pfs.ProjectInfo{
		Project:     &pfs.Project{Name: row.Name},
		Description: row.Description,
		CreatedAt:   projectTimestamp}
	iter.index++
	return nil
}

// ListProject returns a ProjectIterator that exposes a Next() function for retrieving *pfs.ProjectInfo references.
func ListProject(ctx context.Context, tx *pachsql.Tx) (*ProjectIterator, error) {
	limit := 100
	page, err := listProject(ctx, tx, limit, 0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list projects")
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
	return errors.Wrap(err, "failed to create project")
}

// DeleteProject deletes an entry in the core.projects table.
func DeleteProject(ctx context.Context, tx *pachsql.Tx, projectName string) error {
	result, err := tx.ExecContext(ctx, "DELETE FROM core.projects WHERE name = $1;", projectName)
	if err != nil {
		return errors.Wrap(err, "failed to delete project")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "could not get affected rows")
	}
	if rowsAffected == 0 {
		return errors.Wrap(fmt.Errorf("project %s does not exist", projectName), "failed to delete project")
	}
	return nil
}

func DeleteAllProjects(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, "TRUNCATE core.projects;")
	return errors.Wrap(err, "could not delete all project rows")
}

// GetProject is like GetProjectByName, but retrieves an entry using the row id.
func GetProject(ctx context.Context, tx *pachsql.Tx, id ID) (*pfs.ProjectInfo, error) {
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
	var err error
	if err = row.Scan(&project.Project.Name, &project.Description, &createdAt); err != nil {
		return nil, errors.Wrap(err, "failed scanning project row")
	}
	project.CreatedAt, err = types.TimestampProto(createdAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed converting project proto timestamp")
	}
	return project, nil
}

// UpsertProject updates all fields of an existing project entry in the core.projects table by name. If 'upsert' is set to true, UpsertProject()
// will attempt to call CreateProject() if the entry does not exist.
func UpsertProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo) error {
	return updateProject(ctx, tx, project, "name", project.Project.Name, true)
}

// UpdateProject is like UpsertProject, but uses the row id instead of the name. It does not allow upserting.
func UpdateProject(ctx context.Context, tx *pachsql.Tx, id ID, project *pfs.ProjectInfo) error {
	return updateProject(ctx, tx, project, "id", id, false)
}

func updateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo, where string, whereVal interface{}, upsert bool) error {
	res, err := tx.ExecContext(ctx, fmt.Sprintf("UPDATE core.projects SET name = $1, description = $2 WHERE %s = $3;", where),
		project.Project.Name, project.Description, whereVal)
	if err != nil {
		return errors.Wrap(err, "failed to update project")
	}
	numRows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get affected rows")
	}
	if numRows == 0 {
		if upsert {
			return CreateProject(ctx, tx, project)
		}
		return errors.New(fmt.Sprintf("%s not found in core.projects", project.Project.Name))
	}
	return nil
}
