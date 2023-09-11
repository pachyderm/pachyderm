package v2_8_0

import "time"

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
