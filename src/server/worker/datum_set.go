package worker

import (
	"fmt"
	"sort"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	"golang.org/x/net/context"
)

type datumFactory interface {
	Len() int
	Datum(i int) []*Input
}

func newAtomDatumFactory(ctx context.Context, pfsClient pfs.APIClient, input *pps.AtomInput) (datumFactory, error) {
	var result atomDatumFactory
	fileInfos, err := pfsClient.GlobFile(ctx, &pfs.GlobFileRequest{
		Commit:  client.NewCommit(input.Repo, input.Commit),
		Pattern: input.Glob,
	})
	if err != nil {
		return nil, err
	}
	for _, fileInfo := range fileInfos.FileInfo {
		result = append(result, &Input{
			FileInfo: fileInfo,
			Name:     input.Name,
			Lazy:     input.Lazy,
			Branch:   input.Branch,
		})
	}
	sort.Sort(result)
	return result, nil
}

type atomDatumFactory []*Input

func (a atomDatumFactory) Len() int      { return len(a) }
func (a atomDatumFactory) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a atomDatumFactory) Less(i, j int) bool {
	if a[i].FileInfo.SizeBytes == a[j].FileInfo.SizeBytes {
		return a[i].FileInfo.File.Path < a[j].FileInfo.File.Path
	}
	return a[i].FileInfo.SizeBytes < a[j].FileInfo.SizeBytes
}

func (a atomDatumFactory) Datum(i int) []*Input {
	return []*Input{a[i]}
}

type unionDatumFactory struct {
	inputs []datumFactory
}

func newUnionDatumFactory(ctx context.Context, pfsClient pfs.APIClient, union []*pps.Input) (datumFactory, error) {
	result := &unionDatumFactory{}
	for _, input := range union {
		datumFactory, err := newDatumFactory(ctx, pfsClient, input)
		if err != nil {
			return nil, err
		}
		result.inputs = append(result.inputs, datumFactory)
	}
	return result, nil
}

func (d *unionDatumFactory) Len() int {
	result := 0
	for _, datumFactory := range d.inputs {
		result += datumFactory.Len()
	}
	return result
}

func (d *unionDatumFactory) Datum(i int) []*Input {
	for _, datumFactory := range d.inputs {
		if i < datumFactory.Len() {
			return datumFactory.Datum(i)
		}
		i -= datumFactory.Len()
	}
	panic("index out of bounds")
}

type crossDatumFactory struct {
	inputs []datumFactory
}

func (d *crossDatumFactory) Len() int {
	if len(d.inputs) == 0 {
		return 0
	}
	result := d.inputs[0].Len()
	for i := 1; i < len(d.inputs); i++ {
		result *= d.inputs[i].Len()
	}
	return result
}

func (d *crossDatumFactory) Datum(i int) []*Input {
	if i >= d.Len() {
		panic("index out of bounds")
	}
	var result []*Input
	for _, datumFactory := range d.inputs {
		result = append(result, datumFactory.Datum(i%datumFactory.Len())...)
		i /= datumFactory.Len()
	}
	return result
}

func newCrossDatumFactory(ctx context.Context, pfsClient pfs.APIClient, cross []*pps.Input) (datumFactory, error) {
	result := &crossDatumFactory{}
	for _, input := range cross {
		datumFactory, err := newDatumFactory(ctx, pfsClient, input)
		if err != nil {
			return nil, err
		}
		result.inputs = append(result.inputs, datumFactory)
	}
	return result, nil
}

func newDatumFactory(ctx context.Context, pfsClient pfs.APIClient, input *pps.Input) (datumFactory, error) {
	switch {
	case input.Atom != nil:
		return newAtomDatumFactory(ctx, pfsClient, input.Atom)
	case input.Union != nil:
		return newUnionDatumFactory(ctx, pfsClient, input.Union)
	case input.Cross != nil:
		return newCrossDatumFactory(ctx, pfsClient, input.Cross)
	}
	return nil, fmt.Errorf("unrecognized input type")
}
