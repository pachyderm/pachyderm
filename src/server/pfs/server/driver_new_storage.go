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
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"golang.org/x/net/context"
)

const (
	storageTaskNamespace = "storage"
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

func (d *driver) startCommitNewStorageLayer(txnCtx *txnenv.TransactionContext, id string, parent *pfs.Commit, branch string, provenance []*pfs.CommitProvenance, description string) (*pfs.Commit, error) {
	commit, err := d.startCommit(txnCtx, id, parent, branch, provenance, description)
	if err != nil {
		return nil, err
	}
	d.fs = d.storage.New(context.Background(), commit.ID)
	return commit, nil
}

func (d *driver) finishCommitNewStorageLayer(txnCtx *txnenv.TransactionContext, commit *pfs.Commit, description string) (retErr error) {
	if err := d.checkIsAuthorizedInTransaction(txnCtx, commit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	commitInfo, err := d.resolveCommit(txnCtx.Stm, commit)
	if err != nil {
		return err
	}
	if commitInfo.Finished != nil {
		return pfsserver.ErrCommitFinished{commit}
	}
	if description != "" {
		commitInfo.Description = description
	}
	// Close in-memory file set (serializes in-memory part).
	if err := d.fs.Close(); err != nil {
		return err
	}
	defer func() {
		if retErr == nil {
			d.fs = nil
		}
	}()
	// Merge the commit with its parent.
	// (bryce) this should be where the level based compaction scheme is used.
	// right now we are eagerly compacting at every commit.
	fileSets := []string{commitInfo.Commit.ID}
	parentCommit := commitInfo.ParentCommit
	if parentCommit != nil {
		// (bryce) how to handle parent commit that is not closed?
		parentCommitInfo, err := d.resolveCommit(txnCtx.Stm, parentCommit)
		if err != nil {
			return err
		}
		fileSets = append(fileSets, path.Join(parentCommitInfo.Commit.ID, fileset.Compacted))
	}
	if err := d.merge(context.Background(), path.Join(commitInfo.Commit.ID, fileset.Compacted), fileSets); err != nil {
		return err
	}
	// (bryce) need size.
	commitInfo.SizeBytes = uint64(0)
	commitInfo.Finished = now()
	return d.writeFinishedCommit(txnCtx.Stm, commit, commitInfo)
}

// (bryce) should have a context going through here.
func (d *driver) putFilesNewStorageLayer(pachClient *client.APIClient, server *putFileServer) error {
	// (bryce) how to handle no commit started?
	// (bryce) this will not work for large files with the current implementation.
	for req, err := server.Recv(); err == nil; req, err = server.Recv() {
		if err := d.fs.WriteHeader(&tar.Header{Name: req.File.Path}); err != nil {
			return err
		}
		if _, err := io.Copy(d.fs, bytes.NewReader(req.Value)); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) getFileNewStorageLayer(pachClient *client.APIClient, file *pfs.File) (io.Reader, error) {
	// (bryce) path should be cleaned in option function
	fileSet := file.Commit.ID
	r := d.storage.NewReader(context.Background(), path.Join(fileSet, fileset.Compacted), index.WithPrefix(file.Path))
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

func (d *driver) merge(ctx context.Context, outputPath string, prefixes []string) error {
	// (bryce) need to add a resiliency measure for existing incomplete merge for the prefix (master crashed).
	// Setup task.
	task := &work.Task{Id: prefixes[0]}
	var err error
	task.Data, err = serializeMerge(&pfs.Merge{Prefixes: prefixes})
	if err != nil {
		return err
	}
	if err := d.storage.Shard(ctx, prefixes, func(pathRange *index.PathRange) error {
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
	collectFunc := func(_ context.Context, _ *work.Task) error { return nil }
	m := work.NewMaster(d.etcdClient, d.prefix, storageTaskNamespace)
	if err := m.Run(ctx, task, collectFunc); err != nil {
		return err
	}
	inputPrefix := path.Join(tmpPrefix, prefixes[0])
	return d.storage.Merge(ctx, outputPath, []string{inputPrefix})
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

// (bryce) it might potentially make sense to exit if an error occurs in this function
// because each pachd instance that errors here will lose its merge worker without an obvious
// notification for the user (outside of the log message).
func (d *driver) mergeWorker() {
	w := work.NewWorker(d.etcdClient, d.prefix, storageTaskNamespace)

	err := w.Run(context.Background(), func(ctx context.Context, task, subtask *work.Task) (retErr error) {
		merge, shard, err := deserialize(task.Data, subtask.Data)
		if err != nil {
			return err
		}
		outputPath := path.Join(tmpPrefix, task.Id, subtask.Id)
		pathRange := &index.PathRange{
			Lower: shard.Range.Lower,
			Upper: shard.Range.Upper,
		}
		return d.storage.Merge(ctx, outputPath, merge.Prefixes, index.WithRange(pathRange))
	})

	if err != nil {
		log.Printf("error in merge worker: %v", err)
	}
}
