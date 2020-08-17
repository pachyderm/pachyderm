package server

import (
	"bytes"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

// Differ compares two sources and iterates over the items that are not equal.
type Differ struct {
	a, b Source
}

// NewDiffer creates an iterator over the differences between Sources a and b
func NewDiffer(a, b Source) *Differ {
	return &Differ{a: a, b: b}
}

// Iterate compares the entries from `a` and `b` path wise.
// If one side is missing a path, cb is called with the info for the side that has the path
// If both sides have a path, but the content is different, cb is called with the info for both sides at once.
// If both sides have a path, and the content is the same, cb is not called. The info is not part of the diff.
func (d *Differ) Iterate(ctx context.Context, cb func(aFi, bFi *pfs.FileInfo) error) error {
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	aInfos := make(chan *pfs.FileInfo)
	bInfos := make(chan *pfs.FileInfo)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer close(aInfos)
		return d.a.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case aInfos <- fi:
				return nil
			}
		})
	})
	eg.Go(func() error {
		defer close(bInfos)
		return d.b.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case bInfos <- fi:
				return nil
			}
		})
	})
	eg.Go(func() error {
		aFi, aOpen := <-aInfos
		bFi, bOpen := <-bInfos
		for aOpen && bOpen {
			switch {
			case aFi.File.Path < bFi.File.Path:
				if err := cb(aFi, nil); err != nil {
					return err
				}
				aFi, aOpen = <-aInfos
			case bFi.File.Path < aFi.File.Path:
				if err := cb(nil, bFi); err != nil {
					return err
				}
				bFi, bOpen = <-bInfos
			default:
				if !equalFileInfos(aFi, bFi) {
					if err := cb(aFi, bFi); err != nil {
						return err
					}
				}
				aFi, aOpen = <-aInfos
				bFi, bOpen = <-bInfos
			}
		}
		for ; aOpen; aFi, aOpen = <-aInfos {
			if err := cb(aFi, nil); err != nil {
				return err
			}
		}
		for ; bOpen; bFi, bOpen = <-bInfos {
			if err := cb(nil, bFi); err != nil {
				return err
			}
		}
		return nil
	})
	return eg.Wait()
}

func equalFileInfos(aFi, bFi *pfs.FileInfo) bool {
	return bytes.Equal(aFi.Hash, bFi.Hash)
}
