package v2_7_0

import "time"

// CollectionRecord is a record in a collections table.
type CollectionRecord struct {
	Key       string    `db:"key"`
	Proto     []byte    `db:"proto"`
	CreatedAt time.Time `db:"createdat"`
	UpdatedAt time.Time `db:"updatedat"`
}

type Project struct {
	ID          uint64    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
