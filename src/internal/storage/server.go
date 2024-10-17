package storage

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"io"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/taskchain"
	"golang.org/x/sync/semaphore"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pachyderm/pachyderm/v2/src/cdr"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsfile"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	ChunkPrefix     = "chunk/"
	defaultTTL      = client.DefaultTTL
	maxTTL          = 30 * time.Minute
	taskParallelism = 10
	cacheSize       = 128
)

type Env struct {
	DB     *pachsql.DB
	Bucket *blob.Bucket
	Config pachconfig.StorageConfiguration
}

// Server contains the storage layer servers.
type Server struct {
	storage.UnimplementedFilesetServer
	Filesets *fileset.Storage
	Chunks   *chunk.Storage
	Tracker  track.Tracker
}

// New creates a new Server
func New(ctx context.Context, env Env) (*Server, error) {
	// Setup tracker
	tracker := track.NewPostgresTracker(env.DB)

	// chunk
	keyStore := chunk.NewPostgresKeyStore(env.DB)
	secret, err := getOrCreateKey(ctx, keyStore, "default")
	if err != nil {
		return nil, err
	}

	var store kv.Store
	store = kv.NewFromBucket(env.Bucket, maxKeySize, chunk.DefaultMaxChunkSize)
	store = wrapStore(&env.Config, store)
	store = kv.NewPrefixed(store, []byte(ChunkPrefix))
	chunkStorageOpts := makeChunkOptions(&env.Config)
	chunkStorageOpts = append(chunkStorageOpts, chunk.WithSecret(secret))
	chunkStorage := chunk.NewStorage(store, env.Bucket, env.DB, tracker, chunkStorageOpts...)

	// fileset
	filesetStorage := fileset.NewStorage(fileset.NewPostgresStore(env.DB), tracker, chunkStorage, makeFilesetOptions(&env.Config)...)

	return &Server{
		Filesets: filesetStorage,
		Chunks:   chunkStorage,
		Tracker:  tracker,
	}, nil
}

func getOrCreateKey(ctx context.Context, keyStore chunk.KeyStore, name string) ([]byte, error) {
	secret, err := keyStore.Get(ctx, name)
	if !errors.Is(err, sql.ErrNoRows) {
		return secret, errors.EnsureStack(err)
	}
	secret = make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, errors.EnsureStack(err)
	}
	log.Info(ctx, "generated new secret", zap.String("name", name))
	if err := keyStore.Create(ctx, name, secret); err != nil {
		return nil, errors.EnsureStack(err)
	}
	res, err := keyStore.Get(ctx, name)
	return res, errors.EnsureStack(err)
}

func (s *Server) CreateFileset(server storage.Fileset_CreateFilesetServer) error {
	ctx := server.Context()
	var handle *fileset.Handle
	if err := s.Filesets.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		opts := []fileset.UnorderedWriterOption{fileset.WithRenewal(defaultTTL, renewer), fileset.WithValidator(ValidateFilename)}
		uw, err := s.Filesets.NewUnorderedWriter(ctx, opts...)
		if err != nil {
			return err
		}
		for {
			msg, err := server.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			switch mod := msg.Modification.(type) {
			case *storage.CreateFilesetRequest_AppendFile:
				if err := uw.Put(ctx, mod.AppendFile.Path, "", true, bytes.NewReader(mod.AppendFile.Data.Value)); err != nil {
					return err
				}
			case *storage.CreateFilesetRequest_DeleteFile:
				if err := uw.Delete(ctx, mod.DeleteFile.Path, ""); err != nil {
					return err
				}
			case *storage.CreateFilesetRequest_CopyFile:
				if err := s.copyFile(ctx, uw, mod.CopyFile); err != nil {
					return err
				}
			}
		}
		handle, err = uw.Close(ctx)
		return err
	}); err != nil {
		return err
	}
	return server.SendAndClose(&storage.CreateFilesetResponse{
		FilesetId: handle.HexString(),
	})
}

func (s *Server) copyFile(ctx context.Context, uw *fileset.UnorderedWriter, msg *storage.CopyFile) error {
	handle, err := fileset.ParseHandle(msg.FilesetId)
	if err != nil {
		return err
	}
	fs, err := s.Filesets.Open(ctx, []*fileset.Handle{handle})
	if err != nil {
		return err
	}
	srcPath := pfsfile.CleanPath(msg.Src)
	dstPath := srcPath
	if msg.Dst != "" {
		dstPath = pfsfile.CleanPath(msg.Dst)
	}
	fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
		return idx.Path == srcPath || strings.HasPrefix(idx.Path, fileset.Clean(srcPath, true))
	})
	fs = fileset.NewIndexMapper(fs, func(idx *index.Index) *index.Index {
		idx = protoutil.Clone(idx)
		relPath, err := filepath.Rel(srcPath, idx.Path)
		if err != nil {
			panic(err)
		}
		idx.Path = path.Join(dstPath, relPath)
		return idx
	})
	return uw.Copy(ctx, fs, "", false, index.WithPrefix(srcPath))
}

// ReadFileset is not properly documented.
//
//   - TODO: Add file filter error types.
//   - TODO: document.
func (s *Server) ReadFileset(request *storage.ReadFilesetRequest, server storage.Fileset_ReadFilesetServer) error {
	ctx := server.Context()
	return s.readFileset(ctx, request, func(f fileset.File) error {
		w := &writer{
			server: server,
			path:   f.Index().Path,
		}
		bufW := bufio.NewWriterSize(w, grpcutil.MaxMsgPayloadSize)
		if err := f.Content(ctx, bufW); err != nil {
			return err
		}
		return errors.EnsureStack(bufW.Flush())
	})
}

// TODO: Add file filter error types.
func (s *Server) readFileset(ctx context.Context, request *storage.ReadFilesetRequest, cb func(fileset.File) error) error {
	handle, err := fileset.ParseHandle(request.FilesetId)
	if err != nil {
		return err
	}
	fs, err := s.Filesets.Open(ctx, []*fileset.Handle{handle})
	if err != nil {
		return err
	}
	// Compute the intersection of the path range filters.
	pathRange := &index.PathRange{}
	for _, f := range request.Filters {
		switch f := f.Filter.(type) {
		case *storage.FileFilter_PathRange:
			// Return immediately if the path ranges are disjoint.
			if pathRange.Lower != "" && f.PathRange.Upper <= pathRange.Lower ||
				pathRange.Upper != "" && f.PathRange.Lower >= pathRange.Upper {
				return nil
			}
			if f.PathRange.Lower > pathRange.Lower {
				pathRange.Lower = f.PathRange.Lower
			}
			if pathRange.Upper == "" || f.PathRange.Upper != "" && f.PathRange.Upper < pathRange.Upper {
				pathRange.Upper = f.PathRange.Upper
			}
		}
	}
	// Compile regular expressions.
	var regexes []*regexp.Regexp
	for _, f := range request.Filters {
		switch f := f.Filter.(type) {
		case *storage.FileFilter_PathRegex:
			regex, err := regexp.Compile(f.PathRegex)
			if err != nil {
				return errors.EnsureStack(err)
			}
			regexes = append(regexes, regex)
		}
	}
	return fs.Iterate(ctx, func(f fileset.File) error {
		path := f.Index().Path
		for _, r := range regexes {
			if !r.MatchString(path) {
				return nil
			}
		}
		return cb(f)
	}, index.WithRange(pathRange))
}

type writer struct {
	server storage.Fileset_ReadFilesetServer
	path   string
}

func (w *writer) Write(data []byte) (int, error) {
	response := &storage.ReadFilesetResponse{
		Path: w.path,
		Data: wrapperspb.Bytes(data),
	}
	if err := w.server.Send(response); err != nil {
		return 0, err
	}
	return len(data), nil
}

func (s *Server) ReadFilesetCDR(request *storage.ReadFilesetRequest, server storage.Fileset_ReadFilesetCDRServer) error {
	ctx := server.Context()
	cache, err := lru.New[string, string](cacheSize)
	if err != nil {
		return err
	}
	taskChain := taskchain.New(ctx, semaphore.NewWeighted(int64(taskParallelism)))
	if err := s.readFileset(ctx, request, func(f fileset.File) error {
		if len(f.Index().File.DataRefs) == 0 {
			return server.Send(&storage.ReadFilesetCDRResponse{
				Path: f.Index().Path,
				Ref:  &cdr.Ref{Body: &cdr.Ref_Concat{}},
			})
		}
		var refs []*cdr.Ref
		for i, dataRef := range f.Index().File.DataRefs {
			i, dataRef := i, dataRef
			if err := taskChain.CreateTask(func(ctx context.Context) (func() error, error) {
				ref, err := s.Chunks.CDRFromDataRef(ctx, dataRef, cache)
				if err != nil {
					return nil, err
				}
				return func() error {
					refs = append(refs, ref)
					if i == len(f.Index().File.DataRefs)-1 {
						// TODO: Move to cdr package?
						ref := chunk.CreateConcatRef(refs)
						return server.Send(&storage.ReadFilesetCDRResponse{
							Path: f.Index().Path,
							Ref:  ref,
						})
					}
					return nil
				}, nil
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return taskChain.Wait()
}

// RenewFileset is not properly documented.
//
//   - TODO: We should be able to use this and potentially others directly in PFS.
//   - TODO: Document.
func (s *Server) RenewFileset(ctx context.Context, request *storage.RenewFilesetRequest) (*emptypb.Empty, error) {
	handle, err := fileset.ParseHandle(request.FilesetId)
	if err != nil {
		return nil, err
	}
	switch {
	case request.TtlSeconds < 1:
		return nil, errors.Errorf("ttl (%d) must be at least one second", request.TtlSeconds)
	case request.TtlSeconds > int64(maxTTL/time.Second):
		return nil, errors.Errorf("ttl (%ds) exceeds max ttl (%ds)", request.TtlSeconds, maxTTL/time.Second)
	}
	// NOTE: no need to explicitly check overflow because of the bounds checks above.
	ttl := time.Duration(request.TtlSeconds) * time.Second
	_, err = s.Filesets.SetTTL(ctx, handle, ttl)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ComposeFileset(ctx context.Context, request *storage.ComposeFilesetRequest) (*storage.ComposeFilesetResponse, error) {
	var handles []*fileset.Handle
	for _, handle := range request.FilesetIds {
		handle, err := fileset.ParseHandle(handle)
		if err != nil {
			return nil, err
		}
		handles = append(handles, handle)
	}
	ttl := time.Duration(request.TtlSeconds) * time.Second
	handle, err := s.Filesets.Compose(ctx, handles, ttl)
	if err != nil {
		return nil, err
	}
	return &storage.ComposeFilesetResponse{
		FilesetId: handle.HexString(),
	}, nil
}

func (s *Server) ShardFileset(ctx context.Context, request *storage.ShardFilesetRequest) (*storage.ShardFilesetResponse, error) {
	handle, err := fileset.ParseHandle(request.FilesetId)
	if err != nil {
		return nil, err
	}
	fs, err := s.Filesets.Open(ctx, []*fileset.Handle{handle})
	if err != nil {
		return nil, err
	}
	shardConfig := s.Filesets.ShardConfig()
	if request.NumFiles > 0 {
		shardConfig.NumFiles = request.NumFiles
	}
	if request.SizeBytes > 0 {
		shardConfig.SizeBytes = request.SizeBytes
	}
	shards, err := fs.Shards(ctx, index.WithShardConfig(shardConfig))
	if err != nil {
		return nil, err
	}
	var pathRanges []*storage.PathRange
	for _, shard := range shards {
		pathRanges = append(pathRanges, &storage.PathRange{
			Lower: shard.Lower,
			Upper: shard.Upper,
		})
	}
	return &storage.ShardFilesetResponse{
		Shards: pathRanges,
	}, nil
}
