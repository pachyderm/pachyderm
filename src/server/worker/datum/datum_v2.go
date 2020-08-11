package datum

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
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
	PutTar(io.Reader, ...string) error
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
	recoveryCallback func() error
	stats            *pps.ProcessStats
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
		stats:       &pps.ProcessStats{},
	}
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
		if err := MergeProcessStats(d.set.stats.ProcessStats, d.stats); retErr == nil {
			retErr = err
		}
	}()
	if err != nil {
		return d.handleFailed(err)
	}
	d.set.stats.Processed++
	return d.uploadOutput()
}

func (d *Datum) handleFailed(err error) error {
	d.meta.State = State_FAILED
	if d.recoveryCallback != nil {
		err = d.recoveryCallback()
		if err == nil {
			d.meta.State = State_RECOVERED
			d.set.stats.Recovered++
		}
	}
	if d.meta.State == State_FAILED {
		d.set.stats.Failed++
		if d.set.stats.FailedID == "" {
			d.set.stats.FailedID = d.ID
		}
	}
	// TODO: Store error in meta.
	return d.uploadMetaOutput()
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
	d.stats.DownloadBytes = 0
	defer func() {
		d.stats.DownloadTime = types.DurationProto(time.Since(start))
	}()
	for _, input := range d.meta.Inputs {
		if err := sync.PullV2(d.set.pachClient, input.FileInfo.File, path.Join(d.PFSStorageRoot(), input.Name), func(hdr *tar.Header) error {
			d.stats.DownloadBytes += uint64(hdr.Size)
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (d *Datum) Run(ctx context.Context, cb func(ctx context.Context) error) error {
	if d.timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, d.timeout)
		defer cancel()
		return cb(timeoutCtx)
	}
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
	if err := d.uploadMetaOutput(); err != nil {
		return err
	}
	if d.set.pfsOutputClient != nil {
		start := time.Now()
		d.stats.UploadBytes = 0
		defer func() {
			d.stats.UploadTime = types.DurationProto(time.Since(start))
		}()
		if err := d.upload(d.set.pfsOutputClient, path.Join(d.PFSStorageRoot(), OutputPrefix), func(hdr *tar.Header) error {
			d.stats.UploadBytes += uint64(hdr.Size)
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (d *Datum) upload(ptc PutTarClient, storageRoot string, cb ...func(*tar.Header) error) error {
	// TODO: Might make more sense to convert to tar on the fly.
	f, err := os.Create(path.Join(d.set.storageRoot, TmpFileName))
	if err != nil {
		return err
	}
	if err := tarutil.LocalToTar(storageRoot, f, cb...); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	return ptc.PutTar(f, d.ID)
}
