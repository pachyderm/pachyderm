package v2_8_0

import (
	"database/sql/driver"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type Repo struct {
	ID          uint64    `db:"id"`
	Name        string    `db:"name"`
	ProjectName string    `db:"project_name"`
	Description string    `db:"description"`
	RepoType    string    `db:"type"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type NullUint64 struct {
	Uint64 uint64
	Valid  bool
}

func (n *NullUint64) Scan(value any) error {
	if value == nil {
		n.Uint64, n.Valid = 0, false
		return nil
	}
	n.Valid = true
	switch v := value.(type) {
	case uint64:
		n.Uint64 = v
		return nil
	case int64:
		n.Uint64 = uint64(v)
		return nil
	}
	return errors.Errorf("invalid type for NullUint64: %T", value)
}

func (n *NullUint64) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Uint64, nil
}

type Branch struct {
	// Key       string    `db:"key"`
	ID        uint64     `db:"id"`
	Name      string     `db:"name"`
	Head      uint64     `db:"head"`
	RepoID    uint64     `db:"repo_id"`
	TriggerID NullUint64 `db:"trigger_id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

type Edge struct {
	FromID uint64 `db:"from_id"`
	ToID   uint64 `db:"to_id"`
}

type BranchTrigger struct {
	BranchID      uint64 `db:"branch_id"`
	CronSpec      string `db:"cron_spec"`
	RateLimitSpec string `db:"rate_limit_spec"`
	Size          string `db:"size"`
	NumCommits    uint64 `db:"num_commits"`
	All           bool   `db:"all_conditions"`
}
