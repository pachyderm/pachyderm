package pfs

import (
	pfs_client "github.com/pachyderm/pachyderm/v2/src/pfs"
)

type APIServer interface {
	ListRepoWithoutAuth(ctx context.Context) (*pfs.ListRepoResponse, error)
}
