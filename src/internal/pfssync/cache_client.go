package pfssync

import (
	io "io"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

type CacheClient struct {
	*client.APIClient
	mu      sync.Mutex
	cache   *simplelru.LRU
	renewer *renew.StringSet
}

// TODO: Expose configuration for cache size?
// TODO: Dedupe work?
func NewCacheClient(pachClient *client.APIClient, renewer *renew.StringSet) *CacheClient {
	cc := &CacheClient{
		APIClient: pachClient,
		renewer:   renewer,
	}
	cache, err := simplelru.NewLRU(100, cc.onEvicted)
	if err != nil {
		// lru.NewWithEvict only errors for size < 1
		panic(err)
	}
	cc.cache = cache
	return cc
}

func (cc *CacheClient) GetFileTAR(commit *pfs.Commit, path string) (io.ReadCloser, error) {
	key := pfsdb.CommitKey(commit)
	cc.mu.Lock()
	if c, ok := cc.cache.Get(key); ok {
		cc.mu.Unlock()
		return cc.APIClient.GetFileTAR(c.(*pfs.Commit), path)
	}
	cc.mu.Unlock()
	id, err := cc.APIClient.GetFileSet(commit.Branch.Repo.Name, commit.Branch.Name, commit.ID)
	if err != nil {
		return nil, err
	}
	cc.renewer.Add(id)
	commit = client.NewCommit(client.FileSetsRepoName, "", id)
	cc.mu.Lock()
	cc.cache.Add(key, commit)
	cc.mu.Unlock()
	return cc.APIClient.GetFileTAR(commit, path)
}

func (cc *CacheClient) onEvicted(key, value interface{}) {
	cc.renewer.Remove(value.(*pfs.Commit).ID)
}
