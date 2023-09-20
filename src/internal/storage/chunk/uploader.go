package chunk

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/taskchain"
	"google.golang.org/protobuf/proto"

	"golang.org/x/sync/semaphore"
)

const (
	taskParallelism = 100
)

// UploadFunc is a function that provides the metadata for a task and the corresponding set of chunk references.
type UploadFunc = func(interface{}, []*DataRef) error

// Uploader uploads chunks.
// Each upload call creates at least one upload task with the provided metadata.
// Upload tasks are performed asynchronously, which is why the interface is callback based.
// Callbacks will be executed with respect to the order the upload tasks are created.
type Uploader struct {
	ctx       context.Context
	storage   *Storage
	client    Client
	taskChain *taskchain.TaskChain
	noUpload  bool
	cb        UploadFunc
}

func (s *Storage) NewUploader(ctx context.Context, name string, noUpload bool, cb UploadFunc) *Uploader {
	client := NewClient(s.store, s.db, s.tracker, NewRenewer(ctx, s.tracker, name, defaultChunkTTL), s.pool)
	return &Uploader{
		ctx:       ctx,
		storage:   s,
		client:    client,
		taskChain: taskchain.New(ctx, semaphore.NewWeighted(taskParallelism)),
		noUpload:  noUpload,
		cb:        cb,
	}
}

func (u *Uploader) Upload(meta interface{}, r io.Reader) error {
	var dataRefs []*DataRef
	if err := ComputeChunks(r, func(chunkBytes []byte) error {
		return u.taskChain.CreateTask(func(ctx context.Context) (func() error, error) {
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
	return u.taskChain.CreateTask(func(_ context.Context) (func() error, error) {
		return func() error {
			if len(dataRefs) > 0 {
				dataRefs[0].Ref.Edge = true
				dataRefs[len(dataRefs)-1].Ref.Edge = true
			}
			return u.cb(meta, dataRefs)
		}, nil
	})
}

// Copy performs an upload using a list of data references as the data source.
// Stable data references will be reused.
// Unstable data references will have their data downloaded and uploaded similar to a normal upload.
func (u *Uploader) Copy(meta interface{}, dataRefs []*DataRef) error {
	var stableDataRefs, nextDataRefs []*DataRef
	appendDataRefs := func(dataRefs []*DataRef) error {
		return u.taskChain.CreateTask(func(_ context.Context) (func() error, error) {
			return func() error {
				for _, dataRef := range dataRefs {
					stableDataRefs = append(stableDataRefs, proto.Clone(dataRef).(*DataRef))
				}
				return nil
			}, nil
		})
	}
	for len(dataRefs) > 0 {
		if !isStableDataRef(dataRefs[0]) {
			if err := appendDataRefs(nextDataRefs); err != nil {
				return err
			}
			nextDataRefs = nil
			var err error
			dataRefs, err = u.align(u.ctx, dataRefs, func(chunk []byte) error {
				return u.taskChain.CreateTask(func(ctx context.Context) (func() error, error) {
					dataRef, err := upload(ctx, u.client, chunk, nil, u.noUpload)
					if err != nil {
						return nil, err
					}
					return func() error {
						stableDataRefs = append(stableDataRefs, dataRef)
						return nil
					}, nil
				})
			})
			if err != nil {
				return err
			}
			continue
		}
		nextDataRefs = append(nextDataRefs, dataRefs[0])
		dataRefs = dataRefs[1:]
	}
	if err := appendDataRefs(nextDataRefs); err != nil {
		return err
	}
	return u.taskChain.CreateTask(func(_ context.Context) (func() error, error) {
		return func() error {
			if len(stableDataRefs) > 0 {
				stableDataRefs[0].Ref.Edge = true
				stableDataRefs[len(stableDataRefs)-1].Ref.Edge = true
			}
			return u.cb(meta, stableDataRefs)
		}, nil
	})
}

// A data reference is stable if it does not reference an edge chunk and references the full chunk.
func isStableDataRef(dataRef *DataRef) bool {
	return !dataRef.Ref.Edge && dataRef.OffsetBytes == 0 && dataRef.SizeBytes == dataRef.Ref.SizeBytes
}

// align performs a chunk boundary alignment on a list of data references.
// Alignment involves iterating through the passed in list of data references
// until a data reference is found that begins at a chunk boundary. Chunks
// encountered up to that point will be returned through the provided callback.
// The returned list of data references is the slice that begins at the first
// data reference found that begins at a chunk boundary.
func (u *Uploader) align(ctx context.Context, dataRefs []*DataRef, cb func([]byte) error) ([]*DataRef, error) {
	ctx, cancel := pctx.WithCancel(ctx)
	defer cancel()
	r := u.storage.NewReader(ctx, dataRefs, WithPrefetchLimit(3))
	if err := miscutil.WithPipe(func(w2 io.Writer) error {
		return r.Get(w2)
	}, func(r io.Reader) error {
		splitBytesLeft := dataRefs[0].SizeBytes
		return ComputeChunks(r, func(chunk []byte) error {
			chunkBytesLeft := int64(len(chunk))
			if err := cb(chunk); err != nil {
				return err
			}
			for chunkBytesLeft > splitBytesLeft {
				chunkBytesLeft -= splitBytesLeft
				dataRefs = dataRefs[1:]
				splitBytesLeft = dataRefs[0].SizeBytes
			}
			if chunkBytesLeft == splitBytesLeft {
				return errutil.ErrBreak
			}
			splitBytesLeft -= chunkBytesLeft
			return nil
		})
	}); err != nil && !errors.Is(err, errutil.ErrBreak) {
		return nil, err
	}
	return dataRefs[1:], nil
}

func (u *Uploader) CopyByReference(meta interface{}, dataRefs []*DataRef) error {
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
