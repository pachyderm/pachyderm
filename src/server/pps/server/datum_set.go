package server

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	workerpkg "github.com/pachyderm/pachyderm/src/server/pkg/worker"

	"golang.org/x/net/context"
)

type datumFactory interface {
	Next() []*workerpkg.Input
	Len() int
	Datum(i int) []*workerpkg.Input
}

type atomDatumFactory struct {
	inputs []*workerpkg.Input
	index  int
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
		result.inputs = append(result.inputs, &workerpkg.Input{
			FileInfo: fileInfo,
			Name:     input.Name,
			Lazy:     input.Lazy,
		})
	}
	return result, nil
}

func (d *atomDatumFactory) Next() []*workerpkg.Input {
	if d.index > len(d.inputs) {
		return nil
	}
	defer func() { d.index++ }()
	return []*workerpkg.Input{d.inputs[d.index]}
}

func (d *atomDatumFactory) Len() int {
	return len(d.inputs)
}

func (d *atomDatumFactory) Datum(i int) []*workerpkg.Input {
	return []*workerpkg.Input{d.inputs[i]}
}

type unionDatumFactory struct {
	inputs []datumFactory
}

func newUnionDatumFactory(ctx context.Context, pfsClient pfs.APIClient, unionInput *pps.UnionInput) (datumFactory, error) {
	result := &unionDatumFactory{}
	for _, input := range unionInput.Input {
		datumFactory, err := newDatumFactory(ctx, pfsClient, input)
		if err != nil {
			return nil, err
		}
		result.inputs = append(result.inputs, datumFactory)
	}
	return result, nil
}

func (d *unionDatumFactory) Next() []*workerpkg.Input {
	return nil
}

func (d *unionDatumFactory) Len() int {
	result := 0
	for _, datumFactory := range d.inputs {
		result += datumFactory.Len()
	}
	return result
}

func (d *unionDatumFactory) Datum(i int) []*workerpkg.Input {
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

func (d *crossDatumFactory) Next() []*workerpkg.Input {
	return nil
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

func (d *crossDatumFactory) Datum(i int) []*workerpkg.Input {
	if i >= d.Len() {
		panic("index out of bounds")
	}
	var result []*workerpkg.Input
	for _, datumFactory := range d.inputs {
		result = append(result, datumFactory.Datum(i%datumFactory.Len())...)
		i /= datumFactory.Len()
	}
	return result
}

func newCrossDatumFactory(ctx context.Context, pfsClient pfs.APIClient, crossInput *pps.CrossInput) (datumFactory, error) {
	result := &crossDatumFactory{}
	for _, input := range crossInput.Input {
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
