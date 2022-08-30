/*
Package datum handles datums, the core units of work in a pipeline.

We use a specific file system structure to store datums, whether in meta commits or in temporary file sets during job processing.
Every datum has a JSON-marshalled [datum.Meta] at the file path

	/meta/<datum-ID>/meta

When computing a job's list of datums, the ID is prefixed with an index via [datum.WithPrefixIndex] to ensure a cache-friendly ordering.
Meta commits additionally track output files, so that they can be deleted per datum.
These files are under

	/pfs/<datum-ID>/out/

with file structure identical to the output commit (or more specifically, to the state of the worker's file system after the datum).
*/
package datum

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	io "io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfssync"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	workerStats "github.com/pachyderm/pachyderm/v2/src/server/worker/stats"
)

const (
	defaultNumRetries = 3
)

// SetSpec specifies criteria for creating datum sets.
type SetSpec struct {
	Number    int64
	SizeBytes int64
}

// CreateSets creates datum sets from the passed in datum iterator.
func CreateSets(dit Iterator, storageRoot string, setSpec *SetSpec, upload func(func(client.ModifyFile) error) error) error {
	var metas []*Meta
	shouldCreateSet := shouldCreateSetFunc(setSpec)
	if err := dit.Iterate(func(meta *Meta) error {
		metas = append(metas, meta)
		if shouldCreateSet(meta) {
			if err := createSet(metas, storageRoot, upload); err != nil {
				return err
			}
			metas = nil
		}
		return nil
	}); err != nil {
		return errors.EnsureStack(err)
	}
	if len(metas) > 0 {
		return createSet(metas, storageRoot, upload)
	}
	return nil
}

func shouldCreateSetFunc(setSpec *SetSpec) func(*Meta) bool {
	switch {
	case setSpec.Number > 0:
		var num int64
		return func(meta *Meta) bool {
			num++
			if num >= setSpec.Number {
				num = 0
				return true
			}
			return false
		}
	case setSpec.SizeBytes > 0:
		var size int64
		return func(meta *Meta) bool {
			for _, input := range meta.Inputs {
				size += int64(input.FileInfo.SizeBytes)
			}
			if size >= setSpec.SizeBytes {
				size = 0
				return true
			}
			return false
		}
	default:
		return func(meta *Meta) bool {
			return true
		}
	}
}

func createSet(metas []*Meta, storageRoot string, upload func(func(client.ModifyFile) error) error) error {
	return upload(func(mf client.ModifyFile) error {
		return WithSet(nil, storageRoot, func(s *Set) error {
			for _, meta := range metas {
				if err := s.UploadMeta(meta, WithPrefixIndex()); err != nil {
					return err
				}
			}
			return nil
		}, WithMetaOutput(mf))
	})
}

// Set manages a set of datums.
type Set struct {
	cacheClient                       *pfssync.CacheClient
	storageRoot                       string
	metaOutputClient, pfsOutputClient client.ModifyFile
	stats                             *Stats
}

// WithSet provides a scoped environment for a datum set.
func WithSet(cacheClient *pfssync.CacheClient, storageRoot string, cb func(*Set) error, opts ...SetOption) (retErr error) {
	s := &Set{
		cacheClient: cacheClient,
		storageRoot: storageRoot,
		stats:       &Stats{ProcessStats: &pps.ProcessStats{}},
	}
	for _, opt := range opts {
		opt(s)
	}
	if err := os.MkdirAll(storageRoot, 0777); err != nil {
		return errors.EnsureStack(err)
	}
	defer func() {
		if err := os.RemoveAll(storageRoot); retErr == nil {
			retErr = errors.EnsureStack(err)
		}
	}()
	return cb(s)
}

// UploadMeta uploads the meta file for a datum.
func (s *Set) UploadMeta(meta *Meta, opts ...Option) error {
	d := newDatum(s, meta, opts...)
	return d.uploadMetaFile(d.set.metaOutputClient)
}

// WithDatum provides a scoped environment for a datum within the datum set.
// TODO: Handle datum concurrency here, and potentially move symlinking here.
func (s *Set) WithDatum(meta *Meta, cb func(*Datum) error, opts ...Option) error {
	d := newDatum(s, meta, opts...)

	var err error
	for i := 0; i <= d.numRetries; i++ {
		err = d.withData(func() (retErr error) {
			defer func() {
				if retErr == nil || i == d.numRetries {
					retErr = d.finish(retErr)
				}
				duration := time.Duration(d.meta.Stats.ProcessTime.GetNanos()) + time.Duration(d.meta.Stats.ProcessTime.GetSeconds())*time.Second
				labels := workerStats.DatumLabels(d.meta.Job, d.meta.State.String())
				workerStats.DatumProcTime.With(labels).Observe(duration.Seconds())
				workerStats.DatumProcSecondsCount.With(labels).Add(duration.Seconds())
				workerStats.DatumCount.With(labels).Inc()
			}()
			return cb(d)
		})
		if err == nil {
			return nil
		}
	}
	return err
}

// Datum manages a datum.
type Datum struct {
	set              *Set
	ID               string
	meta             *Meta
	storageRoot      string
	numRetries       int
	recoveryCallback func(context.Context) error
	timeout          time.Duration
	IDPrefix         string
}

func newDatum(set *Set, meta *Meta, opts ...Option) *Datum {
	ID := common.DatumID(meta.Inputs)
	d := &Datum{
		set:         set,
		meta:        meta,
		ID:          ID,
		storageRoot: path.Join(set.storageRoot, ID),
		numRetries:  defaultNumRetries,
	}
	d.meta.Stats = &pps.ProcessStats{}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// PFSStorageRoot returns the pfs storage root.
func (d *Datum) PFSStorageRoot() string {
	return path.Join(d.storageRoot, common.PFSPrefix, d.ID)
}

// MetaStorageRoot returns the meta storage root.
func (d *Datum) MetaStorageRoot() string {
	return path.Join(d.storageRoot, common.MetaPrefix, d.ID)
}

func (d *Datum) finish(err error) (retErr error) {
	defer func() {
		if err := MergeProcessStats(d.set.stats.ProcessStats, d.meta.Stats); retErr == nil {
			retErr = err
		}
	}()
	if err != nil {
		d.handleFailed(err)
		return d.uploadMetaOutput()
	}
	d.set.stats.Processed++
	return d.uploadOutput()
}

func (d *Datum) handleFailed(err error) {
	if d.meta.State == State_RECOVERED {
		d.set.stats.Recovered++
		return
	}
	d.meta.State = State_FAILED
	d.meta.Reason = err.Error()
	d.set.stats.Failed++
	if d.set.stats.FailedID == "" {
		d.set.stats.FailedID = d.ID
	}
}

func (d *Datum) withData(cb func() error) (retErr error) {
	// Setup and defer cleanup of pfs directory.
	if err := os.MkdirAll(path.Join(d.PFSStorageRoot(), common.OutputPrefix), 0777); err != nil {
		return errors.EnsureStack(err)
	}
	defer func() {
		if err := os.RemoveAll(d.PFSStorageRoot()); retErr == nil {
			retErr = errors.EnsureStack(err)
		}
	}()
	return pfssync.WithDownloader(d.set.cacheClient, func(downloader pfssync.Downloader) error {
		// TODO: Move to copy file for inputs to datum file set.
		if err := d.downloadData(downloader); err != nil {
			return err
		}
		return cb()
	})
}

func (d *Datum) downloadData(downloader pfssync.Downloader) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		d.meta.Stats.DownloadTime = types.DurationProto(duration)
		labels := workerStats.JobLabels(d.meta.Job)
		workerStats.DatumDownloadSize.With(labels).Observe(float64(d.meta.Stats.DownloadBytes))
		workerStats.DatumDownloadBytesCount.With(labels).Add(float64(d.meta.Stats.DownloadBytes))
		workerStats.DatumDownloadTime.With(labels).Observe(duration.Seconds())
		workerStats.DatumDownloadSecondsCount.With(labels).Add(duration.Seconds())
	}()
	d.meta.Stats.DownloadBytes = 0
	var mu sync.Mutex
	for _, input := range d.meta.Inputs {
		// TODO: Need some validation to catch lazy & empty since they are incompatible.
		// Probably should catch this at the input validation during pipeline creation?
		opts := []pfssync.DownloadOption{
			pfssync.WithHeaderCallback(func(hdr *tar.Header) error {
				mu.Lock()
				defer mu.Unlock()
				d.meta.Stats.DownloadBytes += hdr.Size
				return nil
			}),
		}
		if input.Lazy {
			opts = append(opts, pfssync.WithLazy())
		}
		if input.EmptyFiles {
			opts = append(opts, pfssync.WithEmpty())
		}
		if err := downloader.Download(path.Join(d.PFSStorageRoot(), input.Name), input.FileInfo.File, opts...); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}

// Run provides a scoped environment for the processing of a datum.
func (d *Datum) Run(ctx context.Context, cb func(ctx context.Context) error) error {
	start := time.Now()
	defer func() {
		d.meta.Stats.ProcessTime = types.DurationProto(time.Since(start))
	}()
	if d.timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, d.timeout)
		defer cancel()
		return d.run(timeoutCtx, cb)
	}
	return d.run(ctx, cb)
}

func (d *Datum) run(ctx context.Context, cb func(ctx context.Context) error) (retErr error) {
	defer func() {
		if retErr != nil {
			if d.recoveryCallback != nil {
				// TODO: Set error based on recovery or original? Going with original for now.
				err := d.recoveryCallback(ctx)
				if err == nil {
					d.meta.State = State_RECOVERED
				}
			}
		}
	}()
	return cb(ctx)
}

func (d *Datum) uploadMetaOutput() (retErr error) {
	if d.set.metaOutputClient != nil {
		// Setup and defer cleanup of meta directory.
		if err := os.MkdirAll(d.MetaStorageRoot(), 0777); err != nil {
			return errors.EnsureStack(err)
		}
		defer func() {
			if err := os.RemoveAll(d.MetaStorageRoot()); retErr == nil {
				retErr = errors.EnsureStack(err)
			}
		}()
		if err := d.uploadMetaFile(d.set.metaOutputClient); err != nil {
			return err
		}
		return d.upload(d.set.metaOutputClient, d.storageRoot)
	}
	return nil
}

func (d *Datum) uploadMetaFile(mf client.ModifyFile) error {
	marshaler := &jsonpb.Marshaler{}
	buf := &bytes.Buffer{}
	if err := marshaler.Marshal(buf, d.meta); err != nil {
		return errors.EnsureStack(err)
	}
	fullPath := common.MetaFilePath(d.IDPrefix + d.ID)
	err := mf.PutFile(fullPath, buf, client.WithAppendPutFile(), client.WithDatumPutFile(d.ID))
	return errors.EnsureStack(err)
}

func (d *Datum) uploadOutput() error {
	if d.set.pfsOutputClient != nil {
		start := time.Now()
		d.meta.Stats.UploadBytes = 0
		if err := d.upload(d.set.pfsOutputClient, path.Join(d.PFSStorageRoot(), common.OutputPrefix), func(hdr *tar.Header) error {
			d.meta.Stats.UploadBytes += hdr.Size
			return nil
		}); err != nil {
			return err
		}
		// TODO: stats should probably include meta upload as well
		duration := time.Since(start)
		d.meta.Stats.UploadTime = types.DurationProto(duration)
		labels := workerStats.JobLabels(d.meta.Job)
		workerStats.DatumUploadSize.With(labels).Observe(float64(d.meta.Stats.UploadBytes))
		workerStats.DatumUploadBytesCount.With(labels).Add(float64(d.meta.Stats.UploadBytes))
		workerStats.DatumUploadTime.With(labels).Observe(duration.Seconds())
		workerStats.DatumUploadSecondsCount.With(labels).Add(duration.Seconds())
	}
	return d.uploadMetaOutput()
}

func (d *Datum) upload(mf client.ModifyFile, storageRoot string, cb ...func(*tar.Header) error) (retErr error) {
	if err := miscutil.WithPipe(func(w io.Writer) (retErr error) {
		bufW := bufio.NewWriterSize(w, grpcutil.MaxMsgPayloadSize)
		defer func() {
			if err := bufW.Flush(); retErr == nil {
				retErr = err
			}
		}()
		var opts []tarutil.ExportOption
		if len(cb) > 0 {
			opts = append(opts, tarutil.WithHeaderCallback(cb[0]))
		}
		return tarutil.Export(storageRoot, bufW, opts...)
	}, func(r io.Reader) error {
		err := mf.PutFileTAR(r, client.WithAppendPutFile(), client.WithDatumPutFile(d.ID))
		return errors.EnsureStack(err)
	}); err != nil {
		return err
	}
	return d.handleSymlinks(mf, storageRoot)
}

func (d *Datum) handleSymlinks(mf client.ModifyFile, storageRoot string) error {
	err := filepath.Walk(storageRoot, func(file string, fi os.FileInfo, err error) (retErr error) {
		if err != nil {
			return err
		}
		if file == storageRoot {
			return nil
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			return nil
		}
		dstPath, err := filepath.Rel(storageRoot, file)
		if err != nil {
			return errors.EnsureStack(err)
		}
		file, err = os.Readlink(file)
		if err != nil {
			return errors.EnsureStack(err)
		}
		fi, err = os.Stat(file)
		if err != nil {
			return errors.EnsureStack(err)
		}
		// Upload the local files if they are not from PFS.
		if !strings.HasPrefix(file, d.PFSStorageRoot()) {
			return d.uploadSymlink(mf, dstPath, file, fi)
		}
		relPath, err := filepath.Rel(d.PFSStorageRoot(), file)
		if err != nil {
			return errors.EnsureStack(err)
		}
		pathSplit := strings.Split(relPath, string(os.PathSeparator))
		var input *common.Input
		for _, i := range d.meta.Inputs {
			if i.Name == pathSplit[0] {
				input = i
			}
		}
		// Upload the local files if they are not using the empty or lazy files feature.
		if !(input.EmptyFiles || input.Lazy) {
			return d.uploadSymlink(mf, dstPath, file, fi)
		}
		srcFile := proto.Clone(input.FileInfo.File).(*pfs.File)
		srcFile.Path = path.Join(pathSplit[1:]...)
		return errors.EnsureStack(mf.CopyFile(dstPath, srcFile, client.WithDatumCopyFile(d.ID)))
	})
	return errors.EnsureStack(err)
}

func (d *Datum) uploadSymlink(mf client.ModifyFile, dstPath, file string, fi os.FileInfo) error {
	cb := func(dstPath, file string) (retErr error) {
		f, err := os.Open(file)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer func() {
			if err := f.Close(); retErr == nil {
				retErr = err
			}
		}()
		return errors.EnsureStack(mf.PutFile(dstPath, f, client.WithDatumPutFile(d.ID)))
	}
	if fi.IsDir() {
		dir := file
		err := filepath.Walk(dir, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return nil
			}
			dstSubpath, err := filepath.Rel(dir, file)
			if err != nil {
				return errors.EnsureStack(err)
			}
			return cb(path.Join(dstPath, dstSubpath), file)
		})
		return errors.EnsureStack(err)
	}
	return cb(dstPath, file)
}

// TODO: I think these types would be unecessary if the dependencies were shuffled around a bit.
type fileWalkerFunc func(string) ([]string, error)

// Deleter deletes a datum.
type Deleter func(*Meta) error

// NewDeleter creates a new deleter.
func NewDeleter(metaFileWalker fileWalkerFunc, metaOutputClient, pfsOutputClient client.ModifyFile) Deleter {
	return func(meta *Meta) error {
		ID := common.DatumID(meta.Inputs)
		tagOption := client.WithDatumDeleteFile(ID)
		// Delete the datum's meta file from the meta commit.
		if err := metaOutputClient.DeleteFile(common.MetaFilePath(ID), tagOption); err != nil {
			return errors.EnsureStack(err)
		}
		pfsDir := "/" + path.Join(common.PFSPrefix, ID)
		outDir := path.Join(pfsDir, common.OutputPrefix)
		files, err := metaFileWalker(pfsDir)
		if err != nil {
			if pfsserver.IsFileNotFoundErr(err) {
				return nil
			}
			return err
		}
		for i := range files {
			if err := metaOutputClient.DeleteFile(files[i], tagOption); err != nil {
				return errors.EnsureStack(err)
			}
			// Delete the content output by the datum.
			if strings.HasPrefix(files[i], outDir) {
				// Remove the output directory prefix.
				file, err := filepath.Rel(outDir, files[i])
				if err != nil {
					return errors.EnsureStack(err)
				}
				if err := pfsOutputClient.DeleteFile(file, tagOption); err != nil {
					return errors.EnsureStack(err)
				}
			}
		}
		return nil
	}
}
