package pfs

import (
	"github.com/jmoiron/sqlx"
	pfsgrpc "github.com/pachyderm/pachyderm/v2/src/pfs"
)

// We can add type aliases to improve ergonomics.
// Internal packages just depend on this pfs package instead of the other one.
type (
	CreateRepoRequest     = pfsgrpc.CreateRepoRequest
	ListRepoRequest       = pfsgrpc.ListRepoRequest
	ListRepoResponse      = pfsgrpc.ListRepoResponse
	ClearCommitRequest    = pfsgrpc.ClearCommitRequest
	StartCommitRequest    = pfsgrpc.StartCommitRequest
	AddFilesetRequest     = pfsgrpc.AddFilesetRequest
	GetFilesetRequest     = pfsgrpc.GetFilesetRequest
	CreateFilesetResponse = pfsgrpc.CreateFilesetResponse
)

// API extends the gRPC generated PFS API to add methods that can be executed in a transaction
type API interface {
	pfsgrpc.APIServer

	// Repo
	CreateRepoTx(tx *sqlx.Tx, req CreateRepoRequest) error
	// ...
	ListRepoTx(tx *sqlx.Tx, req ListRepoRequest) (*ListRepoResponse, error)

	// Commit
	StartCommitTx(tx *sqlx.Tx, req StartCommitRequest) error
	// ...
	ClearCommitTx(tx *sqlx.Tx, req ClearCommitRequest) error

	// Fileset
	AddFilesetTx(tx *sqlx.Tx, req AddFilesetRequest) error
	GetFilesetTx(tx *sqlx.Tx, req GetFilesetRequest) (*CreateFilesetResponse, error)
}
