package server

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	workerpkg "github.com/pachyderm/pachyderm/src/server/pkg/worker"

	"golang.org/x/net/context"
)

type datumFactory interface {
	Next() []*workerpkg.Input
	Len() int
}

type atomDatumFactory struct {
	datumList []*workerpkg.Input
	index     int
}

func newAtomDatumFactory(ctx context.Context, pfsClient pfs.APIClient, input *pps.AtomInput) (datumFactory, error) {
	result := &atomDatumFactory{}
	fileInfos, err := pfsClient.GlobFile(ctx, &pfs.GlobFileRequest{
		Commit:  input.Commit,
		Pattern: input.Glob,
	})
	if err != nil {
		return nil, err
	}
	for _, fileInfo := range fileInfos.FileInfo {
		result.datumList = append(result.datumList, &workerpkg.Input{
			FileInfo: fileInfo,
			Name:     input.Name,
			Lazy:     input.Lazy,
		})
	}
	return result, nil
}

func (d *atomDatumFactory) Next() []*workerpkg.Input {
	if d.index > len(d.datumList) {
		return nil
	}
	defer func() { d.index++ }()
	return []*workerpkg.Input{d.datumList[d.index]}
}

func (d *atomDatumFactory) Len() int {
	return len(d.datumList)
}

type crossDatumFactory struct {
	indexes    []int
	datumLists [][]*pfs.FileInfo
	done       bool
}

func (d *crossDatumFactory) Next() []*workerpkg.Input {
	return nil
	// if d.done {
	// 	return nil
	// }

	// defer func() {
	// 	// Increment the indexes
	// 	for i := 0; i < len(d.datumLists); i++ {
	// 		if d.indexes[i] == len(d.datumLists[i])-1 {
	// 			d.indexes[i] = 0
	// 			continue
	// 		}
	// 		d.indexes[i]++
	// 		return
	// 	}
	// 	d.done = true
	// }()

	// var datum []*pfs.FileInfo
	// for i, index := range d.indexes {
	// 	datum = append(datum, d.datumLists[i][index])
	// }
	// return datum
}

func (d *crossDatumFactory) Len() int {
	if len(d.datumLists) == 0 {
		return 0
	}
	result := len(d.datumLists[0])
	for i := 1; i < len(d.datumLists); i++ {
		result *= len(d.datumLists[i])
	}
	return result
}

func (d *crossDatumFactory) Datum(i int) []*workerpkg.Input {
	return nil
}

func newCrossDatumFactory(ctx context.Context, pfsClient pfs.APIClient, input *pps.CrossInput) (datumFactory, error) {
	df := &crossDatumFactory{}
	// for _, input := range inputs {
	// 	fileInfos, err := pfsClient.GlobFile(ctx, &pfs.GlobFileRequest{
	// 		Commit:  input.Commit,
	// 		Pattern: input.Glob,
	// 	})
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if len(fileInfos.FileInfo) > 0 {
	// 		df.datumLists = append(df.datumLists, fileInfos.FileInfo)
	// 		df.indexes = append(df.indexes, 0)
	// 	} else {
	// 		// If any input is empty, we don't return any datums
	// 		df.done = true
	// 		break
	// 	}
	// }
	return df, nil
}

func newDatumFactory(ctx context.Context, pfsClient pfs.APIClient, input *pps.Input) (datumFactory, error) {
	return nil, nil
}
