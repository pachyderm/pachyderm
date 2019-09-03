package datum

import (
	"fmt"
	"io"
	"sort"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

// Factory is an interface which allows you to iterate through the datums
// for a job.
type Factory interface {
	Len() int
	Datum(i int) []*common.Input
}

type pfsFactory struct {
	inputs []*common.Input
}

func newPFSFactory(pachClient *client.APIClient, input *pps.PFSInput) (Factory, error) {
	result := &pfsFactory{}
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
		result.inputs = append(result.inputs, &common.Input{
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

func (d *pfsFactory) Len() int {
	return len(d.inputs)
}

func (d *pfsFactory) Datum(i int) []*common.Input {
	return []*common.Input{d.inputs[i]}
}

type unionFactory struct {
	inputs []Factory
}

func newUnionFactory(pachClient *client.APIClient, union []*pps.Input) (Factory, error) {
	result := &unionFactory{}
	for _, input := range union {
		datumFactory, err := NewFactory(pachClient, input)
		if err != nil {
			return nil, err
		}
		result.inputs = append(result.inputs, datumFactory)
	}
	return result, nil
}

func (d *unionFactory) Len() int {
	result := 0
	for _, datumFactory := range d.inputs {
		result += datumFactory.Len()
	}
	return result
}

func (d *unionFactory) Datum(i int) []*common.Input {
	for _, datumFactory := range d.inputs {
		if i < datumFactory.Len() {
			return datumFactory.Datum(i)
		}
		i -= datumFactory.Len()
	}
	panic("index out of bounds")
}

type crossFactory struct {
	inputs []Factory
}

func (d *crossFactory) Len() int {
	if len(d.inputs) == 0 {
		return 0
	}
	result := d.inputs[0].Len()
	for i := 1; i < len(d.inputs); i++ {
		result *= d.inputs[i].Len()
	}
	return result
}

func (d *crossFactory) Datum(i int) []*common.Input {
	if i >= d.Len() {
		panic("index out of bounds")
	}
	var result []*common.Input
	for _, datumFactory := range d.inputs {
		result = append(result, datumFactory.Datum(i%datumFactory.Len())...)
		i /= datumFactory.Len()
	}
	sortInputs(result)
	return result
}

type gitFactory struct {
	inputs []*common.Input
}

func newGitFactory(pachClient *client.APIClient, input *pps.GitInput) (Factory, error) {
	result := &gitFactory{}
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
		&common.Input{
			FileInfo: fileInfo,
			Name:     input.Name,
			Branch:   input.Branch,
			GitURL:   input.URL,
		},
	)
	return result, nil
}

func (d *gitFactory) Len() int {
	return len(d.inputs)
}

func (d *gitFactory) Datum(i int) []*common.Input {
	return []*common.Input{d.inputs[i]}
}

func newCrossFactory(pachClient *client.APIClient, cross []*pps.Input) (Factory, error) {
	result := &crossFactory{}
	for _, input := range cross {
		datumFactory, err := NewFactory(pachClient, input)
		if err != nil {
			return nil, err
		}
		result.inputs = append(result.inputs, datumFactory)
	}
	return result, nil
}

func newCronFactory(pachClient *client.APIClient, input *pps.CronInput) (Factory, error) {
	return newPFSFactory(pachClient, &pps.PFSInput{
		Name:   input.Name,
		Repo:   input.Repo,
		Branch: "master",
		Commit: input.Commit,
		Glob:   "/*",
	})
}

// NewFactory creates a datumFactory for an input.
func NewFactory(pachClient *client.APIClient, input *pps.Input) (Factory, error) {
	switch {
	case input.Pfs != nil:
		return newPFSFactory(pachClient, input.Pfs)
	case input.Union != nil:
		return newUnionFactory(pachClient, input.Union)
	case input.Cross != nil:
		return newCrossFactory(pachClient, input.Cross)
	case input.Cron != nil:
		return newCronFactory(pachClient, input.Cron)
	case input.Git != nil:
		return newGitFactory(pachClient, input.Git)
	}
	return nil, fmt.Errorf("unrecognized input type")
}
