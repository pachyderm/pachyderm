package worker

import (
	"fmt"
	"io"
	"sort"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
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
}

type pfsDatumIterator struct {
	inputs   []*Input
	location int
}

func newPFSDatumIterator(pachClient *client.APIClient, input *pps.PFSInput) (DatumIterator, error) {
	result := &pfsDatumIterator{}
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
		result.inputs = append(result.inputs, &Input{
			FileInfo:   fileInfo,
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
	// make sure it gets initialized properly
	result.Reset()
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

func (d *pfsDatumIterator) Next() bool {
	d.location++
	return d.location < len(d.inputs)
}

type unionDatumIterator struct {
	inputs   []DatumIterator
	unionIdx int
}

func newUnionDatumIterator(pachClient *client.APIClient, union []*pps.Input) (DatumIterator, error) {
	result := &unionDatumIterator{}
	for _, input := range union {
		datumIterator, err := NewDatumIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		result.inputs = append(result.inputs, datumIterator)
	}
	result.Reset()
	return result, nil
}

func (d *unionDatumIterator) Reset() {
	for _, input := range d.inputs {
		input.Reset()
	}
	d.unionIdx = 0
}

func (d *unionDatumIterator) Len() int {
	result := 0
	for _, datumIterator := range d.inputs {
		result += datumIterator.Len()
	}
	return result
}

func (d *unionDatumIterator) Next() bool {
	if len(d.inputs) == 0 {
		return false
	}
	inputHasNext := d.inputs[d.unionIdx].Next()
	if !inputHasNext {
		d.unionIdx++
		if d.unionIdx < len(d.inputs) {
			return d.inputs[d.unionIdx].Next()
		}
	}
	return false
}

func (d *unionDatumIterator) Datum() []*Input {
	return d.inputs[d.unionIdx].Datum()
}

type crossDatumIterator struct {
	inputs  []DatumIterator
	started bool
}

func newCrossDatumIterator(pachClient *client.APIClient, cross []*pps.Input) (DatumIterator, error) {
	result := &crossDatumIterator{}
	for _, input := range cross {
		datumIterator, err := NewDatumIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		result.inputs = append(result.inputs, datumIterator)
	}
	return result, nil
}

func (d *crossDatumIterator) Reset() {
	for _, input := range d.inputs {
		input.Reset()
	}
	d.started = false
}

func (d *crossDatumIterator) Len() int {
	if len(d.inputs) == 0 {
		return 0
	}
	result := d.inputs[0].Len()
	for i := 1; i < len(d.inputs); i++ {
		result *= d.inputs[i].Len()
	}
	return result
}

func (d *crossDatumIterator) Next() bool {
	// we need to make sure each of the iterators has been started (i.e is not at -1)
	hasNext := true
	if !d.started {
		for _, input := range d.inputs {
			hasNext = hasNext && input.Next()
		}
		if hasNext {
			d.started = true
		}
		return hasNext
	}

	// this is the normal case
	for _, input := range d.inputs {
		// if we're at the end of the "row"
		if !input.Next() {
			// we reset the "row"
			input.Reset()

			if !input.Next() {
				// this should never happen, since the started case should handle this
				return false
			}
			// after resetting this "row", start iterating through the next "row"
		} else {
			return true
		}
	}
	return true
}

func (d *crossDatumIterator) Datum() []*Input {
	var result []*Input
	for _, datumIterator := range d.inputs {
		result = append(result, datumIterator.Datum()...)
	}
	sortInputs(result)
	return result
}

type gitDatumIterator struct {
	inputs   []*Input
	location int
}

func newGitDatumIterator(pachClient *client.APIClient, input *pps.GitInput) (DatumIterator, error) {
	result := &gitDatumIterator{}
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
	result.Reset()
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
