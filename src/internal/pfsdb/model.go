package pfsdb

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pbutil"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/pgjsontypes"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// A separate type is defined for safety so row ids must be explicitly cast for use in another table.

// ProjectID is the row id for a project entry in postgres.
type ProjectID uint64

// RepoID is the row id for a repo entry in postgres.
type RepoID uint64

// CommitID is the row id for a commit entry in postgres.
type CommitID uint64

// BranchID is the row id for a branch entry in postgres.
type BranchID uint64

type CreatedAtUpdatedAt struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ProjectRow struct {
	ID          ProjectID             `db:"id"`
	Name        string                `db:"name"`
	Description string                `db:"description"`
	Metadata    pgjsontypes.StringMap `db:"metadata"`
	CreatedAtUpdatedAt
	CreatedBy sql.NullString `db:"created_by"`
}

func (project *ProjectRow) Pb() *pfs.Project {
	return &pfs.Project{
		Name: project.Name,
	}
}

func (project *ProjectRow) PbInfo() *pfs.ProjectInfo {
	return &pfs.ProjectInfo{
		Project: &pfs.Project{
			Name: project.Name,
		},
		Description: project.Description,
		CreatedAt:   timestamppb.New(project.CreatedAt),
		Metadata:    project.Metadata.Data,
		CreatedBy:   project.CreatedBy.String,
	}
}

func (project ProjectRow) GetCreatedAtUpdatedAt() CreatedAtUpdatedAt {
	return project.CreatedAtUpdatedAt
}

// RepoRow is a row in the pfs.repos table.
type RepoRow struct {
	ID          RepoID     `db:"id"`
	Project     ProjectRow `db:"project"`
	Name        string     `db:"name"`
	Type        string     `db:"type"`
	Description string     `db:"description"`
	CreatedAtUpdatedAt
	BranchesNames string                `db:"branches"`
	Metadata      pgjsontypes.StringMap `db:"metadata"`
}

func (repo RepoRow) GetCreatedAtUpdatedAt() CreatedAtUpdatedAt {
	return repo.CreatedAtUpdatedAt
}

func (repo *RepoRow) Pb() *pfs.Repo {
	return &pfs.Repo{
		Name:    repo.Name,
		Type:    repo.Type,
		Project: repo.Project.Pb(),
	}
}

func (repo *RepoRow) PbInfo() (*pfs.RepoInfo, error) {
	branches, err := parseBranches(repo)
	if err != nil {
		return nil, err
	}
	return &pfs.RepoInfo{
		Repo:        repo.Pb(),
		Description: repo.Description,
		Branches:    branches,
		Created:     timestamppb.New(repo.CreatedAt),
		Metadata:    repo.Metadata.Data,
	}, nil
}

func parseBranches(repo *RepoRow) ([]*pfs.Branch, error) {
	var branches []*pfs.Branch
	if repo.BranchesNames == noBranches {
		return branches, nil
	}
	// after aggregation, braces, quotes, and leading hex prefixes need to be removed from the encoded branch strings.
	for _, branchName := range strings.Split(strings.Trim(repo.BranchesNames, "{}"), ",") {
		branch := &pfs.Branch{
			Name: branchName,
			Repo: repo.Pb(),
		}
		branches = append(branches, branch)
	}
	return branches, nil
}

type CommitRow struct {
	ID    CommitID `db:"int_id"`
	SetID string   `db:"commit_set_id"`

	// todo(fahad): write a migration that changes the 'commit_id' column to key to avoid confusion.
	// Key is the human readable string identifier for a commit.
	Key            string                `db:"commit_id"`
	Origin         string                `db:"origin"`
	Description    string                `db:"description"`
	StartTime      sql.NullTime          `db:"start_time"`
	FinishingTime  sql.NullTime          `db:"finishing_time"`
	FinishedTime   sql.NullTime          `db:"finished_time"`
	CompactingTime sql.NullInt64         `db:"compacting_time_s"`
	ValidatingTime sql.NullInt64         `db:"validating_time_s"`
	Error          string                `db:"error"`
	Size           int64                 `db:"size"`
	Metadata       pgjsontypes.StringMap `db:"metadata"`
	// BranchName is used to derive the BranchID in commit related queries.
	BranchName sql.NullString `db:"branch_name"`
	BranchID   sql.NullInt64  `db:"branch_id"`
	Repo       RepoRow        `db:"repo"`
	CreatedAtUpdatedAt
}

func (c *CommitRow) ToCommit(ctx context.Context, extCtx sqlx.ExtContext) (*Commit, error) {
	if c == nil {
		return nil, nil
	}
	commit := &Commit{
		CommitRow: c,
	}
	if c.ID == 0 {
		return nil, errors.New("cannot convert row to commit if row has no id.")
	}
	if c.Repo.ID != 0 && c.Repo.BranchesNames == "" {
		repo, err := getRepo(ctx, extCtx, c.Repo.ID)
		if err != nil {
			return nil, errors.Wrap(err, "get commit from database row: get repo")
		}
		c.Repo = *repo
	}
	parent, children, err := getCommitRelativeRows(ctx, extCtx, c.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get commit from database row")
	}
	subvenance, err := getSubvenantCommitRows(ctx, extCtx, c.ID, WithMaxDepth(1))
	if err != nil {
		return nil, errors.Wrap(err, "get commit from database row")
	}
	provenance, err := getProvenantCommitRows(ctx, extCtx, c.ID, WithMaxDepth(1))
	if err != nil {
		return nil, errors.Wrap(err, "get commit from database row")
	}
	commit.Children = children
	commit.Parent = parent
	commit.DirectProvenance = provenance
	commit.DirectSubvenance = subvenance
	return commit, nil
}

func (c *CommitRow) FromPb(commit *pfs.Commit) {
	c.BranchName = sql.NullString{Valid: false}
	if commit != nil && commit.Repo != nil {
		c.Repo = RepoRow{
			Name: commit.Repo.Name,
			Type: commit.Repo.Type,
			Project: ProjectRow{
				Name: commit.Repo.Project.Name,
			},
		}
	}
	c.Key = CommitKey(commit)
	c.SetID = commit.Id
	if commit.Branch != nil {
		c.BranchName = sql.NullString{String: commit.Branch.Name, Valid: true}
	}
}

func (c *CommitRow) FromPbInfo(commitInfo *pfs.CommitInfo) {
	c.FromPb(commitInfo.Commit)
	if commitInfo.Origin != nil {
		c.Origin = commitInfo.Origin.Kind.String()
	}
	c.StartTime = pbutil.SanitizeTimestampPb(commitInfo.Started)
	c.FinishingTime = pbutil.SanitizeTimestampPb(commitInfo.Finishing)
	c.FinishedTime = pbutil.SanitizeTimestampPb(commitInfo.Finished)
	c.Description = commitInfo.Description
	c.Error = commitInfo.Error
	c.Metadata = pgjsontypes.StringMap{Data: commitInfo.Metadata}
	if commitInfo.Details != nil {
		c.CompactingTime = pbutil.DurationPbToBigInt(commitInfo.Details.CompactingTime)
		c.ValidatingTime = pbutil.DurationPbToBigInt(commitInfo.Details.ValidatingTime)
		c.Size = commitInfo.Details.SizeBytes
	}
}

func (c CommitRow) GetCreatedAtUpdatedAt() CreatedAtUpdatedAt {
	return c.CreatedAtUpdatedAt
}

func (c *CommitRow) Pb() *pfs.Commit {
	if c == nil {
		return nil
	}
	pb := &pfs.Commit{
		Id:   c.SetID,
		Repo: c.Repo.Pb(),
	}
	if c.BranchName.Valid {
		pb.Branch = &pfs.Branch{
			Repo: pb.Repo,
			Name: c.BranchName.String,
		}
	}
	return pb
}

func (c *CommitRow) CommitOrigin() pfs.CommitOrigin {
	return pfs.CommitOrigin{Kind: pfs.OriginKind(pfs.OriginKind_value[strings.ToUpper(c.Origin)])}
}

// BranchRow is a row in the pfs.branches table.
type BranchRow struct {
	ID       BranchID              `db:"id"`
	Head     CommitRow             `db:"head"`
	Repo     RepoRow               `db:"repo"`
	Name     string                `db:"name"`
	Metadata pgjsontypes.StringMap `db:"metadata"`
	CreatedAtUpdatedAt
}

func (branch BranchRow) GetCreatedAtUpdatedAt() CreatedAtUpdatedAt {
	return branch.CreatedAtUpdatedAt
}

func (branch *BranchRow) Pb() *pfs.Branch {
	return &pfs.Branch{
		Name: branch.Name,
		Repo: branch.Repo.Pb(),
	}
}

type BranchTrigger struct {
	FromBranch    BranchRow `db:"from_branch"`
	ToBranch      BranchRow `db:"to_branch"`
	CronSpec      string    `db:"cron_spec"`
	RateLimitSpec string    `db:"rate_limit_spec"`
	Size          string    `db:"size"`
	NumCommits    int64     `db:"num_commits"`
	AllConditions bool      `db:"all_conditions"`
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
