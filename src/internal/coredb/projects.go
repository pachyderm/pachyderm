package coredb

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/types"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"time"
)

type ProjectIterator struct {
}

// QueryExecer defines an interface for functions shared across sqlx.Tx and sqlx.DB types.
// Functions that take a querier support both running inside a transaction and outside it.
type QueryExecer interface {
	sqlx.QueryerContext
	sqlx.ExecerContext
}

// CreateProject creates an entry in the core.projects table.
func CreateProject(ctx context.Context, queryExecer QueryExecer, project *pfs.ProjectInfo) error {
	_, err := queryExecer.ExecContext(ctx, "INSERT INTO core.projects (name, description) VALUES ($1, $2);", project.Project.Name, project.Description)
	//todo: insert project.authInfo into auth table.
	return errors.Wrap(err, "failed to create project")
}

// DeleteProject deletes an entry in the core.projects table.
func DeleteProject(ctx context.Context, queryExecer QueryExecer, projectName string) error {
	_, err := queryExecer.ExecContext(ctx, "DELETE FROM core.projects WHERE name = $1;", projectName)
	//todo: delete corresponding project.authInfo auth table.
	return errors.Wrap(err, "failed to delete project")
}

// GetProject retrieves an entry from the core.projects table by project name.
func GetProject(ctx context.Context, queryExecer QueryExecer, projectName string) (*pfs.ProjectInfo, error) {
	return getProject(ctx, queryExecer, "name", projectName)
}

// GetProjectByID is like GetProject, but retrieves an entry using the row id.
func GetProjectByID(ctx context.Context, queryExecer QueryExecer, id int64) (*pfs.ProjectInfo, error) {
	return getProject(ctx, queryExecer, "id", id)
}

func getProject(ctx context.Context, queryExecer QueryExecer, where string, whereVal interface{}) (*pfs.ProjectInfo, error) {
	row := queryExecer.QueryRowxContext(ctx, fmt.Sprintf("SELECT name, description, created_at FROM core.projects WHERE %s = $1", where), whereVal)
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
func UpdateProject(ctx context.Context, queryExecer QueryExecer, project *pfs.ProjectInfo, upsert bool) error {
	return updateProject(ctx, queryExecer, project, "name", project.Project.Name, upsert)
}

// UpdateProjectByID is like UpdateProject, but uses the row id instead of the name. It does not allow upserting.
func UpdateProjectByID(ctx context.Context, queryExecer QueryExecer, id int64, project *pfs.ProjectInfo) error {
	return updateProject(ctx, queryExecer, project, "id", id, false)
}

func updateProject(ctx context.Context, queryExecer QueryExecer, project *pfs.ProjectInfo, where string, whereVal interface{}, upsert bool) error {
	res, err := queryExecer.ExecContext(ctx, fmt.Sprintf("UPDATE core.projects SET name = $1, description = $2 WHERE %s = $3;", where),
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
			return CreateProject(ctx, queryExecer, project)
		}
		return errors.New(fmt.Sprintf("%s not found in core.projects", project.Project.Name))
	}
	return nil
}
