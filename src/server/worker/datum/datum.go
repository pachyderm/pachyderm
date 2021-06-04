package datum

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	io "io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfssync"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
)

const (
	// MetaPrefix is the prefix for the meta path.
	MetaPrefix = "meta"
	// MetaFileName is the name of the meta file.
	MetaFileName = "meta"
	// PFSPrefix is the prefix for the pfs path.
	PFSPrefix = "pfs"
	// OutputPrefix is the prefix for the output path.
	OutputPrefix = "out"
	// TmpFileName is the name of the tmp file.
	TmpFileName       = "tmp"
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
		return err
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
	pachClient                        *client.APIClient
	storageRoot                       string
	metaOutputClient, pfsOutputClient client.ModifyFile
	stats                             *Stats
}

// WithSet provides a scoped environment for a datum set.
func WithSet(pachClient *client.APIClient, storageRoot string, cb func(*Set) error, opts ...SetOption) (retErr error) {
	s := &Set{
		pachClient:  pachClient,
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
func (s *Set) WithDatum(ctx context.Context, meta *Meta, cb func(*Datum) error, opts ...Option) error {
	d := newDatum(s, meta, opts...)
	cancelCtx, cancel := context.WithCancel(ctx)
	attemptsLeft := d.numRetries + 1
	return backoff.RetryUntilCancel(cancelCtx, func() error {
		return d.withData(func() (retErr error) {
			defer func() {
				attemptsLeft--
				if retErr == nil || attemptsLeft == 0 {
					retErr = d.finish(retErr)
					cancel()
				}
			}()
			return cb(d)
		})
	}, &backoff.ZeroBackOff{}, func(err error, _ time.Duration) error {
		// TODO: Tagged logger here?
		fmt.Println("withDatum:", err)
		return nil
	})
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
	return path.Join(d.storageRoot, PFSPrefix, d.ID)
}

// MetaStorageRoot returns the meta storage root.
func (d *Datum) MetaStorageRoot() string {
	return path.Join(d.storageRoot, MetaPrefix, d.ID)
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
	if err := os.MkdirAll(path.Join(d.PFSStorageRoot(), OutputPrefix), 0777); err != nil {
		return errors.EnsureStack(err)
	}
	defer func() {
		if err := os.RemoveAll(d.PFSStorageRoot()); retErr == nil {
			retErr = errors.EnsureStack(err)
		}
	}()
	return pfssync.WithDownloader(d.set.pachClient, func(downloader pfssync.Downloader) error {
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
		d.meta.Stats.DownloadTime = types.DurationProto(time.Since(start))
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
				d.meta.Stats.DownloadBytes += uint64(hdr.Size)
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
			return err
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
		return err
	}
	fullPath := path.Join(MetaPrefix, d.ID, MetaFileName)
	return mf.PutFile(fullPath, buf, client.WithAppendPutFile(), client.WithTagPutFile(d.ID))
}

func (d *Datum) uploadOutput() error {
	if d.set.pfsOutputClient != nil {
		start := time.Now()
		d.meta.Stats.UploadBytes = 0
		if err := d.upload(d.set.pfsOutputClient, path.Join(d.PFSStorageRoot(), OutputPrefix), func(hdr *tar.Header) error {
			d.meta.Stats.UploadBytes += uint64(hdr.Size)
			return nil
		}); err != nil {
			return err
		}
		d.meta.Stats.UploadTime = types.DurationProto(time.Since(start))
	}
	return d.uploadMetaOutput()
}

func (d *Datum) upload(mf client.ModifyFile, storageRoot string, cb ...func(*tar.Header) error) (retErr error) {
	return obj.WithPipe(func(w io.Writer) (retErr error) {
		bufW := bufio.NewWriterSize(w, grpcutil.MaxMsgPayloadSize)
		defer func() {
			if err := bufW.Flush(); retErr == nil {
				retErr = err
			}
		}()
		opts := []tarutil.ExportOption{
			tarutil.WithSymlinkCallback(func(dst, src string, copyFunc func() error) error {
				return d.handleSymlink(mf, dst, src, copyFunc)
			}),
		}
		if len(cb) > 0 {
			opts = append(opts, tarutil.WithHeaderCallback(cb[0]))
		}
		return tarutil.Export(storageRoot, bufW, opts...)
	}, func(r io.Reader) error {
		return mf.PutFileTar(r, client.WithAppendPutFile(), client.WithTagPutFile(d.ID))
	})
}

func (d *Datum) handleSymlink(mf client.ModifyFile, dst, src string, copyFunc func() error) error {
	if !strings.HasPrefix(src, d.PFSStorageRoot()) {
		return copyFunc()
	}
	relPath, err := filepath.Rel(d.PFSStorageRoot(), src)
	if err != nil {
		return err
	}
	pathSplit := strings.Split(relPath, string(os.PathSeparator))
	var input *common.Input
	for _, i := range d.meta.Inputs {
		if i.Name == pathSplit[0] {
			input = i
		}
	}
	srcFile := input.FileInfo.File
	srcFile.Path = path.Join(pathSplit[1:]...)
	return mf.CopyFile(dst, srcFile, client.WithTagCopyFile(d.ID))
}

// TODO: I think these types would be unecessary if the dependencies were shuffled around a bit.
type fileWalkerFunc func(string) ([]string, error)

// Deleter deletes a datum.
type Deleter func(*Meta) error

// NewDeleter creates a new deleter.
func NewDeleter(metaFileWalker fileWalkerFunc, metaOutputClient, pfsOutputClient client.ModifyFile) Deleter {
	return func(meta *Meta) error {
		ID := common.DatumID(meta.Inputs)
		// Delete the datum directory in the meta output.
		if err := metaOutputClient.DeleteFile(path.Join(MetaPrefix, ID) + "/"); err != nil {
			return err
		}
		if err := metaOutputClient.DeleteFile(path.Join(PFSPrefix, ID) + "/"); err != nil {
			return err
		}
		// Delete the content output by the datum.
		outputDir := "/" + path.Join(PFSPrefix, ID, OutputPrefix)
		files, err := metaFileWalker(outputDir)
		if err != nil {
			if pfsserver.IsFileNotFoundErr(err) {
				return nil
			}
			return err
		}
		for i := range files {
			// Remove the output directory prefix.
			file, err := filepath.Rel(outputDir, files[i])
			if err != nil {
				return err
			}
			if err := pfsOutputClient.DeleteFile(file, client.WithTagDeleteFile(ID)); err != nil {
				return err
			}
		}
		return nil
	}
}
