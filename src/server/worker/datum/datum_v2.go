package datum

import (
	"bytes"
	"io"
	"os"
	"path"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/sync"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

const (
	TmpFileName   = "tmp"
	InputFileName = "input"
	OutputPrefix  = "pfs/out"
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

func WithDatumOutput(ptc PutTarClient) SetOption {
	return func(s *Set) {
		s.datumOutputClient = ptc
	}
}

func WithRawOutput(ptc PutTarClient) SetOption {
	return func(s *Set) {
		s.rawOutputClient = ptc
	}
}

type Set struct {
	pachClient                         *client.APIClient
	hasher                             DatumHasher
	storageRoot                        string
	datumOutputClient, rawOutputClient PutTarClient
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

func (s *Set) WithDatum(inputs []*common.InputV2, cb func(d *Datum) error) (retErr error) {
	hash := s.hasher.Hash(inputs)
	d := &Datum{
		set:         s,
		inputs:      inputs,
		hash:        hash,
		storageRoot: path.Join(s.storageRoot, hash),
	}
	// Setup and defer cleanup of datum directory.
	if err := os.MkdirAll(d.storageRoot, 0700); err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(d.storageRoot); retErr == nil {
			retErr = err
		}
	}()
	// Download input files.
	for _, input := range inputs {
		if err := sync.PullV2(s.pachClient, input.FileInfo.File, path.Join(d.storageRoot, input.Name)); err != nil {
			return err
		}
	}
	// Defer upload of output.
	// TODO: Upload during error?
	defer func() {
		if err := d.uploadOutput(); retErr == nil {
			retErr = err
		}
	}()
	// TODO: Record failures.
	return cb(d)
}

type Datum struct {
	set    *Set
	inputs []*common.InputV2
	// TODO: Create separate id from input paths for top level directory.
	hash        string
	storageRoot string
}

func (d *Datum) StorageRoot() string {
	return d.storageRoot
}

func (d *Datum) Fail() {

}

func (d *Datum) Recover() {

}

func (d *Datum) uploadOutput() error {
	if d.set.datumOutputClient != nil {
		if err := d.serializeInputs(); err != nil {
			return err
		}
		if err := d.upload(d.set.datumOutputClient, d.storageRoot); err != nil {
			return err
		}
	}
	if d.set.rawOutputClient != nil {
		if err := d.upload(d.set.rawOutputClient, path.Join(d.storageRoot, OutputPrefix)); err != nil {
			return err
		}
	}
	return nil
}

func (d *Datum) serializeInputs() error {
	marshaler := &jsonpb.Marshaler{}
	buf := &bytes.Buffer{}
	for _, input := range d.inputs {
		if err := marshaler.Marshal(buf, input); err != nil {
			return err
		}
	}
	return d.writeFile(InputFileName, buf)
}

func (d *Datum) writeFile(fileName string, r io.Reader) error {
	fullPath := path.Join(d.storageRoot, fileName)
	return sync.WriteFile(fullPath, r)
}

func (d *Datum) upload(ptc PutTarClient, storageRoot string) error {
	// TODO: Might make more sense to convert to tar on the fly.
	f, err := os.Create(path.Join(d.storageRoot, TmpFileName))
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
