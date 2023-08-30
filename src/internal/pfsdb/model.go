package pfsdb

import (
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"
)

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

type Branch struct {
	ID   BranchID `db:"id"`
	Name string   `db:"name"`
	Head uint64   `db:"head"`
	Repo Repo     `db:"repo"`
	CreatedAtUpdatedAt
}
