package server

import (
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
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

const (
	// tmpPrefix is for temporary object paths that store compacted shards.
	tmpPrefix = "tmp"
)

func init() {
	// This is kind of weird, but I think it might be necessary.
	// Since we build our protos with gogoproto, we need to register
	// the types that can be embedded in an Any type with the gogoproto
	// proto package because building the protos just registers them with the
	// normal proto package.
	proto.RegisterType((*pfs.Compaction)(nil), "pfs.Compaction")
	proto.RegisterType((*pfs.Shard)(nil), "pfs.Shard")
}

func (d *driver) startCommitNewStorageLayer(txnCtx *txnenv.TransactionContext, id string, parent *pfs.Commit, branch string, provenance []*pfs.CommitProvenance, description string) (*pfs.Commit, error) {
	commit, err := d.startCommit(txnCtx, id, parent, branch, provenance, description)
	if err != nil {
		return nil, err
	}
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
	// Compact the commit changes into a diff file set.
	commitPath := path.Join(commit.Repo.Name, commit.ID)
	if err := d.compact(context.Background(), path.Join(commitPath, fileset.Diff), []string{commitPath}); err != nil {
		return err
	}
	// Compact the commit changes (diff file set) into the total changes in the commit's ancestry.
	var compactSpec *fileset.CompactSpec
	if commitInfo.ParentCommit == nil {
		compactSpec, err = d.storage.CompactSpec(context.Background(), commitPath)
	} else {
		// (bryce) how to handle parent commit that is not closed?
		parentCommitPath := path.Join(commitInfo.ParentCommit.Repo.Name, commitInfo.ParentCommit.ID)
		compactSpec, err = d.storage.CompactSpec(context.Background(), commitPath, parentCommitPath)
	}
	if err != nil {
		return err
	}
	if err := d.compact(context.Background(), compactSpec.Output, compactSpec.Input); err != nil {
		return err
	}
	// (bryce) need size.
	commitInfo.SizeBytes = uint64(0)
	commitInfo.Finished = now()
	return d.writeFinishedCommit(txnCtx.Stm, commit, commitInfo)
}

func (d *driver) putFilesNewStorageLayer(pachClient *client.APIClient, server *putFileServer) (retErr error) {
	// (bryce) how to handle no commit started?
	// (bryce) a failed put (crash/error) should result with any serialized sub file sets getting
	// cleaned up.
	// (bryce) piping is not ideal, but easy for getting integration started.
	// we will want to eventually just read raw bytes from the grpc stream.
	r, w := io.Pipe()
	var eg errgroup.Group
	var fs *fileset.FileSet
	fsChan := make(chan *fileset.FileSet)
	eg.Go(func() (retErr error) {
		fs = <-fsChan
		defer func() {
			if err := fs.Close(); err != nil {
				retErr = err
			}
		}()
		return fs.Put(r)
	})
	eg.Go(func() error {
		var fs *fileset.FileSet
		for req, err := server.Recv(); err == nil; req, err = server.Recv() {
			if fs == nil {
				repo := req.File.Commit.Repo.Name
				commit := req.File.Commit.ID
				subFileSetStr := fileset.SubFileSetStr(d.subFileSet)
				fsChan <- d.storage.New(pachClient.Ctx(), path.Join(repo, commit, subFileSetStr), subFileSetStr)
				// (bryce) subFileSet will need to be incremented through etcd eventually.
				d.subFileSet++
			}
			if _, err := w.Write(req.Value); err != nil {
				return err
			}
		}
		return w.Close()
	})
	return eg.Wait()
}

func (d *driver) getFileNewStorageLayer(pachClient *client.APIClient, file *pfs.File, w io.Writer) error {
	// (bryce) path should be cleaned in option function
	repo := file.Commit.Repo.Name
	commit := file.Commit.ID
	// (bryce) need exact match option for file path.
	mr, err := d.storage.NewMergeReader(context.Background(), []string{path.Join(repo, commit, fileset.Compacted)}, index.WithPrefix(file.Path))
	if err != nil {
		return err
	}
	return mr.Get(w)
}

func (d *driver) compact(ctx context.Context, outputPath string, prefixes []string) error {
	// (bryce) need some cleanup improvements, probably garbage collection.
	if err := d.storage.Delete(ctx, tmpPrefix); err != nil {
		return err
	}
	// (bryce) need to add a resiliency measure for existing incomplete compaction for the prefix (master crashed).
	// Setup task.
	task := &work.Task{Id: prefixes[0]}
	var err error
	task.Data, err = serializeCompaction(&pfs.Compaction{Prefixes: prefixes})
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
	m := work.NewMaster(d.etcdClient, d.prefix, func(_ context.Context, _ *work.Task) error { return nil })
	if err := m.Run(ctx, task); err != nil {
		return err
	}
	return d.storage.Compact(ctx, outputPath, []string{tmpPrefix})
}

func serializeCompaction(compaction *pfs.Compaction) (*types.Any, error) {
	serializedCompaction, err := proto.Marshal(compaction)
	if err != nil {
		return nil, err
	}
	return &types.Any{
		TypeUrl: "/" + proto.MessageName(compaction),
		Value:   serializedCompaction,
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

func deserialize(compactionAny, shardAny *types.Any) (*pfs.Compaction, *pfs.Shard, error) {
	compaction := &pfs.Compaction{}
	if err := types.UnmarshalAny(compactionAny, compaction); err != nil {
		return nil, nil, err
	}
	shard := &pfs.Shard{}
	if err := types.UnmarshalAny(shardAny, shard); err != nil {
		return nil, nil, err
	}
	return compaction, shard, nil
}

// (bryce) it might potentially make sense to exit if an error occurs in this function
// because each pachd instance that errors here will lose its compaction worker without an obvious
// notification for the user (outside of the log message).
func (d *driver) compactionWorker() {
	w := work.NewWorker(d.etcdClient, d.prefix, func(ctx context.Context, task, subtask *work.Task) (retErr error) {
		compaction, shard, err := deserialize(task.Data, subtask.Data)
		if err != nil {
			return err
		}
		outputPath := path.Join(tmpPrefix, task.Id, subtask.Id)
		pathRange := &index.PathRange{
			Lower: shard.Range.Lower,
			Upper: shard.Range.Upper,
		}
		return d.storage.Compact(ctx, outputPath, compaction.Prefixes, index.WithRange(pathRange))
	})
	if err := w.Run(context.Background()); err != nil {
		log.Printf("error in compaction worker: %v", err)
	}
}
