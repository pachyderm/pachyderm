package pfsdb

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// SliceDiff takes two slices and returns the elements in the first slice that are not in the second slice.
// TODO this can be moved to a more generic package.
func SliceDiff[K comparable, V any](a []V, b []V, key func(V) K) []V {
	m := make(map[K]bool)
	for _, item := range b {
		m[key(item)] = true
	}
	var result []V
	for _, item := range a {
		if !m[key(item)] {
			result = append(result, item)
		}
	}
	return result
}

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
	directProv, err := GetDirectBranchProvenance(ctx, tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "could not get direct branch provenance")
	}
	fullProv, err := GetBranchProvenance(ctx, tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "could not get full branch provenance")
	}
	fullSubv, err := GetBranchSubvenance(ctx, tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "could not get full branch subvenance")
	}
	return &pfs.BranchInfo{
		Branch:           branch.Pb(),
		Head:             branch.Head.Pb(),
		DirectProvenance: directProv,
		Provenance:       fullProv,
		Subvenance:       fullSubv,
	}, nil
}

// UpsertBranch creates a branch if it does not exist, or updates the head if the branch already exists.
// If direct provenance is specified, it will be used to update the branch's provenance relationships.
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
	// Compare current direct provenance to new direct provenance.
	oldDirectProv, err := GetDirectBranchProvenance(ctx, tx, branchID)
	if err != nil {
		return 0, errors.Wrap(err, "could not get direct branch provenance")
	}
	// Add net new direct provenance relationships.
	toAdd := SliceDiff[string, *pfs.Branch](branchInfo.DirectProvenance, oldDirectProv, func(branch *pfs.Branch) string { return branch.Key() })
	for _, branch := range toAdd {
		toID, err := GetBranchID(ctx, tx, branch)
		if err != nil {
			return 0, errors.Wrap(err, "could not get to_id")
		}
		if err := CreateDirectBranchProvenance(ctx, tx, branchID, toID); err != nil {
			return 0, errors.Wrap(err, "could not add direct branch provenance")
		}
	}
	// Remove old direct provenance relationships that are no longer needed.
	toRemove := SliceDiff[string, *pfs.Branch](oldDirectProv, branchInfo.DirectProvenance, func(branch *pfs.Branch) string { return branch.Key() })
	for _, branch := range toRemove {
		toID, err := GetBranchID(ctx, tx, branch)
		if err != nil {
			return 0, errors.Wrap(err, "could not get to_id")
		}
		if err := DeleteDirectBranchProvenance(ctx, tx, branchID, toID); err != nil {
			return 0, errors.Wrap(err, "could not delete direct branch provenance")
		}
	}
	return branchID, nil
}

// GetBranchID returns the unique id of a branch.
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

// GetDirectBranchProvenance returns the direct provenance of a branch, i.e. all branches that it directly depends on.
func GetDirectBranchProvenance(ctx context.Context, tx *pachsql.Tx, id BranchID) ([]*pfs.Branch, error) {
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

// GetBranchProvenance returns the full provenance of a branch, i.e. all branches that it either directly or transitively depends on.
func GetBranchProvenance(ctx context.Context, tx *pachsql.Tx, id BranchID) ([]*pfs.Branch, error) {
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

// GetBranchSubvenance returns the full subvenance of a branch, i.e. all branches that either directly or transitively depend on it.
func GetBranchSubvenance(ctx context.Context, tx *pachsql.Tx, id BranchID) ([]*pfs.Branch, error) {
	var branches []Branch
	if err := tx.SelectContext(ctx, &branches, `
		WITH RECURSIVE subv(from_id, to_id) AS (
		    SELECT from_id, to_id
		    FROM pfs.branch_provenance
		    WHERE to_id = $1
		  UNION ALL
		    SELECT bp.from_id, bp.to_id
		    FROM subv JOIN pfs.branch_provenance bp ON subv.from_id = bp.to_id
		)
		SELECT DISTINCT
		    branch.id,
			branch.name,
			repo.name as "repo.name",
			repo.type as "repo.type",
			project.name as "repo.project.name"
		FROM pfs.branches branch
		    JOIN subv ON branch.id = subv.from_id
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

// AddBranchProvenance adds a provenance relationship between two branches.
func CreateDirectBranchProvenance(ctx context.Context, tx *pachsql.Tx, from, to BranchID) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO pfs.branch_provenance(from_id, to_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, from, to); err != nil {
		return errors.Wrap(err, "could not add branch provenance")
	}
	return nil
}

func CreateDirectBranchProvenanceByName(ctx context.Context, tx *pachsql.Tx, from, to *pfs.Branch) error {
	fromID, err := GetBranchID(ctx, tx, from)
	if err != nil {
		return errors.Wrap(err, "could not get from_id")
	}
	toID, err := GetBranchID(ctx, tx, to)
	if err != nil {
		return errors.Wrap(err, "could not get to_id")
	}
	return CreateDirectBranchProvenance(ctx, tx, fromID, toID)
}

func DeleteDirectBranchProvenance(ctx context.Context, tx *pachsql.Tx, from, to BranchID) error {
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM pfs.branch_provenance
		WHERE from_id = $1 AND to_id = $2
	`, from, to); err != nil {
		return errors.Wrap(err, "could not delete branch provenance")
	}
	return nil
}
