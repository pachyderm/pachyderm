package server

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"go.uber.org/zap"
	"gocloud.dev/blob"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	URLTaskNamespace = "url"
)

func (d *driver) URLWorker(ctx context.Context) {
	ctx = auth.AsInternalUser(ctx, "pfs-url-worker")
	taskSource := d.env.TaskService.NewSource(URLTaskNamespace)
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		err := taskSource.Iterate(ctx, func(ctx context.Context, input *types.Any) (*types.Any, error) {
			switch {
			case types.Is(input, &PutFileURLTask{}):
				putFileURLTask, err := deserializePutFileURLTask(input)
				if err != nil {
					return nil, err
				}
				return d.processPutFileURLTask(ctx, putFileURLTask)
			case types.Is(input, &GetFileURLTask{}):
				getFileURLTask, err := deserializeGetFileURLTask(input)
				if err != nil {
					return nil, err
				}
				return d.processGetFileURLTask(ctx, getFileURLTask)
			default:
				return nil, errors.Errorf("unrecognized any type (%v) in URL worker", input.TypeUrl)
			}
		})
		return errors.EnsureStack(err)
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Error(ctx, "error in URL worker", zap.Error(err))
		return nil
	})
}

func (d *driver) processPutFileURLTask(ctx context.Context, task *PutFileURLTask) (_ *types.Any, retErr error) {
	url, err := obj.ParseURL(task.URL)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	bucket, err := openBucket(ctx, url)
	if err != nil {
		return nil, err
	}
	defer errors.Close(&retErr, bucket, "close bucket")
	prefix := strings.TrimPrefix(url.Object, "/")
	result := &PutFileURLTaskResult{}
	if err := log.LogStep(ctx, "putFileURLTask", func(ctx context.Context) error {
		return d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
			id, err := d.withUnorderedWriter(ctx, renewer, func(uw *fileset.UnorderedWriter) error {
				startOffset := task.StartOffset
				length := int64(-1) // -1 means to read until end of file.
				for i, path := range task.Paths {
					if i != 0 {
						startOffset = 0
					}
					if i == len(task.Paths)-1 && task.EndOffset != int64(-1) {
						length = task.EndOffset - startOffset
					}
					if err := func() error {
						r, err := bucket.NewRangeReader(ctx, path, startOffset, length, nil)
						if err != nil {
							return errors.EnsureStack(err)
						}
						defer errors.Close(&retErr, r, "close reader for bucket %s", url.Bucket)
						return uw.Put(ctx, filepath.Join(task.Dst, strings.TrimPrefix(path, prefix)), task.Datum, true, r)
					}(); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
			result.Id = id.HexString()
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return serializePutFileURLTaskResult(result)
}

func (d *driver) processGetFileURLTask(ctx context.Context, task *GetFileURLTask) (_ *types.Any, retErr error) {
	src, err := d.getFile(ctx, task.File, task.PathRange)
	if err != nil {
		return nil, err
	}
	url, err := obj.ParseURL(task.URL)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	bucket, err := openBucket(ctx, url)
	if err != nil {
		return nil, err
	}
	if url.Object != "" {
		bucket = blob.PrefixedBucket(bucket, strings.Trim(url.Object, "/")+"/")
	}
	defer errors.Close(&retErr, bucket, "close bucket")

	if err := log.LogStep(ctx, "getFileURLTask", func(ctx context.Context) error {
		err := src.Iterate(ctx, func(fi *pfs.FileInfo, file fileset.File) error {
			if fi.FileType != pfs.FileType_FILE {
				return nil
			}
			w, err := bucket.NewWriter(ctx, strings.TrimLeft(fi.File.Path, "/"), nil)
			if err != nil {
				return errors.EnsureStack(err)
			}
			defer errors.Close(&retErr, w, "close writer for bucket %s", url.Bucket)
			return errors.EnsureStack(file.Content(ctx, w))
		})
		return errors.EnsureStack(err)
	}); err != nil {
		return nil, err
	}
	return serializeGetFileURLTaskResult(&GetFileURLTaskResult{})
}

func serializePutFileURLTask(task *PutFileURLTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializePutFileURLTask(taskAny *types.Any) (*PutFileURLTask, error) {
	task := &PutFileURLTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializePutFileURLTaskResult(task *PutFileURLTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializePutFileURLTaskResult(taskAny *types.Any) (*PutFileURLTaskResult, error) {
	task := &PutFileURLTaskResult{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeGetFileURLTask(task *GetFileURLTask) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}

func deserializeGetFileURLTask(taskAny *types.Any) (*GetFileURLTask, error) {
	task := &GetFileURLTask{}
	if err := types.UnmarshalAny(taskAny, task); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return task, nil
}

func serializeGetFileURLTaskResult(task *GetFileURLTaskResult) (*types.Any, error) {
	data, err := proto.Marshal(task)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(task),
		Value:   data,
	}, nil
}
