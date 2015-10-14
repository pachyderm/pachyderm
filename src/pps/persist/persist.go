package persist // import "go.pachyderm.com/pachyderm/src/pps/persist"

import "errors"

var (
	ErrIDSet        = errors.New("pachyderm.pps.persist: ID set")
	ErrIDNotSet     = errors.New("pachyderm.pps.persist: ID not set")
	ErrTimestampSet = errors.New("pachyderm.pps.persist: Timestamp set")
)

func NewRethinkAPIServer(address string, databaseName string) (APIServer, error) {
	apiServer, err := newRethinkAPIServer(address, databaseName)
	if err != nil {
		return nil, err
	}
	return newLogAPIServer(apiServer), nil
}
