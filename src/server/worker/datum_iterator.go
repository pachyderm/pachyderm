package worker

import (
	"fmt"
	"io"
	"sort"

	glob "github.com/pachyderm/ohmyglob"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	"github.com/cevaris/ordered_map"
)

// DatumIterator is an interface which allows you to iterate through the datums
// for a job. A datum iterator keeps track of which datum it is on, which can be Reset()
// The intended use is by using this pattern `for di.Next() { ... datum := di.Datum() ... }`
// Note that since you start the loop by a call to Next(), the datum iterator's location starts at -1
type DatumIterator interface {
	Reset()
	Len() int
	Next() bool
	Datum() []*Input
	DatumN(int) []*Input
}

type pfsDatumIterator struct {
	inputs   []*Input
	location int
}

func newPFSDatumIterator(pachClient *client.APIClient, input *pps.PFSInput) (DatumIterator, error) {
	result := &pfsDatumIterator{}
	// make sure it gets initialized properly
	result.Reset()
	if input.Commit == "" {
		// this can happen if a pipeline with multiple inputs has been triggered
		// before all commits have inputs
		return result, nil
	}
	fs, err := pachClient.GlobFileStream(pachClient.Ctx(), &pfs.GlobFileRequest{
		Commit:  client.NewCommit(input.Repo, input.Commit),
		Pattern: input.Glob,
	})
	if err != nil {
		return nil, err
	}
	for {
		fileInfo, err := fs.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		g := glob.MustCompile(input.Glob, '/')
		joinOn := g.Replace(fileInfo.File.Path, input.JoinOn)
		result.inputs = append(result.inputs, &Input{
			FileInfo:   fileInfo,
			JoinOn:     joinOn,
			Name:       input.Name,
			Lazy:       input.Lazy,
			Branch:     input.Branch,
			EmptyFiles: input.EmptyFiles,
		})
	}
	// We sort the inputs so that the order is deterministic. Note that it's
	// not possible for 2 inputs to have the same path so this is guaranteed to
	// produce a deterministic order.
	sort.Slice(result.inputs, func(i, j int) bool {
		// We sort by descending size first because it can boost performance to
		// process the biggest datums first.
		if result.inputs[i].FileInfo.SizeBytes != result.inputs[j].FileInfo.SizeBytes {
			return result.inputs[i].FileInfo.SizeBytes > result.inputs[j].FileInfo.SizeBytes
		}
		return result.inputs[i].FileInfo.File.Path < result.inputs[j].FileInfo.File.Path
	})
	return result, nil
}

func (d *pfsDatumIterator) Reset() {
	d.location = -1
}

func (d *pfsDatumIterator) Len() int {
	return len(d.inputs)
}

func (d *pfsDatumIterator) Datum() []*Input {
	return []*Input{d.inputs[d.location]}
}

func (d *pfsDatumIterator) DatumN(n int) []*Input {
	return []*Input{d.inputs[n]}
}

func (d *pfsDatumIterator) Next() bool {
	d.location++
	return d.location < len(d.inputs)
}

type listDatumIterator struct {
	inputs   []*Input
	location int
}

func newListDatumIterator(pachClient *client.APIClient, inputs []*Input) (DatumIterator, error) {
	result := &listDatumIterator{}
	// make sure it gets initialized properly
	result.Reset()
	result.inputs = inputs
	return result, nil
}

func (d *listDatumIterator) Reset() {
	d.location = -1
}

func (d *listDatumIterator) Len() int {
	return len(d.inputs)
}

func (d *listDatumIterator) Datum() []*Input {
	return []*Input{d.inputs[d.location]}
}

func (d *listDatumIterator) DatumN(n int) []*Input {
	return []*Input{d.inputs[n]}
}

func (d *listDatumIterator) Next() bool {
	d.location++
	return d.location < len(d.inputs)
}

type unionDatumIterator struct {
	iterators []DatumIterator
	unionIdx  int
	location  int
}

func newUnionDatumIterator(pachClient *client.APIClient, union []*pps.Input) (DatumIterator, error) {
	result := &unionDatumIterator{}
	defer result.Reset()
	for _, input := range union {
		datumIterator, err := NewDatumIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		result.iterators = append(result.iterators, datumIterator)
	}
	return result, nil
}

func (d *unionDatumIterator) Reset() {
	for _, input := range d.iterators {
		input.Reset()
	}
	d.unionIdx = 0
	d.location = -1
}

func (d *unionDatumIterator) Len() int {
	result := 0
	for _, datumIterator := range d.iterators {
		result += datumIterator.Len()
	}
	return result
}

func (d *unionDatumIterator) Next() bool {
	if d.unionIdx >= len(d.iterators) {
		return false
	}
	if !d.iterators[d.unionIdx].Next() {
		d.unionIdx++
		return d.Next()
	}
	d.location++
	return true
}

func (d *unionDatumIterator) Datum() []*Input {
	return d.iterators[d.unionIdx].Datum()
}

func (d *unionDatumIterator) DatumN(n int) []*Input {
	for _, datumIterator := range d.iterators {
		if n < datumIterator.Len() {
			return datumIterator.DatumN(n)
		}
		n -= datumIterator.Len()
	}
	panic("index out of bounds")
}

type crossDatumIterator struct {
	iterators []DatumIterator
	started   bool
	location  int
}

func newCrossDatumIterator(pachClient *client.APIClient, cross []*pps.Input) (DatumIterator, error) {
	result := &crossDatumIterator{}
	defer result.Reset()
	for _, iterator := range cross {
		datumIterator, err := NewDatumIterator(pachClient, iterator)
		if err != nil {
			return nil, err
		}
		result.iterators = append(result.iterators, datumIterator)
	}
	result.location = -1
	return result, nil
}

func newCrossListDatumIterator(pachClient *client.APIClient, cross [][]*Input) (DatumIterator, error) {
	result := &crossDatumIterator{}
	defer result.Reset()
	for _, iterator := range cross {
		datumIterator, err := newListDatumIterator(pachClient, iterator)
		if err != nil {
			return nil, err
		}
		result.iterators = append(result.iterators, datumIterator)
	}
	result.location = -1
	return result, nil
}

func (d *crossDatumIterator) Reset() {
	inhabited := len(d.iterators) > 0
	for _, iterators := range d.iterators {
		iterators.Reset()
		if !iterators.Next() {
			inhabited = false
		}
	}
	if !inhabited {
		d.iterators = nil
	}
	d.location = -1
	d.started = !inhabited
}

func (d *crossDatumIterator) Len() int {
	if len(d.iterators) == 0 {
		return 0
	}
	result := d.iterators[0].Len()
	for i := 1; i < len(d.iterators); i++ {
		result *= d.iterators[i].Len()
	}
	return result
}

func (d *crossDatumIterator) Next() bool {
	if !d.started {
		d.started = true
		d.location++
		return true
	}
	for _, input := range d.iterators {
		// if we're at the end of the "row"
		if !input.Next() {
			// we reset the "row"
			input.Reset()
			// and start it back up
			input.Next()
			// after resetting this "row", start iterating through the next "row"
		} else {
			d.location++
			return true
		}
	}
	return false
}

func (d *crossDatumIterator) Datum() []*Input {
	var result []*Input
	for _, datumIterator := range d.iterators {
		result = append(result, datumIterator.Datum()...)
	}
	sortInputs(result)
	return result
}

func (d *crossDatumIterator) DatumN(n int) []*Input {
	if n >= d.Len() {
		panic("index out of bounds")
	}
	var result []*Input
	for _, datumIterator := range d.iterators {
		result = append(result, datumIterator.DatumN(n%datumIterator.Len())...)
		n /= datumIterator.Len()
	}
	sortInputs(result)
	return result
}

type joinDatumIterator struct {
	datums   [][]*Input
	location int
}

func newJoinDatumIterator(pachClient *client.APIClient, join []*pps.Input) (DatumIterator, error) {
	result := &joinDatumIterator{}
	om := ordered_map.NewOrderedMap()

	for i, input := range join {
		datumIterator, err := NewDatumIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		for datumIterator.Next() {
			x := datumIterator.Datum()
			for _, k := range x {
				tupleI, ok := om.Get(k.JoinOn)
				var tuple [][]*Input
				if !ok {
					tuple = make([][]*Input, len(join))
				} else {
					tuple = tupleI.([][]*Input)
				}
				tuple[i] = append(tuple[i], k)
				om.Set(k.JoinOn, tuple)
			}
		}
	}

	iter := om.IterFunc()
	for kv, ok := iter(); ok; kv, ok = iter() {
		tuple := kv.Value.([][]*Input)
		cross, err := newCrossListDatumIterator(pachClient, tuple)
		if err != nil {
			return nil, err
		}
		for cross.Next() {
			result.datums = append(result.datums, cross.Datum())
		}
	}
	result.location = -1
	return result, nil
}

func (d *joinDatumIterator) Reset() {
	d.location = -1
}

func (d *joinDatumIterator) Len() int {
	return len(d.datums)
}

func (d *joinDatumIterator) Next() bool {
	d.location++
	return d.location < len(d.datums)
}

func (d *joinDatumIterator) Datum() []*Input {
	var result []*Input
	result = append(result, d.datums[d.location]...)
	sortInputs(result)
	return result
}

func (d *joinDatumIterator) DatumN(n int) []*Input {
	d.location = n
	return d.Datum()
}

type gitDatumIterator struct {
	inputs   []*Input
	location int
}

func newGitDatumIterator(pachClient *client.APIClient, input *pps.GitInput) (DatumIterator, error) {
	result := &gitDatumIterator{}
	defer result.Reset()
	if input.Commit == "" {
		// this can happen if a pipeline with multiple inputs has been triggered
		// before all commits have inputs
		return result, nil
	}
	fileInfo, err := pachClient.InspectFile(input.Name, input.Commit, "/commit.json")
	if err != nil {
		return nil, err
	}
	result.inputs = append(
		result.inputs,
		&Input{
			FileInfo: fileInfo,
			Name:     input.Name,
			Branch:   input.Branch,
			GitURL:   input.URL,
		},
	)
	return result, nil
}

func (d *gitDatumIterator) Reset() {
	d.location = -1
}

func (d *gitDatumIterator) Len() int {
	return len(d.inputs)
}

func (d *gitDatumIterator) Datum() []*Input {
	return []*Input{d.inputs[d.location]}
}

func (d *gitDatumIterator) Next() bool {
	d.location++
	return d.location < len(d.inputs)
}

func (d *gitDatumIterator) DatumN(n int) []*Input {
	if n < d.location {
		d.Reset()
	}
	for d.location != n {
		d.Next()
	}
	return d.Datum()
}

func newCronDatumIterator(pachClient *client.APIClient, input *pps.CronInput) (DatumIterator, error) {
	return newPFSDatumIterator(pachClient, &pps.PFSInput{
		Name:   input.Name,
		Repo:   input.Repo,
		Branch: "master",
		Commit: input.Commit,
		Glob:   "/*",
	})
}

// NewDatumIterator creates a datumIterator for an input.
func NewDatumIterator(pachClient *client.APIClient, input *pps.Input) (DatumIterator, error) {
	switch {
	case input.Pfs != nil:
		return newPFSDatumIterator(pachClient, input.Pfs)
	case input.Union != nil:
		return newUnionDatumIterator(pachClient, input.Union)
	case input.Cross != nil:
		return newCrossDatumIterator(pachClient, input.Cross)
	case input.Join != nil:
		return newJoinDatumIterator(pachClient, input.Join)
	case input.Cron != nil:
		return newCronDatumIterator(pachClient, input.Cron)
	case input.Git != nil:
		return newGitDatumIterator(pachClient, input.Git)
	}
	return nil, fmt.Errorf("unrecognized input type")
}

func sortInputs(inputs []*Input) {
	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].Name < inputs[j].Name
	})
}
