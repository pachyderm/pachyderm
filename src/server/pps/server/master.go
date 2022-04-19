package server

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	middleware_auth "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
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
)

const (
	masterLockPath = "_master_lock"
	maxErrCount    = 3 // gives all retried operations ~4.5s total to finish
)

var (
	failures = map[string]bool{
		"InvalidImageName": true,
		"ErrImagePull":     true,
		"Unschedulable":    true,
	}

	zero     int32 // used to turn down RCs in scaleDownWorkersForPipeline
	falseVal bool  // used to delete RCs in deletePipelineResources and restartPipeline()
)

type pipelineEvent struct {
	pipeline  string
	timestamp time.Time
}

type stepError struct {
	error
	retry        bool
	failPipeline bool
}

func newRetriableError(err error, message string) error {
	return stepError{
		error:        errors.Wrap(err, message),
		retry:        true,
		failPipeline: true,
	}
}

func (s stepError) Unwrap() error {
	return s.error
}

type ppsMaster struct {
	env Env
	// for checking worker status during crashing monitor
	etcdPrefix string
	// masterCtx is a context that is cancelled if
	// the current pps master loses its master status
	masterCtx context.Context
	// fields for the pollPipelines, pollPipelinePods, and watchPipelines goros
	pollPipelinesMu sync.Mutex
	pollCancel      func() // protected by pollPipelinesMu
	pollPodsCancel  func() // protected by pollPipelinesMu
	watchCancel     func() // protected by pollPipelinesMu
	pcMgr           *pcManager
	kd              InfraDriver
	sm              PipelineStateManager
	// channel through which pipeline events are passed
	eventCh chan *pipelineEvent
}

func newMaster(ctx context.Context, env Env, etcdPrefix string, kd InfraDriver, sm PipelineStateManager) *ppsMaster {
	return &ppsMaster{
		masterCtx:  ctx,
		env:        env,
		etcdPrefix: etcdPrefix,
		pcMgr:      newPcManager(),
		kd:         kd,
		sm:         sm,
	}
}

// The master process is responsible for creating/deleting workers as
// pipelines are created/removed.
func (a *apiServer) master() {
	masterLock := dlock.NewDLock(a.env.EtcdClient, path.Join(a.etcdPrefix, masterLockPath))
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		// set internal auth for basic operations
		ctx = middleware_auth.AsInternalUser(ctx, "pps-master")
		defer cancel()
		ctx, err := masterLock.Lock(ctx)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer masterLock.Unlock(ctx)
		log.Infof("PPS master: launching master process")
		kd := newKubeDriver(a.env.KubeClient, a.env.Config, a.env.Logger)
		sm := newPipelineStateManager(a.env.DB, a.pipelines, a.txnEnv, a.env.PFSServer)
		m := newMaster(ctx, a.env, a.etcdPrefix, kd, sm)
		m.run()
		return errors.Wrapf(ctx.Err(), "ppsMaster.Run() exited unexpectedly")
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Errorf("PPS master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
	panic("internal error: PPS master has somehow exited. Restarting pod...")
}

func (m *ppsMaster) setPipelineCrashing(ctx context.Context, specCommit *pfs.Commit, reason string) error {
	return m.sm.SetState(ctx, specCommit, pps.PipelineState_PIPELINE_CRASHING, reason)
}

// run() ingests pipeline events from `m.eventCh` that are generated by watching the DB and
// from polling k8s. It then distributes the work to pipelineController goroutines that refresh
// the worker(s) state for each pipeline.
//
// Notes:
// - Since each event `e` instructs the master to set a particular pipeline `p` to its most recently
// declared state, and because pipelines don't share k8s resources, we can run a single goroutine
// for each `p` to increase throughput across pipelines.
//
// - In the case where many events are queued for a given pipeline, we can skip to the
// most recent event in the next `step` the pipelineController executes. This is done using `pipelineController.Bump()`,
// i.e. when an active pipelineController completes execution, it will re-execute with the
// most recently declared state if it has been bumped.
func (m *ppsMaster) run() {
	// close m.eventCh after all cancels have returned and therefore all pollers
	// (which are what write to m.eventCh) have exited
	m.eventCh = make(chan *pipelineEvent, 1)
	defer close(m.eventCh)
	defer m.cancelPCs()
	// start pollers in the background--cancel functions ensure poll/monitor
	// goroutines all definitely stop (either because cancelXYZ returns or because
	// the binary panics)
	m.startPipelinePoller()
	defer m.cancelPipelinePoller()
	m.startPipelinePodsPoller()
	defer m.cancelPipelinePodsPoller()
	m.startPipelineWatcher()
	defer m.cancelPipelineWatcher()

eventLoop:
	for {
		select {
		case e := <-m.eventCh:
			func(e *pipelineEvent) {
				m.pcMgr.Lock()
				defer m.pcMgr.Unlock()
				if pc, ok := m.pcMgr.pcs[e.pipeline]; ok {
					pc.Bump(e.timestamp) // raises flag in pipelineController to run again whenever it finishes
				} else {
					// pc's ctx is cancelled in pipelineController.tryFinish(), to avoid leaking resources
					pcCtx, pcCancel := context.WithCancel(m.masterCtx)
					pc = m.newPipelineController(pcCtx, pcCancel, e.pipeline)
					m.pcMgr.pcs[e.pipeline] = pc
					go pc.Start(e.timestamp)
				}
			}(e)
		case <-m.masterCtx.Done():
			break eventLoop
		}
	}
}

// setPipelineState is a PPS-master-specific helper that wraps
// ppsutil.SetPipelineState in a trace
func setPipelineState(ctx context.Context, db *pachsql.DB, pipelines collection.PostgresCollection, specCommit *pfs.Commit, state pps.PipelineState, reason string) (retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/SetPipelineState", "pipeline", specCommit.Branch.Repo.Name, "new-state", state)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return ppsutil.SetPipelineState(ctx, db, pipelines,
		specCommit, nil, state, reason)
}

func (m *ppsMaster) cancelPCs() {
	m.pcMgr.Lock()
	defer m.pcMgr.Unlock()
	for _, pc := range m.pcMgr.pcs {
		pc.cancel()
	}
}

type PipelineStateManager interface {
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

type stateManager struct {
	db        *pachsql.DB
	pipelines collection.PostgresCollection
	txEnv     *txnenv.TransactionEnv
	pfsApi    pfsserver.APIServer
}

func newPipelineStateManager(
	db *pachsql.DB,
	pipelines collection.PostgresCollection,
	txEnv *txnenv.TransactionEnv,
	pfsApi pfsserver.APIServer) PipelineStateManager {
	return &stateManager{
		db:        db,
		pipelines: pipelines,
		txEnv:     txEnv,
		pfsApi:    pfsApi,
	}
}

// takes pc.ctx
func (sm *stateManager) FetchState(ctx context.Context, pipeline string) (*pps.PipelineInfo, context.Context, error) {
	// query pipelineInfo
	var pi *pps.PipelineInfo
	var err error
	if pi, err = sm.tryLoadLatestPipelineInfo(ctx, pipeline); err != nil && collection.IsErrNotFound(err) {
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

func (sm *stateManager) SetState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) error {
	if err := setPipelineState(ctx, sm.db, sm.pipelines, specCommit, state, reason); err != nil {
		// don't bother failing if we can't set the state
		return stepError{
			error: errors.Wrapf(err, "could not set pipeline state to %v"+
				"(you may need to restart pachd to un-stick the pipeline)", state),
			retry: true,
		}
	}
	return nil
}

func (sm *stateManager) TransitionState(ctx context.Context, specCommit *pfs.Commit, from []pps.PipelineState, to pps.PipelineState, reason string) (retErr error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx,
		"/pps.Master/TransitionPipelineState", "pipeline", specCommit.Branch.Repo.Name,
		"from-state", from, "to-state", to)
	defer func() {
		tracing.TagAnySpan(span, "err", retErr)
		tracing.FinishAnySpan(span)
	}()
	return ppsutil.SetPipelineState(ctx, sm.db, sm.pipelines,
		specCommit, from, to, reason)
}

func (sm *stateManager) Watch(ctx context.Context) (<-chan *watch.Event, func(), error) {
	pipelineWatcher, err := sm.pipelines.ReadOnly(ctx).Watch()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error creating watch")
	}
	return pipelineWatcher.Watch(), pipelineWatcher.Close, nil
}

func (sm *stateManager) ListPipelineInfo(ctx context.Context, f func(*pps.PipelineInfo) error) error {
	return ppsutil.ListPipelineInfo(ctx, sm.pipelines, nil, 0, f)
}

func (sm *stateManager) GetPipelineInfo(ctx context.Context, name string, version int) (*pps.PipelineInfo, error) {
	var pipelineInfo pps.PipelineInfo
	if err := sm.pipelines.ReadOnly(ctx).GetUniqueByIndex(
		ppsdb.PipelinesVersionIndex,
		ppsdb.VersionKey(name, uint64(version)),
		&pipelineInfo); err != nil {
		return nil, errors.Wrapf(err, "couldn't retrieve pipeline information")
	}
	return &pipelineInfo, nil
}

func (sm *stateManager) tryLoadLatestPipelineInfo(ctx context.Context, pipeline string) (*pps.PipelineInfo, error) {
	pi := &pps.PipelineInfo{}
	errCnt := 0
	err := backoff.RetryNotify(func() error {
		return sm.loadLatestPipelineInfo(ctx, pipeline, pi)
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

func (sm *stateManager) loadLatestPipelineInfo(ctx context.Context, pipeline string, message *pps.PipelineInfo) error {
	specCommit, err := ppsutil.FindPipelineSpecCommit(ctx, sm.pfsApi, *sm.txEnv, pipeline)
	if err != nil {
		return errors.Wrapf(err, "could not find spec commit for pipeline %q", pipeline)
	}
	// TODO: which context should be passed here? master's ctx?
	if err := sm.pipelines.ReadOnly(ctx).Get(specCommit, message); err != nil {
		return errors.Wrapf(err, "could not retrieve pipeline info for %q", pipeline)
	}
	return nil
}

type mockStateManager struct {
	specCommits map[string]string              // maps spec commit IDs to pipeline names
	pipelines   map[string]*pps.PipelineInfo   // maps pipeline names to PipelineInfos
	states      map[string][]pps.PipelineState // tracks all of the
	eChan       chan *watch.Event
}

func newMockStateManager() *mockStateManager {
	msm := &mockStateManager{}
	msm.reset()
	return msm
}

func (msm *mockStateManager) SetState(ctx context.Context, specCommit *pfs.Commit, state pps.PipelineState, reason string) error {
	if pipeline, ok := msm.specCommits[specCommit.ID]; ok {
		if pi, ok := msm.pipelines[pipeline]; ok {
			pi.State = state
			return nil
		}
		return errors.New("Ah no pipeline state!!")
	}
	return errors.New("Ah!!")
}

func (msm *mockStateManager) TransitionState(ctx context.Context, specCommit *pfs.Commit, from []pps.PipelineState, to pps.PipelineState, reason string) (retErr error) {
	return msm.SetState(ctx, specCommit, to, reason)
}

func (msm *mockStateManager) FetchState(ctx context.Context, pipeline string) (*pps.PipelineInfo, context.Context, error) {
	if pi, ok := msm.pipelines[pipeline]; ok {
		return pi, ctx, nil
	}
	// TODO: make error realistic
	return nil, nil, errors.New("Nothing here!")
}

func (msm *mockStateManager) Watch(ctx context.Context) (<-chan *watch.Event, func(), error) {
	return msm.eChan, func() {
		close(msm.eChan)
	}, nil
}

func (msm *mockStateManager) ListPipelineInfo(ctx context.Context, f func(*pps.PipelineInfo) error) error {
	for _, pi := range msm.pipelines {
		if err := f(pi); err != nil {
			return err
		}
	}
	return nil
}

func (msm *mockStateManager) GetPipelineInfo(ctx context.Context, name string, version int) (*pps.PipelineInfo, error) {
	if pi, ok := msm.pipelines[name]; ok {
		return pi, nil
	}
	return nil, errors.New("not found")
}

func (msm *mockStateManager) upsertPipeline(pi *pps.PipelineInfo) *pfs.Commit {
	msm.pipelines[pi.Pipeline.Name] = pi
	mockSpecCommit := client.NewCommit(pi.Pipeline.Name, "master", uuid.NewWithoutDashes())
	msm.specCommits[mockSpecCommit.ID] = pi.Pipeline.Name
	if ss, ok := msm.states[pi.Pipeline.Name]; ok {
		ss = append(ss, pi.State)
	} else {
		msm.states[pi.Pipeline.Name] = []pps.PipelineState{pi.State}
	}
	msm.pushWatchEvent(&watch.Event{
		Key:  []byte(fmt.Sprintf("%s@%s", pi.Pipeline.Name, mockSpecCommit.ID)),
		Type: watch.EventPut,
	})
	return mockSpecCommit
}

func (msm *mockStateManager) pushWatchEvent(ev *watch.Event) {
	msm.eChan <- ev
}

func (msm *mockStateManager) reset() {
	msm.specCommits = make(map[string]string)
	msm.pipelines = make(map[string]*pps.PipelineInfo)
	msm.states = make(map[string][]pps.PipelineState)
	msm.eChan = make(chan *watch.Event)
}
