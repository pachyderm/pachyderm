package server

import (
	"io"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/gc"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"golang.org/x/net/context"
)

const (
	// tmpPrefix is for temporary storage paths.
	// TODO Temporary prefix cleanup needs some work.
	// Paths should get cleaned up in the background.
	tmpPrefix            = "tmp"
	storageTaskNamespace = "storage"
)

type driverV2 struct {
	*driver
}

// newDriver is used to create a new Driver instance
func newDriverV2(
	env *serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	etcdPrefix string,
	treeCache *hashtree.Cache,
	storageRoot string,
	memoryRequest int64,
) (*driverV2, error) {
	d1, err := newDriver(env, txnEnv, etcdPrefix, treeCache, storageRoot, memoryRequest)
	if err != nil {
		return nil, err
	}
	d2 := &driverV2{driver: d1}
	objClient, err := NewObjClient(env.Configuration)
	if err != nil {
		return nil, err
	}
	// (bryce) local db for testing.
	db, err := gc.NewLocalDB()
	if err != nil {
		return nil, err
	}
	gcClient, err := gc.NewClient(db)
	if err != nil {
		return nil, err
	}
	chunkStorageOpts := append([]chunk.StorageOption{chunk.WithGarbageCollection(gcClient)}, chunk.ServiceEnvToOptions(env)...)
	d2.storage = fileset.NewStorage(objClient, chunk.NewStorage(objClient, chunkStorageOpts...), fileset.ServiceEnvToOptions(env)...)
	d2.compactionQueue, err = work.NewTaskQueue(context.Background(), d2.etcdClient, d2.prefix, storageTaskNamespace)
	if err != nil {
		return nil, err
	}
	go d2.master(env, objClient, db)
	go d2.compactionWorker()
	return d2, nil
}

func (d *driverV2) finishCommitV2(txnCtx *txnenv.TransactionContext, commit *pfs.Commit, description string) error {
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
	commitPath := path.Join(commit.Repo.Name, commit.ID)
	// Clean up temporary filesets leftover from failed operations.
	if err := d.storage.Delete(txnCtx.Client.Ctx(), path.Join(tmpPrefix, commitPath)); err != nil {
		return err
	}
	// Run compaction task.
	return d.compactionQueue.RunTaskBlock(txnCtx.Client.Ctx(), func(m *work.Master) error {
		if err := backoff.Retry(func() error {
			if err := d.storage.Delete(context.Background(), path.Join(commitPath, fileset.Diff)); err != nil {
				return err
			}
			return d.storage.Delete(context.Background(), path.Join(commitPath, fileset.Compacted))
		}, backoff.NewExponentialBackOff()); err != nil {
			return err
		}
		// Compact the commit changes into a diff file set.
		if _, err := d.compact(m, path.Join(commitPath, fileset.Diff), []string{commitPath}); err != nil {
			return err
		}
		// Compact the commit changes (diff file set) into the total changes in the commit's ancestry.
		var compactSpec *fileset.CompactSpec
		if commitInfo.ParentCommit == nil {
			compactSpec, err = d.storage.CompactSpec(m.Ctx(), commitPath)
		} else {
			parentCommitPath := path.Join(commitInfo.ParentCommit.Repo.Name, commitInfo.ParentCommit.ID)
			compactSpec, err = d.storage.CompactSpec(m.Ctx(), commitPath, parentCommitPath)
		}
		if err != nil {
			return err
		}
		compactRes, err := d.compact(m, compactSpec.Output, compactSpec.Input)
		if err != nil {
			return err
		}
		commitInfo.SizeBytes = uint64(compactRes.OutputSize)
		commitInfo.Finished = types.TimestampNow()
		return d.writeFinishedCommit(txnCtx.Stm, commit, commitInfo)
	})
}

// TODO Integrate delete commit.
//func (d *driver) deleteCommitV2(txnCtx *txnenv.TransactionContext, commit *pfs.Commit) error {
//	return d.storage.Delete(txnCtx.Client.Ctx(), path.Join(commit.Repo.Name, commit.ID))
//}

// TODO Need commit validation and handling of branch names.
func (d *driverV2) withFileSet(ctx context.Context, repo, commit string, f func(*fileset.FileSet) error) (retErr error) {
	// TODO subFileSet will need to be incremented through postgres or etcd.
	d.mu.Lock()
	subFileSetStr := fileset.SubFileSetStr(d.subFileSet)
	subFileSetPath := path.Join(repo, commit, subFileSetStr)
	fs, err := d.storage.New(ctx, path.Join(tmpPrefix, subFileSetPath), subFileSetStr)
	if err != nil {
		return err
	}
	d.subFileSet++
	d.mu.Unlock()
	defer func() {
		if err := d.storage.Delete(ctx, path.Join(tmpPrefix, subFileSetPath)); retErr == nil {
			retErr = err
		}
	}()
	if err := f(fs); err != nil {
		return err
	}
	if err := fs.Close(); err != nil {
		return err
	}
	return d.compactionQueue.RunTaskBlock(ctx, func(m *work.Master) error {
		_, err := d.compact(m, subFileSetPath, []string{path.Join(tmpPrefix, subFileSetPath)})
		return err
	})
}

func (d *driverV2) getTar(ctx context.Context, commit *pfs.Commit, glob string, w io.Writer) error {
	indexOpt, mf, err := parseGlob(cleanPath(glob))
	if err != nil {
		return err
	}
	s := d.storage.NewSource(ctx, compactedCommitPath(commit), indexOpt)
	filter := fileset.NewIndexFilter(s, func(idx *index.Index) bool {
		return mf(idx.Path)
	})
	// TODO: remove absolute paths on the way out?
	// nonAbsolute := &fileset.HeaderMapper{
	// 	R: filter,
	// 	F: func(th *tar.Header) *tar.Header {
	// 		th.Name = "." + th.Name
	// 		return th
	// 	},
	// }
	return fileset.WriteTarStream(ctx, w, filter)
}

func (d *driverV2) listFileV2(pachClient *client.APIClient, file *pfs.File, full bool, history int64, cb func(*pfs.FileInfoV2) error) error {
	ctx := pachClient.Ctx()
	if err := validateFile(file); err != nil {
		return err
	}
	if err := d.checkIsAuthorized(pachClient, file.Commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	name := cleanPath(file.Path)
	s := NewSource(file.Commit, true, func() fileset.FileSource {
		x := d.storage.NewSource(ctx, compactedCommitPath(file.Commit), index.WithPrefix(name))
		x = fileset.NewIndexResolver(x)
		x = fileset.NewIndexFilter(x, func(idx *index.Index) bool {
			if idx.Path == name {
				return true
			}
			if idx.Path == name+"/" {
				return false
			}
			return strings.HasPrefix(idx.Path, name)
		})
		return x
	})
	return s.Iterate(ctx, func(fi *pfs.FileInfoV2, _ fileset.File) error {
		if pathIsChild(name, cleanPath(fi.File.Path)) {
			return cb(fi)
		}
		return nil
	})
}

// TODO Need to figure out path cleaning.
func (d *driverV2) getTarConditional(ctx context.Context, repo, commit, glob string, f func(*FileReader) error) error {
	compactedPaths := []string{path.Join(repo, commit, fileset.Compacted)}
	prefix := globLiteralPrefix(glob)
	mr, err := d.storage.NewMergeReader(ctx, compactedPaths, index.WithPrefix(prefix))
	if err != nil {
		return err
	}
	mf, err := matchFunc(glob)
	if err != nil {
		return err
	}
	var fr *FileReader
	nextFileReader := func(idx *index.Index) error {
		fmr, err := mr.Next()
		if err != nil {
			return err
		}
		if !mf(idx.Path) {
			return nil
		}
		fr = newFileReader(client.NewFile(repo, commit, idx.Path), idx, fmr, mr)
		return nil
	}
	if err := d.storage.ResolveIndexes(ctx, compactedPaths, func(idx *index.Index) error {
		// Ignore index entries for deleted files.
		if len(idx.DataOp.DataRefs) == 0 {
			return nil
		}
		if fr == nil {
			return nextFileReader(idx)
		}
		dir := path.Dir(idx.Path)
		if dir == fr.file.Path {
			fr.updateFileInfo(idx)
			return nil
		}
		if err := f(fr); err != nil {
			return err
		}
		if err := fr.drain(); err != nil {
			return err
		}
		fr = nil
		return nextFileReader(idx)

	}, index.WithPrefix(prefix)); err != nil {
		return err
	}
	if fr != nil {
		return f(fr)
	}
	return nil
}

type compactStats struct {
	OutputSize int64
}

func (d *driverV2) compact(master *work.Master, outputPath string, inputPrefixes []string) (*compactStats, error) {
	ctx := master.Ctx()
	// resolve prefixes into paths
	inputPaths := []string{}
	for _, inputPrefix := range inputPrefixes {
		if err := d.storage.WalkFileSet(ctx, inputPrefix, func(inputPath string) error {
			inputPaths = append(inputPaths, inputPath)
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return d.compactIter(ctx, compactSpec{
		master:     master,
		inputPaths: inputPaths,
		outputPath: outputPath,
		maxFanIn:   d.env.StorageCompactionMaxFanIn,
	})
}

type compactSpec struct {
	master     *work.Master
	outputPath string
	inputPaths []string
	maxFanIn   int
}

// compactIter is one level of compaction.  It will only perform compaction
// if len(inputPaths) <= params.maxFanIn otherwise it will split inputPaths recursively.
func (d *driverV2) compactIter(ctx context.Context, params compactSpec) (_ *compactStats, retErr error) {
	if len(params.inputPaths) <= params.maxFanIn {
		return d.shardedCompact(ctx, params.master, params.outputPath, params.inputPaths)
	}
	scratch := path.Join(tmpPrefix, uuid.NewWithoutDashes())
	defer func() {
		if err := d.storage.Delete(ctx, scratch); retErr == nil {
			retErr = err
		}
	}()
	childOutputPaths := []string{}
	childSize := len(params.inputPaths) / params.maxFanIn
	if len(params.inputPaths)%params.maxFanIn != 0 {
		childSize++
	}
	// TODO: use an errgroup to make the recursion concurrecnt.
	// this requires changing the master to allow multiple calls to RunSubtasks
	// don't forget to pass the errgroups childCtx to compactIter instead of ctx.
	for i := 0; i < params.maxFanIn; i++ {
		start := i * childSize
		end := (i + 1) * childSize
		if end > len(params.inputPaths) {
			end = len(params.inputPaths)
		}
		childOutputPath := path.Join(scratch, strconv.Itoa(i))
		childOutputPaths = append(childOutputPaths, childOutputPath)
		if _, err := d.compactIter(ctx, compactSpec{
			master:     params.master,
			inputPaths: params.inputPaths[start:end],
			outputPath: childOutputPath,
			maxFanIn:   params.maxFanIn,
		}); err != nil {
			return nil, err
		}
	}
	return d.shardedCompact(ctx, params.master, params.outputPath, childOutputPaths)
}

// shardedCompact generates shards for the fileset(s) in inputPaths,
// gives those shards to workers, and waits for them to complete.
// Fan in is bound by len(inputPaths), concatenating shards have
// fan in of one because they are concatenated sequentially.
func (d *driverV2) shardedCompact(ctx context.Context, master *work.Master, outputPath string, inputPaths []string) (_ *compactStats, retErr error) {
	scratch := path.Join(tmpPrefix, uuid.NewWithoutDashes())
	defer func() {
		if err := d.storage.Delete(ctx, scratch); retErr == nil {
			retErr = err
		}
	}()
	compaction := &pfs.Compaction{InputPrefixes: inputPaths}
	var subtasks []*work.Task
	var shardOutputs []string
	if err := d.storage.Shard(ctx, inputPaths, func(pathRange *index.PathRange) error {
		shardOutputPath := path.Join(scratch, strconv.Itoa(len(subtasks)))
		shard, err := serializeShard(&pfs.Shard{
			Compaction: compaction,
			Range: &pfs.PathRange{
				Lower: pathRange.Lower,
				Upper: pathRange.Upper,
			},
			OutputPath: shardOutputPath,
		})
		if err != nil {
			return err
		}
		subtasks = append(subtasks, &work.Task{Data: shard})
		shardOutputs = append(shardOutputs, shardOutputPath)
		return nil
	}); err != nil {
		return nil, err
	}
	if err := master.RunSubtasks(subtasks, func(_ context.Context, taskInfo *work.TaskInfo) error {
		if taskInfo.State == work.State_FAILURE {
			return errors.Errorf(taskInfo.Reason)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return d.concatFileSets(ctx, outputPath, shardOutputs)
}

// concatFileSets concatenates the filesets in inputPaths and writes the result to outputPath
// TODO: move this to the fileset package, and error if the entries are not sorted.
func (d *driverV2) concatFileSets(ctx context.Context, outputPath string, inputPaths []string) (*compactStats, error) {
	var size int64
	fsw := d.storage.NewWriter(ctx, outputPath, fileset.WithIndexCallback(func(idx *index.Index) error {
		size += idx.SizeBytes
		return nil
	}))
	for _, inputPath := range inputPaths {
		fsr := d.storage.NewReader(ctx, inputPath)
		if err := fileset.CopyFiles(fsw, fsr); err != nil {
			return nil, err
		}
	}
	if err := fsw.Close(); err != nil {
		return nil, err
	}
	return &compactStats{OutputSize: size}, nil
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

func deserializeShard(shardAny *types.Any) (*pfs.Shard, error) {
	shard := &pfs.Shard{}
	if err := types.UnmarshalAny(shardAny, shard); err != nil {
		return nil, err
	}
	return shard, nil
}

func (d *driverV2) compactionWorker() {
	ctx := context.Background()
	w := work.NewWorker(d.etcdClient, d.prefix, storageTaskNamespace)
	err := backoff.RetryNotify(func() error {
		return w.Run(ctx, func(ctx context.Context, subtask *work.Task) error {
			shard, err := deserializeShard(subtask.Data)
			if err != nil {
				return err
			}
			pathRange := &index.PathRange{
				Lower: shard.Range.Lower,
				Upper: shard.Range.Upper,
			}
			_, err = d.storage.Compact(ctx, shard.OutputPath, shard.Compaction.InputPrefixes, index.WithRange(pathRange))
			return err
		})
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Printf("error in compaction worker: %v", err)
		return nil
	})
	// Never ending backoff should prevent us from getting here.
	panic(err)
}

func (d *driverV2) globFileV2(pachClient *client.APIClient, commit *pfs.Commit, glob string, cb func(*pfs.FileInfoV2) error) (retErr error) {
	// Validate arguments
	if commit == nil {
		return errors.New("commit cannot be nil")
	}
	if commit.Repo == nil {
		return errors.New("commit repo cannot be nil")
	}
	if err := d.checkIsAuthorized(pachClient, commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	ctx := pachClient.Ctx()
	indexOpt, mf, err := parseGlob(cleanPath(glob))
	if err != nil {
		return err
	}
	s := NewSource(commit, true, func() fileset.FileSource {
		x := d.storage.NewSource(ctx, compactedCommitPath(commit), indexOpt)
		x = fileset.NewIndexResolver(x)
		return fileset.NewIndexFilter(x, func(idx *index.Index) bool {
			return mf(cleanPath(idx.Path))
		})
	})
	return s.Iterate(ctx, func(fi *pfs.FileInfoV2, f fileset.File) error {
		return cb(fi)
	})
}

func (d *driverV2) diffFileV2(pachClient *client.APIClient, oldFile, newFile *pfs.File, cb func(oldFi, newFi *pfs.FileInfoV2) error) error {
	// TODO: move validation to the Validating API Server
	// Validation
	if newFile == nil {
		return errors.New("file cannot be nil")
	}
	if newFile.Commit == nil {
		return errors.New("file commit cannot be nil")
	}
	if newFile.Commit.Repo == nil {
		return errors.New("file commit repo cannot be nil")
	}
	// Do READER authorization check for both newFile and oldFile
	if oldFile != nil && oldFile.Commit != nil {
		if err := d.checkIsAuthorized(pachClient, oldFile.Commit.Repo, auth.Scope_READER); err != nil {
			return err
		}
	}
	if newFile != nil && newFile.Commit != nil {
		if err := d.checkIsAuthorized(pachClient, newFile.Commit.Repo, auth.Scope_READER); err != nil {
			return err
		}
	}
	newCommitInfo, err := d.inspectCommit(pachClient, newFile.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if oldFile == nil {
		oldFile = &pfs.File{}
		// ParentCommit may be nil, that's fine because getTreeForCommit
		// handles nil
		oldFile.Commit = newCommitInfo.ParentCommit
		oldFile.Path = newFile.Path
	}
	ctx := pachClient.Ctx()
	oldCommit := oldFile.Commit
	newCommit := newFile.Commit
	oldName := cleanPath(oldFile.Path)
	newName := cleanPath(newFile.Path)
	old := NewSource(oldCommit, true, func() fileset.FileSource {
		x := d.storage.NewSource(ctx, compactedCommitPath(oldCommit), index.WithPrefix(oldName))
		x = fileset.NewIndexResolver(x)
		x = fileset.NewIndexFilter(x, func(idx *index.Index) bool {
			return idx.Path == oldName || strings.HasPrefix(idx.Path, oldName+"/")
		})
		return x
	})
	new := NewSource(oldCommit, true, func() fileset.FileSource {
		x := d.storage.NewSource(ctx, compactedCommitPath(newCommit), index.WithPrefix(newName))
		x = fileset.NewIndexResolver(x)
		x = fileset.NewIndexFilter(x, func(idx *index.Index) bool {
			return idx.Path == newName || strings.HasPrefix(idx.Path, newName+"/")
		})
		return x
	})
	diff := NewDiffer(old, new)
	return diff.IterateDiff(pachClient.Ctx(), cb)
}
