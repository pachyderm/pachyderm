package pps

import "github.com/pachyderm/pachyderm/src/client/pfs"

type FilterFunc func([]*pfs.FileInfo) bool
