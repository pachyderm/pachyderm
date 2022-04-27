package server

import (
	"fmt"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
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
func (sd *stateDriver) FetchState(ctx context.Context, pipeline string) (*pps.PipelineInfo, context.Context, error) {
	// query pipelineInfo
	var pi *pps.PipelineInfo
	var err error
	if pi, err = sd.tryLoadLatestPipelineInfo(ctx, pipeline); err != nil && collection.IsErrNotFound(err) {
		// if the pipeline info is not found, interpret the operation as a delete
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
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
		"/pps.Master/TransitionPipelineState", "pipeline", specCommit.Branch.Repo.Name,
		"from-state", from, "to-state", to)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return ppsutil.SetPipelineState(ctx, sd.db, sd.pipelines,
		specCommit, from, to, reason)
}

func (sd *stateDriver) Watch(ctx context.Context) (<-chan *watch.Event, func(), error) {
	pipelineWatcher, err := sd.pipelines.ReadOnly(ctx).Watch()
	if err != nil {
		return nil, nil, errors.EnsureStack(err)
	}
	return pipelineWatcher.Watch(), func() { pipelineWatcher.Close() }, nil
}

func (sd *stateDriver) ListPipelineInfo(ctx context.Context, f func(*pps.PipelineInfo) error) error {
	return ppsutil.ListPipelineInfo(ctx, sd.pipelines, nil, 0, f)
}

func (sd *stateDriver) GetPipelineInfo(ctx context.Context, name string, version int) (*pps.PipelineInfo, error) {
	var pipelineInfo pps.PipelineInfo
	if err := sd.pipelines.ReadOnly(ctx).GetUniqueByIndex(
		ppsdb.PipelinesVersionIndex,
		ppsdb.VersionKey(name, uint64(version)),
		&pipelineInfo); err != nil {
		return nil, errors.Wrapf(err, "couldn't retrieve pipeline information")
	}
	return &pipelineInfo, nil
}

func (sd *stateDriver) tryLoadLatestPipelineInfo(ctx context.Context, pipeline string) (*pps.PipelineInfo, error) {
	pi := &pps.PipelineInfo{}
	errCnt := 0
	err := backoff.RetryNotify(func() error {
		return sd.loadLatestPipelineInfo(ctx, pipeline, pi)
	}, backoff.NewExponentialBackOff(), func(err error, d time.Duration) error {
		errCnt++
		// Don't put the pipeline in a failing state if we're in the middle
		// of activating auth, retry in a bit
		if (auth.IsErrNotAuthorized(err) || auth.IsErrNotSignedIn(err)) && errCnt <= maxErrCount {
			log.Warnf("PPS master: could not retrieve pipelineInfo for pipeline %q: %v; retrying in %v",
				pipeline, err, d)
			return nil
		}
		return stepError{
			error: errors.Wrapf(err, "could not load pipelineInfo for pipeline %q", pipeline),
			retry: false,
		}
	})
	return pi, err
}

func (sd *stateDriver) loadLatestPipelineInfo(ctx context.Context, pipeline string, message *pps.PipelineInfo) error {
	specCommit, err := ppsutil.FindPipelineSpecCommit(ctx, sd.pfsApi, *sd.txEnv, pipeline)
	if err != nil {
		return errors.Wrapf(err, "could not find spec commit for pipeline %q", pipeline)
	}
	// TODO: which context should be passed here? master's ctx?
	if err := sd.pipelines.ReadOnly(ctx).Get(specCommit, message); err != nil {
		return errors.Wrapf(err, "could not retrieve pipeline info for %q", pipeline)
	}
	return nil
}

type mockStateDriver struct {
	specCommits map[string]string              // maps spec commit IDs to pipeline names
	pipelines   map[string]*pps.PipelineInfo   // maps pipeline names to PipelineInfos
	states      map[string][]pps.PipelineState // tracks all of the
	eChan       chan *watch.Event
	doneEChan   chan struct{} // supports closing eChan
}

func newMockStateDriver() *mockStateDriver {
	msd := &mockStateDriver{}
	msd.reset()
	msd.eChan = make(chan *watch.Event, 1)
	msd.done = make(chan struct{}, 5)
	return msd
}

func (msd *mockStateDriver) SetState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) error {
	if pipeline, ok := msd.specCommits[specCommit.ID]; ok {
		if pi, ok := msd.pipelines[pipeline]; ok {
			pi.State = state
			msd.states[pipeline] = append(msd.states[pipeline], state)
			msd.pushWatchEvent(pi, watch.EventPut)
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
	return nil, nil, nil
}

func (msd *mockStateDriver) Watch(ctx context.Context) (<-chan *watch.Event, func(), error) {
	go func() {
		select {
		case <-ctx.Done():
			close(msd.eChan)
			return
		case <-msd.doneEChan:
			close(msd.eChan)
			return
		}
	}()
	return msd.eChan, func() {
		select {
		case msd.doneEChan <- struct{}{}:
		default:
		}
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
	pipelineKey := pi.Pipeline.Name
	msd.pipelines[pipelineKey] = pi
	mockSpecCommit := client.NewCommit(pi.Pipeline.Name, "master", uuid.NewWithoutDashes())
	pi.SpecCommit = mockSpecCommit
	msd.specCommits[mockSpecCommit.ID] = pipelineKey
	if ss, ok := msd.states[pipelineKey]; ok {
		msd.states[pipelineKey] = append(ss, pi.State)
	} else {
		msd.states[pipelineKey] = []pps.PipelineState{pi.State}
	}
	msd.pushWatchEvent(pi, watch.EventPut)
	return mockSpecCommit
}

func (msd *mockStateDriver) pushWatchEvent(pi *pps.PipelineInfo, et watch.EventType) {
	msd.eChan <- &watch.Event{
		Key:  []byte(fmt.Sprintf("%s@%s", pi.Pipeline.Name, pi.SpecCommit.ID)),
		Type: et,
	}
}

func (msd *mockStateDriver) cancelWatch() {
	msd.doneEChan <- struct{}{}
}

func (msd *mockStateDriver) reset() {
	msd.specCommits = make(map[string]string)
	msd.pipelines = make(map[string]*pps.PipelineInfo)
	msd.states = make(map[string][]pps.PipelineState)
}
