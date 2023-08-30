package pfsdb

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

type BranchID uint64
type BranchField string
type ListBranchFilter map[BranchField][]string

type BranchInfoWithID struct {
	*pfs.BranchInfo
	ID BranchID
}
type BranchIterator struct {
}

// var _ stream.Iterator[BranchInfoWithID] = &BranchIterator{}

func GetBranch(ctx context.Context, tx *pachsql.Tx, id BranchID) (*pfs.BranchInfo, error) {
	var branch Branch
	if err := tx.GetContext(ctx, &branch, `
		SELECT
			branch.id as id,
			branch.name,
			branch.head,
			branch.created_at,
			branch.updated_at,
			repo.id as "repo.id",
			repo.name as "repo.name",
			project.name as "repo.project.name"
		FROM pfs.branches branch
			JOIN pfs.repos repo ON branch.repo_id = repo.id
			JOIN core.projects project ON repo.project_id = project.id
		WHERE branch.id = $1
	`, id); err != nil {
		return nil, errors.Wrap(err, "could not get branch")
	}
	return &pfs.BranchInfo{
		Branch: &pfs.Branch{Name: branch.Name,
			Repo: &pfs.Repo{Name: branch.Repo.Name,
				Project: &pfs.Project{Name: branch.Repo.Project.Name}}}}, nil
}

func GetBranchID(ctx context.Context, tx *pachsql.Tx, project, repo, branch string) (BranchID, error) {
	return 0, nil
}

func ListBranch(ctx context.Context, tx *pachsql.Tx, filter ListBranchFilter) (*BranchIterator, error) {
	return nil, nil
}

func CreateBranch(ctx context.Context, tx *pachsql.Tx, branch *pfs.BranchInfo) (BranchID, error) {
	var id BranchID
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO pfs.branches(repo_id, name, head)
		VALUES (
			(SELECT repo.id FROM pfs.repos repo JOIN core.projects project ON repo.project_id = project.id WHERE project.name = $1 AND repo.name = $2),
			$3,
			(SELECT int_id FROM pfs.commits WHERE commit_id = $4)
		)
		RETURNING id
	`, branch.Branch.Repo.Project.Name, branch.Branch.Repo.Name, branch.Branch.Name, CommitKey(branch.Head)).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "could not create branch")
	}
	return id, nil
}

func DeleteBranch(ctx context.Context, tx *pachsql.Tx, id BranchID) error {
	return nil
}

func DeleteBranchByName(ctx context.Context, tx *pachsql.Tx, project, repo, branch string) error {
	return nil
}
