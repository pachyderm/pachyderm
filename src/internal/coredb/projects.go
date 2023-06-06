package coredb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"time"
)

// Querier defines an interface for functions shared across sqlx.Tx and sqlx.DB types.
// Functions that take a querier support both running inside a transaction and outside it.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// CreateProject creates an entry in the core.projects table.
func CreateProject(ctx context.Context, querier Querier, project *pfs.ProjectInfo) error {
	_, err := querier.ExecContext(ctx, "INSERT INTO core.projects (name, description) VALUES ($1, $2);", project.Project.Name, project.Description)
	//todo: insert project.authInfo into auth table.
	return errors.Wrap(err, "failed to create project")
}

// DeleteProject deletes an entry in the core.projects table.
func DeleteProject(ctx context.Context, querier Querier, projectName string) error {
	_, err := querier.ExecContext(ctx, "DELETE FROM core.projects WHERE name = $1;", projectName)
	//todo: delete corresponding project.authInfo auth table.
	return errors.Wrap(err, "failed to delete project")
}

// GetProject retrieves an entry from the core.projects table by project name.
func GetProject(ctx context.Context, querier Querier, projectName string) (*pfs.ProjectInfo, error) {
	return getProject(ctx, querier, "name", projectName)
}

// GetProjectByID is like GetProject, but retrieves an entry using the row id.
func GetProjectByID(ctx context.Context, querier Querier, id int64) (*pfs.ProjectInfo, error) {
	return getProject(ctx, querier, "id", id)
}

func getProject(ctx context.Context, querier Querier, where string, whereVal interface{}) (*pfs.ProjectInfo, error) {
	row := querier.QueryRowContext(ctx, fmt.Sprintf("SELECT name, description, created_at FROM core.projects WHERE %s = $1", where), whereVal)
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

// UpdateProject updates all fields of an existing project entry in the core.projects table by name. If 'upsert' is set to true, UpdateProject()
// will attempt to call CreateProject() if the entry does not exist.
func UpdateProject(ctx context.Context, querier Querier, project *pfs.ProjectInfo, upsert bool) error {
	return updateProject(ctx, querier, project, "name", project.Project.Name, upsert)
}

// UpdateProjectByID is like UpdateProject, but uses the row id instead of the name. It does not allow upserting.
func UpdateProjectByID(ctx context.Context, querier Querier, id int64, project *pfs.ProjectInfo) error {
	return updateProject(ctx, querier, project, "id", id, false)
}

func updateProject(ctx context.Context, querier Querier, project *pfs.ProjectInfo, where string, whereVal interface{}, upsert bool) error {
	res, err := querier.ExecContext(ctx, fmt.Sprintf("UPDATE core.projects SET name = $1, description = $2 WHERE %s = $3;", where),
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
			return CreateProject(ctx, querier, project)
		}
		return errors.New(fmt.Sprintf("%s not found in core.projects", project.Project.Name))
	}
	return nil
}
