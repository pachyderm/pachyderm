package pfssync

import (
	io "io"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
	cache, err := simplelru.NewLRU(10, nil)
	if err != nil {
		// lru.NewWithEvict only errors for size < 1
		panic(err)
	}
	cc.cache = cache
	return cc
}

func (cc *CacheClient) GetFileTAR(commit *pfs.Commit, path string) (io.ReadCloser, error) {
	key := pfsdb.CommitKey(commit)
	if c, ok := cc.get(key); ok {
		return cc.APIClient.GetFileTAR(c, path)
	}
	branch := ""
	if commit.Branch != nil {
		branch = commit.Branch.Name
	}
	id, err := cc.APIClient.GetFileSet(commit.Repo.Project.GetName(), commit.Repo.Name, branch, commit.Id)
	if err != nil {
		return nil, err
	}
	if err := cc.renewer.Add(cc.APIClient.Ctx(), id); err != nil {
		return nil, err
	}
	commit = client.NewCommit(commit.Repo.Project.GetName(), client.FileSetsRepoName, "", id)
	cc.put(key, commit)
	return cc.APIClient.GetFileTAR(commit, path)
}

func (cc *CacheClient) get(key string) (*pfs.Commit, bool) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	c, ok := cc.cache.Get(key)
	if !ok {
		return nil, ok
	}
	return c.(*pfs.Commit), ok
}

func (cc *CacheClient) put(key string, commit *pfs.Commit) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.cache.Add(key, commit)
}

func (cc *CacheClient) WithCreateFileSetClient(cb func(client.ModifyFile) error) (*pfs.CreateFileSetResponse, error) {
	resp, err := cc.APIClient.WithCreateFileSetClient(func(mf client.ModifyFile) error {
		ccfsc := &cacheCreateFileSetClient{
			ModifyFile:  mf,
			CacheClient: cc,
		}
		return cb(ccfsc)
	})
	return resp, errors.EnsureStack(err)
}

type cacheCreateFileSetClient struct {
	client.ModifyFile
	*CacheClient
}

func (ccfsc *cacheCreateFileSetClient) CopyFile(dst string, src *pfs.File, opts ...client.CopyFileOption) error {
	newSrc := &pfs.File{
		Path:  src.Path,
		Datum: src.Datum,
	}
	key := pfsdb.CommitKey(src.Commit)
	if c, ok := ccfsc.get(key); ok {
		newSrc.Commit = c
		return errors.EnsureStack(ccfsc.ModifyFile.CopyFile(dst, newSrc, opts...))
	}
	id, err := ccfsc.APIClient.GetFileSet(src.Commit.Repo.Project.GetName(), src.Commit.Repo.Name, "", src.Commit.Id)
	if err != nil {
		return err
	}
	if err := ccfsc.CacheClient.renewer.Add(ccfsc.APIClient.Ctx(), id); err != nil {
		return err
	}
	newSrc.Commit = client.NewCommit(src.Commit.Repo.Project.GetName(), client.FileSetsRepoName, "", id)
	ccfsc.put(key, newSrc.Commit)
	return errors.EnsureStack(ccfsc.ModifyFile.CopyFile(dst, newSrc, opts...))
}
