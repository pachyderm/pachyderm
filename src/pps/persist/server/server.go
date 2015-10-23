package server // import "go.pachyderm.com/pachyderm/src/pps/persist/server"

import (
	"errors"

	"go.pachyderm.com/pachyderm/src/pps/persist"
)

var (
	ErrIDSet        = errors.New("pachyderm.pps.persist.server: ID set")
	ErrIDNotSet     = errors.New("pachyderm.pps.persist.server: ID not set")
	ErrTimestampSet = errors.New("pachyderm.pps.persist.server: Timestamp set")
)

type APIServer interface {
	persist.APIServer
	Close() error
}

func NewRethinkAPIServer(address string, databaseName string) (APIServer, error) {
	return newRethinkAPIServer(address, databaseName)
}
