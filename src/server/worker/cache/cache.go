package cache

import (
	"bytes"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
)

type WorkerCache interface {
	DownloadHashtree(jobID string, tag string) error
	CacheHashtree(jobID string, tag string, r io.Reader) error
	ClearJob(jobID string) error
}

type workerCache struct {
	// hashtreeStorage is the where we store on disk hashtrees
	hashtreeStorage string

	caches map[string]*hashtree.MergeCache
}

func (wc *WorkerCache) CacheHashtree(jobID string, tag string) (retErr error) {
	return nil
}

func (wc *WorkerCache) DownloadHashtree(jobID string, tag string) (retErr error) {
	buf := &bytes.Buffer{}
	if err := d.PachClient().GetTag(tag, buf); err != nil {
		return err
	}
	if err := a.datumCache.Put(tag, buf); err != nil {
		return err
	}
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
	return nil
}
