package v2_8_0

import (
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"time"
)

type Commit struct {
	IntID          uint64        `db:"int_id"`
	CommitSetID    string        `db:"commit_set_id"`
	CommitID       string        `db:"commit_id"`
	RepoID         uint64        `db:"repo_id"`
	BranchID       string        `db:"branch_id_str"`
	Origin         string        `db:"origin"`
	Description    string        `db:"description"`
	StartTime      time.Time     `db:"start_time"`
	FinishingTime  time.Time     `db:"finishing_time"`
	FinishedTime   time.Time     `db:"finished_time"`
	CompactingTime int64         `db:"compacting_time"`
	ValidatingTime int64         `db:"validating_time"`
	Error          string        `db:"error"`
	Size           int64         `db:"size"`
	RepoName       string        `db:"repo_name"`
	RepoType       string        `db:"repo_type"`
	ProjectName    string        `db:"proj_name"`
	ParentCommit   *pfs.Commit   `db:"parent_commit"`
	ChildCommits   []*pfs.Commit `db:"child_commits"`
	CreatedAt      time.Time     `db:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at"`
}

type Repo struct {
	ID          uint64    `db:"id"`
	Name        string    `db:"name"`
	ProjectName string    `db:"project_name"`
	Description string    `db:"description"`
	RepoType    string    `db:"type"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type Branch struct {
	ID        uint64    `db:"id"`
	Name      string    `db:"name"`
	Head      uint64    `db:"head"`
	RepoID    uint64    `db:"repo_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Edge struct {
	FromID uint64 `db:"from_id"`
	ToID   uint64 `db:"to_id"`
}

type BranchTrigger struct {
	FromBranchID  uint64 `db:"from_branch_id"`
	ToBranchID    uint64 `db:"to_branch_id"`
	CronSpec      string `db:"cron_spec"`
	RateLimitSpec string `db:"rate_limit_spec"`
	Size          string `db:"size"`
	NumCommits    uint64 `db:"num_commits"`
	All           bool   `db:"all_conditions"`
}
