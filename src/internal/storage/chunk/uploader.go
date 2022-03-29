package chunk

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"golang.org/x/sync/semaphore"
)

const (
	taskParallelism  = 100
	chunkParallelism = 100
)

type uploadFunc = func(interface{}, []*DataRef) error

type Uploader struct {
	ctx       context.Context
	client    Client
	taskChain *TaskChain
	chunkSem  *semaphore.Weighted
	noUpload  bool
	cb        uploadFunc
}

func (s *Storage) NewUploader(ctx context.Context, name string, noUpload bool, cb uploadFunc) *Uploader {
	client := NewClient(s.store, s.db, s.tracker, NewRenewer(ctx, s.tracker, name, defaultChunkTTL))
	return &Uploader{
		ctx:       ctx,
		client:    client,
		taskChain: NewTaskChain(ctx, semaphore.NewWeighted(taskParallelism)),
		chunkSem:  semaphore.NewWeighted(chunkParallelism),
		noUpload:  noUpload,
		cb:        cb,
	}
}

// TODO: Need to think more about the context / error handling with the nested task chains.
func (u *Uploader) Upload(meta interface{}, r io.Reader) error {
	taskChain := NewTaskChain(u.ctx, u.chunkSem)
	var dataRefs []*DataRef
	if err := ComputeChunks(r, func(chunkBytes []byte) error {
		return taskChain.CreateTask(func(ctx context.Context) (func() error, error) {
			dataRef, err := upload(ctx, u.client, chunkBytes, nil, u.noUpload)
			if err != nil {
				return nil, err
			}
			return func() error {
				dataRefs = append(dataRefs, dataRef)
				return nil
			}, nil
		})
	}); err != nil {
		return err
	}
	return u.taskChain.CreateTask(func(ctx context.Context) (func() error, error) {
		if err := taskChain.Wait(); err != nil {
			return nil, err
		}
		return func() error {
			return u.cb(meta, dataRefs)
		}, nil
	})
}

func (u *Uploader) Copy(meta interface{}, dataRefs []*DataRef) error {
	return u.taskChain.CreateTask(func(_ context.Context) (func() error, error) {
		return func() error {
			return u.cb(meta, dataRefs)
		}, nil
	})
}

func upload(ctx context.Context, client Client, chunkBytes []byte, pointsTo []ID, noUpload bool) (*DataRef, error) {
	md := Metadata{
		Size:     len(chunkBytes),
		PointsTo: pointsTo,
	}
	createFunc := func(ctx context.Context, data []byte) (ID, error) {
		res, err := client.Create(ctx, md, data)
		return res, errors.EnsureStack(err)
	}
	if noUpload {
		createFunc = func(ctx context.Context, data []byte) (ID, error) {
			return Hash(data), nil
		}
	}
	ref, err := Create(ctx, CreateOptions{}, chunkBytes, createFunc)
	if err != nil {
		return nil, err
	}
	contentHash := Hash(chunkBytes)
	return &DataRef{
		Hash:      contentHash,
		Ref:       ref,
		SizeBytes: int64(len(chunkBytes)),
	}, nil
}

func (u *Uploader) Close() error {
	return u.taskChain.Wait()
}
