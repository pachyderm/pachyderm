package datum

import (
	"io"
	"sort"

	glob "github.com/pachyderm/ohmyglob"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/worker/common"

	"github.com/cevaris/ordered_map"
)

// Iterator is an interface which allows you to iterate through the datums
// for a job. A datum iterator keeps track of which datum it is on, which can be Reset()
// The intended use is by using this pattern `for di.Next() { ... datum := di.Datum() ... }`
// Note that since you start the loop by a call to Next(), the datum iterator's location starts at -1
type Iterator interface {
	Reset()
	Len() int
	Next() bool
	Datum() []*common.Input
	DatumN(int) []*common.Input
}

type pfsIterator struct {
	inputs   []*common.Input
	location int
}

func newPFSIterator(pachClient *client.APIClient, input *pps.PFSInput) (Iterator, error) {
	result := &pfsIterator{}
	// make sure it gets initialized properly (location = -1)
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
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		g, err := glob.Compile(input.Glob, '/')
		if err != nil {
			return nil, err
		}
		joinOn := g.Replace(fileInfo.File.Path, input.JoinOn)
		groupBy := g.Replace(fileInfo.File.Path, input.GroupBy)
		result.inputs = append(result.inputs, &common.Input{
			FileInfo:   fileInfo,
			JoinOn:     joinOn,
			GroupBy:    groupBy,
			Name:       input.Name,
			Lazy:       input.Lazy,
			Branch:     input.Branch,
			EmptyFiles: input.EmptyFiles,
			S3:         input.S3,
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

func (d *pfsIterator) Reset() {
	d.location = -1
}

func (d *pfsIterator) Len() int {
	return len(d.inputs)
}

func (d *pfsIterator) Datum() []*common.Input {
	return []*common.Input{d.inputs[d.location]}
}

func (d *pfsIterator) DatumN(n int) []*common.Input {
	return []*common.Input{d.inputs[n]}
}

func (d *pfsIterator) Next() bool {
	if d.location < len(d.inputs) {
		d.location++
	}
	return d.location < len(d.inputs)
}

type listIterator struct {
	inputs   []*common.Input
	location int
}

func newListIterator(pachClient *client.APIClient, inputs []*common.Input) (Iterator, error) {
	result := &listIterator{}
	// make sure it gets initialized properly
	result.Reset()
	result.inputs = inputs
	return result, nil
}

func (d *listIterator) Reset() {
	d.location = -1
}

func (d *listIterator) Len() int {
	return len(d.inputs)
}

func (d *listIterator) Datum() []*common.Input {
	return []*common.Input{d.inputs[d.location]}
}

func (d *listIterator) DatumN(n int) []*common.Input {
	return []*common.Input{d.inputs[n]}
}

func (d *listIterator) Next() bool {
	if d.location < len(d.inputs) {
		d.location++
	}
	return d.location < len(d.inputs)
}

type unionIterator struct {
	iterators []Iterator
	unionIdx  int
	location  int
}

func newUnionIterator(pachClient *client.APIClient, union []*pps.Input) (Iterator, error) {
	result := &unionIterator{}
	defer result.Reset()
	for _, input := range union {
		datumIterator, err := NewIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		result.iterators = append(result.iterators, datumIterator)
	}
	return result, nil
}

func (d *unionIterator) Reset() {
	for _, input := range d.iterators {
		input.Reset()
	}
	d.unionIdx = 0
	d.location = -1
}

func (d *unionIterator) Len() int {
	result := 0
	for _, datumIterator := range d.iterators {
		result += datumIterator.Len()
	}
	return result
}

func (d *unionIterator) Next() bool {
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

func (d *unionIterator) Datum() []*common.Input {
	return d.iterators[d.unionIdx].Datum()
}

func (d *unionIterator) DatumN(n int) []*common.Input {
	for _, datumIterator := range d.iterators {
		if n < datumIterator.Len() {
			return datumIterator.DatumN(n)
		}
		n -= datumIterator.Len()
	}
	panic("index out of bounds")
}

type crossIterator struct {
	iterators     []Iterator
	started, done bool
	location      int
}

func newCrossIterator(pachClient *client.APIClient, cross []*pps.Input) (Iterator, error) {
	result := &crossIterator{}
	defer result.Reset() // Call Next() on all inner iterators once
	for _, iterator := range cross {
		datumIterator, err := NewIterator(pachClient, iterator)
		if err != nil {
			return nil, err
		}
		result.iterators = append(result.iterators, datumIterator)
	}
	result.location = -1
	return result, nil
}

func newCrossListIterator(pachClient *client.APIClient, cross [][]*common.Input) (Iterator, error) {
	result := &crossIterator{}
	defer result.Reset()
	for _, iterator := range cross {
		datumIterator, err := newListIterator(pachClient, iterator)
		if err != nil {
			return nil, err
		}
		result.iterators = append(result.iterators, datumIterator)
	}
	result.location = -1
	return result, nil
}

func (d *crossIterator) Reset() {
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
	d.done = d.started
}

func (d *crossIterator) Len() int {
	if len(d.iterators) == 0 {
		return 0
	}
	result := d.iterators[0].Len()
	for i := 1; i < len(d.iterators); i++ {
		result *= d.iterators[i].Len()
	}
	return result
}

func (d *crossIterator) Next() bool {
	if !d.started {
		d.started = true
		d.location++
		// First call to Next() does nothing, as Reset() calls Next() on all inner
		// datums once already
		return true
	}
	if d.done {
		return false
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
	d.done = true
	return false
}

func (d *crossIterator) Datum() []*common.Input {
	var result []*common.Input
	for _, datumIterator := range d.iterators {
		result = append(result, datumIterator.Datum()...)
	}
	sortInputs(result)
	return result
}

func (d *crossIterator) DatumN(n int) []*common.Input {
	if n >= d.Len() {
		panic("index out of bounds")
	}
	var result []*common.Input
	for _, datumIterator := range d.iterators {
		result = append(result, datumIterator.DatumN(n%datumIterator.Len())...)
		n /= datumIterator.Len()
	}
	sortInputs(result)
	return result
}

type groupIterator struct {
	datums   [][]*common.Input
	location int
}

func newGroupIterator(pachClient *client.APIClient, group []*pps.Input) (Iterator, error) {
	groupMap := make(map[string][]*common.Input)
	keys := make([]string, 0, len(group))
	result := &groupIterator{}
	defer result.Reset()

	// okay, so we have a slice of pps Inputs
	for _, input := range group {
		// turn our inputs into iterators
		datumIterator, err := NewIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		// iterate through each iterator to get the individual datums
		for datumIterator.Next() {
			datums := datumIterator.Datum()
			for _, datum := range datums {
				// put the datums in an map keyed by GroupBy
				groupDatums, ok := groupMap[datum.GroupBy]
				if !ok {
					keys = append(keys, datum.GroupBy)
				}
				groupMap[datum.GroupBy] = append(groupDatums, datum)
			}
		}
	}

	// sort everything by the group_by
	sort.Sort(sort.StringSlice(keys))
	// put each equivalence class into its own datum
	for _, key := range keys {
		result.datums = append(result.datums, groupMap[key])
	}

	return result, nil
}

func (d *groupIterator) Reset() {
	d.location = -1
}

func (d *groupIterator) Len() int {
	return len(d.datums)
}

func (d *groupIterator) Next() bool {
	if d.location < len(d.datums) {
		d.location++
	}
	return d.location < len(d.datums)
}

func (d *groupIterator) Datum() []*common.Input {
	var result []*common.Input
	result = append(result, d.datums[d.location]...)
	sortInputs(result)
	return result
}

func (d *groupIterator) DatumN(n int) []*common.Input {
	d.location = n
	return d.Datum()
}

type joinIterator struct {
	datums   [][]*common.Input
	location int
}

func newJoinIterator(pachClient *client.APIClient, join []*pps.Input) (Iterator, error) {
	result := &joinIterator{}
	om := ordered_map.NewOrderedMap()

	for i, input := range join {
		datumIterator, err := NewIterator(pachClient, input)
		if err != nil {
			return nil, err
		}
		for datumIterator.Next() {
			x := datumIterator.Datum()
			for _, k := range x {
				tupleI, ok := om.Get(k.JoinOn)
				var tuple [][]*common.Input
				if !ok {
					tuple = make([][]*common.Input, len(join))
				} else {
					tuple = tupleI.([][]*common.Input)
				}
				tuple[i] = append(tuple[i], k)
				om.Set(k.JoinOn, tuple)
			}
		}
	}

	iter := om.IterFunc()
	for kv, ok := iter(); ok; kv, ok = iter() {
		tuple := kv.Value.([][]*common.Input)
		cross, err := newCrossListIterator(pachClient, tuple)
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

func (d *joinIterator) Reset() {
	d.location = -1
}

func (d *joinIterator) Len() int {
	return len(d.datums)
}

func (d *joinIterator) Next() bool {
	if d.location < len(d.datums) {
		d.location++
	}
	return d.location < len(d.datums)
}

func (d *joinIterator) Datum() []*common.Input {
	var result []*common.Input
	result = append(result, d.datums[d.location]...)
	sortInputs(result)
	return result
}

func (d *joinIterator) DatumN(n int) []*common.Input {
	d.location = n
	return d.Datum()
}

type gitIterator struct {
	inputs   []*common.Input
	location int
}

func newGitIterator(pachClient *client.APIClient, input *pps.GitInput) (Iterator, error) {
	result := &gitIterator{}
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
		&common.Input{
			FileInfo: fileInfo,
			Name:     input.Name,
			Branch:   input.Branch,
			GitURL:   input.URL,
		},
	)
	return result, nil
}

func (d *gitIterator) Reset() {
	d.location = -1
}

func (d *gitIterator) Len() int {
	return len(d.inputs)
}

func (d *gitIterator) Datum() []*common.Input {
	return []*common.Input{d.inputs[d.location]}
}

func (d *gitIterator) Next() bool {
	if d.location < len(d.inputs) {
		d.location++
	}
	return d.location < len(d.inputs)
}

func (d *gitIterator) DatumN(n int) []*common.Input {
	if n < d.location {
		d.Reset()
	}
	for d.location != n {
		d.Next()
	}
	return d.Datum()
}

func newCronIterator(pachClient *client.APIClient, input *pps.CronInput) (Iterator, error) {
	return newPFSIterator(pachClient, &pps.PFSInput{
		Name:   input.Name,
		Repo:   input.Repo,
		Branch: "master",
		Commit: input.Commit,
		Glob:   "/*",
	})
}

// NewIterator creates an Iterator for an input.
func NewIterator(pachClient *client.APIClient, input *pps.Input) (Iterator, error) {
	switch {
	case input.Pfs != nil:
		return newPFSIterator(pachClient, input.Pfs)
	case input.Union != nil:
		return newUnionIterator(pachClient, input.Union)
	case input.Cross != nil:
		return newCrossIterator(pachClient, input.Cross)
	case input.Join != nil:
		return newJoinIterator(pachClient, input.Join)
	case input.Group != nil:
		return newGroupIterator(pachClient, input.Group)
	case input.Cron != nil:
		return newCronIterator(pachClient, input.Cron)
	case input.Git != nil:
		return newGitIterator(pachClient, input.Git)
	}
	return nil, errors.Errorf("unrecognized input type: %v", input)
}

func sortInputs(inputs []*common.Input) {
	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].Name < inputs[j].Name
	})
}
