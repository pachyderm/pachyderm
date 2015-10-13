package persist // import "go.pachyderm.com/pachyderm/src/pps/persist"

import "errors"

var (
	ErrIDSet    = errors.New("pachyderm.pps.persist: ID set")
	ErrIDNotSet = errors.New("pachyderm.pps.persist: ID not set")
)

func NewRethinkAPIServer(address string, databaseName string) (APIServer, error) {
	return newRethinkAPIServer(address, databaseName)
}
