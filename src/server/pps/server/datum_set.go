package server

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	"golang.org/x/net/context"
)

type datumFactory interface {
	Next() []*pfs.FileInfo
	// Resets the internal indexes so we start reading from the first
	// datum set again.
	Reset()
	Indexes() []int
	Len() int
}

type datumFactoryImpl struct {
	indexes    []int
	datumLists [][]*pfs.FileInfo
	done       bool
}

func (d *datumFactoryImpl) Next() []*pfs.FileInfo {
	if d.done {
		return nil
	}

	defer func() {
		// Increment the indexes
		for i := 0; i < len(d.datumLists); i++ {
			if d.indexes[i] == len(d.datumLists[i])-1 {
				d.indexes[i] = 0
				continue
			}
			d.indexes[i]++
			return
		}
		d.done = true
	}()

	var datum []*pfs.FileInfo
	for i, index := range d.indexes {
		datum = append(datum, d.datumLists[i][index])
	}
	return datum
}

func (d *datumFactoryImpl) Indexes() []int {
	return d.indexes
}

func (d *datumFactoryImpl) Reset() {
	for i := range d.indexes {
		d.indexes[i] = 0
	}
}

func (d *datumFactoryImpl) Len() int {
	if len(d.datumLists) == 0 {
		return 0
	}
	result := len(d.datumLists[0])
	for i := 1; i < len(d.datumLists); i++ {
		result *= len(d.datumLists[i])
	}
	return result
}

func newDatumFactory(ctx context.Context, pfsClient pfs.APIClient, inputs []*pps.JobInput, indexes []int) (datumFactory, error) {
	df := &datumFactoryImpl{}
	for _, input := range inputs {
		fileInfos, err := pfsClient.GlobFile(ctx, &pfs.GlobFileRequest{
			Commit:  input.Commit,
			Pattern: input.Glob,
		})
		if err != nil {
			return nil, err
		}
		if len(fileInfos.FileInfo) > 0 {
			df.datumLists = append(df.datumLists, fileInfos.FileInfo)
			df.indexes = append(df.indexes, 0)
		} else {
			// If any input is empty, we don't return any datums
			df.done = true
			break
		}
	}
	if indexes != nil {
		df.indexes = indexes
	}
	return df, nil
}
