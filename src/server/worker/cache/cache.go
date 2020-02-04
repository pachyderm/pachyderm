package cache

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
)

// WorkerCache is an interface for managing hashtree MergeCaches for multiple
// concurrent jobs
type WorkerCache interface {
	DownloadHashtree(pachClient *client.APIClient, jobID string, tag string) error
	CacheHashtree(jobID string, tag string, r io.Reader) error
	ClearJob(jobID string)
	Get(jobID string, tag string, w io.Writer, filter hashtree.Filter) error
	Merge(jobID string, w *hashtree.Writer, parent io.Reader, filter hashtree.Filter) error
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

func (wc *workerCache) getOrCreateCache(jobID string) (*hashtree.MergeCache, error) {
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

func (wc *workerCache) ClearJob(jobID string) {
	delete(wc.caches, jobID)
}

func (wc *workerCache) CacheHashtree(jobID string, tag string, r io.Reader) error {
	cache, err := wc.getOrCreateCache(jobID)
	if err != nil {
		return err
	}
	return cache.Put(tag, r)
}

func (wc *workerCache) DownloadHashtree(pachClient *client.APIClient, jobID string, tag string) error {
	cache, err := wc.getOrCreateCache(jobID)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := pachClient.GetTag(tag, buf); err != nil {
		return err
	}
	return cache.Put(tag, buf)
	/* TODO: ensure a path exists for this when downloading hashtrees
	if a.pipelineInfo.EnableStats {
		buf.Reset()
		if err := d.PachClient().GetTag(tag+statsTagSuffix, buf); err != nil {
			// We are okay with not finding the stats hashtree.
			// This allows users to enable stats on a pipeline
			// with pre-existing jobs.
			return nil
		}
		return a.datumStatsCache.Put(datumIdx, buf)
	}
	*/
}

func (wc *workerCache) Get(jobID string, tag string, w io.Writer, filter hashtree.Filter) error {
	cache := wc.caches[jobID]
	if cache == nil {
		return fmt.Errorf("hashtree cache not found for job %v", jobID)
	}
	return cache.Get(tag, w, filter)
}

func (wc *workerCache) Merge(jobID string, w *hashtree.Writer, parent io.Reader, filter hashtree.Filter) error {
	cache := wc.caches[jobID]
	if cache == nil {
		return fmt.Errorf("hashtree cache not found for job %v", jobID)
	}

	return cache.Merge(w, parent, filter)
}
