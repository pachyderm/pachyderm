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

const (
	insertProjectQuery     = "INSERT INTO core.projects (name, description, created_at, updated_at) VALUES ($1,$2,$3,$4);"
	updateProjectBaseQuery = "UPDATE core.projects SET name = $1, description = $2, created_at = $3, updated_at = $4"
)

// CreateProject creates an entry in the core.projects table.
func CreateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo) error {
	//todo: this can be swapped out with postgres' created_at trigger later.
	createdAt, err := types.TimestampFromProto(project.CreatedAt)
	if err != nil {
		return errors.EnsureStack(err)
	}
	_, err = tx.ExecContext(ctx, insertProjectQuery, project.Project.Name, project.Description, createdAt, nil)
	//todo: insert project.authInfo into auth table.
	return errors.EnsureStack(err)
}

// GetProject retrieves an entry from the core.projects table by project name.
func GetProject(ctx context.Context, tx *pachsql.Tx, projectName string) (*pfs.ProjectInfo, error) {
	return getProject(ctx, tx, "where name = $1", projectName)
}

// GetProjectByID is like GetProject, but retrieves an entry using the row id.
func GetProjectByID(ctx context.Context, tx *pachsql.Tx, id int64) (*pfs.ProjectInfo, error) {
	return getProject(ctx, tx, "where id = $1", id)
}

func getProject(ctx context.Context, tx *pachsql.Tx, where string, whereVal interface{}) (*pfs.ProjectInfo, error) {
	row := tx.QueryRowContext(ctx, fmt.Sprintf("SELECT name, description, created_at FROM core.projects %s", where), whereVal)
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
// will attempt to call CreateProject() if the entry does not exist. If updateTime is not provided, the current time will be used.
func UpdateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo, upsert bool, updateTime ...*types.Timestamp) error {
	return updateProject(ctx, tx, project, "where name = $5", project.Project.Name, upsert, updateTime...)
}

// UpdateProjectByID is like UpdateProject, but uses the row id instead of the name. It does not allow upserting.
func UpdateProjectByID(ctx context.Context, tx *pachsql.Tx, id int64, project *pfs.ProjectInfo, updateTime ...*types.Timestamp) error {
	return updateProject(ctx, tx, project, "where id = $5", id, false, updateTime...)
}

func updateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo, where string, whereVal interface{}, upsert bool, updateTime ...*types.Timestamp) error {
	createdAt, err := types.TimestampFromProto(project.CreatedAt)
	if err != nil {
		return errors.EnsureStack(err)
	}
	updatedAt := time.Now()
	if updateTime != nil {
		updatedAt, err = types.TimestampFromProto(updateTime[0])
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
	res, err := tx.ExecContext(ctx, fmt.Sprintf("%s %s;", updateProjectBaseQuery, where),
		project.Project.Name, project.Description, createdAt, updatedAt, whereVal)
	if err != nil {
		return errors.EnsureStack(err)
	}
	numRows, err := res.RowsAffected()
	if err != nil {
		return errors.EnsureStack(err)
	}
	if numRows == 0 {
		if upsert {
			_, err := tx.ExecContext(ctx, "INSERT INTO core.projects (name, description, created_at, updated_at) VALUES ($1,$2,$3,$4);",
				project.Project.Name, project.Description, createdAt, updatedAt)
			return errors.EnsureStack(err)
		}
		return errors.New(fmt.Sprintf("%s not found in core.projects", project.Project.Name))
	}
	return errors.EnsureStack(err)
}
