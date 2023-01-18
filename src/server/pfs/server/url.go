package server

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	defaultURLTaskSize = 1000
)

var (
	batchSize = int64(1000000) // batchSize is overridden when testing. Default is 1 Mb
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
		defer func() {
			if err := resp.Body.Close(); retErr == nil {
				retErr = err
			}
		}()
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
		defer func() {
			if err := bucket.Close(); err != nil {
				retErr = multierror.Append(retErr, errors.EnsureStack(err))
			}
		}()
		r, err := bucket.NewReader(ctx, url.Object, nil)
		if err != nil {
			return 0, errors.EnsureStack(err)
		}
		defer func() {
			if err := r.Close(); err != nil {
				retErr = multierror.Append(retErr, errors.Wrapf(err, "error closing reader for bucket %s", url.Bucket))
			}
		}()
		return 0, uw.Put(ctx, dstPath, tag, true, r)
	}
}

func putFileURLRecursive(ctx context.Context, taskService task.Service, uw *fileset.UnorderedWriter, dst, tag string, src *pfs.AddFile_URLSource) error {
	inputChan := make(chan *types.Any)
	outputChan := make(chan *types.Any)
	eg, ctx := errgroup.WithContext(ctx)
	doer := taskService.NewDoer(URLTaskNamespace, uuid.NewWithoutDashes(), nil)
	// Create tasks.
	eg.Go(func() error {
		if err := coordinateTasks(ctx, src.URL, func(startOffset, endOffset int64, startPath, endPath string) error {
			input, err := serializePutFileURLTask(&PutFileURLTask{
				Dst:         dst,
				Datum:       tag,
				URL:         src.URL,
				StartOffset: startOffset,
				EndOffset:   endOffset,
				StartPath:   startPath,
				EndPath:     endPath,
			})
			if err != nil {
				return err
			}
			select {
			case inputChan <- input:
			case <-ctx.Done():
				return errors.EnsureStack(ctx.Err())
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
		defer close(outputChan)
		return task.DoOrdered(ctx, doer, inputChan, outputChan, 100)
	})
	// Deserialize output.
	eg.Go(func() error {
		for output := range outputChan {
			result, err := deserializePutFileURLTaskResult(output)
			if err != nil {
				return err
			}
			fsid, err := fileset.ParseID(result.Id)
			if err != nil {
				return err
			}
			if err := uw.AddFileSet(ctx, *fsid); err != nil {
				return errors.EnsureStack(err)
			}
		}
		return nil
	})
	return errors.EnsureStack(eg.Wait())
}

// coordinateTasks iterates through a list of objects and creates tasks of uniform byte size by sharding/batching objects.
// createTask is passed through as a callback to support nicer testing.
func coordinateTasks(ctx context.Context, URL string, createTask func(startOffset, endOffset int64, startPath, endPath string) error) (retErr error) {
	url, err := obj.ParseURL(URL)
	if err != nil {
		return errors.EnsureStack(err)
	}
	bucket, err := openBucket(ctx, url)
	if err != nil {
		return err
	}
	defer func() {
		if err := bucket.Close(); err != nil {
			retErr = multierror.Append(retErr, errors.EnsureStack(err))
		}
	}()
	startPath, endPath := "", ""
	list := bucket.List(&blob.ListOptions{Prefix: url.Object})
	var listObj *blob.ListObject
	listObj, err = list.Next(ctx)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return errors.Wrapf(err, "no objects were listed")
		}
	}
	startPath = listObj.Key
	remainingObjSize := listObj.Size
	bytesInBatch := int64(0)
	filePos := int64(0)
	for {
		if bytesInBatch+remainingObjSize == 0 && listObj.Size != 0 {
			startPath = ""
			endPath = ""
			filePos = 0
		}
		for bytesInBatch+remainingObjSize < batchSize {
			bytesInBatch += remainingObjSize
			listObj, err = list.Next(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if err := createTask(filePos, batchSize-bytesInBatch, startPath, endPath); err != nil {
						return errors.EnsureStack(err)
					}
					return nil
				}
				return errors.Wrapf(err, "error listing bucket %s", url.Bucket)
			}
			log.Debug(ctx, "object stats", zap.String("object", listObj.Key), zap.Int64("size", listObj.Size))
			remainingObjSize = listObj.Size
			if startPath == "" {
				startPath = listObj.Key
			}
			endPath = listObj.Key
		}
		remainingObjSize -= batchSize - bytesInBatch
		if err := createTask(filePos, batchSize-bytesInBatch, startPath, endPath); err != nil {
			return errors.EnsureStack(err)
		}
		startPath = listObj.Key
		endPath = listObj.Key
		filePos = batchSize - bytesInBatch
		bytesInBatch = 0
		for remainingObjSize >= batchSize {
			if err := createTask(filePos, filePos+batchSize, startPath, endPath); err != nil {
				return errors.EnsureStack(err)
			}
			filePos += batchSize
			remainingObjSize -= batchSize
		}
	}
}

// TODO: Parallelize and decide on appropriate config.
func (d *driver) getFileURL(ctx context.Context, taskService task.Service, URL string, file *pfs.File, basePathRange *pfs.PathRange) (int64, error) {
	if basePathRange == nil {
		basePathRange = &pfs.PathRange{}
	}
	inputChan := make(chan *types.Any)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		// TODO: Add cache?
		doer := taskService.NewDoer(URLTaskNamespace, uuid.NewWithoutDashes(), nil)
		err := doer.Do(ctx, inputChan, func(_ int64, output *types.Any, err error) error { return err })
		return errors.EnsureStack(err)
	})
	var bytesWritten int64
	eg.Go(func() error {
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
				return errors.EnsureStack(ctx.Err())
			}
			return nil
		}
		var count int64
		if err := src.Iterate(ctx, func(fi *pfs.FileInfo, file fileset.File) error {
			if fi.FileType != pfs.FileType_FILE {
				return nil
			}
			bytesWritten += int64(fi.SizeBytes)
			if count >= defaultURLTaskSize {
				pathRange.Upper = file.Index().Path
				if err := createTask(); err != nil {
					return err
				}
				pathRange = &pfs.PathRange{
					Lower: file.Index().Path,
				}
				count = 0
			}
			count++
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
	if err := eg.Wait(); err != nil {
		return 0, errors.EnsureStack(err)
	}
	return bytesWritten, nil
}
