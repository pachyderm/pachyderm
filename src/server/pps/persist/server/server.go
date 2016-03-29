package server

import (
	"errors"

	"github.com/pachyderm/pachyderm/src/server/pps/persist"
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
