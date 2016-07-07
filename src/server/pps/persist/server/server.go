package server

import (
	"errors"

	"github.com/pachyderm/pachyderm/src/server/pps/persist"
)

var (
	// ErrIDSet is used when ID is set.
	ErrIDSet = errors.New("pachyderm.pps.persist.server: ID set")
	// ErrIDNotSet is used when ID is not set.
	ErrIDNotSet = errors.New("pachyderm.pps.persist.server: ID not set")
	// ErrTimestampSet is used when Timestamp is set.
	ErrTimestampSet = errors.New("pachyderm.pps.persist.server: Timestamp set")
)

// APIServer represents Pachyderm's interface to a persistence server such as rethinkdb.
type APIServer interface {
	persist.APIServer
	Close() error
}

// NewRethinkAPIServer creates an APIServer which connects
// to rethinkdb on the given address and database.
func NewRethinkAPIServer(address string, databaseName string) (APIServer, error) {
	return newRethinkAPIServer(address, databaseName)
}
