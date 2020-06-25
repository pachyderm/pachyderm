package datum

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/sync"
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

// TODO: This is copied from the chain package, needs to be refactored somewhere else.
type DatumHasher interface {
	// Hash should essentially wrap the common.HashDatum function, but other
	// implementations may be useful in tests.
	Hash([]*common.InputV2) string
}

type SetOption func(*Set)

func WithMetaOutput(ptc PutTarClient) SetOption {
	return func(s *Set) {
		s.metaOutputClient = ptc
	}
}

func WithPFSOutput(ptc PutTarClient) SetOption {
	return func(s *Set) {
		s.pfsOutputClient = ptc
	}
}

type Set struct {
	pachClient                        *client.APIClient
	hasher                            DatumHasher
	storageRoot                       string
	metaOutputClient, pfsOutputClient PutTarClient
}

func WithSet(pachClient *client.APIClient, storageRoot string, hasher DatumHasher, cb func(s *Set) error, opts ...SetOption) (retErr error) {
	s := &Set{
		pachClient:  pachClient,
		hasher:      hasher,
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
func (s *Set) WithDatum(ctx context.Context, inputs []*common.InputV2, cb func(d *Datum) error, opts ...DatumOption) (retErr error) {
	hash := s.hasher.Hash(inputs)
	d := &Datum{
		set: s,
		meta: &Meta{
			Inputs: inputs,
		},
		hash:            hash,
		metaStorageRoot: path.Join(s.storageRoot, MetaPrefix, hash),
		pfsStorageRoot:  path.Join(s.storageRoot, PFSPrefix, hash),
		numRetries:      defaultNumRetries,
	}
	for _, opt := range opts {
		opt(d)
	}
	attemptsLeft := d.numRetries + 1
	return backoff.RetryUntilCancel(ctx, func() error {
		return d.withData(func() error {
			if err := cb(d); err != nil {
				if attemptsLeft == 0 {
					return d.handleFailedDatum(err)
				}
				attemptsLeft--
				return err
			}
			return d.uploadOutput()
		})
	}, &backoff.ZeroBackOff{}, func(_ error, _ time.Duration) error {
		return nil
	})
}

type DatumOption func(*Datum)

func WithRetry(numRetries int) DatumOption {
	return func(d *Datum) {
		d.numRetries = numRetries
	}
}

func WithRecoveryCallback(cb func() error) DatumOption {
	return func(d *Datum) {
		d.recoveryCallback = cb
	}
}

type Datum struct {
	set  *Set
	meta *Meta
	// TODO: Create separate id from input paths for top level directory.
	hash                            string
	pfsStorageRoot, metaStorageRoot string
	numRetries                      int
	recoveryCallback                func() error
}

func (d *Datum) StorageRoot() string {
	return d.pfsStorageRoot
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
	// Setup and defer cleanup of datum directory.
	if err := os.MkdirAll(d.pfsStorageRoot, 0700); err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(d.pfsStorageRoot); retErr == nil {
			retErr = err
		}
	}()
	// Download input files.
	// TODO: Move to copy file for inputs to datum file set.
	for _, input := range d.meta.Inputs {
		if err := sync.PullV2(d.set.pachClient, input.FileInfo.File, path.Join(d.pfsStorageRoot, input.Name)); err != nil {
			return err
		}
	}
	return cb()
}

func (d *Datum) uploadMetaOutput() error {
	if d.set.metaOutputClient != nil {
		if err := d.uploadMetaFile(); err != nil {
			return err
		}
		return d.upload(d.set.metaOutputClient, d.pfsStorageRoot)
	}
	return nil
}

func (d *Datum) uploadMetaFile() (retErr error) {
	// Setup and defer cleanup of meta directory.
	if err := os.MkdirAll(d.metaStorageRoot, 0700); err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(d.metaStorageRoot); retErr == nil {
			retErr = err
		}
	}()
	marshaler := &jsonpb.Marshaler{}
	buf := &bytes.Buffer{}
	if err := marshaler.Marshal(buf, d.meta); err != nil {
		return err
	}
	fullPath := path.Join(d.metaStorageRoot, MetaFileName)
	if err := sync.WriteFile(fullPath, buf); err != nil {
		return err
	}
	return d.upload(d.set.metaOutputClient, d.metaStorageRoot)
}

func (d *Datum) uploadOutput() error {
	if err := d.uploadMetaOutput(); err != nil {
		return err
	}
	if d.set.pfsOutputClient != nil {
		if err := d.upload(d.set.pfsOutputClient, path.Join(d.pfsStorageRoot, OutputPrefix)); err != nil {
			return err
		}
	}
	return nil
}

func (d *Datum) upload(ptc PutTarClient, storageRoot string) error {
	// TODO: Might make more sense to convert to tar on the fly.
	f, err := os.Create(path.Join(storageRoot, TmpFileName))
	if err != nil {
		return err
	}
	if err := sync.LocalToTar(storageRoot, f); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	return ptc.PutTar(f, d.hash)
}
