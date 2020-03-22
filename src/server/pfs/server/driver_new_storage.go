package server

import (
	"archive/tar"
	"fmt"
	"hash"
	"io"
	"log"
	"path"
	"regexp"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"golang.org/x/net/context"
)

const (
	// tmpPrefix is for temporary object paths that store compacted shards.
	tmpPrefix            = "tmp"
	storageTaskNamespace = "storage"
)

func (d *driver) startCommitNewStorageLayer(txnCtx *txnenv.TransactionContext, id string, parent *pfs.Commit, branch string, provenance []*pfs.CommitProvenance, description string) (*pfs.Commit, error) {
	return d.startCommit(txnCtx, id, parent, branch, provenance, description)
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
	// Run compaction task.
	return d.compactionQueue.RunTaskBlock(txnCtx.Client.Ctx(), func(m *work.Master) error {
		commitPath := path.Join(commit.Repo.Name, commit.ID)
		// (bryce) need some cleanup improvements, probably garbage collection.
		// (bryce) this should be a retry when garbage collection is integrated.
		d.storage.Delete(context.Background(), path.Join(commitPath, fileset.Diff))
		d.storage.Delete(context.Background(), path.Join(commitPath, fileset.Compacted))
		// Compact the commit changes into a diff file set.
		if err := d.compact(m, path.Join(commitPath, fileset.Diff), []string{commitPath}); err != nil {
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
		if err := d.compact(m, compactSpec.Output, compactSpec.Input); err != nil {
			return err
		}
		// (bryce) need size.
		commitInfo.SizeBytes = uint64(0)
		commitInfo.Finished = now()
		return d.writeFinishedCommit(txnCtx.Stm, commit, commitInfo)
	})
}

// (bryce) add commit validation.
// (bryce) a failed put (crash/error) should result with any serialized sub file sets getting
// cleaned up.
// (bryce) probably should prevent / clean files that end with "/", since that will indicate a directory.
func (d *driver) putFilesNewStorageLayer(ctx context.Context, repo, commit string, r io.Reader) (retErr error) {
	// (bryce) subFileSet will need to be incremented through etcd eventually.
	d.mu.Lock()
	subFileSetStr := fileset.SubFileSetStr(d.subFileSet)
	fs := d.storage.New(ctx, path.Join(repo, commit, subFileSetStr), subFileSetStr)
	d.subFileSet++
	d.mu.Unlock()
	defer func() {
		if err := fs.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return fs.Put(r)
}

func (d *driver) getFilesNewStorageLayer(ctx context.Context, repo, commit, glob string, w io.Writer) error {
	// (bryce) glob should be cleaned in option function
	// (bryce) need exact match option for file glob.
	compactedPath := path.Join(repo, commit, fileset.Compacted)
	mr, err := d.storage.NewMergeReader(ctx, []string{compactedPath}, index.WithPrefix(glob))
	if err != nil {
		return err
	}
	return mr.Get(w)
}

var globRegex = regexp.MustCompile(`[*?[\]{}!()@+^]`)

func globLiteralPrefix(glob string) string {
	idx := globRegex.FindStringIndex(glob)
	if idx == nil {
		return glob
	}
	return glob[:idx[0]]
}

func (d *driver) getFilesConditional(ctx context.Context, repo, commit, glob string, f func(*FileReader) error) error {
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

func matchFunc(glob string) (func(string) bool, error) {
	// (bryce) this is a little weird, but it prevents the parent directory from being matched (i.e. /*).
	var parentG *globlib.Glob
	parentGlob, baseGlob := path.Split(glob)
	if len(baseGlob) > 0 {
		var err error
		parentG, err = globlib.Compile(parentGlob, '/')
		if err != nil {
			return nil, err
		}
	}
	g, err := globlib.Compile(glob, '/')
	if err != nil {
		return nil, err
	}
	return func(s string) bool {
		return g.Match(s) && (parentG == nil || !parentG.Match(s))
	}, nil
}

// FileReader is a PFS wrapper for a fileset.MergeReader.
// The primary purpose of this abstraction is to convert from index.Index to
// pfs.FileInfoNewStorage and to convert a set of index hashes to a file hash.
type FileReader struct {
	file      *pfs.File
	idx       *index.Index
	fmr       *fileset.FileMergeReader
	mr        *fileset.MergeReader
	fileCount int
	hash      hash.Hash
}

func newFileReader(file *pfs.File, idx *index.Index, fmr *fileset.FileMergeReader, mr *fileset.MergeReader) *FileReader {
	h := pfs.NewHash()
	for _, dataRef := range idx.DataOp.DataRefs {
		h.Write([]byte(dataRef.Hash))
	}
	return &FileReader{
		file: file,
		idx:  idx,
		fmr:  fmr,
		mr:   mr,
		hash: h,
	}
}

func (fr *FileReader) updateFileInfo(idx *index.Index) {
	fr.fileCount++
	for _, dataRef := range idx.DataOp.DataRefs {
		fr.hash.Write([]byte(dataRef.Hash))
	}
}

func (fr *FileReader) Info() *pfs.FileInfoNewStorage {
	return &pfs.FileInfoNewStorage{
		File: fr.file,
		Hash: pfs.EncodeHash(fr.hash.Sum(nil)),
	}
}

func (fr *FileReader) Get(w io.Writer) error {
	if err := fr.fmr.Get(w); err != nil {
		return err
	}
	for fr.fileCount > 0 {
		fmr, err := fr.mr.Next()
		if err != nil {
			return err
		}
		if err := fmr.Get(w); err != nil {
			return err
		}
		fr.fileCount--
	}
	// Close a tar writer to create tar EOF padding.
	return tar.NewWriter(w).Close()
}

func (fr *FileReader) drain() error {
	for fr.fileCount > 0 {
		if _, err := fr.mr.Next(); err != nil {
			return err
		}
		fr.fileCount--
	}
	return nil
}

func (d *driver) compact(master *work.Master, outputPath string, inputPrefixes []string) (retErr error) {
	scratch := path.Join(tmpPrefix, uuid.NewWithoutDashes())
	defer func() {
		// (bryce) need some cleanup improvements, probably garbage collection.
		if err := d.storage.Delete(context.Background(), scratch); retErr == nil {
			retErr = err
		}
	}()
	compaction := &pfs.Compaction{InputPrefixes: inputPrefixes}
	var subtasks []*work.Task
	if err := d.storage.Shard(master.Ctx(), inputPrefixes, func(pathRange *index.PathRange) error {
		outputPath := path.Join(scratch, strconv.Itoa(len(subtasks)))
		shard, err := serializeShard(&pfs.Shard{
			Compaction: compaction,
			Range: &pfs.PathRange{
				Lower: pathRange.Lower,
				Upper: pathRange.Upper,
			},
			OutputPath: outputPath,
		})
		if err != nil {
			return err
		}
		subtasks = append(subtasks, &work.Task{Data: shard})
		return nil
	}); err != nil {
		return err
	}
	if err := master.RunSubtasks(subtasks, func(_ context.Context, taskInfo *work.TaskInfo) error {
		if taskInfo.State == work.State_FAILURE {
			return fmt.Errorf(taskInfo.Reason)
		}
		return nil
	}); err != nil {
		return err
	}
	return d.storage.Compact(master.Ctx(), outputPath, []string{scratch})
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

// (bryce) it might potentially make sense to exit if an error occurs in this function
// because each pachd instance that errors here will lose its compaction worker without an obvious
// notification for the user (outside of the log message).
// (bryce) ^ maybe just a retry would be good enough.
func (d *driver) compactionWorker() {
	w := work.NewWorker(d.etcdClient, d.prefix, storageTaskNamespace)
	if err := w.Run(context.Background(), func(ctx context.Context, subtask *work.Task) error {
		shard, err := deserializeShard(subtask.Data)
		if err != nil {
			return err
		}
		pathRange := &index.PathRange{
			Lower: shard.Range.Lower,
			Upper: shard.Range.Upper,
		}
		return d.storage.Compact(ctx, shard.OutputPath, shard.Compaction.InputPrefixes, index.WithRange(pathRange))
	}); err != nil {
		log.Printf("error in compaction worker: %v", err)
	}
}
