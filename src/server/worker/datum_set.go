package worker

import (
	"fmt"
	"io"
	"os/exec"
	"sort"

	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
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

type atomDatumFactory struct {
	inputs []*Input
}

func newAtomDatumFactory(pachClient *client.APIClient, input *pps.AtomInput) (DatumFactory, error) {
	result := &atomDatumFactory{}
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

func (d *atomDatumFactory) Len() int {
	return len(d.inputs)
}

func (d *atomDatumFactory) Datum(i int) []*Input {
	return []*Input{d.inputs[i]}
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

func newUnionDatumFactory(pachClient *client.APIClient, union []*pps.Input, logger io.Writer) (DatumFactory, error) {
	result := &unionDatumFactory{}
	for _, input := range union {
		datumFactory, err := NewDatumFactory(pachClient, input, logger)
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

func newCrossDatumFactory(pachClient *client.APIClient, cross []*pps.Input, logger io.Writer) (DatumFactory, error) {
	result := &crossDatumFactory{}
	for _, input := range cross {
		datumFactory, err := NewDatumFactory(pachClient, input, logger)
		if err != nil {
			return nil, err
		}
		result.inputs = append(result.inputs, datumFactory)
	}
	return result, nil
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

func newCronDatumFactory(pachClient *client.APIClient, input *pps.CronInput) (DatumFactory, error) {
	return newPFSDatumFactory(pachClient, &pps.PFSInput{
		Name:   input.Name,
		Repo:   input.Repo,
		Branch: "master",
		Commit: input.Commit,
		Glob:   "/*",
	})
}

type filterDatumFactory struct {
	pred  pps.FilterFunc
	input DatumFactory
}

func (d *filterDatumFactory) Len() int {
	return d.input.Len()
}

func (d *filterDatumFactory) Datum(i int) []*Input {
	result := d.input.Datum(i)
	var fis []*pfs.FileInfo
	for _, in := range result {
		fis = append(fis, in.FileInfo)
	}
	if d.pred(fis) {
		return result
	}
	return nil
}

func newFilterDatumFactory(pachClient *client.APIClient, input *pps.FilterInput, logger io.Writer) (DatumFactory, error) {
	in, err := NewDatumFactory(pachClient, input.Input, logger)
	if err != nil {
		return nil, err
	}
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  pps.FilterHandshake,
		Plugins:          pps.FilterPluginMap,
		Cmd:              exec.Command("sh", "-c", input.Predicate.Source),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Managed:          true,
		Logger: hclog.New(&hclog.LoggerOptions{
			Level:  hclog.Warn,
			Output: logger,
		}),
	})
	grpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}
	raw, err := grpcClient.Dispense("filter_grpc")
	if err != nil {
		return nil, err
	}
	c := raw.(*pps.FilterGRPCClient)
	return &filterDatumFactory{
		pred: func(fis []*pfs.FileInfo) bool {
			result, err := c.Filter(fis)
			if err != nil {
				panic(err)
			}
			return result
		},
		input: in,
	}, nil
}

// NewDatumFactory creates a datumFactory for an input.
func NewDatumFactory(pachClient *client.APIClient, input *pps.Input, logger io.Writer) (DatumFactory, error) {
	switch {
	case input.Atom != nil:
		return newAtomDatumFactory(pachClient, input.Atom)
	case input.Pfs != nil:
		return newPFSDatumFactory(pachClient, input.Pfs)
	case input.Union != nil:
		return newUnionDatumFactory(pachClient, input.Union, logger)
	case input.Cross != nil:
		return newCrossDatumFactory(pachClient, input.Cross, logger)
	case input.Cron != nil:
		return newCronDatumFactory(pachClient, input.Cron)
	case input.Git != nil:
		return newGitDatumFactory(pachClient, input.Git)
	case input.Filter != nil:
		return newFilterDatumFactory(pachClient, input.Filter, logger)
	}
	return nil, fmt.Errorf("unrecognized input type")
}

func sortInputs(inputs []*Input) {
	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].Name < inputs[j].Name
	})
}
