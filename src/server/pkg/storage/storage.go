package storage

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/obj"
)

// Shares are a simple way to control and evenly
// split resource usage at the memory/disk caching tiers.
// The number of shares determines the max concurrency in terms
// of number of file sets that can be operated on at a time.
// Memory and disk is evenly distributed between the shares.
type Share struct {
	mem, disk int64
}

type Storage struct {
	prefix               string
	client               obj.Client
	share                Share
	numShares, maxShares int64
	mu                   sync.Mutex
}

func New(prefix string, memAvailable, diskAvailable, maxShares int64) *Storage {
	// Setup obj storage client and shares.
	// May make sense to instead use ServiceEnv here with obj client.
}

func (s *Storage) NewFileSet(name string, opts ...fileset.Option) *fileset.FileSet {
	// Block if no shares are available.
	// Setup opt for scratch space.
	return fileset.New(name, opts...)
}

func (s *Storage) ListFiles(commit, prefix string) (fileset.Reader, error) {

}

func (s *Storage) GetFiles(commit, prefix string) (fileset.Reader, error) {

}
