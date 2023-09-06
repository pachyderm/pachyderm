package pfsdb

import (
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// RepoID is the row id for a repo entry in postgres.
// A separate type is defined for safety so row ids must be explicitly cast for use in another table.
type RepoID uint64

// CommitID is the row id for a commit entry in postgres.
type CommitID uint64

// BranchID is the row id for a branch entry in postgres.
type BranchID uint64

type CreatedAtUpdatedAt struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Repo is a row in the pfs.repos table.
type Repo struct {
	ID          RepoID         `db:"id"`
	Project     coredb.Project `db:"project"`
	Name        string         `db:"name"`
	Type        string         `db:"type"`
	Description string         `db:"description"`
	CreatedAtUpdatedAt

	// Branches is a string that contains an array of hex-encoded branchInfos. The array is enclosed with curly braces.
	// Each entry is prefixed with '//x' and entries are delimited by a ','
	Branches string `db:"branches"`
}

func (repo *Repo) Pb() *pfs.Repo {
	return &pfs.Repo{
		Name:    repo.Name,
		Type:    repo.Type,
		Project: repo.Project.Pb(),
	}
}

// Commit is a row in the pfs.commits table.
type Commit struct {
	ID          CommitID `db:"id"`
	Repo        Repo     `db:"repo"`
	CommitSetID string   `db:"commit_set_id"`
	CreatedAtUpdatedAt
}

func (commit *Commit) Pb() *pfs.Commit {
	return &pfs.Commit{
		Id:   commit.CommitSetID,
		Repo: commit.Repo.Pb(),
	}
}

// Branch is a row in the pfs.branches table.
type Branch struct {
	ID   BranchID `db:"id"`
	Head Commit   `db:"head"`
	Repo Repo     `db:"repo"`
	Name string   `db:"name"`
	CreatedAtUpdatedAt
}

func (branch *Branch) Pb() *pfs.Branch {
	return &pfs.Branch{
		Name: branch.Name,
		Repo: branch.Repo.Pb(),
	}
}
