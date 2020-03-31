package driver

import (
	"context"
	"path/filepath"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
	"github.com/pachyderm/pachyderm/src/server/worker/cache"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

// MockOptions is a basic data struct containing options used by the MockDriver
type MockOptions struct {
	NumWorkers   int
	NumShards    int
	EtcdPrefix   string
	PipelineInfo *pps.PipelineInfo
	HashtreePath string
}

// MockDriver is an implementation of the Driver interface for use by tests.
// Complicated operations are short-circuited, but etcd operations should still
// work through this.
type MockDriver struct {
	ctx        context.Context
	options    *MockOptions
	etcdClient *etcd.Client

	chunkCaches, chunkStatsCaches cache.WorkerCache
}

// Not used - forces a compile-time error in this file if MockDriver does not
// implement Driver
var _ Driver = &MockDriver{}

// NewMockDriver constructs a MockDriver using the specified fields.
func NewMockDriver(etcdClient *etcd.Client, userOptions *MockOptions) *MockDriver {
	options := &MockOptions{}
	*options = *userOptions

	if options.NumWorkers == 0 {
		options.NumWorkers = 1
	}

	if options.NumShards == 0 {
		options.NumShards = 1
	}

	md := &MockDriver{
		ctx:        context.Background(),
		options:    options,
		etcdClient: etcdClient,
	}

	if options.HashtreePath != "" {
		md.chunkCaches = cache.NewWorkerCache(filepath.Join(options.HashtreePath, "chunk"))
		md.chunkStatsCaches = cache.NewWorkerCache(filepath.Join(options.HashtreePath, "chunkStats"))
	}

	return md
}

// WithContext does nothing aside from cloning the current MockDriver since there
// is no pachClient configured.
func (md *MockDriver) WithContext(ctx context.Context) Driver {
	result := &MockDriver{}
	*result = *md
	result.ctx = ctx
	return result
}

// Jobs returns a collection for the PPS jobs data in etcd
func (md *MockDriver) Jobs() col.Collection {
	return ppsdb.Jobs(md.etcdClient, md.options.EtcdPrefix)
}

// Pipelines returns a collection for the PPS Pipelines data in etcd
func (md *MockDriver) Pipelines() col.Collection {
	return ppsdb.Pipelines(md.etcdClient, md.options.EtcdPrefix)
}

// NewTaskWorker returns a work.Worker instance that can be used for running pipeline tasks.
func (md *MockDriver) NewTaskWorker() *work.Worker {
	return work.NewWorker(md.etcdClient, md.options.EtcdPrefix, workNamespace(md.options.PipelineInfo))
}

// NewTaskQueue returns a work.TaskQueue instance that can be used for distributing pipeline tasks.
func (md *MockDriver) NewTaskQueue() (*work.TaskQueue, error) {
	return work.NewTaskQueue(md.ctx, md.etcdClient, md.options.EtcdPrefix, workNamespace(md.options.PipelineInfo))
}

// PipelineInfo returns the pipeline configuration that the driver was
// initialized with.
func (md *MockDriver) PipelineInfo() *pps.PipelineInfo {
	return md.options.PipelineInfo
}

func (md *MockDriver) Namespace() string {
	return "namespace"
}

// ChunkCaches returns a cache.WorkerCache instance that can be used for caching
// hashtrees in the worker across multiple jobs. If no hashtree storage is
// specified in the MockDriver options, this will be nil.
func (md *MockDriver) ChunkCaches() cache.WorkerCache {
	return md.chunkCaches
}

// ChunkStatsCaches returns a cache.WorkerCache instance that can be used for
// caching hashtrees in the worker across multiple jobs. If no hashtree storage
// is specified in the MockDriver options, this will be nil.
func (md *MockDriver) ChunkStatsCaches() cache.WorkerCache {
	return md.chunkStatsCaches
}

// WithDatumCache calls the given callback with two hashtree merge caches, one
// for datums and one for datum stats. The lifetime of these caches will be
// bound to the callback, and any resources will be cleaned up upon return.
func (md *MockDriver) WithDatumCache(cb func(*hashtree.MergeCache, *hashtree.MergeCache) error) (retErr error) {
	return withDatumCache(md.options.HashtreePath, cb)
}

// InputDir returns the path used to hold the input filesets. Inherit and shadow
// this if you want to actually load data somewhere (make sure that this is
// unique so that tests don't collide).
func (md *MockDriver) InputDir() string {
	return "/pfs"
}

// HashtreeDir returns the path used to hold hashtrees. This is set in the
// options.
func (md *MockDriver) HashtreeDir() string {
	return md.options.HashtreePath
}

// PachClient returns the pachd API client for the driver. This is always `nil`
// for a MockDriver, but you can inherit and shadow this if you want some other
// value.
func (md *MockDriver) PachClient() *client.APIClient {
	return nil
}

// GetExpectedNumWorkers returns the configured number of workers
func (md *MockDriver) GetExpectedNumWorkers() (uint64, error) {
	return uint64(md.options.NumWorkers), nil
}

// NumShards returns the number of hashtree shards configured for the pipeline
func (md *MockDriver) NumShards() int64 {
	return int64(md.options.NumShards)
}

// WithData doesn't do anything except call the given callback.  Inherit and
// shadow this if you actually want to load some data onto the filesystem.
// Make sure to implement this in terms of the `InputDir` method.
func (md *MockDriver) WithData(
	data []*common.Input,
	inputTree *hashtree.Ordered,
	logger logs.TaggedLogger,
	cb func(string, *pps.ProcessStats) error,
) (*pps.ProcessStats, error) {
	stats := &pps.ProcessStats{}
	err := cb("", stats)
	return stats, err
}

// WithActiveData doesn't do anything except call the given callback. Inherit
// and shadow this if you actually want to load some data onto the filesystem.
// Make sure to implement this in terms of the `InputDir` method.
func (md *MockDriver) WithActiveData(inputs []*common.Input, dir string, cb func() error) error {
	return cb()
}

// RunUserCode does nothing.  Inherit and shadow this if you actually want to
// do something for user code
func (md *MockDriver) RunUserCode(logs.TaggedLogger, []string, *pps.ProcessStats, *types.Duration) error {
	return nil
}

// RunUserErrorHandlingCode does nothing.  Inherit and shadow this if you
// actually want to do something for user error-handling code
func (md *MockDriver) RunUserErrorHandlingCode(logs.TaggedLogger, []string, *pps.ProcessStats, *types.Duration) error {
	return nil
}

// UploadOutput does nothing. Inherit and shadow this if you actually want it to
// do something.
func (md *MockDriver) UploadOutput(string, string, logs.TaggedLogger, []*common.Input, *pps.ProcessStats, *hashtree.Ordered) ([]byte, error) {
	return []byte{}, nil
}

// DeleteJob will delete the given job entry from etcd.
func (md *MockDriver) DeleteJob(stm col.STM, jobPtr *pps.EtcdJobInfo) error {
	// The dummy version doesn't bother keeping JobCounts updated properly
	return md.Jobs().ReadWrite(stm).Delete(jobPtr.Job.ID)
}

// UpdateJobState will update the given job's state in etcd.
func (md *MockDriver) UpdateJobState(jobID string, state pps.JobState, reason string) error {
	// The dummy version doesn't bother with stats commits
	_, err := md.NewSTM(func(stm col.STM) error {
		jobPtr := &pps.EtcdJobInfo{}
		if err := md.Jobs().ReadWrite(stm).Get(jobID, jobPtr); err != nil {
			return err
		}
		return ppsutil.UpdateJobState(md.Pipelines().ReadWrite(stm), md.Jobs().ReadWrite(stm), jobPtr, state, reason)
	})
	return err
}

// ReportUploadStats does nothing.
func (md *MockDriver) ReportUploadStats(time.Time, *pps.ProcessStats, logs.TaggedLogger) {
}

// NewSTM calls the given callback under a new STM using the configured etcd
// client.
func (md *MockDriver) NewSTM(cb func(col.STM) error) (*etcd.TxnResponse, error) {
	return col.NewSTM(md.ctx, md.etcdClient, cb)
}
