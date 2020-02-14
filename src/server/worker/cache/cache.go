package cache

import (
	"os"
	"path/filepath"
	"sync"

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
	mutex sync.Mutex

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
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	if cache, ok := wc.caches[jobID]; ok {
		return cache, nil
	}

	newCache, err := hashtree.NewMergeCache(filepath.Join(wc.hashtreeStorage, jobID))
	if err != nil {
		return nil, err
	}

	wc.caches[jobID] = newCache
	return newCache, nil
}

func (wc *workerCache) RemoveCache(jobID string) error {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	if _, ok := wc.caches[jobID]; ok {
		if err := os.RemoveAll(filepath.Join(wc.hashtreeStorage, jobID)); err != nil {
			return err
		}
		delete(wc.caches, jobID)
	}
	return nil
}

func (wc *workerCache) GetCache(jobID string) *hashtree.MergeCache {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	return wc.caches[jobID]
}

func (wc *workerCache) Close() error {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()

	wc.caches = make(map[string]*hashtree.MergeCache)
	return os.RemoveAll(wc.hashtreeStorage)
}
