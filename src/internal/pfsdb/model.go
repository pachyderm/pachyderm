package pfsdb

import (
	"encoding/hex"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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

func (repo *Repo) PbInfo() (*pfs.RepoInfo, error) {
	branches, err := parseBranches(repo.Branches)
	if err != nil {
		return nil, err
	}
	return &pfs.RepoInfo{
		Repo:        repo.Pb(),
		Description: repo.Description,
		Branches:    branches,
		Created:     timestamppb.New(repo.CreatedAt),
	}, nil
}

func parseBranches(branchInfos string) ([]*pfs.Branch, error) {
	var branches []*pfs.Branch
	if branchInfos == noBranches {
		return branches, nil
	}
	// after aggregation, braces, quotes, and leading hex prefixes need to be removed from the encoded branch strings.
	for _, branchStr := range strings.Split(strings.Trim(branchInfos, "{}"), ",") {
		branchHex := strings.Trim(strings.Trim(branchStr, "\""), "\\x")
		decodedString, err := hex.DecodeString(branchHex)
		if err != nil {
			return nil, errors.Wrap(err, "branch not hex encoded")
		}
		branchInfo := &pfs.BranchInfo{}
		if err := proto.Unmarshal(decodedString, branchInfo); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal BranchInfo")
		}
		branches = append(branches, branchInfo.Branch)
	}
	return branches, nil
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

func (branch *Branch) PbInfo() *pfs.BranchInfo {
	return &pfs.BranchInfo{
		Branch: branch.Pb(),
		Head:   branch.Head.Pb(),
	}
}
