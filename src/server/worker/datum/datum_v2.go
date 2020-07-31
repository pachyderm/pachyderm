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
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/sync"
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

// TODO:
// Implement chunk spec configuration (hardcoded just for testing).
func CreateSets(dit IteratorV2, storageRoot string, upload func(func(PutTarClient) error) error) error {
	var metas []*Meta
	if err := dit.Iterate(func(meta *Meta) error {
		metas = append(metas, meta)
		// TODO: Hardcoded just for testing, datum set creation configuration will be applied here.
		if len(metas) > 10 {
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
}

func WithSet(pachClient *client.APIClient, storageRoot string, cb func(*Set) error, opts ...SetOption) (retErr error) {
	s := &Set{
		pachClient:  pachClient,
		storageRoot: storageRoot,
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
	attemptsLeft := d.numRetries + 1
	cancelCtx, cancel := context.WithCancel(ctx)
	return backoff.RetryUntilCancel(cancelCtx, func() error {
		return d.withData(func() error {
			if err := cb(d); err != nil {
				if attemptsLeft == 0 {
					cancel()
					return d.handleFailedDatum(err)
				}
				attemptsLeft--
				return err
			}
			return d.uploadOutput()
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

func (d *Datum) handleFailedDatum(err error) error {
	d.meta.State = State_FAILED
	if d.recoveryCallback != nil {
		err = d.recoveryCallback()
		if err == nil {
			d.meta.State = State_RECOVERED
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
	for _, input := range d.meta.Inputs {
		if err := sync.PullV2(d.set.pachClient, input.FileInfo.File, path.Join(d.PFSStorageRoot(), input.Name)); err != nil {
			return err
		}
	}
	return cb()
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
		if err := d.upload(d.set.pfsOutputClient, path.Join(d.PFSStorageRoot(), OutputPrefix)); err != nil {
			return err
		}
	}
	return nil
}

func (d *Datum) upload(ptc PutTarClient, storageRoot string) error {
	// TODO: Might make more sense to convert to tar on the fly.
	f, err := os.Create(path.Join(d.set.storageRoot, TmpFileName))
	if err != nil {
		return err
	}
	if err := tarutil.LocalToTar(storageRoot, f); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	return ptc.PutTar(f, d.ID)
}
