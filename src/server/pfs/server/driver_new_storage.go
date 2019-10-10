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
	proto.RegisterType((*pfs.Merge)(nil), "pfs.Merge")
	proto.RegisterType((*pfs.Shard)(nil), "pfs.Shard")
}

// (bryce) it might potentially make sense to exit if an error occurs in this function
// because each pachd instance that errors here will lose its merge worker without an obvious
// notification for the user (outside of the log message).
func (d *driver) mergeWorker() {
	w := work.NewWorker(d.etcdClient, d.prefix, func(ctx context.Context, task, subtask *work.Task) (retErr error) {
		_, shard, err := deserialize(task.Data, subtask.Data)
		if err != nil {
			return err
		}
		outputPath := path.Join(tmpPrefix, task.Id, subtask.Id)
		pathRange := &index.PathRange{
			Lower: shard.Range.Lower,
			Upper: shard.Range.Upper,
		}
		return d.storage.Merge(ctx, outputPath, task.Id, index.WithRange(pathRange))
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
	// Setup task.
	task := &work.Task{Id: prefix}
	var err error
	task.Data, err = serializeMerge(&pfs.Merge{})
	if err != nil {
		return err
	}
	if err := d.storage.Shard(ctx, prefix, func(pathRange *index.PathRange) error {
		shard, err := serializeShard(&pfs.Shard{
			Range: &pfs.PathRange{
				Lower: pathRange.Lower,
				Upper: pathRange.Upper,
			},
		})
		if err != nil {
			return err
		}
		task.Subtasks = append(task.Subtasks, &work.Task{
			Id:   strconv.Itoa(len(task.Subtasks)),
			Data: shard,
		})
		return nil
	}); err != nil {
		return err
	}
	// Setup and run master.
	m := work.NewMaster(d.etcdClient, d.prefix, func(_ context.Context, _ *work.Task) error { return nil })
	if err := m.Run(ctx, task); err != nil {
		return err
	}
	outputPath := path.Join(prefix, fileset.FullMergeSuffix)
	inputPrefix := path.Join(tmpPrefix, prefix)
	return d.storage.Merge(ctx, outputPath, inputPrefix)
}

func serializeMerge(merge *pfs.Merge) (*types.Any, error) {
	serializedMerge, err := proto.Marshal(merge)
	if err != nil {
		return nil, err
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(merge),
		Value:   serializedMerge,
	}, nil
}

func serializeShard(shard *pfs.Shard) (*types.Any, error) {
	serializedShard, err := proto.Marshal(shard)
	if err != nil {
		return nil, err
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(shard),
		Value:   serializedShard,
	}, nil
}

func deserialize(mergeAny, shardAny *types.Any) (*pfs.Merge, *pfs.Shard, error) {
	merge := &pfs.Merge{}
	if err := types.UnmarshalAny(mergeAny, merge); err != nil {
		return nil, nil, err
	}
	shard := &pfs.Shard{}
	if err := types.UnmarshalAny(shardAny, shard); err != nil {
		return nil, nil, err
	}
	return merge, shard, nil
}
