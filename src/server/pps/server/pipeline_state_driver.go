package server

import (
	"context"
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/pretty"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type PipelineStateDriver interface {
	// returns PipelineInfo corresponding to the latest pipeline state, a context loaded with the pipeline's auth info, and error
	// NOTE: returns nil, nil, nil if the step is found to be a delete operation
	FetchState(ctx context.Context, pipeline *pps.Pipeline) (*pps.PipelineInfo, context.Context, error)
	// setPipelineState set's pc's state in the collection to 'state'. This will trigger a
	// collection watch event and cause step() to eventually run again.
	SetState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) error
	// TransitionState is similar to SetState, except that it checks whether pipelineInfo @ specCommit
	// is in one of the 'from' states
	TransitionState(ctx context.Context, specCommit *pfs.Commit, from []pps.PipelineState, to pps.PipelineState, reason string) error
	// wraps a Watcher on the pipelines collection
	Watch(ctx context.Context) (<-chan *watch.Event, func(), error)
	// list all PipelineInfos
	ListPipelineInfo(ctx context.Context, f func(*pps.PipelineInfo) error) error
	GetPipelineInfo(ctx context.Context, pipeline *pps.Pipeline, version int) (*pps.PipelineInfo, error)
}

type stateDriver struct {
	db        *pachsql.DB
	pipelines collection.PostgresCollection
	txEnv     *txnenv.TransactionEnv
	pfsApi    pfsserver.APIServer
}

func newPipelineStateDriver(
	db *pachsql.DB,
	pipelines collection.PostgresCollection,
	txEnv *txnenv.TransactionEnv,
	pfsApi pfsserver.APIServer) PipelineStateDriver {
	return &stateDriver{
		db:        db,
		pipelines: pipelines,
		txEnv:     txEnv,
		pfsApi:    pfsApi,
	}
}

// takes pc.ctx
func (sd *stateDriver) FetchState(ctx context.Context, pipeline *pps.Pipeline) (*pps.PipelineInfo, context.Context, error) {
	// query pipelineInfo
	pi, err := sd.tryLoadLatestPipelineInfo(ctx, pipeline)
	if err != nil {
		return nil, nil, errors.Wrap(err, "fetch pipeline state")
	}
	tracing.TagAnySpan(ctx,
		"current-state", pi.State.String(),
		"spec-commit", pretty.CompactPrintCommitSafe(pi.SpecCommit))
	// add pipeline auth
	// the pipelineController's context is authorized as pps master, but we want to switch to the pipeline itself
	// first clear the cached WhoAmI result from the context
	md := metadata.New(map[string]string{auth.ContextTokenKey: pi.AuthToken})
	return pi, metadata.NewOutgoingContext(ctx, md), nil
}

func (sd *stateDriver) SetState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) error {
	if err := setPipelineState(ctx, sd.db, sd.pipelines, specCommit, state, reason); err != nil {
		// don't bother failing if we can't set the state
		return stepError{
			error: errors.Wrapf(err, "could not set pipeline state to %v"+
				"(you may need to restart pachd to un-stick the pipeline)", state),
			retry: true,
		}
	}
	return nil
}

func (sd *stateDriver) TransitionState(ctx context.Context, specCommit *pfs.Commit, from []pps.PipelineState, to pps.PipelineState, reason string) (retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/TransitionPipelineState",
		"project", specCommit.Repo.Project.GetName(),
		"pipeline", specCommit.Repo.Name,
		"from-state", from,
		"to-state", to)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return ppsutil.SetPipelineState(ctx, sd.db, sd.pipelines,
		specCommit, from, to, reason)
}

func (sd *stateDriver) Watch(ctx context.Context) (<-chan *watch.Event, func(), error) {
	pipelineWatcher, err := sd.pipelines.ReadOnly().Watch(ctx)
	if err != nil {
		return nil, nil, errors.EnsureStack(err)
	}
	return pipelineWatcher.Watch(), pipelineWatcher.Close, nil
}

func (sd *stateDriver) ListPipelineInfo(ctx context.Context, f func(*pps.PipelineInfo) error) error {
	return ppsutil.ListPipelineInfo(ctx, sd.pipelines, nil, 0, f)
}

func (sd *stateDriver) GetPipelineInfo(ctx context.Context, pipeline *pps.Pipeline, version int) (*pps.PipelineInfo, error) {
	var pipelineInfo pps.PipelineInfo
	if err := sd.pipelines.ReadOnly().GetUniqueByIndex(
		ctx,
		ppsdb.PipelinesVersionIndex,
		ppsdb.VersionKey(pipeline, uint64(version)),
		&pipelineInfo); err != nil {
		return nil, errors.Wrapf(err, "couldn't retrieve pipeline information")
	}
	return &pipelineInfo, nil
}

func (sd *stateDriver) tryLoadLatestPipelineInfo(ctx context.Context, pipeline *pps.Pipeline) (*pps.PipelineInfo, error) {
	pi := &pps.PipelineInfo{}
	errCnt := 0
	err := backoff.RetryNotify(func() error {
		return sd.loadLatestPipelineInfo(ctx, pipeline, pi)
	}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
		errCnt++
		// Don't put the pipeline in a failing state if we're in the middle
		// of activating auth, retry in a bit
		if (auth.IsErrNotAuthorized(err) || auth.IsErrNotSignedIn(err)) && errCnt <= maxErrCount {
			log.Info(ctx, "could not retrieve pipelineInfo; retrying", zap.Error(err), zap.Duration("retryAfter", d), zap.Int("nErr", errCnt), zap.Int("maxErr", maxErrCount))
			return nil
		}
		return stepError{
			error: errors.Wrapf(err, "could not load pipelineInfo for pipeline %q", pipeline),
			retry: false,
		}
	})
	return pi, err
}

func (sd *stateDriver) loadLatestPipelineInfo(ctx context.Context, pipeline *pps.Pipeline, message *pps.PipelineInfo) error {
	specCommit, err := ppsutil.FindPipelineSpecCommit(ctx, sd.pfsApi, *sd.txEnv, pipeline)
	if err != nil {
		return errors.Wrapf(err, "could not find spec commit for pipeline %q", pipeline)
	}
	if err := sd.pipelines.ReadOnly().Get(ctx, specCommit, message); err != nil {
		return errors.Wrapf(err, "could not retrieve pipeline info for %q", pipeline)
	}
	return nil
}

type mockStateDriver struct {
	pipelines   map[pipelineKey]string              // maps pipeline names to their latest spec commit IDs
	specCommits map[string]*pps.PipelineInfo        // maps spec commit IDs to their pipeline Infos
	states      map[pipelineKey][]pps.PipelineState // tracks all of the
	eChan       chan *watch.Event
	closeEChan  chan struct{} // supports closing eChan
}

func newMockStateDriver() *mockStateDriver {
	d := &mockStateDriver{}
	d.reset()
	d.eChan = make(chan *watch.Event, 1)
	d.closeEChan = make(chan struct{}, 1)
	return d
}

func (d *mockStateDriver) SetState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) error {
	if pi, ok := d.specCommits[specCommit.Id]; ok {
		pi = proto.Clone(pi).(*pps.PipelineInfo)
		pi.State = state
		d.specCommits[specCommit.Id] = pi
		d.states[toKey(pi.Pipeline)] = append(d.states[toKey(pi.Pipeline)], state)
		d.pushWatchEvent(pi, watch.EventPut)
		return nil
	}
	return errors.New("pipeline does not exist")
}

func (d *mockStateDriver) TransitionState(ctx context.Context, specCommit *pfs.Commit, from []pps.PipelineState, to pps.PipelineState, reason string) (retErr error) {
	fromContains := func(ps pps.PipelineState) bool {
		for _, v := range from {
			if v == ps {
				return true
			}
		}
		return false
	}
	if pi, ok := d.specCommits[specCommit.Id]; ok {
		if fromContains(pi.State) {
			return d.SetState(ctx, specCommit, to, reason)
		}
		return ppsutil.PipelineTransitionError{
			Pipeline: pi.Pipeline,
			Expected: from,
			Target:   to,
			Current:  pi.State,
		}
	}
	return errors.New("pipeline does not exist")
}

func (d *mockStateDriver) FetchState(ctx context.Context, pipeline *pps.Pipeline) (*pps.PipelineInfo, context.Context, error) {
	if spec, ok := d.pipelines[toKey(pipeline)]; ok {
		if pi, ok := d.specCommits[spec]; ok {
			return pi, ctx, nil
		}
	}
	return nil, nil, ppsserver.ErrPipelineNotFound{Pipeline: pipeline}
}

func (d *mockStateDriver) Watch(ctx context.Context) (<-chan *watch.Event, func(), error) {
	go func() {
		defer close(d.eChan)
		select {
		case <-ctx.Done():
			d.eChan <- &watch.Event{Type: watch.EventError, Err: context.Cause(ctx)}
			return
		case <-d.closeEChan:
			return
		}
	}()
	return d.eChan, func() {
		select {
		case d.closeEChan <- struct{}{}:
		default:
		}
	}, nil
}

func (d *mockStateDriver) ListPipelineInfo(ctx context.Context, f func(*pps.PipelineInfo) error) error {
	for _, spec := range d.pipelines {
		if pi, ok := d.specCommits[spec]; ok {
			if err := f(pi); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *mockStateDriver) GetPipelineInfo(ctx context.Context, pipeline *pps.Pipeline, version int) (*pps.PipelineInfo, error) {
	if spec, ok := d.pipelines[toKey(pipeline)]; ok {
		if pi, ok := d.specCommits[spec]; ok {
			return pi, nil
		}
	}
	return nil, errors.New("not found")
}

func (d *mockStateDriver) upsertPipeline(pi *pps.PipelineInfo) *pfs.Commit {
	mockSpecCommit := client.NewCommit(pi.Pipeline.Project.GetName(), pi.Pipeline.Name, "master", uuid.NewWithoutDashes())
	pi.SpecCommit = mockSpecCommit
	d.pipelines[toKey(pi.Pipeline)] = pi.SpecCommit.Id
	d.specCommits[mockSpecCommit.Id] = pi
	if ss, ok := d.states[toKey(pi.Pipeline)]; ok {
		d.states[toKey(pi.Pipeline)] = append(ss, pi.State)
	} else {
		d.states[toKey(pi.Pipeline)] = []pps.PipelineState{pi.State}
	}
	d.pushWatchEvent(pi, watch.EventPut)
	return mockSpecCommit
}

func (d *mockStateDriver) pushWatchEvent(pi *pps.PipelineInfo, et watch.EventType) {
	d.eChan <- &watch.Event{
		Key:  []byte(fmt.Sprintf("%s@%s", pi.Pipeline, pi.SpecCommit.Id)),
		Type: et,
	}
}

func (d *mockStateDriver) cancelWatch() {
	d.closeEChan <- struct{}{}
}

func (d *mockStateDriver) reset() {
	d.pipelines = make(map[pipelineKey]string)
	d.specCommits = make(map[string]*pps.PipelineInfo)
	d.states = make(map[pipelineKey][]pps.PipelineState)
}

func (d *mockStateDriver) currentPipelineInfo(key pipelineKey) *pps.PipelineInfo {
	return d.specCommits[d.pipelines[key]]
}
