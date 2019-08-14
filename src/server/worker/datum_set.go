package worker

import (
	"fmt"
	"io"
	"sort"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

// DatumFactory is an interface which allows you to iterate through the datums
// for a job.
type DatumFactory interface {
	Len() int
	Datum(i int) []*Input
}

type pfsDatumFactory struct {
	inputs []*Input
}

func newPFSDatumFactory(pachClient *client.APIClient, input *pps.PFSInput) (DatumFactory, error) {
	result := &pfsDatumFactory{}
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
	return result, nil
}

func (d *pfsDatumFactory) Len() int {
	return len(d.inputs)
}

func (d *pfsDatumFactory) Datum(i int) []*Input {
	return []*Input{d.inputs[i]}
}

type unionDatumFactory struct {
	inputs []DatumFactory
}

func newUnionDatumFactory(pachClient *client.APIClient, union []*pps.Input) (DatumFactory, error) {
	result := &unionDatumFactory{}
	for _, input := range union {
		datumFactory, err := NewDatumFactory(pachClient, input)
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
	inputs []DatumFactory
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
	sortInputs(result)
	return result
}

type gitDatumFactory struct {
	inputs []*Input
}

func newGitDatumFactory(pachClient *client.APIClient, input *pps.GitInput) (DatumFactory, error) {
	result := &gitDatumFactory{}
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

func (d *gitDatumFactory) Len() int {
	return len(d.inputs)
}

func (d *gitDatumFactory) Datum(i int) []*Input {
	return []*Input{d.inputs[i]}
}

func newCrossDatumFactory(pachClient *client.APIClient, cross []*pps.Input) (DatumFactory, error) {
	result := &crossDatumFactory{}
	for _, input := range cross {
		datumFactory, err := NewDatumFactory(pachClient, input)
		if err != nil {
			return nil, err
		}
		result.inputs = append(result.inputs, datumFactory)
	}
	return result, nil
}

func newCronDatumFactory(pachClient *client.APIClient, input *pps.CronInput) (DatumFactory, error) {
	return newPFSDatumFactory(pachClient, &pps.PFSInput{
		Name:   input.Name,
		Repo:   input.Repo,
		Branch: "master",
		Commit: input.Commit,
		Glob:   "/*",
	})
}

// NewDatumFactory creates a datumFactory for an input.
func NewDatumFactory(pachClient *client.APIClient, input *pps.Input) (DatumFactory, error) {
	switch {
	case input.Pfs != nil:
		return newPFSDatumFactory(pachClient, input.Pfs)
	case input.Union != nil:
		return newUnionDatumFactory(pachClient, input.Union)
	case input.Cross != nil:
		return newCrossDatumFactory(pachClient, input.Cross)
	case input.Cron != nil:
		return newCronDatumFactory(pachClient, input.Cron)
	case input.Git != nil:
		return newGitDatumFactory(pachClient, input.Git)
	}
	return nil, fmt.Errorf("unrecognized input type")
}
