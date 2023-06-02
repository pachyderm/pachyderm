package coredb

import (
	"context"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func CreateProject(ctx context.Context, tx *pachsql.Tx, project *pfs.ProjectInfo) error {
	createdAt, err := types.TimestampFromProto(project.CreatedAt)
	if err != nil {
		return errors.EnsureStack(err)
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO core.projects (name, description, created_at) VALUES ($1,$2,$3);",
		project.Project.Name, project.Description, createdAt)
	//todo: insert project.authInfo into auth table.
	return errors.EnsureStack(err)
}

func GetProject(ctx context.Context, tx *pachsql.Tx, projectName string) (*pfs.ProjectInfo, error) {
	row := tx.QueryRowContext(ctx, "SELECT (name, description, created_at) FROM core.projects WHERE name = $1;", projectName)
	project := &pfs.ProjectInfo{Project: &pfs.Project{}}
	if err := row.Scan(&project.Project.Name, &project.Description, &project.CreatedAt); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return project, nil
}
