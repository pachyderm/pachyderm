package server

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/dbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/tracker"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	storageTaskNamespace = "storage"
	tmpRepo              = client.TmpRepoName
	defaultTTL           = client.DefaultTTL
	maxTTL               = 30 * time.Minute
)

type driverV2 struct {
	*driver

	storage         *fileset.Storage
	compactionQueue *work.TaskQueue
}

// newDriver is used to create a new Driver instance
func newDriverV2(env *serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string, treeCache *hashtree.Cache, storageRoot string, memoryRequest int64) (*driverV2, error) {
	d1, err := newDriver(env, txnEnv, etcdPrefix, treeCache, storageRoot, memoryRequest)
	if err != nil {
		return nil, err
	}
	d2 := &driverV2{driver: d1}
	objClient, err := NewObjClient(env.Configuration)
	if err != nil {
		return nil, err
	}
	db, err := newDB()
	if err != nil {
		return nil, err
	}
	tracker := tracker.NewPGTracker(db)
	chunkStorageOpts, err := env.ChunkStorageOptions()
	if err != nil {
		return nil, err
	}
	chunkStorage := chunk.NewStorage(objClient, chunk.NewPostgresStore(db), tracker, chunkStorageOpts...)
	d2.storage = fileset.NewStorage(fileset.NewPostgresStore(db), tracker, chunkStorage, env.FileSetStorageOptions()...)
	d2.compactionQueue, err = work.NewTaskQueue(context.Background(), d2.etcdClient, d2.prefix, storageTaskNamespace)
	if err != nil {
		return nil, err
	}
	go d2.master(env)
	go d2.compactionWorker()
	return d2, nil
}

func newDB() (db *sqlx.DB, retErr error) {
	defer func() {
		if db != nil {
			db.MustExec(`DROP SCHEMA IF EXISTS storage CASCADE`)
			fileset.SetupPostgresStore(db)
			chunk.SetupPostgresStore(db)
			tracker.PGTrackerApplySchema(db)
		}
	}()
	postgresHost, ok := os.LookupEnv("POSTGRES_SERVICE_HOST")
	if !ok {
		// TODO: Probably not the right long term approach here, but this is necessary to handle the mock pachd instance used in tests.
		// It does not run in kubernetes, so we need to fallback on setting up a local database.
		return dbutil.NewDB(dbutil.DBParams{
			Host:   dbutil.DefaultPostgresHost,
			Port:   dbutil.DefaultPostgresPort,
			User:   dbutil.TestPostgresUser,
			DBName: "pgc",
		})
	}
	postgresPortStr, ok := os.LookupEnv("POSTGRES_SERVICE_PORT")
	if !ok {
		return nil, errors.Errorf("postgres service port not found")
	}
	postgresPort, err := strconv.Atoi(postgresPortStr)
	if err != nil {
		return nil, err
	}
	return dbutil.NewDB(dbutil.DBParams{
		Host:   postgresHost,
		Port:   postgresPort,
		User:   "pachyderm",
		Pass:   "elephantastic",
		DBName: "pgc",
	})
}

func (d *driverV2) finishCommitV2(txnCtx *txnenv.TransactionContext, commit *pfs.Commit, description string) error {
	commitInfo, err := d.resolveCommit(txnCtx.Stm, commit)
	if err != nil {
		return err
	}
	if commitInfo.Finished != nil {
		return pfsserver.ErrCommitFinished{commitInfo.Commit}
	}
	commit = commitInfo.Commit
	if description != "" {
		commitInfo.Description = description
	}
	commitPath := commitKey(commit)
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
			parentCommitPath := commitKey(commitInfo.ParentCommit)
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

func (d *driverV2) getSubFileSet() int64 {
	// TODO subFileSet will need to be incremented through postgres or etcd.
	return time.Now().UnixNano()
}

func (d *driverV2) fileOperation(pachClient *client.APIClient, commit *pfs.Commit, cb func(*fileset.UnorderedWriter) error) error {
	ctx := pachClient.Ctx()
	repo := commit.Repo.Name
	var branch string
	if !uuid.IsUUIDWithoutDashes(commit.ID) {
		branch = commit.ID
	}
	commitInfo, err := d.inspectCommit(pachClient, commit, pfs.CommitState_STARTED)
	if err != nil {
		if (!isNotFoundErr(err) && !isNoHeadErr(err)) || branch == "" {
			return err
		}
		return d.oneOffFileOperation(ctx, repo, branch, cb)
	}
	if commitInfo.Finished != nil {
		if branch == "" {
			return pfsserver.ErrCommitFinished{commitInfo.Commit}
		}
		return d.oneOffFileOperation(ctx, repo, branch, cb)
	}
	return d.withCommitWriter(ctx, commitInfo.Commit, cb)
}

// TODO: Cleanup after failure?
func (d *driverV2) oneOffFileOperation(ctx context.Context, repo, branch string, cb func(*fileset.UnorderedWriter) error) error {
	return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txnenv.TransactionContext) (retErr error) {
		commit, err := d.startCommit(txnCtx, "", client.NewCommit(repo, ""), branch, nil, "")
		if err != nil {
			return err
		}
		defer func() {
			if retErr == nil {
				retErr = d.finishCommitV2(txnCtx, commit, "")
			}
		}()
		return d.withCommitWriter(txnCtx.ClientContext, commit, cb)
	})
}

// withCommitWriter calls cb with an unordered writer. All data written to cb is added to the commit, or an error is returned.
func (d *driverV2) withCommitWriter(ctx context.Context, commit *pfs.Commit, cb func(*fileset.UnorderedWriter) error) (retErr error) {
	n := d.getSubFileSet()
	subFileSetStr := fileset.SubFileSetStr(n)
	subFileSetPath := path.Join(commit.Repo.Name, commit.ID, subFileSetStr)
	return d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		id, err := d.withTmpUnorderedWriter(ctx, renewer, false, cb)
		if err != nil {
			return err
		}
		tmpPath := path.Join(tmpRepo, id)
		return d.storage.Copy(ctx, tmpPath, subFileSetPath, 0)
	})
}

func (d *driverV2) withTmpUnorderedWriter(ctx context.Context, renewer *fileset.Renewer, compact bool, cb func(*fileset.UnorderedWriter) error) (string, error) {
	id := uuid.NewWithoutDashes()
	inputPath := path.Join(tmpRepo, id)
	opts := []fileset.UnorderedWriterOption{fileset.WithRenewal(defaultTTL, renewer)}
	defaultTag := fileset.SubFileSetStr(d.getSubFileSet())
	uw, err := d.storage.NewUnorderedWriter(ctx, inputPath, defaultTag, opts...)
	if err != nil {
		return "", err
	}
	if err := cb(uw); err != nil {
		return "", err
	}
	if err := uw.Close(); err != nil {
		return "", err
	}
	if compact {
		outputPath := path.Join(tmpRepo, id, fileset.Compacted)
		_, err := d.storage.Compact(ctx, outputPath, []string{inputPath}, defaultTTL)
		if err != nil {
			return "", err
		}
		renewer.Add(outputPath)
	}
	return id, nil
}

func (d *driverV2) withWriter(pachClient *client.APIClient, commit *pfs.Commit, cb func(string, *fileset.Writer) error) (retErr error) {
	ctx := pachClient.Ctx()
	commitInfo, err := d.inspectCommit(pachClient, commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if commitInfo.Finished != nil {
		return pfsserver.ErrCommitFinished{commitInfo.Commit}
	}
	commit = commitInfo.Commit
	n := d.getSubFileSet()
	subFileSetStr := fileset.SubFileSetStr(n)
	subFileSetPath := path.Join(commit.Repo.Name, commit.ID, subFileSetStr)
	fsw := d.storage.NewWriter(ctx, subFileSetPath)
	if err := cb(subFileSetStr, fsw); err != nil {
		return err
	}
	return fsw.Close()
}

func (d *driverV2) getTar(pachClient *client.APIClient, commit *pfs.Commit, glob string, w io.Writer) error {
	ctx := pachClient.Ctx()
	commitInfo, err := d.inspectCommit(pachClient, commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if commitInfo.Finished == nil {
		return pfsserver.ErrCommitNotFinished{commitInfo.Commit}
	}
	commit = commitInfo.Commit
	indexOpt, mf, err := parseGlob(glob)
	if err != nil {
		return err
	}
	fs, err := d.storage.Open(ctx, []string{compactedCommitPath(commit)}, indexOpt)
	if err != nil {
		return err
	}
	var dir string
	filter := fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
		if dir != "" && strings.HasPrefix(idx.Path, dir) {
			return true
		}
		match := mf(idx.Path)
		if match && fileset.IsDir(idx.Path) {
			dir = idx.Path
		}
		return match
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

func (d *driverV2) listFileV2(pachClient *client.APIClient, file *pfs.File, full bool, history int64, cb func(*pfs.FileInfo) error) error {
	if _, err := d.inspectCommit(pachClient, file.Commit, pfs.CommitState_FINISHED); err != nil {
		return err
	}
	ctx := pachClient.Ctx()
	commitInfo, err := d.inspectCommit(pachClient, file.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if commitInfo.Finished == nil {
		return pfsserver.ErrCommitNotFinished{commitInfo.Commit}
	}
	commit := commitInfo.Commit
	name := cleanPath(file.Path)
	fs, err := d.storage.Open(ctx, []string{compactedCommitPath(commit)}, index.WithPrefix(name))
	if err != nil {
		return err
	}
	fs = d.storage.NewIndexResolver(fs)
	fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
		if idx.Path == "/" {
			return false
		}
		if idx.Path == name {
			return true
		}
		if idx.Path == name+"/" {
			return false
		}
		return strings.HasPrefix(idx.Path, name)
	})
	s := NewSource(commit, fs, true)
	return s.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
		if pathIsChild(name, cleanPath(fi.File.Path)) {
			return cb(fi)
		}
		return nil
	})
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
	var outputSize int64
	if err := d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		res, err := d.compactIter(ctx, compactSpec{
			master:     master,
			inputPaths: inputPaths,
			maxFanIn:   d.env.StorageCompactionMaxFanIn,
		})
		if err != nil {
			return err
		}
		renewer.Add(res.OutputPath)
		outputSize = res.OutputSize
		return d.storage.Copy(ctx, res.OutputPath, outputPath, 0)
	}); err != nil {
		return nil, err
	}
	return &compactStats{OutputSize: outputSize}, nil
}

type compactSpec struct {
	master     *work.Master
	inputPaths []string
	maxFanIn   int
}

type compactResult struct {
	OutputPath string
	OutputSize int64
}

// compactIter is one level of compaction.  It will only perform compaction
// if len(inputPaths) <= params.maxFanIn otherwise it will split inputPaths recursively.
func (d *driverV2) compactIter(ctx context.Context, params compactSpec) (*compactResult, error) {
	if len(params.inputPaths) <= params.maxFanIn {
		return d.shardedCompact(ctx, params.master, params.inputPaths)
	}
	childSize := len(params.inputPaths) / params.maxFanIn
	if len(params.inputPaths)%params.maxFanIn != 0 {
		childSize++
	}
	// TODO: use an errgroup to make the recursion concurrecnt.
	// this requires changing the master to allow multiple calls to RunSubtasks
	// don't forget to pass the errgroups childCtx to compactIter instead of ctx.
	// TODO: change this such that the fan in is maxed at the lower levels first rather
	// than the higher.
	var res *compactResult
	if err := d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		var childOutputPaths []string
		for i := 0; i < params.maxFanIn; i++ {
			start := i * childSize
			if start >= len(params.inputPaths) {
				break
			}
			end := (i + 1) * childSize
			if end > len(params.inputPaths) {
				end = len(params.inputPaths)
			}
			res, err := d.compactIter(ctx, compactSpec{
				master:     params.master,
				inputPaths: params.inputPaths[start:end],
				maxFanIn:   params.maxFanIn,
			})
			if err != nil {
				return err
			}
			renewer.Add(res.OutputPath)
			childOutputPaths = append(childOutputPaths, res.OutputPath)
		}
		var err error
		res, err = d.shardedCompact(ctx, params.master, childOutputPaths)
		return err
	}); err != nil {
		return nil, err
	}
	return res, nil
}

// shardedCompact generates shards for the fileset(s) in inputPaths,
// gives those shards to workers, and waits for them to complete.
// Fan in is bound by len(inputPaths), concatenating shards have
// fan in of one because they are concatenated sequentially.
func (d *driverV2) shardedCompact(ctx context.Context, master *work.Master, inputPaths []string) (*compactResult, error) {
	scratch := path.Join(tmpRepo, uuid.NewWithoutDashes())
	compaction := &pfs.Compaction{InputPrefixes: inputPaths}
	var subtasks []*work.Task
	var shardOutputs []string
	fs, err := d.storage.OpenWithDeletes(ctx, inputPaths)
	if err != nil {
		return nil, err
	}
	if err := d.storage.Shard(ctx, fs, func(pathRange *index.PathRange) error {
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
	var res *compactResult
	if err := d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		if err := master.RunSubtasks(subtasks, func(_ context.Context, taskInfo *work.TaskInfo) error {
			if taskInfo.State == work.State_FAILURE {
				return errors.Errorf(taskInfo.Reason)
			}
			shard, err := deserializeShard(taskInfo.Task.Data)
			if err != nil {
				return err
			}
			renewer.Add(shard.OutputPath)
			return nil
		}); err != nil {
			return err
		}
		var err error
		res, err = d.concatFileSets(ctx, shardOutputs)
		return err
	}); err != nil {
		return nil, err
	}
	return res, nil
}

// concatFileSets concatenates the filesets in inputPaths and writes the result to outputPath
// TODO: move this to the fileset package, and error if the entries are not sorted.
func (d *driverV2) concatFileSets(ctx context.Context, inputPaths []string) (*compactResult, error) {
	outputPath := path.Join(tmpRepo, uuid.NewWithoutDashes())
	var size int64
	fsw := d.storage.NewWriter(ctx, outputPath, fileset.WithIndexCallback(func(idx *index.Index) error {
		size += idx.SizeBytes
		return nil
	}), fileset.WithTTL(defaultTTL))
	for _, inputPath := range inputPaths {
		fs, err := d.storage.OpenWithDeletes(ctx, []string{inputPath})
		if err != nil {
			return nil, err
		}
		if err := fileset.CopyFiles(ctx, fsw, fs); err != nil {
			return nil, err
		}
	}
	if err := fsw.Close(); err != nil {
		return nil, err
	}
	return &compactResult{OutputPath: outputPath, OutputSize: size}, nil
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
			_, err = d.storage.Compact(ctx, shard.OutputPath, shard.Compaction.InputPrefixes, defaultTTL, index.WithRange(pathRange))
			if err != nil {
				panic(err)
			}
			return err
		})
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Printf("error in compaction worker: %v", err)
		return nil
	})
	// Never ending backoff should prevent us from getting here.
	panic(err)
}

func (d *driverV2) globFileV2(pachClient *client.APIClient, commit *pfs.Commit, glob string, cb func(*pfs.FileInfo) error) (retErr error) {
	ctx := pachClient.Ctx()
	commitInfo, err := d.inspectCommit(pachClient, commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if commitInfo.Finished == nil {
		return pfsserver.ErrCommitNotFinished{commitInfo.Commit}
	}
	commit = commitInfo.Commit
	indexOpt, mf, err := parseGlob(glob)
	if err != nil {
		return err
	}
	fs, err := d.storage.Open(ctx, []string{compactedCommitPath(commit)}, indexOpt)
	if err != nil {
		return err
	}
	fs = d.storage.NewIndexResolver(fs)
	s := NewSource(commit, fs, true)
	return s.Iterate(ctx, func(fi *pfs.FileInfo, f fileset.File) error {
		if !mf(fi.File.Path) {
			return nil
		}
		return cb(fi)
	})
}

func (d *driverV2) copyFile(pachClient *client.APIClient, src *pfs.File, dst *pfs.File, overwrite bool) (retErr error) {
	ctx := pachClient.Ctx()
	srcCommitInfo, err := d.inspectCommit(pachClient, src.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if srcCommitInfo.Finished == nil {
		return pfsserver.ErrCommitNotFinished{srcCommitInfo.Commit}
	}
	srcCommit := srcCommitInfo.Commit
	dstCommitInfo, err := d.inspectCommit(pachClient, dst.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if dstCommitInfo.Finished != nil {
		return pfsserver.ErrCommitFinished{dstCommitInfo.Commit}
	}
	dstCommit := dstCommitInfo.Commit
	if overwrite {
		// TODO: after delete merging is sorted out add overwrite support
		return errors.New("overwrite not yet supported")
	}
	srcPath := cleanPath(src.Path)
	dstPath := cleanPath(dst.Path)
	pathTransform := func(x string) string {
		relPath, err := filepath.Rel(srcPath, x)
		if err != nil {
			panic("cannot apply path transform")
		}
		return path.Join(dstPath, relPath)
	}
	fs, err := d.storage.Open(ctx, []string{compactedCommitPath(srcCommit)}, index.WithPrefix(srcPath))
	if err != nil {
		return err
	}
	fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
		return idx.Path == srcPath || strings.HasPrefix(idx.Path, srcPath+"/")
	})
	fs = fileset.NewHeaderMapper(fs, func(th *tar.Header) *tar.Header {
		th.Name = pathTransform(th.Name)
		return th
	})
	fs = fileset.NewDirInserter(fs)
	return d.withWriter(pachClient, dstCommit, func(tag string, dst *fileset.Writer) error {
		return fs.Iterate(ctx, func(f fileset.File) error {
			hdr, err := f.Header()
			if err != nil {
				return err
			}
			return dst.Append(hdr.Name, func(fw *fileset.FileWriter) error {
				hdr, err := f.Header()
				if err != nil {
					return err
				}
				return fileset.WithTarFileWriter(fw, hdr, func(tfw *fileset.TarFileWriter) error {
					tfw.Append(tag)
					return f.Content(tfw)
				})
			})
		})
	})
}

func (d *driverV2) diffFileV2(pachClient *client.APIClient, oldFile, newFile *pfs.File, cb func(oldFi, newFi *pfs.FileInfo) error) error {
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
		oldFile = &pfs.File{
			Commit: newCommitInfo.ParentCommit,
			Path:   newFile.Path,
		}
	}
	ctx := pachClient.Ctx()
	oldCommit := oldFile.Commit
	newCommit := newFile.Commit
	oldName := cleanPath(oldFile.Path)
	if oldName == "/" {
		oldName = ""
	}
	newName := cleanPath(newFile.Path)
	if newName == "/" {
		newName = ""
	}
	var old Source = emptySource{}
	if oldCommit != nil {
		fs, err := d.storage.Open(ctx, []string{compactedCommitPath(oldCommit)}, index.WithPrefix(oldName))
		if err != nil {
			return err
		}
		fs = d.storage.NewIndexResolver(fs)
		fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
			return idx.Path == oldName || strings.HasPrefix(idx.Path, oldName+"/")
		})
		old = NewSource(oldCommit, fs, true)
	}
	fs, err := d.storage.Open(ctx, []string{compactedCommitPath(newCommit)}, index.WithPrefix(newName))
	if err != nil {
		return err
	}
	fs = d.storage.NewIndexResolver(fs)
	fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
		return idx.Path == newName || strings.HasPrefix(idx.Path, newName+"/")
	})
	new := NewSource(newCommit, fs, true)
	diff := NewDiffer(old, new)
	return diff.Iterate(pachClient.Ctx(), cb)
}

func (d *driverV2) inspectFile(pachClient *client.APIClient, file *pfs.File) (*pfs.FileInfo, error) {
	ctx := pachClient.Ctx()
	commitInfo, err := d.inspectCommit(pachClient, file.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return nil, err
	}
	if commitInfo.Finished == nil {
		return nil, pfsserver.ErrCommitNotFinished{commitInfo.Commit}
	}
	commit := commitInfo.Commit
	p := cleanPath(file.Path)
	if p == "/" {
		p = ""
	}
	fs, err := d.storage.Open(ctx, []string{compactedCommitPath(commit)}, index.WithPrefix(p))
	if err != nil {
		return nil, err
	}
	fs = d.storage.NewIndexResolver(fs)
	fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
		return idx.Path == p || strings.HasPrefix(idx.Path, p+"/")
	})
	s := NewSource(commit, fs, true)
	var ret *pfs.FileInfo
	s = NewErrOnEmpty(s, &pfsserver.ErrFileNotFound{File: file})
	if err := s.Iterate(ctx, func(fi *pfs.FileInfo, f fileset.File) error {
		p2 := fi.File.Path
		if p2 == p || p2 == p+"/" {
			ret = fi
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *driverV2) walkFile(pachClient *client.APIClient, file *pfs.File, cb func(*pfs.FileInfo) error) (retErr error) {
	ctx := pachClient.Ctx()
	commitInfo, err := d.inspectCommit(pachClient, file.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if commitInfo.Finished == nil {
		return pfsserver.ErrCommitNotFinished{commitInfo.Commit}
	}
	commit := commitInfo.Commit
	p := cleanPath(file.Path)
	if p == "/" {
		p = ""
	}
	fs, err := d.storage.Open(ctx, []string{compactedCommitPath(commit)}, index.WithPrefix(p))
	if err != nil {
		return err
	}
	fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
		return idx.Path == p || strings.HasPrefix(idx.Path, p+"/")
	})
	s := NewSource(commit, fs, false)
	s = NewErrOnEmpty(s, &pfsserver.ErrFileNotFound{File: file})
	return s.Iterate(ctx, func(fi *pfs.FileInfo, f fileset.File) error {
		return cb(fi)
	})
}

func (d *driverV2) clearCommitV2(pachClient *client.APIClient, commit *pfs.Commit) error {
	ctx := pachClient.Ctx()
	commitInfo, err := d.inspectCommit(pachClient, commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if commitInfo.Finished != nil {
		return errors.Errorf("cannot clear finished commit")
	}
	return d.storage.Delete(ctx, commitPath(commit))
}

func (d *driverV2) deleteRepo(txnCtx *txnenv.TransactionContext, repo *pfs.Repo, force bool) error {
	ctx := txnCtx.ClientContext
	if err := d.storage.WalkFileSet(ctx, repo.Name, func(p string) error {
		return d.storage.Delete(ctx, p)
	}); err != nil {
		return err
	}
	return d.driver.deleteRepo(txnCtx, repo, force)
}

func (d *driverV2) deleteAll(txnCtx *txnenv.TransactionContext) error {
	// Note: d.listRepo() doesn't return the 'spec' repo, so it doesn't get
	// deleted here. Instead, PPS is responsible for deleting and re-creating it
	repoInfos, err := d.listRepo(txnCtx.Client, !includeAuth)
	if err != nil {
		return err
	}
	for _, repoInfo := range repoInfos.RepoInfo {
		if err := d.deleteRepo(txnCtx, repoInfo.Repo, true); err != nil && !auth.IsErrNotAuthorized(err) {
			return err
		}
	}
	return nil
}

func (d *driverV2) deleteCommit(txnCtx *txnenv.TransactionContext, userCommit *pfs.Commit) error {
	// Main txn: Delete all downstream commits, and update subvenance of upstream commits
	// TODO update branches inside this txn, by storing a repo's branches in its
	// RepoInfo or its HEAD commit
	deleted := make(map[string]*pfs.CommitInfo) // deleted commits
	affectedRepos := make(map[string]struct{})  // repos containing deleted commits

	// 1) re-read CommitInfo inside txn
	userCommitInfo, err := d.resolveCommit(txnCtx.Stm, userCommit)
	if err != nil {
		return errors.Wrapf(err, "resolveCommit")
	}

	// 2) Define helper for deleting commits. 'lower' corresponds to
	// pfs.CommitRange.Lower, and is an ancestor of 'upper'
	deleteCommit := func(lower, upper *pfs.Commit) error {
		// Validate arguments
		if lower.Repo.Name != upper.Repo.Name {
			return errors.Errorf("cannot delete commit range with mismatched repos \"%s\" and \"%s\"", lower.Repo.Name, upper.Repo.Name)
		}
		affectedRepos[lower.Repo.Name] = struct{}{}
		commits := d.commits(lower.Repo.Name).ReadWrite(txnCtx.Stm)

		// delete commits on path upper -> ... -> lower (traverse ParentCommits)
		commit := upper
		for {
			if commit == nil {
				return errors.Errorf("encountered nil parent commit in %s/%s...%s", lower.Repo.Name, lower.ID, upper.ID)
			}
			// Store commitInfo in 'deleted' and remove commit from etcd
			commitInfo := &pfs.CommitInfo{}
			if err := commits.Get(commit.ID, commitInfo); err != nil {
				return err
			}
			// If a commit has already been deleted, we don't want to overwrite the existing information, since commitInfo will be nil
			if _, ok := deleted[commit.ID]; !ok {
				deleted[commit.ID] = commitInfo
			}
			if err := commits.Delete(commit.ID); err != nil {
				return err
			}
			// Delete the commit's filesets
			if err := d.storage.Delete(txnCtx.Client.Ctx(), path.Join(commit.Repo.Name, commit.ID)); err != nil {
				return err
			}
			if commit.ID == lower.ID {
				break // check after deletion so we delete 'lower' (inclusive range)
			}
			commit = commitInfo.ParentCommit
		}

		return nil
	}

	// 3) Validate the commit (check that it has no provenance) and delete it
	if provenantOnInput(userCommitInfo.Provenance) {
		return errors.Errorf("cannot delete the commit \"%s/%s\" because it has non-empty provenance", userCommit.Repo.Name, userCommit.ID)
	}
	deleteCommit(userCommitInfo.Commit, userCommitInfo.Commit)

	// 4) Delete all of the downstream commits of 'commit'
	for _, subv := range userCommitInfo.Subvenance {
		deleteCommit(subv.Lower, subv.Upper)
	}

	// 5) Remove the commits in 'deleted' from all remaining upstream commits'
	// subvenance.
	// While 'commit' is required to be an input commit (no provenance),
	// downstream commits from 'commit' may have multiple inputs, and those
	// other inputs must have their subvenance updated
	visited := make(map[string]bool) // visitied upstream (provenant) commits
	for _, deletedInfo := range deleted {
		for _, prov := range deletedInfo.Provenance {
			// Check if we've fixed provCommit already (or if it's deleted and
			// doesn't need to be fixed
			if _, isDeleted := deleted[prov.Commit.ID]; isDeleted || visited[prov.Commit.ID] {
				continue
			}
			visited[prov.Commit.ID] = true

			// fix provCommit's subvenance
			provCI := &pfs.CommitInfo{}
			if err := d.commits(prov.Commit.Repo.Name).ReadWrite(txnCtx.Stm).Update(prov.Commit.ID, provCI, func() error {
				subvTo := 0 // copy subvFrom to subvTo, excepting subv ranges to delete (so that they're overwritten)
			nextSubvRange:
				for subvFrom, subv := range provCI.Subvenance {
					// Compute path (of commit IDs) connecting subv.Upper to subv.Lower
					cur := subv.Upper.ID
					path := []string{cur}
					for cur != subv.Lower.ID {
						// Get CommitInfo for 'cur' (either in 'deleted' or from etcd)
						// and traverse parent
						curInfo, ok := deleted[cur]
						if !ok {
							curInfo = &pfs.CommitInfo{}
							if err := d.commits(subv.Lower.Repo.Name).ReadWrite(txnCtx.Stm).Get(cur, curInfo); err != nil {
								return errors.Wrapf(err, "error reading commitInfo for subvenant \"%s/%s\"", subv.Lower.Repo.Name, cur)
							}
						}
						if curInfo.ParentCommit == nil {
							break
						}
						cur = curInfo.ParentCommit.ID
						path = append(path, cur)
					}

					// move 'subv.Upper' through parents until it points to a non-deleted commit
					for j := range path {
						if _, ok := deleted[subv.Upper.ID]; !ok {
							break
						}
						if j+1 >= len(path) {
							// All commits in subvRange are deleted. Remove entire Range
							// from provCI.Subvenance
							continue nextSubvRange
						}
						subv.Upper.ID = path[j+1]
					}

					// move 'subv.Lower' through children until it points to a non-deleted commit
					for j := len(path) - 1; j >= 0; j-- {
						if _, ok := deleted[subv.Lower.ID]; !ok {
							break
						}
						// We'll eventually get to a non-deleted commit because the
						// 'upper' block didn't exit
						subv.Lower.ID = path[j-1]
					}
					provCI.Subvenance[subvTo] = provCI.Subvenance[subvFrom]
					subvTo++
				}
				provCI.Subvenance = provCI.Subvenance[:subvTo]
				return nil
			}); err != nil {
				return errors.Wrapf(err, "err fixing subvenance of upstream commit %s/%s", prov.Commit.Repo.Name, prov.Commit.ID)
			}
		}
	}

	// 6) Rewrite ParentCommit of deleted commits' children, and
	// ChildCommits of deleted commits' parents
	visited = make(map[string]bool) // visited child/parent commits
	for deletedID, deletedInfo := range deleted {
		if visited[deletedID] {
			continue
		}

		// Traverse downwards until we find the lowest (most ancestral)
		// non-nil, deleted commit
		lowestCommitInfo := deletedInfo
		for {
			if lowestCommitInfo.ParentCommit == nil {
				break // parent is nil
			}
			parentInfo, ok := deleted[lowestCommitInfo.ParentCommit.ID]
			if !ok {
				break // parent is not deleted
			}
			lowestCommitInfo = parentInfo // parent exists and is deleted--go down
		}

		// BFS upwards through graph for all non-deleted children
		var next *pfs.Commit                            // next vertex to search
		queue := []*pfs.Commit{lowestCommitInfo.Commit} // queue of vertices to explore
		liveChildren := make(map[string]struct{})       // live children discovered so far
		for len(queue) > 0 {
			next, queue = queue[0], queue[1:]
			if visited[next.ID] {
				continue
			}
			visited[next.ID] = true
			nextInfo, ok := deleted[next.ID]
			if !ok {
				liveChildren[next.ID] = struct{}{}
				continue
			}
			queue = append(queue, nextInfo.ChildCommits...)
		}

		// Point all non-deleted children at the first valid parent (or nil),
		// and point first non-deleted parent at all non-deleted children
		commits := d.commits(deletedInfo.Commit.Repo.Name).ReadWrite(txnCtx.Stm)
		parent := lowestCommitInfo.ParentCommit
		for child := range liveChildren {
			commitInfo := &pfs.CommitInfo{}
			if err := commits.Update(child, commitInfo, func() error {
				commitInfo.ParentCommit = parent
				return nil
			}); err != nil {
				return errors.Wrapf(err, "err updating child commit %v", lowestCommitInfo.Commit)
			}
		}
		if parent != nil {
			commitInfo := &pfs.CommitInfo{}
			if err := commits.Update(parent.ID, commitInfo, func() error {
				// Add existing live commits in commitInfo.ChildCommits to the
				// live children above lowestCommitInfo, then put them all in
				// 'parent'
				for _, child := range commitInfo.ChildCommits {
					if _, ok := deleted[child.ID]; ok {
						continue
					}
					liveChildren[child.ID] = struct{}{}
				}
				commitInfo.ChildCommits = make([]*pfs.Commit, 0, len(liveChildren))
				for child := range liveChildren {
					commitInfo.ChildCommits = append(commitInfo.ChildCommits, client.NewCommit(parent.Repo.Name, child))
				}
				return nil
			}); err != nil {
				return errors.Wrapf(err, "err rewriting children of ancestor commit %v", lowestCommitInfo.Commit)
			}
		}
	}

	// 7) Traverse affected repos and rewrite all branches so that no branch
	// points to a deleted commit
	var affectedBranches []*pfs.BranchInfo
	repos := d.repos.ReadWrite(txnCtx.Stm)
	for repo := range affectedRepos {
		repoInfo := &pfs.RepoInfo{}
		if err := repos.Get(repo, repoInfo); err != nil {
			return err
		}
		for _, brokenBranch := range repoInfo.Branches {
			// Traverse HEAD commit until we find a non-deleted parent or nil;
			// rewrite branch
			var branchInfo pfs.BranchInfo
			if err := d.branches(brokenBranch.Repo.Name).ReadWrite(txnCtx.Stm).Update(brokenBranch.Name, &branchInfo, func() error {
				prevHead := branchInfo.Head
				for {
					if branchInfo.Head == nil {
						return nil // no commits left in branch
					}
					headCommitInfo, headIsDeleted := deleted[branchInfo.Head.ID]
					if !headIsDeleted {
						break
					}
					branchInfo.Head = headCommitInfo.ParentCommit
				}
				if prevHead != nil && prevHead.ID != branchInfo.Head.ID {
					affectedBranches = append(affectedBranches, &branchInfo)
				}
				return err
			}); err != nil && !col.IsErrNotFound(err) {
				// If err is NotFound, branch is in downstream provenance but
				// doesn't exist yet--nothing to update
				return errors.Wrapf(err, "error updating branch %v/%v", brokenBranch.Repo.Name, brokenBranch.Name)
			}

			// Update repo size if this is the master branch
			if branchInfo.Name == "master" {
				if branchInfo.Head != nil {
					headCommitInfo, err := d.resolveCommit(txnCtx.Stm, branchInfo.Head)
					if err != nil {
						return err
					}
					repoInfo.SizeBytes = headCommitInfo.SizeBytes
				} else {
					// No HEAD commit, set the repo size to 0
					repoInfo.SizeBytes = 0
				}

				if err := repos.Put(repo, repoInfo); err != nil {
					return err
				}
			}
		}
	}

	// 8) propagate the changes to 'branch' and its subvenance. This may start
	// new HEAD commits downstream, if the new branch heads haven't been
	// processed yet
	for _, afBranch := range affectedBranches {
		if err := txnCtx.PropagateCommit(afBranch.Branch, false); err != nil {
			return err
		}
	}

	return nil
}

func (d *driverV2) createTmpFileSet(server pfs.API_CreateTmpFileSetServer) (string, error) {
	ctx := server.Context()
	var id string
	if err := d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		var err error
		id, err = d.withTmpUnorderedWriter(ctx, renewer, true, func(uw *fileset.UnorderedWriter) error {
			req := &pfs.PutTarRequestV2{
				Tag: "",
			}
			_, err := putTar(uw, server, req)
			return err
		})
		return err
	}); err != nil {
		return "", err
	}
	return id, nil
}

func (d *driverV2) renewTmpFileSet(ctx context.Context, id string, ttl time.Duration) error {
	if ttl < time.Second {
		return errors.Errorf("ttl (%d) must be at least one second", ttl)
	}
	if ttl > maxTTL {
		return errors.Errorf("ttl (%d) exceeds max ttl (%d)", ttl, maxTTL)
	}
	// check that it is the correct length, to prevent malicious renewing of multiple filesets
	// len(hex(uuid)) == 32
	if len(id) != 32 {
		return errors.Errorf("invalid id (%s)", id)
	}
	p := path.Join(tmpRepo, id)
	_, err := d.storage.SetTTL(ctx, p, ttl)
	return err
}

func (d *driverV2) inspectCommit(pachClient *client.APIClient, commit *pfs.Commit, blockState pfs.CommitState) (*pfs.CommitInfo, error) {
	if commit.GetRepo().GetName() == tmpRepo {
		cinfo := &pfs.CommitInfo{
			Commit:      commit,
			Description: "Temporary FileSet",
			Finished:    &types.Timestamp{}, // it's always been finished. How did you get the id if it wasn't finished?
		}
		return cinfo, nil
	}
	return d.driver.inspectCommit(pachClient, commit, blockState)
}
