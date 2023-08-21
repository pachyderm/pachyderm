package server

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/docker/go-units"
	"gocloud.dev/blob"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	defaultConcurrency = 50
	maxConcurrency     = 1000
)

var (
	// these are overridden when testing.
	defaultNumObjectsThreshold = 1000
	defaultSizeThreshold       = int64(units.GB)
)

func putFileURL(ctx context.Context, taskService task.Service, uw *fileset.UnorderedWriter, dstPath, tag string, src *pfs.AddFile_URLSource) (n int64, retErr error) {
	url, err := url.Parse(src.URL)
	if err != nil {
		return 0, errors.EnsureStack(err)
	}
	switch url.Scheme {
	case "http", "https":
		client := &http.Client{
			Transport: promutil.InstrumentRoundTripper("putFileURL", http.DefaultTransport),
		}
		req, err := http.NewRequestWithContext(ctx, "GET", src.URL, nil)
		if err != nil {
			return 0, errors.EnsureStack(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return 0, errors.EnsureStack(err)
		} else if resp.StatusCode >= 400 {
			return 0, errors.Errorf("error retrieving content from %q: %s", src.URL, resp.Status)
		}
		defer errors.Close(&retErr, resp.Body, "close http body")
		return 0, uw.Put(ctx, dstPath, tag, true, resp.Body)
	default:
		if src.Recursive {
			return 0, putFileURLRecursive(ctx, taskService, uw, dstPath, tag, src)
		}
		url, err := obj.ParseURL(src.URL)
		if err != nil {
			return 0, errors.EnsureStack(err)
		}
		bucket, err := openBucket(ctx, url)
		if err != nil {
			return 0, err
		}
		defer errors.Close(&retErr, bucket, "close bucket")
		r, err := bucket.NewReader(ctx, url.Object, nil)
		if err != nil {
			return 0, errors.EnsureStack(err)
		}
		defer errors.Close(&retErr, r, "close reader for bucket %v", url.Bucket)
		return 0, uw.Put(ctx, dstPath, tag, true, r)
	}
}

type shardCallback func(paths []string, startOffset, endOffset int64) error

func putFileURLRecursive(ctx context.Context, taskService task.Service, uw *fileset.UnorderedWriter, dst, tag string, src *pfs.AddFile_URLSource) error {
	inputChan := make(chan *anypb.Any)
	eg, ctx := errgroup.WithContext(ctx)
	doer := taskService.NewDoer(URLTaskNamespace, uuid.NewWithoutDashes(), nil)
	// Create tasks.
	eg.Go(func() error {
		if err := shardObjects(ctx, src.URL, func(paths []string, startOffset, endOffset int64) error {
			input, err := serializePutFileURLTask(&PutFileURLTask{
				Dst:         dst,
				Datum:       tag,
				URL:         src.URL,
				Paths:       paths,
				StartOffset: startOffset,
				EndOffset:   endOffset,
			})
			if err != nil {
				return err
			}
			select {
			case inputChan <- input:
			case <-ctx.Done():
				return errors.EnsureStack(context.Cause(ctx))
			}
			return nil
		}); err != nil {
			return err
		}
		close(inputChan)
		return nil
	})
	// Order output from input channel.
	eg.Go(func() error {
		// TODO: Add cache?
		concurrency := src.Concurrency
		// We assume '0' here means a user did not set the concurrency in the request, so we should use our default.
		if src.Concurrency == 0 {
			concurrency = defaultConcurrency
		}
		if src.Concurrency > maxConcurrency {
			concurrency = maxConcurrency
		}
		return task.DoOrdered(ctx, doer, inputChan, int(concurrency), func(_ int64, output *anypb.Any, _ error) error {
			result, err := deserializePutFileURLTaskResult(output)
			if err != nil {
				return err
			}
			fsid, err := fileset.ParseID(result.Id)
			if err != nil {
				return err
			}
			return errors.EnsureStack(uw.AddFileSet(ctx, *fsid))
		})
	})
	return errors.EnsureStack(eg.Wait())
}

// shardObjects iterates through a list of objects and creates tasks by sharding small files
// into a single shard. Files larger than a shard size will be given their own dedicated shard.
func shardObjects(ctx context.Context, URL string, cb shardCallback) (retErr error) {
	url, err := obj.ParseURL(URL)
	if err != nil {
		return errors.EnsureStack(err)
	}
	bucket, err := openBucket(ctx, url)
	if err != nil {
		return err
	}
	defer errors.Close(&retErr, bucket, "close bucket")
	list := bucket.List(&blob.ListOptions{Prefix: url.Object})
	var paths []string
	var size int64
	for {
		listObj, err := list.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return errors.EnsureStack(err)
		}
		paths = append(paths, listObj.Key)
		size += listObj.Size
		if len(paths) >= defaultNumObjectsThreshold || size >= defaultSizeThreshold {
			if err := cb(paths, 0, -1); err != nil {
				return errors.EnsureStack(err)
			}
			paths = nil
			size = 0
		}
	}
	if len(paths) > 0 {
		return cb(paths, 0, -1)
	}
	return nil
}

func (d *driver) getFileURL(ctx context.Context, taskService task.Service, URL string, file *pfs.File, basePathRange *pfs.PathRange) (int64, error) {
	if basePathRange == nil {
		basePathRange = &pfs.PathRange{}
	}
	inputChan := make(chan *anypb.Any)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		// TODO: Add cache?
		doer := taskService.NewDoer(URLTaskNamespace, uuid.NewWithoutDashes(), nil)
		err := doer.Do(ctx, inputChan, func(_ int64, output *anypb.Any, err error) error { return err })
		return errors.EnsureStack(err)
	})
	var bytesWritten int64
	eg.Go(func() error {
		return d.storage.Filesets.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
			fsID, err := d.getFileSet(ctx, file.Commit)
			if err != nil {
				return err
			}
			if err := renewer.Add(ctx, *fsID); err != nil {
				return err
			}
			file := proto.Clone(file).(*pfs.File)
			file.Commit = client.NewRepo(pfs.DefaultProjectName, client.FileSetsRepoName).NewCommit("", fsID.HexString())
			src, err := d.getFile(ctx, file, basePathRange)
			if err != nil {
				return err
			}
			pathRange := &pfs.PathRange{
				Lower: basePathRange.Lower,
			}
			createTask := func() error {
				input, err := serializeGetFileURLTask(&GetFileURLTask{
					URL:       URL,
					File:      file,
					PathRange: pathRange,
				})
				if err != nil {
					return err
				}
				select {
				case inputChan <- input:
				case <-ctx.Done():
					return errors.EnsureStack(context.Cause(ctx))
				}
				return nil
			}
			var numObjects, size int64
			if err := src.Iterate(ctx, func(fi *pfs.FileInfo, file fileset.File) error {
				if fi.FileType != pfs.FileType_FILE {
					return nil
				}
				bytesWritten += fi.SizeBytes
				if numObjects >= int64(defaultNumObjectsThreshold) || size >= defaultSizeThreshold {
					pathRange.Upper = file.Index().Path
					if err := createTask(); err != nil {
						return err
					}
					pathRange = &pfs.PathRange{
						Lower: file.Index().Path,
					}
					numObjects, size = 0, 0
				}
				numObjects++
				size += fi.SizeBytes
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
			pathRange.Upper = basePathRange.Upper
			if err := createTask(); err != nil {
				return err
			}
			close(inputChan)
			return nil
		})
	})
	if err := eg.Wait(); err != nil {
		return 0, errors.EnsureStack(err)
	}
	return bytesWritten, nil
}
