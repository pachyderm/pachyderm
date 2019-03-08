package server

import (
    "fmt"
    "sort"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

func (req *CreateBranchOp) validate(d *driver, pachClient *client.APIClient) error {
	if err := d.checkIsAuthorized(pachClient, req.Branch.Repo, auth.Scope_WRITER); err != nil {
		return err
	}

	// The request must do exactly one of:
	// 1) updating 'branch's provenance (commit is nil OR commit == branch)
	// 2) re-pointing 'branch' at a new commit
	if req.Head != nil {
		// Determine if this is a provenance update
		sameTarget := req.Branch.Repo.Name == req.Head.Repo.Name && req.Branch.Name == req.Head.ID
		if !sameTarget && req.Provenance != nil {
			return fmt.Errorf("cannot point branch \"%s\" at target commit \"%s/%s\" without clearing its provenance",
				req.Branch.Name, req.Head.Repo.Name, req.Head.ID)
		}
	}

    return nil
}

func (req *CreateBranchOp) execute(d *driver, pachClient *client.APIClient, stm col.STM) error {
    // if 'commit' is a branch, resolve it
    var err error
    if req.Head != nil {
        _, err = d.resolveCommit(stm, req.Head) // if 'commit' is a branch, resolve it
        if err != nil {
            // possible that branch exists but has no head commit. This is fine, but
            // branchInfo.Head must also be nil
            if !isNoHeadErr(err) {
                return fmt.Errorf("unable to inspect %s/%s: %v", err, req.Head.Repo.Name, req.Head.ID)
            }
            req.Head = nil
        }
    }

    // Retrieve (and create, if necessary) the current version of this branch
    branches := d.branches(req.Branch.Repo.Name).ReadWrite(stm)
    branchInfo := &pfs.BranchInfo{}
    if err := branches.Upsert(req.Branch.Name, branchInfo, func() error {
        branchInfo.Name = req.Branch.Name // set in case 'branch' is new
        branchInfo.Branch = req.Branch
        branchInfo.Head = req.Head
        branchInfo.DirectProvenance = nil
        for _, provBranch := range req.Provenance {
            add(&branchInfo.DirectProvenance, provBranch)
        }
        return nil
    }); err != nil {
        return err
    }
    repos := d.repos.ReadWrite(stm)
    repoInfo := &pfs.RepoInfo{}
    if err := repos.Update(req.Branch.Repo.Name, repoInfo, func() error {
        add(&repoInfo.Branches, req.Branch)
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
    return d.propagateCommit(stm, req.Branch)
}
