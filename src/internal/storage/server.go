package storage

import (
	"context"
	"crypto/rand"
	"database/sql"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/kv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"go.uber.org/zap"
	"gocloud.dev/blob"
)

const chunkPrefix = "chunk/"

type Env struct {
	DB *pachsql.DB
	// ObjectStore is a client from the obj package
	ObjectStore obj.Client

	// Bucket is an object storage bucket from the Go CDK packages.
	// If set, it takes priority over ObjectStore
	Bucket *blob.Bucket
}

// Server contains the storage layer servers.
// Eventually this will be it's own gRPC server.
type Server struct {
	Filesets *fileset.Storage
	Chunks   *chunk.Storage
	Tracker  track.Tracker
}

// New creates a new Server
func New(env Env, config pachconfig.StorageConfiguration) (*Server, error) {
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
	store = kv.NewMetered(store, "bucket")
	store = wrapStore(&config, store)
	store = kv.NewPrefixed(store, []byte(chunkPrefix))
	chunkStorageOpts := makeChunkOptions(&config)
	chunkStorageOpts = append(chunkStorageOpts, chunk.WithSecret(secret))
	chunkStorage := chunk.NewStorage(store, env.DB, tracker, chunkStorageOpts...)

	// fileset
	filesetStorage := fileset.NewStorage(fileset.NewPostgresStore(env.DB), tracker, chunkStorage, makeFilesetOptions(&config)...)

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
