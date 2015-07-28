package fuse

import (
	"github.com/pachyderm/pachyderm/src/pfs"
)

type Mounter interface {
	Mount(apiClient pfs.ApiClient, repositoryName string, mountPoint string) error
}

func NewMounter() Mounter {
	return newMounter()
}
