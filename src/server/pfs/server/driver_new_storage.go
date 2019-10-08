package server

import (
	"bytes"
	"io"
	"log"
	"path"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"golang.org/x/net/context"
)

func init() {
	// This is kind of weird, but I think it might be necessary.
	// Since we build our protos with gogoproto, we need to register
	// the types that can be embedded in an Any type with the gogoproto
	// proto package because building the protos just registers them with the
	// normal proto package.
	proto.RegisterType((*pfs.WorkMeta)(nil), "pfs.WorkMeta")
	proto.RegisterType((*pfs.ShardMeta)(nil), "pfs.ShardMeta")
}

// (bryce) it might potentially make sense to exit if an error occurs in this function
// because each pachd instance that errors here will lose its merge worker without an obvious
// notification for the user (outside of the log message).
func (d *driver) mergeWorker() {
	w := work.NewWorker(d.etcdClient, d.prefix, func(ctx context.Context, work *work.Work, shard *work.Shard) (retErr error) {
		workMeta, shardMeta, err := deserializeMetas(work.Meta, shard.Meta)
		if err != nil {
			return err
		}
		outputPath := path.Join(tmpPrefix, workMeta.Prefix, shard.Id)
		pathRange := &index.PathRange{
			Lower: shardMeta.Range.Lower,
			Upper: shardMeta.Range.Upper,
		}
		return d.storage.Merge(ctx, outputPath, workMeta.Prefix, index.WithRange(pathRange))
	})
	if err := w.Run(context.Background()); err != nil {
		log.Printf("error in merge worker: %v", err)
	}
}

const (
	// (bryce) this is just a hacky way to start integrating put/get file without full integration
	// with the commit system.
	// The real integration will work off of commits, both for putting/getting files and merging.
	fileSet = "test"
)

var (
	// (bryce) need to expose as configuration.
	memThreshold int64 = 1024 * chunk.MB
)

// (bryce) should have a context going through here.
func (d *driver) putFilesNewStorageLayer(pachClient *client.APIClient, server *putFileServer) error {
	fs := d.storage.New(context.Background(), fileSet, fileset.WithMemThreshold(memThreshold))
	for req, err := server.Recv(); err == nil; req, err = server.Recv() {
		if err := fs.WriteHeader(&tar.Header{Name: req.File.Path}); err != nil {
			return err
		}
		if _, err := io.Copy(fs, bytes.NewReader(req.Value)); err != nil {
			return err
		}
	}
	if err := fs.Close(); err != nil {
		return err
	}
	return d.merge(context.Background(), fileSet)
}

func (d *driver) getFileNewStorageLayer(pachClient *client.APIClient, file *pfs.File) (io.Reader, error) {
	// (bryce) path should be cleaned in option function
	r := d.storage.NewReader(context.Background(), fileSet, index.WithPrefix(file.Path))
	hdr, err := r.Next()
	if err != nil {
		if err == io.EOF {
			return nil, pfsserver.ErrFileNotFound{file}
		}
		return nil, err
	}
	// (bryce) going to want an exact match option for storage layer.
	if hdr.Hdr.Name != file.Path {
		return nil, pfsserver.ErrFileNotFound{file}
	}
	return r, nil
}

func (d *driver) merge(ctx context.Context, prefix string) error {
	// (bryce) need to add a resiliency measure for existing incomplete merge for the prefix (master crashed).
	// Setup shards.
	workMeta, err := serializeWorkMeta(&pfs.WorkMeta{Prefix: prefix})
	if err != nil {
		return err
	}
	var shards []*work.Shard
	if err := d.storage.Shard(ctx, prefix, func(pathRange *index.PathRange) error {
		shardMeta, err := serializeShardMeta(&pfs.ShardMeta{
			Range: &pfs.PathRange{
				Lower: pathRange.Lower,
				Upper: pathRange.Upper,
			},
		})
		if err != nil {
			return err
		}
		shards = append(shards, &work.Shard{
			Id:   strconv.Itoa(len(shards)),
			Meta: shardMeta,
		})
		return nil
	}); err != nil {
		return err
	}
	m := work.NewMaster(d.etcdClient, d.prefix, func(_ context.Context, _ *work.Shard) error { return nil })
	if err := m.Run(ctx, &work.Work{
		Id:     prefix,
		Meta:   workMeta,
		Shards: shards,
	}); err != nil {
		return err
	}
	outputPath := path.Join(prefix, fileset.FullMergeSuffix)
	inputPrefix := path.Join(tmpPrefix, prefix)
	return d.storage.Merge(ctx, outputPath, inputPrefix)
}

func serializeWorkMeta(meta *pfs.WorkMeta) (*types.Any, error) {
	serializedMeta, err := proto.Marshal(meta)
	if err != nil {
		return nil, err
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(meta),
		Value:   serializedMeta,
	}, nil
}

func serializeShardMeta(meta *pfs.ShardMeta) (*types.Any, error) {
	serializedMeta, err := proto.Marshal(meta)
	if err != nil {
		return nil, err
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(meta),
		Value:   serializedMeta,
	}, nil
}

func deserializeMetas(work, shard *types.Any) (*pfs.WorkMeta, *pfs.ShardMeta, error) {
	workMeta := &pfs.WorkMeta{}
	if err := types.UnmarshalAny(work, workMeta); err != nil {
		return nil, nil, err
	}
	shardMeta := &pfs.ShardMeta{}
	if err := types.UnmarshalAny(shard, shardMeta); err != nil {
		return nil, nil, err
	}
	return workMeta, shardMeta, nil
}
