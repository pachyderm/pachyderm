package server

import (
    "fmt"
    "sort"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

type Operation interface {
    // TODO: probably don't need to pass the full client here
    validate(pachClient *client.APIClient, d *driver) error
    execute(pachClient *client.APIClient, d *driver, stm col.STM) error
}

type CreateBranchOp struct {
    branch *pfs.Branch
    commit *pfs.Commit
    provenance []*pfs.Branch
}

func (op *CreateBranchOp) validate(pachClient *client.APIClient, d *driver) error {
	if err := d.checkIsAuthorized(pachClient, op.branch.Repo, auth.Scope_WRITER); err != nil {
		return err
	}

	// The request must do exactly one of:
	// 1) updating 'branch's provenance (commit is nil OR commit == branch)
	// 2) re-pointing 'branch' at a new commit
	if op.commit != nil {
		// Determine if this is a provenance update
		sameTarget := op.branch.Repo.Name == op.commit.Repo.Name && op.branch.Name == op.commit.ID
		if !sameTarget && op.provenance != nil {
			return fmt.Errorf("cannot point branch \"%s\" at target commit \"%s/%s\" without clearing its provenance",
				op.branch.Name, op.commit.Repo.Name, op.commit.ID)
		}
	}

    return nil
}

func (op CreateBranchOp) execute(pachClient *client.APIClient, d *driver, stm col.STM) error {
    // if 'commit' is a branch, resolve it
    var err error
    if op.commit != nil {
        _, err = d.resolveCommit(stm, op.commit) // if 'commit' is a branch, resolve it
        if err != nil {
            // possible that branch exists but has no head commit. This is fine, but
            // branchInfo.Head must also be nil
            if !isNoHeadErr(err) {
                return fmt.Errorf("unable to inspect %s/%s: %v", err, op.commit.Repo.Name, op.commit.ID)
            }
            op.commit = nil
        }
    }

    // Retrieve (and create, if necessary) the current version of this branch
    branches := d.branches(op.branch.Repo.Name).ReadWrite(stm)
    branchInfo := &pfs.BranchInfo{}
    if err := branches.Upsert(op.branch.Name, branchInfo, func() error {
        branchInfo.Name = op.branch.Name // set in case 'branch' is new
        branchInfo.Branch = op.branch
        branchInfo.Head = op.commit
        branchInfo.DirectProvenance = nil
        for _, provBranch := range op.provenance {
            add(&branchInfo.DirectProvenance, provBranch)
        }
        return nil
    }); err != nil {
        return err
    }
    repos := d.repos.ReadWrite(stm)
    repoInfo := &pfs.RepoInfo{}
    if err := repos.Update(op.branch.Repo.Name, repoInfo, func() error {
        add(&repoInfo.Branches, op.branch)
        return nil
    }); err != nil {
        return err
    }

    // Update (or create)
    // 1) 'branch's Provenance
    // 2) the Provenance of all branches in 'branch's Subvenance (in the case of an update), and
    // 3) the Subvenance of all branches in the *old* provenance of 'branch's Subvenance
    toUpdate := []*pfs.BranchInfo{branchInfo}
    for _, subvBranch := range branchInfo.Subvenance {
        subvBranchInfo := &pfs.BranchInfo{}
        if err := d.branches(subvBranch.Repo.Name).ReadWrite(stm).Get(subvBranch.Name, subvBranchInfo); err != nil {
            return err
        }
        toUpdate = append(toUpdate, subvBranchInfo)
    }
    // Sorting is important here because it sorts topologically. This means
    // that when evaluating element i of `toUpdate` all elements < i will
    // have already been evaluated and thus we can safely use their
    // Provenance field.
    sort.Slice(toUpdate, func(i, j int) bool { return len(toUpdate[i].Provenance) < len(toUpdate[j].Provenance) })
    for _, branchInfo := range toUpdate {
        oldProvenance := branchInfo.Provenance
        branchInfo.Provenance = nil
        // Re-compute Provenance
        for _, provBranch := range branchInfo.DirectProvenance {
            if err := d.addBranchProvenance(branchInfo, provBranch, stm); err != nil {
                return err
            }
            provBranchInfo := &pfs.BranchInfo{}
            if err := d.branches(provBranch.Repo.Name).ReadWrite(stm).Get(provBranch.Name, provBranchInfo); err != nil {
                return err
            }
            for _, provBranch := range provBranchInfo.Provenance {
                if err := d.addBranchProvenance(branchInfo, provBranch, stm); err != nil {
                    return err
                }
            }
        }
        if err := d.branches(branchInfo.Branch.Repo.Name).ReadWrite(stm).Put(branchInfo.Branch.Name, branchInfo); err != nil {
            return err
        }
        // Update Subvenance of 'branchInfo's Provenance (incl. all Subvenance)
        for _, oldProvBranch := range oldProvenance {
            if !has(&branchInfo.Provenance, oldProvBranch) {
                // Provenance was deleted, so we delete ourselves from their subvenance
                oldProvBranchInfo := &pfs.BranchInfo{}
                if err := d.branches(oldProvBranch.Repo.Name).ReadWrite(stm).Update(oldProvBranch.Name, oldProvBranchInfo, func() error {
                    del(&oldProvBranchInfo.Subvenance, branchInfo.Branch)
                    return nil
                }); err != nil {
                    return err
                }
            }
        }
    }

    // propagate the head commit to 'branch'. This may also modify 'branch', by
    // creating a new HEAD commit if 'branch's provenance was changed and its
    // current HEAD commit has old provenance
    return d.propagateCommit(stm, op.branch)
}
