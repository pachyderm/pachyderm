package datum

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

const (
	MetaPrefix        = "meta"
	MetaFileName      = "meta"
	PFSPrefix         = "pfs"
	OutputPrefix      = "out"
	TmpFileName       = "tmp"
	defaultNumRetries = 0
)

type PutTarClient interface {
	PutTar(io.Reader, bool, ...string) error
}

const defaultDatumsPerSet = 10

type SetSpec struct {
	Number int
}

func CreateSets(dit IteratorV2, storageRoot string, setSpec *SetSpec, upload func(func(PutTarClient) error) error) error {
	var metas []*Meta
	datumsPerSet := defaultDatumsPerSet
	if setSpec != nil {
		datumsPerSet = setSpec.Number
	}
	if err := dit.Iterate(func(meta *Meta) error {
		metas = append(metas, meta)
		if len(metas) >= datumsPerSet {
			if err := createSet(metas, storageRoot, upload); err != nil {
				return err
			}
			metas = nil
		}
		return nil
	}); err != nil {
		return err
	}
	return createSet(metas, storageRoot, upload)
}

func createSet(metas []*Meta, storageRoot string, upload func(func(PutTarClient) error) error) error {
	return upload(func(ptc PutTarClient) error {
		return WithSet(nil, storageRoot, func(s *Set) error {
			for _, meta := range metas {
				d := newDatum(s, meta)
				if err := d.uploadMetaOutput(); err != nil {
					return err
				}
			}
			return nil
		}, WithMetaOutput(ptc))
	})
}

type Set struct {
	pachClient                        *client.APIClient
	storageRoot                       string
	metaOutputClient, pfsOutputClient PutTarClient
	stats                             *Stats
}

func WithSet(pachClient *client.APIClient, storageRoot string, cb func(*Set) error, opts ...SetOption) (retErr error) {
	s := &Set{
		pachClient:  pachClient,
		storageRoot: storageRoot,
		stats:       &Stats{ProcessStats: &pps.ProcessStats{}},
	}
	for _, opt := range opts {
		opt(s)
	}
	if err := os.MkdirAll(storageRoot, 0700); err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(storageRoot); retErr == nil {
			retErr = err
		}
	}()
	return cb(s)
}

// TODO: Handle datum concurrency here, and potentially move symlinking here.
func (s *Set) WithDatum(ctx context.Context, meta *Meta, cb func(*Datum) error, opts ...DatumOption) error {
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

type Datum struct {
	set              *Set
	ID               string
	meta             *Meta
	storageRoot      string
	numRetries       int
	recoveryCallback func(context.Context) error
	timeout          time.Duration
}

func newDatum(set *Set, meta *Meta, opts ...DatumOption) *Datum {
	// TODO: ID needs more discussion.
	ID := common.DatumIDV2(meta.Inputs)
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

func (d *Datum) PFSStorageRoot() string {
	return path.Join(d.storageRoot, PFSPrefix, d.ID)
}

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
	if err := os.MkdirAll(path.Join(d.PFSStorageRoot(), OutputPrefix), 0700); err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(d.PFSStorageRoot()); retErr == nil {
			retErr = err
		}
	}()
	// Download input files.
	// TODO: Move to copy file for inputs to datum file set.
	if err := d.downloadData(); err != nil {
		return err
	}
	return cb()
}

func (d *Datum) downloadData() error {
	start := time.Now()
	d.meta.Stats.DownloadBytes = 0
	defer func() {
		d.meta.Stats.DownloadTime = types.DurationProto(time.Since(start))
	}()
	for _, input := range d.meta.Inputs {
		if err := sync.PullV2(d.set.pachClient, input.FileInfo.File, path.Join(d.PFSStorageRoot(), input.Name), func(hdr *tar.Header) error {
			d.meta.Stats.DownloadBytes += uint64(hdr.Size)
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

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
		if err := os.MkdirAll(d.MetaStorageRoot(), 0700); err != nil {
			return err
		}
		defer func() {
			if err := os.RemoveAll(d.MetaStorageRoot()); retErr == nil {
				retErr = err
			}
		}()
		marshaler := &jsonpb.Marshaler{}
		buf := &bytes.Buffer{}
		if err := marshaler.Marshal(buf, d.meta); err != nil {
			return err
		}
		fullPath := path.Join(d.MetaStorageRoot(), MetaFileName)
		if err := ioutil.WriteFile(fullPath, buf.Bytes(), 0700); err != nil {
			return err
		}
		return d.upload(d.set.metaOutputClient, d.storageRoot)
	}
	return nil
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

func (d *Datum) upload(ptc PutTarClient, storageRoot string, cb ...func(*tar.Header) error) error {
	// TODO: Might make more sense to convert to tar on the fly.
	f, err := os.Create(path.Join(d.set.storageRoot, TmpFileName))
	if err != nil {
		return err
	}
	if err := tarutil.Export(storageRoot, f, cb...); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	return ptc.PutTar(f, false, d.ID)
}

// TODO: I think these types would be unecessary if the dependencies were shuffled around a bit.
type fileWalkerFunc func(string) ([]string, error)

type DeleteClient interface {
	DeleteFiles([]string, ...string) error
}

type Deleter func(*Meta) error

func NewDeleter(metaFileWalker fileWalkerFunc, metaOutputClient, pfsOutputClient DeleteClient) Deleter {
	return func(meta *Meta) error {
		ID := common.DatumIDV2(meta.Inputs)
		// Delete the datum directory in the meta output.
		if err := metaOutputClient.DeleteFiles([]string{path.Join(MetaPrefix, ID) + "/"}); err != nil {
			return err
		}
		if err := metaOutputClient.DeleteFiles([]string{path.Join(PFSPrefix, ID) + "/"}); err != nil {
			return err
		}
		// Delete the content output by the datum.
		outputDir := "/" + path.Join(PFSPrefix, ID, OutputPrefix)
		files, err := metaFileWalker(outputDir)
		if err != nil {
			return err
		}
		// Remove the output directory prefix.
		for i := range files {
			var err error
			files[i], err = filepath.Rel(outputDir, files[i])
			if err != nil {
				return err
			}
		}
		return pfsOutputClient.DeleteFiles(files, ID)
	}
}
