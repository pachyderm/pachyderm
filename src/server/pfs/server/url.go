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
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		// TODO: Add cache?
		doer := taskService.NewDoer(URLTaskNamespace, uuid.NewWithoutDashes(), nil)
		err := doer.Do(ctx, inputChan, func(_ int64, output *types.Any, err error) error {
			if err != nil {
				return err
			}
			result, err := deserializePutFileURLTaskResult(output)
			if err != nil {
				return err
			}
			fsid, err := fileset.ParseID(result.Id)
			if err != nil {
				return err
			}
			return uw.AddFileSet(ctx, *fsid)
		})
		return errors.EnsureStack(err)
	})
	eg.Go(func() (retErr error) {
		url, err := obj.ParseURL(src.URL)
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
		var paths []string
		createTask := func() error {
			input, err := serializePutFileURLTask(&PutFileURLTask{
				Dst:   dst,
				Datum: tag,
				URL:   src.URL,
				Paths: paths,
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
		list := bucket.List(&blob.ListOptions{Prefix: url.Object})
		var listObj *blob.ListObject
		for {
			listObj, err = list.Next(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return errors.Wrapf(err, "error listing bucket %s", url.Bucket)
			}
			log.Debug(ctx, "object stats", zap.String("object", listObj.Key), zap.Int64("size", listObj.Size))
			if len(paths) >= defaultURLTaskSize {
				if err := createTask(); err != nil {
					return errors.EnsureStack(err)
				}
				paths = nil
			}
			paths = append(paths, listObj.Key)
		}
		if len(paths) > 0 {
			if err := createTask(); err != nil {
				return err
			}
		}
		close(inputChan)
		return nil
	})
	return errors.EnsureStack(eg.Wait())
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
