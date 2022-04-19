package server

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"golang.org/x/net/context"
)

type PipelineStateDriver interface {
	// returns PipelineInfo corresponding to the latest pipeline state, a context loaded with the pipeline's auth info, and error
	// NOTE: returns nil, nil, nil if the step is found to be a delete operation
	FetchState(ctx context.Context, pipeline string) (*pps.PipelineInfo, context.Context, error)
	// setPipelineState set's pc's state in the collection to 'state'. This will trigger a
	// collection watch event and cause step() to eventually run again.
	SetState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) error
	// TransitionState is similar to setPipelineState, except that it sets
	// 'from' and logs a different trace
	TransitionState(ctx context.Context, specCommit *pfs.Commit, from []pps.PipelineState, to pps.PipelineState, reason string) error
	// wraps a Watcher on the pipelines collection
	Watch(ctx context.Context) (<-chan *watch.Event, func(), error)
	// list all PipelineInfos
	ListPipelineInfo(ctx context.Context, f func(*pps.PipelineInfo) error) error
	GetPipelineInfo(ctx context.Context, name string, version int) (*pps.PipelineInfo, error)
}

type mockStateDriver struct {
	specCommits map[string]string              // maps spec commit IDs to pipeline names
	pipelines   map[string]*pps.PipelineInfo   // maps pipeline names to PipelineInfos
	states      map[string][]pps.PipelineState // tracks all of the
	eChan       chan *watch.Event
}

func newMockStateDriver() *mockStateDriver {
	msd := &mockStateDriver{}
	msd.reset()
	return msd
}

func (msd *mockStateDriver) SetState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) error {
	if pipeline, ok := msd.specCommits[specCommit.ID]; ok {
		if pi, ok := msd.pipelines[pipeline]; ok {
			pi.State = state
			msd.states[pipeline] = append(msd.states[pipeline], state)
			return nil
		}
		return errors.New("Ah no pipeline state!!")
	}
	return errors.New("Ah!!")
}

func (msd *mockStateDriver) TransitionState(ctx context.Context, specCommit *pfs.Commit, from []pps.PipelineState, to pps.PipelineState, reason string) (retErr error) {
	return msd.SetState(ctx, specCommit, to, reason)
}

func (msd *mockStateDriver) FetchState(ctx context.Context, pipeline string) (*pps.PipelineInfo, context.Context, error) {
	if pi, ok := msd.pipelines[pipeline]; ok {
		return pi, ctx, nil
	}
	// TODO: make error realistic
	return nil, nil, stepError{
		error: errors.Wrapf(errors.New("pipeline not found"), "could not load pipelineInfo for pipeline %q", pipeline),
		retry: false,
	}
}

func (msd *mockStateDriver) Watch(ctx context.Context) (<-chan *watch.Event, func(), error) {
	return msd.eChan, func() {
		close(msd.eChan)
	}, nil
}

func (msd *mockStateDriver) ListPipelineInfo(ctx context.Context, f func(*pps.PipelineInfo) error) error {
	for _, pi := range msd.pipelines {
		if err := f(pi); err != nil {
			return err
		}
	}
	return nil
}

func (msd *mockStateDriver) GetPipelineInfo(ctx context.Context, name string, version int) (*pps.PipelineInfo, error) {
	if pi, ok := msd.pipelines[name]; ok {
		return pi, nil
	}
	return nil, errors.New("not found")
}

func (msd *mockStateDriver) upsertPipeline(pi *pps.PipelineInfo) *pfs.Commit {
	msd.pipelines[pi.Pipeline.Name] = pi
	mockSpecCommit := client.NewCommit(pi.Pipeline.Name, "master", uuid.NewWithoutDashes())
	pi.SpecCommit = mockSpecCommit
	msd.specCommits[mockSpecCommit.ID] = pi.Pipeline.Name
	if ss, ok := msd.states[pi.Pipeline.Name]; ok {
		msd.states[pi.Pipeline.Name] = append(ss, pi.State)
	} else {
		msd.states[pi.Pipeline.Name] = []pps.PipelineState{pi.State}
	}
	msd.pushWatchEvent(&watch.Event{
		Key:  []byte(fmt.Sprintf("%s@%s", pi.Pipeline.Name, mockSpecCommit.ID)),
		Type: watch.EventPut,
	})
	return mockSpecCommit
}

func (msd *mockStateDriver) pushWatchEvent(ev *watch.Event) {
	msd.eChan <- ev
}

func (msd *mockStateDriver) reset() {
	msd.specCommits = make(map[string]string)
	msd.pipelines = make(map[string]*pps.PipelineInfo)
	msd.states = make(map[string][]pps.PipelineState)
	msd.eChan = make(chan *watch.Event)
}
