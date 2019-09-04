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

type DummyDriver struct {
	options    *DummyOptions
	etcdClient *etcd.Client
}

type DummyOptions struct {
	NumWorkers int
	EtcdPrefix string
}

// Not used - forces a compile-time error in this file if DummyDriver does not
// implement Driver
func ensureInterface() Driver {
	return &DummyDriver{}
}

func NewDummyDriver(etcdClient *etcd.Client, options *DummyOptions) *DummyDriver {
	return &DummyDriver{
		options:    options,
		etcdClient: etcdClient,
	}
}

func (dd *DummyDriver) WithCtx(context.Context) Driver {
	result := &DummyDriver{}
	*result = *dd
	return result
}

func (dd *DummyDriver) Jobs() col.Collection {
	return ppsdb.Jobs(dd.etcdClient, dd.options.EtcdPrefix)
}

func (dd *DummyDriver) Pipelines() col.Collection {
	return ppsdb.Pipelines(dd.etcdClient, dd.options.EtcdPrefix)
}

func (dd *DummyDriver) Plans() col.Collection {
	return col.NewCollection(dd.etcdClient, path.Join(dd.options.EtcdPrefix, planPrefix), nil, &common.Plan{}, nil, nil)
}

func (dd *DummyDriver) Shards() col.Collection {
	return col.NewCollection(dd.etcdClient, path.Join(dd.options.EtcdPrefix, shardPrefix), nil, &common.ShardInfo{}, nil, nil)
}

func (dd *DummyDriver) Chunks(jobID string) col.Collection {
	return col.NewCollection(dd.etcdClient, path.Join(dd.options.EtcdPrefix, chunkPrefix, jobID), nil, &common.ChunkState{}, nil, nil)
}

func (dd *DummyDriver) Merges(jobID string) col.Collection {
	return col.NewCollection(dd.etcdClient, path.Join(dd.options.EtcdPrefix, mergePrefix, jobID), nil, &common.MergeState{}, nil, nil)
}

func (dd *DummyDriver) GetExpectedNumWorkers() (int, error) {
	return dd.options.NumWorkers, nil
}

func (dd *DummyDriver) WithProvisionedNode(ctx context.Context, data []*common.Input, inputTree *hashtree.Ordered, logger logs.TaggedLogger, cb func(*pps.ProcessStats) error) (*pps.ProcessStats, error) {
	stats := &pps.ProcessStats{}
	err := cb(stats)
	return stats, err
}

func (dd *DummyDriver) RunUserCode(context.Context, logs.TaggedLogger, []string, *pps.ProcessStats, *types.Duration) error {
	// Inherit and shadow this if you actually want to do something for user code
	return nil
}

func (dd *DummyDriver) RunUserErrorHandlingCode(context.Context, logs.TaggedLogger, []string, *pps.ProcessStats, *types.Duration) error {
	// Inherit and shadow this if you actually want to do something for user error-handling code
	return nil
}

func (dd *DummyDriver) DeleteJob(stm col.STM, jobPtr *pps.EtcdJobInfo) error {
	// The dummy version doesn't bother keeping JobCounts updated properly
	return dd.Jobs().ReadWrite(stm).Delete(jobPtr.Job.ID)
}

func (dd *DummyDriver) UpdateJobState(ctx context.Context, jobID string, statsCommit *pfs.Commit, state pps.JobState, reason string) error {
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

func (dd *DummyDriver) ReportUploadStats(time.Time, *pps.ProcessStats, logs.TaggedLogger) {
	return
}

func (dd *DummyDriver) NewSTM(ctx context.Context, cb func(col.STM) error) (*etcd.TxnResponse, error) {
	return col.NewSTM(ctx, dd.etcdClient, cb)
}
