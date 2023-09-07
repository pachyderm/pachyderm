package pfsdb

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// GetBranch returns a branch by id.
func GetBranch(ctx context.Context, tx *pachsql.Tx, id BranchID) (*pfs.BranchInfo, error) {
	branch := &Branch{}
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
	directProv, err := GetDirectBranchProv(ctx, tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "could not get direct branch provenance")
	}
	fullProv, err := GetBranchProv(ctx, tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "could not get full branch provenance")
	}
	return &pfs.BranchInfo{
		Branch:           branch.Pb(),
		Head:             branch.Head.Pb(),
		Provenance:       fullProv,
		DirectProvenance: directProv,
	}, nil
}

// UpsertBranch creates a branch if it does not exist, or updates the head commit if it does.
func UpsertBranch(ctx context.Context, tx *pachsql.Tx, branchInfo *pfs.BranchInfo) (BranchID, error) {
	if branchInfo.Branch.Repo.Name == "" {
		return 0, errors.Errorf("repo name required")
	}
	if branchInfo.Branch.Repo.Type == "" {
		return 0, errors.Errorf("repo type required")
	}
	if branchInfo.Branch.Repo.Project.Name == "" {
		return 0, errors.Errorf("project name required")
	}
	if branchInfo.Head.Id == "" {
		return 0, errors.Errorf("head commit required")
	}
	var branchID BranchID
	if err := tx.QueryRowContext(ctx,
		`
		INSERT INTO pfs.branches(repo_id, name, head)
		VALUES (
			(SELECT repo.id FROM pfs.repos repo JOIN core.projects project ON repo.project_id = project.id WHERE project.name = $1 AND repo.name = $2 AND repo.type = $3),
			$4,
			(SELECT int_id FROM pfs.commits WHERE commit_id = $5)
		)
		ON CONFLICT (repo_id, name) DO UPDATE SET head = EXCLUDED.head
		RETURNING id
		`,
		branchInfo.Branch.Repo.Project.Name,
		branchInfo.Branch.Repo.Name,
		branchInfo.Branch.Repo.Type,
		branchInfo.Branch.Name,
		CommitKey(branchInfo.Head),
	).Scan(&branchID); err != nil {
		return 0, errors.Wrap(err, "could not create branch")
	}
	return branchID, nil
}

func GetBranchID(ctx context.Context, tx *pachsql.Tx, branch *pfs.Branch) (BranchID, error) {
	var id BranchID
	if err := tx.QueryRowContext(ctx, `
		SELECT branch.id
		FROM pfs.branches branch
			JOIN pfs.repos repo ON branch.repo_id = repo.id
			JOIN core.projects project ON repo.project_id = project.id
		WHERE project.name = $1 AND repo.name = $2 AND repo.type = $3 AND branch.name = $4
	`, branch.Repo.Project.Name, branch.Repo.Name, branch.Repo.Type, branch.Name,
	).Scan(&id); err != nil {
		return 0, errors.Wrapf(err, "could not get id for branch %v", branch)
	}
	return id, nil
}

func GetDirectBranchProv(ctx context.Context, tx *pachsql.Tx, id BranchID) ([]*pfs.Branch, error) {
	var branches []Branch
	if err := tx.SelectContext(ctx, &branches, `
		SELECT
			branch.id,
			branch.name,
			repo.name as "repo.name",
			repo.type as "repo.type",
			project.name as "repo.project.name"
		FROM pfs.branch_provenance bp
		    JOIN pfs.branches branch ON bp.to_id = branch.id
			JOIN pfs.repos repo ON branch.repo_id = repo.id
			JOIN core.projects project ON repo.project_id = project.id
		WHERE bp.from_id = $1
	`, id); err != nil {
		return nil, errors.Wrap(err, "could not get direct branch provenance")
	}
	var branchPbs []*pfs.Branch
	for _, branch := range branches {
		branchPbs = append(branchPbs, branch.Pb())
	}
	return branchPbs, nil
}

func GetBranchProv(ctx context.Context, tx *pachsql.Tx, id BranchID) ([]*pfs.Branch, error) {
	var branches []Branch
	if err := tx.SelectContext(ctx, &branches, `
		WITH RECURSIVE prov(from_id, to_id) AS (
		    SELECT from_id, to_id
		    FROM pfs.branch_provenance
		    WHERE from_id = $1
		  UNION ALL
		    SELECT bp.from_id, bp.to_id
		    FROM prov JOIN pfs.branch_provenance bp ON prov.to_id = bp.from_id
		)
		SELECT DISTINCT
		    branch.id,
			branch.name,
			repo.name as "repo.name",
			repo.type as "repo.type",
			project.name as "repo.project.name"
		FROM pfs.branches branch
		    JOIN prov ON branch.id = prov.to_id
			JOIN pfs.repos repo ON branch.repo_id = repo.id
		    JOIN core.projects project ON repo.project_id = project.id
		WHERE branch.id != $1
	`, id); err != nil {
		return nil, errors.Wrap(err, "could not get branch provenance")
	}
	var branchPbs []*pfs.Branch
	for _, branch := range branches {
		branchPbs = append(branchPbs, branch.Pb())
	}
	return branchPbs, nil
}

func AddDirectBranchProv(ctx context.Context, tx *pachsql.Tx, from, to *pfs.Branch) error {
	fromID, err := GetBranchID(ctx, tx, from)
	if err != nil {
		return errors.Wrap(err, "could not get from_id")
	}
	toID, err := GetBranchID(ctx, tx, to)
	if err != nil {
		return errors.Wrap(err, "could not get to_id")
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO pfs.branch_provenance(from_id, to_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, fromID, toID); err != nil {
		return errors.Wrap(err, "could not add branch provenance")
	}
	return nil
}
