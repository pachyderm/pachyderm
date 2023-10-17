package pfsdb

import (
	"database/sql"
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

type Commit struct {
	ID             CommitID      `db:"int_id"`
	CommitSetID    string        `db:"commit_set_id"`
	CommitID       string        `db:"commit_id"`
	Origin         string        `db:"origin"`
	Description    string        `db:"description"`
	StartTime      sql.NullTime  `db:"start_time"`
	FinishingTime  sql.NullTime  `db:"finishing_time"`
	FinishedTime   sql.NullTime  `db:"finished_time"`
	CompactingTime sql.NullInt64 `db:"compacting_time_s"`
	ValidatingTime sql.NullInt64 `db:"validating_time_s"`
	Error          string        `db:"error"`
	Size           int64         `db:"size"`
	// BranchName is used to derive the BranchID in commit related queries.
	BranchName sql.NullString `db:"branch_name"`
	BranchID   sql.NullInt64  `db:"branch_id"`
	Repo       Repo           `db:"repo"`
	CreatedAtUpdatedAt
}

func (commit Commit) GetCreatedAtUpdatedAt() CreatedAtUpdatedAt {
	return commit.CreatedAtUpdatedAt
}

func (commit *Commit) Pb() *pfs.Commit {
	pb := &pfs.Commit{
		Id:   commit.CommitSetID,
		Repo: commit.Repo.Pb(),
	}
	if commit.BranchName.Valid {
		pb.Branch = &pfs.Branch{
			Repo: pb.Repo,
			Name: commit.BranchName.String,
		}
	}
	return pb
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

type BranchTrigger struct {
	FromBranch    Branch `db:"from_branch"`
	ToBranch      Branch `db:"to_branch"`
	CronSpec      string `db:"cron_spec"`
	RateLimitSpec string `db:"rate_limit_spec"`
	Size          string `db:"size"`
	NumCommits    int64  `db:"num_commits"`
	AllConditions bool   `db:"all_conditions"`
}

func (trigger *BranchTrigger) Pb() *pfs.Trigger {
	return &pfs.Trigger{
		Branch:        trigger.ToBranch.Name,
		CronSpec:      trigger.CronSpec,
		RateLimitSpec: trigger.RateLimitSpec,
		Size:          trigger.Size,
		Commits:       trigger.NumCommits,
		All:           trigger.AllConditions,
	}
}
