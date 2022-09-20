package transform

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/driver"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/logs"
)

// Set this to true to enable worker log statements to go to stdout
const debug = false

func defaultPipelineInfo() *pps.PipelineInfo {
	name := "testPipeline"
	return &pps.PipelineInfo{
		Pipeline: client.NewPipeline(name),
		Details: &pps.PipelineInfo_Details{
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
		},
	}
}

type testEnv struct {
	*realenv.RealEnv
	logger *logs.MockLogger
	driver driver.Driver
}

// testDriver is identical to a real driver except it overloads egress, which is
// tricky to do in a test environment
type testDriver struct {
	inner driver.Driver
}

func (td *testDriver) Jobs() col.PostgresCollection {
	return td.inner.Jobs()
}
func (td *testDriver) Pipelines() col.PostgresCollection {
	return td.inner.Pipelines()
}
func (td *testDriver) NewTaskSource() task.Source {
	return td.inner.NewTaskSource()
}
func (td *testDriver) NewTaskDoer(groupID string, cache task.Cache) task.Doer {
	return td.inner.NewTaskDoer(groupID, cache)
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
	res, err := td.inner.ExpectedNumWorkers()
	return res, errors.EnsureStack(err)
}
func (td *testDriver) WithContext(ctx context.Context) driver.Driver {
	return &testDriver{td.inner.WithContext(ctx)}
}
func (td *testDriver) WithActiveData(inputs []*common.Input, dir string, cb func() error) error {
	return errors.EnsureStack(td.inner.WithActiveData(inputs, dir, cb))
}
func (td *testDriver) UserCodeEnv(jobID string, commit *pfs.Commit, inputs []*common.Input) []string {
	return td.inner.UserCodeEnv(jobID, commit, inputs)
}
func (td *testDriver) RunUserCode(ctx context.Context, logger logs.TaggedLogger, env []string) error {
	return errors.EnsureStack(td.inner.RunUserCode(ctx, logger, env))
}
func (td *testDriver) RunUserErrorHandlingCode(ctx context.Context, logger logs.TaggedLogger, env []string) error {
	return errors.EnsureStack(td.inner.RunUserErrorHandlingCode(ctx, logger, env))
}
func (td *testDriver) DeleteJob(sqlTx *pachsql.Tx, ji *pps.JobInfo) error {
	return errors.EnsureStack(td.inner.DeleteJob(sqlTx, ji))
}
func (td *testDriver) UpdateJobState(job *pps.Job, state pps.JobState, reason string) error {
	return errors.EnsureStack(td.inner.UpdateJobState(job, state, reason))
}
func (td *testDriver) NewSQLTx(cb func(*pachsql.Tx) error) error {
	return errors.EnsureStack(td.inner.NewSQLTx(cb))
}
func (td *testDriver) GetContainerImageID(ctx context.Context, containerName string) (string, error) {
	return "mockImage", nil
}

// newTestEnv provides a test env with etcd and pachd instances and connected
// clients, plus a worker driver for performing worker operations.
func newTestEnv(t *testing.T, pipelineInfo *pps.PipelineInfo, realEnv *realenv.RealEnv) *testEnv {
	logger := logs.NewMockLogger()
	if debug {
		logger.Writer = os.Stdout
	}
	workerDir := filepath.Join(realEnv.Directory, "worker")
	driver, err := driver.NewDriver(
		realEnv.ServiceEnv,
		realEnv.PachClient,
		pipelineInfo,
		workerDir,
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(realEnv.PachClient.Ctx())
	t.Cleanup(cancel)
	driver = driver.WithContext(ctx)

	return &testEnv{
		RealEnv: realEnv,
		logger:  logger,
		driver:  &testDriver{driver},
	}
}
