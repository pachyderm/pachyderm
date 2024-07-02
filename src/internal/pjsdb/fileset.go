package pjsdb

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

// validFileset forces the caller to validate filesets before they can be passed to the pjsdb API.
type validFileset struct {
	ID     fileset.FileSet
	handle string
	hash   []byte // optional
}

func validateFileset(ctx context.Context, s *fileset.Storage, filesetHandle string, shouldhash bool) (*validFileset, error) {
	fs := &validFileset{handle: filesetHandle}
	var err error
	fs.ID, err = openFileset(ctx, s, filesetHandle)
	if err != nil {
		return nil, errors.Join(&InvalidFilesetError{ID: filesetHandle, Reason: "empty handle"}, errors.Wrap(err, "validate fileset"))
	}
	if !shouldhash {
		return fs, nil
	}
	fs.hash, err = hashFileset(ctx, fs.ID)
	if err != nil {
		return nil, errors.Wrap(err, "validate filesets")
	}
	return fs, nil
}

func openFileset(ctx context.Context, s *fileset.Storage, filesetHandle string) (fileset.FileSet, error) {
	if filesetHandle == "" {
		return nil, FilesetHandleIsEmptyError
	}
	id, err := fileset.ParseID(filesetHandle)
	if err != nil {
		return nil, errors.Wrap(err, "open fileset: parsing fileset id")
	}
	fs, err := s.Open(ctx, []fileset.ID{*id})
	if err != nil {
		return nil, errors.Wrap(err, "open fileset: opening fileset")
	}
	return fs, nil
}

func hashFileset(ctx context.Context, fs fileset.FileSet) ([]byte, error) {
	hash := pachhash.New()
	if err := fs.Iterate(ctx, func(f fileset.File) error {
		fHash, err := f.Hash(ctx)
		if err != nil {
			return errors.Wrap(err, "hashing file "+f.Index().Path)
		}
		// naively assuming that this is better than xor'ing a bunch of hashes together.
		if _, err := hash.Write(fHash); err != nil {
			return errors.Wrap(err, "writing file hash"+f.Index().Path)
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "hash fileset: iterating over files")
	}
	return hash.Sum(nil), nil
}
