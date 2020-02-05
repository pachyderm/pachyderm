package cache

import (
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
)

// WorkerCache is an interface for managing hashtree MergeCaches for multiple
// concurrent jobs
type WorkerCache interface {
	GetOrCreateCache(jobID string) (*hashtree.MergeCache, error)
	GetCache(jobID string) *hashtree.MergeCache
	RemoveCache(jobID string) error
}

type workerCache struct {
	// hashtreeStorage is the where we store on disk hashtrees
	hashtreeStorage string

	caches map[string]*hashtree.MergeCache
}

// NewWorkerCache constructs a WorkerCache for maintaining hashtree caches for
// multiple concurrent jobs
func NewWorkerCache(hashtreeStorage string) WorkerCache {
	return &workerCache{
		hashtreeStorage: hashtreeStorage,
		caches:          make(map[string]*hashtree.MergeCache),
	}
}

func (wc *workerCache) GetOrCreateCache(jobID string) (*hashtree.MergeCache, error) {
	if cache, ok := wc.caches[jobID]; ok {
		return cache, nil
	}
	cachePath := filepath.Join(wc.hashtreeStorage, jobID)
	if err := os.MkdirAll(cachePath, 0777); err != nil {
		return nil, err
	}

	newCache := hashtree.NewMergeCache(filepath.Join(wc.hashtreeStorage, jobID))
	wc.caches[jobID] = newCache
	return newCache, nil
}

func (wc *workerCache) RemoveCache(jobID string) error {
	if _, ok := wc.caches[jobID]; ok {
		if err := os.RemoveAll(filepath.Join(wc.hashtreeStorage, jobID)); err != nil {
			return err
		}
		delete(wc.caches, jobID)
	}
	return nil
}

func (wc *workerCache) GetCache(jobID string) *hashtree.MergeCache {
	return wc.caches[jobID]
}
