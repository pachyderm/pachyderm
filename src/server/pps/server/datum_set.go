package server

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	"golang.org/x/net/context"
)

type datumSetFactory interface {
	Next() []*pfs.FileInfo
	// Resets the internal indexes so we start reading from the first
	// datum set again.
	Reset()
	Indexes() []int
}

type datumSetFactoryImpl struct {
	indexes    []int
	datumLists [][]*pfs.FileInfo
	done       bool
}

func (d *datumSetFactoryImpl) Next() []*pfs.FileInfo {
	if d.done {
		return nil
	}

	defer func() {
		// Increment the indexes
		for i := 0; i < len(d.datumLists); i++ {
			if d.indexes[i] == len(d.datumLists[i])-1 {
				continue
			}
			d.indexes[i]++
			return
		}
		d.done = true
	}()

	var datumSet []*pfs.FileInfo
	for i, index := range d.indexes {
		datumSet = append(datumSet, d.datumLists[i][index])
	}
	return datumSet
}

func (d *datumSetFactoryImpl) Indexes() []int {
	return d.indexes
}

func (d *datumSetFactoryImpl) Reset() {
	for i := range d.indexes {
		d.indexes[i] = 0
	}
}

func newDatumSetFactory(ctx context.Context, pfsClient pfs.APIClient, inputs []*pps.JobInput, indexes []int) (datumSetFactory, error) {
	dsf := &datumSetFactoryImpl{}
	for _, input := range inputs {
		fileInfos, err := pfsClient.GlobFile(ctx, &pfs.GlobFileRequest{
			Commit:  input.Commit,
			Pattern: input.Glob,
		})
		if err != nil {
			return nil, err
		}
		if len(fileInfos.FileInfo) > 0 {
			dsf.datumLists = append(dsf.datumLists, fileInfos.FileInfo)
			dsf.indexes = append(dsf.indexes, 0)
		}
	}
	if indexes != nil {
		dsf.indexes = indexes
	}
	return dsf, nil
}
