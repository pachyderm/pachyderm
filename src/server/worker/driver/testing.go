package driver

import (
	"context"
	"path"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/logs"
)

type MockOptions struct {
	NumWorkers int
	EtcdPrefix string
}

type MockDriver struct {
	options    *MockOptions
	etcdClient *etcd.Client
}

// Not used - forces a compile-time error in this file if MockDriver does not
// implement Driver
var _ Driver = &MockDriver{}

func NewMockDriver(etcdClient *etcd.Client, options *MockOptions) *MockDriver {
	return &MockDriver{
		options:    options,
		etcdClient: etcdClient,
	}
}

func (dd *MockDriver) WithCtx(context.Context) Driver {
	result := &MockDriver{}
	*result = *dd
	return result
}

func (dd *MockDriver) Jobs() col.Collection {
	return ppsdb.Jobs(dd.etcdClient, dd.options.EtcdPrefix)
}

func (dd *MockDriver) Pipelines() col.Collection {
	return ppsdb.Pipelines(dd.etcdClient, dd.options.EtcdPrefix)
}

func (dd *MockDriver) Plans() col.Collection {
	return col.NewCollection(dd.etcdClient, path.Join(dd.options.EtcdPrefix, planPrefix), nil, &common.Plan{}, nil, nil)
}

func (dd *MockDriver) Shards() col.Collection {
	return col.NewCollection(dd.etcdClient, path.Join(dd.options.EtcdPrefix, shardPrefix), nil, &common.ShardInfo{}, nil, nil)
}

func (dd *MockDriver) Chunks(jobID string) col.Collection {
	return col.NewCollection(dd.etcdClient, path.Join(dd.options.EtcdPrefix, chunkPrefix, jobID), nil, &common.ChunkState{}, nil, nil)
}

func (dd *MockDriver) Merges(jobID string) col.Collection {
	return col.NewCollection(dd.etcdClient, path.Join(dd.options.EtcdPrefix, mergePrefix, jobID), nil, &common.MergeState{}, nil, nil)
}

func (dd *MockDriver) GetExpectedNumWorkers() (int, error) {
	return dd.options.NumWorkers, nil
}

func (dd *MockDriver) WithProvisionedNode(ctx context.Context, data []*common.Input, inputTree *hashtree.Ordered, logger logs.TaggedLogger, cb func(*pps.ProcessStats) error) (*pps.ProcessStats, error) {
	stats := &pps.ProcessStats{}
	err := cb(stats)
	return stats, err
}

func (dd *MockDriver) RunUserCode(context.Context, logs.TaggedLogger, []string, *pps.ProcessStats, *types.Duration) error {
	// Inherit and shadow this if you actually want to do something for user code
	return nil
}

func (dd *MockDriver) RunUserErrorHandlingCode(context.Context, logs.TaggedLogger, []string, *pps.ProcessStats, *types.Duration) error {
	// Inherit and shadow this if you actually want to do something for user error-handling code
	return nil
}

func (dd *MockDriver) DeleteJob(stm col.STM, jobPtr *pps.EtcdJobInfo) error {
	// The dummy version doesn't bother keeping JobCounts updated properly
	return dd.Jobs().ReadWrite(stm).Delete(jobPtr.Job.ID)
}

func (dd *MockDriver) UpdateJobState(ctx context.Context, jobID string, statsCommit *pfs.Commit, state pps.JobState, reason string) error {
	// The dummy version doesn't bother with stats commits
	_, err := dd.NewSTM(ctx, func(stm col.STM) error {
		jobPtr := &pps.EtcdJobInfo{}
		if err := dd.Jobs().ReadWrite(stm).Get(jobID, jobPtr); err != nil {
			return err
		}
		return ppsutil.UpdateJobState(dd.Pipelines().ReadWrite(stm), dd.Jobs().ReadWrite(stm), jobPtr, state, reason)
	})
	return err
}

func (dd *MockDriver) ReportUploadStats(time.Time, *pps.ProcessStats, logs.TaggedLogger) {
	return
}

func (dd *MockDriver) NewSTM(ctx context.Context, cb func(col.STM) error) (*etcd.TxnResponse, error) {
	return col.NewSTM(ctx, dd.etcdClient, cb)
}
