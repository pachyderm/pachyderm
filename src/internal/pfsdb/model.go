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

type Repo struct {
	Project     coredb.Project `db:"project"`
	ID          RepoID         `db:"id"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	Type        string         `db:"type"`
	CreatedAtUpdatedAt
}

func (repo *Repo) RepoPb() *pfs.Repo {
	return &pfs.Repo{
		Name:    repo.Name,
		Type:    repo.Type,
		Project: repo.Project.ProjectPb(),
	}
}

type Commit struct {
	ID          CommitID `db:"id"`
	Repo        Repo     `db:"repo"`
	CommitSetID string   `db:"commit_set_id"`
	CreatedAtUpdatedAt
}

func (commit *Commit) CommitPb() *pfs.Commit {
	return &pfs.Commit{
		Id:   commit.CommitSetID,
		Repo: commit.Repo.RepoPb(),
	}
}

type Branch struct {
	ID   BranchID `db:"id"`
	Head Commit   `db:"head"`
	Repo Repo     `db:"repo"`
	Name string   `db:"name"`
	CreatedAtUpdatedAt
}

func (branch *Branch) BranchPb() *pfs.Branch {
	return &pfs.Branch{
		Name: branch.Name,
		Repo: branch.Repo.RepoPb(),
	}
}
