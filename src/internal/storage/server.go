package storage

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"io"
	"regexp"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
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
	ChunkPrefix = "chunk/"
	defaultTTL  = client.DefaultTTL
	maxTTL      = 30 * time.Minute
)

type Env struct {
	DB *pachsql.DB
	// ObjectStore is a client from the obj package
	ObjectStore obj.Client

	// Bucket is an object storage bucket from the Go CDK packages.
	// If set, it takes priority over ObjectStore
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
func New(env Env) (*Server, error) {
	// Setup tracker
	tracker := track.NewPostgresTracker(env.DB)

	// chunk
	keyStore := chunk.NewPostgresKeyStore(env.DB)
	secret, err := getOrCreateKey(context.TODO(), keyStore, "default")
	if err != nil {
		return nil, err
	}

	var store kv.Store
	if env.Bucket != nil {
		store = kv.NewFromBucket(env.Bucket, maxKeySize, chunk.DefaultMaxChunkSize)
	} else {
		store = kv.NewFromObjectClient(env.ObjectStore, maxKeySize, chunk.DefaultMaxChunkSize)
	}
	store = wrapStore(&env.Config, store)
	store = kv.NewPrefixed(store, []byte(ChunkPrefix))
	chunkStorageOpts := makeChunkOptions(&env.Config)
	chunkStorageOpts = append(chunkStorageOpts, chunk.WithSecret(secret))
	chunkStorage := chunk.NewStorage(store, env.DB, tracker, chunkStorageOpts...)

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

// TODO: Copy file.
func (s *Server) CreateFileset(server storage.Fileset_CreateFilesetServer) error {
	ctx := server.Context()
	var id *fileset.ID
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
			}
		}
		id, err = uw.Close(ctx)
		return err
	}); err != nil {
		return err
	}
	return server.SendAndClose(&storage.CreateFilesetResponse{
		FilesetId: id.HexString(),
	})
}

// TODO: Add file filter error types.
func (s *Server) ReadFileset(request *storage.ReadFilesetRequest, server storage.Fileset_ReadFilesetServer) error {
	ctx := server.Context()
	id, err := fileset.ParseID(request.FilesetId)
	if err != nil {
		return err
	}
	fs, err := s.Filesets.Open(ctx, []fileset.ID{*id})
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
		w := &writer{
			server: server,
			path:   path,
		}
		bufW := bufio.NewWriterSize(w, grpcutil.MaxMsgPayloadSize)
		if err := f.Content(ctx, bufW); err != nil {
			return err
		}
		return errors.EnsureStack(bufW.Flush())
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

// TODO: We should be able to use this and potentially others directly in PFS.
func (s *Server) RenewFileset(ctx context.Context, request *storage.RenewFilesetRequest) (*emptypb.Empty, error) {
	id, err := fileset.ParseID(request.FilesetId)
	if err != nil {
		return nil, err
	}
	ttl := time.Duration(request.TtlSeconds) * time.Second
	if ttl < time.Second {
		return nil, errors.Errorf("ttl (%d) must be at least one second", ttl)
	}
	if ttl > maxTTL {
		return nil, errors.Errorf("ttl (%d) exceeds max ttl (%d)", ttl, maxTTL)
	}
	_, err = s.Filesets.SetTTL(ctx, *id, ttl)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ComposeFileset(ctx context.Context, request *storage.ComposeFilesetRequest) (*storage.ComposeFilesetResponse, error) {
	var ids []fileset.ID
	for _, id := range request.FilesetIds {
		id, err := fileset.ParseID(id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, *id)
	}
	ttl := time.Duration(request.TtlSeconds) * time.Second
	id, err := s.Filesets.Compose(ctx, ids, ttl)
	if err != nil {
		return nil, err
	}
	return &storage.ComposeFilesetResponse{
		FilesetId: id.HexString(),
	}, nil
}

func (s *Server) ShardFileset(ctx context.Context, request *storage.ShardFilesetRequest) (*storage.ShardFilesetResponse, error) {
	id, err := fileset.ParseID(request.FilesetId)
	if err != nil {
		return nil, err
	}
	fs, err := s.Filesets.Open(ctx, []fileset.ID{*id})
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
