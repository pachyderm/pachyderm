package coredb

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"time"
)

// CreateProject creates an entry in the core.projects table.
func CreateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO core.projects (name, description) VALUES ($1, $2);", project.Project.Name, project.Description)
	//todo: insert project.authInfo into auth table.
	return errors.EnsureStack(err)
}

// DeleteProject deletes an entry in the core.projects table.
func DeleteProject(ctx context.Context, tx *pachsql.Tx, projectName string) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM core.projects WHERE name = $1;", projectName)
	//todo: delete corresponding project.authInfo auth table.
	return errors.EnsureStack(err)
}

// GetProject retrieves an entry from the core.projects table by project name.
func GetProject(ctx context.Context, tx *pachsql.Tx, projectName string) (*pfs.ProjectInfo, error) {
	return getProject(ctx, tx, "name", projectName)
}

// GetProjectByID is like GetProject, but retrieves an entry using the row id.
func GetProjectByID(ctx context.Context, tx *pachsql.Tx, id int64) (*pfs.ProjectInfo, error) {
	return getProject(ctx, tx, "id", id)
}

func getProject(ctx context.Context, tx *pachsql.Tx, where string, whereVal interface{}) (*pfs.ProjectInfo, error) {
	row := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT name, description, created_at FROM core.projects WHERE %s = $1", where), whereVal)
	project := &pfs.ProjectInfo{Project: &pfs.Project{}}
	var createdAt time.Time
	var err error
	if err = row.Scan(&project.Project.Name, &project.Description, &createdAt); err != nil {
		return nil, errors.EnsureStack(err)
	}
	project.CreatedAt, err = types.TimestampProto(createdAt)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return project, nil
}

// UpdateProject updates all fields of an existing project entry in the core.projects table by name. If 'upsert' is set to true, UpdateProject()
// will attempt to call CreateProject() if the entry does not exist.
func UpdateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo, upsert bool) error {
	return updateProject(ctx, tx, project, "name", project.Project.Name, upsert)
}

// UpdateProjectByID is like UpdateProject, but uses the row id instead of the name. It does not allow upserting.
func UpdateProjectByID(ctx context.Context, tx *pachsql.Tx, id int64, project *pfs.ProjectInfo) error {
	return updateProject(ctx, tx, project, "id", id, false)
}

func updateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo, where string, whereVal interface{}, upsert bool) error {
	res, err := tx.ExecContext(ctx, fmt.Sprintf("UPDATE core.projects SET name = $1, description = $2 WHERE %s = $3;", where),
		project.Project.Name, project.Description, whereVal)
	if err != nil {
		return errors.EnsureStack(err)
	}
	numRows, err := res.RowsAffected()
	if err != nil {
		return errors.EnsureStack(err)
	}
	if numRows == 0 {
		if upsert {
			return CreateProject(ctx, tx, project)
		}
		return errors.New(fmt.Sprintf("%s not found in core.projects", project.Project.Name))
	}
	return errors.EnsureStack(err)
}
