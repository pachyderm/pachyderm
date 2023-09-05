package pfsdb

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func GetBranch(ctx context.Context, tx *pachsql.Tx, id BranchID) (*Branch, error) {
	branch := new(Branch)
	if err := tx.GetContext(ctx, branch, `
		SELECT
			branch.name,
			branch.created_at,
			branch.updated_at,
			repo.name as "repo.name",
			repo.type as "repo.type",
			project.name as "repo.project.name",
			commit.commit_set_id as "head.commit_set_id",
			repo.name as "head.repo.name",
			repo.type as "head.repo.type",
			project.name as "head.repo.project.name"
		FROM pfs.branches branch
			JOIN pfs.repos repo ON branch.repo_id = repo.id
			JOIN core.projects project ON repo.project_id = project.id
			JOIN pfs.commits commit ON branch.head = commit.int_id
		WHERE branch.id = $1
	`, id); err != nil {
		return nil, errors.Wrap(err, "could not get branch")
	}
	return branch, nil
}

func UpsertBranch(ctx context.Context, tx *pachsql.Tx, branch *Branch) (BranchID, error) {
	var query string
	if branch.Repo.Name != "" && branch.Repo.Type != "" && branch.Repo.Project.Name != "" && branch.Head.CommitSetID != "" {
		query = `
			INSERT INTO pfs.branches(repo_id, name, head)
			VALUES (
				(SELECT repo.id FROM pfs.repos repo JOIN core.projects project ON repo.project_id = project.id WHERE project.name = $1 AND repo.name = $2 AND repo.type = $3),
				$4,
				(SELECT int_id FROM pfs.commits WHERE commit_id = $5)
			)
			ON CONFLICT (repo_id, name) DO UPDATE SET head = EXCLUDED.head
			RETURNING id
		`
	} else {
		return 0, errors.Errorf("branch must have a repo name and head commit")
	}

	var branchID BranchID
	head := &pfs.Commit{Repo: &pfs.Repo{Name: branch.Repo.Name, Type: branch.Repo.Type, Project: &pfs.Project{Name: branch.Repo.Project.Name}}, Id: branch.Head.CommitSetID}
	if err := tx.QueryRowContext(ctx, query, branch.Repo.Project.Name, branch.Repo.Name, branch.Repo.Type, branch.Name, CommitKey(head)).Scan(&branchID); err != nil {
		return 0, errors.Wrap(err, "could not create branch")
	}
	return branchID, nil
}
