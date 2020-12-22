package transform

import (
	"context"
	"path/filepath"

	etcd "github.com/coreos/etcd/clientv3"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsconsts"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/work"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

func defaultPipelineInfo() *pps.PipelineInfo {
	name := "testPipeline"
	return &pps.PipelineInfo{
		Pipeline:     client.NewPipeline(name),
		OutputBranch: "master",
		Transform: &pps.Transform{
			Cmd:        []string{"bash"},
			Stdin:      []string{"cp inputRepo/* out"},
			WorkingDir: client.PPSInputPrefix,
		},
		ParallelismSpec: &pps.ParallelismSpec{
			Constant: 1,
		},
		ResourceRequests: &pps.ResourceSpec{
			Memory: "100M",
			Cpu:    0.5,
		},
		Input: &pps.Input{
			Pfs: &pps.PFSInput{
				Name:   "inputRepo",
				Repo:   "inputRepo",
				Branch: "master",
				Glob:   "/*",
			},
		},
		SpecCommit: client.NewCommit(ppsconsts.SpecRepo, name),
	}
}

type testEnv struct {
	*testpachd.RealEnv
	logger *logs.MockLogger
	driver driver.Driver
}

// testDriver is identical to a real driver except it overloads egress, which is
// tricky to do in a test environment
type testDriver struct {
	inner driver.Driver
}

// Fuck golang
func (td *testDriver) Jobs() col.Collection {
	return td.inner.Jobs()
}
func (td *testDriver) Pipelines() col.Collection {
	return td.inner.Pipelines()
}
func (td *testDriver) NewTaskWorker() *work.Worker {
	return td.inner.NewTaskWorker()
}
func (td *testDriver) NewTaskQueue() (*work.TaskQueue, error) {
	return td.inner.NewTaskQueue()
}
func (td *testDriver) PipelineInfo() *pps.PipelineInfo {
	return td.inner.PipelineInfo()
}
func (td *testDriver) Namespace() string {
	return td.inner.Namespace()
}
func (td *testDriver) InputDir() string {
	return td.inner.InputDir()
}
func (td *testDriver) PachClient() *client.APIClient {
	return td.inner.PachClient()
}
func (td *testDriver) ExpectedNumWorkers() (int64, error) {
	return td.inner.ExpectedNumWorkers()
}
func (td *testDriver) WithContext(ctx context.Context) driver.Driver {
	return &testDriver{td.inner.WithContext(ctx)}
}
func (td *testDriver) WithActiveData(inputs []*common.Input, dir string, cb func() error) error {
	return td.inner.WithActiveData(inputs, dir, cb)
}
func (td *testDriver) UserCodeEnv(job string, commit *pfs.Commit, inputs []*common.Input) []string {
	return td.inner.UserCodeEnv(job, commit, inputs)
}
func (td *testDriver) RunUserCode(ctx context.Context, logger logs.TaggedLogger, env []string) error {
	return td.inner.RunUserCode(ctx, logger, env)
}
func (td *testDriver) RunUserErrorHandlingCode(ctx context.Context, logger logs.TaggedLogger, env []string) error {
	return td.inner.RunUserErrorHandlingCode(ctx, logger, env)
}
func (td *testDriver) DeleteJob(stm col.STM, ji *pps.EtcdJobInfo) error {
	return td.inner.DeleteJob(stm, ji)
}
func (td *testDriver) UpdateJobState(job string, state pps.JobState, reason string) error {
	return td.inner.UpdateJobState(job, state, reason)
}
func (td *testDriver) NewSTM(cb func(col.STM) error) (*etcd.TxnResponse, error) {
	return td.inner.NewSTM(cb)
}

// withTestEnv provides a test env with etcd and pachd instances and connected
// clients, plus a worker driver for performing worker operations.
func withTestEnv(db *sqlx.DB, pipelineInfo *pps.PipelineInfo, cb func(*testEnv) error) error {
	return testpachd.WithRealEnv(db, func(realEnv *testpachd.RealEnv) error {
		logger := logs.NewMockLogger()
		workerDir := filepath.Join(realEnv.Directory, "worker")
		driver, err := driver.NewDriver(
			pipelineInfo,
			realEnv.PachClient,
			realEnv.EtcdClient,
			"/pachyderm_test",
			workerDir,
			"namespace",
		)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(realEnv.PachClient.Ctx())
		defer cancel()
		driver = driver.WithContext(ctx)

		env := &testEnv{
			RealEnv: realEnv,
			logger:  logger,
			driver:  &testDriver{driver},
		}

		return cb(env)
	})
}
